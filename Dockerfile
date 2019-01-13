
# Multi-stage build setup (https://docs.docker.com/develop/develop-images/multistage-build/)

# Stage 1 (to create a "build" image, ~850MB)
FROM golang:1.10.1 AS builder
RUN go version

MAINTAINER Victor Jungbauer <victor.jungbauer@gmail.com>
RUN apt-get update -qq
RUN apt-get install -y -qq git curl wget

# install npm
RUN apt-get install -y -qq npm
RUN ln -s /usr/bin/nodejs /usr/bin/node

# install bower
RUN npm install --global bower

COPY . /go/src/github.com/vic999/pirrigo/
WORKDIR /go/src/github.com/vic999/pirrigo/
RUN set -x && \
    go get -v

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -a -o app .

# Stage 2 (to create a downsized "container executable", ~7MB)

# If you need SSL certificates for HTTPS, replace `FROM SCRATCH` with:
#
#   FROM alpine:3.7
#   RUN apk --no-cache add ca-certificates
#
FROM scratch
WORKDIR /root/
COPY --from=builder /go/src/github.com/vic999/pirrigo/app .

EXPOSE 8000
ENTRYPOINT ["./app"]

