# GSoC 2026 — Apache Answer Contribution Log

> **Contributor:** Yash Chauhan (@Yash-Chauhan22)
> **Upstream Project:** [apache/answer](https://github.com/apache/answer)
> **Fork:** [Yash-Chauhan22/answer](https://github.com/Yash-Chauhan22/answer)
> **Program:** Google Summer of Code 2026

---

## Repository Setup

| Task | Status |
|------|--------|
| Fork apache/answer | Done |
| Clone fork locally | Done |
| Add upstream remote | Done |
| Install Go toolchain | In progress |
| Install Node.js + pnpm | In progress |
| Create .env config | Done |
| Create gsoc-dev branch | Pending |

---

## Git Remotes

`
origin   -> https://github.com/Yash-Chauhan22/answer.git  (your fork)
upstream -> https://github.com/apache/answer.git           (main project)
`

### Syncing with upstream

`bash
git fetch upstream
git checkout main
git merge upstream/main
git push origin main
`

---

## Branch Strategy

| Branch | Purpose |
|--------|---------|
| main | Mirrors upstream apache/answer main |
| gsoc-dev | Active GSoC development base |
| feature/xxx | Individual feature branches |

---

## Running the Project

### Option 1: Docker (Quickest)

`bash
docker-compose up -d
# Access at http://localhost:9080
`

### Option 2: Full Local Dev

Backend:
`bash
go mod download
go run ./cmd/answer/... run -C ./configs/
`

Frontend (new terminal):
`bash
cd ui
pnpm install
pnpm dev
`

---

## Important Links

| Resource | URL |
|----------|-----|
| Project Website | https://answer.apache.org |
| Upstream GitHub | https://github.com/apache/answer |
| Contributing Guide | https://answer.apache.org/community/contributing |
| Discord | https://discord.gg/a6PZZbfnFx |
| Dev Mailing List | dev@answer.apache.org |

---

## Contribution Notes

- PRs go to apache/answer (upstream), not your fork
- Each PR needs an associated GitHub issue
- Conventional commits: feat:, fix:, docs:, refactor:, test:
- Apache License header required on new files (CI checks this)
