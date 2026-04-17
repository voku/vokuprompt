# vokuprompt

`vokuprompt` turns a weak task request into a compiled execution prompt for a coding agent.

## What vokuprompt is

- A small compiler from category + pattern definitions into an execution-ready prompt.
- A way to make task framing reviewable through `patterns.json`, `categories.json`, and `placeholders.json`.
- A tool for both agents and humans to inspect how prompt shape changes by category.

## What vokuprompt is not

- Not an embedded LLM.
- Not an auto-detector of task intent.
- Not a general workflow runner or task tracker.
- Not decorative docs: the JSON registries and skill file are the product boundary.

## Installation / Build

```bash
git clone https://github.com/voku/vokuprompt.git
cd vokuprompt
go test ./... -v
go build ./cmd/vokuprompt
```

Run directly without installing:

```bash
go run ./cmd/vokuprompt categories
go run ./cmd/vokuprompt optimize --category bugfix
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
  "execution_request": "Analyze the original prompt, improve it, and execute the improved prompt now.\n\nReturn:\n1. failure analysis\n2. improved prompt\n3. execution result"
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

## How an agent should use it

1. Run `vokuprompt categories`.
2. Choose the right category.
3. Run `vokuprompt optimize --category <name>`.
4. Read `compiled_prompt` and `placeholder_manifest`.
5. Resolve placeholders using repository facts plus `placeholders.json`.
6. Append `execution_request`.
7. Execute the filled prompt.

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
