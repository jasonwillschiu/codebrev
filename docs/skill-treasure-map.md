# Skill: Treasure Map (Blast Radius + No Forgotten Deps)

Use this skill when you’re about to make a change and want a small, high-signal “treasure map” that prevents missed dependencies and reduces breakage.

## Prereqs

This skill is intended to be usable by anyone on the team and assumes:

- `codebrev` is installed and can generate `codebrev.md`
- `codebrev.md` already exists for the target repo (refresh it if it might be stale)
- `rg` (ripgrep) is installed
- `ast-grep` is installed

## Inputs

- Repo root: `REPO_ROOT`
- Change request (1–3 sentences): `CHANGE_REQUEST`
- Seeds (optional): `SEED_FILES` (1–5 file paths you expect to touch)
- Hard size caps: `NODE_BUDGET=20`, `EDGE_BUDGET=30`

## Step 0 — Ensure `codebrev.md` is current

This skill assumes `codebrev.md` already exists. If it might be stale, regenerate it with `codebrev` for `REPO_ROOT` before you start mapping.

## Step 1 — LLM prompt (Treasure Map builder)

Copy/paste the prompt below into the LLM that will build the map. It should treat `codebrev.md` as the primary source of truth for discovery and only open raw files when needed to resolve ambiguity.

---

### Prompt: Build a Treasure Map for the change

You are building a **task-scoped treasure map** to implement the change safely with **minimal forgotten dependencies**.

Goal: Produce an actionable map that answers:
1) What files/contracts will likely change?
2) What *must* be checked/updated to avoid breakage?
3) What is the smallest verification checklist to prove the change is correct?

Constraints:
- **Hard cap**: max `{NODE_BUDGET}` nodes and `{EDGE_BUDGET}` edges in any Mermaid graph.
- Prefer **contracts/shape edges** over “imports everything” edges.
- Collapse broad modules into boundary nodes; don’t expand high fan-in infra unless directly edited.
- If uncertain, resolve by inspecting source files (don’t guess).

Available tools:
- `codebrev.md` (repo index)
- `rg`/search
- `ast-grep` (validate edges; don’t use it as the primary graph generator)
- Read specific files to remove ambiguity

Repository:
- Root: `{REPO_ROOT}`
- Change request: `{CHANGE_REQUEST}`
- Seed files (if provided): `{SEED_FILES}`

Process (do not skip):
1) Read `codebrev.md` and identify:
   - likely touched packages/files
   - relevant **Contracts**: tagged structs, routes, event subjects/messages, DTOs
   - “Change Impact Analysis” entries for the touched surfaces
2) Propose a candidate dependency set, then **validate key edges**:
   - Use `ast-grep` (preferred) or targeted file reads to confirm:
     - call sites (handlers → services → repos)
     - DTO construction/consumption
     - event payload production/consumption (NATS subjects, message structs)
     - schema/serialization field names (json tags)
3) Prune deterministically until within budget:
   - Hop limit: 2 from seeds/contracts
   - Drop low-risk/trivial edges (logging/metrics, generic helpers)
   - Collapse directories into module nodes unless directly edited
   - Stop expansion at boundary nodes (router bootstrap, infra clients, generic middleware)

Deliverables (exact format):

1) **Treasure Map (Mermaid)**: one graph only, within budget.
   - Nodes are files or collapsed modules.
   - Edges must be typed: `calls`, `api_shape`, `dto`, `event_payload`, `schema`, `cache_key`, `template_props`.
   - Mark boundary nodes explicitly in labels (e.g., `Boundary: core/router`).

2) **Contracts to Watch** (bulleted list)
   - Each contract includes: name, producer(s), consumer(s), why it matters.
   - Examples: request/response DTO, message payload struct, route path, subject name, JSON field set.

3) **Edit Checklist** (5–12 bullets)
   - Concrete “touch/verify” items with file paths and what to verify.

4) **Verification Plan**
   - Minimal commands/tests to run (repo-specific).
   - If none exist, include “smoke checks” (build, run, or targeted compilation).

5) **Open Questions / Ambiguities**
   - List any missing info that blocks safe execution and which file(s) to open to resolve it.

Do not write implementation code. This output is the map for the implementer.

---

## Step 2 — Edge validation patterns (ast-grep first)

Use these as *validation* queries after `codebrev.md` suggests candidates.

### Go: call sites / wiring

- Find call sites of a function/type:
  - `rg -n "NewThing\\(" -S .`
  - `ast-grep` to confirm it’s a real call, not a comment/string.

### Go: JSON/DTO field usage

- Confirm a field is read/written (especially renamed fields):
  - `rg -n "\\.FieldName\\b" -S .`
  - `rg -n "json:\"field_name\"" -S .`

### NATS / subjects / message payloads

- Confirm a subject string or message type is used across boundaries:
  - `rg -n "subjects\\.|Subject" -S .`
  - `rg -n "Publish\\(|Subscribe\\(" -S .`

## Step 3 — Ambiguity clearing (read files deliberately)

When you hit uncertainty, open only the minimal files that resolve it:

- Entry points: router/bootstrap, handler registration, worker registry.
- Contract definitions: `contracts/*` message/DTO files.
- Producers/consumers: the service that emits and the handler/worker that consumes.
- Any “generated” file only if it defines an API surface you must update; otherwise treat as boundary.

## Output target

- Save the resulting map as `docs/treasure-map-[good-name-for-this-edit].md` alongside the repo root.
  - Keep it small and actionable; if it grows, prune harder rather than adding pages.
