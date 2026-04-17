# vokuprompt

Use `vokuprompt` to turn a weak task request into a compiled prompt contract for a coding agent.

## Workflow

1. Run `vokuprompt categories` to see available categories.
2. Choose the best category for the task from the deterministic category registry.
   - `bugfix` for minimal patches with explicit verification.
   - `performance` for dominant-workload measurement and benchmark-backed speedups.
   - `refactor` for safe restructuring, containment, and deletion-before-extension.
   - `review` for failure analysis, missing evidence, and challenge-oriented critique.
   - `tests` for failing-test-first work where tests are the proof.
3. Run `vokuprompt optimize --category bugfix` (or another listed category).
4. Inspect the returned `compiled_prompt`, `placeholder_manifest`, and `execution_request` so you know both the prompt contract and the meta execution layer.
5. Resolve each placeholder with current task context plus `placeholders.json` from the repository root.
6. Use the returned `execution_request` as the meta prompt that tells you how to build the final executable prompt after placeholder resolution.
7. Execute the resolved prompt contract now.
8. Return:
   1. category confirmation
   2. placeholder resolution summary
   3. final executable prompt
   4. execution result

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
- Run [VALIDATION] and show raw output.

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
- Run go test ./... -v and show raw output.

Review:
- Double-check the minimal patch and validate existing categories and optimize flags plus unchanged compiler selection behavior remains unchanged.

Use the selected category as the deterministic task frame.
1. Build the final executable prompt by resolving every required placeholder in compiled_prompt from repository facts and the current task context.
2. Keep the selected category and compiled structure intact; do not silently rewrite the contract.
3. If a required placeholder cannot be resolved safely, stop and ask for the missing input.
4. After placeholder resolution, execute the final prompt.

Return:
1. category confirmation
2. placeholder resolution summary
3. final executable prompt
4. execution result
```

## Example 2

### Original weak prompt

`make the search path faster without risky changes`

### Chosen category

`performance`

### Command invocation

```bash
vokuprompt categories
vokuprompt optimize --category performance
```

### Returned compiled prompt

```text
Goal:
[TASK]

Context:
[CONTEXT_TARGET]

Constraints:
- Do not claim a speedup without realistic dominant-workload evidence, and do not modify unrelated [SCOPE_ELEMENTS].

Execution:
- Measure the dominant workload first, then make one benchmark-backed change over [UNIT] at a time and continue until [DONE_CONDITION].

Validation:
- Run [VALIDATION] against realistic benchmarks that reflect the dominant workload and show raw output.

Review:
- Confirm the reported speedup comes from the dominant workload and that [STABLE_INTERFACE] remains unchanged.
```

### Placeholder resolution

- `[TASK]` → Fix the search-path performance regression with the smallest safe change and prove the gain with realistic benchmarks.
- `[CONTEXT_TARGET]` → the search endpoint hot path, query preparation, and the benchmark that captures the regression.
- `[SCOPE_ELEMENTS]` → public API response fields, unrelated handlers, storage migrations, and non-search workflows.
- `[UNIT]` → one measured hotspot or benchmark-backed micro-change.
- `[DONE_CONDITION]` → the targeted search benchmark shows the intended latency improvement under the dominant workload and `go test ./... -v` still passes.
- `[VALIDATION]` → `go test ./... -v && go test ./... -run Search -bench Search -benchmem`
- `[STABLE_INTERFACE]` → the existing search response schema, optimize CLI contract, and deterministic prompt rendering.

### Final executed prompt

```text
Goal:
Fix the search-path performance regression with the smallest safe change and prove the gain with realistic benchmarks.

Context:
the search endpoint hot path, query preparation, and the benchmark that captures the regression.

Constraints:
- Do not claim a speedup without realistic dominant-workload evidence, and do not modify unrelated public API response fields, unrelated handlers, storage migrations, and non-search workflows.

Execution:
- Measure the dominant workload first, then make one benchmark-backed change over one measured hotspot or benchmark-backed micro-change at a time and continue until the targeted search benchmark shows the intended latency improvement under the dominant workload and go test ./... -v still passes.

Validation:
- Run go test ./... -v && go test ./... -run Search -bench Search -benchmem against realistic benchmarks that reflect the dominant workload and show raw output.

Review:
- Confirm the reported speedup comes from the dominant workload and that the existing search response schema, optimize CLI contract, and deterministic prompt rendering remains unchanged.

Use the selected category as the deterministic task frame.
1. Build the final executable prompt by resolving every required placeholder in compiled_prompt from repository facts and the current task context.
2. Keep the selected category and compiled structure intact; do not silently rewrite the contract.
3. If a required placeholder cannot be resolved safely, stop and ask for the missing input.
4. After placeholder resolution, execute the final prompt.

Return:
1. category confirmation
2. placeholder resolution summary
3. final executable prompt
4. execution result
```
