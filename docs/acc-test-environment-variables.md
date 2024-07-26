# Acceptance Testing Environment Variable Dictionary

Environment variables (beyond standard AWS Go SDK ones) used by acceptance testing. See also the `internal/acctest` package.

| Variable | Description |
|----------|-------------|
| `ACM_CERTIFICATE_ROOT_DOMAIN` | Root domain name to use with ACM Certificate testing. |
| `ACM_CERTIFICATE_MULTIPLE_ISSUED_DOMAIN` | Domain name of ACM Certificate with multiple issued certificates. **DEPRECATED:** Should be replaced with `aws_acm_certificate` resource usage in tests. |
| `ACM_CERTIFICATE_MULTIPLE_ISSUED_MOST_RECENT_ARN` | Amazon Resource Name of most recent ACM Certificate with multiple issued certificates. **DEPRECATED:** Should be replaced with `aws_acm_certificate` resource usage in tests. |
| `ACM_CERTIFICATE_SINGLE_ISSUED_DOMAIN` | Domain name of ACM Certificate with a single issued certificate. **DEPRECATED:** Should be replaced with `aws_acm_certificate` resource usage in tests. |
| `ACM_CERTIFICATE_SINGLE_ISSUED_MOST_RECENT_ARN` | Amazon Resource Name of most recent ACM Certificate with a single issued certificate. **DEPRECATED:** Should be replaced with `aws_acm_certificate` resource usage in tests. |
| `ADM_CLIENT_ID` | Identifier for Amazon Device Manager Client in Pinpoint testing. |
| `AMPLIFY_DOMAIN_NAME` | Domain name to use for Amplify domain association testing. |
| `AMPLIFY_GITHUB_ACCESS_TOKEN` | GitHub access token used for AWS Amplify testing. |
| `AMPLIFY_GITHUB_REPOSITORY` | GitHub repository used for AWS Amplify testing. |
| `ADM_CLIENT_SECRET` | Secret for Amazon Device Manager Client in Pinpoint testing. |
| `APNS_BUNDLE_ID` | Identifier for Apple Push Notification Service Bundle in Pinpoint testing. |
| `APNS_CERTIFICATE` | Certificate (PEM format) for Apple Push Notification Service in Pinpoint testing. |
| `APNS_CERTIFICATE_PRIVATE_KEY` | Private key for Apple Push Notification Service in Pinpoint testing. |
| `APNS_SANDBOX_BUNDLE_ID` | Identifier for Sandbox Apple Push Notification Service Bundle in Pinpoint testing. |
| `APNS_SANDBOX_CERTIFICATE` | Certificate (PEM format) for Sandbox Apple Push Notification Service in Pinpoint testing. |
| `APNS_SANDBOX_CERTIFICATE_PRIVATE_KEY` | Private key for Sandbox Apple Push Notification Service in Pinpoint testing. |
| `APNS_SANDBOX_CREDENTIAL` | Credential contents for Sandbox Apple Push Notification Service in SNS Application Platform testing. Conflicts with `APNS_SANDBOX_CREDENTIAL_PATH`. |
| `APNS_SANDBOX_CREDENTIAL_PATH` | Path to credential for Sandbox Apple Push Notification Service in SNS Application Platform testing. Conflicts with `APNS_SANDBOX_CREDENTIAL`. |
| `APNS_SANDBOX_PRINCIPAL` | Principal contents for Sandbox Apple Push Notification Service in SNS Application Platform testing. Conflicts with `APNS_SANDBOX_PRINCIPAL_PATH`. |
| `APNS_SANDBOX_PRINCIPAL_PATH` | Path to the principal for Sandbox Apple Push Notification Service in SNS Application Platform testing. Conflicts with `APNS_SANDBOX_PRINCIPAL`. |
| `APNS_SANDBOX_TEAM_ID` | Identifier for Sandbox Apple Push Notification Service Team in Pinpoint testing. |
| `APNS_SANDBOX_TOKEN_KEY` | Token key file content (.p8 format) for Sandbox Apple Push Notification Service in Pinpoint testing. |
| `APNS_SANDBOX_TOKEN_KEY_ID` | Identifier for Sandbox Apple Push Notification Service Token Key in Pinpoint testing. |
| `APNS_TEAM_ID` | Identifier for Apple Push Notification Service Team in Pinpoint testing. |
| `APNS_TOKEN_KEY` | Token key file content (.p8 format) for Apple Push Notification Service in Pinpoint testing. |
| `APNS_TOKEN_KEY_ID` | Identifier for Apple Push Notification Service Token Key in Pinpoint testing. |
| `APNS_VOIP_BUNDLE_ID` | Identifier for VOIP Apple Push Notification Service Bundle in Pinpoint testing. |
| `APNS_VOIP_CERTIFICATE` | Certificate (PEM format) for VOIP Apple Push Notification Service in Pinpoint testing. |
| `APNS_VOIP_CERTIFICATE_PRIVATE_KEY` | Private key for VOIP Apple Push Notification Service in Pinpoint testing. |
| `APNS_VOIP_TEAM_ID` | Identifier for VOIP Apple Push Notification Service Team in Pinpoint testing. |
| `APNS_VOIP_TOKEN_KEY` | Token key file content (.p8 format) for VOIP Apple Push Notification Service in Pinpoint testing. |
| `APNS_VOIP_TOKEN_KEY_ID` | Identifier for VOIP Apple Push Notification Service Token Key in Pinpoint testing. |
| `APPRUNNER_CUSTOM_DOMAIN` | A custom domain endpoint (root domain, subdomain, or wildcard) for AppRunner Custom Domain Association testing. |
| `AUDITMANAGER_DEREGISTER_ACCOUNT_ON_DESTROY` | Flag to execute tests that will disable AuditManager in the account upon destruction. |
| `AUDITMANAGER_ORGANIZATION_ADMIN_ACCOUNT_ID` | Organization admin account identifier for use in AuditManager testing. |
| `AWS_ALTERNATE_ACCESS_KEY_ID` | AWS access key ID with access to a secondary AWS account for tests requiring multiple accounts. Requires `AWS_ALTERNATE_SECRET_ACCESS_KEY`. Conflicts with `AWS_ALTERNATE_PROFILE`. |
| `AWS_ALTERNATE_SECRET_ACCESS_KEY` | AWS secret access key with access to a secondary AWS account for tests requiring multiple accounts. Requires `AWS_ALTERNATE_ACCESS_KEY_ID`. Conflicts with `AWS_ALTERNATE_PROFILE`. |
| `AWS_ALTERNATE_PROFILE` | AWS profile with access to a secondary AWS account for tests requiring multiple accounts. Conflicts with `AWS_ALTERNATE_ACCESS_KEY_ID` and `AWS_ALTERNATE_SECRET_ACCESS_KEY`. |
| `AWS_ALTERNATE_REGION` | Secondary AWS region for tests requiring multiple regions. Defaults to `us-east-1`. |
| `AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_BODY` | Certificate body of publicly trusted certificate for API Gateway Domain Name testing. |
| `AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_CHAIN` | Certificate chain of publicly trusted certificate for API Gateway Domain Name testing. |
| `AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_PRIVATE_KEY` | Private key of publicly trusted certificate for API Gateway Domain Name testing. |
| `AWS_API_GATEWAY_DOMAIN_NAME_REGIONAL_CERTIFICATE_NAME_ENABLED` | Flag to enable API Gateway Domain Name regional certificate upload testing. |
| `AWS_CODEBUILD_BITBUCKET_SOURCE_LOCATION` | BitBucket source URL for CodeBuild testing. CodeBuild must have access to this repository via OAuth or Source Credentials. Defaults to `https://terraform@bitbucket.org/terraform/aws-test.git`. |
| `AWS_CODEBUILD_GITHUB_SOURCE_LOCATION` | GitHub source URL for CodeBuild testing. CodeBuild must have access to this repository via OAuth or Source Credentials. Defaults to `https://github.com/hashibot-test/aws-test.git`. |
| `AWS_DEFAULT_REGION` | Primary AWS region for tests. Defaults to `us-west-2`. |
| `AWS_DETECTIVE_MEMBER_EMAIL` | Email address for Detective Member testing. A valid email address associated with an AWS root account is required for tests to pass. |
| `AWS_EC2_CLIENT_VPN_LIMIT` | Concurrency limit for Client VPN acceptance tests. [Default is 5](https://docs.aws.amazon.com/vpn/latest/clientvpn-admin/limits.html) if not specified. |
| `AWS_EC2_EIP_PUBLIC_IPV4_POOL` | Identifier for EC2 Public IPv4 Pool for EC2 EIP testing. |
| `AWS_EC2_TRANSIT_GATEWAY_LIMIT` | Concurrency limit for Transit Gateway acceptance tests. [Default is 5](https://docs.aws.amazon.com/vpc/latest/tgw/transit-gateway-quotas.html) if not specified. |
| `AWS_EC2_VERIFIED_ACCESS_INSTANCE_LIMIT` | Concurrency limit for Verified Access acceptance tests. [Default is 5](https://docs.aws.amazon.com/verified-access/latest/ug/verified-access-quotas.html) if not specified. |
| `AWS_GUARDDUTY_MEMBER_ACCOUNT_ID` | Identifier of AWS Account for GuardDuty Member testing. **DEPRECATED:** Should be replaced with standard alternate account handling for tests. |
| `AWS_GUARDDUTY_MEMBER_EMAIL` | Email address for GuardDuty Member testing. **DEPRECATED:** It may be possible to use a placeholder email address instead. |
| `AWS_LAMBDA_IMAGE_LATEST_ID` | ECR repository image URI (tagged as `latest`) for Lambda container image acceptance tests. |
| `AWS_LAMBDA_IMAGE_V1_ID` | ECR repository image URI (tagged as `v1`) for Lambda container image acceptance tests. |
| `AWS_LAMBDA_IMAGE_V2_ID` | ECR repository image URI (tagged as `v2`) for Lambda container image acceptance tests. |
| `AWS_THIRD_ACCESS_KEY_ID` | AWS access key ID with access to a third AWS account for tests requiring multiple accounts. Requires `AWS_THIRD_SECRET_ACCESS_KEY`. Conflicts with `AWS_THIRD_PROFILE`. |
| `AWS_THIRD_SECRET_ACCESS_KEY` | AWS secret access key with access to a third AWS account for tests requiring multiple accounts. Requires `AWS_THIRD_ACCESS_KEY_ID`. Conflicts with `AWS_THIRD_PROFILE`. |
| `AWS_THIRD_PROFILE` | AWS profile with access to a third AWS account for tests requiring multiple accounts. Conflicts with `AWS_THIRD_ACCESS_KEY_ID` and `AWS_THIRD_SECRET_ACCESS_KEY`. |
| `AWS_THIRD_REGION` | Third AWS region for tests requiring multiple regions. Defaults to `us-east-2`. |
| `DX_CONNECTION_ID` | Identifier for Direct Connect Connection testing. |
| `DX_VIRTUAL_INTERFACE_ID` | Identifier for Direct Connect Virtual Interface testing. |
| `EC2_SECURITY_GROUP_RULES_PER_GROUP_LIMIT` | EC2 Quota for Rules per Security Group. Defaults to 50. **DEPRECATED:** Can be augmented or replaced with Service Quotas lookup. |
| `EVENT_BRIDGE_PARTNER_EVENT_BUS_NAME` | Amazon EventBridge partner event bus name. |
| `EVENT_BRIDGE_PARTNER_EVENT_SOURCE_NAME` | Amazon EventBridge partner event source name. |
| `FINSPACE_MANAGED_KX_LICENSE_ENABLED` | Enables tests requiring a license to provision managed KX resources. |
| `GCM_API_KEY` | API Key for Google Cloud Messaging in Pinpoint and SNS Platform Application testing. |
| `GITHUB_TOKEN` | GitHub token for CodePipeline testing. |
| `GLOBALACCERATOR_BYOIP_IPV4_ADDRESS` | IPv4 address from a BYOIP CIDR of AWS Account used for testing Global Accelerator's BYOIP accelerator. |
| `GRAFANA_SSO_GROUP_ID` | AWS SSO group ID for Grafana testing. |
| `GRAFANA_SSO_USER_ID` | AWS SSO user ID for Grafana testing. |
| `MACIE_MEMBER_ACCOUNT_ID` | Identifier of AWS Account for Macie Member testing. **DEPRECATED:** Should be replaced with standard alternate account handling for tests. |
| `QUICKSIGHT_NAMESPACE` | QuickSight namespace name for testing. |
| `QUICKSIGHT_ATHENA_TESTING_ENABLED` | Enable QuickSight tests dependent on Amazon Athena resources. |
| `ROUTE53DOMAINS_DOMAIN_NAME` | Registered domain for Route 53 Domains testing. |
| `RESOURCEEXPLORER_INDEX_TYPE` | Index Type for Resource Explorer 2 Search datasource testing. |
| `SAGEMAKER_IMAGE_VERSION_BASE_IMAGE` | SageMaker base image to use for tests. |
| `SERVICEQUOTAS_INCREASE_ON_CREATE_QUOTA_CODE` | Quota Code for Service Quotas testing (submits support case). |
| `SERVICEQUOTAS_INCREASE_ON_CREATE_SERVICE_CODE` | Service Code for Service Quotas testing (submits support case). |
| `SERVICEQUOTAS_INCREASE_ON_CREATE_VALUE` | Value of quota increase for Service Quotas testing (submits support case). |
| `SES_DOMAIN_IDENTITY_ROOT_DOMAIN` | Root domain name of publicly accessible and Route 53 configurable domain for SES Domain Identity testing. |
| `SES_DEDICATED_IP` | Dedicated IP address for testing IP assignment with a "Standard" (non-managed) SES dedicated IP pool. |
| `SWF_DOMAIN_TESTING_ENABLED` | Enables SWF Domain testing (API does not support deletions). |
| `TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN` | Email address for Organizations Account testing. |
| `TEST_AWS_SES_VERIFIED_EMAIL_ARN` | Verified SES Email Identity for use in Cognito User Pool testing. |
| `TF_ACC` | Enables Go tests containing `resource.Test()` and `resource.ParallelTest()`. |
| `TF_ACC_ASSUME_ROLE_ARN` | Amazon Resource Name of existing IAM Role to use for limited permissions acceptance testing. |
| `TF_AWS_LICENSE_MANAGER_GRANT_HOME_REGION` | Region where a License Manager license is imported. |
| `TF_AWS_LICENSE_MANAGER_GRANT_LICENSE_ARN` | ARN for a License Manager license imported into the current account. |
| `TF_AWS_LICENSE_MANAGER_GRANT_PRINCIPAL` | ARN of a principal to share the License Manager license with. Either a root user, Organization, or Organizational Unit. |
| `TF_TEST_CLOUDFRONT_RETAIN` | Flag to disable but dangle CloudFront Distributions during testing to reduce feedback time (must be manually destroyed afterwards) |
