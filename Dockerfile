# ── Estágio 1: Compilação ─────────────────────────────────────────────────────
# Usa imagem com Go completo só para compilar — não vai para produção.
# Resultado: imagem de build ~600MB, imagem final ~10MB.
FROM golang:1.23-alpine AS builder
WORKDIR /app

# Copia go.mod e go.sum ANTES do código fonte.
# O Docker cacheia cada layer: se as dependências não mudaram,
# o 'go mod download' não roda de novo — build muito mais rápido.
COPY go.mod go.sum ./
RUN go mod download

# Agora copia o restante do código e compila.
COPY . .

# CGO_ENABLED=0  → binário estático, sem dependências de .so
# GOOS=linux     → compila para Linux (necessário se você desenvolve em macOS/Windows)
# -ldflags -s -w → remove tabela de símbolos e info de debug → binário menor
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o api ./cmd/api

# ── Estágio 2: Runtime ────────────────────────────────────────────────────────
# distroless/static: sem shell, sem apt, sem nada além do binário.
# Superfície de ataque mínima — ideal para produção em ECS.
FROM gcr.io/distroless/static-debian12

# Copia apenas o binário compilado do estágio anterior.
COPY --from=builder /app/api /api

# Documenta qual porta a app expõe.
# Não abre a porta — isso é responsabilidade do ECS Task Definition.
EXPOSE 8080

# Executa o binário diretamente, sem shell intermediário.
ENTRYPOINT ["/api"]