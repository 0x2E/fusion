FROM alpine:3.21.0
ARG TARGETOS=linux
ARG TARGETARCH
LABEL org.opencontainers.image.source="https://github.com/0x2E/fusion"

RUN addgroup -S fusion && adduser -S -D -H -h /fusion -G fusion fusion && \
    mkdir -p /data && chown -R fusion:fusion /data

WORKDIR /fusion
COPY --chown=fusion:fusion --chmod=755 build/fusion-${TARGETOS}-${TARGETARCH} ./fusion
EXPOSE 8080
VOLUME ["/data"]
ENV DB="/data/fusion.db"
HEALTHCHECK --interval=10s --timeout=3s --start-period=2s --retries=3 \
  CMD wget -q -O /dev/null http://127.0.0.1:8080/api/oidc/enabled || exit 1
# TODO: Temporarily run as root until legacy /data DB files owned by root are migrated.
CMD [ "./fusion" ]
