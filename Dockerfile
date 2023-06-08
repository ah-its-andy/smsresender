FROM alpine:3.15.0

WORKDIR /app/

COPY exec exec

ENTRYPOINT ["/app/exec"]