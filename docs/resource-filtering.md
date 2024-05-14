<!-- markdownlint-configure-file { "code-block-style": false } -->
# Adding Resource Filtering Support

AWS provides server-side filtering across many services and resources, which can be used when listing resources of that type, for example in the implementation of a data source.
See the [EC2 Listing and filtering your resources page](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Using_Filtering.html#Filtering_Resources_CLI) for information about how server-side filtering can be used with EC2 resources.

To determine if the supporting AWS API supports this functionality:

- Open the AWS Go SDK documentation for the service, e.g., for [`service/rds`](https://docs.aws.amazon.com/sdk-for-go/api/service/rds/). Note: there can be a delay between the AWS announcement and the updated AWS Go SDK documentation.
- Determine if the service API includes functionality for filtering resources (usually a `Filters` argument to a `DescribeThing` API call).

Implementing server-side filtering support for Terraform AWS Provider resources requires the following, each with its own section below:

- _Generated Service Filtering Code_: In the internal code generators (e.g., `internal/generate/namevaluesfilters`), implementation and customization of how a service handles filtering, which is standardized for the resources.
- _Resource Filtering Code Implementation_: In the resource's equivalent data source code (e.g., `internal/service/{servicename}/thing_data_source.go`), implementation of `filter` schema attribute, along with handling in the `Read` function.
- _Resource Filtering Documentation Implementation_: In the resource's equivalent data source documentation (e.g., `website/docs/d/service_thing.html.markdown`), addition of `filter` argument

## Adding Service to Filter Generating Code

This step is only necessary for the first implementation and may have been previously completed. If so, move on to the next section.

More details about this code generation can be found in the [namevaluesfilters documentation](https://github.com/hashicorp/terraform-provider-aws/blob/main/internal/generate/namevaluesfilters/README.md).

Add the AWS Go SDK service name (e.g., `rds`) to `sliceServiceNames` in `internal/generate/namevaluesfilters/generators/servicefilters/main.go`.

- Run `make gen` (`go generate ./...`) and ensure there are no errors via `make test` (`go test ./...`)

### Resource Filter Code Implementation

- In the resource's equivalent data source Go file (e.g., `internal/service/ec2/internet_gateway_data_source.go`), add the following Go import: `"github.com/hashicorp/terraform-provider-aws/internal/generate/namevaluesfilters"`
- In the resource schema, add `"filter": namevaluesfilters.Schema(),`
- Implement the logic to build the list of filters:

=== "Terraform Plugin SDK V2"
    ```go
    input := &ec2.DescribeInternetGatewaysInput{}

    // Filters based on attributes.
    filters := namevaluesfilters.New(map[string]string{
    	"internet-gateway-id": d.Get("internet_gateway_id").(string),
    })
    // Add filters based on key-value tags (N.B. Not applicable to all AWS services that support filtering)
    filters.Add(namevaluesfilters.EC2Tags(keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()))
    // Add filters based on the custom filtering "filter" attribute.
    filters.Add(d.Get("filter").(*schema.Set))

    input.Filters = filters.EC2Filters()
    ```

## Resource Filtering Documentation Implementation

- In the resource's equivalent data source documentation (e.g., `website/docs/d/internet_gateway.html.markdown`), add the following to the arguments reference:

```markdown
* `filter` - (Optional) Custom filter block as described below.

More complex filters can be expressed using one or more `filter` sub-blocks, which take the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInternetGateways.html).

* `values` - (Required) Set of values that are accepted for the given field.
  An Internet Gateway will be selected if any one of the given values matches.
```
