FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -a -o server cmd/main.go

FROM gcr.io/distroless/base-debian10

COPY --from=builder /app/server /server

EXPOSE 8080 2222

CMD ["/server"]