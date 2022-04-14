## New Region

While region validation is automatically added with SDK updates, new regions
are generally limited in which services they support. Below are some
manually sourced values from documentation. Amazon employees can code search
previous region values to find new region values in internal packages like
RIPStaticConfig if they are not documented yet.

- [ ] Check [Elastic Load Balancing endpoints and quotas](https://docs.aws.amazon.com/general/latest/gr/elb.html#elb_region) and add Route53 Hosted Zone ID if available to [`internal/service/elb/hosted_zone_id_data_source.go`](../../internal/service/elb/hosted_zone_id_data_source.go)
- [ ] Check [Amazon Simple Storage Service endpoints and quotas](https://docs.aws.amazon.com/general/latest/gr/s3.html#s3_region) and add Route53 Hosted Zone ID if available to [`internal/service/s3/hosted_zones.go`](../../internal/service/s3/hosted_zones.go)
- [ ] Check [CloudTrail Supported Regions docs](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-supported-regions.html#cloudtrail-supported-regions) and add AWS Account ID if available to [`internal/service/cloudtrail/service_account_data_source.go`](../../internal/service/cloudtrail/service_account_data_source.go)
- [ ] Check [Elastic Load Balancing Access Logs docs](https://docs.aws.amazon.com/elasticloadbalancing/latest/classic/enable-access-logs.html#attach-bucket-policy) and add Elastic Load Balancing Account ID if available to [`internal/service/elb/service_account_data_source.go`](../../internal/service/elb/service_account_data_source.go)
- [ ] Check [Redshift Database Audit Logging docs](https://docs.aws.amazon.com/redshift/latest/mgmt/db-auditing.html#db-auditing-bucket-permissions) and add AWS Account ID if available to [`internal/service/redshift/service_account_data_source.go`](../../internal/service/redshift/service_account_data_source.go)
- [ ] Check [AWS Elastic Beanstalk endpoints and quotas](https://docs.aws.amazon.com/general/latest/gr/elasticbeanstalk.html#elasticbeanstalk_region) and add Route53 Hosted Zone ID if available to [`internal/service/elasticbeanstalk/hosted_zone_data_source.go`](../../internal/service/elasticbeanstalk/hosted_zone_data_source.go)
- [ ] Check [Sagemaker docs](https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-algo-docker-registry-paths.html) and add AWS Account IDs if available to [`internal/service/sagemaker/prebuilt_ecr_image_data_source.go`](../../internal/service/sagemaker/prebuilt_ecr_image_data_source.go)
