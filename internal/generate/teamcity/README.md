# TeamCity Acceptance Test Configuration Generator

This generator generates the configuration for the default TeamCity service acceptance test list,
located at `.teamcity/components/generated/services_all.kt`.
Can be invoked using either `make gen` along with all other generators or
`go generate ./internal/generate/teamcity/...` to run this generator specifically.

## Configuration

The generator creates a TeamCity build configuration for each service listed in `names/names_data.csv`.
By default, the service acceptance tests do not use the VPC Lock and use the default parallelism.
These setting can be overridden for each service by adding a `service` entry in the file `acctest_services.hcl`.

The service entry has the following parameters:

* `vpc_lock` - (Optional) Set to `true` to enabled the VPC Lock for the service.
* `parallelism` - (Optional) Set this to specify a parallelism value for the acceptance tests for this service.

For example, the `appstream` service uses both the VPC Lock and a maximum of 10 parallel tests:

```hcl
service "appstream" {
    vpc_lock    = true
    parallelism = 10
}
```
