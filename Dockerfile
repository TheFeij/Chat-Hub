# build stage
FROM golang:1.22.2-alpine3.19 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go

# run stage
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/main /app
COPY config/ /app/config/
COPY static/ /app/static/


EXPOSE 8080
CMD ["/app/main"]