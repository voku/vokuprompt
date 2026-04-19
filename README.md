# vokuprompt

`vokuprompt` turns a weak task request into a compiled prompt contract for a coding agent.

## Quick Start

Get the latest published Linux bundle:

```bash
curl -fsSL https://voku.github.io/vokuprompt/vokuprompt-linux-amd64.tar.gz -o /tmp/vokuprompt-linux-amd64.tar.gz
curl -fsSL https://voku.github.io/vokuprompt/vokuprompt-linux-amd64.tar.gz.sha256 -o /tmp/vokuprompt-linux-amd64.tar.gz.sha256
(cd /tmp && sha256sum -c vokuprompt-linux-amd64.tar.gz.sha256)
mkdir -p /tmp/vokuprompt
tar -xzf /tmp/vokuprompt-linux-amd64.tar.gz -C /tmp/vokuprompt
/tmp/vokuprompt/vokuprompt-linux-amd64/vokuprompt categories
```

Published bundle links:

- Linux bundle: https://voku.github.io/vokuprompt/vokuprompt-linux-amd64.tar.gz
- SHA-256: https://voku.github.io/vokuprompt/vokuprompt-linux-amd64.tar.gz.sha256

## What vokuprompt is

- A deterministic compiler: the caller fetches categories, chooses one, and `vokuprompt` compiles the matching prompt contract.
- A way to make task framing reviewable through `patterns.json`, `categories.json`, and `placeholders.json`.
- A tool for both agents and humans to inspect how prompt shape changes by category.
- A deterministic way to force post-task memory capture contracts when meaningful new understanding must be written back to a repo-local memory store.

## What vokuprompt is not

- Not an embedded LLM.
- Not an auto-detector of task intent: the agent must fetch categories and choose one before running `optimize`.
- Not a general workflow runner or task tracker.
- Not the memory store itself: it only compiles the deterministic write-back contract.
- Not decorative docs: the JSON registries and skill file are the product boundary.

## Installation / Build

```bash
git clone https://github.com/voku/vokuprompt.git
cd vokuprompt
go test ./... -v
go build ./cmd/vokuprompt
```

The latest installable GitHub Actions bundle is published by the `Build Artifact` workflow on `main` and mirrored to GitHub Pages for a stable download URL.
It contains:

- the `vokuprompt` Linux binary
- `categories.json`
- `patterns.json`
- `placeholders.json`
- `skills/vokuprompt/SKILL.md`

Run directly without installing:

```bash
go run ./cmd/vokuprompt categories
go run ./cmd/vokuprompt optimize --category bugfix
```

## Copy-paste prompt for Copilot setup steps

If you want another coding-agent repository to preload `vokuprompt` from the published bundle, give the agent this prompt:

```text
Add a `.github/workflows/copilot-setup-steps.yml` workflow for GitHub Copilot cloud agent.

Requirements:
- Use a single `copilot-setup-steps` job on `ubuntu-latest`.
- Use least-privilege permissions, including `actions: read` and `contents: read`.
- Download `https://voku.github.io/vokuprompt/vokuprompt-linux-amd64.tar.gz`.
- Extract the tarball into `/usr/local/share/vokuprompt`.
- Keep the binary at `/usr/local/share/vokuprompt/vokuprompt`, make it executable, and symlink `/usr/local/bin/vokuprompt` to it.
- Keep `categories.json`, `patterns.json`, `placeholders.json`, and `skills/vokuprompt/SKILL.md` under `/usr/local/share/vokuprompt`.
- Verify the install with:
  `vokuprompt categories`
- Make the workflow runnable via `workflow_dispatch` and also validate changes when the workflow file is edited.

Return the full workflow file, ready to paste.
```

## categories example

```bash
go run ./cmd/vokuprompt categories
```

Example output:

```json
{
  "categories": [
    {
      "name": "bugfix",
      "description": "Fix a bug with minimal patch and explicit proof."
    },
    {
      "name": "performance",
      "description": "Improve speed with dominant-workload realism and benchmark proof."
    },
    {
      "name": "refactor",
      "description": "Restructure safely with containment, simplification, and interface stability."
    },
    {
      "name": "review",
      "description": "Critically analyze a change, challenge assumptions, and surface missing evidence."
    },
    {
      "name": "tests",
      "description": "Drive the task with a failing test first and use tests as proof."
    },
    {
      "name": "operational_contract",
      "description": "Execute as a multi-pass operational contract with missingness checks, verification loops, and explicit continuation."
    }
  ]
}
```

## optimize example

```bash
go run ./cmd/vokuprompt optimize --category bugfix
```

Example output shape:

```json
{
  "category": "bugfix",
  "selected_patterns": [
    "goal_context_constraints_done",
    "scope_containment",
    "step_loop",
    "verification_prompt",
    "double_check"
  ],
  "compiled_prompt": "Goal:\n[TASK]\n\nContext:\n[CONTEXT_TARGET]\n\nConstraints:\n- Do not modify unrelated [SCOPE_ELEMENTS].\n\nExecution:\n- Work step by step over [UNIT] and continue until [DONE_CONDITION].\n\nValidation:\n- Run [VALIDATION] and show raw output.\n\nReview:\n- Double-check the minimal patch and validate [STABLE_INTERFACE] remains unchanged.",
  "placeholder_manifest": [
    {
      "name": "CONTEXT_TARGET",
      "required": true,
      "node_types": ["Framing"],
      "source_patterns": ["goal_context_constraints_done"]
    }
  ],
  "execution_request": "Treat the selected category as fixed input chosen from the deterministic category registry.\n1. Use placeholder_manifest as the source of truth for which placeholders must be resolved now.\n2. Build the final executable prompt by resolving every required placeholder in compiled_prompt from repository facts and the current task context.\n3. Keep the selected category and compiled structure intact; do not silently rewrite the contract.\n4. If a required placeholder cannot be resolved safely, stop and ask for the missing input.\n5. After placeholder resolution, execute the final prompt.\n\nReturn:\n1. category confirmation\n2. placeholder resolution summary\n3. final executable prompt\n4. execution result"
}
```

## explain / inspect example

If you want trust and debuggability, inspect the selection trace:

```bash
go run ./cmd/vokuprompt optimize --category bugfix --explain
```

`--explain` keeps the normal optimize response and adds:

- patterns rejected because of conflicts
- patterns rejected because of role limits
- required placeholders in one quick list

This is useful for debugging, reviewability, and understanding why one category compiles differently from another.

## Memory-forcing use case

When non-trivial discovery, debugging, refactoring, or implementation work creates durable new understanding, run a memory-oriented category after the primary task so the agent compiles a deterministic write-back contract.

Example commands:

```bash
go run ./cmd/vokuprompt categories
go run ./cmd/vokuprompt optimize --category code_discovery
go run ./cmd/vokuprompt optimize --category debugging_digest
go run ./cmd/vokuprompt optimize --category architecture_memory
go run ./cmd/vokuprompt optimize --category handoff_memory
```

Use these categories to force evidence-backed, privacy-aware write-back into the repo-local memory store already chosen by the host repository. `vokuprompt` stays the generic compiler for the contract and does not define a new persistence model.

## How an agent should use it

1. Run `vokuprompt categories`.
2. Choose the right category from the deterministic category registry.
3. Run `vokuprompt optimize --category <name>`.
4. Read `compiled_prompt`, `placeholder_manifest`, and `execution_request`.
5. Resolve placeholders using repository facts plus `placeholders.json`.
6. Use `execution_request` as the meta prompt that tells the agent how to build the final executable prompt after placeholder resolution.
7. Execute the resolved prompt contract.
8. If meaningful new understanding emerged, run a memory-oriented category and complete the repo-local write-back before stopping.

For the agent-oriented flow, see `skills/vokuprompt/SKILL.md`.

## How a human can test it manually

```bash
go test ./... -v
go run ./cmd/vokuprompt categories
go run ./cmd/vokuprompt optimize --category performance
go run ./cmd/vokuprompt optimize --category review --explain
```

Then compare:

- `selected_patterns`
- `compiled_prompt`
- `placeholder_manifest`
- `explanation.rejected_by_conflict`
- `explanation.rejected_by_role_limits`

You should see category-specific differences in prompt tone and structure.

## Knowledge base / product boundary

These files are not supporting decoration; they are the product boundary:

- `patterns.json`
- `categories.json`
- `placeholders.json`
- `skills/vokuprompt/SKILL.md`

Together they define:

- what categories exist
- how prompts are composed
- which placeholders must be resolved
- how agents are expected to use the tool
