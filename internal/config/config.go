package config

import (
    "fmt"
    "os"
)

// Config agrupa todas as configurações da aplicação.
// Adicionar aqui qualquer nova variável de ambiente necessária.
type Config struct {
    Port        string // porta HTTP que o servidor vai escutar
    DatabaseURL string // connection string do PostgreSQL
    RedisURL    string // host:port do Redis
    BaseURL     string // domínio base para montar a URL curta (ex: https://srt.ly)
}

// Load lê as env vars e aborta se alguma obrigatória estiver faltando.
// Isso garante que a app nunca sobe em estado inválido.
func Load() *Config {
    return &Config{
        Port:        getEnv("PORT", "8080"),          // 8080 é o padrão
        DatabaseURL: mustGetEnv("DATABASE_URL"),       // obrigatória
        RedisURL:    mustGetEnv("REDIS_URL"),           // obrigatória
        BaseURL:     getEnv("BASE_URL", "http://localhost:8080"),
    }
}

// getEnv retorna a env var ou um valor padrão se não existir.
func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}

// mustGetEnv aborta a aplicação se a variável não estiver definida.
// Fail-fast: melhor crashar na inicialização que ter comportamento inesperado em runtime.
func mustGetEnv(key string) string {
    v := os.Getenv(key)
    if v == "" {
        panic(fmt.Sprintf("variável de ambiente obrigatória não definida: %s", key))
    }
    return v
}