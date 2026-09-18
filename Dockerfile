# Runtime image - hanya berisi binary compiled
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata curl

WORKDIR /root/

COPY task-manager .

RUN mkdir -p /root/logs

EXPOSE 8080

CMD ["./task-manager"]
