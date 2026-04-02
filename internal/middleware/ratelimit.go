package middleware

import (
    "fmt"
    "net/http"
    "time"

    "github.com/redis/go-redis/v9"
)

// RateLimiter armazena o cliente Redis e as configurações do limite.
type RateLimiter struct {
    client   *redis.Client
    limit    int           // máximo de requisições permitidas na janela
    window   time.Duration // tamanho da janela de tempo (ex: 1 minuto)
}

// NewRateLimiter cria um RateLimiter.
//
// Exemplo de uso no main.go:
//
//	rl := middleware.NewRateLimiter(redisClient, 60, time.Minute)
//	mux.Handle("POST /shorten", rl.Limit(http.HandlerFunc(h.Shorten)))
//
// Parâmetros:
//   - client: cliente Redis já inicializado
//   - limit:  número máximo de requisições por janela por IP
//   - window: duração da janela (ex: time.Minute para 60 req/min)
func NewRateLimiter(client *redis.Client, limit int, window time.Duration) *RateLimiter {
    return &RateLimiter{
        client: client,
        limit:  limit,
        window: window,
    }
}

// Limit retorna um middleware que aplica rate limiting por IP.
//
// Algoritmo: sliding window counter usando Redis INCR + EXPIRE.
// Cada IP tem uma chave no Redis com um contador que incrementa a cada
// requisição. Na primeira requisição da janela, define o TTL igual à
// duração da janela. Quando o contador passa do limite, retorna 429.
//
// Vantagem deste algoritmo: simples, com apenas 2 comandos Redis por
// requisição (INCR + EXPIRE na primeira vez, só INCR nas demais).
// Limitação: não é um sliding window puro — o contador reseta de uma vez
// quando a chave expira. Para rate limiting mais preciso, use o
// algoritmo de token bucket com Lua scripts.
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Usa o IP real do cliente como identificador.
        // Em produção atrás do ALB, o IP real vem no header X-Forwarded-For.
        ip := r.RemoteAddr
        if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
            ip = forwarded
        }

        // Chave Redis: "ratelimit:<IP>" — única por IP por janela.
        key := fmt.Sprintf("ratelimit:%s", ip)
        ctx := r.Context()

        // INCR atomicamente incrementa o contador (ou cria com valor 1).
        // É atômico: sem race condition mesmo com múltiplos containers.
        count, err := rl.client.Incr(ctx, key).Result()
        if err != nil {
            // Se o Redis estiver fora, deixa a requisição passar.
            // Fail-open: melhor aceitar do que derrubar a aplicação inteira.
            next.ServeHTTP(w, r)
            return
        }

        // Na primeira requisição da janela (count == 1), define o TTL.
        // Nas demais, a chave já tem TTL — não precisamos redefinir.
        if count == 1 {
            rl.client.Expire(ctx, key, rl.window)
        }

        // Se o contador passou do limite, rejeita com 429 Too Many Requests.
        if int(count) > rl.limit {
            // Informa ao cliente quantos segundos deve esperar (padrão RFC 7231).
            ttl, _ := rl.client.TTL(ctx, key).Result()
            w.Header().Set("Retry-After", fmt.Sprintf("%.0f", ttl.Seconds()))
            w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limit))
            w.Header().Set("X-RateLimit-Remaining", "0")
            http.Error(w, "limite de requisições excedido", http.StatusTooManyRequests)
            return
        }

        // Adiciona headers informativos para o cliente saber o estado do limite.
        remaining := rl.limit - int(count)
        w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limit))
        w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

        // Passa para o próximo handler.
        next.ServeHTTP(w, r)
    })
}