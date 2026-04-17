# Changelog

## v1.0.0

### Added

- Prompt compiler: `vokuprompt optimize --category <name>` compiles a category-specific execution prompt from `patterns.json` and `categories.json`.
- Category listing: `vokuprompt categories` returns all available categories with descriptions.
- Five built-in categories: `bugfix`, `performance`, `refactor`, `review`, `tests`.
- Placeholder manifest in optimize output: lists every placeholder, its required status, section type, and source pattern.
- Placeholder registry: `placeholders.json` with description, examples, resolution guidance, expected format, and preferred sources for every placeholder.
- Execution request: optimize output includes a ready-to-use `execution_request` that instructs analyze → improve → execute.
- Explain mode: `vokuprompt optimize --category <name> --explain` adds a selection trace with:
  - `required_patterns`: which patterns were mandatory for this category.
  - `rejected_by_conflict`: patterns dropped because a selected pattern conflicts with them.
  - `rejected_by_role_limits`: patterns dropped because their role slot was already filled, including `blocked_by` naming the winner.
  - `required_placeholders`: flat list of all placeholders that must be resolved.
- Agent skill: `skills/vokuprompt/SKILL.md` with workflow, resolution rules, and two worked examples.
- End-to-end example fixtures for all five categories in `examples/`.
- Public README with onboarding, build instructions, command examples, and knowledge-base file map.
- Full test suite: compiler determinism, section ordering, placeholder manifest contract, example fixture verification, explain contract, and category behavior.

### Design decisions

- No embedded LLM. The tool is a deterministic compiler; the agent or human using it brings the execution context.
- No automatic category detection. The caller chooses the category; the tool compiles the prompt.
- Analysis section renders before Validation in all prompts so the review category reads: execute → analyze → validate → challenge.
- Pattern weights and role limits are the only selection mechanism. Conflicts are declared explicitly in `patterns.json`.
