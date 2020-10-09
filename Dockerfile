FROM quay.io/giantswarm/alpine:3.11-giantswarm

USER giantswarm
COPY ./apptestctl /usr/local/bin/apptestctl

ENTRYPOINT ["apptestctl"]
