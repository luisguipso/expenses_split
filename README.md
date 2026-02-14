# Contas

Aplicação web para divisão de despesas domésticas de forma proporcional ao salário de cada membro. Backend em Go, frontend em React + TypeScript.

## Funcionalidades

- Cadastro e autenticação de usuários (JWT)
- Criação de domicílios com convite por código
- Contas fixas e despesas variáveis (compartilhadas ou pessoais)
- Divisão proporcional por salário
- Cálculo de saldo: quem deve pagar e quem deve receber
- Acertos financeiros otimizados (mínimo de transferências)

## Pré-requisitos

- Go 1.23+
- Node.js 18+
- Docker e Docker Compose
- PostgreSQL 16 (via Docker)

## Setup

### 1. Subir o banco de dados

```bash
make docker-up
```

Isso inicia o PostgreSQL na porta `5432` com usuário `contas`, senha `contas` e banco `contas`.

### 2. Variáveis de ambiente

O servidor carrega variáveis de um arquivo `.env` na raiz do projeto. Crie um se quiser sobrescrever os valores padrão:

```env
PORT=8080
DATABASE_URL=postgres://contas:contas@localhost:5432/contas?sslmode=disable
JWT_SECRET=dev-secret-change-in-production
FRONTEND_URL=http://localhost:5173
```

Todos os valores acima já são os padrões — o `.env` é opcional para desenvolvimento local.

### 3. Instalar dependências do frontend

```bash
make web-install
```

## Executando

### Backend (modo desenvolvimento)

```bash
make dev
```

O servidor inicia em `http://localhost:8080`. As migrações do banco rodam automaticamente.

### Frontend (modo desenvolvimento)

```bash
make web-dev
```

O Vite inicia em `http://localhost:5173` com hot-reload.

### Backend + banco juntos

```bash
make up
```

Sobe o Docker (PostgreSQL) e inicia o servidor Go.

### Build de produção

```bash
# Backend
make build        # gera bin/api

# Frontend
make web-build    # gera web/dist/

# Docker (backend multi-stage)
docker compose up --build
```

## Testes

### Testes unitários

```bash
make test
```

Roda todos os testes unitários (sem dependência de banco).

### Testes de integração

Os testes de integração rodam contra um PostgreSQL real. É necessário criar o banco de teste primeiro:

```bash
# 1. Certifique-se de que o PostgreSQL está rodando
make docker-up

# 2. Crie o banco de teste (apenas uma vez)
docker exec -it contas-db-1 psql -U contas -c "CREATE DATABASE contas_test;"

# 3. Rode os testes
make test-integration
```

A variável `CONTAS_TEST_DATABASE_URL` é configurada automaticamente pelo Makefile para `postgres://contas:contas@localhost:5432/contas_test?sslmode=disable`.

Os testes de integração cobrem:
- Autenticação (registro, login, refresh token)
- Domicílios (criação, convite, salário)
- Categorias (CRUD)
- Contas fixas (CRUD, paid_by)
- Despesas (CRUD, filtros por mês/categoria)
- Resumo e acertos financeiros
- Fluxo completo (cross-feature)

### Todos os testes

```bash
make test && make test-integration
```

## Estrutura do projeto

```
├── cmd/api/             # Entrypoint do servidor
├── internal/
│   ├── domain/          # Entidades e interfaces
│   ├── handler/         # Handlers HTTP (Echo)
│   ├── repository/      # Acesso ao banco (pgx)
│   ├── service/         # Lógica de negócio
│   ├── middleware/       # Auth JWT
│   ├── migrate/         # Migrações embarcadas
│   ├── mock/            # Mocks para testes unitários
│   └── integration/     # Testes de integração
├── migrations/          # Arquivos de migração (referência)
├── pkg/                 # Pacotes compartilhados
├── web/                 # Frontend React + TypeScript
├── docker-compose.yml
├── Dockerfile
└── Makefile
```

## Migrações

As migrações são aplicadas automaticamente ao iniciar o servidor. Elas ficam embarcadas no binário via `go:embed` a partir de `internal/migrate/sql/`.

Para aplicar manualmente:

```bash
make migrate-up
make migrate-down
```
