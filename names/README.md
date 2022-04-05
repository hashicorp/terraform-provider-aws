# names

Package `names` provides AWS service-name information that is critical to the Terraform AWS Provider working correctly. If you are unsure about a change you are making, please do not hesitate to ask!

**NOTE:** The information in `names_data.csv` affects code working correctly, generators, documentation, website navigation, etc. Please triple-check for correctness before making changes!

The core of the `names` package is the `names_data.csv`, which contains raw, comma-separated data about naming in the AWS Provider, AWS Go SDKs v1 and v2, and AWS CLI. The file is dynamically embedded at build time in the AWS Provider. _The information it contains must be correct._ Please double-check any changes.

Consumers of `names` include:

* Package `provider` (`internal/provider`)
* Package `conns` (`internal/conns`)
* AWS Provider generators
* `skaff` tool

The columns of `names_data.csv` are as follows:

| Index | Name | Use | Description |
| --- | --- | --- | --- |
| 0 | AWSCLIV2Command | Reference | Service command in AWS CLI v2 |
| 1 | AWSCLIV2CommandNoDashes | Reference | Same as AWSCLIV2Command without dashes |
| 2 | GoV1Package | Reference | AWS SDK for Go v1 package name |
| 3 | GoV2Package | Reference | AWS SDK for Go v2 package name |
| 4 | ProviderPackageActual | Code | Actual current name of TF AWS Provider package (_e.g._, if `internal/service/amp`, then `amp`); blank if it doesn't yet exist; takes precedence over ProviderPackageCorrect if both are defined |
| 5 | ProviderPackageCorrect | Code | Shorter of AWSCLIV2CommandNoDashes and GoV2Package; should _not_ be blank if either exists; same as (Service Identifier)[https://github.com/hashicorp/terraform-provider-aws/blob/main/docs/contributing/naming.md#service-identifier]; what the TF AWS Provider package name _should be_; ProviderPackageActual takes precedence |
| 6 | Aliases | Code | _Semicolon_-separated list of name variations (_e.g._, for AMP, `prometheus;prometheusservice`). **DO NOT** include the primary name as it's included for you. Blank if none. |
| 7 | ProviderNameUpper | Code | ProviderPackageActual, if it exists, otherwise ProviderPackageCorrect, correctly capitalized |
| 8 | GoV1ClientName | Code | _Exact name_ (_i.e._, spelling and capitalization) of the AWS SDK for Go v1 client type (_e.g._, see the [`New()` return type](https://docs.aws.amazon.com/sdk-for-go/api/service/ses/#New) for SES) |
| 9 | SkipClientGenerate | Code | Several service clients need special configuration; use a non-empty value to skip them |
| 10 | SDKVersion | Code | Whether service currently uses AWS SDK for Go v1 or v2; use `1` or `2` |
| 11 | ResourcePrefixActual | Code | Regular expression to match anomalous resource prefixes; only use if ResourcePrefixCorrect is not suitable (_e.g._, `aws_codepipeline_` won't work as there is only one resource named `aws_codepipeline`); takes precedence over ResourcePrefixCorrect |
| 12 | ResourcePrefixCorrect | Code | Regular expression to match what resource prefixes _should be_ (_i.e._, `aws_` + ProviderPackageCorrect + `_`); used if ResourcePrefixActual is blank |
| 13 | HumanFriendly | Code | Human-friendly name of service as used by AWS; used in website navigation and error messages |
| 14 | Brand | Code | Either `Amazon`, `AWS`, or blank (rare) as used by AWS; used in error messages |
| 15 | Exclude | Code | Whether or not service should be included; if included (blank), ProviderPackageBoth must have a value |
| 16 | AllowedSubcategory | Code | Include in `website/allowed-subcategories.txt`. Overrides Exclude. Some excluded pseudo-services (e.g., VPC which is part of EC2) are subcategories. If non-blank _or_ Exclude is blank, will be included. |
| 17 | DeprecatedEnvVar | Code | Deprecated environment variable name |
| 18 | EnvVar | Code | Current environment variable associated with service |
| 19 | Note | Reference | Very brief note usually to explain why excluded |

For more information about service naming, see (the Naming Guide)[https://github.com/hashicorp/terraform-provider-aws/blob/main/docs/contributing/naming.md#service-identifier].
