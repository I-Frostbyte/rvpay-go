## Required Reading

Read only the following documents before beginning work:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md

Use these documents as the source of truth.

Do not inspect the repository recursively.

Only inspect additional directories when explicitly instructed below.

Never review:

- .git/
- .github/
- third_party/
- grpc/go/
- vendor/
- node_modules/
- testdata/
- coverage/
- bin/
- tmp/
- docs/ (except the documents listed above)

Use README.md as the repository map.

## Completion Rules

Before completing this task verify that:

- Existing functionality has not been broken.
- Existing public APIs remain compatible unless instructed otherwise.
- Existing package naming conventions are preserved.
- Existing logging style is preserved.
- Existing error handling style is preserved.
- Existing folder structure is preserved.
- Existing code is extended rather than rewritten.
- No generated code has been manually edited.
- No secrets have been committed.
- All newly created files compile.
- All tests affected by this task pass.

If a prerequisite from an earlier agent is missing, stop and explain why rather than creating a partial implementation.