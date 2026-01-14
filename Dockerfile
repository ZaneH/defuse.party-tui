FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git protoc curl

RUN curl -sSL "https://github.com/bufbuild/buf/releases/download/v1.47.2/buf-Linux-x86_64" -o /usr/local/bin/buf && \
    chmod +x /usr/local/bin/buf

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

COPY . .

RUN buf generate

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o tui-server ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates python3

WORKDIR /root/

COPY --from=builder /app/tui-server .
COPY web /root/web
COPY start.sh .

RUN chmod +x start.sh

EXPOSE 2222 8080

CMD ["./start.sh"]
