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
vibe-gen init --lang javascript --framework react --name my-app --dry-run
```

## Supported Frameworks

| Language   | Framework      | Type     |
|------------|----------------|----------|
| Go         | Echo           | API      |
| Go         | Gin            | API      |
| Go         | Fiber          | API      |
| Go         | Cron Job       | Worker   |
| JavaScript | ExpressJS      | API      |
| JavaScript | NestJS         | API      |
| JavaScript | Cron Scheduler | Worker   |
| JavaScript | React/Vite     | Frontend |
| JavaScript | Next.js        | Frontend |
| JavaScript | Vue.js/Vite    | Frontend |
| JavaScript | Svelte/Vite    | Frontend |
| JavaScript | NuxtJS         | Frontend |

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
make run-tui  # run interactive TUI without installing
```
