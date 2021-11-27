#!/bin/bash

MOCKCICD_SRCDIRPATH="tmp/src" \
MOCKCICD_GITREPOURL="https://github.com/jhwbarlow/instant-search-demo.git" \
MOCKCICD_GITBRANCH="master" \
MOCKCICD_IMAGENAME="myregistry.com/jhwbarlow/algolia-instant-search-demo" \
MOCKCICD_HELMCHARTPATH="chart/algolia-instant-search-demo" \
MOCKCICD_HELMK8SNAMESPACE="algolia" \
MOCKCICD_HELMRELEASENAME="algolia-instant-search-demo" \
MOCKCICD_INSTALLTIMEOUT="5m" \
MOCKCICD_POLLPERIOD="1m" \
go run main.go	