package handler

import (
    "encoding/json"
    "net/http"
    "time"
)

// HealthResponse é o corpo retornado pelo endpoint de healthcheck.
// Inclui timestamp para facilitar o debug de quando o check foi feito.
type HealthResponse struct {
    Status    string `json:"status"`     // sempre "ok" se a app está viva
    Timestamp string `json:"timestamp"`  // horário UTC do check
}

// Health trata GET /health.
//
// Este endpoint é consultado pelo ALB (Application Load Balancer) da AWS
// e pelo próprio ECS para saber se o container está vivo e pronto para
// receber tráfego. Se retornar qualquer coisa diferente de 200, o ECS
// para de enviar requisições para esse container e o reinicia.
//
// Regra importante: este handler nunca deve fazer I/O (sem banco, sem Redis).
// Ele só precisa confirmar que o processo Go está respondendo.
// Checagens de dependências (DB, cache) pertencem a um endpoint /ready separado.
func (h *URLHandler) Health(w http.ResponseWriter, r *http.Request) {
    // Monta a resposta com o status e o timestamp atual em UTC.
    resp := HealthResponse{
        Status:    "ok",
        Timestamp: time.Now().UTC().Format(time.RFC3339),
    }

    // Retorna 200 com Content-Type correto.
    // O ALB interpreta qualquer código 2xx como "saudável".
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}