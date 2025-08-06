# build frontend
FROM node:24 AS fe
WORKDIR /src
RUN npm i -g pnpm
COPY .git .git/
COPY frontend ./frontend
COPY scripts.sh .
RUN ./scripts.sh build-frontend

# build backend
FROM golang:1.24 AS be
# Add Arguments for target OS and architecture (provided by buildx)
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src
COPY . ./
COPY --from=fe /src/frontend/build ./frontend/build/
RUN ./scripts.sh build-backend ${TARGETOS} ${TARGETARCH}

# deploy
FROM alpine:3.21.0
LABEL org.opencontainers.image.source="https://github.com/0x2E/fusion"
WORKDIR /fusion
COPY --from=be /src/build/fusion ./
EXPOSE 8080
RUN mkdir /data
ENV DB="/data/fusion.db"
CMD [ "./fusion" ]
