FROM quay.io/giantswarm/alpine:3.18.4-giantswarm

USER giantswarm
COPY ./apptestctl /usr/local/bin/apptestctl

ENTRYPOINT ["apptestctl"]
