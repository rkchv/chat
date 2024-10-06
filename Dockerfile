FROM golang:1.20-alpine3.19 as builder
LABEL authors="ivansemeniv"

COPY . /rkchv/chat/src
WORKDIR /rkchv/chat/src

RUN go mod download
RUN go build -o ./bin/chat_server cmd/grpc-server/main.go

FROM alpine:3.19.2
WORKDIR /root/
COPY --from=builder /rkchv/chat/src/bin/chat_server .

CMD ["./chat_server"]
