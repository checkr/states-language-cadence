FROM golang:1.13.1 as builder

COPY . /build
WORKDIR /build

ENV GOOS=linux
ENV CGO_ENABLED=0
ENV GO_BUILD_OPTS="-a -installsuffix cgo"

RUN make mod-vendor-unpack bin

# ----

FROM alpine:3.6

RUN apk add --no-cache libc6-compat ca-certificates

COPY --from=builder /build/bin/* /bgccore/bin/
