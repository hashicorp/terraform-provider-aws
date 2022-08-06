# skaff

`skaff` is a Terraform AWS Provider scaffolding tool.

## Overview workflow steps

1. Figure out what you're trying to do:
    * Create a resource or a data source?
    * AWS Go SDK v1 or v2 code?
    * Name of the new resource or data source?
2. Use `skaff` to generate provider code
3. Go through the generated code completing code and customizing for the AWS Go SDK API
4. Run, test, refine
5. Remove "TIP" comments
6. Submit code in pull request

## Running `skaff`

1. Use Git to clone the GitHub [https://github.com/hashicorp/terraform-provider-aws](hashicorp/terraform-provider-aws) repository.
2. `cd skaff`
3. `go install .`
4. Change directories to the service where your new resource will reside. _E.g._, `cd ../internal/service/mq`.
5. To get help, enter `skaff` without arguments.
6. Generate a resource. _E.g._, `skaff resource --name BrokerReboot` (or equivalently `skaff resource -n BrokerReboot`).

## Usage 

### Help
```
$ skaff --help
Usage:
  skaff [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  datasource  Create scaffolding for a data source
  help        Help about any command
  resource    Create scaffolding for a resource

Flags:
  -h, --help   help for skaff
```

### Autocompletion
Generate the autocompletion script for skaff for the specified shell
```
$ skaff completion --help
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
```
$ skaff datasource --help
Usage:
  skaff datasource [flags]

Flags:
  -c, --clear-comments     Do not include instructional comments in source
  -f, --force              Force creation, overwriting existing files
  -h, --help               help for datasource
  -n, --name string        Name of the entity
  -s, --snakename string   If skaff doesn't get it right, explicitly give name in snake case (e.g., db_vpc_instance)
```

### Resource
Create scaffolding for a resource
```
$ skaff resource --help
Usage:
  skaff resource [flags]

Flags:
  -c, --clear-comments     Do not include instructional comments in source
  -f, --force              Force creation, overwriting existing files
  -h, --help               help for resource
  -n, --name string        Name of the entity
  -s, --snakename string   If skaff doesn't get it right, explicitly give name in snake case (e.g., db_vpc_instance)
```