# Multi-stage build setup (https://docs.docker.com/develop/develop-images/multistage-build/)

# Stage 1 (to create a "build" image, ~850MB)
FROM golang:1.10.1 AS go-builder
RUN go version

MAINTAINER Victor Jungbauer <victor.jungbauer@gmail.com>

COPY . /go/src/github.com/vic999/pirrigo/
WORKDIR /go/src/github.com/vic999/pirrigo/
RUN set -x && \
    go get -v

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -a -o app .

# Stage 2 (Use Bower to handle Angular)
FROM node:8 AS bower-builder
ENV NODE_ENV=development
RUN mkdir -p /app
WORKDIR /app
RUN yarn global add bower gulp
COPY static/. .

RUN yarn install && bower install --allow-root
#COPY . /app
#RUN npm run build


# Stage 3 (to create a downsized "container executable", ~7MB)

# If you need SSL certificates for HTTPS, replace `FROM SCRATCH` with:
#
FROM alpine:3.7
RUN apk --no-cache add ca-certificates
#
#FROM scratch
WORKDIR /root/
COPY --from=go-builder /go/src/github.com/vic999/pirrigo/app .
COPY init .
COPY --from=bower-builder /app/ ./static
COPY templates .

EXPOSE 8000
CMD ["./app", "./init/appconfig.json"]
