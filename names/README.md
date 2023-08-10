# names

Package `names` provides AWS service-name information that is critical to the Terraform AWS Provider working correctly. If you are unsure about a change you are making, please do not hesitate to ask!

**NOTE:** The information in `names_data.csv` affects the provider, generators, documentation, website navigation, etc. working correctly. _Please do not make any changes until you understand the table below._

The core of the `names` package is `names_data.csv`, which contains raw, comma-separated data about naming in the AWS Provider, AWS Go SDKs v1 and v2, and AWS CLI. The file is dynamically embedded at build time in the AWS Provider and referenced by generators when generating code. _The information it contains must be correct._ Please double-check any changes.

Consumers of `names` include:

* Package `provider` (`internal/provider`)
* Package `conns` (`internal/conns`)
* AWS Provider generators
* `skaff` tool

After any edits to `names_data.csv`, run `make gen`. Doing so regenerates code and performs checks on `names_data.csv`.

The columns of `names_data.csv` are as follows:

| Index | Name | Use | Description |
| --- | --- | --- | --- |
| 0 | **AWSCLIV2Command** | Reference | Service command in [AWS CLI v2](https://awscli.amazonaws.com/v2/documentation/api/latest/index.html) |
| 1 | **AWSCLIV2CommandNoDashes** | Reference | Same as **AWSCLIV2Command** without dashes |
| 2 | **GoV1Package** | Code | [AWS SDK for Go v1](https://docs.aws.amazon.com/sdk-for-go/api/) package name |
| 3 | **GoV2Package** | Code | [AWS SDK for Go v2](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2) package name |
| 4 | **ProviderPackageActual** | Code | Actual TF AWS provide package name _if_ **ProviderPackageCorrect** is not used; takes precedence over **ProviderPackageCorrect** if both are defined |
| 5 | **ProviderPackageCorrect** | Code | Shorter of **AWSCLIV2CommandNoDashes** and **GoV2Package**; should _not_ be blank if either exists; same as [Service Identifier](https://hashicorp.github.io/terraform-provider-aws/naming/#service-identifier); what the TF AWS Provider package name _should be_; **ProviderPackageActual** takes precedence |
| 6 | **SplitPackageRealPackage** | Code | If multiple "services" live in one service, this is the package where the service's Go files live (_e.g._, VPC is part of EC2) |
| 7 | **Aliases** | Code | _Semicolon_-separated list of name variations (_e.g._, for "AMP", `prometheus;prometheusservice`). Do not include **ProviderPackageActual** (or **ProviderPackageCorrect**, if blank) since that will create duplicates in the [Custom Endpoints guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/custom-service-endpoints). |
| 8 | **ProviderNameUpper** | Code | [Correctly capitalized](https://hashicorp.github.io/terraform-provider-aws/naming/#mixedcaps) **ProviderPackageActual**, if it exists, otherwise **ProviderPackageCorrect** |
| 9 | **GoV1ClientTypeName** | Code | _Exact name_ (_i.e._, spelling and capitalization) of the AWS SDK for Go v1 client type (_e.g._, see the [`New()` return type](https://docs.aws.amazon.com/sdk-for-go/api/service/ses/#New) for SES) |
| 10 | **SkipClientGenerate** | Code | Some service clients need special configuration rather than the default generated configuration; use a non-empty value to skip generation but you must then manually configure the client in `internal/conns/config.go` |
| 11 | **ClientSDKV1** | Code | Whether, in the TF AWS Provider, the service currently uses AWS SDK for Go v1; use `1` or leave empty |
| 12 | **ClientSDKV2** | Code | Whether, in the TF AWS Provider, the service currently uses AWS SDK for Go v2; use `2` or leave emtpy |
| 13 | **ResourcePrefixActual** | Code | Regular expression to match anomalous TF resource name prefixes (_e.g._, for the resource name `aws_config_config_rule`, `aws_config_` will match all resources); only use if **ResourcePrefixCorrect** is not suitable (_e.g._, `aws_codepipeline_` won't work as there is only one resource named `aws_codepipeline`); takes precedence over **ResourcePrefixCorrect** |
| 14 | **ResourcePrefixCorrect** | Code | Regular expression to match what resource name prefixes _should be_ (_i.e._, `aws_` + **ProviderPackageCorrect** + `_`); used if **ResourcePrefixActual** is blank |
| 15 | **FilePrefix** | Code | If multiple "services" live in one service, this is the prefix that files must have to be associated with this sub-service (_e.g._, VPC files in the EC2 service are prefixed with `vpc_`); see also **SplitPackageRealPackage** |
| 16 | **DocPrefix** | Code | _Semicolon_-separated list of prefixes for service documentation files in `website/docs/r` and `website/docs/d`; usually only one prefix, _i.e._, `<**ProviderPackageCorrect**>_` |
| 17 | **HumanFriendly** | Code | [REQUIRED] Human-friendly name of service as used by AWS; documentation `subcategory` must exactly match this value; used in website navigation and error messages |
| 18 | **Brand** | Code | Either `Amazon`, `AWS`, or blank (rare) as used by AWS; used in error messages |
| 19 | **Exclude** | Code | Whether or not the service should be included; if included (blank), **ProviderPackageActual** or **ProviderPackageCorrect** must have a value |
| 20 | **AllowedSubcategory** | Code | If **Exclude** is non-blank, whether to include **HumanFriendly** in `website/allowed-subcategories.txt` anyway. In other words, if non-blank, overrides **Exclude** in some situations. Some excluded pseudo-services (_e.g._, VPC is part of EC2) are still subcategories. Only applies if **Exclude** is non-blank. |
| 21 | **DeprecatedEnvVar** | Code | Deprecated environment variable name |
| 22 | **EnvVar** | Code | Current environment variable associated with service |
| 23 | **Note** | Reference | Very brief note usually to explain why excluded |

For more information about service naming, see [the Naming Guide](https://hashicorp.github.io/terraform-provider-aws/naming/#service-identifier).
