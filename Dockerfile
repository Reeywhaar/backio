FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod .
COPY main.go .
COPY internal/ internal/
RUN go build -o backio .

FROM alpine:latest
RUN apk add --no-cache rclone
COPY --from=builder /app/backio /backio
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh /backio
ENV PORT=8080
ENTRYPOINT ["/entrypoint.sh"]
