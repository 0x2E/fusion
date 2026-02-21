FROM alpine:3.21.0
ARG TARGETOS=linux
ARG TARGETARCH
LABEL org.opencontainers.image.source="https://github.com/patrickjmcd/reedme"

RUN addgroup -S reedme && adduser -S -D -H -h /reedme -G reedme reedme && \
    mkdir -p /data && chown -R reedme:reedme /data

WORKDIR /reedme
COPY --chown=reedme:reedme --chmod=755 build/reedme-${TARGETOS}-${TARGETARCH} ./reedme
EXPOSE 8080
VOLUME ["/data"]
# SQLite default path (used when REEDME_DATABASE_URL is not set)
ENV DB="/data/reedme.db"
# PostgreSQL: set REEDME_DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=disable to use PostgreSQL instead
HEALTHCHECK --interval=10s --timeout=3s --start-period=2s --retries=3 \
  CMD wget -q -O /dev/null http://127.0.0.1:8080/api/oidc/enabled || exit 1
# TODO: Temporarily run as root until legacy /data DB files owned by root are migrated.
CMD [ "./reedme" ]
