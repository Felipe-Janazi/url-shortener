package model

import "time"

// URL representa um registro na tabela 'urls' do PostgreSQL.
// As tags `db:` são usadas pelo pgx para mapear colunas automaticamente.
// As tags `json:` controlam a serialização nas respostas HTTP.
type URL struct {
    ID          string    `db:"id"           json:"id"`
    Code        string    `db:"code"         json:"code"`          // ex: "aB3xK7q"
    OriginalURL string    `db:"original_url" json:"original_url"`  // URL completa original
    Clicks      int       `db:"clicks"       json:"clicks"`        // contador de acessos
    CreatedAt   time.Time `db:"created_at"   json:"created_at"`
    ExpiresAt   time.Time `db:"expires_at"   json:"expires_at,omitempty"`
}

// ShortenRequest é o body esperado no POST /shorten.
type ShortenRequest struct {
    URL string `json:"url"` // URL a ser encurtada
}

// ShortenResponse é o body retornado no POST /shorten.
type ShortenResponse struct {
    ShortURL    string `json:"short_url"`    // ex: https://srt.ly/aB3xK7q
    Code        string `json:"code"`         // só o código curto
    OriginalURL string `json:"original_url"` // URL original para confirmação
    ExpiresAt   string `json:"expires_at"`   // data de expiração
}