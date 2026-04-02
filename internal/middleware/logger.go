package middleware

import (
    "log/slog"
    "net/http"
    "time"
)

// responseWriter é um wrapper sobre http.ResponseWriter que captura
// o status code escrito pelo handler. O http.ResponseWriter padrão
// não expõe o status depois que WriteHeader foi chamado.
type responseWriter struct {
    http.ResponseWriter        // embute o ResponseWriter original
    statusCode          int    // status HTTP capturado
    bytesWritten        int64  // bytes escritos no body
}

// WriteHeader intercepta a chamada original e salva o status code.
// Se o handler não chamar WriteHeader explicitamente, o Go usa 200 — 
// por isso inicializamos statusCode com 200 no construtor.
func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}

// Write intercepta a escrita do body para contabilizar bytes trafegados.
func (rw *responseWriter) Write(b []byte) (int, error) {
    n, err := rw.ResponseWriter.Write(b)
    rw.bytesWritten += int64(n)
    return n, err
}

// Logger retorna um middleware HTTP que registra cada requisição com:
// método, path, status, latência e bytes de resposta.
//
// Uso no main.go:
//
//	mux.Handle("/", middleware.Logger(outroHandler))
//
// O log usa slog (logger estruturado do Go 1.21+), que produz saída
// em JSON quando configurado — facilita ingestão no CloudWatch Logs.
func Logger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Registra o momento em que a requisição chegou.
        start := time.Now()

        // Envolve o ResponseWriter original com nosso wrapper
        // para capturar o status code e o tamanho da resposta.
        wrapped := &responseWriter{
            ResponseWriter: w,
            statusCode:     http.StatusOK, // default: 200
        }

        // Chama o próximo handler da cadeia com o writer interceptado.
        next.ServeHTTP(wrapped, r)

        // Calcula a latência total após o handler terminar.
        latency := time.Since(start)

        // Loga a requisição de forma estruturada.
        // Campos estruturados (key=value) são indexáveis no CloudWatch/Datadog.
        slog.Info("request",
            "method",   r.Method,
            "path",     r.URL.Path,
            "status",   wrapped.statusCode,
            "latency",  latency.String(),     // ex: "1.234ms"
            "bytes",    wrapped.bytesWritten,
            "remote",   r.RemoteAddr,
            "ua",       r.UserAgent(),
        )
    })
}