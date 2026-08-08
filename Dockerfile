FROM golang:1.26-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -mod=vendor -o /out/relay ./cmd/relay

FROM alpine:3.20
RUN adduser -D -H -u 10001 relay
WORKDIR /app
COPY --from=build /out/relay /app/relay
USER relay
ENV LISTEN_ADDR=0.0.0.0:8080
EXPOSE 8080
ENTRYPOINT ["/app/relay"]
