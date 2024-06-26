# DRS Initialize Example

This is an example of using the Terraform AWS Provider to manually initialize your account to use DRS. For more information see the [AWS instructions](https://docs.aws.amazon.com/drs/latest/userguide/getting-started-initializing.html).

Running the example:

1. Run `terraform apply`
2. Assume the `AWSElasticDisasterRecoveryInitializerRole` role
3. Use the AWS CLI to initialize: `aws drs initialize-service`
