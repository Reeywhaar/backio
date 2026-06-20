FROM alpine:latest
RUN apk add --no-cache curl
COPY test.sh /test.sh
RUN chmod +x /test.sh
CMD ["/test.sh"]
