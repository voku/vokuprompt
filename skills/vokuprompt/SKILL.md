# vokuprompt

Use `vokuprompt` to turn a weak task request into a compiled execution prompt for a coding agent.

## Workflow

1. Run `vokuprompt categories` to see available categories.
2. Choose the best category for the task.
3. Run `vokuprompt optimize --category bugfix` (or another listed category).
4. Inspect the returned `compiled_prompt` and `placeholder_manifest`.
5. Resolve each placeholder with current task context plus `placeholders.json` from the repository root.
6. Append the returned `execution_request`.
7. Execute the filled prompt now.
8. Return:
   1. failure analysis
   2. improved prompt
   3. execution result

## Placeholder resolution rules

- Use the placeholder manifest from `vokuprompt optimize` as the source of truth for what must be filled now.
- Use `placeholders.json` to understand each placeholder's meaning, required/optional status, and expected resolution style.
- Prefer repository facts, the user request, and already-run validation commands over guesswork.
- If a placeholder cannot be resolved safely, stop and ask for the missing input instead of inventing it.

## Example 1

### Original weak prompt

`fix the failing bug and keep the change small`

### Chosen category

`bugfix`

### Command invocation

```bash
vokuprompt categories
vokuprompt optimize --category bugfix
```

### Returned compiled prompt

```text
Goal:
[TASK]

Context:
[CONTEXT_TARGET]

Constraints:
- Do not modify unrelated [SCOPE_ELEMENTS].

Execution:
- Work step by step over [UNIT] and continue until [DONE_CONDITION].

Validation:
- Run verification with [VALIDATION] and show raw output.

Review:
- Double-check the minimal patch and validate [STABLE_INTERFACE] remains unchanged.
```

### Placeholder resolution

- `[TASK]` → Fix the nil dereference in the optimizer with the smallest stable patch.
- `[CONTEXT_TARGET]` → `internal/cli/optimize.go` and its contract tests.
- `[SCOPE_ELEMENTS]` → unrelated commands, repo docs outside the task, and engine selection behavior.
- `[UNIT]` → one failing contract check, then the minimal code change, then rerun tests.
- `[DONE_CONDITION]` → `go test ./... -v` passes and the optimize JSON contract includes the new field.
- `[VALIDATION]` → `go test ./... -v`
- `[STABLE_INTERFACE]` → existing `categories` and `optimize` flags plus unchanged compiler selection behavior.

### Final executed prompt

```text
Goal:
Fix the nil dereference in the optimizer with the smallest stable patch.

Context:
internal/cli/optimize.go and its contract tests.

Constraints:
- Do not modify unrelated commands, repo docs outside the task, or engine selection behavior.

Execution:
- Work step by step over one failing contract check, then the minimal code change, then rerun tests and continue until go test ./... -v passes and the optimize JSON contract includes the new field.

Validation:
- Run verification with go test ./... -v and show raw output.

Review:
- Double-check the minimal patch and validate existing categories and optimize flags plus unchanged compiler selection behavior remains unchanged.

Analyze the original prompt, improve it, and execute the improved prompt now.

Return:
1. failure analysis
2. improved prompt
3. execution result
```

## Example 2

### Original weak prompt

`add agent docs so people know how to use this tool`

### Chosen category

`bugfix` (the current shipped category for minimal, reviewable repo changes)

### Command invocation

```bash
vokuprompt categories
vokuprompt optimize --category bugfix
```

### Returned compiled prompt

Use the same `compiled_prompt` and `placeholder_manifest` shape returned by `vokuprompt optimize`.

### Placeholder resolution

- `[TASK]` → Add agent-facing docs, a placeholder registry, and examples without redesigning the compiler.
- `[CONTEXT_TARGET]` → `skills/vokuprompt/SKILL.md`, `placeholders.json`, and optimize response docs/tests.
- `[SCOPE_ELEMENTS]` → core compiler logic, category selection behavior, and unrelated repository files.
- `[UNIT]` → one agent-facing file or contract test at a time.
- `[DONE_CONDITION]` → the skill doc matches the CLI, placeholders are documented, and all tests pass.
- `[VALIDATION]` → `go test ./... -v`
- `[STABLE_INTERFACE]` → the existing CLI behavior, selected pattern ordering, and compiled prompt format.

### Final executed prompt

```text
Goal:
Add agent-facing docs, a placeholder registry, and examples without redesigning the compiler.

Context:
skills/vokuprompt/SKILL.md, placeholders.json, and optimize response docs/tests.

Constraints:
- Do not modify core compiler logic, category selection behavior, or unrelated repository files.

Execution:
- Work step by step over one agent-facing file or contract test at a time and continue until the skill doc matches the CLI, placeholders are documented, and all tests pass.

Validation:
- Run verification with go test ./... -v and show raw output.

Review:
- Double-check the minimal patch and validate the existing CLI behavior, selected pattern ordering, and compiled prompt format remains unchanged.

Analyze the original prompt, improve it, and execute the improved prompt now.

Return:
1. failure analysis
2. improved prompt
3. execution result
```
