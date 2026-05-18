---
name: fixdocs
description: "Fix Terraform provider end user documentation issues detected by swissshepherd (ss). Removes an ignored target from the config, runs ss, validates findings, fixes the documentation, and commits."
---

<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Skill: Fix End User Documentation With `swissshepherd`

Fix Terraform provider documentation issues by removing targets from the swissshepherd ignore list, validating findings against the schema, and correcting the documentation.

## When to use

Trigger this skill when the user:
- Says "fix docs", "fixdocs", "run swissshepherd", "run ss", or similar
- Asks to remove a target from the swissshepherd ignore list
- Asks to fix documentation for a specific resource or data source
- Provides swissshepherd output and asks to fix the findings

## Inputs

Optional:
- A target name (e.g., `aws_s3_bucket`) or prefix (e.g., `aws_s3_`)
- A type (`resource`, `data_source`, `ephemeral`, etc.)
- Number of targets to process (default: 1)

If no target is specified, pick the next one from `ignore_targets` in `.ci/swissshepherd-weak.hcl`.

## Procedure

### Phase 1: Verify clean baseline

```bash
make swissshepherd
```

This MUST output "All checks passed." before proceeding. If it doesn't, stop and tell the user the baseline is dirty.

### Phase 2: Select and expose target

1. Open `.ci/swissshepherd-weak.hcl`
2. Find the target in an `ignore_targets` list (within a `check` block)
3. Remove the target line from the list
4. Run swissshepherd to see findings:

```bash
swissshepherd --config .ci/swissshepherd-weak.hcl --target <name> --type <type>
```

If "All checks passed" — the target was already clean. Commit the config removal and move to the next target.

### Phase 3: Validate findings

For each finding, determine if it's valid by checking the schema source of truth:

- **Coverage errors** ("not documented", "does not exist in schema"): Read the resource's Go source (`internal/service/<service>/<resource>.go` or `*_data_source.go`) to confirm the attribute/block exists in the schema.
- **Heading errors**: Check what the heading currently says vs what ss expects.
- **Label errors**: Check if the attribute is Required, Optional, or Computed in the schema.
- **Byline errors**: Compare against the expected bylines in the config.

If a finding appears to be a swissshepherd bug (schema says one thing, ss reports another), note it and skip — do NOT fix the doc incorrectly.

### Phase 4: Fix the documentation

Open the doc file (path is in the ss output) and apply fixes:

| Finding | Fix |
|---------|-----|
| "block X is not documented" | Add a `### \`block_name\` Block` section with its attributes listed |
| "attribute X should be documented in Attribute Reference" | Add to Attribute Reference section |
| "attribute X should not appear in Argument Reference" | Move from Arguments to Attributes |
| "documented attribute X does not exist in schema" | Remove from docs (it's phantom) |
| "missing (Required) or (Optional) label" | Add the correct label based on schema |
| "heading ... should be ..." | Rename to the suggested heading |
| "byline does not match expected texts" | Replace with a standard byline |
| "reference-style link definition" | Convert `[ref]: url` to inline `[text](url)` |

#### Documentation style rules

Authoritative reference: [docs/end-user-documentation.md](../../../docs/end-user-documentation.md). When this skill and that document disagree, the document wins.

### Phase 5: Verify fix

```bash
swissshepherd --config .ci/swissshepherd-weak.hcl --target <name> --type <type>
```

Must output "All checks passed." If not, iterate on remaining findings.

### Phase 6: Full verification

```bash
make swissshepherd
```

Must output "All checks passed." to confirm no regressions.

### Phase 7: Commit

Stage and commit:

```bash
git add .ci/swissshepherd-weak.hcl website/docs/
git commit -m "<resource_name>: Fix documentation per swissshepherd"
```

Use the resource name without the `aws_` prefix in the commit message scope when it matches a single service. For multi-target batches, use the service name.

## Important constraints

- **Schema is truth.** Never "fix" a finding by silencing it if the schema confirms the issue.
- **One target per commit.** Each removed ignore target gets its own commit for clean git history.
- **Always use `--config .ci/swissshepherd-weak.hcl`** — running without config produces 20,000+ findings.
- **Check both check blocks.** A target may appear in `ignore_targets` under `check "schema_docs"` AND `check "import_section"` (or others). Remove from all.
- **Don't touch `ignore_contents_check`** unless the user explicitly asks — those are structural exceptions.
- **Preserve file structure.** Don't rewrite entire doc files. Make minimal, targeted edits.
- **Nested blocks.** When ss reports "block X.Y is not documented", the doc needs a subsection under the parent block's section. Use `` ### `y` Block `` nested contextually after the parent.

## Example session

User: "fix docs for aws_s3_bucket_lifecycle_configuration"

1. Verify `make swissshepherd` passes
2. Remove `aws_s3_bucket_lifecycle_configuration` from `ignore_targets` in the `schema_docs` check block
3. Run `swissshepherd --config .ci/swissshepherd-weak.hcl --target aws_s3_bucket_lifecycle_configuration --type resource`
4. See: `ERROR [schema_docs] ... block "rule.filter" is not documented`
5. Check Go source — confirm `filter` block exists under `rule`
6. Add `` ### `filter` Block `` section with its attributes
7. Re-run ss — passes
8. Run `make swissshepherd` — passes
9. Commit
