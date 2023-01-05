<a name="v1.6.21"></a>
# [v1.6.21](https://github.com/rhysd/actionlint/releases/tag/v1.6.21) - 09 Oct 2022

- Check contexts availability. Some contexts limit where they can be used. For example, `jobs.<job_id>.env` workflow key does not allow accessing `env` context, but `jobs.<job_id>.steps.env` allows. See [the official document](https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability) for the complete list of contexts availability. ([#180](https://github.com/rhysd/actionlint/issues/180))
  ```yaml
  ...

  env:
    TOPLEVEL: ...

  jobs:
    test:
      runs-on: ubuntu-latest
      env:
        # ERROR: 'env' context is not available here
        JOB_LEVEL: ${{ env.TOPLEVEL }}
      steps:
        - env:
            # OK: 'env' context is available here
            STEP_LEVEL: ${{ env.TOPLEVEL }}
          ...
  ```
  actionlint reports the context is not available and what contexts are available as follows:
  ```
  test.yaml:11:22: context "env" is not allowed here. available contexts are "github", "inputs", "matrix", "needs", "secrets", "strategy". see https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability for more details [expression]
     |
  11 |       JOB_LEVEL: ${{ env.TOPLEVEL }}
     |                      ^~~~~~~~~~~~
  ```
- Check special functions availability. Some functions limit where they can be used. For example, status functions like `success()` or `failure()` are only available in conditions of `if:`. See [the official document](https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability) for the complete list of special functions availability. ([#214](https://github.com/rhysd/actionlint/issues/214))
  ```yaml
  ...

  steps:
    # ERROR: 'success()' function is not available here
    - run: echo 'Success? ${{ success() }}'
      # OK: 'success()' function is available here
      if: success()
  ```
  actionlint reports `success()` is not available and where the function is available as follows:
  ```
  test.yaml:8:33: calling function "success" is not allowed here. "success" is only available in "jobs.<job_id>.if", "jobs.<job_id>.steps.if". see https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability for more details [expression]
    |
  8 |       - run: echo 'Success? ${{ success() }}'
    |                                 ^~~~~~~~~
  ```
- Fix `inputs` context is not available in `run-name:` section. ([#223](https://github.com/rhysd/actionlint/issues/223))
- Allow dynamic shell configuration like `shell: ${{ env.SHELL }}`.
- Fix no error is reported when `on:` does not exist at toplevel. ([#232](https://github.com/rhysd/actionlint/issues/232))
- Fix an error position is not correct when the error happens at root node of workflow AST.
- Fix an incorrect empty event is parsed when `on:` section is empty.
- Fix the error message when parsing an unexpected key on toplevel. ([#231](https://github.com/rhysd/actionlint/issues/231), thanks [@norwd](https://github.com/norwd))
- Add `in_progress` type to `workflow_run` webhook event trigger.
- Describe [the actionlint extension](https://extensions.panic.com/extensions/org.netwrk/org.netwrk.actionlint/) for [Nova.app](https://nova.app) in [the usage document](https://github.com/rhysd/actionlint/blob/main/docs/usage.md#nova). ([#222](https://github.com/rhysd/actionlint/issues/222), thanks [@jbergstroem](https://github.com/jbergstroem))
- Note [Super-Linter](https://github.com/github/super-linter) uses a different place for configuration file. ([#227](https://github.com/rhysd/actionlint/issues/227), thanks [@per-oestergaard](https://github.com/per-oestergaard))
- Add `actions/setup-dotnet@v3` to popular actions data set.
- [`generate-availability` script](https://github.com/rhysd/actionlint/tree/main/scripts/generate-availability) was created to scrape the information about contexts and special functions availability from the official document. The information is available through `actionlint.WorkflowKeyAvailability()` Go API. This script is run once a week on CI to keep the information up-to-date.



[Changes][v1.6.21]


<a name="v1.6.20"></a>
# [v1.6.20](https://github.com/rhysd/actionlint/releases/tag/v1.6.20) - 30 Sep 2022

- Support `run-name` which [GitHub introduced recently](https://github.blog/changelog/2022-09-26-github-actions-dynamic-names-for-workflow-runs/). It is a name of workflow run dynamically configured. See [the official document](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#run-name) for more details. ([#220](https://github.com/rhysd/actionlint/issues/220))
  ```yaml
  on: push
  run-name: Deploy by @${{ github.actor }}

  jobs:
    ...
  ```
- Add `end_column` property to JSON representation of error. The property indicates a column of the end position of `^~~~~~~` indicator in snippet. Note that `end_column` is equal to `column` when the indicator cannot be shown. ([#219](https://github.com/rhysd/actionlint/issues/219))
  ```console
  $ actionlint -format '{{json .}}' test.yaml | jq
  [
    {
      "message": "property \"unknown_prop\" is not defined in object type {arch: string; debug: string; name: string; os: string; temp: string; tool_cache: string; workspace: string}",
      "filepath": "test.yaml",
      "line": 7,
      "column": 23,
      "kind": "expression",
      "snippet": "      - run: echo ${{ runner.unknown_prop }}\n                      ^~~~~~~~~~~~~~~~~~~",
      "end_column": 41
    }
  ]
  ```
- Overhaul the workflow parser to parse workflow keys in case-insensitive. This is a work derived from the fix of [#216](https://github.com/rhysd/actionlint/issues/216). Now the parser parses all workflow keys in case-insensitive way correctly. Note that permission names at `permissions:` are exceptionally case-sensitive.
  - This fixes properties of `inputs` for `workflow_dispatch` were not case-insensitive.
  - This fixes inputs and outputs of local actions were not handled in case-insensitive way.
- Update popular actions data set. `actions/stale@v6` was newly added.

[Changes][v1.6.20]


<a name="v1.6.19"></a>
# [v1.6.19](https://github.com/rhysd/actionlint/releases/tag/v1.6.19) - 22 Sep 2022

- Fix inputs, outputs, and secrets of reusable workflow should be case-insensitive. ([#216](https://github.com/rhysd/actionlint/issues/216))
  ```yaml
  # .github/workflows/reusable.yaml
  on:
    workflow_call:
      inputs:
        INPUT_UPPER:
          type: string
        input_lower:
          type: string
      secrets:
        SECRET_UPPER:
        secret_lower:
  ...

  # .github/workflows/test.yaml
  ...

  jobs:
    caller:
      uses: ./.github/workflows/reusable.yaml
      # Inputs and secrets are case-insensitive. So all the followings should be OK
      with:
        input_upper: ...
        INPUT_LOWER: ...
      secrets:
        secret_upper: ...
        SECRET_LOWER: ...
  ```
- Describe [how to install specific version of `actionlint` binary with the download script](https://github.com/rhysd/actionlint/blob/main/docs/install.md#download-script). ([#218](https://github.com/rhysd/actionlint/issues/218))

[Changes][v1.6.19]


<a name="v1.6.18"></a>
# [v1.6.18](https://github.com/rhysd/actionlint/releases/tag/v1.6.18) - 17 Sep 2022

- This release much enhances checks for local reusable workflow calls. Note that these checks are done for local reusable workflows (starting with `./`). ([#179](https://github.com/rhysd/actionlint/issues/179)).
  - Detect missing required inputs/secrets and undefined inputs/secrets at `jobs.<job_id>.with` and `jobs.<job_id>.secrets`. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#check-inputs-and-secrets-in-workflow-call) for more details.
    ```yaml
    # .github/workflows/reusable.yml
    on:
      workflow_call:
        inputs:
          name:
            type: string
            required: true
        secrets:
          password:
            required: true
    ...

    # .github/workflows/test.yml
    ...

    jobs:
      missing-required:
        uses: ./.github/workflows/reusable.yml
        with:
          # ERROR: Undefined input "user"
          user: rhysd
          # ERROR: Required input "name" is missing
        secrets:
          # ERROR: Undefined secret "credentials"
          credentials: my-token
          # ERROR: Required secret "password" is missing
    ```
  - Type check for reusable workflow inputs at `jobs.<job_id>.with`. Types are defined at `on.workflow_call.inputs.<name>.type` in reusable workflow. actionlint checks types of expressions in workflow calls. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#check-inputs-and-secrets-in-workflow-call) for more details.
    ```yaml
    # .github/workflows/reusable.yml
    on:
      workflow_call:
        inputs:
          id:
            type: number
          message:
            type: string
    ...

    # .github/workflows/test.yml
    ...

    jobs:
      type-checks:
        uses: ./.github/workflows/reusable.yml
        with:
          # ERROR: Cannot assign string value to number input. format() returns string value
          id: ${{ format('runner name is {0}', runner.name) }}
          # ERROR: Cannot assign null to string input. If you want to pass string "null", use ${{ 'null' }}
          message: null
    ```
  - Detect local reusable workflow which does not exist at `jobs.<job_id>.uses`. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#check-workflow-call-syntax) for more details.
    ```yaml
    jobs:
      test:
        # ERROR: This workflow file does not exist
        with: ./.github/workflows/does-not-exist.yml
    ```
  - Check `needs.<job_id>.outputs.<output_id>` in downstream jobs of workflow call jobs. The outputs object is now typed strictly based on `on.workflow_call.outputs.<name>` in the called reusable workflow. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#check-outputs-of-workflow-call-in-downstream-jobs) for more details.
    ```yaml
    # .github/workflows/get-build-info.yml
    on:
      workflow_call:
        outputs:
          version:
            value: ...
            description: version of software
    ...

    # .github/workflows/test.yml
    ...

    jobs:
      # This job's outputs object is typed as {version: string}
      get_build_info:
        uses: ./.github/workflows/get-build-info.yml
      downstream:
        needs: [get_build_info]
        runs-on: ubuntu-latest
        steps:
          # OK. `version` is defined in the reusable workflow
          - run: echo '${{ needs.get_build_info.outputs.version }}'
          # ERROR: `tag` is not defined in the reusable workflow
          - run: echo '${{ needs.get_build_info.outputs.tag }}'
    ```
- Add missing properties in contexts and improve types of some properties looking at [the official contexts document](https://docs.github.com/en/actions/learn-github-actions/contexts#github-context).
  - `github.action_status`
  - `runner.debug`
  - `services.<service_id>.ports`
- Fix `on.workflow_call.inputs.<name>.description` and `on.workflow_call.secrets.<name>.description` were incorrectly mandatory. They are actually optional.
- Report parse errors when parsing `action.yml` in local actions. They were ignored in previous versions.
- Sort the order of properties in an object type displayed in error message. In previous versions, actionlint sometimes displayed `{a: true, b: string}`, or it displayed `{b: string, a: true}` for the same object type. This randomness was caused by random iteration of map values in Go.
- Update popular actions data set to the latest.

[Changes][v1.6.18]


<a name="v1.6.17"></a>
# [v1.6.17](https://github.com/rhysd/actionlint/releases/tag/v1.6.17) - 28 Aug 2022

- Allow workflow calls are available in matrix jobs. See [the official announcement](https://github.blog/changelog/2022-08-22-github-actions-improvements-to-reusable-workflows-2/) for more details. ([#197](https://github.com/rhysd/actionlint/issues/197))
  ```yaml
  jobs:
    ReuseableMatrixJobForDeployment:
      strategy:
        matrix:
          target: [dev, stage, prod]
      uses: octocat/octo-repo/.github/workflows/deployment.yml@main
      with:
        target: ${{ matrix.target }}
  ```
- Allow nested workflow calls. See [the official announcement](https://github.blog/changelog/2022-08-22-github-actions-improvements-to-reusable-workflows-2/) for more details. ([#201](https://github.com/rhysd/actionlint/issues/201))
  ```yaml
  on: workflow_call

  jobs:
    call-another-reusable:
      uses: path/to/another-reusable.yml@v1
  ```
- Fix job outputs should be passed to `needs.*.outputs` of only direct children. Until v1.6.16, they are passed to any downstream jobs. ([#151](https://github.com/rhysd/actionlint/issues/151))
  ```yaml
  jobs:
    first:
      runs-on: ubuntu-latest
      outputs:
        first: 'output from first job'
      steps:
        - run: echo 'first'

    second:
      needs: [first]
      runs-on: ubuntu-latest
      outputs:
        second: 'output from second job'
      steps:
        - run: echo 'second'

    third:
      needs: [second]
      runs-on: ubuntu-latest
      steps:
        - run: echo '${{ toJSON(needs.second.outputs) }}'
        # ERROR: `needs.first` does not exist, but v1.6.16 reported no error
        - run: echo '${{ toJSON(needs.first.outputs) }}'
  ```
  When you need both `needs.first` and `needs.second`, add the both to `needs:`.
  ```yaml
    third:
      needs: [first, second]
      runs-on: ubuntu-latest
      steps:
        # OK
        -  echo '${{ toJSON(needs.first.outputs) }}'
  ```
- Fix `}}` in string literals are detected as end marker of placeholder `${{ }}`. ([#205](https://github.com/rhysd/actionlint/issues/205))
  ```yaml
  jobs:
    test:
      runs-on: ubuntu-latest
      strategy:
        # This caused an incorrect error until v1.6.16
        matrix: ${{ fromJSON('{"foo":{}}') }}
  ```
- Fix `working-directory:` should not be available with `uses:` in steps. `working-directory:` is only available with `run:`. ([#207](https://github.com/rhysd/actionlint/issues/207))
  ```yaml
  steps:
    - uses: actions/checkout@v3
      # ERROR: `working-directory:` is not available here
      working-directory: ./foo
  ```
- The working directory for running `actionlint` command can be set via [`WorkingDir` field of `LinterOptions` struct](https://pkg.go.dev/github.com/rhysd/actionlint#LinterOptions). When it is empty, the return value from `os.Getwd` will be used.
- Update popular actions data set. `actions/configure-pages@v2` was added.
- Use Go 1.19 on CI by default. It is used to build release binaries.
- Update dependencies (go-yaml/yaml v3.0.1).
- Update playground dependencies (except for CodeMirror v6).

[Changes][v1.6.17]


<a name="v1.6.16"></a>
# [v1.6.16](https://github.com/rhysd/actionlint/releases/tag/v1.6.16) - 19 Aug 2022

- Allow an empty object at `permissions:`. You can use it to disable permissions for all of the available scopes. ([#170](https://github.com/rhysd/actionlint/issues/170), [#171](https://github.com/rhysd/actionlint/issues/171), thanks [@peaceiris](https://github.com/peaceiris))
  ```yaml
  permissions: {}
  ```
- Support `github.triggering_actor` context value. ([#190](https://github.com/rhysd/actionlint/issues/190), thanks [@stefreak](https://github.com/stefreak))
- Rename `step-id` rule to `id` rule. Now the rule checks both job IDs and step IDs. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#id-naming-convention) for more details. ([#182](https://github.com/rhysd/actionlint/issues/182))
  ```yaml
  jobs:
    # ERROR: '.' cannot be contained in ID
    v1.2.3:
      runs-on: ubuntu-latest
      steps:
        - run: echo 'job ID with version'
          # ERROR: ID cannot contain spaces
          id: echo for test
    # ERROR: ID cannot start with numbers
    2d-game:
      runs-on: ubuntu-latest
      steps:
        - run: echo 'oops'
  ```
- Accessing `env` context in `jobs.<id>.if` is now reported as error. ([#155](https://github.com/rhysd/actionlint/issues/155))
  ```yaml
  jobs:
    test:
      runs-on: ubuntu-latest
      # ERROR: `env` is not available here
      if: ${{ env.DIST == 'arch' }}
      steps:
        - run: ...
  ```
- Fix actionlint wrongly typed some matrix value when the matrix is expanded with `${{ }}`. For example, `matrix.foo` in the following code is typed as `{x: string}`, but it should be `any` because it is initialized with the value from `fromJSON`. ([#145](https://github.com/rhysd/actionlint/issues/145))
  ```yaml
  strategy:
    matrix:
      foo: ${{ fromJSON(...) }}
      exclude:
        - foo:
            x: y
  ```
- Fix incorrect type check when multiple runner labels are set to `runs-on:` via expanding `${{ }}` for selecting self-hosted runners. ([#164](https://github.com/rhysd/actionlint/issues/164))
  ```yaml
  jobs:
    test:
      strategy:
        matrix:
          include:
            - labels: ["self-hosted", "macOS", "X64"]
            - labels: ["self-hosted", "linux"]
      # actionlint incorrectly reported type error here
      runs-on: ${{ matrix.labels }}
  ```
- Fix usage of local actions (`uses: ./path/to/action`) was not checked when multiple workflow files were passed to `actionlint` command. ([#173](https://github.com/rhysd/actionlint/issues/173))
- Allow `description:` is missing in `secrets:` of reusable workflow call definition since it is optional. ([#174](https://github.com/rhysd/actionlint/issues/174))
- Fix type of propery of `github.event.inputs` is string unlike `inputs` context. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#workflow-dispatch-event-validation) for more details. ([#181](https://github.com/rhysd/actionlint/issues/181))
  ```yaml
  on:
    workflow_dispatch:
      inputs:
        is-valid:
          # Type of `inputs.is-valid` is bool
          # Type of `github.event.inputs.is-valid` is string
          type: boolean
  ```
- Fix crash when a value is expanded with `${{ }}` at `continue-on-error:`. ([#193](https://github.com/rhysd/actionlint/issues/193))
- Fix some error was caused by some other error. For example, the following code reported two errors. '" is not available for string literal' error caused another 'one placeholder should be included in boolean value string' error. This was caused because the `${{ x == "foo" }}` placeholder was not counted due to the previous type error.
  ```yaml
  if: ${{ x == "foo" }}
  ```
- Add support for [`merge_group` workflow trigger](https://github.blog/changelog/2022-08-18-merge-group-webhook-event-and-github-actions-workflow-trigger/).
- Add official actions to manage GitHub Pages to popular actions data set.
  - `actions/configure-pages@v1`
  - `actions/deploy-pages@v1`
  - `actions/upload-pages-artifact@v1`
- Update popular actions data set to the latest. Several new major versions and new inputs of actions were added to it.
- Describe how to install actionlint via [Chocolatey](https://chocolatey.org/), [scoop](https://scoop.sh/), and [AUR](https://aur.archlinux.org/) in [the installation document](https://github.com/rhysd/actionlint/blob/main/docs/install.md). ([#167](https://github.com/rhysd/actionlint/issues/167), [#168](https://github.com/rhysd/actionlint/issues/168), thanks [@sitiom](https://github.com/sitiom))
- [VS Code extension for actionlint](https://marketplace.visualstudio.com/items?itemName=arahata.linter-actionlint) was created by [@arahatashun](https://github.com/arahatashun). See [the document](https://github.com/rhysd/actionlint/blob/main/docs/usage.md#vs-code) for more details.
- Describe how to use [the Docker image](https://hub.docker.com/r/rhysd/actionlint) at step of GitHub Actions workflow. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/usage.md#use-actionlint-on-github-actions) for the details. ([#146](https://github.com/rhysd/actionlint/issues/146))
  ```yaml
  - uses: docker://rhysd/actionlint:latest
    with:
      args: -color
  ```
- Clarify the behavior if empty strings are set to some command line options in documents. `-shellcheck=` disables shellcheck integration and `-pyflakes=` disables pyflakes integration. ([#156](https://github.com/rhysd/actionlint/issues/156))
- Update Go module dependencies.

[Changes][v1.6.16]


<a name="v1.6.15"></a>
# [v1.6.15](https://github.com/rhysd/actionlint/releases/tag/v1.6.15) - 28 Jun 2022

- Fix referring `env` context from `env:` at step level caused an error. `env:` at toplevel and job level cannot refer `env` context, but `env:` at step level can. ([#158](https://github.com/rhysd/actionlint/issues/158))
  ```yaml
  on: push

  env:
    # ERROR: 'env:' at toplevel cannot refer 'env' context
    ERROR1: ${{ env.PATH }}

  jobs:
    my_job:
      runs-on: ubuntu-latest
      env:
        # ERROR: 'env:' at job level cannot refer 'env' context
        ERROR2: ${{ env.PATH }}
      steps:
        - run: echo "$THIS_IS_OK"
          env:
            # OK: 'env:' at step level CAN refer 'env' context
            THIS_IS_OK: ${{ env.PATH }}
  ```
- [Docker image for linux/arm64](https://hub.docker.com/layers/rhysd/actionlint/1.6.15/images/sha256-f63ee59f1846abce86ca9de1d41a1fc22bc7148d14b788cb455a9594d83e73f7?context=repo) is now provided. It is useful for M1 Mac users. ([#159](https://github.com/rhysd/actionlint/issues/159), thanks [@politician](https://github.com/politician))
- Fix [the download script](https://github.com/rhysd/actionlint/blob/main/scripts/download-actionlint.bash) did not respect the version specified via the first argument. ([#162](https://github.com/rhysd/actionlint/issues/162), thanks [@mateiidavid](https://github.com/mateiidavid))

[Changes][v1.6.15]


<a name="v1.6.14"></a>
# [v1.6.14](https://github.com/rhysd/actionlint/releases/tag/v1.6.14) - 26 Jun 2022

- Some filters are exclusive in events at `on:`. Now actionlint checks the exclusive filters are used in the same event. `paths` and `paths-ignore`, `branches` and `branches-ignore`, `tags` and `tags-ignore` are exclusive. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#webhook-events-validation) for the details.
  ```yaml
  on:
    push:
      # ERROR: Both 'paths' and 'paths-ignore' filters cannot be used for the same event
      paths: ...
      paths-ignore: ...
  ```
- Some event filters are checked more strictly. Some filters are only available with specific events. Now actionlint checks the limitation. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#webhook-events-validation) for complete list of such filters.
  ```yaml
  on:
    release:
      # ERROR: 'tags' filter is only available for 'push' event
      tags: v*.*.*
  ```
- Paths starting/ending with spaces are now reported as error.
- Inputs of workflow which specify both `default` and `required` are now reported as error. When `required` is specified at input of workflow call, a caller of it must specify value of the input. So the default value will never be used. ([#154](https://github.com/rhysd/actionlint/issues/154), thanks [@sksat](https://github.com/sksat))
  ```yaml
  on:
    workflow_call:
      inputs:
        my_input:
          description: test
          type: string
          # ERROR: The default value 'aaa' will never be used
          required: true
          default: aaa
  ```
- Fix inputs of `workflow_dispatch` are set to `inputs` context as well as `github.event.inputs`. This was added by [the recent change of GitHub Actions](https://github.blog/changelog/2022-06-10-github-actions-inputs-unified-across-manual-and-reusable-workflows/). ([#152](https://github.com/rhysd/actionlint/issues/152))
  ```yaml
  on:
    workflow_dispatch:
      inputs:
        my_input:
          type: string
          required: true
  jobs:
    my_job:
      runs-on: ubuntu-latest
      steps:
        - run: echo ${{ github.event.inputs.my_input }}
        # Now the input is also set to `inputs` context
        - run: echo ${{ inputs.my_input }}
  ```
- Improve that `env` context is now not defined in values of `env:`, `id:` and `uses:`. actionlint now reports usage of `env` context in such places as type errors. ([#158](https://github.com/rhysd/actionlint/issues/158))
  ```yaml
  runs-on: ubuntu-latest
  env:
    FOO: aaa
  steps:
    # ERROR: 'env' context is not defined in values of 'env:', 'id:' and 'uses:'
    - uses: test/${{ env.FOO }}@main
      env:
        BAR: ${{ env.FOO }}
      id: foo-${{ env.FOO }}
  ```
- `actionlint` command gains `-stdin-filename` command line option. When it is specified, the file name is used on reading input from stdin instead of `<stdin>`. ([#157](https://github.com/rhysd/actionlint/issues/157), thanks [@arahatashun](https://github.com/arahatashun))
  ```sh
  # Error message shows foo.yml as file name where the error happened
  ... | actionlint -stdin-filename foo.yml -
  ```
- [The download script](https://github.com/rhysd/actionlint/blob/main/docs/install.md#download-script) allows to specify a directory path to install `actionlint` executable with the second argument of the script. For example, the following command downloads `/path/to/bin/actionlint`:
  ```sh
  # Downloads the latest stable version at `/path/to/bin/actionlint`
  bash <(curl https://raw.githubusercontent.com/rhysd/actionlint/main/scripts/download-actionlint.bash) latest /path/to/bin
  # Downloads actionlint v1.6.14 at `/path/to/bin/actionlint`
  bash <(curl https://raw.githubusercontent.com/rhysd/actionlint/main/scripts/download-actionlint.bash) 1.6.14 /path/to/bin
  ```
- Update popular actions data set including `goreleaser-action@v3`, `setup-python@v4`, `aks-set-context@v3`.
- Update Go dependencies including go-yaml/yaml v3.

[Changes][v1.6.14]


<a name="v1.6.13"></a>
# [v1.6.13](https://github.com/rhysd/actionlint/releases/tag/v1.6.13) - 18 May 2022

- [`secrets: inherit` in reusable workflow](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#onworkflow_callsecretsinherit) is now supported ([#138](https://github.com/rhysd/actionlint/issues/138))
  ```yaml
  on:
    workflow_dispatch:

  jobs:
    pass-secrets-to-workflow:
      uses: ./.github/workflows/called-workflow.yml
      secrets: inherit
  ```
  This means that actionlint cannot know the workflow inherits secrets or not when checking a reusable workflow. To support `secrets: inherit` without giving up on checking `secrets` context, actionlint assumes the followings. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#check-types-of-inputs-and-secrets-in-reusable-workflow) for the details.
  - when `secrets:` is omitted in a reusable workflow, the workflow inherits secrets from a caller
  - when `secrets:` exists in a reusable workflow, the workflow inherits no other secret
- [`macos-12` runner](https://github.blog/changelog/2022-04-25-github-actions-public-beta-of-macos-12-for-github-hosted-runners-is-now-available/) is now supported ([#134](https://github.com/rhysd/actionlint/issues/134), thanks [@shogo82148](https://github.com/shogo82148))
- [`ubuntu-22.04` runner](https://github.blog/changelog/2022-05-10-github-actions-beta-of-ubuntu-22-04-for-github-hosted-runners-is-now-available/) is now supported ([#142](https://github.com/rhysd/actionlint/issues/142), thanks [@shogo82148](https://github.com/shogo82148))
- `concurrency` is available on reusable workflow call ([#136](https://github.com/rhysd/actionlint/issues/136))
  ```yaml
  jobs:
    checks:
      concurrency:
        group: ${{ github.ref }}-${{ github.workflow }}
        cancel-in-progress: true
      uses: ./path/to/workflow.yaml
  ```
- [pre-commit](https://pre-commit.com/) hook now uses a fixed version of actionlint. For example, the following configuration continues to use actionlint v1.6.13 even if v1.6.14 is released. ([#116](https://github.com/rhysd/actionlint/issues/116))
  ```yaml
  repos:
    - repo: https://github.com/rhysd/actionlint
      rev: v1.6.13
      hooks:
        - id: actionlint-docker
  ```
- Update popular actions data set including new versions of `docker/*`, `haskell/actions/setup`,  `actions/setup-go`, ... ([#140](https://github.com/rhysd/actionlint/issues/140), thanks [@bflad](https://github.com/bflad))
- Update Go module dependencies


[Changes][v1.6.13]


<a name="v1.6.12"></a>
# [v1.6.12](https://github.com/rhysd/actionlint/releases/tag/v1.6.12) - 14 Apr 2022

- Fix `secrets.ACTIONS_RUNNER_DEBUG` and `secrets.ACTIONS_STEP_DEBUG` are not pre-defined in a reusable workflow. ([#130](https://github.com/rhysd/actionlint/issues/130))
- Fix checking permissions is outdated. `pages` and `discussions` permissions were added and `metadata` permission was removed. ([#131](https://github.com/rhysd/actionlint/issues/131), thanks [@suzuki-shunsuke](https://github.com/suzuki-shunsuke))
- Disable [SC2157](https://github.com/koalaman/shellcheck/wiki/SC2157) shellcheck rule to avoid a false positive due to [the replacement of `${{ }}`](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#shellcheck-integration-for-run) in script. For example, in the below script `-z ${{ env.FOO }}` was replaced with `-z ______________` and it caused 'always false due to literal strings' error. ([#113](https://github.com/rhysd/actionlint/issues/113))
  ```yaml
  - run: |
      if [[ -z ${{ env.FOO }} ]]; then
        echo "FOO is empty"
      fi
  ```
- Add codecov-action@v3 to popular actions data set.

[Changes][v1.6.12]


<a name="v1.6.11"></a>
# [v1.6.11](https://github.com/rhysd/actionlint/releases/tag/v1.6.11) - 05 Apr 2022

- Fix crash on making [outputs in JSON format](https://github.com/rhysd/actionlint/blob/main/docs/usage.md#format-error-messages) with `actionlint -format '{{json .}}'`. ([#128](https://github.com/rhysd/actionlint/issues/128))
- Allow any outputs from `actions/github-script` action because it allows to set arbitrary outputs via calling `core.setOutput()` in JavaScript. ([#104](https://github.com/rhysd/actionlint/issues/104))
  ```yaml
  - id: test
    uses: actions/github-script@v5
    with:
      script: |
        core.setOutput('answer', 42);
  - run: |
      echo "The answer is ${{ steps.test.outputs.answer }}"
  ```
- Add support for Go 1.18. All released binaries were built with Go 1.18 compiler. The bottom supported version is Go 1.16 and it's not been changed.
- Update popular actions data set (`actions/cache`, `code-ql-actions/*`, ...)
- Update some Go module dependencies

[Changes][v1.6.11]


<a name="v1.6.10"></a>
# [v1.6.10](https://github.com/rhysd/actionlint/releases/tag/v1.6.10) - 11 Mar 2022

- Support outputs in reusable workflow call. See [the official document](https://docs.github.com/en/actions/using-workflows/reusing-workflows#using-outputs-from-a-reusable-workflow) for the usage of the outputs syntax. ([#119](https://github.com/rhysd/actionlint/issues/119), [#121](https://github.com/rhysd/actionlint/issues/121))
  Example of reusable workflow definition:
  ```yaml
  on:
    workflow_call:
      outputs:
        some_output:
          description: "Some awesome output"
          value: 'result value of workflow call'
  jobs:
    job:
      runs-on: ubuntu-latest
      steps:
        ...
  ```
  Example of reusable workflow call:
  ```yaml
  jobs:
    job1:
      uses: ./.github/workflows/some_workflow.yml
    job2:
      runs-on: ubuntu-latest
      needs: job1
      steps:
        - run: echo ${{ needs.job1.outputs.some_output }}
  ```
- Support checking `jobs` context, which is only available in `on.workflow_call.outputs.<name>.value`. Outputs of jobs can be referred via the context. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#check-outputs-of-reusable-workflow) for more details.
  ```yaml
  on:
    workflow_call:
      outputs:
        image-version:
          description: "Docker image version"
          # ERROR: 'imagetag' does not exist (typo of 'image_tag')
          value: ${{ jobs.gen-image-version.outputs.imagetag }}
  jobs:
    gen-image-version:
      runs-on: ubuntu-latest
      outputs:
        image_tag: "${{ steps.get_tag.outputs.tag }}"
      steps:
        - run: ./output_image_tag.sh
          id: get_tag
  ```
- Add new major releases in `actions/*` actions including `actions/checkout@v3`, `actions/setup-go@v3`, `actions/setup-python@v3`, ...
- Check job IDs. They must start with a letter or `_` and contain only alphanumeric characters, `-` or `_`. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#job-id-naming-convention) for more details. ([#80](https://github.com/rhysd/actionlint/issues/80))
  ```yaml
  on: push
  jobs:
    # ERROR: '.' cannot be contained in job ID
    foo-v1.2.3:
      runs-on: ubuntu-latest
      steps:
        - run: 'job ID with version'
  ```
- Fix `windows-latest` now means `windows-2022` runner. See [virtual-environments#4856](https://github.com/actions/virtual-environments/issues/4856) for the details. ([#120](https://github.com/rhysd/actionlint/issues/120))
- Update [the playground](https://rhysd.github.io/actionlint/) dependencies to the latest.
- Update Go module dependencies

[Changes][v1.6.10]


<a name="v1.6.9"></a>
# [v1.6.9](https://github.com/rhysd/actionlint/releases/tag/v1.6.9) - 24 Feb 2022

- Support [`runner.arch` context value](https://docs.github.com/en/actions/learn-github-actions/contexts#runner-context). (thanks [@shogo82148](https://github.com/shogo82148), [#101](https://github.com/rhysd/actionlint/issues/101))
  ```yaml
  steps:
    - run: ./do_something_64bit.sh
      if: ${{ runner.arch == 'x64' }}
  ```
- Support [calling reusable workflows in local directories](https://docs.github.com/en/actions/using-workflows/reusing-workflows#calling-a-reusable-workflow). (thanks [@jsok](https://github.com/jsok), [#107](https://github.com/rhysd/actionlint/issues/107))
  ```yaml
  jobs:
    call-workflow-in-local-repo:
      uses: ./.github/workflows/useful_workflow.yml
  ```
- Add [a document](https://github.com/rhysd/actionlint/blob/main/docs/install.md#asdf) to install actionlint via [asdf](https://asdf-vm.com/) version manager. (thanks [@crazy-matt](https://github.com/crazy-matt), [#99](https://github.com/rhysd/actionlint/issues/99))
- Fix using `secrets.GITHUB_TOKEN` caused a type error when some other secret is defined. (thanks [@mkj-is](https://github.com/mkj-is), [#106](https://github.com/rhysd/actionlint/issues/106))
- Fix nil check is missing on parsing `uses:` step. (thanks [@shogo82148](https://github.com/shogo82148), [#102](https://github.com/rhysd/actionlint/issues/102))
- Fix some documents including broken links. (thanks [@ohkinozomu](https://github.com/ohkinozomu), [#105](https://github.com/rhysd/actionlint/issues/105))
- Update popular actions data set to the latest. More arguments are added to many actions. And a few actions had new major versions.
- Update webhook payload data set to the latest. `requested_action` type was added to `check_run` hook. `requested` and `rerequested` types were removed from `check_suite` hook. `updated` type was removed from `project` hook.


[Changes][v1.6.9]


<a name="v1.6.8"></a>
# [v1.6.8](https://github.com/rhysd/actionlint/releases/tag/v1.6.8) - 15 Nov 2021

- [Untrusted inputs](https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions) detection can detect untrusted inputs in object filter syntax. For example, `github.event.*.body` filters `body` properties and it includes the untrusted input `github.event.comment.body`. actionlint detects such filters and causes an error. The error message includes all untrusted input names which are filtered by the object filter so that you can know what inputs are untrusted easily. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#untrusted-inputs) for more details.
  Input example:
  ```yaml
  - name: Get comments
    run: echo '${{ toJSON(github.event.*.body) }}'
  ```
  Error message:
  ```
  object filter extracts potentially untrusted properties "github.event.comment.body", "github.event.discussion.body", "github.event.issue.body", ...
  ```
  Instead you should do:
  ```yaml
  - name: Get comments
    run: echo "$JSON"
    env:
      JSON: {{ toJSON(github.event.*.body) }}
  ```
- Support [the new input type syntax for `workflow_dispatch` event](https://github.blog/changelog/2021-11-10-github-actions-input-types-for-manual-workflows/), which was introduced recently. You can declare types of inputs on triggering a workflow manually. actionlint does two things with this new syntax.
  - actionlint checks the syntax. Unknown input types, invalid default values, missing options for 'choice' type.
    ```yaml
    inputs:
      # Unknown input type
      id:
        type: number
      # ERROR: No options for 'choice' input type
      kind:
        type: choice
      name:
        type: choice
        options:
          - Tama
          - Mike
        # ERROR: Default value is not in options
        default: Chobi
      verbose:
        type: boolean
        # ERROR: Boolean value must be 'true' or 'false'
        default: yes
    ```
  - actionlint give a strict object type to `github.event.inputs` so that a type checker can check unknown input names and type mismatches on using the value.
    ```yaml
    on:
      workflow_dispatch:
        inputs:
          message:
            type: string
          verbose:
            type: boolean
    # Type of `github.event.inputs` is {"message": string; "verbose": bool}
    jobs:
      test:
        runs-on: ubuntu-latest
        steps:
          # ERROR: Undefined input
          - run: echo "${{ github.event.inputs.massage }}"
          # ERROR: Bool value is not available for object key
          - run: echo "${{ env[github.event.inputs.verbose] }}"
    ```
  - See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#check-workflow-dispatch-events) for more details.
- Add missing properties in `github` context. See [the contexts document](https://docs.github.com/en/actions/learn-github-actions/contexts#github-context) to know the full list of properties.
  - `github.ref_name` (thanks [@dihmandrake](https://github.com/dihmandrake), [#72](https://github.com/rhysd/actionlint/issues/72))
  - `github.ref_protected`
  - `github.ref_type`
- Filtered array by object filters is typed more strictly.
  ```
  # `env` is a map object { string => string }
  # Previously typed as array<any> now it is typed as array<string>
  env.*
  ```
- Update Go module dependencies and playground dependencies.

[Changes][v1.6.8]


<a name="v1.6.7"></a>
# [v1.6.7](https://github.com/rhysd/actionlint/releases/tag/v1.6.7) - 08 Nov 2021

- Fix missing property `name` in `runner` context object (thanks [@ioanrogers](https://github.com/ioanrogers), [#67](https://github.com/rhysd/actionlint/issues/67)).
- Fix a false positive on type checking at `x.*` object filtering syntax where the receiver is an object. actionlint previously only allowed arrays as receiver of object filtering ([#66](https://github.com/rhysd/actionlint/issues/66)).
  ```ruby
  fromJSON('{"a": "from a", "b": "from b"}').*
  # => ["from a", "from b"]

  fromJSON('{"a": {"x": "from a.x"}, "b": {"x": "from b.x"}}').*.x
  # => ["from a.x", "from b.x"]
  ```
- Add [rust-cache](https://github.com/Swatinem/rust-cache) as new popular action.
- Remove `bottle: unneeded` from Homebrew formula (thanks [@oppara](https://github.com/oppara), [#63](https://github.com/rhysd/actionlint/issues/63)).
- Support `branch_protection_rule` webhook again.
- Update popular actions data set to the latest ([#64](https://github.com/rhysd/actionlint/issues/64), [#70](https://github.com/rhysd/actionlint/issues/70)).

[Changes][v1.6.7]


<a name="v1.6.6"></a>
# [v1.6.6](https://github.com/rhysd/actionlint/releases/tag/v1.6.6) - 17 Oct 2021

- `inputs` and `secrets` objects are now typed looking at `workflow_call` event at `on:`. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#check-types-of-inputs-and-secrets-in-reusable-workflow) for more details.
  - `inputs` object is typed with definitions at `on.workflow_call.inputs`. When the workflow is not callable, it is typed at `{}` (empty object) so any `inputs.*` access causes a type error.
  - `secrets` object is typed with definitions at `on.workflow_call.secrets`.
  ```yaml
  on:
    workflow_call:
      # `inputs` object is typed {url: string; lucky_number: number}
      inputs:
        url:
          description: 'your URL'
          type: string
        lucky_number:
          description: 'your lucky number'
          type: number
      # `secrets` object is typed {user: string; credential: string}
      secrets:
        user:
          description: 'your user name'
        credential:
          description: 'your credential'
  jobs:
    test:
      runs-on: ubuntu-20.04
      steps:
        - name: Send data
          # ERROR: uri is typo of url
          run: curl ${{ inputs.uri }} -d ${{ inputs.lucky_number }}
          env:
            # ERROR: credentials is typo of credential
            TOKEN: ${{ secrets.credentials }}
  ```
- `id-token` is added to permissions (thanks [@cmmarslender](https://github.com/cmmarslender), [#62](https://github.com/rhysd/actionlint/issues/62)).
- Report an error on nested workflow calls since it is [not allowed](https://docs.github.com/en/actions/learn-github-actions/reusing-workflows#limitations).
  ```yaml
  on:
    # This workflow is reusable
    workflow_call:

  jobs:
    test:
      # ERROR: Nested workflow call is not allowed
      uses: owner/repo/path/to/workflow.yml@ref
  ```
- Parse `uses:` at reusable workflow call more strictly following `{owner}/{repo}/{path}@{ref}` format.
- Popular actions data set was updated to the latest ([#61](https://github.com/rhysd/actionlint/issues/61)).
- Dependencies of playground were updated to the latest (including eslint v8).

[Changes][v1.6.6]


<a name="v1.6.5"></a>
# [v1.6.5](https://github.com/rhysd/actionlint/releases/tag/v1.6.5) - 08 Oct 2021

- Support [reusable workflows](https://docs.github.com/en/actions/learn-github-actions/reusing-workflows) syntax which is now in beta. Only very basic syntax checks are supported at this time. Please see [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#check-reusable-workflows) to know checks for reusable workflow syntax.
  - Example of `workflow_call` event
    ```yaml
    on:
      workflow_call:
        inputs:
          name:
            description: your name
            type: string
        secrets:
          token:
            required: true

    jobs:
      ...
    ```
  - Example of reusable workflow call with `uses:` at `job.<job_id>`
    ```yaml
    on: ...
    jobs:
      hello:
        uses: owner/repo/path/to/workflow.yml@main
        with:
          name: Octocat
        secrets:
          token: ${{ secrets.token }}
    ```
- Support `github.run_attempt` property in `${{ }}` expression ([#57](https://github.com/rhysd/actionlint/issues/57)).
- Add support for `windows-2022` runner which is now in [public beta](https://github.com/actions/virtual-environments/issues/3949).
- Remove support for `ubuntu-16.04` runner which was [removed from GitHub Actions at the end of September](https://github.com/actions/virtual-environments/issues/3287).
- Ignore [SC2154](https://github.com/koalaman/shellcheck/wiki/SC2154) shellcheck rule which can cause false positive ([#53](https://github.com/rhysd/actionlint/issues/53)).
- Fix error position was not correct when required keys are not existing in job configuration.
- Update popular actions data set. New major versions of github-script and lock-threads actions are supported ([#59](https://github.com/rhysd/actionlint/issues/59)).
- Fix document (thanks [@fornwall](https://github.com/fornwall) at [#52](https://github.com/rhysd/actionlint/issues/52), thanks [@equal-l2](https://github.com/equal-l2) at [#56](https://github.com/rhysd/actionlint/issues/56)).
  - Now actionlint is [an official package of Homebrew](https://formulae.brew.sh/formula/actionlint). Simply executing `brew install actionlint` can install actionlint.

[Changes][v1.6.5]


<a name="v1.6.4"></a>
# [v1.6.4](https://github.com/rhysd/actionlint/releases/tag/v1.6.4) - 21 Sep 2021

- Implement 'map' object types `{ string => T }`, where all properties of the object are typed as `T`. Since a key of object is always string, left hand side of `=>` is fixed to `string`. For example, `env` context only has string properties so it is typed as `{ string => string}`. Previously its properties were typed `any`.
  ```yaml
  # typed as string (previously any)
  env.FOO

  # typed as { id: string; network: string; ports: object; } (previously any)
  job.services.redis
  ```
- `github.event.discussion.title` and `github.event.discussion.body` are now checked as untrusted inputs.
- Update popular actions data set. ([#50](https://github.com/rhysd/actionlint/issues/50), [#51](https://github.com/rhysd/actionlint/issues/51))
- Update webhooks payload data set. `branch_protection_rule` hook was dropped from the list due to [github/docs@179a6d3](https://github.com/github/docs/commit/179a6d334e92b9ade8626ef42a546dae66b49951). ([#50](https://github.com/rhysd/actionlint/issues/50), [#51](https://github.com/rhysd/actionlint/issues/51))

[Changes][v1.6.4]


<a name="v1.6.3"></a>
# [v1.6.3](https://github.com/rhysd/actionlint/releases/tag/v1.6.3) - 04 Sep 2021

- Improve guessing a type of matrix value. When a matrix contains numbers and strings, previously the type fell back to `any`. Now it is deduced as string.
  ```yaml
  strategy:
    matrix:
      # matrix.node is now deduced as `string` instead of `any`
      node: [14, 'latest']
  ```
- Fix types of `||` and `&&` expressions. Previously they were typed as `bool` but it was not correct. Correct type is sum of types of both sides of the operator like TypeScript. For example, type of `'foo' || 'bar'` is a string, and `github.event && matrix` is an object.
- actionlint no longer reports an error when a local action does not exist in the repository. It is a popular pattern that a local action directory is cloned while a workflow running. ([#25](https://github.com/rhysd/actionlint/issues/25), [#40](https://github.com/rhysd/actionlint/issues/40))
- Disable [SC2050](https://github.com/koalaman/shellcheck/wiki/SC2050) shellcheck rule since it causes some false positive. ([#45](https://github.com/rhysd/actionlint/issues/45))
- Fix `-version` did not work when running actionlint via [the Docker image](https://github.com/rhysd/actionlint/blob/main/docs/usage.md#docker) ([#47](https://github.com/rhysd/actionlint/issues/47)).
- Fix pre-commit hook file name. (thanks [@xsc27](https://github.com/xsc27), [#38](https://github.com/rhysd/actionlint/issues/38))
- [New `branch_protection_rule` event](https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads#branch_protection_rule) is supported. ([#48](https://github.com/rhysd/actionlint/issues/48))
- Update popular actions data set. ([#41](https://github.com/rhysd/actionlint/issues/41), [#48](https://github.com/rhysd/actionlint/issues/48))
- Update Go library dependencies.
- Update playground dependencies.

[Changes][v1.6.3]


<a name="v1.6.2"></a>
# [v1.6.2](https://github.com/rhysd/actionlint/releases/tag/v1.6.2) - 23 Aug 2021

- actionlint now checks evaluated values at `${{ }}` are not an object nor an array since they are not useful. See [the check document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#check-type-check-expression) for more details.
```yaml
# ERROR: This will always be replaced with `echo 'Object'`
- run: echo '${{ runner }}'
# OK: Serialize an object into JSON to check the content
- run: echo '${{ toJSON(runner) }}'
```
- Add [pre-commit](https://pre-commit.com/) support. pre-commit is a framework for managing Git `pre-commit` hooks. See [the usage document](https://github.com/rhysd/actionlint/blob/main/docs/usage.md#pre-commit) for more details. (thanks [@xsc27](https://github.com/xsc27) for adding the integration at [#33](https://github.com/rhysd/actionlint/issues/33)) ([#23](https://github.com/rhysd/actionlint/issues/23))
- Add [an official Docker image](https://hub.docker.com/repository/docker/rhysd/actionlint). The Docker image contains shellcheck and pyflakes as dependencies. Now actionlint can be run with `docker run` command easily. See [the usage document](https://github.com/rhysd/actionlint/blob/main/docs/usage.md#docker) for more details. (thanks [@xsc27](https://github.com/xsc27) for the help at [#34](https://github.com/rhysd/actionlint/issues/34))
```sh
docker run --rm -v $(pwd):/repo --workdir /repo rhysd/actionlint:latest -color
```
- Go 1.17 is now a default compiler to build actionlint. Built binaries are faster than before by 2~7% when the process is CPU-bound. Sizes of built binaries are about 2% smaller. Note that Go 1.16 continues to be supported.
- `windows/arm64` target is added to released binaries thanks to Go 1.17.
- Now any value can be converted into bool implicitly. Previously this was not permitted as actionlint provides stricter type check. However it is not useful that a condition like `if: github.event.foo` causes a type error.
- Fix a prefix operator cannot be applied repeatedly like `!!42`.
- Fix a potential crash when type checking on expanding an object with `${{ }}` like `matrix: ${{ fromJSON(env.FOO) }}`
- Update popular actions data set ([#36](https://github.com/rhysd/actionlint/issues/36))

[Changes][v1.6.2]


<a name="v1.6.1"></a>
# [v1.6.1](https://github.com/rhysd/actionlint/releases/tag/v1.6.1) - 16 Aug 2021

- [Problem Matchers](https://github.com/actions/toolkit/blob/master/docs/problem-matchers.md) is now officially supported by actionlint, which annotates errors from actionlint on GitHub as follows. The matcher definition is maintained at [`.github/actionlint-matcher.json`](https://github.com/rhysd/actionlint/blob/main/.github/actionlint-matcher.json) by [script](https://github.com/rhysd/actionlint/tree/main/scripts/generate-actionlint-matcher). For the usage, see [the document](https://github.com/rhysd/actionlint/blob/main/docs/usage.md#problem-matchers).

<img src="https://github.com/rhysd/ss/blob/master/actionlint/problem-matcher.png?raw=true" alt="annotation by Problem Matchers" width="715" height="221"/>

- `runner_label` rule now checks conflicts in labels at `runs-on`. For example, there is no runner which meats both `ubuntu-latest` and `windows-latest`. This kind of misconfiguration sometimes happen when a beginner misunderstands the usage of `runs-on:`. To run a job on each runners, `matrix:` should be used. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#check-runner-labels) for more information.

```yaml
on: push
jobs:
  test:
    # These labels match to no runner
    runs-on: [ubuntu-latest, windows-latest]
    steps:
      - run: echo ...
```

- Reduce memory footprint (around 16%) on starting `actionlint` command by removing unnecessary data from `PopularActions` global variable. This also slightly reduces binary size (about 3.7% at `playground/main.wasm`).
- Fix accessing `steps.*` objects in job's `environment:` configuration caused a type error ([#30](https://github.com/rhysd/actionlint/issues/30)).
- Fix checking that action's input names at `with:` were not in case insensitive ([#31](https://github.com/rhysd/actionlint/issues/31)).
- Ignore outputs of [getsentry/paths-filter](https://github.com/getsentry/paths-filter). It is a fork of [dorny/paths-filter](https://github.com/dorny/paths-filter). actionlint cannot check the outputs statically because it sets outputs dynamically.
- Add [Azure/functions-action](https://github.com/Azure/functions-action) to popular actions.
- Update popular actions data set ([#29](https://github.com/rhysd/actionlint/issues/29)).

[Changes][v1.6.1]


<a name="v1.6.0"></a>
# [v1.6.0](https://github.com/rhysd/actionlint/releases/tag/v1.6.0) - 11 Aug 2021

- Check potentially untrusted inputs to prevent [a script injection vulnerability](https://securitylab.github.com/research/github-actions-untrusted-input/) at `run:` and `script` input of [actions/github-script](https://github.com/actions/github-script). See [the rule document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#untrusted-inputs) for more explanations and workflow example. (thanks [@azu](https://github.com/azu) for the feature request at [#19](https://github.com/rhysd/actionlint/issues/19))

Incorrect code

```yaml
- run: echo '${{ github.event.pull_request.title }}'
```

should be replaced with

```yaml
- run: echo "issue ${TITLE}"
  env:
    TITLE: ${{github.event.issue.title}}
```

- Add `-format` option to `actionlint` command. It allows to flexibly format error messages as you like with [Go template syntax](https://pkg.go.dev/text/template). See [the usage document](https://github.com/rhysd/actionlint/blob/main/docs/usage.md#format) for more details. (thanks [@ybiquitous](https://github.com/ybiquitous) for the feature request at [#20](https://github.com/rhysd/actionlint/issues/20))

Simple example to output error messages as JSON:

```sh
actionlint -format '{{json .}}'
```

More compliated example to output error messages as markdown:

```sh
actionlint -format '{{range $ := .}}### Error at line {{$.Line}}, col {{$.Column}} of `{{$.Filepath}}`\n\n{{$.Message}}\n\n```\n{{$.Snippet}}\n```\n\n{{end}}'
```

- Documents are reorganized. Long `README.md` is separated into several document files ([#28](https://github.com/rhysd/actionlint/issues/28))
  - [`README.md`](https://github.com/rhysd/actionlint/blob/main/README.md): Introduction, Quick start, Document links
  - [`docs/checks.md`](https://github.com/rhysd/actionlint/tree/main/docs/checks.md): Full list of all checks done by actionlint with example inputs, outputs, and playground links
  - [`docs/install.md`](https://github.com/rhysd/actionlint/tree/main/docs/install.md): Installation instruction
  - [`docs/usage.md`](https://github.com/rhysd/actionlint/tree/main/docs/usage.md): Advanced usage of `actionlint` command, usage of playground, integration with [reviewdog](https://github.com/reviewdog/reviewdog), [Problem Matchers](https://github.com/actions/toolkit/blob/master/docs/problem-matchers.md), [super-linter](https://github.com/github/super-linter)
  - [`docs/config.md`](https://github.com/rhysd/actionlint/tree/main/docs/config.md): About configuration file
  - [`doc/api.md`](https://github.com/rhysd/actionlint/tree/main/docs/api.md): Using actionlint as Go library
  - [`doc/reference.md`](https://github.com/rhysd/actionlint/tree/main/docs/reference.md): Links to resources
- Fix checking shell names was not case-insensitive, for example `PowerShell` was detected as invalid shell name
- Update popular actions data set to the latest
- Make lexer errors on checking `${{ }}` expressions more meaningful

[Changes][v1.6.0]


<a name="v1.5.3"></a>
# [v1.5.3](https://github.com/rhysd/actionlint/releases/tag/v1.5.3) - 04 Aug 2021

- Now actionlint allows to use any operators outside `${{ }}` on `if:` condition like `if: github.repository_owner == 'rhysd'` ([#22](https://github.com/rhysd/actionlint/issues/22)). [The official document](https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idif) said that using any operator outside `${{ }}` was invalid even if it was on `if:` condition. However, [github/docs#8786](https://github.com/github/docs/pull/8786) clarified that the document was not correct.

[Changes][v1.5.3]


<a name="v1.5.2"></a>
# [v1.5.2](https://github.com/rhysd/actionlint/releases/tag/v1.5.2) - 02 Aug 2021

- Outputs of [dorny/paths-filter](https://github.com/dorny/paths-filter) are now not typed strictly because the action dynamically sets outputs which are not defined in its `action.yml`. actionlint cannot check such outputs statically ([#18](https://github.com/rhysd/actionlint/issues/18)).
- [The table](https://github.com/rhysd/actionlint/blob/main/all_webhooks.go) for checking [Webhooks supported by GitHub Actions](https://docs.github.com/en/actions/reference/events-that-trigger-workflows#webhook-events) is now generated from the official document automatically with [script](https://github.com/rhysd/actionlint/tree/main/scripts/generate-webhook-events). The table continues to be updated weekly by [the CI workflow](https://github.com/rhysd/actionlint/actions/workflows/generate.yaml).
- Improve error messages while lexing expressions as follows.
- Fix column numbers are off-by-one on some lexer errors.
- Fix checking invalid numbers where some digit follows zero in a hex number (e.g. `0x01`) or an exponent part of number (e.g. `1e0123`).
- Fix a parse error message when some tokens still remain after parsing finishes.
- Refactor the expression lexer to lex an input incrementally. It slightly reduces memory consumption.

Lex error until v1.5.1:

```test.yaml:9:26: got unexpected character '+' while lexing expression, expecting '_', '\'', '}', '(', ')', '[', ']', '.', '!', '<', '>', '=', '&', '|', '*', ',', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z' [expression]```

Lex error from v1.5.2:

```test.yaml:9:26: got unexpected character '+' while lexing expression, expecting 'a'..'z', 'A'..'Z', '0'..'9', ''', '}', '(', ')', '[', ']', '.', '!', '<', '>', '=', '&', '|', '*', ',', '_' [expression]```

[Changes][v1.5.2]


<a name="v1.5.1"></a>
# [v1.5.1](https://github.com/rhysd/actionlint/releases/tag/v1.5.1) - 29 Jul 2021

- Improve checking the intervals of scheduled events ([#14](https://github.com/rhysd/actionlint/issues/14), [#15](https://github.com/rhysd/actionlint/issues/15)). Since GitHub Actions [limits the interval to once every 5 minutes](https://github.blog/changelog/2019-11-01-github-actions-scheduled-jobs-maximum-frequency-is-changing/), actionlint now reports an error when a workflow is configured to be run once per less than 5 minutes.
- Skip checking inputs of [octokit/request-action](https://github.com/octokit/request-action) since it allows to specify arbitrary inputs though they are not defined in its `action.yml` ([#16](https://github.com/rhysd/actionlint/issues/16)).
  - Outputs of the action are still be typed strictly. Only its inputs are not checked.
- The help text of `actionlint` is now hosted online: https://rhysd.github.io/actionlint/usage.html
- Add new fuzzing target for parsing glob patterns.

[Changes][v1.5.1]


<a name="v1.5.0"></a>
# [v1.5.0](https://github.com/rhysd/actionlint/releases/tag/v1.5.0) - 26 Jul 2021

- `action` rule now validates inputs of popular actions at `with:`. When a required input is not specified or an undefined input is specified, actionlint will report it.
  - Popular actions are updated automatically once a week and the data set is embedded to executable directly. The check does not need any network request and does not affect performance of actionlint. Sources of the actions are listed [here](https://github.com/rhysd/actionlint/blob/main/scripts/generate-popular-actions/main.go#L51). If you have some request to support new action, please report it at [the issue form](https://github.com/rhysd/actionlint/issues/new).
  - Please see [the document](https://github.com/rhysd/actionlint#check-popular-action-inputs) for example ([Playground](https://rhysd.github.io/actionlint#eJyFj0EKwjAQRfc9xV8I1UJbcJmVK+8xDYOpqUlwEkVq725apYgbV8PMe/Dne6cQkpiiOPtOVAFEljhP4Jqc1D4LqUsupnqgmS1IIgd5W0CNJCwKpGPvnbSatOHDbf/BwL2PRq0bYPmR9efXBdiMIwyJOfYDy7asqrZqBq9tucM0/TWXyF81UI5F0wbSlk4s67u5mMKFLL8A+h9EEw==)).
- `expression` rule now types outputs of popular actions (type of `steps.{id}.outputs` object) more strictly.
  - For example, `actions/cache@v2` sets `cache-hit` output. The outputs object is typed as `{ cache-hit: any }`. Previously it was typed as `any` which means no further type check was performed.
  - Please see the second example of [the document](https://github.com/rhysd/actionlint#check-contextual-step-object) ([Playground](https://rhysd.github.io/actionlint#eJyNTksKwjAQ3fcUbyFUC0nBZVauvIakMZjY0gRnokjp3W3TUl26Gt53XugVYiJXFPfQkCoAtsTzBR6pJxEmQ2pSz0l0etayRGwjLS5AIJElBW3Yh55qo42zp+dxlQF/Vcjkxrw8O7UhoLVvhd0wwGlyZ99Z2pdVVVeyC6YtDxjHH3PUUxiyjtq0+mZpmzENVrDGhVyVN8r8V4bEMfGKhPP8bfw7dlliH1xHWso=)).
- Outputs of local actions (their names start with `./`) are also typed more strictly as well as popular actions.
- Metadata (`action.yml`) of local actions are now cached to avoid reading and parsing `action.yml` files repeatedly for the same action.
- Add new rule `permissions` to check [permission scopes](https://docs.github.com/en/actions/reference/authentication-in-a-workflow#permissions-for-the-github_token) for default `secrets.GITHUB_TOKEN`. Please see [the document](https://github.com/rhysd/actionlint#permissions) for more details ([Playground](https://rhysd.github.io/actionlint/#eJxNjd0NwyAMhN89xS3AAmwDxBK0FCOMlfUDiVr16aTv/qR5dNNM1Hl8imqRph7nKJOJXhLVEzBZ51ZgWFMnq2TR2jRXw/Zu63/gBkDKnN7ftQethPF6GByOEOuDdXL/ldw+8eCUBZlrlQvntjLp)).
- Structure of [`actionlint.Permissions`](https://pkg.go.dev/github.com/rhysd/actionlint#Permissions) struct was changed. A parser no longer checks values of `permissions:` configuration. The check is now done by `permissions` rule.

[Changes][v1.5.0]


<a name="v1.4.3"></a>
# [v1.4.3](https://github.com/rhysd/actionlint/releases/tag/v1.4.3) - 21 Jul 2021

- Support new Webhook events [`discussion` and `discussion_comment`](https://docs.github.com/en/actions/reference/events-that-trigger-workflows#discussion) ([#8](https://github.com/rhysd/actionlint/issues/8)).
- Read file concurrently with limiting concurrency to number of CPUs. This improves performance when checking many files and disabling shellcheck/pyflakes integration.
- Support Linux based on musl libc by [the download script](https://github.com/rhysd/actionlint/blob/main/scripts/download-actionlint.bash) ([#5](https://github.com/rhysd/actionlint/issues/5)).
- Reduce number of goroutines created while running shellcheck/pyflakes processes. This has small impact on memory usage when your workflows have many `run:` steps.
- Reduce built binary size by splitting an external library which is only used for debugging into a separate command line tool.
- Introduce several micro benchmark suites to track performance.
- Enable code scanning for Go/TypeScript/JavaScript sources in actionlint repository.

[Changes][v1.4.3]


<a name="v1.4.2"></a>
# [v1.4.2](https://github.com/rhysd/actionlint/releases/tag/v1.4.2) - 16 Jul 2021

- Fix executables in the current directory may be used unexpectedly to run `shellcheck` or `pyflakes` on Windows. This behavior could be security vulnerability since an attacker might put malicious executables in shared directories. actionlint searched an executable with [`exec.LookPath`](https://pkg.go.dev/os/exec#LookPath), but it searched the current directory on Windows as [golang/go#43724](https://github.com/golang/go/issues/43724) pointed. Now actionlint uses [`execabs.LookPath`](https://pkg.go.dev/golang.org/x/sys/execabs#LookPath) instead, which does not have the issue. (ref: [sharkdp/bat#1724](https://github.com/sharkdp/bat/pull/1724))
- Fix issue caused by running so many processes concurrently. Since checking workflows by actionlint is highly parallelized, checking many workflow files makes too many `shellcheck` processes and opens many files in parallel. This hit OS resources limitation (issue [#3](https://github.com/rhysd/actionlint/issues/3)). Now reading files is serialized and number of processes run concurrently is limited for fixing the issue. Note that checking workflows is still done in parallel so this fix does not affect actionlint's performance.
- Ensure cleanup processes even if actionlint stops due to some fatal issue while visiting a workflow tree.
- Improve fatal error message to know which workflow file caused the error.
- [Playground](https://rhysd.github.io/actionlint/) improvements
  - "Permalink" button was added to make permalink directly linked to the current workflow source code. The source code is embedded in hash of the URL.
  - "Check" button and URL input form was added to check workflow files on https://github.com or https://gist.github.com easily. Visit a workflow file on GitHub, copy the URL, paste it to the input form and click the button. It instantly fetches the workflow file content and checks it with actionlint.
  - `u=` URL parameter was added to specify GitHub or Gist URL like https://rhysd.github.io/actionlint/?u=https://github.com/rhysd/actionlint/blob/main/.github/workflows/ci.yaml

[Changes][v1.4.2]


<a name="v1.4.1"></a>
# [v1.4.1](https://github.com/rhysd/actionlint/releases/tag/v1.4.1) - 12 Jul 2021

- A pre-built executable for `darwin/arm64` (Apple M1) was added to CI ([#1](https://github.com/rhysd/actionlint/issues/1))
  - Managing `actionlint` command with Homebrew on M1 Mac is now available. See [the instruction](https://github.com/rhysd/actionlint#homebrew-on-macos) for more details
  - Since the author doesn't have M1 Mac and GitHub Actions does not support M1 Mac yet, the built binary is not tested
- Pre-built executables are now built with Go 1.16 compiler (previously it was 1.15)
- Fix error message is sometimes not in one line when the error message was caused by go-yaml/yaml parser
- Fix playground does not work on Safari browsers on both iOS and Mac since they don't support `WebAssembly.instantiateStreaming()` yet
- Make URLs in error messages clickable on playground
- Code base of playground was migrated from JavaScript to Typescript along with improving error handlings

[Changes][v1.4.1]


<a name="v1.4.0"></a>
# [v1.4.0](https://github.com/rhysd/actionlint/releases/tag/v1.4.0) - 09 Jul 2021

- New rule to validate [glob pattern syntax](https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#filter-pattern-cheat-sheet) to filter branches, tags and paths. For more details, see [documentation](https://github.com/rhysd/actionlint#check-glob-pattern).
  - syntax errors like missing closing brackets for character range `[..]`
  - invalid usage like `?` following `*`, invalid character range `[9-1]`, ...
  - invalid character usage for Git ref names (branch name, tag name)
    - ref name cannot start/end with `/`
    - ref name cannot contain `[`, `:`, `\`, ...
- Fix column of error position is off by one when the error is caused by quoted strings like `'...'` or `"..."`.
- Add `--norc` option to `shellcheck` command to check shell scripts in `run:` in order not to be affected by any user configuration.
- Improve some error messages
- Explain playground in `man` manual

[Changes][v1.4.0]


<a name="v1.3.2"></a>
# [v1.3.2](https://github.com/rhysd/actionlint/releases/tag/v1.3.2) - 04 Jul 2021

- [actionlint playground](https://rhysd.github.io/actionlint) was implemented thanks to WebAssembly. actionlint is now available on browser without installing anything. The playground does not send user's workflow content to any remote server.
- Some margins are added to code snippets in error message. See below examples. I believe it's easier to recognize code in bunch of error messages than before.
- Line number is parsed from YAML syntax error. Since errors from [go-yaml/go](https://github.com/go-yaml/yaml) don't have position information, previously YAML syntax errors are reported at line:0, col:0. Now line number is parsed from error message and set correctly (if error message includes line number).
- Code snippet is shown in error message even if column number of the error position is unknown.
- Fix error message on detecting duplicate of step IDs.
- Fix and improve validating arguments of `format()` calls.
- All rule documents have links to actionlint playground with example code.
- `man` manual covers usage of actionlint on CI services.

Error message until v1.3.1:

```
test.yaml:4:13: invalid CRON format "0 */3 * *" in schedule event: Expected exactly 5 fields, found 4: 0 */3 * * [events]
4|     - cron: '0 */3 * *'
 |             ^~
```

Error message at v1.3.2:

```
test.yaml:4:13: invalid CRON format "0 */3 * *" in schedule event: Expected exactly 5 fields, found 4: 0 */3 * * [events]
  |
4 |     - cron: '0 */3 * *'
  |             ^~
```

[Changes][v1.3.2]


<a name="v1.3.1"></a>
# [v1.3.1](https://github.com/rhysd/actionlint/releases/tag/v1.3.1) - 30 Jun 2021

- Files are checked in parallel. This made actionlint around 1.3x faster with 3 workflow files in my environment
- Manual for `man` command was added. `actionlint.1` is included in released archives. If you installed actionlint via Homebrew, the manual is also installed automatically
- `-version` now reports how the binary was built (Go version, arch, os, ...)
- Added [`Command`](https://pkg.go.dev/github.com/rhysd/actionlint#Command) struct to manage entire command lifecycle
- Order of checked files is now stable. When all the workflows in the current repository are checked, the order is sorted by file names
- Added fuzz target for rule checkers

[Changes][v1.3.1]


<a name="v1.3.0"></a>
# [v1.3.0](https://github.com/rhysd/actionlint/releases/tag/v1.3.0) - 26 Jun 2021

- `-version` now outputs how the executable was installed.
- Fix errors output to stdout was not colorful on Windows.
- Add new `-color` flag to force to enable colorful outputs. This is useful when running actionlint on GitHub Actions since scripts at `run:` don't enable colors.
- `Linter.LintFiles` and `Linter.LintFile` methods take `project` parameter to explicitly specify what project the files belong to. Leaving it `nil` automatically detects projects from their file paths.
- `LintOptions.NoColor` is replaced by `LintOptions.Color`.

Example of `-version` output:

```console
$ brew install actionlint
$ actionlint -version
1.3.0
downloaded from release page

$ go install github.com/rhysd/actionlint/cmd/actionlint@v1.3.0
go: downloading github.com/rhysd/actionlint v1.3.0
$ actionlint -version
v1.3.0
built from source
```

Example of running actionlint on GitHub Actions forcing to enable color output:

```yaml
- name: Check workflow files
  run: |
    bash <(curl https://raw.githubusercontent.com/rhysd/actionlint/main/scripts/download-actionlint.bash)
    ./actionlint -color
  shell: bash
```

[Changes][v1.3.0]


<a name="v1.2.0"></a>
# [v1.2.0](https://github.com/rhysd/actionlint/releases/tag/v1.2.0) - 25 Jun 2021

- [pyflakes](https://github.com/PyCQA/pyflakes) integration was added. If `pyflakes` is installed on your system, actionlint checks Python scripts in `run:` (when `shell: python`) with it. See [the rule document](https://github.com/rhysd/actionlint#check-pyflakes-integ) for more details.
- Error handling while running rule checkers was improved. When some internal error occurs while applying rules, actionlint stops correctly due to the error. Previously, such errors were only shown in debug logs and actionlint continued checks.
- Fixed sanitizing `${{ }}` expressions in scripts before passing them to shellcheck or pyflakes. Previously expressions were not correctly sanitized when `}}` came before `${{`.

[Changes][v1.2.0]


<a name="v1.1.2"></a>
# [v1.1.2](https://github.com/rhysd/actionlint/releases/tag/v1.1.2) - 21 Jun 2021

- Run `shellcheck` command for scripts at `run:` in parallel. Since executing an external process is heavy and running shellcheck was bottleneck of actionlint, this brought better performance. In my environment, it was **more than 3x faster** than before.
- Sort errors by their positions in the source file.

[Changes][v1.1.2]


<a name="v1.1.1"></a>
# [v1.1.1](https://github.com/rhysd/actionlint/releases/tag/v1.1.1) - 20 Jun 2021

- [`download-actionlint.yaml`](https://github.com/rhysd/actionlint/blob/main/scripts/download-actionlint.bash) now sets `executable` output when it is run in GitHub Actions environment. Please see [instruction in 'Install' document](https://github.com/rhysd/actionlint#ci-services) for the usage.
- Redundant type `ArrayDerefType` was removed. Instead, [`Deref` field](https://pkg.go.dev/github.com/rhysd/actionlint#ArrayType) is now provided in `ArrayType`.
- Fix crash on broken YAML input.
- `actionlint -version` returns correct version string even if the `actionlint` command was installed via `go install`.

[Changes][v1.1.1]


<a name="v1.1.0"></a>
# [v1.1.0](https://github.com/rhysd/actionlint/releases/tag/v1.1.0) - 19 Jun 2021

- Ignore [SC1091](https://github.com/koalaman/shellcheck/wiki/SC1091) and [SC2194](https://github.com/koalaman/shellcheck/wiki/SC2194) on running shellcheck. These are reported as false positives due to sanitization of `${{ ... }}`. See [the check doc](https://github.com/rhysd/actionlint#check-shellcheck-integ) to know the sanitization.
- actionlint replaces `${{ }}` in `run:` scripts before passing them to shellcheck. v1.0.0 replaced `${{ }}` with whitespaces, but it caused syntax errors in some scripts (e.g. `if ${{ ... }}; then ...`). Instead, v1.1.0 replaces `${{ }}` with underscores. For example, `${{ matrix.os }}` is replaced with `________________`.
- Add [`download-actionlint.bash`](https://github.com/rhysd/actionlint/blob/main/scripts/download-actionlint.bash) script to download pre-built binaries easily on CI services. See [installation document](https://github.com/rhysd/actionlint#on-ci) for the usage.
- Better error message on lexing `"` in `${{ }}` expression since double quote is usually misused for string delimiters
- `-ignore` option can now be specified multiple times.
- Fix `github.repositoryUrl` was not correctly resolved in `${{ }}` expression
- Reports an error when `if:` condition does not use `${{ }}` but the expression contains any operators. [The official document](https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepsif) prohibits this explicitly to avoid conflicts with YAML syntax.
- Clarify that the version of this repository is for `actionlint` CLI tool, not for library. It means that the APIs may have breaking changes on minor or patch version bumps.
- Add more tests and refactor some code. Enumerating quoted items in error message is now done more efficiently and in deterministic order.

[Changes][v1.1.0]


<a name="v1.0.0"></a>
# [v1.0.0](https://github.com/rhysd/actionlint/releases/tag/v1.0.0) - 16 Jun 2021

First release :tada:

See documentation for more details:

- [Installation](https://github.com/rhysd/actionlint#install)
- [Usage](https://github.com/rhysd/actionlint#usage)
- [Checks done by actionlint](https://github.com/rhysd/actionlint#checks)

[Changes][v1.0.0]


[v1.6.21]: https://github.com/rhysd/actionlint/compare/v1.6.20...v1.6.21
[v1.6.20]: https://github.com/rhysd/actionlint/compare/v1.6.19...v1.6.20
[v1.6.19]: https://github.com/rhysd/actionlint/compare/v1.6.18...v1.6.19
[v1.6.18]: https://github.com/rhysd/actionlint/compare/v1.6.17...v1.6.18
[v1.6.17]: https://github.com/rhysd/actionlint/compare/v1.6.16...v1.6.17
[v1.6.16]: https://github.com/rhysd/actionlint/compare/v1.6.15...v1.6.16
[v1.6.15]: https://github.com/rhysd/actionlint/compare/v1.6.14...v1.6.15
[v1.6.14]: https://github.com/rhysd/actionlint/compare/v1.6.13...v1.6.14
[v1.6.13]: https://github.com/rhysd/actionlint/compare/v1.6.12...v1.6.13
[v1.6.12]: https://github.com/rhysd/actionlint/compare/v1.6.11...v1.6.12
[v1.6.11]: https://github.com/rhysd/actionlint/compare/v1.6.10...v1.6.11
[v1.6.10]: https://github.com/rhysd/actionlint/compare/v1.6.9...v1.6.10
[v1.6.9]: https://github.com/rhysd/actionlint/compare/v1.6.8...v1.6.9
[v1.6.8]: https://github.com/rhysd/actionlint/compare/v1.6.7...v1.6.8
[v1.6.7]: https://github.com/rhysd/actionlint/compare/v1.6.6...v1.6.7
[v1.6.6]: https://github.com/rhysd/actionlint/compare/v1.6.5...v1.6.6
[v1.6.5]: https://github.com/rhysd/actionlint/compare/v1.6.4...v1.6.5
[v1.6.4]: https://github.com/rhysd/actionlint/compare/v1.6.3...v1.6.4
[v1.6.3]: https://github.com/rhysd/actionlint/compare/v1.6.2...v1.6.3
[v1.6.2]: https://github.com/rhysd/actionlint/compare/v1.6.1...v1.6.2
[v1.6.1]: https://github.com/rhysd/actionlint/compare/v1.6.0...v1.6.1
[v1.6.0]: https://github.com/rhysd/actionlint/compare/v1.5.3...v1.6.0
[v1.5.3]: https://github.com/rhysd/actionlint/compare/v1.5.2...v1.5.3
[v1.5.2]: https://github.com/rhysd/actionlint/compare/v1.5.1...v1.5.2
[v1.5.1]: https://github.com/rhysd/actionlint/compare/v1.5.0...v1.5.1
[v1.5.0]: https://github.com/rhysd/actionlint/compare/v1.4.3...v1.5.0
[v1.4.3]: https://github.com/rhysd/actionlint/compare/v1.4.2...v1.4.3
[v1.4.2]: https://github.com/rhysd/actionlint/compare/v1.4.1...v1.4.2
[v1.4.1]: https://github.com/rhysd/actionlint/compare/v1.4.0...v1.4.1
[v1.4.0]: https://github.com/rhysd/actionlint/compare/v1.3.2...v1.4.0
[v1.3.2]: https://github.com/rhysd/actionlint/compare/v1.3.1...v1.3.2
[v1.3.1]: https://github.com/rhysd/actionlint/compare/v1.3.0...v1.3.1
[v1.3.0]: https://github.com/rhysd/actionlint/compare/v1.2.0...v1.3.0
[v1.2.0]: https://github.com/rhysd/actionlint/compare/v1.1.2...v1.2.0
[v1.1.2]: https://github.com/rhysd/actionlint/compare/v1.1.1...v1.1.2
[v1.1.1]: https://github.com/rhysd/actionlint/compare/v1.1.0...v1.1.1
[v1.1.0]: https://github.com/rhysd/actionlint/compare/v1.0.0...v1.1.0
[v1.0.0]: https://github.com/rhysd/actionlint/tree/v1.0.0

 <!-- Generated by https://github.com/rhysd/changelog-from-release -->
