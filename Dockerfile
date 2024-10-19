FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN go build -o scalingo-api-test ./cmd/api

EXPOSE 5000

CMD ["./scalingo-api-test"]