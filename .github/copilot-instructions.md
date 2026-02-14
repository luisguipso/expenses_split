# Copilot Instructions — Contas

## Engineering Principles

- Follow **SOLID** principles: single responsibility, open/closed, Liskov substitution, interface segregation, dependency inversion
- Apply **YAGNI** — don't build features or abstractions that aren't needed yet
- Apply **DRY** — extract shared logic, avoid code duplication across layers
- Use **clean architecture** with layers: domain → service → handler → repository
- Dependencies always point inward — all layers depend on domain interfaces, never on concrete implementations
- Constructors return interfaces, not concrete types

## Backend (Go)

- Use typed response structs; never use `map[string]interface{}`
- Use integer cents (`int64`) for all currency values — no floats
- Handle errors explicitly; wrap with context using `fmt.Errorf("operation: %w", err)`
- Keep functions small and focused on a single responsibility
- Use table-driven tests where applicable
- Mock dependencies via domain interfaces (see `internal/mock/`)

## Frontend (TypeScript / React)

- Use TypeScript strict mode — no `any` types unless absolutely unavoidable
- Define explicit interfaces/types for all API responses and inputs
- Keep components focused — extract reusable logic into custom hooks or utility functions
- Use TanStack Query for server state management
- All user-facing text in **pt-BR** (Brazilian Portuguese)
- Format currency as BRL using `Intl.NumberFormat` or `.toLocaleString('pt-BR', ...)`

## Workflow

- After every implementation, **compile and run tests**:
  - Backend: `go build ./...` and `go test ./... -count=1`
  - Frontend: `npx tsc --noEmit`
- Only commit code **after all tests and type checks pass**
- Write concise, descriptive commit messages following conventional commits (`feat:`, `fix:`, `refactor:`, etc.)
- Include `Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>` trailer in commits

## Code Style

- Only add comments where behavior is non-obvious; don't comment trivial code
- Prefer small, incremental changes over large rewrites
- When adding a new feature, follow the existing patterns in the codebase (entities, DTOs, mocks, handler helpers, test structure)
