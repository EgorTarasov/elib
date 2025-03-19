FROM golang:1.24-alpine AS builder


COPY . .
RUN GO_ENABLED=0  go build \
	-o /go/bin/main ./cmd/app/main.go


FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /go/bin/main .
RUN chown root:root main
