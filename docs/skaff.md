# Provider Scaffolding (skaff)

`skaff` is a Terraform AWS Provider scaffolding command line tool.
It generates resource, data source, or function source files, along with test files which adhere to the latest best practices.
These files are heavily commented with instructions, serving as the best way to get started with provider development.

## Overview workflow steps

1. Figure out what you're trying to do:
    * Resource, data source, or function?
    * [Name it](naming.md).
    !!! tip
        Net-new resources should be implemented with AWS SDK Go V2 and the Terraform Plugin Framework (e.g. the default `skaff` settings).
        See [AWS Go SDK Versions](aws-go-sdk-versions.md), [Terraform Plugin Development Packages](terraform-plugin-development-packages.md), and [this issue](https://github.com/hashicorp/terraform-provider-aws/issues/32917) for additional information.
1. Use `skaff` to generate provider code.
1. Go through the generated code, customizing as necessary.
1. Run, test, refine.
1. Remove "TIP" comments.
1. Submit a pull request.

## Running `skaff`

1. Clone the [Terraform AWS Provider](https://github.com/hashicorp/terraform-provider-aws) repository.
1. Install `skaff`.

    ```sh
    make skaff
    ```

1. Change into the appropriate directory.
    - For resources and data sources, this is the service directory where the new entity will reside, e.g. `internal/service/mq`.
    - For functions, this is `internal/functions`.
1. Generate the resource, data source or function. For example,
    - `skaff resource --name BrokerReboot`.
    - `skaff datasource --name IAMRole`.
    - `skaff function --name ARNParse`.

To get help, enter `skaff` without arguments.

## Usage

### Help

```console
skaff --help
```

```
Usage:
  skaff [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  datasource  Create scaffolding for a data source
  function    Create scaffolding for a function
  help        Help about any command
  resource    Create scaffolding for a resource

Flags:
  -h, --help   help for skaff
```

### Autocompletion

Generate the autocompletion script for `skaff` for the specified shell

```console
skaff completion --help
```

```
Usage:
  skaff completion [command]

Available Commands:
  bash        Generate the autocompletion script for bash
  fish        Generate the autocompletion script for fish
  powershell  Generate the autocompletion script for powershell
  zsh         Generate the autocompletion script for zsh

Flags:
  -h, --help   help for completion

Use "skaff completion [command] --help" for more information about a command
```

### Data Source

Create scaffolding for a data source

```console
skaff datasource --help
```

```
Create scaffolding for a data source

Usage:
  skaff datasource [flags]

Flags:
  -c, --clear-comments     do not include instructional comments in source
  -f, --force              force creation, overwriting existing files
  -h, --help               help for datasource
  -t, --include-tags       Indicate that this resource has tags and the code for tagging should be generated
  -n, --name string        name of the entity
  -p, --plugin-sdkv2       generate for Terraform Plugin SDK V2
  -s, --snakename string   if skaff doesn't get it right, explicitly give name in snake case (e.g., db_vpc_instance)
  -o, --v1                 generate for AWS Go SDK v1 (some existing services)
```

### Function

Create scaffolding for a function.

```console
skaff function --help
```

```
Create scaffolding for a function

Usage:
  skaff function [flags]

Flags:
  -c, --clear-comments       do not include instructional comments in source
  -d, --description string   description of the function
  -f, --force                force creation, overwriting existing files
  -h, --help                 help for function
  -n, --name string          name of the function
  -s, --snakename string     if skaff doesn't get it right, explicitly give name in snake case (e.g., arn_build)
```

### Resource

Create scaffolding for a resource

```console
skaff resource --help
```

```
Create scaffolding for a resource

Usage:
  skaff resource [flags]

Flags:
  -c, --clear-comments     do not include instructional comments in source
  -f, --force              force creation, overwriting existing files
  -h, --help               help for resource
  -t, --include-tags       Indicate that this resource has tags and the code for tagging should be generated
  -n, --name string        name of the entity
  -p, --plugin-sdkv2       generate for Terraform Plugin SDK V2
  -s, --snakename string   if skaff doesn't get it right, explicitly give name in snake case (e.g., db_vpc_instance)
  -o, --v1                 generate for AWS Go SDK v1 (some existing services)
```
