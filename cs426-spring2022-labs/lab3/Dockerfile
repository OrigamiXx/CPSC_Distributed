# syntax=docker/dockerfile:1

FROM golang:1.16-alpine AS build

WORKDIR /app

COPY . ./
RUN go mod download

RUN go build -o /video_rec_service_server ./video_rec_service/server/server.go


FROM alpine:latest

WORKDIR /

COPY --from=build /video_rec_service_server /video_rec_service

EXPOSE 8080

ENTRYPOINT [ "/video_rec_service" ]