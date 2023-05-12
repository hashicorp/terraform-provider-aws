# Adding a Newly Released AWS Region

New regions can typically be used immediately with the provider, with two important caveats:

- Regions often need to be explicitly enabled via the AWS console. See [ap-east-1 launch blog](https://aws.amazon.com/blogs/aws/now-open-aws-asia-pacific-hong-kong-region/) for an example of how to enable a new region for use.
- Until the provider is aware of the new region, automatic region validation will fail. In order to use the region before validation support is added to the provider you will need to disable region validation by doing the following:

```terraform
provider "aws" {
  # ... potentially other configuration ...

  region                 = "me-south-1"
  skip_region_validation = true
}
```

## Enabling Region Validation

Support for region validation requires that the provider has an updated AWS Go SDK dependency that includes the new region. These are added to the AWS Go SDK `aws/endpoints/defaults.go` file and generally noted in the AWS Go SDK `CHANGELOG` as `aws/endpoints: Updated Regions`. This also needs to be done in the core Terraform binary itself to enable it for the S3 backend. The provider currently takes a dependency on both v1 AND v2 of the AWS Go SDK, as we start to base new (and migrate) resources on v2. Many of the authentication and provider level configuration interactions are also located in the aws-go-sdk-base library. As all of these things take direct dependencies and as a result there ends up being quite a few places these dependency updates need to be made.

### Update aws-go-sdk-base

[aws-go-sdk-base](https://github.com/hashicorp/aws-sdk-go-base)

- Update [aws-go-sdk](https://github.com/aws/aws-sdk-go)
- Update [aws-go-sdk-v2](https://github.com/aws/aws-sdk-go-v2)

### Update Terraform AWS Provider

[provider](https://github.com/hashicorp/terraform-provider-aws)

- Update [aws-go-sdk](https://github.com/aws/aws-sdk-go)
- Update [aws-go-sdk-v2](https://github.com/aws/aws-sdk-go-v2)
- Update [aws-go-sdk-base](https://github.com/hashicorp/aws-sdk-go-base)

### Update Terraform Core (S3 Backend)

[core](https://github.com/hashicorp/terraform)

- Update [aws-go-sdk](https://github.com/aws/aws-sdk-go)
- Update [aws-go-sdk-v2](https://github.com/aws/aws-sdk-go-v2)
- Update [aws-go-sdk-base](https://github.com/hashicorp/aws-sdk-go-base)

```shell
go get github.com/aws/aws-sdk-go@v#.#.#
go mod tidy
```

See the [Changelog Process](changelog-process.md) document for example changelog format.

## Update Region Specific values in static Data Sources

Some data sources include static values specific to regions that are not available via a standard AWS API call. These will need to be manually updated. AWS employees can code search previous region values to find new region values in internal packages like RIPStaticConfig if they are not documented yet.

- Check [Elastic Load Balancing endpoints and quotas](https://docs.aws.amazon.com/general/latest/gr/elb.html#elb_region) and add Route53 Hosted Zone ID if available to [`internal/service/elb/hosted_zone_id_data_source.go`](https://github.com/hashicorp/terraform-provider-aws/tree/main/internal/service/elb/hosted_zone_id_data_source.go) and [`internal/service/elbv2/hosted_zone_id_data_source.go`](https://github.com/hashicorp/terraform-provider-aws/tree/main/internal/service/elbv2/hosted_zone_id_data_source.go)
- Check [Amazon Simple Storage Service endpoints and quotas](https://docs.aws.amazon.com/general/latest/gr/s3.html#s3_region) and add Route53 Hosted Zone ID if available to [`internal/service/s3/hosted_zones.go`](https://github.com/hashicorp/terraform-provider-aws/tree/main/internal/service/s3/hosted_zones.go)
- Check [CloudTrail Supported Regions docs](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-supported-regions.html#cloudtrail-supported-regions) and add AWS Account ID if available to [`internal/service/cloudtrail/service_account_data_source.go`](https://github.com/hashicorp/terraform-provider-aws/tree/main/internal/service/cloudtrail/service_account_data_source.go)
- ~~Check [Elastic Load Balancing Access Logs docs](https://docs.aws.amazon.com/elasticloadbalancing/latest/classic/enable-access-logs.html#attach-bucket-policy) and add Elastic Load Balancing Account ID if available to [`internal/service/elb/service_account_data_source.go`](https://github.com/hashicorp/terraform-provider-aws/tree/main/internal/service/elb/service_account_data_source.go)~~
- ~~Check [Redshift Database Audit Logging docs](https://docs.aws.amazon.com/redshift/latest/mgmt/db-auditing.html#db-auditing-bucket-permissions) and add AWS Account ID if available to [`internal/service/redshift/service_account_data_source.go`](https://github.com/hashicorp/terraform-provider-aws/tree/main/internal/service/redshift/service_account_data_source.go)~~
- Check [AWS Elastic Beanstalk endpoints and quotas](https://docs.aws.amazon.com/general/latest/gr/elasticbeanstalk.html) and add Route53 Hosted Zone ID if available to [`internal/service/elasticbeanstalk/hosted_zone_data_source.go`](https://github.com/hashicorp/terraform-provider-aws/tree/main/internal/service/elasticbeanstalk/hosted_zone_data_source.go)
- Check [SageMaker docs](https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-algo-docker-registry-paths.html) and add AWS Account IDs if available to [`internal/service/sagemaker/prebuilt_ecr_image_data_source.go`](https://github.com/hashicorp/terraform-provider-aws/tree/main/internal/service/sagemaker/prebuilt_ecr_image_data_source.go)
