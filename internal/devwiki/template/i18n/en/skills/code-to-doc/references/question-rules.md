# User Question Rules

Try a few independent searches first. If the problem remains unresolved, stop and ask the user in the following cases.

## Mandatory ask cases

- Only the feature name is known, and multiple candidate implementations remain after several searches.
- The provided API URL, page path, or key function cannot be found in the local codebase.
- Dynamic registration, reflection, configuration, code generation, or templates prevent reliable static confirmation.
- A critical external system, gateway, or third-party dependency is outside the local repo and has no local docs.
- The target file already exists and the user must choose update versus new document.
- Existing `wiki/` or `raw/` content conflicts with the current code and the latest truth cannot be inferred safely.

## Ask format

Keep questions short and specific. Ask only for the missing anchor.

Prefer prompts like:

- "`<feature-name>` currently matches two entry paths: `a` and `b`. Which path should I trace first?"
- "I could not find API `<URL>` in the local repo. Is it implemented in a gateway, another repo, or an external service?"
- "I can see the caller but not the downstream implementation. Can you provide a key function, module name, or related endpoint?"

## Forbidden question patterns

- Do not ask before attempting your own searches.
- Do not ask vague questions like "can you provide more context?"
- Do not interrupt repeatedly when the trace can continue without user help.
