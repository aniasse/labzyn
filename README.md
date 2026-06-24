# Labzyn

Copilote agentique open-source pour la recherche scientifique sur HPC.
NL → jobs Slurm (TAOUEY) → résultats reproductibles, avec un modèle de données
**métadonnées-driven** (objets & champs dynamiques) inspiré de
[Twenty](https://twenty.com) — réimplémenté en propre, **backend Go haute performance**.

> Reconstruction from-scratch (20 juin 2026) remplaçant le fork Twenty.
> Étude d'architecture : [`docs/architecture-study.md`](docs/architecture-study.md).

## Monorepo

```
labzyn/
  packages/
    server/      # backend Go (API, moteur de métadonnées, MCP, agents)
    front/       # React + Vite + shadcn (design Twenty-like)   [à venir]
  gateway/       # passerelle HPC Go :8088 → TAOUEY/Slurm        [réutilisé]
  docs/
```

## Architecture (philosophie Twenty, en Go)

- **Catalogue de métadonnées** : `metadata.object` + `metadata.field` + `metadata.view`
  décrivent des objets et champs définis à l'exécution.
- **Stockage MVP** : `core.record(data jsonb)` indexé GIN → schéma dynamique sans DDL.
  **v2** : DDL réel par workspace (`workspace_<id>`, vraies colonnes) pour la perf max,
  comme Twenty.
- **API REST générique** pilotée par le catalogue (`/api/v1/{object}`).
- **MCP** (JSON-RPC/SSE) → outils CRUD auto + `submit_slurm_job` → gateway `:8088`.

## Lancer le serveur

```bash
cd packages/server
cp ../../.env.example .env       # ajuster si besoin
go run ./cmd/server
# -> http://localhost:4000/healthz
```

Postgres requis (réutilise l'instance locale, base `labzyn`).

## Stack

Go 1.21+ · pgx v5 · chi · Postgres 16 · (front : React/Vite/shadcn).
