FROM node:10
ARG USER=9999
ARG GROUP=9999
COPY --chown=${USER}:${GROUP} . /app/
WORKDIR /app
RUN npm install
USER ${USER}:${GROUP}
ENTRYPOINT [ "npm", "start" ]
