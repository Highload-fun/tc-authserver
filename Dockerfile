FROM golang:1.22-alpine

RUN apk add --no-cache wget

WORKDIR /build

COPY / .

RUN go build -o main main.go

EXPOSE 8080

CMD ["./main"]