# build backend
FROM golang:1.22 as be
WORKDIR /src
COPY . ./
RUN go build -o fusion-server ./cmd/server/*

# build frontend
FROM node:21 as fe
WORKDIR /src
COPY ./frontend ./
RUN npm i && npm run build

# deploy
FROM debian:12
RUN apt-get update && apt-get install -y sqlite3
WORKDIR /fusion
COPY .env ./
COPY --from=be /src/fusion-server ./
COPY --from=fe /src/build ./frontend/
EXPOSE 8080
RUN mkdir /data
ENV DB="/data/fusion.db"
CMD [ "./fusion-server" ]

