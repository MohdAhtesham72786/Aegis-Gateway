FROM golang:1.20-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN apk add --no-cache git
COPY . .
RUN go build -o /aegis ./cmd/aegis

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build /aegis /usr/local/bin/aegis
WORKDIR /app
COPY ./policies /app/policies
EXPOSE 8080
CMD ["/usr/local/bin/aegis"]
