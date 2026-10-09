# vibe-gen

> AI-ready project boilerplate generator. Generates fully structured, documented projects with built-in AI agent context files.

## Installation

```bash
# From source
cd vibe-gen
make install
```

This installs `vibe-gen` to your `$GOPATH/bin`. Make sure `$GOPATH/bin` is in your `$PATH`.

## Usage

### Interactive TUI (recommended)

```bash
vibe-gen
```

Navigate with arrow keys, select language → framework → project name → destination → confirm.

### Non-interactive (CI / scripts)

```bash
vibe-gen init \
  --lang go \
  --framework echo \
  --name my-service \
  --module github.com/myorg/my-service \
  --dest ./my-service

# Preview without writing files
vibe-gen init --lang go --framework gin --name my-app --dry-run

# Frameworks with variants (Next.js: app | pages, default app)
vibe-gen init --lang typescript --framework next --variant pages --name my-web
```

Languages: `go`, `javascript`, `typescript`. TypeScript templates are the source of truth; the JavaScript
ones are derived from them with `make gen-js` (NestJS is TypeScript-only).

## Supported Frameworks

| Language                 | Framework                         | Type     | Auth in the browser                  |
|--------------------------|-----------------------------------|----------|--------------------------------------|
| Go                       | Echo, Gin, Fiber                  | API      | —                                    |
| Go                       | Cron Job                          | Worker   | —                                    |
| JavaScript / TypeScript  | Express 5, Hono (Node)            | API      | —                                    |
| TypeScript               | NestJS 12                         | API      | —                                    |
| JavaScript / TypeScript  | Cron Scheduler                    | Worker   | —                                    |
| JavaScript / TypeScript  | Next.js 16 (App / Pages Router)   | Frontend | BFF, sealed `__Host-` cookies        |
| JavaScript / TypeScript  | SvelteKit 3                       | Frontend | BFF, sealed `__Host-` cookies        |
| JavaScript / TypeScript  | Nuxt 4                            | Frontend | BFF, sealed `__Host-` cookies        |
| JavaScript / TypeScript  | React 19 / Vue 3 (Vite SPA)       | Frontend | API `AUTH_COOKIE_MODE`, token in memory |

Express and Hono share one Node API core (`templates/typescript/_shared/api`): the same auth, repositories,
migrations and JSON contract as the Go templates.

Only templates marked `Ready` in `internal/registry/registry.go` are offered: they are hidden from the TUI
and rejected by `init` until they pass `make smoke`.

## What Gets Generated

Every project includes:

### Source Code
- Framework-specific skeleton with proper layering (handler → service → repository)
- Sample domain with full CRUD pattern — adapt and copy for your own domains
- Auth domain with JWT (RS256) for API projects
- Config loading (YAML + env var priority, secrets from env only)
- Database package (GORM/Sequelize) with connection pooling
- Middleware: JWT auth, RBAC, rate limiting, CORS

### Developer Files
| File | Purpose |
|---|---|
| `CLAUDE.md` | AI agent entry point — rules, doc map, workflow |
| `.env.example` | All env vars with empty placeholders |
| `config.example.yml` | YAML config template (non-secret values) |
| `Makefile` | `make run`, `make build`, local DB/Redis containers |
| `Dockerfile` | Multi-stage production build |

### AI Documentation (`docs/`)
| Doc | Purpose |
|---|---|
| `architecture.md` | System design and layer responsibilities |
| `coding-standards.md` | Language-specific standards |
| `security.md` | Secret handling, auth, input validation |
| `testing.md` | Strategy, tooling, naming |
| `api.md` | Response format, status codes, pagination |
| `database.md` | Naming, migrations, indexing |
| `observability.md` | Logging, tracing, metrics |
| `performance.md` | DB, caching, concurrency |
| `decisions.md` | Architecture Decision Records (ADR) |
| `progress.md` | AI agent context handoff file |
| `roadmap.md` | Project phases |
| `vibe-workflow.md` | Step-by-step AI dev workflow |
| `git.md` | Commit format, safety rules |

Frontend projects also get `DESIGN.md` at the root (developer-authored direction, gated by its `Status`),
`docs/design_system.md` (engineering rules: tokens, light/dark theme, states, forms, motion) and
`.semgrep/antislop.yml` (the slop lint that `make sast` runs). See the root `CLAUDE.md`, section 5.

## Next Steps After Generation

After running `vibe-gen`, follow the printed instructions:

```bash
cd <project-slug>

# For Go projects:
cp .env.example .env          # fill in your local values
cp config.example.yml config.yml
go mod tidy
make run

# For JavaScript projects:
cp .env.example .env
npm install
npm run dev
```

## Development

```bash
cd vibe-gen
make build    # build to ./out/vibe-gen
make test     # run tests
make smoke    # generate every ready template and build/vet/fmt/test/govulncheck/semgrep it (needs semgrep on PATH);
              # also proves every slop-lint rule fires on internal/smoke/testdata/antislop
make smoke FRAMEWORK=gin   # one template only
make gen-js         # derive JavaScript templates from TypeScript ones (Node 24)
make gen-js-check   # fail if the committed JavaScript templates are stale
make run-tui  # run interactive TUI without installing
```

### When `make smoke` fails on govulncheck

The Go smoke runs `govulncheck@latest`, so it turns red whenever a new advisory is published: the alarm doing its job,
not a flaky test. Read the report, then:

- **A module** (`Found in: golang.org/x/net@v0.59.0`): raise that pin in `templates/go/{echo,gin,fiber}/go.mod.tmpl`
  (cron-job pins no x/net).
- **The standard library** (`Found in: net/http@go1.27`): raise the `go` line in all four
  `templates/go/*/go.mod.tmpl` to the `Fixed in` release (for example `go 1.27.2`). With `GOTOOLCHAIN=auto` (the
  default) Go 1.21+ downloads that toolchain on demand, so the smoke does not depend on your local patch version. With
  `GOTOOLCHAIN=local`, which the official `golang` Docker images set, an older toolchain fails with
  `go.mod requires go >= 1.27.2`: that is the point, but it means the Dockerfile's builder image must already have that
  release (`docker run --rm golang:1.27-alpine go version`). Do not use the `toolchain` line for this: `GOTOOLCHAIN=local`
  ignores it without a warning.

Edit those lines by hand: `go mod tidy` in a generated project also rewrites unrelated lines (it marks `swag` as
indirect). Then run `make smoke FRAMEWORK=<echo|gin|fiber|cron-job>` for each Go template.

### Writing TypeScript templates

`templates/typescript/<fw>/` is the only source; `make gen-js` writes `templates/javascript/<fw>/`
(types stripped with sucrase, then prettier). Commit both. For the conversion to stay lossless:

- Type-only imports use `import type`; no `enum`, `namespace`, parameter properties or decorators (erasable syntax only).
- Relative imports may end in `.ts`/`.tsx` (rewritten to `.js`/`.jsx`) or be extensionless; never mix styles in one template.
- Vue: declare props at runtime (`defineProps({ ... })`), not with a type argument (stripping it would drop the props); no TypeScript inside `<template>`, and no `{{ }}` mustache (it collides with Go templates): use `v-text` / `v-bind`.
- Go template tags in `.ts`/`.tsx` and `<script lang="ts">` code are value tags only (`{{.ProjectSlug}}`).
  Branch on the language with `{{if .IsTypeScript}}` only in files gen-js copies verbatim (Makefile, docs, `.js` configs).
- `tsconfig*.json` and `*.d.ts` are dropped; TypeScript-only dev dependencies and the `typecheck` script are removed from `package.json`.
- NestJS is TypeScript-only and has no JavaScript output.
- TypeScript version: 7.x (native) everywhere except Vue, Nuxt and SvelteKit, which pin 6.x: `vue-tsc` and
  `svelte-check` still need the TypeScript JS API that 7.x no longer ships.
