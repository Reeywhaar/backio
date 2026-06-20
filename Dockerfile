FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod .
COPY main.go .
RUN go build -o server .

FROM alpine:latest
RUN apk add --no-cache rclone
COPY --from=builder /app/server /server
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh /server
ENV PORT=8080
ENTRYPOINT ["/entrypoint.sh"]
