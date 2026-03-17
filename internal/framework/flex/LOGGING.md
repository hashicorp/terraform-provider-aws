# AutoFlex Logging
<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

Logging in AutoFlex is intended to assist in debugging flattening and expanding values.

## Path

As AutoFlex walks the value to be flattened or expanded, it keeps track of the current path, including object fields and collection indexes.
The path is tracked as a [Terraform Plugin Framework `path.Path`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework/path#Path).

## Logging Keys

The logging keys are structured to help interpretation.
All keys start with `autoflex`,
except for the key `error`.

Keys differentiate whether they apply to the source or target value using the `source` or `target` elements.

New logging keys should follow the same pattern.

| Key                         | Description |
+-----------------------------+-------------+
| `autoflex.source.type`      | The fully qualified type of the source at `path` |
| `autoflex.source.fieldname` | When flattening or expanding within a struct or object value, the current field in the source |
| `autoflex.source.path`      | The current path within the source |
| `autoflex.source.size`      | When flattening or expanding a collection type, the size of the source collection |
| `autoflex.target.type`      | The fully qualified type of the target at `path` |
| `autoflex.target.fieldname` | When flattening or expanding within a struct or object value, the current field in the target |
| `autoflex.target.path`      | The current path within the target |
| `error`                     | Error value for any errors |
