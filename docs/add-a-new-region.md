## Add a New Region

New regions can typically be used immediately with the provider, with two caveats:

- Regions often need to be explicitly enabled via the AWS console. See [ap-east-1](https://aws.amazon.com/blogs/aws/now-open-aws-asia-pacific-hong-kong-region/) for an example of how to do that.
- Until the provider is aware of the new region, automatic region validation will fail. In order to use the region before validation support is added to the provider you will need to disable region validation:

```
provider "aws" {
  # ... potentially other configuration ...

  region                 = "me-south-1"
  skip_region_validation = true
}
```

### Enabling Region Validation

Support for region validation needs to be added in two places, the provider itself and the S3 backend which is currently part of Terraform. In both cases, simply updating the AWS Go SDK dependency to a version which supports that region will be enough to enable region validation. These are added to the AWS Go SDK `aws/endpoints/defaults.go` file and generally noted in the AWS Go SDK `CHANGELOG` as `aws/endpoints: Updated Regions`. 

Some datasources will need to be manually updated to include 

Typically our process for new regions is as follows:

- Create new (if not existing) Terraform AWS Provider issue: Support Automatic Region Validation for `XX-XXXXX-#` (Location)
- Create new (if not existing) Terraform S3 Backend issue: backend/s3: Support Automatic Region Validation for `XX-XXXXX-#` (Location)
- [Enable the new region in an AWS testing account](https://docs.aws.amazon.com/general/latest/gr/rande-manage.html#rande-manage-enable) and verify AWS Go SDK update works with the new region with `export AWS_DEFAULT_REGION=XX-XXXXX-#` with the new region and run the `TestAccDataSourceAwsRegion_` acceptance testing or by building the provider and testing a configuration like the following:

```hcl
provider "aws" {
  region = "me-south-1"
}

data "aws_region" "current" {}

output "region" {
  value = data.aws_region.current.name
}
```

- Merge AWS Go SDK update in Terraform AWS Provider and close issue with the following information:

``````markdown
Support for automatic validation of this new region has been merged and will release with version <x.y.z> of the Terraform AWS Provider, later this week.

---

Please note that this new region requires [a manual process to enable](https://docs.aws.amazon.com/general/latest/gr/rande-manage.html#rande-manage-enable). Once enabled in the console, it takes a few minutes for everything to work properly.

If the region is not enabled properly, or the enablement process is still in progress, you can receive errors like these:

```console
$ terraform apply

Error: error validating provider credentials: error calling sts:GetCallerIdentity: InvalidClientTokenId: The security token included in the request is invalid.
    status code: 403, request id: 142f947b-b2c3-11e9-9959-c11ab17bcc63

  on main.tf line 1, in provider "aws":
   1: provider "aws" {
```

---

To use this new region before support has been added to Terraform AWS Provider version in use, you must disable the provider's automatic region validation via:

```hcl
provider "aws" {
  # ... potentially other configuration ...

  region                 = "me-south-1"
  skip_region_validation = true
}
```
``````

- Update the Terraform AWS Provider `CHANGELOG` with the following:

```markdown
NOTES:

* provider: Region validation now automatically supports the new `XX-XXXXX-#` (Location) region. For AWS operations to work in the new region, the region must be explicitly enabled as outlined in the [AWS Documentation](https://docs.aws.amazon.com/general/latest/gr/rande-manage.html#rande-manage-enable). When the region is not enabled, the Terraform AWS Provider will return errors during credential validation (e.g., `error validating provider credentials: error calling sts:GetCallerIdentity: InvalidClientTokenId: The security token included in the request is invalid`) or AWS operations will throw their own errors (e.g., `data.aws_availability_zones.available: Error fetching Availability Zones: AuthFailure: AWS was not able to validate the provided access credentials`). [GH-####]

ENHANCEMENTS:

* provider: Support automatic region validation for `XX-XXXXX-#` [GH-####]
```

- Follow the [Contributing Guide](./contribution-checklists.md#new-region) to submit updates for various data sources to support the new region
- Submit the dependency update to the Terraform S3 Backend by running the following:

```shell
go get github.com/aws/aws-sdk-go@v#.#.#
go mod tidy
```

- Create a S3 Bucket in the new region and verify AWS Go SDK update works with new region by building the Terraform S3 Backend and testing a configuration like the following:

```hcl
terraform {
  backend "s3" {
    bucket = "XXX"
    key    = "test"
    region = "me-south-1"
  }
}

output "test" {
  value = timestamp()
}
```

- After approval, merge AWS Go SDK update in Terraform S3 Backend and close issue with the following information:

``````markdown
Support for automatic validation of this new region has been merged and will release with the next version of the Terraform.

This was verified on a build of Terraform with the update:

```hcl
terraform {
  backend "s3" {
    bucket = "XXX"
    key    = "test"
    region = "me-south-1"
  }
}

output "test" {
  value = timestamp()
}
```

Outputs:

```console
$ terraform init
...
Terraform has been successfully initialized!
```

---

Please note that this new region requires [a manual process to enable](https://docs.aws.amazon.com/general/latest/gr/rande-manage.html#rande-manage-enable). Once enabled in the console, it takes a few minutes for everything to work properly.

If the region is not enabled properly, or the enablement process is still in progress, you can receive errors like these:

```console
$ terraform init

Initializing the backend...

Error: error validating provider credentials: error calling sts:GetCallerIdentity: InvalidClientTokenId: The security token included in the request is invalid.
```

---

To use this new region before this update is released, you must disable the Terraform S3 Backend's automatic region validation via:

```hcl
terraform {
  # ... potentially other configuration ...

  backend "s3" {
    # ... other configuration ...

    region                 = "me-south-1"
    skip_region_validation = true
  }
}
```
``````

- Update the Terraform S3 Backend `CHANGELOG` with the following:

```markdown
NOTES:

* backend/s3: Region validation now automatically supports the new `XX-XXXXX-#` (Location) region. For AWS operations to work in the new region, the region must be explicitly enabled as outlined in the [AWS Documentation](https://docs.aws.amazon.com/general/latest/gr/rande-manage.html#rande-manage-enable). When the region is not enabled, the Terraform S3 Backend will return errors during credential validation (e.g., `error validating provider credentials: error calling sts:GetCallerIdentity: InvalidClientTokenId: The security token included in the request is invalid`). [GH-####]

ENHANCEMENTS:

* backend/s3: Support automatic region validation for `XX-XXXXX-#` [GH-####]
```

While region validation is automatically added with SDK updates, new regions
are generally limited in which services they support. Below are some
manually sourced values from documentation. Amazon employees can code search
previous region values to find new region values in internal packages like
RIPStaticConfig if they are not documented yet.

- Check [Elastic Load Balancing endpoints and quotas](https://docs.aws.amazon.com/general/latest/gr/elb.html#elb_region) and add Route53 Hosted Zone ID if available to [`internal/service/elb/hosted_zone_id_data_source.go`](../../internal/service/elb/hosted_zone_id_data_source.go)
- Check [Amazon Simple Storage Service endpoints and quotas](https://docs.aws.amazon.com/general/latest/gr/s3.html#s3_region) and add Route53 Hosted Zone ID if available to [`internal/service/s3/hosted_zones.go`](../../internal/service/s3/hosted_zones.go)
- Check [CloudTrail Supported Regions docs](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-supported-regions.html#cloudtrail-supported-regions) and add AWS Account ID if available to [`internal/service/cloudtrail/service_account_data_source.go`](../../internal/service/cloudtrail/service_account_data_source.go)
- Check [Elastic Load Balancing Access Logs docs](https://docs.aws.amazon.com/elasticloadbalancing/latest/classic/enable-access-logs.html#attach-bucket-policy) and add Elastic Load Balancing Account ID if available to [`internal/service/elb/service_account_data_source.go`](../../internal/service/elb/service_account_data_source.go)
- Check [Redshift Database Audit Logging docs](https://docs.aws.amazon.com/redshift/latest/mgmt/db-auditing.html#db-auditing-bucket-permissions) and add AWS Account ID if available to [`internal/service/redshift/service_account_data_source.go`](../../internal/service/redshift/service_account_data_source.go)
- Check [AWS Elastic Beanstalk endpoints and quotas](https://docs.aws.amazon.com/general/latest/gr/elasticbeanstalk.html#elasticbeanstalk_region) and add Route53 Hosted Zone ID if available to [`internal/service/elasticbeanstalk/hosted_zone_data_source.go`](../../internal/service/elasticbeanstalk/hosted_zone_data_source.go)
- Check [Sagemaker docs](https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-algo-docker-registry-paths.html) and add AWS Account IDs if available to [`internal/service/sagemaker/prebuilt_ecr_image_data_source.go`](../../internal/service/sagemaker/prebuilt_ecr_image_data_source.go)
