# skaff

**WARNING:** We are actively developing this tool. We may completely change it, abandon it, or delete it without notice. *We do not recommend using this tool at this time.*

The `skaff` tool is a Terraform AWS Provider scaffolding tool.

To use `skaff`, starting in the `terraform-provider-aws` directory:

1. `cd skaff`
2. `go install .`
3. Go to the service where your new resource will reside. _E.g._, `cd ../internal/service/mq`.
4. To get help, enter `skaff` without arguments.
5. Generate a resource with helpful comments. _E.g._, `skaff resource --name BrokerReboot` (or equivalently `skaff resource -n BrokerReboot`).
