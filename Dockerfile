FROM gsoci.azurecr.io/giantswarm/alpine:3.20.3-giantswarm

ARG TARGETARCH

USER giantswarm
COPY ./apptestctl-linux-${TARGETARCH} /usr/local/bin/apptestctl

ENTRYPOINT ["apptestctl"]
