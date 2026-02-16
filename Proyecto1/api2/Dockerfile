FROM golang:1.23-alpine AS builder
WORKDIR /app
RUN go mod init api2 && go mod tidy
COPY *.go ./
RUN go build -o main .

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8002
CMD ["./main"]
