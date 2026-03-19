FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY main.go .
RUN go mod init api1
RUN go build -o api1 main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/api1 .
EXPOSE 8080
CMD ["./api1"]