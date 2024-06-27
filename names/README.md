# names

Package `names` provides AWS service-name information that is critical to the Terraform AWS Provider working correctly. If you are unsure about a change you are making, please do not hesitate to ask!

`NOTE:` The information in `data/names_data.hcl` affects the provider, generators, documentation, website navigation, etc. working correctly. _Please do not make any changes until you understand the table below._

The core of the `names` package is `data/names_data.hcl`, which contains HCL data about naming in the AWS Provider, AWS Go SDKs v1 and v2, and AWS CLI. The file is dynamically embedded at build time in the AWS Provider and referenced by generators when generating code. _The information it contains must be correct._ Please double-check any changes.

Consumers of `names` include:

* Package `provider` (`internal/provider`)
* Package `conns` (`internal/conns`)
* AWS Provider generators
* `skaff` tool

After any edits to `data/names_data.hcl`, run `make gen`. Doing so regenerates code and performs checks on `data/names_data.hcl`.

The schema of the attributes and blocks of `data/names_data.hcl` are as follows:

```hcl
service "" {

  // If both of these attributes are the same as the service block's name, this block will be ommitted
  cli_v2_command { 
    aws_cli_v2_command           = ""
    aws_cli_v2_command_no_dashes = ""
  }

  // If both of these attributes are the same as the service block's name, this block will be ommitted
  go_packages { 
    v1_package = ""
    v2_package = ""
  } 

  // If any blocks below here have attirbutes with empty strings or false bools, they will be ommitted
  // Blocks with zero attributes will be ommitted 
  sdk {
    id             = "" 
    client_version = [] 
  }

  names {
    aliases             = [""] // This can also be excluded if it is empty
    provider_name_upper = ""
    human_friendly      = ""
  }

  client {
    go_v1_client_typename = ""
    skip_client_generate  = bool
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = bool
  }

  resource_prefix {
    actual  = ""
    correct = ""
  }

  provider_package_correct = ""
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = [""]
  brand               = ""
  exclude             = bool
  not_implemented     = bool
  allowed_subcategory = bool
  note                = ""
}

```

The explanation of the attributes of `data/names_data.hcl` are as follows:

| Name | Use | Description |
| --- | --- | --- |
| **ProviderPackageActual** | Code | Actual TF AWS provide package name _if_ `provider_package_correct` is not used; takes precedence over `provider_package_correct` for service block name if both are defined |
| `aws_cli_v2_command` | Reference | Service command in [AWS CLI v2](https://awscli.amazonaws.com/v2/documentation/api/latest/index.html) |
| `aws_cli_v2_command_no_dashes` | Reference | Same as `aws_cli_v2_command` without dashes |
| `v1_package` | Code | [AWS SDK for Go v1](https://docs.aws.amazon.com/sdk-for-go/api/) package name |
| `v2_package` | Code | [AWS SDK for Go v2](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2) package name |
| `id` | Code | Represents the ServiceID of a AWS service which is a unique identifier of a specific service |
| `client_version` | Code | HCL int list containing if in the TF AWS Provider, the service currently uses AWS SDK for Go v1 and/or v2; each integer represents the correlating version|
| `aliases` | Code | HCL string list of name variations (_e.g._, for "AMP", `prometheus,prometheusservice`). Do not include **ProviderPackageActual (or `provider_package_correct`, if blank) since that will create duplicates in the [Custom Endpoints guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/custom-service-endpoints). |
| `provider_name_upper` | Code | [Correctly capitalized](https://hashicorp.github.io/terraform-provider-aws/naming/#mixed-caps) `ProviderPackageActual`, if it exists, otherwise `provider_package_correct` |
| `human_friendly` | Code | [REQUIRED] Human-friendly name of service as used by AWS; documentation `subcategory` must exactly match this value; used in website navigation and error messages |
| `go_v1_client_typename` | Code | _Exact name_ (_i.e._, spelling and capitalization) of the AWS SDK for Go v1 client type (_e.g._, see the [`New()` return type](https://docs.aws.amazon.com/sdk-for-go/api/service/ses/#New) for SES). Also excluded when service only supports AWS SDK for Go v2|
| `skip_client_generate` | Code | Some service clients need special configuration rather than the default generated configuration; use a non-empty value to skip generation but you must then manually configure the client in `internal/conns/config.go` |
| `deprecated_env_var` | Code | Deprecated `AWS_<service>_ENDPOINT` envvar defined for some services |
| `tf_aws_env_var` | Code | `TF_AWS_<service>_ENDPOINT` envvar defined for some services |
| `endpoint_api_call` | Code | Command for the AWS cli for describing the current service |
| `endpoint_api_params` | Code | Used in `service_endpoints_gen_test.go` files for API calls that require a configured value |
| `endpoint_region_override` | Code | Specified alternate regional [endpoint]([https://docs.aws.amazon.com/general/latest/gr/rande.html) for API requests |
| `endpoint_only` | Code | Bool based on if `not_implemented` is non-blank, whether the service endpoint should be included in the provider `endpoints` configuration |
| `resource_prefix_actual` | Code | Regular expression to match anomalous TF resource name prefixes (_e.g._, for the resource name `aws_config_config_rule`, `aws_config_` will match all resources); only use if `resource_prefix_correct` is not suitable (_e.g._, `aws_codepipeline_` won't work as there is only one resource named `aws_codepipeline`); takes precedence over `resource_prefix_correct` |
| `resource_prefix_correct` | Code | Regular expression to match what resource name prefixes _should be_ (_i.e._, `aws_` + `provider_package_correct` + `_`); used if `resource_prefix_actual` is blank |
| `provider_package_correct` | Code | Shorter of `aws_cli_v2_command_no_dashes` and `v2_package`; should _not_ be blank if either exists; same as [Service Identifier](https://hashicorp.github.io/terraform-provider-aws/naming/#service-identifier); what the TF AWS Provider package name _should be_; `ProviderPackageActual` takes precedence |
| `split_package_real_package` | Code | If multiple "services" live in one service, this is the package where the service's Go files live (_e.g._, VPC is part of EC2) |
| `file_prefix` | Code | If multiple "services" live in one service, this is the prefix that files must have to be associated with this sub-service (_e.g._, VPC files in the EC2 service are prefixed with `vpc_`); see also `split_packages_real_packages` |
| `doc_prefix` | Code | Hcl string list of prefixes for service documentation files in `website/docs/r` and `website/docs/d`; usually only one prefix, _i.e._, `<`provider_package_correct`>_` |
| `brand` | Code | Either `Amazon`, `AWS`, or blank (rare) as used by AWS; used in error messages |
| `exclude` | Code | Bool based on whether the service should be included; if included (blank), `ProviderPackageActual` or `provider_package_correct` must have a value |
| `allowed_subcategory` | Code | Bool based on if `Exclude` is non-blank, whether to include `human_friendly` in `website/allowed-subcategories.txt` anyway. In other words, if non-blank, overrides `exclude` in some situations. Some excluded pseudo-services (_e.g._, VPC is part of EC2) are still subcategories. Only applies if `Exclude` is non-blank. |
| `not_implemented` | Code | Bool based on whether the service is implemented by the provider |
| `note` | Reference | Very brief note usually to explain why excluded |

For more information about service naming, see [the Naming Guide](https://hashicorp.github.io/terraform-provider-aws/naming/#service-identifier).
