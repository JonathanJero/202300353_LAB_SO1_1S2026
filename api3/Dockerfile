FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY main.go .
RUN go mod init api3
RUN go build -o api3 main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/api3 .
EXPOSE 8080
CMD ["./api3"]