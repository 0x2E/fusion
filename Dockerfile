# build frontend
FROM node:21 as fe
WORKDIR /src
COPY ./frontend ./
RUN npm i && npm run build

# build backend
FROM golang:1.22 as be
WORKDIR /src
COPY . ./
COPY --from=fe /src/build ./frontend/build/
RUN go build -o fusion ./cmd/server/*

# deploy
FROM debian:12
RUN apt-get update && apt-get install -y sqlite3
WORKDIR /fusion
COPY .env ./
COPY --from=be /src/fusion ./
EXPOSE 8080
RUN mkdir /data
ENV DB="/data/fusion.db"
CMD [ "./fusion" ]

