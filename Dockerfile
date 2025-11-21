# syntax=docker/dockerfile:1

FROM golang:1.23

WORKDIR /app


COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /game-server ./cmd/server/main.go

EXPOSE 8000

CMD ["/game-server"]
