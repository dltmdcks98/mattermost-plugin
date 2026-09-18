# Mattermost Plugin Codex Instructions

This file is a Codex copy adapted from `.agent/HERMES_POLICIES.md`. Keep the Hermes policy source read-only.

- This repository manages the Mattermost `/hermes` slash-command plugin and its Hermes Agent API integration.
- Inspect plugin code under `hermes/` first and run `make -C hermes test` and `make -C hermes bundle` when feasible.
- Treat slash-command responses as a user-facing operational interface; verify error messages, timeouts, API endpoints, authentication, and token handling.
- Confirm that `dist/` or bundle output is an intended release artifact before committing it. Exclude incidental build output.
- When changing Hermes API routes, check effects on Mattermost DM/channel invocation rules and `@hermes` notification behavior.
