# Code Tracing Playbook

## Core rules

- Check documents first, then code, then ask the user.
- Trace until the business loop closes.
- Stop at generic infrastructure only after confirming there is no feature-specific logic inside it.

## Standard procedure

1. Confirm the feature name and the intended scope.
2. Search `wiki/` and `raw/` for matching capability, design, postmortem, and code-summary documents.
3. Enter the codebase through the user-provided anchor.
4. For each link in the chain, record:
   - current file
   - current function / class / route
   - upstream callers
   - downstream callees
   - critical branches and conditions
   - data structures, configuration, tables, and external APIs
5. Move one layer at a time until the full business path is covered.
6. Cross-check current code behavior against existing docs.
7. Fill the template with confirmed facts and explicit open questions.

## Trace order by anchor type

### If the user gives an API URL

1. Find the route, middleware, controller, or handler.
2. Find parsing, auth, and request validation.
3. Find the service / use-case / domain logic.
4. Find storage, cache, queue, and external calls.
5. Find response assembly and error branches.

### If the user gives a key file

1. Read the full file and identify public entries.
2. Find callers and callees for each relevant entry.
3. Follow critical branches instead of stopping at helper boundaries.
4. Expand into upstream and downstream files as needed.

### If the user gives a key function

1. Find the definition.
2. Find all callers.
3. Find all important callees.
4. Track inputs, outputs, and side effects.

### If the user gives a page path or route

1. Find the frontend route and entry component.
2. Find page state, forms, and action handlers.
3. Find API client wrappers.
4. Continue into the backend using the API trace order.

## Generic infrastructure that usually does not need deep tracing

- shared logging
- generic auth middleware
- ORM base wrappers
- generic HTTP clients
- common utility helpers

Keep tracing if any of them contain feature-specific behavior.
