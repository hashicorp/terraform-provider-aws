# skaff

**WARNING:** We are actively developing this tool. We may completely change it, abandon it, or delete it without notice. *We do not recommend using this tool at this time.*

The `skaff` tool is a Terraform AWS Provider scaffolding tool.

To use `skaff`, starting in the `terraform-provider-aws` directory:

1. `cd skaff`
2. `go install .`
3. Go to the service where your new resource will reside. _E.g._, `cd ../internal/service/mq`.
4. To get help, enter `skaff` without arguments.
5. Generate a resource with helpful comments. _E.g._, `skaff resource --name BrokerReboot` (or equivalently `skaff resource -n BrokerReboot`).

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