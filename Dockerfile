# Build the go application into a binary
FROM golang:alpine as builder
WORKDIR /app
ADD . ./
RUN CGO_ENABLED=0 GOOS=linux go build -mod vendor -a -installsuffix cgo -o bin/gocache-server ./gocacheserver/main
RUN apk --update add --no-cache ca-certificates

FROM scratch
ENV APP_HOME=/app
ENV PORT=6379
ENV MAX_CACHE_SIZE=100000
ENV AUTOSAVE="false"
WORKDIR ${APP_HOME}
COPY --from=builder /app/bin/gocache-server ./bin/gocache-server
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
EXPOSE ${PORT}
ENTRYPOINT ["/app/bin/gocache-server"]