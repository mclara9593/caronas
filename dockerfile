# Etapa 1: build
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copia primeiro só os arquivos de dependência pra aproveitar cache do Docker
COPY go.mod go.sum ./
RUN go mod download

# Copia o resto do código
COPY . .

# Compila o binário do servidor (ajuste o caminho se seu main.go do servidor
# estiver em outro lugar, ex: cmd/server/main.go)
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/servidor ./cmd/server

# Etapa 2: imagem final, só com o binário
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/servidor .

EXPOSE 9593

CMD ["./servidor", "-addr=:9593"]