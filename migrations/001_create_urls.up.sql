-- 001_create_urls.up.sql

-- Habilita a extensão para gerar UUIDs automaticamente no Postgres.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Tabela principal que armazena todos os mapeamentos de URL.
CREATE TABLE urls (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    code         VARCHAR(20) UNIQUE NOT NULL, -- ex: "aB3xK7q"
    original_url TEXT        NOT NULL,
    clicks       INTEGER     NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at   TIMESTAMPTZ NOT NULL
);

-- Índice no campo 'code': toda busca de redirecionamento usa esse campo.
-- Sem esse índice, cada GET /{code} faria um full table scan na tabela inteira.
CREATE INDEX idx_urls_code ON urls(code);

-- Índice parcial para URLs já expiradas: facilita um job de limpeza periódica.
CREATE INDEX idx_urls_expired ON urls(expires_at) WHERE expires_at < NOW();