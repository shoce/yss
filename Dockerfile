
# https://hub.docker.com/_/golang/tags
FROM golang:1.25-alpine AS build
ENV CGO_ENABLED=0
RUN mkdir -p /yss/
COPY *.go go.mod go.sum /yss/
WORKDIR /yss/
RUN go version
RUN go get -v
RUN ls -l -a
RUN go build -o yss .
RUN ls -l -a


# https://hub.docker.com/_/alpine/tags
FROM alpine:3
RUN apk add --no-cache gcompat && ln -s -f -v ld-linux-x86-64.so.2 /lib/libresolv.so.2
COPY --from=build /yss/yss /bin/yss
RUN ls -l -a /bin/yss
RUN mkdir /yss/
WORKDIR /yss/
ENTRYPOINT ["/bin/yss"]


