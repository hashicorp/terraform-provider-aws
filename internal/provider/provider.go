package provider

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	awsbase "github.com/hashicorp/aws-sdk-go-base/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/experimental/nullable"
	"github.com/hashicorp/terraform-provider-aws/internal/service/accessanalyzer"
	"github.com/hashicorp/terraform-provider-aws/internal/service/account"
	"github.com/hashicorp/terraform-provider-aws/internal/service/acm"
	"github.com/hashicorp/terraform-provider-aws/internal/service/acmpca"
	"github.com/hashicorp/terraform-provider-aws/internal/service/amp"
	"github.com/hashicorp/terraform-provider-aws/internal/service/amplify"
	"github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
	"github.com/hashicorp/terraform-provider-aws/internal/service/apigatewayv2"
	"github.com/hashicorp/terraform-provider-aws/internal/service/appautoscaling"
	"github.com/hashicorp/terraform-provider-aws/internal/service/appconfig"
	"github.com/hashicorp/terraform-provider-aws/internal/service/appflow"
	"github.com/hashicorp/terraform-provider-aws/internal/service/appintegrations"
	"github.com/hashicorp/terraform-provider-aws/internal/service/applicationinsights"
	"github.com/hashicorp/terraform-provider-aws/internal/service/appmesh"
	"github.com/hashicorp/terraform-provider-aws/internal/service/apprunner"
	"github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
	"github.com/hashicorp/terraform-provider-aws/internal/service/appsync"
	"github.com/hashicorp/terraform-provider-aws/internal/service/athena"
	"github.com/hashicorp/terraform-provider-aws/internal/service/autoscaling"
	"github.com/hashicorp/terraform-provider-aws/internal/service/autoscalingplans"
	"github.com/hashicorp/terraform-provider-aws/internal/service/backup"
	"github.com/hashicorp/terraform-provider-aws/internal/service/batch"
	"github.com/hashicorp/terraform-provider-aws/internal/service/budgets"
	"github.com/hashicorp/terraform-provider-aws/internal/service/ce"
	"github.com/hashicorp/terraform-provider-aws/internal/service/chime"
	"github.com/hashicorp/terraform-provider-aws/internal/service/cloud9"
	"github.com/hashicorp/terraform-provider-aws/internal/service/cloudcontrol"
	"github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	"github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/internal/service/cloudhsmv2"
	"github.com/hashicorp/terraform-provider-aws/internal/service/cloudsearch"
	"github.com/hashicorp/terraform-provider-aws/internal/service/cloudtrail"
	"github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatch"
	"github.com/hashicorp/terraform-provider-aws/internal/service/codeartifact"
	"github.com/hashicorp/terraform-provider-aws/internal/service/codebuild"
	"github.com/hashicorp/terraform-provider-aws/internal/service/codecommit"
	"github.com/hashicorp/terraform-provider-aws/internal/service/codepipeline"
	"github.com/hashicorp/terraform-provider-aws/internal/service/codestarconnections"
	"github.com/hashicorp/terraform-provider-aws/internal/service/codestarnotifications"
	"github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidentity"
	"github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidp"
	"github.com/hashicorp/terraform-provider-aws/internal/service/configservice"
	"github.com/hashicorp/terraform-provider-aws/internal/service/connect"
	"github.com/hashicorp/terraform-provider-aws/internal/service/cur"
	"github.com/hashicorp/terraform-provider-aws/internal/service/dataexchange"
	"github.com/hashicorp/terraform-provider-aws/internal/service/datapipeline"
	"github.com/hashicorp/terraform-provider-aws/internal/service/datasync"
	"github.com/hashicorp/terraform-provider-aws/internal/service/dax"
	"github.com/hashicorp/terraform-provider-aws/internal/service/deploy"
	"github.com/hashicorp/terraform-provider-aws/internal/service/detective"
	"github.com/hashicorp/terraform-provider-aws/internal/service/devicefarm"
	"github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	"github.com/hashicorp/terraform-provider-aws/internal/service/dlm"
	"github.com/hashicorp/terraform-provider-aws/internal/service/dms"
	"github.com/hashicorp/terraform-provider-aws/internal/service/docdb"
	"github.com/hashicorp/terraform-provider-aws/internal/service/ds"
	"github.com/hashicorp/terraform-provider-aws/internal/service/dynamodb"
	"github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/service/ecr"
	"github.com/hashicorp/terraform-provider-aws/internal/service/ecrpublic"
	"github.com/hashicorp/terraform-provider-aws/internal/service/ecs"
	"github.com/hashicorp/terraform-provider-aws/internal/service/efs"
	"github.com/hashicorp/terraform-provider-aws/internal/service/eks"
	"github.com/hashicorp/terraform-provider-aws/internal/service/elasticache"
	"github.com/hashicorp/terraform-provider-aws/internal/service/elasticbeanstalk"
	"github.com/hashicorp/terraform-provider-aws/internal/service/elasticsearch"
	"github.com/hashicorp/terraform-provider-aws/internal/service/elastictranscoder"
	"github.com/hashicorp/terraform-provider-aws/internal/service/elb"
	"github.com/hashicorp/terraform-provider-aws/internal/service/elbv2"
	"github.com/hashicorp/terraform-provider-aws/internal/service/emr"
	"github.com/hashicorp/terraform-provider-aws/internal/service/emrcontainers"
	"github.com/hashicorp/terraform-provider-aws/internal/service/emrserverless"
	"github.com/hashicorp/terraform-provider-aws/internal/service/events"
	"github.com/hashicorp/terraform-provider-aws/internal/service/firehose"
	"github.com/hashicorp/terraform-provider-aws/internal/service/fms"
	"github.com/hashicorp/terraform-provider-aws/internal/service/fsx"
	"github.com/hashicorp/terraform-provider-aws/internal/service/gamelift"
	"github.com/hashicorp/terraform-provider-aws/internal/service/glacier"
	"github.com/hashicorp/terraform-provider-aws/internal/service/globalaccelerator"
	"github.com/hashicorp/terraform-provider-aws/internal/service/glue"
	"github.com/hashicorp/terraform-provider-aws/internal/service/grafana"
	"github.com/hashicorp/terraform-provider-aws/internal/service/guardduty"
	"github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/internal/service/identitystore"
	"github.com/hashicorp/terraform-provider-aws/internal/service/imagebuilder"
	"github.com/hashicorp/terraform-provider-aws/internal/service/inspector"
	"github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/internal/service/kafka"
	"github.com/hashicorp/terraform-provider-aws/internal/service/kafkaconnect"
	"github.com/hashicorp/terraform-provider-aws/internal/service/kendra"
	"github.com/hashicorp/terraform-provider-aws/internal/service/keyspaces"
	"github.com/hashicorp/terraform-provider-aws/internal/service/kinesis"
	"github.com/hashicorp/terraform-provider-aws/internal/service/kinesisanalytics"
	"github.com/hashicorp/terraform-provider-aws/internal/service/kinesisanalyticsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/service/kinesisvideo"
	"github.com/hashicorp/terraform-provider-aws/internal/service/kms"
	"github.com/hashicorp/terraform-provider-aws/internal/service/lakeformation"
	"github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
	"github.com/hashicorp/terraform-provider-aws/internal/service/lexmodels"
	"github.com/hashicorp/terraform-provider-aws/internal/service/licensemanager"
	"github.com/hashicorp/terraform-provider-aws/internal/service/lightsail"
	"github.com/hashicorp/terraform-provider-aws/internal/service/location"
	"github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/internal/service/macie"
	"github.com/hashicorp/terraform-provider-aws/internal/service/macie2"
	"github.com/hashicorp/terraform-provider-aws/internal/service/mediaconvert"
	"github.com/hashicorp/terraform-provider-aws/internal/service/mediapackage"
	"github.com/hashicorp/terraform-provider-aws/internal/service/mediastore"
	"github.com/hashicorp/terraform-provider-aws/internal/service/memorydb"
	"github.com/hashicorp/terraform-provider-aws/internal/service/meta"
	"github.com/hashicorp/terraform-provider-aws/internal/service/mq"
	"github.com/hashicorp/terraform-provider-aws/internal/service/mwaa"
	"github.com/hashicorp/terraform-provider-aws/internal/service/neptune"
	"github.com/hashicorp/terraform-provider-aws/internal/service/networkfirewall"
	"github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
	"github.com/hashicorp/terraform-provider-aws/internal/service/opensearch"
	"github.com/hashicorp/terraform-provider-aws/internal/service/opsworks"
	"github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	"github.com/hashicorp/terraform-provider-aws/internal/service/outposts"
	"github.com/hashicorp/terraform-provider-aws/internal/service/pinpoint"
	"github.com/hashicorp/terraform-provider-aws/internal/service/pricing"
	"github.com/hashicorp/terraform-provider-aws/internal/service/qldb"
	"github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/internal/service/ram"
	"github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/internal/service/redshiftdata"
	"github.com/hashicorp/terraform-provider-aws/internal/service/resourcegroups"
	"github.com/hashicorp/terraform-provider-aws/internal/service/resourcegroupstaggingapi"
	"github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	"github.com/hashicorp/terraform-provider-aws/internal/service/route53domains"
	"github.com/hashicorp/terraform-provider-aws/internal/service/route53recoverycontrolconfig"
	"github.com/hashicorp/terraform-provider-aws/internal/service/route53recoveryreadiness"
	"github.com/hashicorp/terraform-provider-aws/internal/service/route53resolver"
	"github.com/hashicorp/terraform-provider-aws/internal/service/rum"
	"github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
	"github.com/hashicorp/terraform-provider-aws/internal/service/s3outposts"
	"github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/internal/service/schemas"
	"github.com/hashicorp/terraform-provider-aws/internal/service/secretsmanager"
	"github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	"github.com/hashicorp/terraform-provider-aws/internal/service/serverlessrepo"
	"github.com/hashicorp/terraform-provider-aws/internal/service/servicecatalog"
	"github.com/hashicorp/terraform-provider-aws/internal/service/servicediscovery"
	"github.com/hashicorp/terraform-provider-aws/internal/service/servicequotas"
	"github.com/hashicorp/terraform-provider-aws/internal/service/ses"
	"github.com/hashicorp/terraform-provider-aws/internal/service/sfn"
	"github.com/hashicorp/terraform-provider-aws/internal/service/shield"
	"github.com/hashicorp/terraform-provider-aws/internal/service/signer"
	"github.com/hashicorp/terraform-provider-aws/internal/service/simpledb"
	"github.com/hashicorp/terraform-provider-aws/internal/service/sns"
	"github.com/hashicorp/terraform-provider-aws/internal/service/sqs"
	"github.com/hashicorp/terraform-provider-aws/internal/service/ssm"
	"github.com/hashicorp/terraform-provider-aws/internal/service/ssoadmin"
	"github.com/hashicorp/terraform-provider-aws/internal/service/storagegateway"
	"github.com/hashicorp/terraform-provider-aws/internal/service/sts"
	"github.com/hashicorp/terraform-provider-aws/internal/service/swf"
	"github.com/hashicorp/terraform-provider-aws/internal/service/synthetics"
	"github.com/hashicorp/terraform-provider-aws/internal/service/timestreamwrite"
	"github.com/hashicorp/terraform-provider-aws/internal/service/transfer"
	"github.com/hashicorp/terraform-provider-aws/internal/service/waf"
	"github.com/hashicorp/terraform-provider-aws/internal/service/wafregional"
	"github.com/hashicorp/terraform-provider-aws/internal/service/wafv2"
	"github.com/hashicorp/terraform-provider-aws/internal/service/worklink"
	"github.com/hashicorp/terraform-provider-aws/internal/service/workspaces"
	"github.com/hashicorp/terraform-provider-aws/internal/service/xray"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Provider returns a *schema.Provider.
func Provider() *schema.Provider {
	// TODO: Move the validation to this, requires conditional schemas
	// TODO: Move the configuration to this, requires validation

	// The actual provider
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"access_key": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				Description: "The access key for API operations. You can retrieve this\n" +
					"from the 'Security & Credentials' section of the AWS console.",
			},
			"allowed_account_ids": {
				Type:          schema.TypeSet,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Optional:      true,
				ConflictsWith: []string{"forbidden_account_ids"},
				Set:           schema.HashString,
			},
			"assume_role":                   assumeRoleSchema(),
			"assume_role_with_web_identity": assumeRoleWithWebIdentitySchema(),
			"custom_ca_bundle": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "File containing custom root and intermediate certificates. " +
					"Can also be configured using the `AWS_CA_BUNDLE` environment variable. " +
					"(Setting `ca_bundle` in the shared config file is not supported.)",
			},
			"default_tags": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Configuration block with settings to default resource tags across all resources.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tags": {
							Type:        schema.TypeMap,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Resource tags to default across all resources",
						},
					},
				},
			},
			"ec2_metadata_service_endpoint": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "Address of the EC2 metadata service endpoint to use. " +
					"Can also be configured using the `AWS_EC2_METADATA_SERVICE_ENDPOINT` environment variable.",
			},
			"ec2_metadata_service_endpoint_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "Protocol to use with EC2 metadata service endpoint." +
					"Valid values are `IPv4` and `IPv6`. Can also be configured using the `AWS_EC2_METADATA_SERVICE_ENDPOINT_MODE` environment variable.",
			},
			"endpoints": endpointsSchema(),
			"forbidden_account_ids": {
				Type:          schema.TypeSet,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Optional:      true,
				ConflictsWith: []string{"allowed_account_ids"},
				Set:           schema.HashString,
			},
			"http_proxy": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The address of an HTTP proxy to use when accessing the AWS API. " +
					"Can also be configured using the `HTTP_PROXY` or `HTTPS_PROXY` environment variables.",
			},
			"ignore_tags": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Configuration block with settings to ignore resource tags across all resources.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"keys": {
							Type:        schema.TypeSet,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Description: "Resource tag keys to ignore across all resources.",
						},
						"key_prefixes": {
							Type:        schema.TypeSet,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Description: "Resource tag key prefixes to ignore across all resources.",
						},
					},
				},
			},
			"insecure": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "Explicitly allow the provider to perform \"insecure\" SSL requests. If omitted, " +
					"default value is `false`",
			},
			"max_retries": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  25,
				Description: "The maximum number of times an AWS API request is\n" +
					"being executed. If the API request still fails, an error is\n" +
					"thrown.",
			},
			"profile": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				Description: "The profile for API operations. If not set, the default profile\n" +
					"created with `aws configure` will be used.",
			},
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The region where AWS operations will take place. Examples\n" +
					"are us-east-1, us-west-2, etc.", // lintignore:AWSAT003,
			},
			"s3_force_path_style": {
				Type:       schema.TypeBool,
				Optional:   true,
				Default:    false,
				Deprecated: "Use s3_use_path_style instead.",
				Description: "Set this to true to enable the request to use path-style addressing,\n" +
					"i.e., https://s3.amazonaws.com/BUCKET/KEY. By default, the S3 client will\n" +
					"use virtual hosted bucket addressing when possible\n" +
					"(https://BUCKET.s3.amazonaws.com/KEY). Specific to the Amazon S3 service.",
			},
			"s3_use_path_style": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "Set this to true to enable the request to use path-style addressing,\n" +
					"i.e., https://s3.amazonaws.com/BUCKET/KEY. By default, the S3 client will\n" +
					"use virtual hosted bucket addressing when possible\n" +
					"(https://BUCKET.s3.amazonaws.com/KEY). Specific to the Amazon S3 service.",
			},
			"secret_key": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				Description: "The secret key for API operations. You can retrieve this\n" +
					"from the 'Security & Credentials' section of the AWS console.",
			},
			"shared_config_files": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of paths to shared config files. If not set, defaults to [~/.aws/config].",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"shared_credentials_file": {
				Type:          schema.TypeString,
				Optional:      true,
				Default:       "",
				Deprecated:    "Use shared_credentials_files instead.",
				ConflictsWith: []string{"shared_credentials_files"},
				Description:   "The path to the shared credentials file. If not set, defaults to ~/.aws/credentials.",
			},
			"shared_credentials_files": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"shared_credentials_file"},
				Description:   "List of paths to shared credentials files. If not set, defaults to [~/.aws/credentials].",
				Elem:          &schema.Schema{Type: schema.TypeString},
			},
			"skip_credentials_validation": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "Skip the credentials validation via STS API. " +
					"Used for AWS API implementations that do not have STS available/implemented.",
			},
			"skip_get_ec2_platforms": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "Skip getting the supported EC2 platforms. " +
					"Used by users that don't have ec2:DescribeAccountAttributes permissions.",
			},
			"skip_metadata_api_check": {
				Type:         nullable.TypeNullableBool,
				Optional:     true,
				ValidateFunc: nullable.ValidateTypeStringNullableBool,
				Description: "Skip the AWS Metadata API check. " +
					"Used for AWS API implementations that do not have a metadata api endpoint.",
			},
			"skip_region_validation": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "Skip static validation of region name. " +
					"Used by users of alternative AWS-like APIs or users w/ access to regions that are not public (yet).",
			},
			"skip_requesting_account_id": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "Skip requesting the account ID. " +
					"Used for AWS API implementations that do not have IAM/STS API and/or metadata API.",
			},
			"sts_region": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				Description: "The region where AWS STS operations will take place. Examples\n" +
					"are us-east-1 and us-west-2.", // lintignore:AWSAT003,
			},
			"token": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				Description: "session token. A session token is only required if you are\n" +
					"using temporary security credentials.",
			},
			"use_dualstack_endpoint": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Resolve an endpoint with DualStack capability",
			},
			"use_fips_endpoint": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Resolve an endpoint with FIPS capability",
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"aws_acm_certificate": acm.DataSourceCertificate(),

			"aws_acmpca_certificate_authority": acmpca.DataSourceCertificateAuthority(),
			"aws_acmpca_certificate":           acmpca.DataSourceCertificate(),

			"aws_api_gateway_api_key":     apigateway.DataSourceAPIKey(),
			"aws_api_gateway_domain_name": apigateway.DataSourceDomainName(),
			"aws_api_gateway_export":      apigateway.DataSourceExport(),
			"aws_api_gateway_resource":    apigateway.DataSourceResource(),
			"aws_api_gateway_rest_api":    apigateway.DataSourceRestAPI(),
			"aws_api_gateway_sdk":         apigateway.DataSourceSdk(),
			"aws_api_gateway_vpc_link":    apigateway.DataSourceVPCLink(),

			"aws_apigatewayv2_api":    apigatewayv2.DataSourceAPI(),
			"aws_apigatewayv2_apis":   apigatewayv2.DataSourceAPIs(),
			"aws_apigatewayv2_export": apigatewayv2.DataSourceExport(),

			"aws_appmesh_mesh":            appmesh.DataSourceMesh(),
			"aws_appmesh_virtual_service": appmesh.DataSourceVirtualService(),

			"aws_autoscaling_group":    autoscaling.DataSourceGroup(),
			"aws_autoscaling_groups":   autoscaling.DataSourceGroups(),
			"aws_launch_configuration": autoscaling.DataSourceLaunchConfiguration(),

			"aws_backup_framework":   backup.DataSourceFramework(),
			"aws_backup_plan":        backup.DataSourcePlan(),
			"aws_backup_report_plan": backup.DataSourceReportPlan(),
			"aws_backup_selection":   backup.DataSourceSelection(),
			"aws_backup_vault":       backup.DataSourceVault(),

			"aws_batch_compute_environment": batch.DataSourceComputeEnvironment(),
			"aws_batch_job_queue":           batch.DataSourceJobQueue(),
			"aws_batch_scheduling_policy":   batch.DataSourceSchedulingPolicy(),

			"aws_ce_cost_category": ce.DataSourceCostCategory(),
			"aws_ce_tags":          ce.DataSourceTags(),

			"aws_cloudcontrolapi_resource": cloudcontrol.DataSourceResource(),

			"aws_cloudformation_export": cloudformation.DataSourceExport(),
			"aws_cloudformation_stack":  cloudformation.DataSourceStack(),
			"aws_cloudformation_type":   cloudformation.DataSourceType(),

			"aws_cloudfront_cache_policy":                   cloudfront.DataSourceCachePolicy(),
			"aws_cloudfront_distribution":                   cloudfront.DataSourceDistribution(),
			"aws_cloudfront_function":                       cloudfront.DataSourceFunction(),
			"aws_cloudfront_log_delivery_canonical_user_id": cloudfront.DataSourceLogDeliveryCanonicalUserID(),
			"aws_cloudfront_origin_access_identities":       cloudfront.DataSourceOriginAccessIdentities(),
			"aws_cloudfront_origin_access_identity":         cloudfront.DataSourceOriginAccessIdentity(),
			"aws_cloudfront_origin_request_policy":          cloudfront.DataSourceOriginRequestPolicy(),
			"aws_cloudfront_realtime_log_config":            cloudfront.DataSourceRealtimeLogConfig(),
			"aws_cloudfront_response_headers_policy":        cloudfront.DataSourceResponseHeadersPolicy(),

			"aws_cloudhsm_v2_cluster": cloudhsmv2.DataSourceCluster(),

			"aws_cloudtrail_service_account": cloudtrail.DataSourceServiceAccount(),

			"aws_cloudwatch_event_bus":        events.DataSourceBus(),
			"aws_cloudwatch_event_connection": events.DataSourceConnection(),
			"aws_cloudwatch_event_source":     events.DataSourceSource(),

			"aws_cloudwatch_log_group":  logs.DataSourceGroup(),
			"aws_cloudwatch_log_groups": logs.DataSourceGroups(),

			"aws_codeartifact_authorization_token": codeartifact.DataSourceAuthorizationToken(),
			"aws_codeartifact_repository_endpoint": codeartifact.DataSourceRepositoryEndpoint(),

			"aws_codecommit_approval_rule_template": codecommit.DataSourceApprovalRuleTemplate(),
			"aws_codecommit_repository":             codecommit.DataSourceRepository(),

			"aws_codestarconnections_connection": codestarconnections.DataSourceConnection(),

			"aws_cognito_user_pool_client":              cognitoidp.DataSourceUserPoolClient(),
			"aws_cognito_user_pool_clients":             cognitoidp.DataSourceUserPoolClients(),
			"aws_cognito_user_pool_signing_certificate": cognitoidp.DataSourceUserPoolSigningCertificate(),
			"aws_cognito_user_pools":                    cognitoidp.DataSourceUserPools(),

			"aws_connect_bot_association":             connect.DataSourceBotAssociation(),
			"aws_connect_contact_flow":                connect.DataSourceContactFlow(),
			"aws_connect_contact_flow_module":         connect.DataSourceContactFlowModule(),
			"aws_connect_hours_of_operation":          connect.DataSourceHoursOfOperation(),
			"aws_connect_instance":                    connect.DataSourceInstance(),
			"aws_connect_lambda_function_association": connect.DataSourceLambdaFunctionAssociation(),
			"aws_connect_prompt":                      connect.DataSourcePrompt(),
			"aws_connect_queue":                       connect.DataSourceQueue(),
			"aws_connect_quick_connect":               connect.DataSourceQuickConnect(),
			"aws_connect_routing_profile":             connect.DataSourceRoutingProfile(),
			"aws_connect_security_profile":            connect.DataSourceSecurityProfile(),
			"aws_connect_user_hierarchy_structure":    connect.DataSourceUserHierarchyStructure(),

			"aws_cur_report_definition": cur.DataSourceReportDefinition(),

			"aws_datapipeline_pipeline":            datapipeline.DataSourcePipeline(),
			"aws_datapipeline_pipeline_definition": datapipeline.DataSourcePipelineDefinition(),

			"aws_docdb_engine_version":        docdb.DataSourceEngineVersion(),
			"aws_docdb_orderable_db_instance": docdb.DataSourceOrderableDBInstance(),

			"aws_dx_connection": directconnect.DataSourceConnection(),
			"aws_dx_gateway":    directconnect.DataSourceGateway(),
			"aws_dx_location":   directconnect.DataSourceLocation(),
			"aws_dx_locations":  directconnect.DataSourceLocations(),

			"aws_directory_service_directory": ds.DataSourceDirectory(),

			"aws_dynamodb_table": dynamodb.DataSourceTable(),

			"aws_ami":                                        ec2.DataSourceAMI(),
			"aws_ami_ids":                                    ec2.DataSourceAMIIDs(),
			"aws_availability_zone":                          ec2.DataSourceAvailabilityZone(),
			"aws_availability_zones":                         ec2.DataSourceAvailabilityZones(),
			"aws_customer_gateway":                           ec2.DataSourceCustomerGateway(),
			"aws_ebs_default_kms_key":                        ec2.DataSourceEBSDefaultKMSKey(),
			"aws_ebs_encryption_by_default":                  ec2.DataSourceEBSEncryptionByDefault(),
			"aws_ebs_snapshot":                               ec2.DataSourceEBSSnapshot(),
			"aws_ebs_snapshot_ids":                           ec2.DataSourceEBSSnapshotIDs(),
			"aws_ebs_volume":                                 ec2.DataSourceEBSVolume(),
			"aws_ebs_volumes":                                ec2.DataSourceEBSVolumes(),
			"aws_ec2_client_vpn_endpoint":                    ec2.DataSourceClientVPNEndpoint(),
			"aws_ec2_coip_pool":                              ec2.DataSourceCoIPPool(),
			"aws_ec2_coip_pools":                             ec2.DataSourceCoIPPools(),
			"aws_ec2_host":                                   ec2.DataSourceHost(),
			"aws_ec2_instance_type_offering":                 ec2.DataSourceInstanceTypeOffering(),
			"aws_ec2_instance_type_offerings":                ec2.DataSourceInstanceTypeOfferings(),
			"aws_ec2_instance_type":                          ec2.DataSourceInstanceType(),
			"aws_ec2_instance_types":                         ec2.DataSourceInstanceTypes(),
			"aws_ec2_local_gateway_route_table":              ec2.DataSourceLocalGatewayRouteTable(),
			"aws_ec2_local_gateway_route_tables":             ec2.DataSourceLocalGatewayRouteTables(),
			"aws_ec2_local_gateway_virtual_interface":        ec2.DataSourceLocalGatewayVirtualInterface(),
			"aws_ec2_local_gateway_virtual_interface_group":  ec2.DataSourceLocalGatewayVirtualInterfaceGroup(),
			"aws_ec2_local_gateway_virtual_interface_groups": ec2.DataSourceLocalGatewayVirtualInterfaceGroups(),
			"aws_ec2_local_gateway":                          ec2.DataSourceLocalGateway(),
			"aws_ec2_local_gateways":                         ec2.DataSourceLocalGateways(),
			"aws_ec2_managed_prefix_list":                    ec2.DataSourceManagedPrefixList(),
			"aws_ec2_serial_console_access":                  ec2.DataSourceSerialConsoleAccess(),
			"aws_ec2_spot_price":                             ec2.DataSourceSpotPrice(),
			"aws_ec2_transit_gateway":                        ec2.DataSourceTransitGateway(),
			"aws_ec2_transit_gateway_connect":                ec2.DataSourceTransitGatewayConnect(),
			"aws_ec2_transit_gateway_connect_peer":           ec2.DataSourceTransitGatewayConnectPeer(),
			"aws_ec2_transit_gateway_dx_gateway_attachment":  ec2.DataSourceTransitGatewayDxGatewayAttachment(),
			"aws_ec2_transit_gateway_multicast_domain":       ec2.DataSourceTransitGatewayMulticastDomain(),
			"aws_ec2_transit_gateway_peering_attachment":     ec2.DataSourceTransitGatewayPeeringAttachment(),
			"aws_ec2_transit_gateway_route_table":            ec2.DataSourceTransitGatewayRouteTable(),
			"aws_ec2_transit_gateway_route_tables":           ec2.DataSourceTransitGatewayRouteTables(),
			"aws_ec2_transit_gateway_vpc_attachment":         ec2.DataSourceTransitGatewayVPCAttachment(),
			"aws_ec2_transit_gateway_vpc_attachments":        ec2.DataSourceTransitGatewayVPCAttachments(),
			"aws_ec2_transit_gateway_vpn_attachment":         ec2.DataSourceTransitGatewayVPNAttachment(),
			"aws_eip":                                        ec2.DataSourceEIP(),
			"aws_eips":                                       ec2.DataSourceEIPs(),
			"aws_instance":                                   ec2.DataSourceInstance(),
			"aws_instances":                                  ec2.DataSourceInstances(),
			"aws_internet_gateway":                           ec2.DataSourceInternetGateway(),
			"aws_key_pair":                                   ec2.DataSourceKeyPair(),
			"aws_launch_template":                            ec2.DataSourceLaunchTemplate(),
			"aws_nat_gateway":                                ec2.DataSourceNATGateway(),
			"aws_nat_gateways":                               ec2.DataSourceNATGateways(),
			"aws_network_acls":                               ec2.DataSourceNetworkACLs(),
			"aws_network_interface":                          ec2.DataSourceNetworkInterface(),
			"aws_network_interfaces":                         ec2.DataSourceNetworkInterfaces(),
			"aws_prefix_list":                                ec2.DataSourcePrefixList(),
			"aws_route_table":                                ec2.DataSourceRouteTable(),
			"aws_route_tables":                               ec2.DataSourceRouteTables(),
			"aws_route":                                      ec2.DataSourceRoute(),
			"aws_security_group":                             ec2.DataSourceSecurityGroup(),
			"aws_security_groups":                            ec2.DataSourceSecurityGroups(),
			"aws_subnet_ids":                                 ec2.DataSourceSubnetIDs(),
			"aws_subnet":                                     ec2.DataSourceSubnet(),
			"aws_subnets":                                    ec2.DataSourceSubnets(),
			"aws_vpc_dhcp_options":                           ec2.DataSourceVPCDHCPOptions(),
			"aws_vpc_endpoint_service":                       ec2.DataSourceVPCEndpointService(),
			"aws_vpc_endpoint":                               ec2.DataSourceVPCEndpoint(),
			"aws_vpc_ipam_pool":                              ec2.DataSourceIPAMPool(),
			"aws_vpc_ipam_preview_next_cidr":                 ec2.DataSourceIPAMPreviewNextCIDR(),
			"aws_vpc_peering_connection":                     ec2.DataSourceVPCPeeringConnection(),
			"aws_vpc_peering_connections":                    ec2.DataSourceVPCPeeringConnections(),
			"aws_vpc":                                        ec2.DataSourceVPC(),
			"aws_vpcs":                                       ec2.DataSourceVPCs(),
			"aws_vpn_gateway":                                ec2.DataSourceVPNGateway(),

			"aws_ecr_authorization_token": ecr.DataSourceAuthorizationToken(),
			"aws_ecr_image":               ecr.DataSourceImage(),
			"aws_ecr_repository":          ecr.DataSourceRepository(),

			"aws_ecrpublic_authorization_token": ecrpublic.DataSourceAuthorizationToken(),

			"aws_ecs_cluster":              ecs.DataSourceCluster(),
			"aws_ecs_container_definition": ecs.DataSourceContainerDefinition(),
			"aws_ecs_service":              ecs.DataSourceService(),
			"aws_ecs_task_definition":      ecs.DataSourceTaskDefinition(),

			"aws_efs_access_point":  efs.DataSourceAccessPoint(),
			"aws_efs_access_points": efs.DataSourceAccessPoints(),
			"aws_efs_file_system":   efs.DataSourceFileSystem(),
			"aws_efs_mount_target":  efs.DataSourceMountTarget(),

			"aws_eks_addon":         eks.DataSourceAddon(),
			"aws_eks_addon_version": eks.DataSourceAddonVersion(),
			"aws_eks_cluster":       eks.DataSourceCluster(),
			"aws_eks_clusters":      eks.DataSourceClusters(),
			"aws_eks_cluster_auth":  eks.DataSourceClusterAuth(),
			"aws_eks_node_group":    eks.DataSourceNodeGroup(),
			"aws_eks_node_groups":   eks.DataSourceNodeGroups(),

			"aws_elasticache_cluster":           elasticache.DataSourceCluster(),
			"aws_elasticache_replication_group": elasticache.DataSourceReplicationGroup(),
			"aws_elasticache_user":              elasticache.DataSourceUser(),

			"aws_elastic_beanstalk_application":    elasticbeanstalk.DataSourceApplication(),
			"aws_elastic_beanstalk_hosted_zone":    elasticbeanstalk.DataSourceHostedZone(),
			"aws_elastic_beanstalk_solution_stack": elasticbeanstalk.DataSourceSolutionStack(),

			"aws_elasticsearch_domain": elasticsearch.DataSourceDomain(),

			"aws_elb":                 elb.DataSourceLoadBalancer(),
			"aws_elb_hosted_zone_id":  elb.DataSourceHostedZoneID(),
			"aws_elb_service_account": elb.DataSourceServiceAccount(),

			// Adding the Aliases for the ALB -> LB Rename
			"aws_alb":               elbv2.DataSourceLoadBalancer(),
			"aws_alb_listener":      elbv2.DataSourceListener(),
			"aws_alb_target_group":  elbv2.DataSourceTargetGroup(),
			"aws_lb":                elbv2.DataSourceLoadBalancer(),
			"aws_lb_hosted_zone_id": elbv2.DataSourceHostedZoneID(),
			"aws_lb_listener":       elbv2.DataSourceListener(),
			"aws_lb_target_group":   elbv2.DataSourceTargetGroup(),

			"aws_emr_release_labels": emr.DataSourceReleaseLabels(),

			"aws_emrcontainers_virtual_cluster": emrcontainers.DataSourceVirtualCluster(),

			"aws_kinesis_firehose_delivery_stream": firehose.DataSourceDeliveryStream(),

			"aws_globalaccelerator_accelerator": globalaccelerator.DataSourceAccelerator(),

			"aws_glue_connection":                       glue.DataSourceConnection(),
			"aws_glue_data_catalog_encryption_settings": glue.DataSourceDataCatalogEncryptionSettings(),
			"aws_glue_script":                           glue.DataSourceScript(),

			"aws_grafana_workspace": grafana.DataSourceWorkspace(),

			"aws_guardduty_detector": guardduty.DataSourceDetector(),

			"aws_iam_account_alias":           iam.DataSourceAccountAlias(),
			"aws_iam_group":                   iam.DataSourceGroup(),
			"aws_iam_instance_profile":        iam.DataSourceInstanceProfile(),
			"aws_iam_instance_profiles":       iam.DataSourceInstanceProfiles(),
			"aws_iam_openid_connect_provider": iam.DataSourceOpenIDConnectProvider(),
			"aws_iam_policy":                  iam.DataSourcePolicy(),
			"aws_iam_policy_document":         iam.DataSourcePolicyDocument(),
			"aws_iam_role":                    iam.DataSourceRole(),
			"aws_iam_roles":                   iam.DataSourceRoles(),
			"aws_iam_saml_provider":           iam.DataSourceSAMLProvider(),
			"aws_iam_server_certificate":      iam.DataSourceServerCertificate(),
			"aws_iam_session_context":         iam.DataSourceSessionContext(),
			"aws_iam_user":                    iam.DataSourceUser(),
			"aws_iam_user_ssh_key":            iam.DataSourceUserSSHKey(),
			"aws_iam_users":                   iam.DataSourceUsers(),

			"aws_identitystore_group": identitystore.DataSourceGroup(),
			"aws_identitystore_user":  identitystore.DataSourceUser(),

			"aws_imagebuilder_component":                     imagebuilder.DataSourceComponent(),
			"aws_imagebuilder_components":                    imagebuilder.DataSourceComponents(),
			"aws_imagebuilder_container_recipe":              imagebuilder.DataSourceContainerRecipe(),
			"aws_imagebuilder_container_recipes":             imagebuilder.DataSourceContainerRecipes(),
			"aws_imagebuilder_distribution_configuration":    imagebuilder.DataSourceDistributionConfiguration(),
			"aws_imagebuilder_distribution_configurations":   imagebuilder.DataSourceDistributionConfigurations(),
			"aws_imagebuilder_image":                         imagebuilder.DataSourceImage(),
			"aws_imagebuilder_image_pipeline":                imagebuilder.DataSourceImagePipeline(),
			"aws_imagebuilder_image_pipelines":               imagebuilder.DataSourceImagePipelines(),
			"aws_imagebuilder_image_recipe":                  imagebuilder.DataSourceImageRecipe(),
			"aws_imagebuilder_image_recipes":                 imagebuilder.DataSourceImageRecipes(),
			"aws_imagebuilder_infrastructure_configuration":  imagebuilder.DataSourceInfrastructureConfiguration(),
			"aws_imagebuilder_infrastructure_configurations": imagebuilder.DataSourceInfrastructureConfigurations(),

			"aws_inspector_rules_packages": inspector.DataSourceRulesPackages(),

			"aws_iot_endpoint": iot.DataSourceEndpoint(),

			"aws_msk_broker_nodes":  kafka.DataSourceBrokerNodes(),
			"aws_msk_cluster":       kafka.DataSourceCluster(),
			"aws_msk_configuration": kafka.DataSourceConfiguration(),
			"aws_msk_kafka_version": kafka.DataSourceVersion(),

			"aws_mskconnect_connector":            kafkaconnect.DataSourceConnector(),
			"aws_mskconnect_custom_plugin":        kafkaconnect.DataSourceCustomPlugin(),
			"aws_mskconnect_worker_configuration": kafkaconnect.DataSourceWorkerConfiguration(),

			"aws_kendra_faq":   kendra.DataSourceFaq(),
			"aws_kendra_index": kendra.DataSourceIndex(),

			"aws_kinesis_stream":          kinesis.DataSourceStream(),
			"aws_kinesis_stream_consumer": kinesis.DataSourceStreamConsumer(),

			"aws_kms_alias":      kms.DataSourceAlias(),
			"aws_kms_ciphertext": kms.DataSourceCiphertext(),
			"aws_kms_key":        kms.DataSourceKey(),
			"aws_kms_public_key": kms.DataSourcePublicKey(),
			"aws_kms_secret":     kms.DataSourceSecret(),
			"aws_kms_secrets":    kms.DataSourceSecrets(),

			"aws_lakeformation_data_lake_settings": lakeformation.DataSourceDataLakeSettings(),
			"aws_lakeformation_permissions":        lakeformation.DataSourcePermissions(),
			"aws_lakeformation_resource":           lakeformation.DataSourceResource(),

			"aws_lambda_alias":               lambda.DataSourceAlias(),
			"aws_lambda_code_signing_config": lambda.DataSourceCodeSigningConfig(),
			"aws_lambda_function_url":        lambda.DataSourceFunctionURL(),
			"aws_lambda_function":            lambda.DataSourceFunction(),
			"aws_lambda_invocation":          lambda.DataSourceInvocation(),
			"aws_lambda_layer_version":       lambda.DataSourceLayerVersion(),

			"aws_lex_bot":       lexmodels.DataSourceBot(),
			"aws_lex_bot_alias": lexmodels.DataSourceBotAlias(),
			"aws_lex_intent":    lexmodels.DataSourceIntent(),
			"aws_lex_slot_type": lexmodels.DataSourceSlotType(),

			"aws_location_map":         location.DataSourceMap(),
			"aws_location_place_index": location.DataSourcePlaceIndex(),

			"aws_arn":                     meta.DataSourceARN(),
			"aws_billing_service_account": meta.DataSourceBillingServiceAccount(),
			"aws_default_tags":            meta.DataSourceDefaultTags(),
			"aws_ip_ranges":               meta.DataSourceIPRanges(),
			"aws_partition":               meta.DataSourcePartition(),
			"aws_region":                  meta.DataSourceRegion(),
			"aws_regions":                 meta.DataSourceRegions(),
			"aws_service":                 meta.DataSourceService(),

			"aws_memorydb_acl":             memorydb.DataSourceACL(),
			"aws_memorydb_cluster":         memorydb.DataSourceCluster(),
			"aws_memorydb_parameter_group": memorydb.DataSourceParameterGroup(),
			"aws_memorydb_snapshot":        memorydb.DataSourceSnapshot(),
			"aws_memorydb_subnet_group":    memorydb.DataSourceSubnetGroup(),
			"aws_memorydb_user":            memorydb.DataSourceUser(),

			"aws_mq_broker":                         mq.DataSourceBroker(),
			"aws_mq_broker_instance_type_offerings": mq.DataSourceBrokerInstanceTypeOfferings(),

			"aws_neptune_engine_version":        neptune.DataSourceEngineVersion(),
			"aws_neptune_orderable_db_instance": neptune.DataSourceOrderableDBInstance(),

			"aws_networkmanager_connection":                   networkmanager.DataSourceConnection(),
			"aws_networkmanager_connections":                  networkmanager.DataSourceConnections(),
			"aws_networkmanager_core_network_policy_document": networkmanager.DataSourceCoreNetworkPolicyDocument(),
			"aws_networkmanager_device":                       networkmanager.DataSourceDevice(),
			"aws_networkmanager_devices":                      networkmanager.DataSourceDevices(),
			"aws_networkmanager_global_network":               networkmanager.DataSourceGlobalNetwork(),
			"aws_networkmanager_global_networks":              networkmanager.DataSourceGlobalNetworks(),
			"aws_networkmanager_link":                         networkmanager.DataSourceLink(),
			"aws_networkmanager_links":                        networkmanager.DataSourceLinks(),
			"aws_networkmanager_site":                         networkmanager.DataSourceSite(),
			"aws_networkmanager_sites":                        networkmanager.DataSourceSites(),

			"aws_opensearch_domain": opensearch.DataSourceDomain(),

			"aws_organizations_delegated_administrators": organizations.DataSourceDelegatedAdministrators(),
			"aws_organizations_delegated_services":       organizations.DataSourceDelegatedServices(),
			"aws_organizations_organization":             organizations.DataSourceOrganization(),
			"aws_organizations_organizational_units":     organizations.DataSourceOrganizationalUnits(),
			"aws_organizations_resource_tags":            organizations.DataSourceResourceTags(),

			"aws_outposts_asset":                  outposts.DataSourceOutpostAsset(),
			"aws_outposts_assets":                 outposts.DataSourceOutpostAssets(),
			"aws_outposts_outpost":                outposts.DataSourceOutpost(),
			"aws_outposts_outpost_instance_type":  outposts.DataSourceOutpostInstanceType(),
			"aws_outposts_outpost_instance_types": outposts.DataSourceOutpostInstanceTypes(),
			"aws_outposts_outposts":               outposts.DataSourceOutposts(),
			"aws_outposts_site":                   outposts.DataSourceSite(),
			"aws_outposts_sites":                  outposts.DataSourceSites(),

			"aws_pricing_product": pricing.DataSourceProduct(),

			"aws_qldb_ledger": qldb.DataSourceLedger(),

			"aws_ram_resource_share": ram.DataSourceResourceShare(),

			"aws_ses_active_receipt_rule_set": ses.DataSourceActiveReceiptRuleSet(),
			"aws_ses_domain_identity":         ses.DataSourceDomainIdentity(),
			"aws_ses_email_identity":          ses.DataSourceEmailIdentity(),

			"aws_db_cluster_snapshot":       rds.DataSourceClusterSnapshot(),
			"aws_db_event_categories":       rds.DataSourceEventCategories(),
			"aws_db_instance":               rds.DataSourceInstance(),
			"aws_db_proxy":                  rds.DataSourceProxy(),
			"aws_db_snapshot":               rds.DataSourceSnapshot(),
			"aws_db_subnet_group":           rds.DataSourceSubnetGroup(),
			"aws_rds_certificate":           rds.DataSourceCertificate(),
			"aws_rds_cluster":               rds.DataSourceCluster(),
			"aws_rds_engine_version":        rds.DataSourceEngineVersion(),
			"aws_rds_orderable_db_instance": rds.DataSourceOrderableInstance(),

			"aws_redshift_cluster":             redshift.DataSourceCluster(),
			"aws_redshift_cluster_credentials": redshift.DataSourceClusterCredentials(),
			"aws_redshift_orderable_cluster":   redshift.DataSourceOrderableCluster(),
			"aws_redshift_service_account":     redshift.DataSourceServiceAccount(),
			"aws_redshift_subnet_group":        redshift.DataSourceSubnetGroup(),

			"aws_resourcegroupstaggingapi_resources": resourcegroupstaggingapi.DataSourceResources(),

			"aws_route53_delegation_set":          route53.DataSourceDelegationSet(),
			"aws_route53_traffic_policy_document": route53.DataSourceTrafficPolicyDocument(),
			"aws_route53_zone":                    route53.DataSourceZone(),

			"aws_route53_resolver_endpoint": route53resolver.DataSourceEndpoint(),
			"aws_route53_resolver_rule":     route53resolver.DataSourceRule(),
			"aws_route53_resolver_rules":    route53resolver.DataSourceRules(),

			"aws_canonical_user_id": s3.DataSourceCanonicalUserID(),
			"aws_s3_bucket":         s3.DataSourceBucket(),
			"aws_s3_object":         s3.DataSourceObject(),
			"aws_s3_objects":        s3.DataSourceObjects(),
			"aws_s3_bucket_object":  s3.DataSourceBucketObject(),  // DEPRECATED: use aws_s3_object instead
			"aws_s3_bucket_objects": s3.DataSourceBucketObjects(), // DEPRECATED: use aws_s3_objects instead
			"aws_s3_bucket_policy":  s3.DataSourceBucketPolicy(),

			"aws_sagemaker_prebuilt_ecr_image": sagemaker.DataSourcePrebuiltECRImage(),

			"aws_secretsmanager_secret":          secretsmanager.DataSourceSecret(),
			"aws_secretsmanager_secret_rotation": secretsmanager.DataSourceSecretRotation(),
			"aws_secretsmanager_secret_version":  secretsmanager.DataSourceSecretVersion(),
			"aws_secretsmanager_secrets":         secretsmanager.DataSourceSecrets(),

			"aws_serverlessapplicationrepository_application": serverlessrepo.DataSourceApplication(),

			"aws_servicecatalog_constraint":            servicecatalog.DataSourceConstraint(),
			"aws_servicecatalog_launch_paths":          servicecatalog.DataSourceLaunchPaths(),
			"aws_servicecatalog_portfolio_constraints": servicecatalog.DataSourcePortfolioConstraints(),
			"aws_servicecatalog_portfolio":             servicecatalog.DataSourcePortfolio(),
			"aws_servicecatalog_product":               servicecatalog.DataSourceProduct(),

			"aws_service_discovery_dns_namespace": servicediscovery.DataSourceDNSNamespace(),

			"aws_servicequotas_service":       servicequotas.DataSourceService(),
			"aws_servicequotas_service_quota": servicequotas.DataSourceServiceQuota(),

			"aws_sfn_activity":      sfn.DataSourceActivity(),
			"aws_sfn_state_machine": sfn.DataSourceStateMachine(),

			"aws_signer_signing_job":     signer.DataSourceSigningJob(),
			"aws_signer_signing_profile": signer.DataSourceSigningProfile(),

			"aws_sns_topic": sns.DataSourceTopic(),

			"aws_sqs_queue": sqs.DataSourceQueue(),

			"aws_ssm_document":            ssm.DataSourceDocument(),
			"aws_ssm_instances":           ssm.DataSourceInstances(),
			"aws_ssm_maintenance_windows": ssm.DataSourceMaintenanceWindows(),
			"aws_ssm_parameter":           ssm.DataSourceParameter(),
			"aws_ssm_parameters_by_path":  ssm.DataSourceParametersByPath(),
			"aws_ssm_patch_baseline":      ssm.DataSourcePatchBaseline(),

			"aws_ssoadmin_instances":      ssoadmin.DataSourceInstances(),
			"aws_ssoadmin_permission_set": ssoadmin.DataSourcePermissionSet(),

			"aws_storagegateway_local_disk": storagegateway.DataSourceLocalDisk(),

			"aws_caller_identity": sts.DataSourceCallerIdentity(),

			"aws_transfer_server": transfer.DataSourceServer(),

			"aws_waf_ipset":           waf.DataSourceIPSet(),
			"aws_waf_rule":            waf.DataSourceRule(),
			"aws_waf_rate_based_rule": waf.DataSourceRateBasedRule(),
			"aws_waf_web_acl":         waf.DataSourceWebACL(),

			"aws_wafregional_ipset":           wafregional.DataSourceIPSet(),
			"aws_wafregional_rule":            wafregional.DataSourceRule(),
			"aws_wafregional_rate_based_rule": wafregional.DataSourceRateBasedRule(),
			"aws_wafregional_web_acl":         wafregional.DataSourceWebACL(),

			"aws_wafv2_ip_set":            wafv2.DataSourceIPSet(),
			"aws_wafv2_regex_pattern_set": wafv2.DataSourceRegexPatternSet(),
			"aws_wafv2_rule_group":        wafv2.DataSourceRuleGroup(),
			"aws_wafv2_web_acl":           wafv2.DataSourceWebACL(),

			"aws_workspaces_bundle":    workspaces.DataSourceBundle(),
			"aws_workspaces_directory": workspaces.DataSourceDirectory(),
			"aws_workspaces_image":     workspaces.DataSourceImage(),
			"aws_workspaces_workspace": workspaces.DataSourceWorkspace(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"aws_accessanalyzer_analyzer": accessanalyzer.ResourceAnalyzer(),

			"aws_account_alternate_contact": account.ResourceAlternateContact(),

			"aws_acm_certificate":            acm.ResourceCertificate(),
			"aws_acm_certificate_validation": acm.ResourceCertificateValidation(),

			"aws_acmpca_certificate":                       acmpca.ResourceCertificate(),
			"aws_acmpca_certificate_authority":             acmpca.ResourceCertificateAuthority(),
			"aws_acmpca_certificate_authority_certificate": acmpca.ResourceCertificateAuthorityCertificate(),
			"aws_acmpca_policy":                            acmpca.ResourcePolicy(),

			"aws_applicationinsights_application": applicationinsights.ResourceApplication(),

			"aws_prometheus_workspace":                amp.ResourceWorkspace(),
			"aws_prometheus_alert_manager_definition": amp.ResourceAlertManagerDefinition(),
			"aws_prometheus_rule_group_namespace":     amp.ResourceRuleGroupNamespace(),

			"aws_amplify_app":                 amplify.ResourceApp(),
			"aws_amplify_backend_environment": amplify.ResourceBackendEnvironment(),
			"aws_amplify_branch":              amplify.ResourceBranch(),
			"aws_amplify_domain_association":  amplify.ResourceDomainAssociation(),
			"aws_amplify_webhook":             amplify.ResourceWebhook(),

			"aws_api_gateway_account":               apigateway.ResourceAccount(),
			"aws_api_gateway_api_key":               apigateway.ResourceAPIKey(),
			"aws_api_gateway_authorizer":            apigateway.ResourceAuthorizer(),
			"aws_api_gateway_base_path_mapping":     apigateway.ResourceBasePathMapping(),
			"aws_api_gateway_client_certificate":    apigateway.ResourceClientCertificate(),
			"aws_api_gateway_deployment":            apigateway.ResourceDeployment(),
			"aws_api_gateway_documentation_part":    apigateway.ResourceDocumentationPart(),
			"aws_api_gateway_documentation_version": apigateway.ResourceDocumentationVersion(),
			"aws_api_gateway_domain_name":           apigateway.ResourceDomainName(),
			"aws_api_gateway_gateway_response":      apigateway.ResourceGatewayResponse(),
			"aws_api_gateway_integration":           apigateway.ResourceIntegration(),
			"aws_api_gateway_integration_response":  apigateway.ResourceIntegrationResponse(),
			"aws_api_gateway_method":                apigateway.ResourceMethod(),
			"aws_api_gateway_method_response":       apigateway.ResourceMethodResponse(),
			"aws_api_gateway_method_settings":       apigateway.ResourceMethodSettings(),
			"aws_api_gateway_model":                 apigateway.ResourceModel(),
			"aws_api_gateway_request_validator":     apigateway.ResourceRequestValidator(),
			"aws_api_gateway_resource":              apigateway.ResourceResource(),
			"aws_api_gateway_rest_api":              apigateway.ResourceRestAPI(),
			"aws_api_gateway_rest_api_policy":       apigateway.ResourceRestAPIPolicy(),
			"aws_api_gateway_stage":                 apigateway.ResourceStage(),
			"aws_api_gateway_usage_plan":            apigateway.ResourceUsagePlan(),
			"aws_api_gateway_usage_plan_key":        apigateway.ResourceUsagePlanKey(),
			"aws_api_gateway_vpc_link":              apigateway.ResourceVPCLink(),

			"aws_apigatewayv2_api":                  apigatewayv2.ResourceAPI(),
			"aws_apigatewayv2_api_mapping":          apigatewayv2.ResourceAPIMapping(),
			"aws_apigatewayv2_authorizer":           apigatewayv2.ResourceAuthorizer(),
			"aws_apigatewayv2_deployment":           apigatewayv2.ResourceDeployment(),
			"aws_apigatewayv2_domain_name":          apigatewayv2.ResourceDomainName(),
			"aws_apigatewayv2_integration":          apigatewayv2.ResourceIntegration(),
			"aws_apigatewayv2_integration_response": apigatewayv2.ResourceIntegrationResponse(),
			"aws_apigatewayv2_model":                apigatewayv2.ResourceModel(),
			"aws_apigatewayv2_route":                apigatewayv2.ResourceRoute(),
			"aws_apigatewayv2_route_response":       apigatewayv2.ResourceRouteResponse(),
			"aws_apigatewayv2_stage":                apigatewayv2.ResourceStage(),
			"aws_apigatewayv2_vpc_link":             apigatewayv2.ResourceVPCLink(),

			"aws_appconfig_application":                  appconfig.ResourceApplication(),
			"aws_appconfig_configuration_profile":        appconfig.ResourceConfigurationProfile(),
			"aws_appconfig_deployment":                   appconfig.ResourceDeployment(),
			"aws_appconfig_deployment_strategy":          appconfig.ResourceDeploymentStrategy(),
			"aws_appconfig_environment":                  appconfig.ResourceEnvironment(),
			"aws_appconfig_hosted_configuration_version": appconfig.ResourceHostedConfigurationVersion(),

			"aws_appautoscaling_policy":           appautoscaling.ResourcePolicy(),
			"aws_appautoscaling_scheduled_action": appautoscaling.ResourceScheduledAction(),
			"aws_appautoscaling_target":           appautoscaling.ResourceTarget(),

			"aws_appflow_connector_profile": appflow.ResourceConnectorProfile(),
			"aws_appflow_flow":              appflow.ResourceFlow(),

			"aws_appintegrations_event_integration": appintegrations.ResourceEventIntegration(),

			"aws_appmesh_gateway_route":   appmesh.ResourceGatewayRoute(),
			"aws_appmesh_mesh":            appmesh.ResourceMesh(),
			"aws_appmesh_route":           appmesh.ResourceRoute(),
			"aws_appmesh_virtual_gateway": appmesh.ResourceVirtualGateway(),
			"aws_appmesh_virtual_node":    appmesh.ResourceVirtualNode(),
			"aws_appmesh_virtual_router":  appmesh.ResourceVirtualRouter(),
			"aws_appmesh_virtual_service": appmesh.ResourceVirtualService(),

			"aws_apprunner_vpc_connector":                      apprunner.ResourceVPCConnector(),
			"aws_apprunner_auto_scaling_configuration_version": apprunner.ResourceAutoScalingConfigurationVersion(),
			"aws_apprunner_connection":                         apprunner.ResourceConnection(),
			"aws_apprunner_custom_domain_association":          apprunner.ResourceCustomDomainAssociation(),
			"aws_apprunner_service":                            apprunner.ResourceService(),

			"aws_appstream_directory_config":        appstream.ResourceDirectoryConfig(),
			"aws_appstream_fleet":                   appstream.ResourceFleet(),
			"aws_appstream_fleet_stack_association": appstream.ResourceFleetStackAssociation(),
			"aws_appstream_image_builder":           appstream.ResourceImageBuilder(),
			"aws_appstream_stack":                   appstream.ResourceStack(),
			"aws_appstream_user":                    appstream.ResourceUser(),
			"aws_appstream_user_stack_association":  appstream.ResourceUserStackAssociation(),

			"aws_appsync_api_cache":                   appsync.ResourceAPICache(),
			"aws_appsync_api_key":                     appsync.ResourceAPIKey(),
			"aws_appsync_datasource":                  appsync.ResourceDataSource(),
			"aws_appsync_domain_name":                 appsync.ResourceDomainName(),
			"aws_appsync_domain_name_api_association": appsync.ResourceDomainNameAPIAssociation(),
			"aws_appsync_function":                    appsync.ResourceFunction(),
			"aws_appsync_graphql_api":                 appsync.ResourceGraphQLAPI(),
			"aws_appsync_resolver":                    appsync.ResourceResolver(),

			"aws_athena_database":     athena.ResourceDatabase(),
			"aws_athena_data_catalog": athena.ResourceDataCatalog(),
			"aws_athena_named_query":  athena.ResourceNamedQuery(),
			"aws_athena_workgroup":    athena.ResourceWorkGroup(),

			"aws_autoscaling_attachment":     autoscaling.ResourceAttachment(),
			"aws_autoscaling_group":          autoscaling.ResourceGroup(),
			"aws_autoscaling_group_tag":      autoscaling.ResourceGroupTag(),
			"aws_autoscaling_lifecycle_hook": autoscaling.ResourceLifecycleHook(),
			"aws_autoscaling_notification":   autoscaling.ResourceNotification(),
			"aws_autoscaling_policy":         autoscaling.ResourcePolicy(),
			"aws_autoscaling_schedule":       autoscaling.ResourceSchedule(),
			"aws_launch_configuration":       autoscaling.ResourceLaunchConfiguration(),

			"aws_autoscalingplans_scaling_plan": autoscalingplans.ResourceScalingPlan(),

			"aws_backup_framework":                backup.ResourceFramework(),
			"aws_backup_global_settings":          backup.ResourceGlobalSettings(),
			"aws_backup_plan":                     backup.ResourcePlan(),
			"aws_backup_region_settings":          backup.ResourceRegionSettings(),
			"aws_backup_report_plan":              backup.ResourceReportPlan(),
			"aws_backup_selection":                backup.ResourceSelection(),
			"aws_backup_vault":                    backup.ResourceVault(),
			"aws_backup_vault_lock_configuration": backup.ResourceVaultLockConfiguration(),
			"aws_backup_vault_notifications":      backup.ResourceVaultNotifications(),
			"aws_backup_vault_policy":             backup.ResourceVaultPolicy(),

			"aws_batch_compute_environment": batch.ResourceComputeEnvironment(),
			"aws_batch_job_definition":      batch.ResourceJobDefinition(),
			"aws_batch_job_queue":           batch.ResourceJobQueue(),
			"aws_batch_scheduling_policy":   batch.ResourceSchedulingPolicy(),

			"aws_budgets_budget":        budgets.ResourceBudget(),
			"aws_budgets_budget_action": budgets.ResourceBudgetAction(),

			"aws_ce_anomaly_monitor":      ce.ResourceAnomalyMonitor(),
			"aws_ce_anomaly_subscription": ce.ResourceAnomalySubscription(),
			"aws_ce_cost_allocation_tag":  ce.ResourceCostAllocationTag(),
			"aws_ce_cost_category":        ce.ResourceCostCategory(),

			"aws_chime_voice_connector":                         chime.ResourceVoiceConnector(),
			"aws_chime_voice_connector_group":                   chime.ResourceVoiceConnectorGroup(),
			"aws_chime_voice_connector_logging":                 chime.ResourceVoiceConnectorLogging(),
			"aws_chime_voice_connector_origination":             chime.ResourceVoiceConnectorOrigination(),
			"aws_chime_voice_connector_streaming":               chime.ResourceVoiceConnectorStreaming(),
			"aws_chime_voice_connector_termination":             chime.ResourceVoiceConnectorTermination(),
			"aws_chime_voice_connector_termination_credentials": chime.ResourceVoiceConnectorTerminationCredentials(),

			"aws_cloud9_environment_ec2":        cloud9.ResourceEnvironmentEC2(),
			"aws_cloud9_environment_membership": cloud9.ResourceEnvironmentMembership(),

			"aws_cloudcontrolapi_resource": cloudcontrol.ResourceResource(),

			"aws_cloudformation_stack":              cloudformation.ResourceStack(),
			"aws_cloudformation_stack_set":          cloudformation.ResourceStackSet(),
			"aws_cloudformation_stack_set_instance": cloudformation.ResourceStackSetInstance(),
			"aws_cloudformation_type":               cloudformation.ResourceType(),

			"aws_cloudfront_cache_policy":                   cloudfront.ResourceCachePolicy(),
			"aws_cloudfront_distribution":                   cloudfront.ResourceDistribution(),
			"aws_cloudfront_field_level_encryption_config":  cloudfront.ResourceFieldLevelEncryptionConfig(),
			"aws_cloudfront_field_level_encryption_profile": cloudfront.ResourceFieldLevelEncryptionProfile(),
			"aws_cloudfront_function":                       cloudfront.ResourceFunction(),
			"aws_cloudfront_key_group":                      cloudfront.ResourceKeyGroup(),
			"aws_cloudfront_monitoring_subscription":        cloudfront.ResourceMonitoringSubscription(),
			"aws_cloudfront_origin_access_identity":         cloudfront.ResourceOriginAccessIdentity(),
			"aws_cloudfront_origin_request_policy":          cloudfront.ResourceOriginRequestPolicy(),
			"aws_cloudfront_public_key":                     cloudfront.ResourcePublicKey(),
			"aws_cloudfront_realtime_log_config":            cloudfront.ResourceRealtimeLogConfig(),
			"aws_cloudfront_response_headers_policy":        cloudfront.ResourceResponseHeadersPolicy(),

			"aws_cloudhsm_v2_cluster": cloudhsmv2.ResourceCluster(),
			"aws_cloudhsm_v2_hsm":     cloudhsmv2.ResourceHSM(),

			"aws_cloudsearch_domain":                       cloudsearch.ResourceDomain(),
			"aws_cloudsearch_domain_service_access_policy": cloudsearch.ResourceDomainServiceAccessPolicy(),

			"aws_cloudtrail":                  cloudtrail.ResourceCloudTrail(),
			"aws_cloudtrail_event_data_store": cloudtrail.ResourceEventDataStore(),

			"aws_cloudwatch_composite_alarm": cloudwatch.ResourceCompositeAlarm(),
			"aws_cloudwatch_dashboard":       cloudwatch.ResourceDashboard(),
			"aws_cloudwatch_metric_alarm":    cloudwatch.ResourceMetricAlarm(),
			"aws_cloudwatch_metric_stream":   cloudwatch.ResourceMetricStream(),

			"aws_cloudwatch_event_api_destination": events.ResourceAPIDestination(),
			"aws_cloudwatch_event_archive":         events.ResourceArchive(),
			"aws_cloudwatch_event_bus":             events.ResourceBus(),
			"aws_cloudwatch_event_bus_policy":      events.ResourceBusPolicy(),
			"aws_cloudwatch_event_connection":      events.ResourceConnection(),
			"aws_cloudwatch_event_permission":      events.ResourcePermission(),
			"aws_cloudwatch_event_rule":            events.ResourceRule(),
			"aws_cloudwatch_event_target":          events.ResourceTarget(),

			"aws_cloudwatch_log_destination":         logs.ResourceDestination(),
			"aws_cloudwatch_log_destination_policy":  logs.ResourceDestinationPolicy(),
			"aws_cloudwatch_log_group":               logs.ResourceGroup(),
			"aws_cloudwatch_log_metric_filter":       logs.ResourceMetricFilter(),
			"aws_cloudwatch_log_resource_policy":     logs.ResourceResourcePolicy(),
			"aws_cloudwatch_log_stream":              logs.ResourceStream(),
			"aws_cloudwatch_log_subscription_filter": logs.ResourceSubscriptionFilter(),
			"aws_cloudwatch_query_definition":        logs.ResourceQueryDefinition(),

			"aws_rum_app_monitor": rum.ResourceAppMonitor(),

			"aws_codeartifact_domain":                        codeartifact.ResourceDomain(),
			"aws_codeartifact_domain_permissions_policy":     codeartifact.ResourceDomainPermissionsPolicy(),
			"aws_codeartifact_repository":                    codeartifact.ResourceRepository(),
			"aws_codeartifact_repository_permissions_policy": codeartifact.ResourceRepositoryPermissionsPolicy(),

			"aws_codebuild_project":           codebuild.ResourceProject(),
			"aws_codebuild_resource_policy":   codebuild.ResourceResourcePolicy(),
			"aws_codebuild_report_group":      codebuild.ResourceReportGroup(),
			"aws_codebuild_source_credential": codebuild.ResourceSourceCredential(),
			"aws_codebuild_webhook":           codebuild.ResourceWebhook(),

			"aws_codecommit_approval_rule_template":             codecommit.ResourceApprovalRuleTemplate(),
			"aws_codecommit_approval_rule_template_association": codecommit.ResourceApprovalRuleTemplateAssociation(),
			"aws_codecommit_repository":                         codecommit.ResourceRepository(),
			"aws_codecommit_trigger":                            codecommit.ResourceTrigger(),

			"aws_codedeploy_app":               deploy.ResourceApp(),
			"aws_codedeploy_deployment_config": deploy.ResourceDeploymentConfig(),
			"aws_codedeploy_deployment_group":  deploy.ResourceDeploymentGroup(),

			"aws_codepipeline":         codepipeline.ResourceCodePipeline(),
			"aws_codepipeline_webhook": codepipeline.ResourceWebhook(),

			"aws_codestarconnections_connection": codestarconnections.ResourceConnection(),
			"aws_codestarconnections_host":       codestarconnections.ResourceHost(),

			"aws_codestarnotifications_notification_rule": codestarnotifications.ResourceNotificationRule(),

			"aws_cognito_identity_pool":                        cognitoidentity.ResourcePool(),
			"aws_cognito_identity_pool_provider_principal_tag": cognitoidentity.ResourcePoolProviderPrincipalTag(),
			"aws_cognito_identity_pool_roles_attachment":       cognitoidentity.ResourcePoolRolesAttachment(),

			"aws_cognito_identity_provider":          cognitoidp.ResourceIdentityProvider(),
			"aws_cognito_resource_server":            cognitoidp.ResourceResourceServer(),
			"aws_cognito_risk_configuration":         cognitoidp.ResourceRiskConfiguration(),
			"aws_cognito_user":                       cognitoidp.ResourceUser(),
			"aws_cognito_user_group":                 cognitoidp.ResourceUserGroup(),
			"aws_cognito_user_in_group":              cognitoidp.ResourceUserInGroup(),
			"aws_cognito_user_pool":                  cognitoidp.ResourceUserPool(),
			"aws_cognito_user_pool_client":           cognitoidp.ResourceUserPoolClient(),
			"aws_cognito_user_pool_domain":           cognitoidp.ResourceUserPoolDomain(),
			"aws_cognito_user_pool_ui_customization": cognitoidp.ResourceUserPoolUICustomization(),

			"aws_config_aggregate_authorization":       configservice.ResourceAggregateAuthorization(),
			"aws_config_config_rule":                   configservice.ResourceConfigRule(),
			"aws_config_configuration_aggregator":      configservice.ResourceConfigurationAggregator(),
			"aws_config_configuration_recorder":        configservice.ResourceConfigurationRecorder(),
			"aws_config_configuration_recorder_status": configservice.ResourceConfigurationRecorderStatus(),
			"aws_config_conformance_pack":              configservice.ResourceConformancePack(),
			"aws_config_delivery_channel":              configservice.ResourceDeliveryChannel(),
			"aws_config_organization_conformance_pack": configservice.ResourceOrganizationConformancePack(),
			"aws_config_organization_custom_rule":      configservice.ResourceOrganizationCustomRule(),
			"aws_config_organization_managed_rule":     configservice.ResourceOrganizationManagedRule(),
			"aws_config_remediation_configuration":     configservice.ResourceRemediationConfiguration(),

			"aws_connect_bot_association":             connect.ResourceBotAssociation(),
			"aws_connect_contact_flow":                connect.ResourceContactFlow(),
			"aws_connect_contact_flow_module":         connect.ResourceContactFlowModule(),
			"aws_connect_instance":                    connect.ResourceInstance(),
			"aws_connect_hours_of_operation":          connect.ResourceHoursOfOperation(),
			"aws_connect_lambda_function_association": connect.ResourceLambdaFunctionAssociation(),
			"aws_connect_queue":                       connect.ResourceQueue(),
			"aws_connect_quick_connect":               connect.ResourceQuickConnect(),
			"aws_connect_routing_profile":             connect.ResourceRoutingProfile(),
			"aws_connect_security_profile":            connect.ResourceSecurityProfile(),
			"aws_connect_user_hierarchy_group":        connect.ResourceUserHierarchyGroup(),
			"aws_connect_user_hierarchy_structure":    connect.ResourceUserHierarchyStructure(),

			"aws_cur_report_definition": cur.ResourceReportDefinition(),

			"aws_dataexchange_data_set": dataexchange.ResourceDataSet(),
			"aws_dataexchange_revision": dataexchange.ResourceRevision(),

			"aws_datapipeline_pipeline":            datapipeline.ResourcePipeline(),
			"aws_datapipeline_pipeline_definition": datapipeline.ResourcePipelineDefinition(),

			"aws_datasync_agent":                            datasync.ResourceAgent(),
			"aws_datasync_location_efs":                     datasync.ResourceLocationEFS(),
			"aws_datasync_location_fsx_lustre_file_system":  datasync.ResourceLocationFSxLustreFileSystem(),
			"aws_datasync_location_fsx_openzfs_file_system": datasync.ResourceLocationFSxOpenZFSFileSystem(),
			"aws_datasync_location_fsx_windows_file_system": datasync.ResourceLocationFSxWindowsFileSystem(),
			"aws_datasync_location_hdfs":                    datasync.ResourceLocationHDFS(),
			"aws_datasync_location_nfs":                     datasync.ResourceLocationNFS(),
			"aws_datasync_location_s3":                      datasync.ResourceLocationS3(),
			"aws_datasync_location_smb":                     datasync.ResourceLocationSMB(),
			"aws_datasync_task":                             datasync.ResourceTask(),

			"aws_dax_cluster":         dax.ResourceCluster(),
			"aws_dax_parameter_group": dax.ResourceParameterGroup(),
			"aws_dax_subnet_group":    dax.ResourceSubnetGroup(),

			"aws_devicefarm_device_pool":       devicefarm.ResourceDevicePool(),
			"aws_devicefarm_instance_profile":  devicefarm.ResourceInstanceProfile(),
			"aws_devicefarm_network_profile":   devicefarm.ResourceNetworkProfile(),
			"aws_devicefarm_project":           devicefarm.ResourceProject(),
			"aws_devicefarm_test_grid_project": devicefarm.ResourceTestGridProject(),
			"aws_devicefarm_upload":            devicefarm.ResourceUpload(),

			"aws_detective_graph":               detective.ResourceGraph(),
			"aws_detective_invitation_accepter": detective.ResourceInvitationAccepter(),
			"aws_detective_member":              detective.ResourceMember(),

			"aws_dx_bgp_peer":                                  directconnect.ResourceBGPPeer(),
			"aws_dx_connection":                                directconnect.ResourceConnection(),
			"aws_dx_connection_association":                    directconnect.ResourceConnectionAssociation(),
			"aws_dx_connection_confirmation":                   directconnect.ResourceConnectionConfirmation(),
			"aws_dx_gateway":                                   directconnect.ResourceGateway(),
			"aws_dx_gateway_association":                       directconnect.ResourceGatewayAssociation(),
			"aws_dx_gateway_association_proposal":              directconnect.ResourceGatewayAssociationProposal(),
			"aws_dx_hosted_connection":                         directconnect.ResourceHostedConnection(),
			"aws_dx_hosted_private_virtual_interface":          directconnect.ResourceHostedPrivateVirtualInterface(),
			"aws_dx_hosted_private_virtual_interface_accepter": directconnect.ResourceHostedPrivateVirtualInterfaceAccepter(),
			"aws_dx_hosted_public_virtual_interface":           directconnect.ResourceHostedPublicVirtualInterface(),
			"aws_dx_hosted_public_virtual_interface_accepter":  directconnect.ResourceHostedPublicVirtualInterfaceAccepter(),
			"aws_dx_hosted_transit_virtual_interface":          directconnect.ResourceHostedTransitVirtualInterface(),
			"aws_dx_hosted_transit_virtual_interface_accepter": directconnect.ResourceHostedTransitVirtualInterfaceAccepter(),
			"aws_dx_lag":                       directconnect.ResourceLag(),
			"aws_dx_private_virtual_interface": directconnect.ResourcePrivateVirtualInterface(),
			"aws_dx_public_virtual_interface":  directconnect.ResourcePublicVirtualInterface(),
			"aws_dx_transit_virtual_interface": directconnect.ResourceTransitVirtualInterface(),

			"aws_dlm_lifecycle_policy": dlm.ResourceLifecyclePolicy(),

			"aws_dms_certificate":              dms.ResourceCertificate(),
			"aws_dms_endpoint":                 dms.ResourceEndpoint(),
			"aws_dms_event_subscription":       dms.ResourceEventSubscription(),
			"aws_dms_replication_instance":     dms.ResourceReplicationInstance(),
			"aws_dms_replication_subnet_group": dms.ResourceReplicationSubnetGroup(),
			"aws_dms_replication_task":         dms.ResourceReplicationTask(),

			"aws_docdb_cluster":                 docdb.ResourceCluster(),
			"aws_docdb_cluster_instance":        docdb.ResourceClusterInstance(),
			"aws_docdb_cluster_parameter_group": docdb.ResourceClusterParameterGroup(),
			"aws_docdb_cluster_snapshot":        docdb.ResourceClusterSnapshot(),
			"aws_docdb_event_subscription":      docdb.ResourceEventSubscription(),
			"aws_docdb_global_cluster":          docdb.ResourceGlobalCluster(),
			"aws_docdb_subnet_group":            docdb.ResourceSubnetGroup(),

			"aws_directory_service_conditional_forwarder": ds.ResourceConditionalForwarder(),
			"aws_directory_service_directory":             ds.ResourceDirectory(),
			"aws_directory_service_log_subscription":      ds.ResourceLogSubscription(),

			"aws_dynamodb_contributor_insights":          dynamodb.ResourceContributorInsights(),
			"aws_dynamodb_global_table":                  dynamodb.ResourceGlobalTable(),
			"aws_dynamodb_kinesis_streaming_destination": dynamodb.ResourceKinesisStreamingDestination(),
			"aws_dynamodb_table":                         dynamodb.ResourceTable(),
			"aws_dynamodb_table_item":                    dynamodb.ResourceTableItem(),
			"aws_dynamodb_tag":                           dynamodb.ResourceTag(),

			"aws_ami":                                              ec2.ResourceAMI(),
			"aws_ami_copy":                                         ec2.ResourceAMICopy(),
			"aws_ami_from_instance":                                ec2.ResourceAMIFromInstance(),
			"aws_ami_launch_permission":                            ec2.ResourceAMILaunchPermission(),
			"aws_customer_gateway":                                 ec2.ResourceCustomerGateway(),
			"aws_default_network_acl":                              ec2.ResourceDefaultNetworkACL(),
			"aws_default_route_table":                              ec2.ResourceDefaultRouteTable(),
			"aws_default_security_group":                           ec2.ResourceDefaultSecurityGroup(),
			"aws_default_subnet":                                   ec2.ResourceDefaultSubnet(),
			"aws_default_vpc":                                      ec2.ResourceDefaultVPC(),
			"aws_default_vpc_dhcp_options":                         ec2.ResourceDefaultVPCDHCPOptions(),
			"aws_ebs_default_kms_key":                              ec2.ResourceEBSDefaultKMSKey(),
			"aws_ebs_encryption_by_default":                        ec2.ResourceEBSEncryptionByDefault(),
			"aws_ebs_snapshot":                                     ec2.ResourceEBSSnapshot(),
			"aws_ebs_snapshot_copy":                                ec2.ResourceEBSSnapshotCopy(),
			"aws_ebs_snapshot_import":                              ec2.ResourceEBSSnapshotImport(),
			"aws_ebs_volume":                                       ec2.ResourceEBSVolume(),
			"aws_ec2_availability_zone_group":                      ec2.ResourceAvailabilityZoneGroup(),
			"aws_ec2_capacity_reservation":                         ec2.ResourceCapacityReservation(),
			"aws_ec2_carrier_gateway":                              ec2.ResourceCarrierGateway(),
			"aws_ec2_client_vpn_authorization_rule":                ec2.ResourceClientVPNAuthorizationRule(),
			"aws_ec2_client_vpn_endpoint":                          ec2.ResourceClientVPNEndpoint(),
			"aws_ec2_client_vpn_network_association":               ec2.ResourceClientVPNNetworkAssociation(),
			"aws_ec2_client_vpn_route":                             ec2.ResourceClientVPNRoute(),
			"aws_ec2_fleet":                                        ec2.ResourceFleet(),
			"aws_ec2_host":                                         ec2.ResourceHost(),
			"aws_ec2_local_gateway_route":                          ec2.ResourceLocalGatewayRoute(),
			"aws_ec2_local_gateway_route_table_vpc_association":    ec2.ResourceLocalGatewayRouteTableVPCAssociation(),
			"aws_ec2_managed_prefix_list":                          ec2.ResourceManagedPrefixList(),
			"aws_ec2_managed_prefix_list_entry":                    ec2.ResourceManagedPrefixListEntry(),
			"aws_ec2_network_insights_path":                        ec2.ResourceNetworkInsightsPath(),
			"aws_ec2_serial_console_access":                        ec2.ResourceSerialConsoleAccess(),
			"aws_ec2_subnet_cidr_reservation":                      ec2.ResourceSubnetCIDRReservation(),
			"aws_ec2_tag":                                          ec2.ResourceTag(),
			"aws_ec2_traffic_mirror_filter":                        ec2.ResourceTrafficMirrorFilter(),
			"aws_ec2_traffic_mirror_filter_rule":                   ec2.ResourceTrafficMirrorFilterRule(),
			"aws_ec2_traffic_mirror_session":                       ec2.ResourceTrafficMirrorSession(),
			"aws_ec2_traffic_mirror_target":                        ec2.ResourceTrafficMirrorTarget(),
			"aws_ec2_transit_gateway":                              ec2.ResourceTransitGateway(),
			"aws_ec2_transit_gateway_connect":                      ec2.ResourceTransitGatewayConnect(),
			"aws_ec2_transit_gateway_connect_peer":                 ec2.ResourceTransitGatewayConnectPeer(),
			"aws_ec2_transit_gateway_multicast_domain":             ec2.ResourceTransitGatewayMulticastDomain(),
			"aws_ec2_transit_gateway_multicast_domain_association": ec2.ResourceTransitGatewayMulticastDomainAssociation(),
			"aws_ec2_transit_gateway_multicast_group_member":       ec2.ResourceTransitGatewayMulticastGroupMember(),
			"aws_ec2_transit_gateway_multicast_group_source":       ec2.ResourceTransitGatewayMulticastGroupSource(),
			"aws_ec2_transit_gateway_peering_attachment":           ec2.ResourceTransitGatewayPeeringAttachment(),
			"aws_ec2_transit_gateway_peering_attachment_accepter":  ec2.ResourceTransitGatewayPeeringAttachmentAccepter(),
			"aws_ec2_transit_gateway_prefix_list_reference":        ec2.ResourceTransitGatewayPrefixListReference(),
			"aws_ec2_transit_gateway_route":                        ec2.ResourceTransitGatewayRoute(),
			"aws_ec2_transit_gateway_route_table":                  ec2.ResourceTransitGatewayRouteTable(),
			"aws_ec2_transit_gateway_route_table_association":      ec2.ResourceTransitGatewayRouteTableAssociation(),
			"aws_ec2_transit_gateway_route_table_propagation":      ec2.ResourceTransitGatewayRouteTablePropagation(),
			"aws_ec2_transit_gateway_vpc_attachment":               ec2.ResourceTransitGatewayVPCAttachment(),
			"aws_ec2_transit_gateway_vpc_attachment_accepter":      ec2.ResourceTransitGatewayVPCAttachmentAccepter(),
			"aws_egress_only_internet_gateway":                     ec2.ResourceEgressOnlyInternetGateway(),
			"aws_eip":                                              ec2.ResourceEIP(),
			"aws_eip_association":                                  ec2.ResourceEIPAssociation(),
			"aws_flow_log":                                         ec2.ResourceFlowLog(),
			"aws_instance":                                         ec2.ResourceInstance(),
			"aws_internet_gateway":                                 ec2.ResourceInternetGateway(),
			"aws_internet_gateway_attachment":                      ec2.ResourceInternetGatewayAttachment(),
			"aws_key_pair":                                         ec2.ResourceKeyPair(),
			"aws_launch_template":                                  ec2.ResourceLaunchTemplate(),
			"aws_main_route_table_association":                     ec2.ResourceMainRouteTableAssociation(),
			"aws_nat_gateway":                                      ec2.ResourceNATGateway(),
			"aws_network_acl":                                      ec2.ResourceNetworkACL(),
			"aws_network_acl_association":                          ec2.ResourceNetworkACLAssociation(),
			"aws_network_acl_rule":                                 ec2.ResourceNetworkACLRule(),
			"aws_network_interface":                                ec2.ResourceNetworkInterface(),
			"aws_network_interface_attachment":                     ec2.ResourceNetworkInterfaceAttachment(),
			"aws_network_interface_sg_attachment":                  ec2.ResourceNetworkInterfaceSGAttachment(),
			"aws_placement_group":                                  ec2.ResourcePlacementGroup(),
			"aws_route":                                            ec2.ResourceRoute(),
			"aws_route_table":                                      ec2.ResourceRouteTable(),
			"aws_route_table_association":                          ec2.ResourceRouteTableAssociation(),
			"aws_security_group":                                   ec2.ResourceSecurityGroup(),
			"aws_security_group_rule":                              ec2.ResourceSecurityGroupRule(),
			"aws_snapshot_create_volume_permission":                ec2.ResourceSnapshotCreateVolumePermission(),
			"aws_spot_datafeed_subscription":                       ec2.ResourceSpotDataFeedSubscription(),
			"aws_spot_fleet_request":                               ec2.ResourceSpotFleetRequest(),
			"aws_spot_instance_request":                            ec2.ResourceSpotInstanceRequest(),
			"aws_subnet":                                           ec2.ResourceSubnet(),
			"aws_volume_attachment":                                ec2.ResourceVolumeAttachment(),
			"aws_vpc":                                              ec2.ResourceVPC(),
			"aws_vpc_dhcp_options":                                 ec2.ResourceVPCDHCPOptions(),
			"aws_vpc_dhcp_options_association":                     ec2.ResourceVPCDHCPOptionsAssociation(),
			"aws_vpc_endpoint":                                     ec2.ResourceVPCEndpoint(),
			"aws_vpc_endpoint_connection_accepter":                 ec2.ResourceVPCEndpointConnectionAccepter(),
			"aws_vpc_endpoint_connection_notification":             ec2.ResourceVPCEndpointConnectionNotification(),
			"aws_vpc_endpoint_policy":                              ec2.ResourceVPCEndpointPolicy(),
			"aws_vpc_endpoint_route_table_association":             ec2.ResourceVPCEndpointRouteTableAssociation(),
			"aws_vpc_endpoint_security_group_association":          ec2.ResourceVPCEndpointSecurityGroupAssociation(),
			"aws_vpc_endpoint_service":                             ec2.ResourceVPCEndpointService(),
			"aws_vpc_endpoint_service_allowed_principal":           ec2.ResourceVPCEndpointServiceAllowedPrincipal(),
			"aws_vpc_endpoint_subnet_association":                  ec2.ResourceVPCEndpointSubnetAssociation(),
			"aws_vpc_ipam":                                         ec2.ResourceIPAM(),
			"aws_vpc_ipam_organization_admin_account":              ec2.ResourceIPAMOrganizationAdminAccount(),
			"aws_vpc_ipam_pool":                                    ec2.ResourceIPAMPool(),
			"aws_vpc_ipam_pool_cidr_allocation":                    ec2.ResourceIPAMPoolCIDRAllocation(),
			"aws_vpc_ipam_pool_cidr":                               ec2.ResourceIPAMPoolCIDR(),
			"aws_vpc_ipam_preview_next_cidr":                       ec2.ResourceIPAMPreviewNextCIDR(),
			"aws_vpc_ipam_scope":                                   ec2.ResourceIPAMScope(),
			"aws_vpc_ipv4_cidr_block_association":                  ec2.ResourceVPCIPv4CIDRBlockAssociation(),
			"aws_vpc_ipv6_cidr_block_association":                  ec2.ResourceVPCIPv6CIDRBlockAssociation(),
			"aws_vpc_peering_connection":                           ec2.ResourceVPCPeeringConnection(),
			"aws_vpc_peering_connection_accepter":                  ec2.ResourceVPCPeeringConnectionAccepter(),
			"aws_vpc_peering_connection_options":                   ec2.ResourceVPCPeeringConnectionOptions(),
			"aws_vpn_connection":                                   ec2.ResourceVPNConnection(),
			"aws_vpn_connection_route":                             ec2.ResourceVPNConnectionRoute(),
			"aws_vpn_gateway":                                      ec2.ResourceVPNGateway(),
			"aws_vpn_gateway_attachment":                           ec2.ResourceVPNGatewayAttachment(),
			"aws_vpn_gateway_route_propagation":                    ec2.ResourceVPNGatewayRoutePropagation(),

			"aws_ecr_lifecycle_policy":                ecr.ResourceLifecyclePolicy(),
			"aws_ecr_pull_through_cache_rule":         ecr.ResourcePullThroughCacheRule(),
			"aws_ecr_registry_policy":                 ecr.ResourceRegistryPolicy(),
			"aws_ecr_registry_scanning_configuration": ecr.ResourceRegistryScanningConfiguration(),
			"aws_ecr_replication_configuration":       ecr.ResourceReplicationConfiguration(),
			"aws_ecr_repository":                      ecr.ResourceRepository(),
			"aws_ecr_repository_policy":               ecr.ResourceRepositoryPolicy(),

			"aws_ecrpublic_repository":        ecrpublic.ResourceRepository(),
			"aws_ecrpublic_repository_policy": ecrpublic.ResourceRepositoryPolicy(),

			"aws_ecs_account_setting_default":    ecs.ResourceAccountSettingDefault(),
			"aws_ecs_capacity_provider":          ecs.ResourceCapacityProvider(),
			"aws_ecs_cluster":                    ecs.ResourceCluster(),
			"aws_ecs_cluster_capacity_providers": ecs.ResourceClusterCapacityProviders(),
			"aws_ecs_service":                    ecs.ResourceService(),
			"aws_ecs_tag":                        ecs.ResourceTag(),
			"aws_ecs_task_definition":            ecs.ResourceTaskDefinition(),
			"aws_ecs_task_set":                   ecs.ResourceTaskSet(),

			"aws_efs_access_point":              efs.ResourceAccessPoint(),
			"aws_efs_backup_policy":             efs.ResourceBackupPolicy(),
			"aws_efs_file_system":               efs.ResourceFileSystem(),
			"aws_efs_file_system_policy":        efs.ResourceFileSystemPolicy(),
			"aws_efs_mount_target":              efs.ResourceMountTarget(),
			"aws_efs_replication_configuration": efs.ResourceReplicationConfiguration(),

			"aws_eks_addon":                    eks.ResourceAddon(),
			"aws_eks_cluster":                  eks.ResourceCluster(),
			"aws_eks_fargate_profile":          eks.ResourceFargateProfile(),
			"aws_eks_identity_provider_config": eks.ResourceIdentityProviderConfig(),
			"aws_eks_node_group":               eks.ResourceNodeGroup(),

			"aws_elasticache_cluster":                  elasticache.ResourceCluster(),
			"aws_elasticache_global_replication_group": elasticache.ResourceGlobalReplicationGroup(),
			"aws_elasticache_parameter_group":          elasticache.ResourceParameterGroup(),
			"aws_elasticache_replication_group":        elasticache.ResourceReplicationGroup(),
			"aws_elasticache_security_group":           elasticache.ResourceSecurityGroup(),
			"aws_elasticache_subnet_group":             elasticache.ResourceSubnetGroup(),
			"aws_elasticache_user":                     elasticache.ResourceUser(),
			"aws_elasticache_user_group":               elasticache.ResourceUserGroup(),
			"aws_elasticache_user_group_association":   elasticache.ResourceUserGroupAssociation(),

			"aws_elastic_beanstalk_application":            elasticbeanstalk.ResourceApplication(),
			"aws_elastic_beanstalk_application_version":    elasticbeanstalk.ResourceApplicationVersion(),
			"aws_elastic_beanstalk_configuration_template": elasticbeanstalk.ResourceConfigurationTemplate(),
			"aws_elastic_beanstalk_environment":            elasticbeanstalk.ResourceEnvironment(),

			"aws_elasticsearch_domain":              elasticsearch.ResourceDomain(),
			"aws_elasticsearch_domain_policy":       elasticsearch.ResourceDomainPolicy(),
			"aws_elasticsearch_domain_saml_options": elasticsearch.ResourceDomainSAMLOptions(),

			"aws_elastictranscoder_pipeline": elastictranscoder.ResourcePipeline(),
			"aws_elastictranscoder_preset":   elastictranscoder.ResourcePreset(),

			"aws_app_cookie_stickiness_policy":        elb.ResourceAppCookieStickinessPolicy(),
			"aws_elb":                                 elb.ResourceLoadBalancer(),
			"aws_elb_attachment":                      elb.ResourceAttachment(),
			"aws_lb_cookie_stickiness_policy":         elb.ResourceCookieStickinessPolicy(),
			"aws_lb_ssl_negotiation_policy":           elb.ResourceSSLNegotiationPolicy(),
			"aws_load_balancer_backend_server_policy": elb.ResourceBackendServerPolicy(),
			"aws_load_balancer_listener_policy":       elb.ResourceListenerPolicy(),
			"aws_load_balancer_policy":                elb.ResourcePolicy(),
			"aws_proxy_protocol_policy":               elb.ResourceProxyProtocolPolicy(),

			"aws_alb":                         elbv2.ResourceLoadBalancer(),
			"aws_alb_listener":                elbv2.ResourceListener(),
			"aws_alb_listener_certificate":    elbv2.ResourceListenerCertificate(),
			"aws_alb_listener_rule":           elbv2.ResourceListenerRule(),
			"aws_alb_target_group":            elbv2.ResourceTargetGroup(),
			"aws_alb_target_group_attachment": elbv2.ResourceTargetGroupAttachment(),
			"aws_lb":                          elbv2.ResourceLoadBalancer(),
			"aws_lb_listener":                 elbv2.ResourceListener(),
			"aws_lb_listener_certificate":     elbv2.ResourceListenerCertificate(),
			"aws_lb_listener_rule":            elbv2.ResourceListenerRule(),
			"aws_lb_target_group":             elbv2.ResourceTargetGroup(),
			"aws_lb_target_group_attachment":  elbv2.ResourceTargetGroupAttachment(),

			"aws_emr_cluster":                emr.ResourceCluster(),
			"aws_emr_instance_fleet":         emr.ResourceInstanceFleet(),
			"aws_emr_instance_group":         emr.ResourceInstanceGroup(),
			"aws_emr_managed_scaling_policy": emr.ResourceManagedScalingPolicy(),
			"aws_emr_security_configuration": emr.ResourceSecurityConfiguration(),
			"aws_emr_studio":                 emr.ResourceStudio(),
			"aws_emr_studio_session_mapping": emr.ResourceStudioSessionMapping(),

			"aws_emrcontainers_virtual_cluster": emrcontainers.ResourceVirtualCluster(),

			"aws_emrserverless_application": emrserverless.ResourceApplication(),

			"aws_kinesis_firehose_delivery_stream": firehose.ResourceDeliveryStream(),

			"aws_fms_admin_account": fms.ResourceAdminAccount(),
			"aws_fms_policy":        fms.ResourcePolicy(),

			"aws_fsx_backup":                        fsx.ResourceBackup(),
			"aws_fsx_lustre_file_system":            fsx.ResourceLustreFileSystem(),
			"aws_fsx_data_repository_association":   fsx.ResourceDataRepositoryAssociation(),
			"aws_fsx_ontap_file_system":             fsx.ResourceOntapFileSystem(),
			"aws_fsx_ontap_storage_virtual_machine": fsx.ResourceOntapStorageVirtualMachine(),
			"aws_fsx_ontap_volume":                  fsx.ResourceOntapVolume(),
			"aws_fsx_openzfs_file_system":           fsx.ResourceOpenzfsFileSystem(),
			"aws_fsx_openzfs_volume":                fsx.ResourceOpenzfsVolume(),
			"aws_fsx_openzfs_snapshot":              fsx.ResourceOpenzfsSnapshot(),
			"aws_fsx_windows_file_system":           fsx.ResourceWindowsFileSystem(),

			"aws_gamelift_alias":              gamelift.ResourceAlias(),
			"aws_gamelift_build":              gamelift.ResourceBuild(),
			"aws_gamelift_fleet":              gamelift.ResourceFleet(),
			"aws_gamelift_game_server_group":  gamelift.ResourceGameServerGroup(),
			"aws_gamelift_game_session_queue": gamelift.ResourceGameSessionQueue(),
			"aws_gamelift_script":             gamelift.ResourceScript(),

			"aws_glacier_vault":      glacier.ResourceVault(),
			"aws_glacier_vault_lock": glacier.ResourceVaultLock(),

			"aws_globalaccelerator_accelerator":    globalaccelerator.ResourceAccelerator(),
			"aws_globalaccelerator_endpoint_group": globalaccelerator.ResourceEndpointGroup(),
			"aws_globalaccelerator_listener":       globalaccelerator.ResourceListener(),

			"aws_glue_catalog_database":                 glue.ResourceCatalogDatabase(),
			"aws_glue_catalog_table":                    glue.ResourceCatalogTable(),
			"aws_glue_classifier":                       glue.ResourceClassifier(),
			"aws_glue_connection":                       glue.ResourceConnection(),
			"aws_glue_crawler":                          glue.ResourceCrawler(),
			"aws_glue_data_catalog_encryption_settings": glue.ResourceDataCatalogEncryptionSettings(),
			"aws_glue_dev_endpoint":                     glue.ResourceDevEndpoint(),
			"aws_glue_job":                              glue.ResourceJob(),
			"aws_glue_ml_transform":                     glue.ResourceMLTransform(),
			"aws_glue_partition":                        glue.ResourcePartition(),
			"aws_glue_partition_index":                  glue.ResourcePartitionIndex(),
			"aws_glue_registry":                         glue.ResourceRegistry(),
			"aws_glue_resource_policy":                  glue.ResourceResourcePolicy(),
			"aws_glue_schema":                           glue.ResourceSchema(),
			"aws_glue_security_configuration":           glue.ResourceSecurityConfiguration(),
			"aws_glue_trigger":                          glue.ResourceTrigger(),
			"aws_glue_user_defined_function":            glue.ResourceUserDefinedFunction(),
			"aws_glue_workflow":                         glue.ResourceWorkflow(),

			"aws_grafana_license_association":          grafana.ResourceLicenseAssociation(),
			"aws_grafana_role_association":             grafana.ResourceRoleAssociation(),
			"aws_grafana_workspace":                    grafana.ResourceWorkspace(),
			"aws_grafana_workspace_saml_configuration": grafana.ResourceWorkspaceSAMLConfiguration(),

			"aws_guardduty_detector":                   guardduty.ResourceDetector(),
			"aws_guardduty_filter":                     guardduty.ResourceFilter(),
			"aws_guardduty_invite_accepter":            guardduty.ResourceInviteAccepter(),
			"aws_guardduty_ipset":                      guardduty.ResourceIPSet(),
			"aws_guardduty_member":                     guardduty.ResourceMember(),
			"aws_guardduty_organization_admin_account": guardduty.ResourceOrganizationAdminAccount(),
			"aws_guardduty_organization_configuration": guardduty.ResourceOrganizationConfiguration(),
			"aws_guardduty_publishing_destination":     guardduty.ResourcePublishingDestination(),
			"aws_guardduty_threatintelset":             guardduty.ResourceThreatintelset(),

			"aws_iam_access_key":                  iam.ResourceAccessKey(),
			"aws_iam_account_alias":               iam.ResourceAccountAlias(),
			"aws_iam_account_password_policy":     iam.ResourceAccountPasswordPolicy(),
			"aws_iam_group":                       iam.ResourceGroup(),
			"aws_iam_group_membership":            iam.ResourceGroupMembership(),
			"aws_iam_group_policy":                iam.ResourceGroupPolicy(),
			"aws_iam_group_policy_attachment":     iam.ResourceGroupPolicyAttachment(),
			"aws_iam_instance_profile":            iam.ResourceInstanceProfile(),
			"aws_iam_openid_connect_provider":     iam.ResourceOpenIDConnectProvider(),
			"aws_iam_policy":                      iam.ResourcePolicy(),
			"aws_iam_policy_attachment":           iam.ResourcePolicyAttachment(),
			"aws_iam_role":                        iam.ResourceRole(),
			"aws_iam_role_policy":                 iam.ResourceRolePolicy(),
			"aws_iam_role_policy_attachment":      iam.ResourceRolePolicyAttachment(),
			"aws_iam_saml_provider":               iam.ResourceSAMLProvider(),
			"aws_iam_server_certificate":          iam.ResourceServerCertificate(),
			"aws_iam_service_linked_role":         iam.ResourceServiceLinkedRole(),
			"aws_iam_service_specific_credential": iam.ResourceServiceSpecificCredential(),
			"aws_iam_signing_certificate":         iam.ResourceSigningCertificate(),
			"aws_iam_user":                        iam.ResourceUser(),
			"aws_iam_user_group_membership":       iam.ResourceUserGroupMembership(),
			"aws_iam_user_login_profile":          iam.ResourceUserLoginProfile(),
			"aws_iam_user_policy":                 iam.ResourceUserPolicy(),
			"aws_iam_user_policy_attachment":      iam.ResourceUserPolicyAttachment(),
			"aws_iam_user_ssh_key":                iam.ResourceUserSSHKey(),
			"aws_iam_virtual_mfa_device":          iam.ResourceVirtualMFADevice(),

			"aws_imagebuilder_component":                    imagebuilder.ResourceComponent(),
			"aws_imagebuilder_container_recipe":             imagebuilder.ResourceContainerRecipe(),
			"aws_imagebuilder_distribution_configuration":   imagebuilder.ResourceDistributionConfiguration(),
			"aws_imagebuilder_image":                        imagebuilder.ResourceImage(),
			"aws_imagebuilder_image_pipeline":               imagebuilder.ResourceImagePipeline(),
			"aws_imagebuilder_image_recipe":                 imagebuilder.ResourceImageRecipe(),
			"aws_imagebuilder_infrastructure_configuration": imagebuilder.ResourceInfrastructureConfiguration(),

			"aws_inspector_assessment_target":   inspector.ResourceAssessmentTarget(),
			"aws_inspector_assessment_template": inspector.ResourceAssessmentTemplate(),
			"aws_inspector_resource_group":      inspector.ResourceResourceGroup(),

			"aws_iot_authorizer":                 iot.ResourceAuthorizer(),
			"aws_iot_certificate":                iot.ResourceCertificate(),
			"aws_iot_indexing_configuration":     iot.ResourceIndexingConfiguration(),
			"aws_iot_logging_options":            iot.ResourceLoggingOptions(),
			"aws_iot_policy":                     iot.ResourcePolicy(),
			"aws_iot_policy_attachment":          iot.ResourcePolicyAttachment(),
			"aws_iot_provisioning_template":      iot.ResourceProvisioningTemplate(),
			"aws_iot_role_alias":                 iot.ResourceRoleAlias(),
			"aws_iot_thing":                      iot.ResourceThing(),
			"aws_iot_thing_group":                iot.ResourceThingGroup(),
			"aws_iot_thing_group_membership":     iot.ResourceThingGroupMembership(),
			"aws_iot_thing_principal_attachment": iot.ResourceThingPrincipalAttachment(),
			"aws_iot_thing_type":                 iot.ResourceThingType(),
			"aws_iot_topic_rule":                 iot.ResourceTopicRule(),
			"aws_iot_topic_rule_destination":     iot.ResourceTopicRuleDestination(),

			"aws_msk_cluster":                  kafka.ResourceCluster(),
			"aws_msk_configuration":            kafka.ResourceConfiguration(),
			"aws_msk_scram_secret_association": kafka.ResourceScramSecretAssociation(),

			"aws_mskconnect_connector":            kafkaconnect.ResourceConnector(),
			"aws_mskconnect_custom_plugin":        kafkaconnect.ResourceCustomPlugin(),
			"aws_mskconnect_worker_configuration": kafkaconnect.ResourceWorkerConfiguration(),

			"aws_kendra_experience":                   kendra.ResourceExperience(),
			"aws_kendra_faq":                          kendra.ResourceFaq(),
			"aws_kendra_index":                        kendra.ResourceIndex(),
			"aws_kendra_query_suggestions_block_list": kendra.ResourceQuerySuggestionsBlockList(),
			"aws_kendra_thesaurus":                    kendra.ResourceThesaurus(),

			"aws_keyspaces_keyspace": keyspaces.ResourceKeyspace(),
			"aws_keyspaces_table":    keyspaces.ResourceTable(),

			"aws_kinesis_stream":          kinesis.ResourceStream(),
			"aws_kinesis_stream_consumer": kinesis.ResourceStreamConsumer(),

			"aws_kinesis_analytics_application":           kinesisanalytics.ResourceApplication(),
			"aws_kinesisanalyticsv2_application":          kinesisanalyticsv2.ResourceApplication(),
			"aws_kinesisanalyticsv2_application_snapshot": kinesisanalyticsv2.ResourceApplicationSnapshot(),

			"aws_kinesis_video_stream": kinesisvideo.ResourceStream(),

			"aws_kms_alias":                kms.ResourceAlias(),
			"aws_kms_ciphertext":           kms.ResourceCiphertext(),
			"aws_kms_external_key":         kms.ResourceExternalKey(),
			"aws_kms_grant":                kms.ResourceGrant(),
			"aws_kms_key":                  kms.ResourceKey(),
			"aws_kms_replica_external_key": kms.ResourceReplicaExternalKey(),
			"aws_kms_replica_key":          kms.ResourceReplicaKey(),

			"aws_lakeformation_data_lake_settings": lakeformation.ResourceDataLakeSettings(),
			"aws_lakeformation_lf_tag":             lakeformation.ResourceLFTag(),
			"aws_lakeformation_permissions":        lakeformation.ResourcePermissions(),
			"aws_lakeformation_resource":           lakeformation.ResourceResource(),

			"aws_lambda_alias":                          lambda.ResourceAlias(),
			"aws_lambda_code_signing_config":            lambda.ResourceCodeSigningConfig(),
			"aws_lambda_event_source_mapping":           lambda.ResourceEventSourceMapping(),
			"aws_lambda_function":                       lambda.ResourceFunction(),
			"aws_lambda_function_event_invoke_config":   lambda.ResourceFunctionEventInvokeConfig(),
			"aws_lambda_function_url":                   lambda.ResourceFunctionURL(),
			"aws_lambda_invocation":                     lambda.ResourceInvocation(),
			"aws_lambda_layer_version":                  lambda.ResourceLayerVersion(),
			"aws_lambda_layer_version_permission":       lambda.ResourceLayerVersionPermission(),
			"aws_lambda_permission":                     lambda.ResourcePermission(),
			"aws_lambda_provisioned_concurrency_config": lambda.ResourceProvisionedConcurrencyConfig(),

			"aws_lex_bot":       lexmodels.ResourceBot(),
			"aws_lex_bot_alias": lexmodels.ResourceBotAlias(),
			"aws_lex_intent":    lexmodels.ResourceIntent(),
			"aws_lex_slot_type": lexmodels.ResourceSlotType(),

			"aws_licensemanager_association":           licensemanager.ResourceAssociation(),
			"aws_licensemanager_license_configuration": licensemanager.ResourceLicenseConfiguration(),

			"aws_lightsail_container_service":                    lightsail.ResourceContainerService(),
			"aws_lightsail_container_service_deployment_version": lightsail.ResourceContainerServiceDeploymentVersion(),
			"aws_lightsail_domain":                               lightsail.ResourceDomain(),
			"aws_lightsail_instance":                             lightsail.ResourceInstance(),
			"aws_lightsail_instance_public_ports":                lightsail.ResourceInstancePublicPorts(),
			"aws_lightsail_key_pair":                             lightsail.ResourceKeyPair(),
			"aws_lightsail_static_ip":                            lightsail.ResourceStaticIP(),
			"aws_lightsail_static_ip_attachment":                 lightsail.ResourceStaticIPAttachment(),

			"aws_location_map":         location.ResourceMap(),
			"aws_location_place_index": location.ResourcePlaceIndex(),
			"aws_location_tracker":     location.ResourceTracker(),

			"aws_macie_member_account_association": macie.ResourceMemberAccountAssociation(),
			"aws_macie_s3_bucket_association":      macie.ResourceS3BucketAssociation(),

			"aws_macie2_account":                    macie2.ResourceAccount(),
			"aws_macie2_classification_job":         macie2.ResourceClassificationJob(),
			"aws_macie2_custom_data_identifier":     macie2.ResourceCustomDataIdentifier(),
			"aws_macie2_findings_filter":            macie2.ResourceFindingsFilter(),
			"aws_macie2_invitation_accepter":        macie2.ResourceInvitationAccepter(),
			"aws_macie2_member":                     macie2.ResourceMember(),
			"aws_macie2_organization_admin_account": macie2.ResourceOrganizationAdminAccount(),

			"aws_media_convert_queue": mediaconvert.ResourceQueue(),

			"aws_media_package_channel": mediapackage.ResourceChannel(),

			"aws_media_store_container":        mediastore.ResourceContainer(),
			"aws_media_store_container_policy": mediastore.ResourceContainerPolicy(),

			"aws_memorydb_acl":             memorydb.ResourceACL(),
			"aws_memorydb_cluster":         memorydb.ResourceCluster(),
			"aws_memorydb_parameter_group": memorydb.ResourceParameterGroup(),
			"aws_memorydb_snapshot":        memorydb.ResourceSnapshot(),
			"aws_memorydb_subnet_group":    memorydb.ResourceSubnetGroup(),
			"aws_memorydb_user":            memorydb.ResourceUser(),

			"aws_mq_broker":        mq.ResourceBroker(),
			"aws_mq_configuration": mq.ResourceConfiguration(),

			"aws_mwaa_environment": mwaa.ResourceEnvironment(),

			"aws_neptune_cluster":                 neptune.ResourceCluster(),
			"aws_neptune_cluster_endpoint":        neptune.ResourceClusterEndpoint(),
			"aws_neptune_cluster_instance":        neptune.ResourceClusterInstance(),
			"aws_neptune_cluster_parameter_group": neptune.ResourceClusterParameterGroup(),
			"aws_neptune_cluster_snapshot":        neptune.ResourceClusterSnapshot(),
			"aws_neptune_event_subscription":      neptune.ResourceEventSubscription(),
			"aws_neptune_parameter_group":         neptune.ResourceParameterGroup(),
			"aws_neptune_subnet_group":            neptune.ResourceSubnetGroup(),

			"aws_networkfirewall_firewall":              networkfirewall.ResourceFirewall(),
			"aws_networkfirewall_firewall_policy":       networkfirewall.ResourceFirewallPolicy(),
			"aws_networkfirewall_logging_configuration": networkfirewall.ResourceLoggingConfiguration(),
			"aws_networkfirewall_resource_policy":       networkfirewall.ResourceResourcePolicy(),
			"aws_networkfirewall_rule_group":            networkfirewall.ResourceRuleGroup(),

			"aws_networkmanager_connection":                               networkmanager.ResourceConnection(),
			"aws_networkmanager_customer_gateway_association":             networkmanager.ResourceCustomerGatewayAssociation(),
			"aws_networkmanager_device":                                   networkmanager.ResourceDevice(),
			"aws_networkmanager_global_network":                           networkmanager.ResourceGlobalNetwork(),
			"aws_networkmanager_link":                                     networkmanager.ResourceLink(),
			"aws_networkmanager_link_association":                         networkmanager.ResourceLinkAssociation(),
			"aws_networkmanager_site":                                     networkmanager.ResourceSite(),
			"aws_networkmanager_transit_gateway_connect_peer_association": networkmanager.ResourceTransitGatewayConnectPeerAssociation(),
			"aws_networkmanager_transit_gateway_registration":             networkmanager.ResourceTransitGatewayRegistration(),

			"aws_opensearch_domain":              opensearch.ResourceDomain(),
			"aws_opensearch_domain_policy":       opensearch.ResourceDomainPolicy(),
			"aws_opensearch_domain_saml_options": opensearch.ResourceDomainSAMLOptions(),

			"aws_opsworks_application":       opsworks.ResourceApplication(),
			"aws_opsworks_custom_layer":      opsworks.ResourceCustomLayer(),
			"aws_opsworks_ecs_cluster_layer": opsworks.ResourceECSClusterLayer(),
			"aws_opsworks_ganglia_layer":     opsworks.ResourceGangliaLayer(),
			"aws_opsworks_haproxy_layer":     opsworks.ResourceHAProxyLayer(),
			"aws_opsworks_instance":          opsworks.ResourceInstance(),
			"aws_opsworks_java_app_layer":    opsworks.ResourceJavaAppLayer(),
			"aws_opsworks_memcached_layer":   opsworks.ResourceMemcachedLayer(),
			"aws_opsworks_mysql_layer":       opsworks.ResourceMySQLLayer(),
			"aws_opsworks_nodejs_app_layer":  opsworks.ResourceNodejsAppLayer(),
			"aws_opsworks_permission":        opsworks.ResourcePermission(),
			"aws_opsworks_php_app_layer":     opsworks.ResourcePHPAppLayer(),
			"aws_opsworks_rails_app_layer":   opsworks.ResourceRailsAppLayer(),
			"aws_opsworks_rds_db_instance":   opsworks.ResourceRDSDBInstance(),
			"aws_opsworks_stack":             opsworks.ResourceStack(),
			"aws_opsworks_static_web_layer":  opsworks.ResourceStaticWebLayer(),
			"aws_opsworks_user_profile":      opsworks.ResourceUserProfile(),

			"aws_organizations_account":                 organizations.ResourceAccount(),
			"aws_organizations_delegated_administrator": organizations.ResourceDelegatedAdministrator(),
			"aws_organizations_organization":            organizations.ResourceOrganization(),
			"aws_organizations_organizational_unit":     organizations.ResourceOrganizationalUnit(),
			"aws_organizations_policy":                  organizations.ResourcePolicy(),
			"aws_organizations_policy_attachment":       organizations.ResourcePolicyAttachment(),

			"aws_pinpoint_adm_channel":               pinpoint.ResourceADMChannel(),
			"aws_pinpoint_apns_channel":              pinpoint.ResourceAPNSChannel(),
			"aws_pinpoint_apns_sandbox_channel":      pinpoint.ResourceAPNSSandboxChannel(),
			"aws_pinpoint_apns_voip_channel":         pinpoint.ResourceAPNSVoIPChannel(),
			"aws_pinpoint_apns_voip_sandbox_channel": pinpoint.ResourceAPNSVoIPSandboxChannel(),
			"aws_pinpoint_app":                       pinpoint.ResourceApp(),
			"aws_pinpoint_baidu_channel":             pinpoint.ResourceBaiduChannel(),
			"aws_pinpoint_email_channel":             pinpoint.ResourceEmailChannel(),
			"aws_pinpoint_event_stream":              pinpoint.ResourceEventStream(),
			"aws_pinpoint_gcm_channel":               pinpoint.ResourceGCMChannel(),
			"aws_pinpoint_sms_channel":               pinpoint.ResourceSMSChannel(),

			"aws_qldb_ledger": qldb.ResourceLedger(),
			"aws_qldb_stream": qldb.ResourceStream(),

			"aws_quicksight_data_source":      quicksight.ResourceDataSource(),
			"aws_quicksight_group":            quicksight.ResourceGroup(),
			"aws_quicksight_group_membership": quicksight.ResourceGroupMembership(),
			"aws_quicksight_user":             quicksight.ResourceUser(),

			"aws_ram_principal_association":   ram.ResourcePrincipalAssociation(),
			"aws_ram_resource_association":    ram.ResourceResourceAssociation(),
			"aws_ram_resource_share":          ram.ResourceResourceShare(),
			"aws_ram_resource_share_accepter": ram.ResourceResourceShareAccepter(),

			"aws_db_cluster_snapshot":                       rds.ResourceClusterSnapshot(),
			"aws_db_event_subscription":                     rds.ResourceEventSubscription(),
			"aws_db_instance":                               rds.ResourceInstance(),
			"aws_db_instance_automated_backups_replication": rds.ResourceInstanceAutomatedBackupsReplication(),
			"aws_db_instance_role_association":              rds.ResourceInstanceRoleAssociation(),
			"aws_db_option_group":                           rds.ResourceOptionGroup(),
			"aws_db_parameter_group":                        rds.ResourceParameterGroup(),
			"aws_db_proxy":                                  rds.ResourceProxy(),
			"aws_db_proxy_default_target_group":             rds.ResourceProxyDefaultTargetGroup(),
			"aws_db_proxy_endpoint":                         rds.ResourceProxyEndpoint(),
			"aws_db_proxy_target":                           rds.ResourceProxyTarget(),
			"aws_db_security_group":                         rds.ResourceSecurityGroup(),
			"aws_db_snapshot":                               rds.ResourceSnapshot(),
			"aws_db_snapshot_copy":                          rds.ResourceSnapshotCopy(),
			"aws_db_subnet_group":                           rds.ResourceSubnetGroup(),
			"aws_rds_cluster":                               rds.ResourceCluster(),
			"aws_rds_cluster_activity_stream":               rds.ResourceClusterActivityStream(),
			"aws_rds_cluster_endpoint":                      rds.ResourceClusterEndpoint(),
			"aws_rds_cluster_instance":                      rds.ResourceClusterInstance(),
			"aws_rds_cluster_parameter_group":               rds.ResourceClusterParameterGroup(),
			"aws_rds_cluster_role_association":              rds.ResourceClusterRoleAssociation(),
			"aws_rds_global_cluster":                        rds.ResourceGlobalCluster(),

			"aws_redshift_authentication_profile":        redshift.ResourceAuthenticationProfile(),
			"aws_redshift_cluster":                       redshift.ResourceCluster(),
			"aws_redshift_cluster_iam_roles":             redshift.ResourceClusterIAMRoles(),
			"aws_redshift_endpoint_access":               redshift.ResourceEndpointAccess(),
			"aws_redshift_event_subscription":            redshift.ResourceEventSubscription(),
			"aws_redshift_hsm_client_certificate":        redshift.ResourceHSMClientCertificate(),
			"aws_redshift_hsm_configuration":             redshift.ResourceHSMConfiguration(),
			"aws_redshift_parameter_group":               redshift.ResourceParameterGroup(),
			"aws_redshift_scheduled_action":              redshift.ResourceScheduledAction(),
			"aws_redshift_security_group":                redshift.ResourceSecurityGroup(),
			"aws_redshift_snapshot_copy_grant":           redshift.ResourceSnapshotCopyGrant(),
			"aws_redshift_snapshot_schedule":             redshift.ResourceSnapshotSchedule(),
			"aws_redshift_snapshot_schedule_association": redshift.ResourceSnapshotScheduleAssociation(),
			"aws_redshift_subnet_group":                  redshift.ResourceSubnetGroup(),
			"aws_redshift_usage_limit":                   redshift.ResourceUsageLimit(),

			"aws_redshiftdata_statement": redshiftdata.ResourceStatement(),

			"aws_resourcegroups_group": resourcegroups.ResourceGroup(),

			"aws_route53_delegation_set":                route53.ResourceDelegationSet(),
			"aws_route53_health_check":                  route53.ResourceHealthCheck(),
			"aws_route53_hosted_zone_dnssec":            route53.ResourceHostedZoneDNSSEC(),
			"aws_route53_key_signing_key":               route53.ResourceKeySigningKey(),
			"aws_route53_query_log":                     route53.ResourceQueryLog(),
			"aws_route53_record":                        route53.ResourceRecord(),
			"aws_route53_traffic_policy":                route53.ResourceTrafficPolicy(),
			"aws_route53_traffic_policy_instance":       route53.ResourceTrafficPolicyInstance(),
			"aws_route53_vpc_association_authorization": route53.ResourceVPCAssociationAuthorization(),
			"aws_route53_zone":                          route53.ResourceZone(),
			"aws_route53_zone_association":              route53.ResourceZoneAssociation(),

			"aws_route53domains_registered_domain": route53domains.ResourceRegisteredDomain(),

			"aws_route53recoverycontrolconfig_cluster":         route53recoverycontrolconfig.ResourceCluster(),
			"aws_route53recoverycontrolconfig_control_panel":   route53recoverycontrolconfig.ResourceControlPanel(),
			"aws_route53recoverycontrolconfig_routing_control": route53recoverycontrolconfig.ResourceRoutingControl(),
			"aws_route53recoverycontrolconfig_safety_rule":     route53recoverycontrolconfig.ResourceSafetyRule(),

			"aws_route53recoveryreadiness_cell":            route53recoveryreadiness.ResourceCell(),
			"aws_route53recoveryreadiness_readiness_check": route53recoveryreadiness.ResourceReadinessCheck(),
			"aws_route53recoveryreadiness_recovery_group":  route53recoveryreadiness.ResourceRecoveryGroup(),
			"aws_route53recoveryreadiness_resource_set":    route53recoveryreadiness.ResourceResourceSet(),

			"aws_route53_resolver_dnssec_config":                   route53resolver.ResourceDNSSECConfig(),
			"aws_route53_resolver_endpoint":                        route53resolver.ResourceEndpoint(),
			"aws_route53_resolver_firewall_config":                 route53resolver.ResourceFirewallConfig(),
			"aws_route53_resolver_firewall_domain_list":            route53resolver.ResourceFirewallDomainList(),
			"aws_route53_resolver_firewall_rule":                   route53resolver.ResourceFirewallRule(),
			"aws_route53_resolver_firewall_rule_group":             route53resolver.ResourceFirewallRuleGroup(),
			"aws_route53_resolver_firewall_rule_group_association": route53resolver.ResourceFirewallRuleGroupAssociation(),
			"aws_route53_resolver_query_log_config":                route53resolver.ResourceQueryLogConfig(),
			"aws_route53_resolver_query_log_config_association":    route53resolver.ResourceQueryLogConfigAssociation(),
			"aws_route53_resolver_rule":                            route53resolver.ResourceRule(),
			"aws_route53_resolver_rule_association":                route53resolver.ResourceRuleAssociation(),

			"aws_s3_bucket":                                      s3.ResourceBucket(),
			"aws_s3_bucket_accelerate_configuration":             s3.ResourceBucketAccelerateConfiguration(),
			"aws_s3_bucket_acl":                                  s3.ResourceBucketACL(),
			"aws_s3_bucket_analytics_configuration":              s3.ResourceBucketAnalyticsConfiguration(),
			"aws_s3_bucket_cors_configuration":                   s3.ResourceBucketCorsConfiguration(),
			"aws_s3_bucket_intelligent_tiering_configuration":    s3.ResourceBucketIntelligentTieringConfiguration(),
			"aws_s3_bucket_inventory":                            s3.ResourceBucketInventory(),
			"aws_s3_bucket_lifecycle_configuration":              s3.ResourceBucketLifecycleConfiguration(),
			"aws_s3_bucket_logging":                              s3.ResourceBucketLogging(),
			"aws_s3_bucket_metric":                               s3.ResourceBucketMetric(),
			"aws_s3_bucket_notification":                         s3.ResourceBucketNotification(),
			"aws_s3_bucket_object_lock_configuration":            s3.ResourceBucketObjectLockConfiguration(),
			"aws_s3_bucket_ownership_controls":                   s3.ResourceBucketOwnershipControls(),
			"aws_s3_bucket_policy":                               s3.ResourceBucketPolicy(),
			"aws_s3_bucket_public_access_block":                  s3.ResourceBucketPublicAccessBlock(),
			"aws_s3_bucket_replication_configuration":            s3.ResourceBucketReplicationConfiguration(),
			"aws_s3_bucket_request_payment_configuration":        s3.ResourceBucketRequestPaymentConfiguration(),
			"aws_s3_bucket_server_side_encryption_configuration": s3.ResourceBucketServerSideEncryptionConfiguration(),
			"aws_s3_bucket_versioning":                           s3.ResourceBucketVersioning(),
			"aws_s3_bucket_website_configuration":                s3.ResourceBucketWebsiteConfiguration(),
			"aws_s3_object":                                      s3.ResourceObject(),
			"aws_s3_object_copy":                                 s3.ResourceObjectCopy(),
			"aws_s3_bucket_object":                               s3.ResourceBucketObject(), // DEPRECATED: use aws_s3_object instead

			"aws_s3_access_point":                             s3control.ResourceAccessPoint(),
			"aws_s3control_access_point_policy":               s3control.ResourceAccessPointPolicy(),
			"aws_s3_account_public_access_block":              s3control.ResourceAccountPublicAccessBlock(),
			"aws_s3control_bucket":                            s3control.ResourceBucket(),
			"aws_s3control_bucket_lifecycle_configuration":    s3control.ResourceBucketLifecycleConfiguration(),
			"aws_s3control_bucket_policy":                     s3control.ResourceBucketPolicy(),
			"aws_s3control_multi_region_access_point":         s3control.ResourceMultiRegionAccessPoint(),
			"aws_s3control_multi_region_access_point_policy":  s3control.ResourceMultiRegionAccessPointPolicy(),
			"aws_s3control_object_lambda_access_point":        s3control.ResourceObjectLambdaAccessPoint(),
			"aws_s3control_object_lambda_access_point_policy": s3control.ResourceObjectLambdaAccessPointPolicy(),

			"aws_s3outposts_endpoint": s3outposts.ResourceEndpoint(),

			"aws_sagemaker_app":                                       sagemaker.ResourceApp(),
			"aws_sagemaker_app_image_config":                          sagemaker.ResourceAppImageConfig(),
			"aws_sagemaker_code_repository":                           sagemaker.ResourceCodeRepository(),
			"aws_sagemaker_device":                                    sagemaker.ResourceDevice(),
			"aws_sagemaker_device_fleet":                              sagemaker.ResourceDeviceFleet(),
			"aws_sagemaker_domain":                                    sagemaker.ResourceDomain(),
			"aws_sagemaker_endpoint":                                  sagemaker.ResourceEndpoint(),
			"aws_sagemaker_endpoint_configuration":                    sagemaker.ResourceEndpointConfiguration(),
			"aws_sagemaker_feature_group":                             sagemaker.ResourceFeatureGroup(),
			"aws_sagemaker_flow_definition":                           sagemaker.ResourceFlowDefinition(),
			"aws_sagemaker_human_task_ui":                             sagemaker.ResourceHumanTaskUI(),
			"aws_sagemaker_image":                                     sagemaker.ResourceImage(),
			"aws_sagemaker_image_version":                             sagemaker.ResourceImageVersion(),
			"aws_sagemaker_model":                                     sagemaker.ResourceModel(),
			"aws_sagemaker_model_package_group":                       sagemaker.ResourceModelPackageGroup(),
			"aws_sagemaker_model_package_group_policy":                sagemaker.ResourceModelPackageGroupPolicy(),
			"aws_sagemaker_notebook_instance":                         sagemaker.ResourceNotebookInstance(),
			"aws_sagemaker_notebook_instance_lifecycle_configuration": sagemaker.ResourceNotebookInstanceLifeCycleConfiguration(),
			"aws_sagemaker_project":                                   sagemaker.ResourceProject(),
			"aws_sagemaker_studio_lifecycle_config":                   sagemaker.ResourceStudioLifecycleConfig(),
			"aws_sagemaker_user_profile":                              sagemaker.ResourceUserProfile(),
			"aws_sagemaker_workforce":                                 sagemaker.ResourceWorkforce(),
			"aws_sagemaker_workteam":                                  sagemaker.ResourceWorkteam(),

			"aws_schemas_discoverer": schemas.ResourceDiscoverer(),
			"aws_schemas_registry":   schemas.ResourceRegistry(),
			"aws_schemas_schema":     schemas.ResourceSchema(),

			"aws_secretsmanager_secret":          secretsmanager.ResourceSecret(),
			"aws_secretsmanager_secret_policy":   secretsmanager.ResourceSecretPolicy(),
			"aws_secretsmanager_secret_rotation": secretsmanager.ResourceSecretRotation(),
			"aws_secretsmanager_secret_version":  secretsmanager.ResourceSecretVersion(),

			"aws_securityhub_account":                    securityhub.ResourceAccount(),
			"aws_securityhub_action_target":              securityhub.ResourceActionTarget(),
			"aws_securityhub_insight":                    securityhub.ResourceInsight(),
			"aws_securityhub_invite_accepter":            securityhub.ResourceInviteAccepter(),
			"aws_securityhub_member":                     securityhub.ResourceMember(),
			"aws_securityhub_organization_admin_account": securityhub.ResourceOrganizationAdminAccount(),
			"aws_securityhub_organization_configuration": securityhub.ResourceOrganizationConfiguration(),
			"aws_securityhub_product_subscription":       securityhub.ResourceProductSubscription(),
			"aws_securityhub_standards_control":          securityhub.ResourceStandardsControl(),
			"aws_securityhub_standards_subscription":     securityhub.ResourceStandardsSubscription(),
			"aws_securityhub_finding_aggregator":         securityhub.ResourceFindingAggregator(),

			"aws_serverlessapplicationrepository_cloudformation_stack": serverlessrepo.ResourceCloudFormationStack(),

			"aws_servicecatalog_budget_resource_association":     servicecatalog.ResourceBudgetResourceAssociation(),
			"aws_servicecatalog_constraint":                      servicecatalog.ResourceConstraint(),
			"aws_servicecatalog_organizations_access":            servicecatalog.ResourceOrganizationsAccess(),
			"aws_servicecatalog_portfolio":                       servicecatalog.ResourcePortfolio(),
			"aws_servicecatalog_portfolio_share":                 servicecatalog.ResourcePortfolioShare(),
			"aws_servicecatalog_principal_portfolio_association": servicecatalog.ResourcePrincipalPortfolioAssociation(),
			"aws_servicecatalog_product":                         servicecatalog.ResourceProduct(),
			"aws_servicecatalog_product_portfolio_association":   servicecatalog.ResourceProductPortfolioAssociation(),
			"aws_servicecatalog_provisioned_product":             servicecatalog.ResourceProvisionedProduct(),
			"aws_servicecatalog_provisioning_artifact":           servicecatalog.ResourceProvisioningArtifact(),
			"aws_servicecatalog_service_action":                  servicecatalog.ResourceServiceAction(),
			"aws_servicecatalog_tag_option":                      servicecatalog.ResourceTagOption(),
			"aws_servicecatalog_tag_option_resource_association": servicecatalog.ResourceTagOptionResourceAssociation(),

			"aws_service_discovery_http_namespace":        servicediscovery.ResourceHTTPNamespace(),
			"aws_service_discovery_instance":              servicediscovery.ResourceInstance(),
			"aws_service_discovery_private_dns_namespace": servicediscovery.ResourcePrivateDNSNamespace(),
			"aws_service_discovery_public_dns_namespace":  servicediscovery.ResourcePublicDNSNamespace(),
			"aws_service_discovery_service":               servicediscovery.ResourceService(),

			"aws_servicequotas_service_quota": servicequotas.ResourceServiceQuota(),

			"aws_ses_active_receipt_rule_set":      ses.ResourceActiveReceiptRuleSet(),
			"aws_ses_configuration_set":            ses.ResourceConfigurationSet(),
			"aws_ses_domain_dkim":                  ses.ResourceDomainDKIM(),
			"aws_ses_domain_identity":              ses.ResourceDomainIdentity(),
			"aws_ses_domain_identity_verification": ses.ResourceDomainIdentityVerification(),
			"aws_ses_domain_mail_from":             ses.ResourceDomainMailFrom(),
			"aws_ses_email_identity":               ses.ResourceEmailIdentity(),
			"aws_ses_event_destination":            ses.ResourceEventDestination(),
			"aws_ses_identity_notification_topic":  ses.ResourceIdentityNotificationTopic(),
			"aws_ses_identity_policy":              ses.ResourceIdentityPolicy(),
			"aws_ses_receipt_filter":               ses.ResourceReceiptFilter(),
			"aws_ses_receipt_rule":                 ses.ResourceReceiptRule(),
			"aws_ses_receipt_rule_set":             ses.ResourceReceiptRuleSet(),
			"aws_ses_template":                     ses.ResourceTemplate(),

			"aws_sfn_activity":      sfn.ResourceActivity(),
			"aws_sfn_state_machine": sfn.ResourceStateMachine(),

			"aws_shield_protection":                          shield.ResourceProtection(),
			"aws_shield_protection_group":                    shield.ResourceProtectionGroup(),
			"aws_shield_protection_health_check_association": shield.ResourceProtectionHealthCheckAssociation(),

			"aws_signer_signing_job":                signer.ResourceSigningJob(),
			"aws_signer_signing_profile":            signer.ResourceSigningProfile(),
			"aws_signer_signing_profile_permission": signer.ResourceSigningProfilePermission(),

			"aws_simpledb_domain": simpledb.ResourceDomain(),

			"aws_sns_platform_application": sns.ResourcePlatformApplication(),
			"aws_sns_sms_preferences":      sns.ResourceSMSPreferences(),
			"aws_sns_topic":                sns.ResourceTopic(),
			"aws_sns_topic_policy":         sns.ResourceTopicPolicy(),
			"aws_sns_topic_subscription":   sns.ResourceTopicSubscription(),

			"aws_sqs_queue":        sqs.ResourceQueue(),
			"aws_sqs_queue_policy": sqs.ResourceQueuePolicy(),

			"aws_ssm_activation":                ssm.ResourceActivation(),
			"aws_ssm_association":               ssm.ResourceAssociation(),
			"aws_ssm_document":                  ssm.ResourceDocument(),
			"aws_ssm_maintenance_window":        ssm.ResourceMaintenanceWindow(),
			"aws_ssm_maintenance_window_target": ssm.ResourceMaintenanceWindowTarget(),
			"aws_ssm_maintenance_window_task":   ssm.ResourceMaintenanceWindowTask(),
			"aws_ssm_parameter":                 ssm.ResourceParameter(),
			"aws_ssm_patch_baseline":            ssm.ResourcePatchBaseline(),
			"aws_ssm_patch_group":               ssm.ResourcePatchGroup(),
			"aws_ssm_resource_data_sync":        ssm.ResourceResourceDataSync(),

			"aws_ssoadmin_account_assignment":           ssoadmin.ResourceAccountAssignment(),
			"aws_ssoadmin_managed_policy_attachment":    ssoadmin.ResourceManagedPolicyAttachment(),
			"aws_ssoadmin_permission_set":               ssoadmin.ResourcePermissionSet(),
			"aws_ssoadmin_permission_set_inline_policy": ssoadmin.ResourcePermissionSetInlinePolicy(),

			"aws_storagegateway_cache":                   storagegateway.ResourceCache(),
			"aws_storagegateway_cached_iscsi_volume":     storagegateway.ResourceCachediSCSIVolume(),
			"aws_storagegateway_file_system_association": storagegateway.ResourceFileSystemAssociation(),
			"aws_storagegateway_gateway":                 storagegateway.ResourceGateway(),
			"aws_storagegateway_nfs_file_share":          storagegateway.ResourceNFSFileShare(),
			"aws_storagegateway_smb_file_share":          storagegateway.ResourceSMBFileShare(),
			"aws_storagegateway_stored_iscsi_volume":     storagegateway.ResourceStorediSCSIVolume(),
			"aws_storagegateway_tape_pool":               storagegateway.ResourceTapePool(),
			"aws_storagegateway_upload_buffer":           storagegateway.ResourceUploadBuffer(),
			"aws_storagegateway_working_storage":         storagegateway.ResourceWorkingStorage(),

			"aws_swf_domain": swf.ResourceDomain(),

			"aws_synthetics_canary": synthetics.ResourceCanary(),

			"aws_timestreamwrite_database": timestreamwrite.ResourceDatabase(),
			"aws_timestreamwrite_table":    timestreamwrite.ResourceTable(),

			"aws_transfer_access":   transfer.ResourceAccess(),
			"aws_transfer_server":   transfer.ResourceServer(),
			"aws_transfer_ssh_key":  transfer.ResourceSSHKey(),
			"aws_transfer_user":     transfer.ResourceUser(),
			"aws_transfer_workflow": transfer.ResourceWorkflow(),

			"aws_waf_byte_match_set":          waf.ResourceByteMatchSet(),
			"aws_waf_geo_match_set":           waf.ResourceGeoMatchSet(),
			"aws_waf_ipset":                   waf.ResourceIPSet(),
			"aws_waf_rate_based_rule":         waf.ResourceRateBasedRule(),
			"aws_waf_regex_match_set":         waf.ResourceRegexMatchSet(),
			"aws_waf_regex_pattern_set":       waf.ResourceRegexPatternSet(),
			"aws_waf_rule":                    waf.ResourceRule(),
			"aws_waf_rule_group":              waf.ResourceRuleGroup(),
			"aws_waf_size_constraint_set":     waf.ResourceSizeConstraintSet(),
			"aws_waf_sql_injection_match_set": waf.ResourceSQLInjectionMatchSet(),
			"aws_waf_web_acl":                 waf.ResourceWebACL(),
			"aws_waf_xss_match_set":           waf.ResourceXSSMatchSet(),

			"aws_wafregional_byte_match_set":          wafregional.ResourceByteMatchSet(),
			"aws_wafregional_geo_match_set":           wafregional.ResourceGeoMatchSet(),
			"aws_wafregional_ipset":                   wafregional.ResourceIPSet(),
			"aws_wafregional_rate_based_rule":         wafregional.ResourceRateBasedRule(),
			"aws_wafregional_regex_match_set":         wafregional.ResourceRegexMatchSet(),
			"aws_wafregional_regex_pattern_set":       wafregional.ResourceRegexPatternSet(),
			"aws_wafregional_rule":                    wafregional.ResourceRule(),
			"aws_wafregional_rule_group":              wafregional.ResourceRuleGroup(),
			"aws_wafregional_size_constraint_set":     wafregional.ResourceSizeConstraintSet(),
			"aws_wafregional_sql_injection_match_set": wafregional.ResourceSQLInjectionMatchSet(),
			"aws_wafregional_web_acl":                 wafregional.ResourceWebACL(),
			"aws_wafregional_web_acl_association":     wafregional.ResourceWebACLAssociation(),
			"aws_wafregional_xss_match_set":           wafregional.ResourceXSSMatchSet(),

			"aws_wafv2_ip_set":                        wafv2.ResourceIPSet(),
			"aws_wafv2_regex_pattern_set":             wafv2.ResourceRegexPatternSet(),
			"aws_wafv2_rule_group":                    wafv2.ResourceRuleGroup(),
			"aws_wafv2_web_acl":                       wafv2.ResourceWebACL(),
			"aws_wafv2_web_acl_association":           wafv2.ResourceWebACLAssociation(),
			"aws_wafv2_web_acl_logging_configuration": wafv2.ResourceWebACLLoggingConfiguration(),

			"aws_worklink_fleet": worklink.ResourceFleet(),
			"aws_worklink_website_certificate_authority_association": worklink.ResourceWebsiteCertificateAuthorityAssociation(),

			"aws_workspaces_directory": workspaces.ResourceDirectory(),
			"aws_workspaces_ip_group":  workspaces.ResourceIPGroup(),
			"aws_workspaces_workspace": workspaces.ResourceWorkspace(),

			"aws_xray_encryption_config": xray.ResourceEncryptionConfig(),
			"aws_xray_group":             xray.ResourceGroup(),
			"aws_xray_sampling_rule":     xray.ResourceSamplingRule(),
		},
	}

	provider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		terraformVersion := provider.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}
		return providerConfigure(ctx, d, terraformVersion)
	}

	return provider
}

func providerConfigure(ctx context.Context, d *schema.ResourceData, terraformVersion string) (interface{}, diag.Diagnostics) {
	config := conns.Config{
		AccessKey:                      d.Get("access_key").(string),
		DefaultTagsConfig:              expandProviderDefaultTags(d.Get("default_tags").([]interface{})),
		CustomCABundle:                 d.Get("custom_ca_bundle").(string),
		EC2MetadataServiceEndpoint:     d.Get("ec2_metadata_service_endpoint").(string),
		EC2MetadataServiceEndpointMode: d.Get("ec2_metadata_service_endpoint_mode").(string),
		Endpoints:                      make(map[string]string),
		HTTPProxy:                      d.Get("http_proxy").(string),
		IgnoreTagsConfig:               expandProviderIgnoreTags(d.Get("ignore_tags").([]interface{})),
		Insecure:                       d.Get("insecure").(bool),
		MaxRetries:                     d.Get("max_retries").(int),
		Profile:                        d.Get("profile").(string),
		Region:                         d.Get("region").(string),
		S3UsePathStyle:                 d.Get("s3_use_path_style").(bool) || d.Get("s3_force_path_style").(bool),
		SecretKey:                      d.Get("secret_key").(string),
		SkipCredsValidation:            d.Get("skip_credentials_validation").(bool),
		SkipGetEC2Platforms:            d.Get("skip_get_ec2_platforms").(bool),
		SkipRegionValidation:           d.Get("skip_region_validation").(bool),
		SkipRequestingAccountId:        d.Get("skip_requesting_account_id").(bool),
		STSRegion:                      d.Get("sts_region").(string),
		TerraformVersion:               terraformVersion,
		Token:                          d.Get("token").(string),
		UseDualStackEndpoint:           d.Get("use_dualstack_endpoint").(bool),
		UseFIPSEndpoint:                d.Get("use_fips_endpoint").(bool),
	}

	if raw := d.Get("shared_config_files").([]interface{}); len(raw) != 0 {
		l := make([]string, len(raw))
		for i, v := range raw {
			l[i] = v.(string)
		}
		config.SharedConfigFiles = l
	}

	if v := d.Get("shared_credentials_file").(string); v != "" {
		config.SharedCredentialsFiles = []string{v}
	}
	if raw := d.Get("shared_credentials_files").([]interface{}); len(raw) != 0 {
		l := make([]string, len(raw))
		for i, v := range raw {
			l[i] = v.(string)
		}
		config.SharedCredentialsFiles = l
	}

	if l, ok := d.Get("assume_role").([]interface{}); ok && len(l) > 0 && l[0] != nil {
		config.AssumeRole = expandAssumeRole(l[0].(map[string]interface{}))
		log.Printf("[INFO] assume_role configuration set: (ARN: %q, SessionID: %q, ExternalID: %q)", config.AssumeRole.RoleARN, config.AssumeRole.SessionName, config.AssumeRole.ExternalID)
	}

	if l, ok := d.Get("assume_role_with_web_identity").([]interface{}); ok && len(l) > 0 && l[0] != nil {
		config.AssumeRoleWithWebIdentity = expandAssumeRoleWithWebIdentity(l[0].(map[string]interface{}))
		log.Printf("[INFO] assume_role_with_web_identity configuration set: (ARN: %q, SessionID: %q)", config.AssumeRoleWithWebIdentity.RoleARN, config.AssumeRoleWithWebIdentity.SessionName)
	}

	endpointsSet := d.Get("endpoints").(*schema.Set)
	if err := expandEndpoints(endpointsSet.List(), config.Endpoints); err != nil {
		return nil, diag.FromErr(err)
	}

	if v, ok := d.GetOk("allowed_account_ids"); ok {
		for _, accountIDRaw := range v.(*schema.Set).List() {
			config.AllowedAccountIds = append(config.AllowedAccountIds, accountIDRaw.(string))
		}
	}

	if v, ok := d.GetOk("forbidden_account_ids"); ok {
		for _, accountIDRaw := range v.(*schema.Set).List() {
			config.ForbiddenAccountIds = append(config.ForbiddenAccountIds, accountIDRaw.(string))
		}
	}

	if v, null, _ := nullable.Bool(d.Get("skip_metadata_api_check").(string)).Value(); !null {
		if v {
			config.EC2MetadataServiceEnableState = imds.ClientDisabled
		} else {
			config.EC2MetadataServiceEnableState = imds.ClientEnabled
		}
	}

	return config.Client(ctx)
}

func assumeRoleSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"duration": {
					Type:          schema.TypeString,
					Optional:      true,
					Description:   "The duration, between 15 minutes and 12 hours, of the role session. Valid time units are ns, us (or µs), ms, s, h, or m.",
					ValidateFunc:  validAssumeRoleDuration,
					ConflictsWith: []string{"assume_role.0.duration_seconds"},
				},
				"duration_seconds": {
					Type:          schema.TypeInt,
					Optional:      true,
					Deprecated:    "Use assume_role.duration instead",
					Description:   "The duration, in seconds, of the role session.",
					ValidateFunc:  validation.IntBetween(900, 43200),
					ConflictsWith: []string{"assume_role.0.duration"},
				},
				"external_id": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "A unique identifier that might be required when you assume a role in another account.",
					ValidateFunc: validation.All(
						validation.StringLenBetween(2, 1224),
						validation.StringMatch(regexp.MustCompile(`[\w+=,.@:\/\-]*`), ""),
					),
				},
				"policy": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "IAM Policy JSON describing further restricting permissions for the IAM Role being assumed.",
					ValidateFunc: validation.StringIsJSON,
				},
				"policy_arns": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "Amazon Resource Names (ARNs) of IAM Policies describing further restricting permissions for the IAM Role being assumed.",
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: verify.ValidARN,
					},
				},
				"role_arn": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "Amazon Resource Name (ARN) of an IAM Role to assume prior to making API calls.",
					ValidateFunc: verify.ValidARN,
				},
				"session_name": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "An identifier for the assumed role session.",
					ValidateFunc: validAssumeRoleSessionName,
				},
				"tags": {
					Type:        schema.TypeMap,
					Optional:    true,
					Description: "Assume role session tags.",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"transitive_tag_keys": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "Assume role session tag keys to pass to any subsequent sessions.",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
			},
		},
	}
}

func assumeRoleWithWebIdentitySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"duration": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "The duration, between 15 minutes and 12 hours, of the role session. Valid time units are ns, us (or µs), ms, s, h, or m.",
					ValidateFunc: validAssumeRoleDuration,
				},
				"policy": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "IAM Policy JSON describing further restricting permissions for the IAM Role being assumed.",
					ValidateFunc: validation.StringIsJSON,
				},
				"policy_arns": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "Amazon Resource Names (ARNs) of IAM Policies describing further restricting permissions for the IAM Role being assumed.",
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: verify.ValidARN,
					},
				},
				"role_arn": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "Amazon Resource Name (ARN) of an IAM Role to assume prior to making API calls.",
					ValidateFunc: verify.ValidARN,
				},
				"session_name": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "An identifier for the assumed role session.",
					ValidateFunc: validAssumeRoleSessionName,
				},
				"web_identity_token": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(4, 20000),
					ExactlyOneOf: []string{"assume_role_with_web_identity.0.web_identity_token", "assume_role_with_web_identity.0.web_identity_token_file"},
				},
				"web_identity_token_file": {
					Type:         schema.TypeString,
					Optional:     true,
					ExactlyOneOf: []string{"assume_role_with_web_identity.0.web_identity_token", "assume_role_with_web_identity.0.web_identity_token_file"},
				},
			},
		},
	}
}

func endpointsSchema() *schema.Schema {
	endpointsAttributes := make(map[string]*schema.Schema)

	for _, serviceKey := range names.Aliases() {
		endpointsAttributes[serviceKey] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Use this to override the default service endpoint URL",
		}
	}

	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: endpointsAttributes,
		},
	}
}

func expandAssumeRole(m map[string]interface{}) *awsbase.AssumeRole {
	assumeRole := awsbase.AssumeRole{}

	if v, ok := m["duration"].(string); ok && v != "" {
		duration, _ := time.ParseDuration(v)
		assumeRole.Duration = duration
	}

	if v, ok := m["duration_seconds"].(int); ok && v != 0 {
		assumeRole.Duration = time.Duration(v) * time.Second
	}

	if v, ok := m["external_id"].(string); ok && v != "" {
		assumeRole.ExternalID = v
	}

	if v, ok := m["policy"].(string); ok && v != "" {
		assumeRole.Policy = v
	}

	if policyARNSet, ok := m["policy_arns"].(*schema.Set); ok && policyARNSet.Len() > 0 {
		for _, policyARNRaw := range policyARNSet.List() {
			policyARN, ok := policyARNRaw.(string)

			if !ok {
				continue
			}

			assumeRole.PolicyARNs = append(assumeRole.PolicyARNs, policyARN)
		}
	}

	if v, ok := m["role_arn"].(string); ok && v != "" {
		assumeRole.RoleARN = v
	}

	if v, ok := m["session_name"].(string); ok && v != "" {
		assumeRole.SessionName = v
	}

	if tagMapRaw, ok := m["tags"].(map[string]interface{}); ok && len(tagMapRaw) > 0 {
		assumeRole.Tags = make(map[string]string)

		for k, vRaw := range tagMapRaw {
			v, ok := vRaw.(string)

			if !ok {
				continue
			}

			assumeRole.Tags[k] = v
		}
	}

	if transitiveTagKeySet, ok := m["transitive_tag_keys"].(*schema.Set); ok && transitiveTagKeySet.Len() > 0 {
		for _, transitiveTagKeyRaw := range transitiveTagKeySet.List() {
			transitiveTagKey, ok := transitiveTagKeyRaw.(string)

			if !ok {
				continue
			}

			assumeRole.TransitiveTagKeys = append(assumeRole.TransitiveTagKeys, transitiveTagKey)
		}
	}

	return &assumeRole
}

func expandAssumeRoleWithWebIdentity(m map[string]interface{}) *awsbase.AssumeRoleWithWebIdentity {
	assumeRole := awsbase.AssumeRoleWithWebIdentity{}

	if v, ok := m["duration"].(string); ok && v != "" {
		duration, _ := time.ParseDuration(v)
		assumeRole.Duration = duration
	}

	if v, ok := m["duration_seconds"].(int); ok && v != 0 {
		assumeRole.Duration = time.Duration(v) * time.Second
	}

	if v, ok := m["policy"].(string); ok && v != "" {
		assumeRole.Policy = v
	}

	if policyARNSet, ok := m["policy_arns"].(*schema.Set); ok && policyARNSet.Len() > 0 {
		for _, policyARNRaw := range policyARNSet.List() {
			policyARN, ok := policyARNRaw.(string)

			if !ok {
				continue
			}

			assumeRole.PolicyARNs = append(assumeRole.PolicyARNs, policyARN)
		}
	}

	if v, ok := m["role_arn"].(string); ok && v != "" {
		assumeRole.RoleARN = v
	}

	if v, ok := m["session_name"].(string); ok && v != "" {
		assumeRole.SessionName = v
	}

	if v, ok := m["web_identity_token"].(string); ok && v != "" {
		assumeRole.WebIdentityToken = v
	}

	if v, ok := m["web_identity_token_file"].(string); ok && v != "" {
		assumeRole.WebIdentityTokenFile = v
	}

	return &assumeRole
}

func expandProviderDefaultTags(l []interface{}) *tftags.DefaultConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	defaultConfig := &tftags.DefaultConfig{}
	m := l[0].(map[string]interface{})

	if v, ok := m["tags"].(map[string]interface{}); ok {
		defaultConfig.Tags = tftags.New(v)
	}
	return defaultConfig
}

func expandProviderIgnoreTags(l []interface{}) *tftags.IgnoreConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	ignoreConfig := &tftags.IgnoreConfig{}
	m := l[0].(map[string]interface{})

	if v, ok := m["keys"].(*schema.Set); ok {
		ignoreConfig.Keys = tftags.New(v.List())
	}

	if v, ok := m["key_prefixes"].(*schema.Set); ok {
		ignoreConfig.KeyPrefixes = tftags.New(v.List())
	}

	return ignoreConfig
}

func expandEndpoints(endpointsSetList []interface{}, out map[string]string) error {
	for _, endpointsSetI := range endpointsSetList {
		endpoints := endpointsSetI.(map[string]interface{})

		for _, hclKey := range names.Aliases() {
			var serviceKey string
			var err error
			if serviceKey, err = names.ProviderPackageForAlias(hclKey); err != nil {
				return fmt.Errorf("failed to assign endpoint (%s): %w", hclKey, err)
			}

			if out[serviceKey] == "" && endpoints[hclKey].(string) != "" {
				out[serviceKey] = endpoints[hclKey].(string)
			}
		}
	}

	for _, service := range names.ProviderPackages() {
		if out[service] != "" {
			continue
		}

		envvar := names.EnvVar(service)
		if envvar != "" {
			if v := os.Getenv(envvar); v != "" {
				out[service] = v
				continue
			}
		}
		if envvarDeprecated := names.DeprecatedEnvVar(service); envvarDeprecated != "" {
			if v := os.Getenv(envvarDeprecated); v != "" {
				log.Printf("[WARN] The environment variable %q is deprecated. Use %q instead.", envvarDeprecated, envvar)
				out[service] = v
			}
		}
	}

	return nil
}
