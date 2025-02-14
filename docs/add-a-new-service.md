<!-- markdownlint-configure-file { "code-block-style": false } -->
# Adding a New AWS Service

AWS frequently launches new services, and Terraform support is frequently desired by the community shortly after launch. Depending on the API surface area of the new service, this could be a major undertaking. The following steps should be followed to prepare for adding the resources that allow for Terraform management of that service.

## Perform Service Design

Before adding a new service to the provider it's a good idea to familiarize yourself with the primary workflows practitioners are likely to want to accomplish with the provider to ensure the provider design can solve this. It's not always necessary to cover 100% of the AWS service offering to unblock most workflows.

You should have an idea of what resources and data sources should be added, their dependencies and relative importance concerning the workflow. This should give you an idea of the order in which resources are to be added. It's important to note that generally, we like to review and merge resources in isolation, and avoid combining multiple new resources in one Pull Request.

Using the AWS API documentation as a reference, identify the various APIs that correspond to the CRUD operations which consist of the management surface for that resource. These will be the set of APIs called from the new resource. The API's model attributes will correspond to your resource schema.

From there begin to map out the list of resources you would like to implement, and note your plan on the GitHub issue relating to the service (or create one if one does not exist) for the community and maintainers to feedback.

## Add a Service Client

Before new resources are submitted, please raise a separate pull request containing just the new AWS SDK for Go service client.

To add an AWS SDK for Go service client:

1. Check the file `names/data/names_data.hcl` for the service.

1. If the service is there and the `not_implemented` attribute does not exist, you are ready to implement the first [resource](./add-a-new-resource.md) or [data source](./add-a-new-datasource.md).

1. If the service is there and the `not_implemented` attribute is true, remove it and submit the client pull request as described below.

1. Otherwise, determine the service identifier using the rule described in [the Naming Guide](naming.md#service-identifier).

1. In `names/data/names_data.hcl`, add a new hcl block with all the requested information for the service following the guidance in the [`names` README](https://github.com/hashicorp/terraform-provider-aws/blob/main/names/README.md).

    !!! tip
        Be very careful when adding or changing data in `names_data.hcl`!
        The Provider and generators depend on the file being correct.
        We strongly recommend using an editor with HCL support.

Once the names data is ready, create a new service directory with the appropriate service name.

```console
mkdir internal/service/<service>
```

Add a new file `internal/service/<service>/generate.go` with the following content. This will generate the structs required for [resource self-registration](./add-a-new-resource.md#register-resource-to-the-provider).

```go
// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package <service>
```

Next, generate the client and ensure all dependencies are fetched.

```console
make gen
```

```console
go mod tidy
```

At this point a pull request with the re-generated files and new service client can be submitted.

Once the service client has been added, implement the first [resource](./add-a-new-resource.md) or [data source](./add-a-new-datasource.md) in a separate PR.

## Adding a Custom Service Client

If the service API's endpoint must be accessed via a single AWS Region, then:

1. Add a `endpoint_region_overrides` map attribute to the `endpoint_info` for the service in [`names/data/names_data.hcl`](https://github.com/hashicorp/terraform-provider-aws/blob/main/names/README.md)

```terraform
  endpoint_info {
    endpoint_api_call = ...
    endpoint_region_overrides = {
      "aws" = "us-east-1"
    }
  }
```

1. Run `make gen`

## Customizing a new Service Client

If an AWS service must be customized after creation, for example, retry handling must be changed, then:

1. Add a file `internal/<service>/service_package.go` that contains an API client customization function, for example:

```go
package apigateway

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

func (p *servicePackage) withExtraOptions(_ context.Context, config map[string]any) []func(*apigateway.Options) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return []func(*apigateway.Options){
		func(o *apigateway.Options) {
			o.Retryer = conns.AddIsErrorRetryables(cfg.Retryer().(aws.RetryerV2), retry.IsErrorRetryableFunc(func(err error) aws.Ternary {
				// Many operations can return an error such as:
				//   ConflictException: Unable to complete operation due to concurrent modification. Please try again later.
				// Handle them all globally for the service client.
				if errs.IsAErrorMessageContains[*types.ConflictException](err, "try again later") {
					return aws.TrueTernary
				}
				return aws.UnknownTernary // Delegate to configured Retryer.
			}))
		},
	}
}
```
