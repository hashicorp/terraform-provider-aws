actionlint
==========
[![CI Badge][]][CI]
[![API Document][api-badge]][apidoc]

[actionlint][repo] is a static checker for GitHub Actions workflow files. [Try it online!][playground]

Features:

- **Syntax check for workflow files** to check unexpected or missing keys following [workflow syntax][syntax-doc]
- **Strong type check for `${{ }}` expressions** to catch several semantic errors like access to not existing property,
  type mismatches, ...
- **Actions usage check** to check that inputs at `with:` and outputs in `steps.{id}.outputs` are correct
- **Reusable workflow check** to check inputs/outputs/secrets of reusable workflows and workflow calls
- **[shellcheck][] and [pyflakes][] integrations** for scripts at `run:`
- **Security checks**; [script injection][script-injection-doc] by untrusted inputs, hard-coded credentials
- **Other several useful checks**; [glob syntax][filter-pattern-doc] validation, dependencies check for `needs:`,
  runner label validation, cron syntax validation, ...

See [the full list](docs/checks.md) of checks done by actionlint.

<img src="https://github.com/rhysd/ss/blob/master/actionlint/main.gif?raw=true" alt="actionlint reports 7 errors" width="806" height="492"/>

**Example of broken workflow:**

```yaml
on:
  push:
    branch: main
    tags:
      - 'v\d+'
jobs:
  test:
    strategy:
      matrix:
        os: [macos-latest, linux-latest]
    runs-on: ${{ matrix.os }}
    steps:
      - run: echo "Checking commit '${{ github.event.head_commit.message }}'"
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node_version: 16.x
      - uses: actions/cache@v3
        with:
          path: ~/.npm
          key: ${{ matrix.platform }}-node-${{ hashFiles('**/package-lock.json') }}
        if: ${{ github.repository.permissions.admin == true }}
      - run: npm install && npm test
```

**actionlint reports 7 errors:**

```
test.yaml:3:5: unexpected key "branch" for "push" section. expected one of "branches", "branches-ignore", "paths", "paths-ignore", "tags", "tags-ignore", "types", "workflows" [syntax-check]
  |
3 |     branch: main
  |     ^~~~~~~
test.yaml:5:11: character '\' is invalid for branch and tag names. only special characters [, ?, +, *, \ ! can be escaped with \. see `man git-check-ref-format` for more details. note that regular expression is unavailable. note: filter pattern syntax is explained at https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#filter-pattern-cheat-sheet [glob]
  |
5 |       - 'v\d+'
  |           ^~~~
test.yaml:10:28: label "linux-latest" is unknown. available labels are "windows-latest", "windows-2022", "windows-2019", "windows-2016", "ubuntu-latest", "ubuntu-22.04", "ubuntu-20.04", "ubuntu-18.04", "macos-latest", "macos-12", "macos-12.0", "macos-11", "macos-11.0", "macos-10.15", "self-hosted", "x64", "arm", "arm64", "linux", "macos", "windows". if it is a custom label for self-hosted runner, set list of labels in actionlint.yaml config file [runner-label]
   |
10 |         os: [macos-latest, linux-latest]
   |                            ^~~~~~~~~~~~~
test.yaml:13:41: "github.event.head_commit.message" is potentially untrusted. avoid using it directly in inline scripts. instead, pass it through an environment variable. see https://docs.github.com/en/actions/learn-github-actions/security-hardening-for-github-actions for more details [expression]
   |
13 |       - run: echo "Checking commit '${{ github.event.head_commit.message }}'"
   |                                         ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
test.yaml:17:11: input "node_version" is not defined in action "actions/setup-node@v3". available inputs are "always-auth", "architecture", "cache", "cache-dependency-path", "check-latest", "node-version", "node-version-file", "registry-url", "scope", "token" [action]
   |
17 |           node_version: 16.x
   |           ^~~~~~~~~~~~~
test.yaml:21:20: property "platform" is not defined in object type {os: string} [expression]
   |
21 |           key: ${{ matrix.platform }}-node-${{ hashFiles('**/package-lock.json') }}
   |                    ^~~~~~~~~~~~~~~
test.yaml:22:17: receiver of object dereference "permissions" must be type of object but got "string" [expression]
   |
22 |         if: ${{ github.repository.permissions.admin == true }}
   |                 ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
```

## Why?

- **Running a workflow is time consuming.** You need to push the changes and wait until the workflow runs on GitHub even if
  it contains some trivial mistakes. [act][] is useful to debug the workflow locally. But it is not suitable for CI and still
  time consuming when your workflow gets larger.
- **Checks of workflow files by GitHub are very loose.** It reports no error even if unexpected keys are in mappings
  (meant that some typos in keys). And also it reports no error when accessing to property which is actually not existing.
  For example `matrix.foo` when no `foo` is defined in `matrix:` section, it is evaluated to `null` and causes no error.
- **Some mistakes silently break a workflow.** Most common case I saw is specifying missing property to cache key. In the
  case cache silently does not work properly but a workflow itself runs without error. So you might not notice the mistake
  forever.

## Quick start

Install `actionlint` command by downloading [the released binary][releases] or by Homebrew or by `go install`. See
[the installation document](docs/install.md) for more details like how to manage the command with several package managers
or run via Docker container.

```sh
go install github.com/rhysd/actionlint/cmd/actionlint@latest
```

Basically all you need to do is run the `actionlint` command in your repository. actionlint automatically detects workflows and
checks errors. actionlint focuses on finding out mistakes. It tries to catch errors as much as possible and make false positives
as minimal as possible.

```sh
actionlint
```

Another option to try actionlint is [the online playground][playground]. Your browser can run actionlint through WebAssembly.

See [the usage document](docs/usage.md) for more details.

## Documents

- [Checks](docs/checks.md): Full list of all checks done by actionlint with example inputs, outputs, and playground links.
- [Installation](docs/install.md): Installation instructions. Prebuilt binaries, Homebrew package, a Docker image, building from
  source, a download script (for CI) are available.
- [Usage](docs/usage.md): How to use `actionlint` command locally or on GitHub Actions, the online playground, an official Docker
  image, and integrations with reviewdog, Problem Matchers, super-linter, pre-commit, VS Code.
- [Configuration](docs/config.md): How to configure actionlint behavior. Currently only labels of self-hosted runners can be
  configured.
- [Go API](docs/api.md): How to use actionlint as Go library.
- [References](docs/reference.md): Links to resources.

## Bug reporting

When you see some bugs or false positives, it is helpful to [file a new issue][issue-form] with a minimal example
of input. Giving me some feedbacks like feature requests or ideas of additional checks is also welcome.

## License

actionlint is distributed under [the MIT license](./LICENSE.txt).

[CI Badge]: https://github.com/rhysd/actionlint/workflows/CI/badge.svg?branch=main&event=push
[CI]: https://github.com/rhysd/actionlint/actions?query=workflow%3ACI+branch%3Amain
[api-badge]: https://pkg.go.dev/badge/github.com/rhysd/actionlint.svg
[apidoc]: https://pkg.go.dev/github.com/rhysd/actionlint
[repo]: https://github.com/rhysd/actionlint
[playground]: https://rhysd.github.io/actionlint/
[shellcheck]: https://github.com/koalaman/shellcheck
[pyflakes]: https://github.com/PyCQA/pyflakes
[act]: https://github.com/nektos/act
[syntax-doc]: https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions
[filter-pattern-doc]: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#filter-pattern-cheat-sheet
[script-injection-doc]: https://docs.github.com/en/actions/learn-github-actions/security-hardening-for-github-actions#understanding-the-risk-of-script-injections
[issue-form]: https://github.com/rhysd/actionlint/issues/new
[releases]: https://github.com/rhysd/actionlint/releases
