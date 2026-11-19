# Repository Guidelines

## Project Structure & Module Organization

Apache Answer combines a Go backend with a React/TypeScript frontend. `cmd/answer` is the executable entry point; `internal/` contains controllers, services, repositories, schemas, and migrations. Shared Go utilities live in `pkg/`, and extension interfaces live in `plugin/`. Frontend pages, components, stores, and API clients are under `ui/src/`; static assets are in `ui/public/`, and production output is embedded from `ui/build/`. Translations live in `i18n/`, configuration examples in `configs/`, and deployment resources in `charts/`. Go tests generally sit beside their implementation; repository integration tests live in `internal/repo/repo_test/`.

## Build, Test, and Development Commands

Use Go 1.25+, Node.js matching `ui/package.json` (`^20.19.0 || >=22.12.0`), and pnpm 9.7.0. `make lint` also requires Docker or Podman installed and running for the ASF header formatter.

- `make generate`: install generators, regenerate Go code/API documentation, and tidy modules.
- `make ui`: install frontend dependencies and build the UI.
- `make build`: regenerate code and build the `answer` executable.
- `./answer init -C ./data` then `./answer run -C ./data`: initialize and run locally.
- Before `cd ui && pnpm start` (Vite, port 3000), set `REACT_APP_API_URL=http://localhost:8080/` in `ui/.env.development.local`, matching your backend address. This overrides the checked-in private proxy target; restart Vite after changes.
- `make test`: run repository integration tests.
- `go test ./...`: run the full Go test suite.
- `make lint`: regenerate code, check ASF headers, and run golangci-lint.
- `cd ui && pnpm lint`: run ESLint with automatic fixes; `pnpm prettier` formats frontend sources.

## Coding Style & Naming Conventions

Format Go with `gofmt`; use tabs and conventional exported `PascalCase` names. Follow existing snake_case Go filenames and package organization. Frontend code uses two spaces, single quotes, trailing commas, and an 80-column Prettier target. Use PascalCase component directories and follow neighboring files. Respect `.editorconfig`, ESLint, and ASF license headers. Regenerate generated files through tooling rather than editing them manually.

## Testing Guidelines

Use Go's `testing` package and existing Testify assertions. Name files `*_test.go` and functions `TestXxx`. Add regression coverage for changed behavior. Repository tests default to SQLite; MySQL/PostgreSQL variants use `TEST_DB_DRIVER` and Docker. No fixed coverage threshold or frontend test script is configured; verify UI changes with `pnpm build` and browser checks.

## Commit & Pull Request Guidelines

Follow history's conventional prefixes: `fix:`, `feat:`, `build:`, `test:`, or `chore:`, with optional scope, such as `fix(security): enforce visibility`. Keep commits focused. Follow the PR template: link the issue with `Fixes #...` and describe proposed changes. Include validation results and screenshots for visible UI changes. Review generated changes before committing; CI checks linting and unexpected generated diffs.
