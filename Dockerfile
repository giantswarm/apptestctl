FROM quay.io/giantswarm/alpine:3.15.0-giantswarm

USER giantswarm
COPY ./apptestctl /usr/local/bin/apptestctl

ENTRYPOINT ["apptestctl"]
