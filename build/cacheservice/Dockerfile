FROM golang:1.21-alpine AS build
WORKDIR /app
COPY . .
RUN apk add --no-cache git
RUN go build -o /app/bin/service ./cmd/orderservice


FROM alpine:3.17
COPY --from=build /app/bin/service /service
EXPOSE 50051
ENTRYPOINT ["/service"]