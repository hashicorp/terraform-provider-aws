# Team Working Board

Used to automate the AWS Provider Team's working board.

## Usage

### Inputs

| Input | Required | Description |
| ----- | -------- | ----------- |
| `github_token` | true | The token used to authenticate with the GitHub API |
| `item_url` | true | The URL of the Issue or Pull Request |
| `move_to_top` | false | Whether to move the item to the top of the list |
| `status` | false | The Status the item should be set to |
| `view` | false | The View the item should be assigned to |

### Example

```yaml
steps:
  - name: Checkout
    uses: actions/checkout@692973e3d937129bcbf40652eb9f2f61becf3332 # v4.1.7
    with:
      sparse-checkout: .github/actions/team_working_board

  - name: Add Issue to Working Board
    uses: ./.github/actions/team_working_board
    with:
      github_token: ${{ secrets.GITHUB_TOKEN }}
      item_url: ${{ github.event.pull_request.html_url }}
      status: "To Do"
      view: "working-board"
```
