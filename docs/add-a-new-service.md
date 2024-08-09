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

If an AWS service must be created in a non-standard way, for example, the service API's endpoint must be accessed via a single AWS Region, then:

1. Make the `skip_client_generate` attribute `true` for the service in [`names/data/names_data.hcl`](https://github.com/hashicorp/terraform-provider-aws/blob/main/names/README.md)

1. Run `make gen`

1. Add a file `internal/<service>/service_package.go` that contains an API client factory function, for example:

=== "AWS Go SDK V2 (Preferred)"

    ```go
    package costoptimizationhub

    import (
        "context"

        "github.com/aws/aws-sdk-go-v2/aws"
        "github.com/aws/aws-sdk-go-v2/service/costoptimizationhub"
        "github.com/hashicorp/terraform-provider-aws/names"
    )

    // NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
    func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*costoptimizationhub.Client, error) {
        cfg := *(config["aws_sdkv2_config"].(*aws.Config))

        return costoptimizationhub.NewFromConfig(cfg,
            costoptimizationhub.WithEndpointResolverV2(newEndpointResolverSDKv2()),
            withBaseEndpoint(config[names.AttrEndpoint].(string)),
            func(o *costoptimizationhub.Options) {
                if config["partition"].(string) == names.StandardPartitionID {
                    // Cost Optimization Hub endpoint is available only in us-east-1 Region.
                    if cfg.Region != names.USEast1RegionID {
                        tflog.Info(ctx, "overriding region", map[string]any{
                            "original_region": cfg.Region,
                            "override_region": names.USEast1RegionID,
                        })
                        o.Region = names.USEast1RegionID
                    }
                }
            },
        ), nil
    }
    ```

=== "AWS Go SDK V1"

    ```go
    package globalaccelerator
    
    import (
        "context"
    
        aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
        endpoints_sdkv1 "github.com/aws/aws-sdk-go/aws/endpoints"
        session_sdkv1 "github.com/aws/aws-sdk-go/aws/session"
        globalaccelerator_sdkv1 "github.com/aws/aws-sdk-go/service/globalaccelerator"
    )
    
    // NewConn returns a new AWS SDK for Go v1 client for this service package's AWS API.
    func (p *servicePackage) NewConn(ctx context.Context) (*globalaccelerator_sdkv1.GlobalAccelerator, error) {
        sess := p.config["session"].(*session_sdkv1.Session)

        cfg := aws.Config{}

        if endpoint := config[names.AttrEndpoint].(string); endpoint != "" {
            tflog.Debug(ctx, "setting endpoint", map[string]any{
                "tf_aws.endpoint": endpoint,
            })
            cfg.Endpoint = aws.String(endpoint)
        } else {
            cfg.EndpointResolver = newEndpointResolverSDKv1(ctx)
        }
    
        // Force "global" services to correct Regions.
        if config["partition"].(string) == endpoints.AwsPartitionID {
            if aws.StringValue(cfg.Region) != endpoints.UsWest2RegionID {
                tflog.Info(ctx, "overriding region", map[string]any{
                    "original_region": aws.StringValue(cfg.Region),
                    "override_region": endpoints.UsWest2RegionID,
                })
                cfg.Region = aws.String(endpoints.UsWest2RegionID)
            }
        }
    
        return globalaccelerator_sdkv1.New(sess.Copy(config)), nil
    }
    ```

## Customizing a new Service Client

If an AWS service must be customized after creation, for example, retry handling must be changed, then:

1. Add a file `internal/<service>/service_package.go` that contains an API client customization function, for example:

=== "AWS Go SDK V1"

    ```go
    package chime
    
    import (
    	"context"
    
    	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
    	request_sdkv1 "github.com/aws/aws-sdk-go/aws/request"
    	chime_sdkv1 "github.com/aws/aws-sdk-go/service/chime"
    	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
    )
    
    // CustomizeConn customizes a new AWS SDK for Go v1 client for this service package's AWS API.
    func (p *servicePackage) CustomizeConn(ctx context.Context, conn *chime_sdkv1.Chime) (*chime_sdkv1.Chime, error) {
    	conn.Handlers.Retry.PushBack(func(r *request_sdkv1.Request) {
    		// When calling CreateVoiceConnector across multiple resources,
    		// the API can randomly return a BadRequestException without an explanation
    		if r.Operation.Name == "CreateVoiceConnector" {
    			if tfawserr.ErrMessageContains(r.Error, chime_sdkv1.ErrCodeBadRequestException, "Service received a bad request") {
    				r.Retryable = aws_sdkv1.Bool(true)
    			}
    		}
    	})
    
    	return conn, nil
    }
    ```
