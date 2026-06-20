# Labzyn — Étude de Twenty & Architecture cible (Go, haute perf)

> Objectif : reconstruire Labzyn **from scratch** en Go haute-performance, en **réutilisant le design et la philosophie de Twenty** (réimplémentés, pas copiés → pas de contrainte AGPL). Document d'étude issu de l'analyse du fork local (`deepsearch/platform`).

---

## 1. Design system Twenty (tokens extraits — prêts à porter)

Source : `packages/twenty-ui/src/theme/constants/`.

### Typographie
- **Police** : `Inter, sans-serif`
- **Tailles** (rem) : `xxs 0.625` · `xs 0.85` · `sm 0.92` · `md 1` · `lg 1.23` · `xl 1.54` · `xxl 1.85`
- **Graisses** : regular 400 · medium 500 · semiBold 600

### Espacement
- **Unité de base = 4px**. `spacing(n) = n*4px`. Gap entre frères : `2px`.

### Border radius
- `xs 2px` · `sm 4px` · `md 8px` · `xl 20px` · `xxl 40px` · `pill 999px` · `rounded 100%`

### Couleurs
- **Grayscale** en `color(display-p3 …)` (gamut large P3) : gray1 = blanc → gray12 = `0.2` (texte principal). Texte secondaire gray11, tertiaire gray9.
- **Accents** = palettes **Radix UI** en P3, step `9` (red, blue=`indigo9`, green, purple, amber, etc.). → on installe `@radix-ui/colors`.
- Thèmes Light + Dark, exposés en **CSS variables** (`themeCssVariables.ts`).

### Divers UI
- Animations : instant 0.075s · fast 0.15s · normal 0.3s · slow 1.5s
- Table : padding cellule 8px, colonne checkbox 32px. Panneau latéral : 500px.

> **Porter vers Labzyn** : config Tailwind + CSS variables. shadcn/ui pour les primitives (mêmes radius/typo). Garder Inter, l'unité 4px, les radius, le P3, la palette Radix → rendu visuellement identique sans copier de code.

---

## 2. Architecture Twenty (vue d'ensemble)

Monorepo (nx) ~14k fichiers TS/TSX utiles. Packages clés :
- **twenty-front** (7853 f.) — React + Recoil, UI métadonnées-driven.
- **twenty-server** (6203 f.) — NestJS, moteur de métadonnées, API GraphQL générée, MCP.
- **twenty-ui** (533 f.) — design system (ce qu'on réutilise visuellement).
- **twenty-shared** (759 f.) — types partagés (dont `FieldMetadataType`).

### Le cœur : moteur de métadonnées
`packages/twenty-server/src/engine/metadata-modules/` — entités catalogue :
- `object-metadata` — définition d'un objet (table logique).
- `field-metadata` — définition d'un champ (+ type, options).
- `view`, `view-field`, `view-filter`, `view-sort`, `view-group` — vues table/kanban.
- `page-layout*`, `navigation-menu-item`, `role`/`permission*` — UI & droits.
- familles `flat-*` — représentation **en mémoire** des métadonnées (cache rapide).

### Modèle de stockage des DONNÉES — **point décisif**
Twenty fait du **DDL dynamique réel**, PAS du JSONB :
- `engine/twenty-orm/workspace-schema-manager/` → `getWorkspaceSchemaName` = **un schéma Postgres par workspace** (`workspace_<id>`).
- `build-sql-column-definition.util.ts` + `fieldMetadataTypeToColumnType` → quand on crée/édite un champ, ça génère un vrai `ALTER TABLE … ADD COLUMN` typé.
- Avantage : requêtes natives, index réels, contraintes SQL → **haute performance**. Coût : un **moteur de migration DDL** à écrire (le morceau le plus dur).

---

## 3. Système de types de champs (`FieldMetadataType`, 24 types)

Source : `packages/twenty-shared/src/types/FieldMetadataType.ts`.

**Scalaires** : `TEXT`, `NUMBER`, `NUMERIC`, `BOOLEAN`, `DATE`, `DATE_TIME`, `UUID`, `RATING`, `POSITION`, `RAW_JSON`, `RICH_TEXT`, `ARRAY`, `TS_VECTOR`.
**Énumérés** : `SELECT`, `MULTI_SELECT`.
**Composites** (plusieurs sous-colonnes) : `ACTOR`, `ADDRESS`, `CURRENCY`, `EMAILS`, `FULL_NAME`, `LINKS`, `PHONES`.
**Fichiers** : `FILES`.
**Relations** : `RELATION`, `MORPH_RELATION`.

> **Pour Labzyn** : un sous-ensemble suffit au départ → TEXT, NUMBER, NUMERIC, BOOLEAN, DATE/DATE_TIME, SELECT, MULTI_SELECT, RELATION, RAW_JSON, FILES, RICH_TEXT, RATING. Les composites (ADDRESS, CURRENCY…) viennent plus tard. Mapping type→colonne Postgres à reproduire.

---

## 4. MCP (Model Context Protocol) — `engine/api/mcp/`

Twenty expose un **serveur MCP** (`MCP_SERVER_INFO = "Twenty MCP Server" v0.1.0`) :
- **Transport** : JSON-RPC 2.0 sur **SSE** (`write-sse-event.util.ts`, `wrap-jsonrpc-response.util.ts`).
- **Auth** : `mcp-auth.guard.ts` + middleware de garde de méthode.
- **Services** : `mcp-protocol.service` (handshake/protocol), `mcp-tool-executor.service` (exécution).
- **Outils** : `list-object-metadata-names`, `list-skills` + outils **CRUD générés dynamiquement par objet** (annotations read-only / execute). Instructions serveur construites dynamiquement (`build-mcp-server-instructions.util.ts`).

> **Pour Labzyn** : un agent (Labzyn/Scholar/Methodologist/Analyst) parle au backend via MCP → outils auto `list_objects`, `query <object>`, `create/update <object>`, + outil métier `submit_slurm_job` qui appelle le **gateway Go :8088 → TAOUEY**. En Go : JSON-RPC + SSE est simple à servir.

---

## 5. Architecture cible Labzyn (Go, haute perf)

### Décision de stockage (la plus structurante)
Tu veux **dynamique + haute perf**. Deux voies :

| | A. JSONB + catalogue | B. DDL dynamique réel (voie Twenty) |
|---|---|---|
| Effort | Faible | Élevé (moteur migration) |
| Perf | Bonne (GIN) | Maximale (colonnes/index natifs) |
| Contraintes SQL | App-level | Natives |
| Reco | **MVP** | **Cible v2** |

→ **Plan : démarrer en A (JSONB), avec une couche d'abstraction qui permet de basculer vers B objet par objet** quand un objet devient « chaud ». On ne se ferme aucune porte.

### Stack
- **Backend** : Go (net/http 1.22 routing), **pgx** (driver) + **sqlc** pour le catalogue typé, **squirrel** pour les requêtes dynamiques. Postgres 16.
- **Auth** : JWT + workspaces + rôles (maison, léger).
- **Front** : React + Vite + **shadcn/ui** + tokens Twenty-portés. State : Zustand/TanStack Query. i18n FR/EN **propre dès le départ** (lacune actuelle du fork).
- **Agents** : orchestrateur Go → MCP (JSON-RPC/SSE) → outils → gateway `:8088` → TAOUEY (Slurm). Réutilise l'existant.

### Schéma catalogue (tables fixes)
```
workspaces(id, name, ...)
users(id, email, password_hash, ...) / memberships(user_id, workspace_id, role)
objects(id, workspace_id, name_singular, name_plural, label_fr, label_en, icon, is_system)
fields(id, object_id, name, type, label_fr, label_en, options jsonb, is_required, is_unique, relation_target, ...)
views(id, object_id, kind, filters jsonb, sorts jsonb)
records(id, object_id, workspace_id, data jsonb, position, created_at, updated_at)  -- voie A
```

### Modules backend
- `metadata` : CRUD objets/champs/vues (= « Settings > Data model »).
- `records` : CRUD générique `/api/{object}` + filtres/tri/pagination/recherche, validation depuis le catalogue.
- `auth`, `workspaces`, `roles`.
- `mcp` : serveur JSON-RPC/SSE, outils générés depuis le catalogue.
- `agents` : orchestration LLM → MCP → `slurm.submit` → gateway.

### Modules front (métadonnées-driven, comme Twenty)
- `RecordTable` / `RecordBoard` / `RecordDetail` génériques (lisent les métadonnées).
- `DataModelSettings` (éditeur objets/champs).
- `CommandMenu` (Ctrl+K), nav latérale.
- `AgentChat`.

---

## 6. Jalons
1. **Squelette Go** : repo, pgx, migrations catalogue, `/healthz`.
2. **Metadata + Records API** (JSONB) : créer objet « Project », champs, CRUD records.
3. **Front design system** : Tailwind+tokens Twenty, layout, Ctrl+K.
4. **Table/Fiche génériques** branchées sur l'API → 1er objet de bout en bout.
5. **Auth + workspaces + rôles**.
6. **MCP + agent** → gateway `:8088` → TAOUEY.
7. (v2) **Moteur DDL dynamique** pour objets chauds.

---

## 7. Ce qu'on NE reprend PAS de Twenty
CRM (companies/people/opportunities), email/calendar sync, messaging, workflows lourds, Zapier, le monorepo nx géant, NestJS. → Labzyn = lean, focalisé recherche scientifique + HPC.
