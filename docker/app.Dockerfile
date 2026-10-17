FROM golang:1.24.2-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o main main.go

FROM builder AS runner

WORKDIR /app

COPY --from=builder /app/main .

CMD ["./main"]
