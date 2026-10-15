<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# CHANGELOG Entries

`CHANGELOG.md` is generated from small fragment files in this directory using
[`go-changelog`](https://github.com/hashicorp/go-changelog). Add **one file per
pull request**, named for the PR number: `.changelog/<PR_NUMBER>.txt`.

> [!NOTE]
> This tooling is slated to be replaced by [Changie](https://changie.dev/). See
> [`docs/design-decisions/changie-migration.md`](../docs/design-decisions/changie-migration.md).
> Keep new entries in this format until that migration lands.

> [!IMPORTANT]
> The example blocks in this file are **indented** on purpose. `go-changelog`
> treats every file in this directory as a changelog entry and extracts any
> fenced block that begins with ` ```release-note ` at the **start of a line**.
> Indenting the examples keeps this README from being parsed as an entry (which
> would otherwise emit bogus `#README.md` links into `CHANGELOG.md`). In a real
> `.changelog/<PR_NUMBER>.txt` file, the fence must start at column 0 with no
> leading whitespace.

## When is an entry required?

Only for **user-facing** changes:

- New resource, data source, ephemeral resource, list resource, function, or action
- Bug fixes
- Enhancements (new arguments/attributes, new valid values, behavior improvements)
- Deprecations, breaking changes, or anything users must be told about

No entry is needed for tests, documentation-only changes, CI, sweepers, internal
refactors, or dependency bumps that have no user-facing effect.

## Tips

- **Imperative voice.** Write "Add", "Fix", "Support", "Remove" — not "Adds", "Added", or "Fixes". (The `new-*` types are the exception: just the name.)
- **Brief and specific.** One line describing the user-visible change. Name the concrete error, argument, or behavior rather than "various fixes".
- **Backtick identifiers.** Wrap resource names, arguments, attributes, and literal values in backticks.
- **Don't start descriptions with `The`, `A`, or `An`**.
- **One file per PR**, named `<PR_NUMBER>.txt`; combine multiple related notes as separate blocks in that single file.

## Format

Each entry is one or more fenced blocks tagged with a type (shown indented here;
in a real entry file the fence starts at column 0):

    ```release-note:TYPE
    <component>: <description>
    ```

`<component>` is the resource/data source/etc. name (for example
`resource/aws_s3_bucket`, `data-source/aws_s3_bucket`) or `provider` for
provider-wide changes. A single `.txt` file may contain multiple blocks.

## Types

Grouped by the `CHANGELOG.md` heading each type renders into. For the `new-*`
types the body is just the resource/function/action/guide name (no component
prefix, no verb).

### FEATURES

    ```release-note:new-resource
    aws_bedrockagentcore_api_key_credential_provider
    ```

    ```release-note:new-data-source
    aws_s3control_access_points
    ```

    ```release-note:new-ephemeral
    aws_sts_web_identity_token
    ```

    ```release-note:new-list-resource
    aws_db_subnet_group
    ```

    ```release-note:new-function
    arn_parse
    ```

    ```release-note:new-action
    aws_sfn_start_execution
    ```

    ```release-note:new-guide
    Tag Policy Compliance
    ```

### ENHANCEMENTS

    ```release-note:enhancement
    resource/aws_ssm_resource_data_sync: Add `s3_destination.destination_data_sharing` argument
    ```

### BUG FIXES

    ```release-note:bug
    resource/aws_glue_catalog_table: Fix `Invalid address to set` errors when reading `partition_keys.parameters`
    ```

### NOTES

    ```release-note:note
    resource/aws_dms_s3_endpoint: The `kms_key_arn` attribute has been deprecated. Use `server_side_encryption_kms_key_id` instead
    ```

### BREAKING CHANGES

    ```release-note:breaking-change
    resource/aws_db_instance: `character_set_name` can no longer be set with `replicate_source_db`, `restore_to_point_in_time`, `s3_import`, or `snapshot_identifier`
    ```
