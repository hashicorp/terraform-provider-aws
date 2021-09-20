package conns

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/accessanalyzer"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/aws/aws-sdk-go/service/amplify"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/aws/aws-sdk-go/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go/service/applicationinsights"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/aws/aws-sdk-go/service/auditmanager"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscalingplans"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/aws/aws-sdk-go/service/chime"
	"github.com/aws/aws-sdk-go/service/cloud9"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/cloudhsmv2"
	"github.com/aws/aws-sdk-go/service/cloudsearch"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/codeartifact"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/aws/aws-sdk-go/service/codedeploy"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/aws/aws-sdk-go/service/codestarconnections"
	"github.com/aws/aws-sdk-go/service/codestarnotifications"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/aws/aws-sdk-go/service/costandusagereportservice"
	"github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/aws/aws-sdk-go/service/dataexchange"
	"github.com/aws/aws-sdk-go/service/datapipeline"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/aws/aws-sdk-go/service/dax"
	"github.com/aws/aws-sdk-go/service/detective"
	"github.com/aws/aws-sdk-go/service/devicefarm"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/aws/aws-sdk-go/service/dlm"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecrpublic"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/aws/aws-sdk-go/service/elastictranscoder"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/aws/aws-sdk-go/service/emrcontainers"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-sdk-go/service/fms"
	"github.com/aws/aws-sdk-go/service/forecastservice"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/aws/aws-sdk-go/service/glacier"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/aws/aws-sdk-go/service/greengrass"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/identitystore"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/aws/aws-sdk-go/service/inspector"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/aws/aws-sdk-go/service/iotanalytics"
	"github.com/aws/aws-sdk-go/service/iotevents"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
	"github.com/aws/aws-sdk-go/service/kinesisanalyticsv2"
	"github.com/aws/aws-sdk-go/service/kinesisvideo"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/aws/aws-sdk-go/service/licensemanager"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/aws/aws-sdk-go/service/locationservice"
	"github.com/aws/aws-sdk-go/service/macie"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/aws/aws-sdk-go/service/managedblockchain"
	"github.com/aws/aws-sdk-go/service/marketplacecatalog"
	"github.com/aws/aws-sdk-go/service/mediaconnect"
	"github.com/aws/aws-sdk-go/service/mediaconvert"
	"github.com/aws/aws-sdk-go/service/medialive"
	"github.com/aws/aws-sdk-go/service/mediapackage"
	"github.com/aws/aws-sdk-go/service/mediastore"
	"github.com/aws/aws-sdk-go/service/mediastoredata"
	"github.com/aws/aws-sdk-go/service/memorydb"
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/aws/aws-sdk-go/service/mwaa"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/outposts"
	"github.com/aws/aws-sdk-go/service/personalize"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/aws/aws-sdk-go/service/qldb"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/aws/aws-sdk-go/service/resourcegroups"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53domains"
	"github.com/aws/aws-sdk-go/service/route53recoverycontrolconfig"
	"github.com/aws/aws-sdk-go/service/route53recoveryreadiness"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/aws/aws-sdk-go/service/s3outposts"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/aws/aws-sdk-go/service/schemas"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/aws/aws-sdk-go/service/serverlessapplicationrepository"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/aws/aws-sdk-go/service/servicequotas"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/aws/aws-sdk-go/service/shield"
	"github.com/aws/aws-sdk-go/service/signer"
	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/swf"
	"github.com/aws/aws-sdk-go/service/synthetics"
	"github.com/aws/aws-sdk-go/service/timestreamwrite"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/aws/aws-sdk-go/service/worklink"
	"github.com/aws/aws-sdk-go/service/workmail"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/aws/aws-sdk-go/service/xray"
	awsbase "github.com/hashicorp/aws-sdk-go-base"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	tftags "github.com/hashicorp/terraform-provider-aws/aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/version"
)

type Config struct {
	AccessKey     string
	SecretKey     string
	CredsFilename string
	Profile       string
	Token         string
	Region        string
	MaxRetries    int

	AssumeRoleARN               string
	AssumeRoleDurationSeconds   int
	AssumeRoleExternalID        string
	AssumeRolePolicy            string
	AssumeRolePolicyARNs        []string
	AssumeRoleSessionName       string
	AssumeRoleTags              map[string]string
	AssumeRoleTransitiveTagKeys []string

	AllowedAccountIds   []string
	ForbiddenAccountIds []string

	DefaultTagsConfig *tftags.DefaultConfig
	Endpoints         map[string]string
	IgnoreTagsConfig  *tftags.IgnoreConfig
	Insecure          bool

	SkipCredsValidation     bool
	SkipGetEC2Platforms     bool
	SkipRegionValidation    bool
	SkipRequestingAccountId bool
	SkipMetadataApiCheck    bool
	S3ForcePathStyle        bool

	TerraformVersion string
}

type AWSClient struct {
	AccessAnalyzerConn               *accessanalyzer.AccessAnalyzer
	AccountID                        string
	ACMConn                          *acm.ACM
	ACMPCAConn                       *acmpca.ACMPCA
	AmplifyConn                      *amplify.Amplify
	APIGatewayConn                   *apigateway.APIGateway
	APIGatewayV2Conn                 *apigatewayv2.ApiGatewayV2
	ApplicationAutoScalingConn       *applicationautoscaling.ApplicationAutoScaling
	AppConfigConn                    *appconfig.AppConfig
	ApplicationInsightsConn          *applicationinsights.ApplicationInsights
	AppMeshConn                      *appmesh.AppMesh
	AppRunnerConn                    *apprunner.AppRunner
	AppStreamConn                    *appstream.AppStream
	AppSyncConn                      *appsync.AppSync
	AthenaConn                       *athena.Athena
	AuditManagerConn                 *auditmanager.AuditManager
	AutoScalingConn                  *autoscaling.AutoScaling
	AutoScalingPlansConn             *autoscalingplans.AutoScalingPlans
	BackupConn                       *backup.Backup
	BatchConn                        *batch.Batch
	BudgetsConn                      *budgets.Budgets
	CloudFormationConn               *cloudformation.CloudFormation
	ChimeConn                        *chime.Chime
	Cloud9Conn                       *cloud9.Cloud9
	CloudFrontConn                   *cloudfront.CloudFront
	CloudHSMV2Conn                   *cloudhsmv2.CloudHSMV2
	CloudSearchConn                  *cloudsearch.CloudSearch
	CloudTrailConn                   *cloudtrail.CloudTrail
	CloudWatchConn                   *cloudwatch.CloudWatch
	CloudWatchEventsConn             *cloudwatchevents.CloudWatchEvents
	CloudWatchLogsConn               *cloudwatchlogs.CloudWatchLogs
	CodeArtifactConn                 *codeartifact.CodeArtifact
	CodeBuildConn                    *codebuild.CodeBuild
	CodeCommitConn                   *codecommit.CodeCommit
	CodeDeployConn                   *codedeploy.CodeDeploy
	CodePipelineConn                 *codepipeline.CodePipeline
	CodeStarConnectionsConn          *codestarconnections.CodeStarConnections
	CodeStarNotificationsConn        *codestarnotifications.CodeStarNotifications
	CognitoIdentityConn              *cognitoidentity.CognitoIdentity
	CognitoIDPConn                   *cognitoidentityprovider.CognitoIdentityProvider
	ConfigConn                       *configservice.ConfigService
	ConnectConn                      *connect.Connect
	CURConn                          *costandusagereportservice.CostandUsageReportService
	DataExchangeConn                 *dataexchange.DataExchange
	DataPipelineConn                 *datapipeline.DataPipeline
	DataSyncConn                     *datasync.DataSync
	DAXConn                          *dax.DAX
	DefaultTagsConfig                *tftags.DefaultConfig
	DetectiveConn                    *detective.Detective
	DeviceFarmConn                   *devicefarm.DeviceFarm
	DLMConn                          *dlm.DLM
	DMSConn                          *databasemigrationservice.DatabaseMigrationService
	DNSSuffix                        string
	DocDBConn                        *docdb.DocDB
	DirectoryServiceConn             *directoryservice.DirectoryService
	DirectConnectConn                *directconnect.DirectConnect
	DynamoDBConn                     *dynamodb.DynamoDB
	EC2Conn                          *ec2.EC2
	ECRConn                          *ecr.ECR
	ECRPublicConn                    *ecrpublic.ECRPublic
	ECSConn                          *ecs.ECS
	EFSConn                          *efs.EFS
	EKSConn                          *eks.EKS
	ElastiCacheConn                  *elasticache.ElastiCache
	ElasticBeanstalkConn             *elasticbeanstalk.ElasticBeanstalk
	ElasticTranscoderConn            *elastictranscoder.ElasticTranscoder
	ELBConn                          *elb.ELB
	ELBV2Conn                        *elbv2.ELBV2
	EMRConn                          *emr.EMR
	EMRContainersConn                *emrcontainers.EMRContainers
	ElasticSearchConn                *elasticsearch.ElasticsearchService
	FirehoseConn                     *firehose.Firehose
	FMSConn                          *fms.FMS
	ForecastConn                     *forecastservice.ForecastService
	FSxConn                          *fsx.FSx
	GameLiftConn                     *gamelift.GameLift
	GlacierConn                      *glacier.Glacier
	GlobalAcceleratorConn            *globalaccelerator.GlobalAccelerator
	GlueConn                         *glue.Glue
	GuardDutyConn                    *guardduty.GuardDuty
	GreengrassConn                   *greengrass.Greengrass
	IAMConn                          *iam.IAM
	IdentityStoreConn                *identitystore.IdentityStore
	IgnoreTagsConfig                 *tftags.IgnoreConfig
	ImageBuilderConn                 *imagebuilder.Imagebuilder
	InspectorConn                    *inspector.Inspector
	IoTConn                          *iot.IoT
	IoTAnalyticsConn                 *iotanalytics.IoTAnalytics
	IoTEventsConn                    *iotevents.IoTEvents
	KafkaConn                        *kafka.Kafka
	KinesisAnalyticsConn             *kinesisanalytics.KinesisAnalytics
	KinesisAnalyticsV2Conn           *kinesisanalyticsv2.KinesisAnalyticsV2
	KinesisConn                      *kinesis.Kinesis
	KinesisVideoConn                 *kinesisvideo.KinesisVideo
	KMSConn                          *kms.KMS
	LakeFormationConn                *lakeformation.LakeFormation
	LambdaConn                       *lambda.Lambda
	LexModelBuildingConn             *lexmodelbuildingservice.LexModelBuildingService
	LicenseManagerConn               *licensemanager.LicenseManager
	LightsailConn                    *lightsail.Lightsail
	LocationConn                     *locationservice.LocationService
	MacieConn                        *macie.Macie
	Macie2Conn                       *macie2.Macie2
	ManagedBlockchainConn            *managedblockchain.ManagedBlockchain
	MarketplaceCatalogConn           *marketplacecatalog.MarketplaceCatalog
	MediaConnectConn                 *mediaconnect.MediaConnect
	MediaConvertConn                 *mediaconvert.MediaConvert
	MediaConvertAccountConn          *mediaconvert.MediaConvert
	MediaLiveConn                    *medialive.MediaLive
	MediaPackageConn                 *mediapackage.MediaPackage
	MediaStoreConn                   *mediastore.MediaStore
	MediaStoreDataConn               *mediastoredata.MediaStoreData
	MemoryDBConn                     *memorydb.MemoryDB
	MQConn                           *mq.MQ
	MWAAConn                         *mwaa.MWAA
	NeptuneConn                      *neptune.Neptune
	NetworkFirewallConn              *networkfirewall.NetworkFirewall
	NetworkManagerConn               *networkmanager.NetworkManager
	OpsWorksConn                     *opsworks.OpsWorks
	OrganizationsConn                *organizations.Organizations
	OutpostsConn                     *outposts.Outposts
	Partition                        string
	PersonalizeConn                  *personalize.Personalize
	PrometheusConn                   *prometheusservice.PrometheusService
	PinpointConn                     *pinpoint.Pinpoint
	PricingConn                      *pricing.Pricing
	QLDBConn                         *qldb.QLDB
	QuickSightConn                   *quicksight.QuickSight
	Route53Conn                      *route53.Route53
	RAMConn                          *ram.RAM
	RDSConn                          *rds.RDS
	RedshiftConn                     *redshift.Redshift
	Region                           string
	ResourceGroupsConn               *resourcegroups.ResourceGroups
	ResourceGroupsTaggingConn        *resourcegroupstaggingapi.ResourceGroupsTaggingAPI
	ReverseDNSPrefix                 string
	Route53DomainsConn               *route53domains.Route53Domains
	Route53RecoveryControlConfigConn *route53recoverycontrolconfig.Route53RecoveryControlConfig
	Route53RecoveryReadinessConn     *route53recoveryreadiness.Route53RecoveryReadiness
	Route53ResolverConn              *route53resolver.Route53Resolver
	S3Conn                           *s3.S3
	S3ConnURICleaningDisabled        *s3.S3
	S3ControlConn                    *s3control.S3Control
	S3OutpostsConn                   *s3outposts.S3Outposts
	SageMakerConn                    *sagemaker.SageMaker
	ServiceCatalogConn               *servicecatalog.ServiceCatalog
	SchemasConn                      *schemas.Schemas
	ServiceDiscoveryConn             *servicediscovery.ServiceDiscovery
	SecretsManagerConn               *secretsmanager.SecretsManager
	SecurityHubConn                  *securityhub.SecurityHub
	ServerlessAppRepoConn            *serverlessapplicationrepository.ServerlessApplicationRepository
	ServiceQuotasConn                *servicequotas.ServiceQuotas
	SESConn                          *ses.SES
	SFNConn                          *sfn.SFN
	ShieldConn                       *shield.Shield
	SignerConn                       *signer.Signer
	SimpleDBConn                     *simpledb.SimpleDB
	SNSConn                          *sns.SNS
	SQSConn                          *sqs.SQS
	SSMConn                          *ssm.SSM
	SSOAdminConn                     *ssoadmin.SSOAdmin
	StorageGatewayConn               *storagegateway.StorageGateway
	STSConn                          *sts.STS
	SupportedPlatforms               []string
	SWFConn                          *swf.SWF
	SyntheticsConn                   *synthetics.Synthetics
	TerraformVersion                 string
	TimestreamWriteConn              *timestreamwrite.TimestreamWrite
	TransferConn                     *transfer.Transfer
	WAFConn                          *waf.WAF
	WAFRegionalConn                  *wafregional.WAFRegional
	WAFV2Conn                        *wafv2.WAFV2
	WorkLinkConn                     *worklink.WorkLink
	WorkMailConn                     *workmail.WorkMail
	WorkSpacesConn                   *workspaces.WorkSpaces
	XRayConn                         *xray.XRay
}

// PartitionHostname returns a hostname with the provider domain suffix for the partition
// e.g. PREFIX.amazonaws.com
// The prefix should not contain a trailing period.
func (client *AWSClient) PartitionHostname(prefix string) string {
	return fmt.Sprintf("%s.%s", prefix, client.DNSSuffix)
}

// RegionalHostname returns a hostname with the provider domain suffix for the region and partition
// e.g. PREFIX.us-west-2.amazonaws.com
// The prefix should not contain a trailing period.
func (client *AWSClient) RegionalHostname(prefix string) string {
	return fmt.Sprintf("%s.%s.%s", prefix, client.Region, client.DNSSuffix)
}

// Client configures and returns a fully initialized AWSClient
func (c *Config) Client() (interface{}, error) {
	// Get the auth and region. This can fail if keys/regions were not
	// specified and we're attempting to use the environment.
	if !c.SkipRegionValidation {
		if err := awsbase.ValidateRegion(c.Region); err != nil {
			return nil, err
		}
	}

	awsbaseConfig := &awsbase.Config{
		AccessKey:                   c.AccessKey,
		AssumeRoleARN:               c.AssumeRoleARN,
		AssumeRoleDurationSeconds:   c.AssumeRoleDurationSeconds,
		AssumeRoleExternalID:        c.AssumeRoleExternalID,
		AssumeRolePolicy:            c.AssumeRolePolicy,
		AssumeRolePolicyARNs:        c.AssumeRolePolicyARNs,
		AssumeRoleSessionName:       c.AssumeRoleSessionName,
		AssumeRoleTags:              c.AssumeRoleTags,
		AssumeRoleTransitiveTagKeys: c.AssumeRoleTransitiveTagKeys,
		CallerDocumentationURL:      "https://registry.terraform.io/providers/hashicorp/aws",
		CallerName:                  "Terraform AWS Provider",
		CredsFilename:               c.CredsFilename,
		DebugLogging:                logging.IsDebugOrHigher(),
		IamEndpoint:                 c.Endpoints["iam"],
		Insecure:                    c.Insecure,
		MaxRetries:                  c.MaxRetries,
		Profile:                     c.Profile,
		Region:                      c.Region,
		SecretKey:                   c.SecretKey,
		SkipCredsValidation:         c.SkipCredsValidation,
		SkipMetadataApiCheck:        c.SkipMetadataApiCheck,
		SkipRequestingAccountId:     c.SkipRequestingAccountId,
		StsEndpoint:                 c.Endpoints["sts"],
		Token:                       c.Token,
		UserAgentProducts: []*awsbase.UserAgentProduct{
			{Name: "APN", Version: "1.0"},
			{Name: "HashiCorp", Version: "1.0"},
			{Name: "Terraform", Version: c.TerraformVersion, Extra: []string{"+https://www.terraform.io"}},
			{Name: "terraform-provider-aws", Version: version.ProviderVersion, Extra: []string{"+https://registry.terraform.io/providers/hashicorp/aws"}},
		},
	}

	sess, accountID, Partition, err := awsbase.GetSessionWithAccountIDAndPartition(awsbaseConfig)
	if err != nil {
		return nil, fmt.Errorf("error configuring Terraform AWS Provider: %w", err)
	}

	if accountID == "" {
		log.Printf("[WARN] AWS account ID not found for provider. See https://www.terraform.io/docs/providers/aws/index.html#skip_requesting_account_id for implications.")
	}

	if err := awsbase.ValidateAccountID(accountID, c.AllowedAccountIds, c.ForbiddenAccountIds); err != nil {
		return nil, err
	}

	DNSSuffix := "amazonaws.com"
	if p, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), c.Region); ok {
		DNSSuffix = p.DNSSuffix()
	}

	client := &AWSClient{
		AccessAnalyzerConn:               accessanalyzer.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["accessanalyzer"])})),
		AccountID:                        accountID,
		ACMConn:                          acm.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["acm"])})),
		ACMPCAConn:                       acmpca.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["acmpca"])})),
		AmplifyConn:                      amplify.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["amplify"])})),
		APIGatewayConn:                   apigateway.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["apigateway"])})),
		APIGatewayV2Conn:                 apigatewayv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["apigateway"])})),
		ApplicationAutoScalingConn:       applicationautoscaling.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["applicationautoscaling"])})),
		AppConfigConn:                    appconfig.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["appconfig"])})),
		ApplicationInsightsConn:          applicationinsights.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["applicationinsights"])})),
		AppMeshConn:                      appmesh.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["appmesh"])})),
		AppRunnerConn:                    apprunner.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["apprunner"])})),
		AppStreamConn:                    appstream.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["appstream"])})),
		AppSyncConn:                      appsync.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["appsync"])})),
		AthenaConn:                       athena.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["athena"])})),
		AuditManagerConn:                 auditmanager.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["auditmanager"])})),
		AutoScalingConn:                  autoscaling.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["autoscaling"])})),
		AutoScalingPlansConn:             autoscalingplans.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["autoscalingplans"])})),
		BackupConn:                       backup.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["backup"])})),
		BatchConn:                        batch.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["batch"])})),
		BudgetsConn:                      budgets.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["budgets"])})),
		CloudFormationConn:               cloudformation.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cloudformation"])})),
		ChimeConn:                        chime.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["chime"])})),
		Cloud9Conn:                       cloud9.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cloud9"])})),
		CloudFrontConn:                   cloudfront.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cloudfront"])})),
		CloudHSMV2Conn:                   cloudhsmv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cloudhsm"])})),
		CloudSearchConn:                  cloudsearch.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cloudsearch"])})),
		CloudTrailConn:                   cloudtrail.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cloudtrail"])})),
		CloudWatchConn:                   cloudwatch.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cloudwatch"])})),
		CloudWatchEventsConn:             cloudwatchevents.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cloudwatchevents"])})),
		CloudWatchLogsConn:               cloudwatchlogs.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cloudwatchlogs"])})),
		CodeArtifactConn:                 codeartifact.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["codeartifact"])})),
		CodeBuildConn:                    codebuild.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["codebuild"])})),
		CodeCommitConn:                   codecommit.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["codecommit"])})),
		CodeDeployConn:                   codedeploy.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["codedeploy"])})),
		CodePipelineConn:                 codepipeline.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["codepipeline"])})),
		CodeStarConnectionsConn:          codestarconnections.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["codestarconnections"])})),
		CodeStarNotificationsConn:        codestarnotifications.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["codestarnotifications"])})),
		CognitoIdentityConn:              cognitoidentity.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cognitoidentity"])})),
		CognitoIDPConn:                   cognitoidentityprovider.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cognitoidp"])})),
		ConfigConn:                       configservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["configservice"])})),
		ConnectConn:                      connect.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["connect"])})),
		CURConn:                          costandusagereportservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cur"])})),
		DataExchangeConn:                 dataexchange.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["dataexchange"])})),
		DataPipelineConn:                 datapipeline.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["datapipeline"])})),
		DataSyncConn:                     datasync.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["datasync"])})),
		DAXConn:                          dax.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["dax"])})),
		DefaultTagsConfig:                c.DefaultTagsConfig,
		DetectiveConn:                    detective.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["detective"])})),
		DeviceFarmConn:                   devicefarm.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["devicefarm"])})),
		DLMConn:                          dlm.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["dlm"])})),
		DMSConn:                          databasemigrationservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["dms"])})),
		DNSSuffix:                        DNSSuffix,
		DocDBConn:                        docdb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["docdb"])})),
		DirectoryServiceConn:             directoryservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["ds"])})),
		DirectConnectConn:                directconnect.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["directconnect"])})),
		DynamoDBConn:                     dynamodb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["dynamodb"])})),
		EC2Conn:                          ec2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["ec2"])})),
		ECRConn:                          ecr.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["ecr"])})),
		ECRPublicConn:                    ecrpublic.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["ecrpublic"])})),
		ECSConn:                          ecs.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["ecs"])})),
		EFSConn:                          efs.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["efs"])})),
		EKSConn:                          eks.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["eks"])})),
		ElastiCacheConn:                  elasticache.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["elasticache"])})),
		ElasticBeanstalkConn:             elasticbeanstalk.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["elasticbeanstalk"])})),
		ElasticTranscoderConn:            elastictranscoder.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["elastictranscoder"])})),
		ELBConn:                          elb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["elb"])})),
		ELBV2Conn:                        elbv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["elb"])})),
		EMRConn:                          emr.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["emr"])})),
		EMRContainersConn:                emrcontainers.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["emrcontainers"])})),
		ElasticSearchConn:                elasticsearch.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["es"])})),
		FirehoseConn:                     firehose.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["firehose"])})),
		FMSConn:                          fms.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["fms"])})),
		ForecastConn:                     forecastservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["forecast"])})),
		FSxConn:                          fsx.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["fsx"])})),
		GameLiftConn:                     gamelift.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["gamelift"])})),
		GlacierConn:                      glacier.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["glacier"])})),
		GlueConn:                         glue.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["glue"])})),
		GuardDutyConn:                    guardduty.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["guardduty"])})),
		GreengrassConn:                   greengrass.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["greengrass"])})),
		IAMConn:                          iam.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["iam"])})),
		IdentityStoreConn:                identitystore.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["identitystore"])})),
		IgnoreTagsConfig:                 c.IgnoreTagsConfig,
		ImageBuilderConn:                 imagebuilder.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["imagebuilder"])})),
		InspectorConn:                    inspector.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["inspector"])})),
		IoTConn:                          iot.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["iot"])})),
		IoTAnalyticsConn:                 iotanalytics.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["iotanalytics"])})),
		IoTEventsConn:                    iotevents.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["iotevents"])})),
		KafkaConn:                        kafka.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["kafka"])})),
		KinesisAnalyticsConn:             kinesisanalytics.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["kinesisanalytics"])})),
		KinesisAnalyticsV2Conn:           kinesisanalyticsv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["kinesisanalyticsv2"])})),
		KinesisConn:                      kinesis.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["kinesis"])})),
		KinesisVideoConn:                 kinesisvideo.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["kinesisvideo"])})),
		KMSConn:                          kms.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["kms"])})),
		LakeFormationConn:                lakeformation.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["lakeformation"])})),
		LambdaConn:                       lambda.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["lambda"])})),
		LexModelBuildingConn:             lexmodelbuildingservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["lexmodels"])})),
		LicenseManagerConn:               licensemanager.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["licensemanager"])})),
		LightsailConn:                    lightsail.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["lightsail"])})),
		LocationConn:                     locationservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["location"])})),
		MacieConn:                        macie.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["macie"])})),
		Macie2Conn:                       macie2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["macie2"])})),
		ManagedBlockchainConn:            managedblockchain.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["managedblockchain"])})),
		MarketplaceCatalogConn:           marketplacecatalog.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["marketplacecatalog"])})),
		MediaConnectConn:                 mediaconnect.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["mediaconnect"])})),
		MediaConvertConn:                 mediaconvert.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["mediaconvert"])})),
		MediaLiveConn:                    medialive.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["medialive"])})),
		MediaPackageConn:                 mediapackage.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["mediapackage"])})),
		MediaStoreConn:                   mediastore.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["mediastore"])})),
		MediaStoreDataConn:               mediastoredata.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["mediastoredata"])})),
		MemoryDBConn:                     memorydb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["memorydb"])})),
		MQConn:                           mq.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["mq"])})),
		MWAAConn:                         mwaa.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["mwaa"])})),
		NeptuneConn:                      neptune.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["neptune"])})),
		NetworkFirewallConn:              networkfirewall.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["networkfirewall"])})),
		NetworkManagerConn:               networkmanager.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["networkmanager"])})),
		OpsWorksConn:                     opsworks.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["opsworks"])})),
		OrganizationsConn:                organizations.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["organizations"])})),
		OutpostsConn:                     outposts.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["outposts"])})),
		Partition:                        Partition,
		PersonalizeConn:                  personalize.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["personalize"])})),
		PrometheusConn:                   prometheusservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["prometheusservice"])})),
		PinpointConn:                     pinpoint.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["pinpoint"])})),
		PricingConn:                      pricing.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["pricing"])})),
		QLDBConn:                         qldb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["qldb"])})),
		QuickSightConn:                   quicksight.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["quicksight"])})),
		RAMConn:                          ram.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["ram"])})),
		RDSConn:                          rds.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["rds"])})),
		RedshiftConn:                     redshift.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["redshift"])})),
		Region:                           c.Region,
		ResourceGroupsConn:               resourcegroups.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["resourcegroups"])})),
		ResourceGroupsTaggingConn:        resourcegroupstaggingapi.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["resourcegroupstaggingapi"])})),
		ReverseDNSPrefix:                 ReverseDNS(DNSSuffix),
		Route53DomainsConn:               route53domains.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["route53domains"])})),
		Route53RecoveryControlConfigConn: route53recoverycontrolconfig.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["route53recoverycontrolconfig"])})),
		Route53RecoveryReadinessConn:     route53recoveryreadiness.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["route53recoveryreadiness"])})),
		Route53ResolverConn:              route53resolver.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["route53resolver"])})),
		S3ControlConn:                    s3control.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["s3control"])})),
		S3OutpostsConn:                   s3outposts.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["s3outposts"])})),
		SageMakerConn:                    sagemaker.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["sagemaker"])})),
		ServiceCatalogConn:               servicecatalog.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["servicecatalog"])})),
		SchemasConn:                      schemas.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["schemas"])})),
		ServiceDiscoveryConn:             servicediscovery.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["servicediscovery"])})),
		SecretsManagerConn:               secretsmanager.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["secretsmanager"])})),
		SecurityHubConn:                  securityhub.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["securityhub"])})),
		ServerlessAppRepoConn:            serverlessapplicationrepository.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["serverlessrepo"])})),
		ServiceQuotasConn:                servicequotas.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["servicequotas"])})),
		SESConn:                          ses.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["ses"])})),
		SFNConn:                          sfn.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["stepfunctions"])})),
		SignerConn:                       signer.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["signer"])})),
		SimpleDBConn:                     simpledb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["sdb"])})),
		SNSConn:                          sns.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["sns"])})),
		SQSConn:                          sqs.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["sqs"])})),
		SSMConn:                          ssm.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["ssm"])})),
		SSOAdminConn:                     ssoadmin.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["ssoadmin"])})),
		StorageGatewayConn:               storagegateway.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["storagegateway"])})),
		STSConn:                          sts.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["sts"])})),
		SWFConn:                          swf.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["swf"])})),
		SyntheticsConn:                   synthetics.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["synthetics"])})),
		TerraformVersion:                 c.TerraformVersion,
		TimestreamWriteConn:              timestreamwrite.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["timestreamwrite"])})),
		TransferConn:                     transfer.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["transfer"])})),
		WAFConn:                          waf.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["waf"])})),
		WAFRegionalConn:                  wafregional.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["wafregional"])})),
		WAFV2Conn:                        wafv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["wafv2"])})),
		WorkLinkConn:                     worklink.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["worklink"])})),
		WorkMailConn:                     workmail.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["workmail"])})),
		WorkSpacesConn:                   workspaces.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["workspaces"])})),
		XRayConn:                         xray.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["xray"])})),
	}

	// "Global" services that require customizations
	globalAcceleratorConfig := &aws.Config{
		Endpoint: aws.String(c.Endpoints["globalaccelerator"]),
	}
	route53Config := &aws.Config{
		Endpoint: aws.String(c.Endpoints["route53"]),
	}
	route53RecoveryControlConfigConfig := &aws.Config{
		Endpoint: aws.String(c.Endpoints["route53recoverycontrolconfig"]),
	}
	route53RecoveryReadinessConfig := &aws.Config{
		Endpoint: aws.String(c.Endpoints["route53recoveryreadiness"]),
	}
	shieldConfig := &aws.Config{
		Endpoint: aws.String(c.Endpoints["shield"]),
	}

	// Services that require multiple client configurations
	s3Config := &aws.Config{
		Endpoint:         aws.String(c.Endpoints["s3"]),
		S3ForcePathStyle: aws.Bool(c.S3ForcePathStyle),
	}

	client.S3Conn = s3.New(sess.Copy(s3Config))

	s3Config.DisableRestProtocolURICleaning = aws.Bool(true)
	client.S3ConnURICleaningDisabled = s3.New(sess.Copy(s3Config))

	// Force "global" services to correct regions
	switch Partition {
	case endpoints.AwsPartitionID:
		globalAcceleratorConfig.Region = aws.String(endpoints.UsWest2RegionID)
		route53Config.Region = aws.String(endpoints.UsEast1RegionID)
		route53RecoveryControlConfigConfig.Region = aws.String(endpoints.UsWest2RegionID)
		route53RecoveryReadinessConfig.Region = aws.String(endpoints.UsWest2RegionID)
		shieldConfig.Region = aws.String(endpoints.UsEast1RegionID)
	case endpoints.AwsCnPartitionID:
		// The AWS Go SDK is missing endpoint information for Route 53 in the AWS China partition.
		// This can likely be removed in the future.
		if aws.StringValue(route53Config.Endpoint) == "" {
			route53Config.Endpoint = aws.String("https://api.route53.cn")
		}
		route53Config.Region = aws.String(endpoints.CnNorthwest1RegionID)
	case endpoints.AwsUsGovPartitionID:
		route53Config.Region = aws.String(endpoints.UsGovWest1RegionID)
	}

	client.GlobalAcceleratorConn = globalaccelerator.New(sess.Copy(globalAcceleratorConfig))
	client.Route53Conn = route53.New(sess.Copy(route53Config))
	client.Route53RecoveryControlConfigConn = route53recoverycontrolconfig.New(sess.Copy(route53RecoveryControlConfigConfig))
	client.Route53RecoveryReadinessConn = route53recoveryreadiness.New(sess.Copy(route53RecoveryReadinessConfig))
	client.ShieldConn = shield.New(sess.Copy(shieldConfig))

	client.APIGatewayConn.Handlers.Retry.PushBack(func(r *request.Request) {
		// Many operations can return an error such as:
		//   ConflictException: Unable to complete operation due to concurrent modification. Please try again later.
		// Handle them all globally for the service client.
		if tfawserr.ErrMessageContains(r.Error, apigateway.ErrCodeConflictException, "try again later") {
			r.Retryable = aws.Bool(true)
		}
	})

	// Workaround for https://github.com/aws/aws-sdk-go/issues/1472
	client.ApplicationAutoScalingConn.Handlers.Retry.PushBack(func(r *request.Request) {
		if !strings.HasPrefix(r.Operation.Name, "Describe") && !strings.HasPrefix(r.Operation.Name, "List") {
			return
		}
		if tfawserr.ErrCodeEquals(r.Error, applicationautoscaling.ErrCodeFailedResourceAccessException) {
			r.Retryable = aws.Bool(true)
		}
	})

	// StartDeployment operations can return a ConflictException
	// if ongoing deployments are in-progress, thus we handle them
	// here for the service client.
	client.AppConfigConn.Handlers.Retry.PushBack(func(r *request.Request) {
		if r.Operation.Name == "StartDeployment" {
			if tfawserr.ErrCodeEquals(r.Error, appconfig.ErrCodeConflictException) {
				r.Retryable = aws.Bool(true)
			}
		}
	})

	client.AppSyncConn.Handlers.Retry.PushBack(func(r *request.Request) {
		if r.Operation.Name == "CreateGraphqlApi" {
			if tfawserr.ErrMessageContains(r.Error, appsync.ErrCodeConcurrentModificationException, "a GraphQL API creation is already in progress") {
				r.Retryable = aws.Bool(true)
			}
		}
	})

	client.ChimeConn.Handlers.Retry.PushBack(func(r *request.Request) {
		// When calling CreateVoiceConnector across multiple resources,
		// the API can randomly return a BadRequestException without explanation
		if r.Operation.Name == "CreateVoiceConnector" {
			if tfawserr.ErrMessageContains(r.Error, chime.ErrCodeBadRequestException, "Service received a bad request") {
				r.Retryable = aws.Bool(true)
			}
		}
	})

	client.CloudHSMV2Conn.Handlers.Retry.PushBack(func(r *request.Request) {
		if tfawserr.ErrMessageContains(r.Error, cloudhsmv2.ErrCodeCloudHsmInternalFailureException, "request was rejected because of an AWS CloudHSM internal failure") {
			r.Retryable = aws.Bool(true)
		}
	})

	client.ConfigConn.Handlers.Retry.PushBack(func(r *request.Request) {
		// When calling Config Organization Rules API actions immediately
		// after Organization creation, the API can randomly return the
		// OrganizationAccessDeniedException error for a few minutes, even
		// after succeeding a few requests.
		switch r.Operation.Name {
		case "DeleteOrganizationConfigRule", "DescribeOrganizationConfigRules", "DescribeOrganizationConfigRuleStatuses", "PutOrganizationConfigRule":
			if !tfawserr.ErrMessageContains(r.Error, configservice.ErrCodeOrganizationAccessDeniedException, "This action can be only made by AWS Organization's master account.") {
				return
			}

			// We only want to retry briefly as the default max retry count would
			// excessively retry when the error could be legitimate.
			// We currently depend on the DefaultRetryer exponential backoff here.
			// ~10 retries gives a fair backoff of a few seconds.
			if r.RetryCount < 9 {
				r.Retryable = aws.Bool(true)
			} else {
				r.Retryable = aws.Bool(false)
			}
		case "DeleteOrganizationConformancePack", "DescribeOrganizationConformancePacks", "DescribeOrganizationConformancePackStatuses", "PutOrganizationConformancePack":
			if !tfawserr.ErrCodeEquals(r.Error, configservice.ErrCodeOrganizationAccessDeniedException) {
				if r.Operation.Name == "DeleteOrganizationConformancePack" && tfawserr.ErrCodeEquals(err, configservice.ErrCodeResourceInUseException) {
					r.Retryable = aws.Bool(true)
				}
				return
			}

			// We only want to retry briefly as the default max retry count would
			// excessively retry when the error could be legitimate.
			// We currently depend on the DefaultRetryer exponential backoff here.
			// ~10 retries gives a fair backoff of a few seconds.
			if r.RetryCount < 9 {
				r.Retryable = aws.Bool(true)
			} else {
				r.Retryable = aws.Bool(false)
			}
		}
	})

	client.CloudFormationConn.Handlers.Retry.PushBack(func(r *request.Request) {
		if tfawserr.ErrMessageContains(r.Error, cloudformation.ErrCodeOperationInProgressException, "Another Operation on StackSet") {
			r.Retryable = aws.Bool(true)
		}
	})

	// See https://github.com/aws/aws-sdk-go/pull/1276
	client.DynamoDBConn.Handlers.Retry.PushBack(func(r *request.Request) {
		if r.Operation.Name != "PutItem" && r.Operation.Name != "UpdateItem" && r.Operation.Name != "DeleteItem" {
			return
		}
		if tfawserr.ErrMessageContains(r.Error, dynamodb.ErrCodeLimitExceededException, "Subscriber limit exceeded:") {
			r.Retryable = aws.Bool(true)
		}
	})

	client.EC2Conn.Handlers.Retry.PushBack(func(r *request.Request) {
		if r.Operation.Name == "CreateClientVpnEndpoint" {
			if tfawserr.ErrMessageContains(r.Error, "OperationNotPermitted", "Endpoint cannot be created while another endpoint is being created") {
				r.Retryable = aws.Bool(true)
			}
		}

		if r.Operation.Name == "CreateVpnConnection" {
			if tfawserr.ErrMessageContains(r.Error, "VpnConnectionLimitExceeded", "maximum number of mutating objects has been reached") {
				r.Retryable = aws.Bool(true)
			}
		}

		if r.Operation.Name == "CreateVpnGateway" {
			if tfawserr.ErrMessageContains(r.Error, "VpnGatewayLimitExceeded", "maximum number of mutating objects has been reached") {
				r.Retryable = aws.Bool(true)
			}
		}

		if r.Operation.Name == "AttachVpnGateway" || r.Operation.Name == "DetachVpnGateway" {
			if tfawserr.ErrMessageContains(r.Error, "InvalidParameterValue", "This call cannot be completed because there are pending VPNs or Virtual Interfaces") {
				r.Retryable = aws.Bool(true)
			}
		}
	})

	client.FMSConn.Handlers.Retry.PushBack(func(r *request.Request) {
		// Acceptance testing creates and deletes resources in quick succession.
		// The FMS onboarding process into Organizations is opaque to consumers.
		// Since we cannot reasonably check this status before receiving the error,
		// set the operation as retryable.
		switch r.Operation.Name {
		case "AssociateAdminAccount":
			if tfawserr.ErrMessageContains(r.Error, fms.ErrCodeInvalidOperationException, "Your AWS Organization is currently offboarding with AWS Firewall Manager. Please submit onboard request after offboarded.") {
				r.Retryable = aws.Bool(true)
			}
		case "DisassociateAdminAccount":
			if tfawserr.ErrMessageContains(r.Error, fms.ErrCodeInvalidOperationException, "Your AWS Organization is currently onboarding with AWS Firewall Manager and cannot be offboarded.") {
				r.Retryable = aws.Bool(true)
			}
		}
	})

	client.KafkaConn.Handlers.Retry.PushBack(func(r *request.Request) {
		if tfawserr.ErrMessageContains(r.Error, kafka.ErrCodeTooManyRequestsException, "Too Many Requests") {
			r.Retryable = aws.Bool(true)
		}
	})

	client.KinesisConn.Handlers.Retry.PushBack(func(r *request.Request) {
		if r.Operation.Name == "CreateStream" {
			if tfawserr.ErrMessageContains(r.Error, kinesis.ErrCodeLimitExceededException, "simultaneously be in CREATING or DELETING") {
				r.Retryable = aws.Bool(true)
			}
		}
		if r.Operation.Name == "CreateStream" || r.Operation.Name == "DeleteStream" {
			if tfawserr.ErrMessageContains(r.Error, kinesis.ErrCodeLimitExceededException, "Rate exceeded for stream") {
				r.Retryable = aws.Bool(true)
			}
		}
	})

	client.OrganizationsConn.Handlers.Retry.PushBack(func(r *request.Request) {
		// Retry on the following error:
		// ConcurrentModificationException: AWS Organizations can't complete your request because it conflicts with another attempt to modify the same entity. Try again later.
		if tfawserr.ErrMessageContains(r.Error, organizations.ErrCodeConcurrentModificationException, "Try again later") {
			r.Retryable = aws.Bool(true)
		}
	})

	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/17996
	client.SecurityHubConn.Handlers.Retry.PushBack(func(r *request.Request) {
		switch r.Operation.Name {
		case "EnableOrganizationAdminAccount":
			if tfawserr.ErrCodeEquals(r.Error, securityhub.ErrCodeResourceConflictException) {
				r.Retryable = aws.Bool(true)
			}
		}
	})

	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/19215
	client.SSOAdminConn.Handlers.Retry.PushBack(func(r *request.Request) {
		if r.Operation.Name == "AttachManagedPolicyToPermissionSet" || r.Operation.Name == "DetachManagedPolicyFromPermissionSet" {
			if tfawserr.ErrCodeEquals(r.Error, ssoadmin.ErrCodeConflictException) {
				r.Retryable = aws.Bool(true)
			}
		}
	})

	client.StorageGatewayConn.Handlers.Retry.PushBack(func(r *request.Request) {
		// InvalidGatewayRequestException: The specified gateway proxy network connection is busy.
		if tfawserr.ErrMessageContains(r.Error, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified gateway proxy network connection is busy") {
			r.Retryable = aws.Bool(true)
		}
	})

	client.WAFV2Conn.Handlers.Retry.PushBack(func(r *request.Request) {
		if tfawserr.ErrMessageContains(r.Error, wafv2.ErrCodeWAFInternalErrorException, "Retry your request") {
			r.Retryable = aws.Bool(true)
		}

		if tfawserr.ErrMessageContains(r.Error, wafv2.ErrCodeWAFServiceLinkedRoleErrorException, "Retry") {
			r.Retryable = aws.Bool(true)
		}

		if r.Operation.Name == "CreateIPSet" || r.Operation.Name == "CreateRegexPatternSet" ||
			r.Operation.Name == "CreateRuleGroup" || r.Operation.Name == "CreateWebACL" {
			// WAFv2 supports tag on create which can result in the below error codes according to the documentation
			if tfawserr.ErrMessageContains(r.Error, wafv2.ErrCodeWAFTagOperationException, "Retry your request") {
				r.Retryable = aws.Bool(true)
			}
			if tfawserr.ErrMessageContains(err, wafv2.ErrCodeWAFTagOperationInternalErrorException, "Retry your request") {
				r.Retryable = aws.Bool(true)
			}
		}
	})

	if !c.SkipGetEC2Platforms {
		supportedPlatforms, err := GetSupportedEC2Platforms(client.EC2Conn)
		if err != nil {
			// We intentionally fail *silently* because there's a chance
			// user just doesn't have ec2:DescribeAccountAttributes permissions
			log.Printf("[WARN] Unable to get supported EC2 platforms: %s", err)
		} else {
			client.SupportedPlatforms = supportedPlatforms
		}
	}

	return client, nil
}

func HasEC2Classic(platforms []string) bool {
	for _, p := range platforms {
		if p == "EC2" {
			return true
		}
	}
	return false
}

func GetSupportedEC2Platforms(conn *ec2.EC2) ([]string, error) {
	attrName := "supported-platforms"

	input := ec2.DescribeAccountAttributesInput{
		AttributeNames: []*string{aws.String(attrName)},
	}
	attributes, err := conn.DescribeAccountAttributes(&input)
	if err != nil {
		return nil, err
	}

	var platforms []string
	for _, attr := range attributes.AccountAttributes {
		if *attr.AttributeName == attrName {
			for _, v := range attr.AttributeValues {
				platforms = append(platforms, *v.AttributeValue)
			}
			break
		}
	}

	if len(platforms) == 0 {
		return nil, fmt.Errorf("No EC2 platforms detected")
	}

	return platforms, nil
}

// ReverseDNS switches a DNS hostname to reverse DNS and vice-versa.
func ReverseDNS(hostname string) string {
	parts := strings.Split(hostname, ".")

	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}

	return strings.Join(parts, ".")
}

// This is a global MutexKV for use within this plugin.
var GlobalMutexKV = NewMutexKV()
