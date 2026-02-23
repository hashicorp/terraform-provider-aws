# Next Version

Composite action that determines the next version based on branch and event context.

## Inputs

- `manual-version` (optional): Manual version override (e.g., `1.2.3` or `1.2.3-beta1`)

## Outputs

- `core-version`: Core version without prerelease (e.g., `1.2.3`)
- `prerelease`: Prerelease identifier (e.g., `beta1`)
- `version`: Full version string (e.g., `1.2.3-beta1` or `1.2.3`)

## Version Determination Logic

| Branch | Event | Condition | Version | Prerelease |
|--------|-------|-----------|---------|------------|
| `main` | PR merge | Normal PR | `changie next minor` | - |
| `main` | PR merge | From `release/*` | `version/VERSION` | - |
| `release/*` | Any | First beta (latest=0.0.0) | `version/VERSION` | `beta1` |
| `release/*` | Any | Existing prerelease | Same version | Increment number |
| `release/*` | Any | GA exists | `changie next minor` | `beta1` |
| Any | workflow_dispatch | Manual input | Parsed from input | Parsed from input |

## Usage

```yaml
- name: Determine Next Version
  id: version
  uses: ./.github/actions/next_version
  with:
    manual-version: ${{ inputs.version }}  # optional

- name: Use Version
  run: |
    echo "Core Version: ${{ steps.version.outputs.core-version }}"
    echo "Prerelease: ${{ steps.version.outputs.prerelease }}"
    echo "Full Version: ${{ steps.version.outputs.version }}"
```