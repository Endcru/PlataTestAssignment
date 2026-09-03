FROM golang:alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/currency-exchange

ENV CONFIG_PATH=/app/config/docker.yaml

EXPOSE 8090

CMD ["./main"]