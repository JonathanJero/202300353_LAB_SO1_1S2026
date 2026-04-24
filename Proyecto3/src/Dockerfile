FROM golang:1.21 as builder

WORKDIR /usr/src/app

RUN apt-get update && apt-get install -y protobuf-compiler && rm -rf /var/lib/apt/lists/*
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.32.0 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
ENV PATH="/go/bin:$PATH"

COPY proto/ /usr/src/app/proto/
RUN cd proto && protoc --go_out=. --go_opt=paths=source_relative \
           --go-grpc_out=. --go-grpc_opt=paths=source_relative \
           wartweets.proto

COPY go.mod go.sum* /usr/src/app/
COPY go-ingest/ /usr/src/app/go-ingest/
COPY go-writer/ /usr/src/app/go-writer/
COPY go-consumer/ /usr/src/app/go-consumer/

RUN go mod tidy
ARG TARGET_DIR
RUN CGO_ENABLED=0 GOOS=linux go build -o supervisor ./${TARGET_DIR}/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /usr/src/app/supervisor .
CMD ["./supervisor"]
