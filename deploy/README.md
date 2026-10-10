# Deployment Configurations

This directory contains cloud deployment manifests and infrastructure-as-code definitions.

## Render Blueprint (`deploy/render.yaml`)

[Render Blueprints](https://render.com/docs/blueprint-spec) provide declarative Infrastructure as Code:
- **Managed Database (`items-db`)**: Dedicated PostgreSQL instance (`items_db`).
- **Web Service (`items-api`)**: Minimal container built from `./Dockerfile`, listening on port `8080` with `/healthz` health check probe.
- **Automated Database Migrations**: `preDeployCommand: /bin/goose -dir /database/migrations postgres "$DATABASE_URL" up` runs the Goose migration binary prior to live traffic cutover.
- **Dynamic Configuration**: `DATABASE_URL` is securely wired directly from `items-db.connectionString`.

> **Render Setup**: When connecting your GitHub repository in the Render Dashboard, specify `deploy/render.yaml` as the Blueprint Path.

---

## Continuous Deployment (`.github/workflows/cd.yml.disabled`)

Continuous deployment workflow template triggering Render Deploy Hooks via secret `RENDER_DEPLOY_HOOK_URL`.

To enable:
1. Configure `RENDER_DEPLOY_HOOK_URL` in repository **Settings -> Secrets and variables -> Actions**.
2. Rename `.github/workflows/cd.yml.disabled` to `.github/workflows/cd.yml`.
