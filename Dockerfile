FROM quay.io/giantswarm/alpine:3.17.3-giantswarm

USER giantswarm
COPY ./apptestctl /usr/local/bin/apptestctl

ENTRYPOINT ["apptestctl"]
