# AGENTS.md

## Purpose
- This guide helps coding agents work safely and consistently in this repo.
- It documents how to build, lint, and test, plus the local style conventions.

## Repo Map
- `api/` Go HTTP API (Chi router, Jet SQL builder, PostgreSQL)
- `web/` Next.js 16 app router frontend (Tailwind + shadcn UI)
- `Makefile` root convenience targets and `Procfile` for dev orchestration

## Key Entrypoints
- API server entry: `api/main.go`
- API routing: `api/internal/*/routes.go`
- API handlers: `api/internal/*/handlers.go`
- API business logic: `api/internal/*/service.go`
- API repositories: `api/internal/repositories/*`
- API middleware: `api/internal/middleware/*`
- Web root layout: `web/app/layout.tsx`
- Web routes: `web/app/**/page.tsx`
- Web server actions: `web/lib/actions.ts`
- Web API client: `web/lib/sdk.ts`
- Web auth guard: `web/lib/auth.ts`

## Build / Run / Lint / Test

### Root (repo)
- Dev all services: `make dev` (uses Overmind + `Procfile`)
- Dev API only: `make dev/api`
- Dev web only: `make dev/web`
- Build all: `make build`
- Run API: `make run`
- Clean artifacts: `make clean`
- Procfile commands: `api` runs `make dev` and `web` runs `pnpm dev`.

### API (Go) — run from `api/`
- Build binary: `make build`
- Run built binary: `make run`
- Dev with live reload: `make dev` (uses `air`)
- Local Postgres: `docker-compose up` (see `api/docker-compose.yml`).
- Generate Jet SQL code: `make jet` (requires `DATABASE_URL`)
- Migrations:
  - Generate: `make db-generate name=<migration name>` (requires `DATABASE_URL`)
  - Apply: `make db-migrate` (requires `DATABASE_URL`)
  - Status: `make db-status` (requires `DATABASE_URL`)
  - Validate: `make db-validate` (requires `DATABASE_URL`)
  - Hash: `make db-hash`
- OpenAPI: `make openapi`

### API tests (Go)
- Full test suite: `go test ./...`
- Single package tests: `go test ./internal/auth`
- Single test name: `go test ./internal/auth -run TestName`
- To target a specific test across all packages: `go test ./... -run TestName`.
- No test files are currently present (`*_test.go` not found), so these are future-ready.

### Web (Next.js) — run from `web/`
- Dev server: `pnpm dev` (port 3001 in `package.json`)
- Production build: `pnpm build`
- Start production server: `pnpm start`
- Lint: `pnpm lint`
- Package manager: `pnpm` is used (`pnpm-lock.yaml` present).

### Web lint/test notes
- ESLint is configured via `web/eslint.config.mjs` (Next core-web-vitals + TS).
- Single-file lint: `pnpm lint -- app/login/page.tsx`
- No test runner is configured in `web/package.json`.

## Code Style — Go API
- Formatting: always `gofmt` the code; keep standard Go formatting.
- Imports: group standard library, internal (`auth/internal/...`), and external deps.
- Project layout: routers mount handlers; handlers call services; services call repos.
- IDs: use `ulid.ULID` and helpers in `internal/utils/ulidutil` for conversions. IDs are externally prefixed like `user_(ULID)`, but internally raw ULID
- SQL: use Jet (`internal/jet/postgres/public/model` and `.../table`) instead of raw SQL.
- Repos: DB access lives in `internal/repositories/*`, using Jet and returning `apperror`.
- Repository methods should log failures and map to `apperror`.
- Services: maintain the contract for instantiating a service in `internal/*/service.go`.
- Handlers: contain the business logic. parse JSON via `httputil.ParseBody` and send errors via `httputil.HandleError`.
- JSON responses: use `httputil.JSONResponse` for consistent headers.
- Errors: prefer `apperror.New*` constructors for HTTP-safe errors.
- DB errors: map postgres errors via `apperror.FromPGError` when appropriate.
- Logging: repository failures log with context using `log.Printf`.
- Events: security/audit events use `events.Log` with contextual IDs.
- Context: pass `context.Context` through services for logging/event metadata.
- Auth concerns (tokens, hashes, cookie names) live in `internal/auth/*`.

## Code Style — Web (Next.js)
- App Router: server components by default; add `"use client"` only when needed.
- Server actions: defined in `web/lib/actions.ts` with `"use server"` at top.
- Auth gating: use `withAuth` from `web/lib/auth.ts` for protected pages/roles.
- API calls: use `web/lib/sdk.ts` (returns `{ ok, status, data, error }`).
- Avoid calling the API directly from components; go through the SDK helper.
- Error UX: server actions set flash messages and `redirect` on failure.
- Routing: prefer `redirect()` and `sanitizeRelativePath` for safe destinations.
- Callback URLs: validate via `isAllowedCallbackUrl` and build with `withQuery` helpers.
- Styling: Tailwind + shadcn; tokens live in `web/app/globals.css`.
- Class merging: use `cn()` from `web/lib/utils.ts` for conditional classes.
- Fonts: `Space Grotesk` and `IBM Plex Mono` via `app/layout.tsx`.
- Imports: use alias `@/` for internal modules (see `tsconfig.json`).
- Types: prefer `import type` for type-only imports in TS/TSX files.
- Components: use named exports (`export function Foo`) and `type` for props.
- File naming: kebab-case filenames are common (`user-row.tsx`, `reset-tokens.go`).
- Forms: use server actions with `FormData`; redirect instead of returning JSON.
- Cookies: read/write via `next/headers` `cookies()`; pass `cookieStore.toString()` to SDK.

## Architecture Notes
- API routing mounts feature routers in `api/main.go` with `r.Mount`.
- Middleware stack includes request ID, real IP, client info, logging, recovery, timeout, and CORS.
- Client info is captured in `internal/middleware/clientinfo.go` and read via `httputil.ClientInfoFromContext`.
- Auth middleware validates bearer tokens or cookies and stores claims in context.
- Access/refresh cookie names are `access_token` and `refresh_token`.
- Frontend config lives in `web/lib/config.ts` (API base URL, docs URL, cookie domain, allowed hosts).
- Callback host validation is enforced by `ALLOWED_CALLBACK_HOSTS` and helpers in `web/lib/callback.ts`.
- Flash messages are stored in a `flash` cookie via `web/lib/flash.ts` and rendered by `FlashToaster`.
- Use `web/lib/http.ts` helpers (`withQuery`) for building URLs with query params.
- `sdkRequest` uses JSON bodies and `cache: "no-store"` for API calls.
- Protected pages should gate access early and redirect with flash on failure.

## Formatting and Conventions
- Keep string quotes double-quoted in TS/JS/TSX files.
- Maintain the existing semicolon style in each file (some files omit them).
- Use `type` aliases for props/params (see `LoginPageProps`, `UpdateUserParams`).
- Use `Params`/`Response` naming for request/response DTOs in Go services.
- Keep Tailwind class lists readable; group related utilities by purpose.
- Use `className` strings for Tailwind; avoid inline styles unless required.
- When editing shadcn components, preserve the variant + `cva` pattern.

## Error Handling Patterns
- Go HTTP: return `apperror.HTTPError` from services and let handlers map to HTTP.
- Go HTTP: use `httputil.HandleError` to normalize error responses.
- Web: check `result.ok` and `result.data` before use; handle failures with flash+redirect.
- Avoid leaking internal error details to clients; use friendly messages.
- For auth failures, prefer generic messages like "Invalid credentials".

## Generated / Third-Party Files
- `api/internal/jet/**` is generated by Jet; regenerate via `make jet`.
- `api/docs/openapi.json` is generated by `make openapi` (do not hand-edit).
- `web/components/ui/**` follows shadcn UI conventions (use shadcn if re-generating).
- `web/.next/**` and `api/bin/**` are build outputs.

## Environment Notes
- Web env values are documented in `web/.env.example`.
- API env values are documented in `api/.env.example`.
- API tasks like migrations require `DATABASE_URL` (see `api/Makefile`).
- API server config is loaded from env in `api/main.go`.
