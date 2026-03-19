FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY main.go .
RUN go mod init api2
RUN go build -o api2 main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/api2 .
EXPOSE 8081
CMD ["./api2"]