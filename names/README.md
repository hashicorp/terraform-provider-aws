# names

Package `names` provides AWS service-name information that is critical to the Terraform AWS Provider working correctly. If you are unsure about a change you are making, please do not hesitate to ask!

The core of the `names` package is the `names_data.csv`, which contains raw, comma-separated data about naming in the AWS Provider, AWS Go SDKs v1 and v2, and AWS CLI. The file that is dynamically embedded at build time in the AWS Provider. _The information it contains must be correct._ Please double-check any changes.

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
| 4 | ProviderPackageActual | Reference | Actual current name of TF AWS Provider package (_e.g._, if `internal/service/amp`, then "amp"); blank if doesn't exist |
| 5 | ProviderPackageCorrect | Reference | What the TF AWS Provider package name should be based on shorter of AWSCLIV2CommandNoDashes and GoV2Package |
| 6 | ProviderPackageBoth | Code | (Service Identifier)[https://github.com/hashicorp/terraform-provider-aws/blob/main/docs/contributing/naming.md#service-identifier], _i.e._, ProviderPackageActual, if non-blank; otherwise, ProviderPackageCorrect |
| 7 | ProviderNameUpper | Code | Correctly capitalized name; case-insensitive comparison of this and ProviderPackageBoth is always equal |
| 8 | Aliases | Code | _Semicolon_-separated list of name variations (_e.g._, for AMP, `prometheus;prometheusservice`). **DO NOT** include ProviderPackageBoth as it will be included for you. Leave blank if none. |
| 9 | GoV1ClientName | Code | Exact name of the AWS SDK for Go v1 client type (_e.g._, see the [`New()` return type](https://docs.aws.amazon.com/sdk-for-go/api/service/ses/#New) for SES) |
| 10 | HumanFriendly | Code | Human-friendly name of service as used by AWS |
| 11 | Brand | Code | Either `Amazon`, `AWS`, or blank (rare) as used by AWS |
| 12 | Note | Reference | Very brief note usually to explain why excluded |
| 13 | Exclude | Code | Whether or not service should be included; if included (blank), ProviderPackageBoth must have a value |
| 14 | DeprecatedEnvVar | Code | Deprecated environment variable name |
| 15 | EnvVar | Code | Current environment variable associated with service |
| 16 | SDKVersion | Code | Whether service currently uses AWS SDK for Go v1 or v2; use `1` or `2` |
| 17 | SkipClientGenerate | Code | Several service clients need special configuration; use a non-empty value to skip them |

For more information about service naming, including the first seven columns, see (the Naming Guide)[https://github.com/hashicorp/terraform-provider-aws/blob/main/docs/contributing/naming.md#service-identifier].
