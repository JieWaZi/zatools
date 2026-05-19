# Source Priority

Use sources in this order by default:

1. `wiki/`
   - Start with existing capability, feature, workflow, and troubleshooting pages to avoid re-discovering known context.
2. `raw/`
   - Review requirement docs, design docs, feature notes, and test docs to understand prior intent.
3. Local codebase
   - Confirm the real implementation, fill gaps, and correct drift in older pages.
4. User clarification
   - Ask only when independent investigation cannot continue safely or a conflict must be resolved.

## Usage rules

- `wiki/` and `raw/` are clues and history, not the final truth.
- Final claims must be grounded in current code.
- If code and docs disagree, say so explicitly in the proposal or generated workflow page.
