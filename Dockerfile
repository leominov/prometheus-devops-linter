FROM golang:1.15.8-alpine3.13 AS build

RUN apk --no-cache add git
RUN go get -v github.com/prometheus/promu

WORKDIR /go/github.com/leominov/prometheus-devops-linter
COPY . /go/github.com/leominov/prometheus-devops-linter
RUN /go/bin/promu build --prefix=/go

FROM alpine:3.13.2

COPY --from=build /go/prometheus-devops-linter /usr/bin/
