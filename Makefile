# Variáveis configuráveis — podem ser sobrescritas na linha de comando.
# Exemplo: make migrate DB_URL=postgres://outro-host/db
DB_URL ?= postgres://user:pass@localhost:5432/shortener?sslmode=disable
BINARY  = api

# Sobe PostgreSQL e Redis em background, depois inicia a API com as env vars corretas.
# Usar 'make run' é o jeito mais rápido de subir tudo localmente.
.PHONY: run
run:
	docker compose up -d postgres redis
	DATABASE_URL=$(DB_URL) REDIS_URL=localhost:6379 go run ./cmd/api

# Roda todos os testes com o race detector ativado.
# -race detecta acessos concorrentes sem sincronização — fundamental em Go.
# -count=1 desativa o cache de resultados — garante que os testes sempre executam.
.PHONY: test
test:
	go test ./... -v -race -count=1

# Aplica todas as migrations pendentes usando golang-migrate.
# Instale com: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
.PHONY: migrate
migrate:
	migrate -path ./migrations -database "$(DB_URL)" up

# Reverte a última migration aplicada.
# Útil durante desenvolvimento quando você quer ajustar o schema.
.PHONY: migrate-down
migrate-down:
	migrate -path ./migrations -database "$(DB_URL)" down 1

# Compila o binário com as mesmas flags do Dockerfile.
# Use isso para testar o build de produção localmente antes de fazer push.
.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o $(BINARY) ./cmd/api

# Remove o binário compilado e derruba todos os containers com seus volumes.
# Use com cuidado: 'down -v' apaga os dados do PostgreSQL local.
.PHONY: clean
clean:
	rm -f $(BINARY)
	docker compose down -v