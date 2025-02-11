# Community Check

Check a username to see if it's in one of our community lists. We use this to help automate tasks within the repository.

## Usage

### Inputs

| Input               | Required | Description                                   |
| ------------------- | -------- | --------------------------------------------- |
| `user_login`        | true     | The GitHub username to check                  |
| `core_contributors` | false    | The base64 encoded list of Core Contributors  |
| `maintainers`       | false    | The base64 encoded list of maintainers        |
| `partners`          | false    | The base64 encoded list of partner contritors |

### Outputs

| Output             | Default | Description                               |
| ------------------ | ------- | ----------------------------------------- |
| `core_contributor` | `null`  | Whether the user is a Core Contributor    |
| `maintainer`       | `null`  | Whether the user is a maintainer          |
| `partner`          | `null`  | Whether the user is a partner contributor |

### Example

```yaml
steps:
  - name: Checkout
    uses: actions/checkout@692973e3d937129bcbf40652eb9f2f61becf3332 # v4.1.7
    with:
      sparse-checkout: .github/actions/community_check

  - name: Community Check
    id: community_check
    uses: ./.github/actions/community_check
    with:
      user_login: ${{ github.event.issue.user.login }}
      maintainers: ${{ secrets.MAINTAINERS }}
      core_contributors: ${{ secrets.CORE_CONTRIBUTORS }}
      partners: ${{ secrets.PARTNERS }}
```
