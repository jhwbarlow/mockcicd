#!/bin/bash

set -e
set -o pipefail

readonly POLL_PERIOD="1m"
readonly GIT_REPO="https://github.com/algolia/instant-search-demo.git"
readonly GIT_BRANCH="master"
readonly SRC_DIR="tmp/src"
readonly IMAGE_NAME="myregistry.com/jhwbarlow/algolia-instant-search-demo"
readonly K8S_NAMESPACE="algolia"
readonly HELM_CHART_DIR="chart/algolia-instant-search-demo"
readonly HELM_TIMEOUT="5m"
readonly HELM_RELEASE_NAME="algolia-instant-search-demo"

base_dir="$(dirname "$(realpath "${BASH_SOURCE[0]}")")"
readonly BASE_DIR="$base_dir"
unset base_dir

_install_via_helm() {  
  local -r image_tag="$1"
 
  echo "Installing via Helm..."
  
  if ! helm upgrade \
        --install \
        --atomic \
        --create-namespace \
        --timeout "$HELM_TIMEOUT" \
        -n "$K8S_NAMESPACE" \
        --set image.repository="$IMAGE_NAME" \
        --set image.tag="$image_tag" \
        "$HELM_RELEASE_NAME" \
        "$HELM_CHART_DIR"
  then
    echo "ERROR: Helm upgrade failed!"
    return 1
  fi    
}

_build_image() {
  local -r image_tag="$1"

  echo "Building image..."
  
  if ! docker build \
        -f docker/Dockerfile \
        -t "${IMAGE_NAME}:${image_tag}" \
        "$SRC_DIR"
  then
    echo "ERROR: Docker build failed!"
    return 1
  fi
}

_push_image() {
  local -r image_tag="$1"

  echo "Pushing image..."

  if ! docker push "${IMAGE_NAME}:${image_tag}"
  then
    echo "ERROR: Docker push failed!"
    return 1
  fi
}

_poll_and_install() {
  if ! [ -d "$SRC_DIR" ]
  then
    echo "making src dir"
    mkdir -p "$SRC_DIR"
  else
    echo "cleaning src dir"
    rm -rf "$SRC_DIR"
    mkdir -p "$SRC_DIR"
  fi

  cd "$SRC_DIR" || return 1 # Should not fail as we ensure the directory already exists
  git clone -b "$GIT_BRANCH" --single-branch "$GIT_REPO" .

  local short_hash
  short_hash=$(git rev-parse --short HEAD)

  pushd .   
  cd "$BASE_DIR"  
  _build_image "$short_hash"
  _push_image "$short_hash"
  _install_via_helm "$short_hash"
  popd
  
  while :
  do
    local local_hash
    local_hash=$(git rev-parse HEAD)

    local remote_hash
    remote_hash=$(git ls-remote "$GIT_REPO" "$GIT_BRANCH" | cut -f1)

    echo "local repo at commit [${local_hash}] - remote at [${remote_hash}]"

    if [ "$remote_hash" != "$local_hash" ]
    then
      echo "change to remote master- remote now at [${remote_hash}]"
      git pull      
      short_hash=$(git rev-parse --short HEAD)

      pushd .
      cd "$BASE_DIR"
      if ! { _build_image "$short_hash" && \
            _push_image "$short_hash" && \
            _install_via_helm "$short_hash"
          }
      then
        echo "ERROR: Build and install of version [${short_hash}] failed - continuing to use old version"        
      fi

      popd            
    fi

    sleep "$POLL_PERIOD"
  done
}

#######################################

 _poll_and_install
