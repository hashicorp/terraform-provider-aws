package provider

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/mutexkv"
)

// Provider returns a *schema.Provider.
func Provider() *schema.Provider {
	// TODO: Move the validation to this, requires conditional schemas
	// TODO: Move the configuration to this, requires validation

	// The actual provider
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"access_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: descriptions["access_key"],
			},

			"secret_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: descriptions["secret_key"],
			},

			"profile": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: descriptions["profile"],
			},

			"assume_role": AssumeRoleSchema(),

			"shared_credentials_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: descriptions["shared_credentials_file"],
			},

			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: descriptions["token"],
			},

			"region": {
				Type:     schema.TypeString,
				Required: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"AWS_REGION",
					"AWS_DEFAULT_REGION",
				}, nil),
				Description:  descriptions["region"],
				InputDefault: "us-east-1", // lintignore:AWSAT003
			},

			"max_retries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     25,
				Description: descriptions["max_retries"],
			},

			"allowed_account_ids": {
				Type:          schema.TypeSet,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Optional:      true,
				ConflictsWith: []string{"forbidden_account_ids"},
				Set:           schema.HashString,
			},

			"forbidden_account_ids": {
				Type:          schema.TypeSet,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Optional:      true,
				ConflictsWith: []string{"allowed_account_ids"},
				Set:           schema.HashString,
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

			"endpoints": EndpointsSchema(),

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
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: descriptions["insecure"],
			},

			"skip_credentials_validation": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: descriptions["skip_credentials_validation"],
			},

			"skip_get_ec2_platforms": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: descriptions["skip_get_ec2_platforms"],
			},

			"skip_region_validation": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: descriptions["skip_region_validation"],
			},

			"skip_requesting_account_id": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: descriptions["skip_requesting_account_id"],
			},

			"skip_metadata_api_check": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: descriptions["skip_metadata_api_check"],
			},

			"s3_force_path_style": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: descriptions["s3_force_path_style"],
			},
		},
	}

	provider.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		terraformVersion := provider.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}
		return providerConfigure(d, terraformVersion)
	}

	for {
		if registrationComplete {
			break
		}

		time.Sleep(10 * time.Second)
	}

	provider.DataSourcesMap = dataSources
	provider.ResourcesMap = resources

	return provider
}

// This is a global MutexKV for use within this plugin.
var MutexKV = mutexkv.NewMutexKV()

var descriptions map[string]string
var EndpointServiceNames []string

func init() {
	descriptions = map[string]string{
		"region": "The region where AWS operations will take place. Examples\n" +
			"are us-east-1, us-west-2, etc.", // lintignore:AWSAT003

		"access_key": "The access key for API operations. You can retrieve this\n" +
			"from the 'Security & Credentials' section of the AWS console.",

		"secret_key": "The secret key for API operations. You can retrieve this\n" +
			"from the 'Security & Credentials' section of the AWS console.",

		"profile": "The profile for API operations. If not set, the default profile\n" +
			"created with `aws configure` will be used.",

		"shared_credentials_file": "The path to the shared credentials file. If not set\n" +
			"this defaults to ~/.aws/credentials.",

		"token": "session token. A session token is only required if you are\n" +
			"using temporary security credentials.",

		"max_retries": "The maximum number of times an AWS API request is\n" +
			"being executed. If the API request still fails, an error is\n" +
			"thrown.",

		"endpoint": "Use this to override the default service endpoint URL",

		"insecure": "Explicitly allow the provider to perform \"insecure\" SSL requests. If omitted," +
			"default value is `false`",

		"skip_credentials_validation": "Skip the credentials validation via STS API. " +
			"Used for AWS API implementations that do not have STS available/implemented.",

		"skip_get_ec2_platforms": "Skip getting the supported EC2 platforms. " +
			"Used by users that don't have ec2:DescribeAccountAttributes permissions.",

		"skip_region_validation": "Skip static validation of region name. " +
			"Used by users of alternative AWS-like APIs or users w/ access to regions that are not public (yet).",

		"skip_requesting_account_id": "Skip requesting the account ID. " +
			"Used for AWS API implementations that do not have IAM/STS API and/or metadata API.",

		"skip_medatadata_api_check": "Skip the AWS Metadata API check. " +
			"Used for AWS API implementations that do not have a metadata api endpoint.",

		"s3_force_path_style": "Set this to true to force the request to use path-style addressing,\n" +
			"i.e., http://s3.amazonaws.com/BUCKET/KEY. By default, the S3 client will\n" +
			"use virtual hosted bucket addressing when possible\n" +
			"(http://BUCKET.s3.amazonaws.com/KEY). Specific to the Amazon S3 service.",
	}

	EndpointServiceNames = []string{
		"accessanalyzer",
		"acm",
		"acmpca",
		"amplify",
		"apigateway",
		"appconfig",
		"applicationautoscaling",
		"applicationinsights",
		"appmesh",
		"apprunner",
		"appstream",
		"appsync",
		"athena",
		"auditmanager",
		"autoscaling",
		"autoscalingplans",
		"backup",
		"batch",
		"budgets",
		"chime",
		"cloud9",
		"cloudformation",
		"cloudfront",
		"cloudhsm",
		"cloudsearch",
		"cloudtrail",
		"cloudwatch",
		"cloudwatchevents",
		"cloudwatchlogs",
		"codeartifact",
		"codebuild",
		"codecommit",
		"codedeploy",
		"codepipeline",
		"codestarconnections",
		"cognitoidentity",
		"cognitoidp",
		"configservice",
		"connect",
		"cur",
		"dataexchange",
		"datapipeline",
		"datasync",
		"dax",
		"detective",
		"devicefarm",
		"directconnect",
		"dlm",
		"dms",
		"docdb",
		"ds",
		"dynamodb",
		"ec2",
		"ecr",
		"ecrpublic",
		"ecs",
		"efs",
		"eks",
		"elasticache",
		"elasticbeanstalk",
		"elastictranscoder",
		"elb",
		"emr",
		"emrcontainers",
		"es",
		"firehose",
		"fms",
		"forecast",
		"fsx",
		"gamelift",
		"glacier",
		"globalaccelerator",
		"glue",
		"greengrass",
		"guardduty",
		"iam",
		"identitystore",
		"imagebuilder",
		"inspector",
		"iot",
		"iotanalytics",
		"iotevents",
		"kafka",
		"kinesis",
		"kinesisanalytics",
		"kinesisanalyticsv2",
		"kinesisvideo",
		"kms",
		"lakeformation",
		"lambda",
		"lexmodels",
		"licensemanager",
		"lightsail",
		"macie",
		"macie2",
		"managedblockchain",
		"marketplacecatalog",
		"mediaconnect",
		"mediaconvert",
		"medialive",
		"mediapackage",
		"mediastore",
		"mediastoredata",
		"mq",
		"mwaa",
		"neptune",
		"networkfirewall",
		"networkmanager",
		"opsworks",
		"organizations",
		"outposts",
		"personalize",
		"pinpoint",
		"pricing",
		"qldb",
		"quicksight",
		"ram",
		"rds",
		"redshift",
		"resourcegroups",
		"resourcegroupstaggingapi",
		"route53",
		"route53domains",
		"route53resolver",
		"s3",
		"s3control",
		"s3outposts",
		"sagemaker",
		"schemas",
		"sdb",
		"secretsmanager",
		"securityhub",
		"serverlessrepo",
		"servicecatalog",
		"servicediscovery",
		"servicequotas",
		"ses",
		"shield",
		"signer",
		"sns",
		"sqs",
		"ssm",
		"ssoadmin",
		"stepfunctions",
		"storagegateway",
		"sts",
		"swf",
		"synthetics",
		"timestreamwrite",
		"transfer",
		"waf",
		"wafregional",
		"wafv2",
		"worklink",
		"workmail",
		"workspaces",
		"xray",
	}
}
