
# https://hub.docker.com/_/golang/tags
FROM golang:1.23.4-alpine3.20 AS build
RUN mkdir -p /root/yss/
COPY *.go go.mod /root/yss/
WORKDIR /root/yss/
RUN go version
RUN go get -v
RUN ls -l -a
RUN go build -o yss .
RUN ls -l -a


# https://hub.docker.com/_/alpine/tags
FROM alpine:3.20.3
RUN apk add --no-cache gcompat && ln -s -f -v ld-linux-x86-64.so.2 /lib/libresolv.so.2
COPY --from=build /root/yss/yss /bin/yss
RUN ls -l -a /bin/yss
WORKDIR /root/
ENTRYPOINT ["/bin/yss"]


