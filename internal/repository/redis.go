package repository

import (
    "context"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

type RedisRepo struct {
    client *redis.Client
}

// NewRedis cria o cliente e verifica a conexão na inicialização.
func NewRedis(addr string) *RedisRepo {
    client := redis.NewClient(&redis.Options{
        Addr:         addr,
        DialTimeout:  3 * time.Second, // falha rápido se o Redis não responder
        ReadTimeout:  2 * time.Second,
        WriteTimeout: 2 * time.Second,
        PoolSize:     20, // deve ser >= MaxConns do Postgres para não ter gargalo
    })

    if err := client.Ping(context.Background()).Err(); err != nil {
        panic(fmt.Sprintf("redis não acessível: %v", err))
    }

    return &RedisRepo{client: client}
}

// Get busca a URL original no cache pelo código.
// Retorna erro se a chave não existir (cache miss) — o service trata isso como fallback.
func (r *RedisRepo) Get(ctx context.Context, code string) (string, error) {
    // Prefixo "url:" evita colisão com outras chaves se o Redis for compartilhado.
    return r.client.Get(ctx, "url:"+code).Result()
}

// Set armazena o mapeamento código → URL original com TTL.
// Após o TTL expirar, o Redis deleta a chave automaticamente — sem cleanup manual.
func (r *RedisRepo) Set(ctx context.Context, code, originalURL string, ttl time.Duration) error {
    return r.client.Set(ctx, "url:"+code, originalURL, ttl).Err()
}