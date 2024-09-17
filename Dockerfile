FROM gsoci.azurecr.io/giantswarm/alpine:3.20.3-giantswarm

USER giantswarm
COPY ./apptestctl /usr/local/bin/apptestctl

ENTRYPOINT ["apptestctl"]
