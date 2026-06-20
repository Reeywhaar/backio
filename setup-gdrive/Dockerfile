FROM alpine:latest
RUN apk add --no-cache rclone socat
COPY setup.sh /setup.sh
RUN chmod +x /setup.sh
CMD ["/setup.sh"]
