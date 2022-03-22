package conns

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/route53domains"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/accessanalyzer"
	"github.com/aws/aws-sdk-go/service/account"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/aws/aws-sdk-go/service/alexaforbusiness"
	"github.com/aws/aws-sdk-go/service/amplify"
	"github.com/aws/aws-sdk-go/service/amplifybackend"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/aws/aws-sdk-go/service/appflow"
	"github.com/aws/aws-sdk-go/service/appintegrationsservice"
	"github.com/aws/aws-sdk-go/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go/service/applicationcostprofiler"
	"github.com/aws/aws-sdk-go/service/applicationdiscoveryservice"
	"github.com/aws/aws-sdk-go/service/applicationinsights"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/aws/aws-sdk-go/service/appregistry"
	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/aws/aws-sdk-go/service/auditmanager"
	"github.com/aws/aws-sdk-go/service/augmentedairuntime"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscalingplans"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/aws/aws-sdk-go/service/braket"
	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/aws/aws-sdk-go/service/chime"
	"github.com/aws/aws-sdk-go/service/cloud9"
	"github.com/aws/aws-sdk-go/service/cloudcontrolapi"
	"github.com/aws/aws-sdk-go/service/clouddirectory"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/cloudhsmv2"
	"github.com/aws/aws-sdk-go/service/cloudsearch"
	"github.com/aws/aws-sdk-go/service/cloudsearchdomain"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchrum"
	"github.com/aws/aws-sdk-go/service/codeartifact"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/aws/aws-sdk-go/service/codedeploy"
	"github.com/aws/aws-sdk-go/service/codeguruprofiler"
	"github.com/aws/aws-sdk-go/service/codegurureviewer"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/aws/aws-sdk-go/service/codestar"
	"github.com/aws/aws-sdk-go/service/codestarconnections"
	"github.com/aws/aws-sdk-go/service/codestarnotifications"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitosync"
	"github.com/aws/aws-sdk-go/service/comprehend"
	"github.com/aws/aws-sdk-go/service/comprehendmedical"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/aws/aws-sdk-go/service/connectcontactlens"
	"github.com/aws/aws-sdk-go/service/connectparticipant"
	"github.com/aws/aws-sdk-go/service/costandusagereportservice"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/aws/aws-sdk-go/service/dataexchange"
	"github.com/aws/aws-sdk-go/service/datapipeline"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/aws/aws-sdk-go/service/dax"
	"github.com/aws/aws-sdk-go/service/detective"
	"github.com/aws/aws-sdk-go/service/devicefarm"
	"github.com/aws/aws-sdk-go/service/devopsguru"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/aws/aws-sdk-go/service/dlm"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodbstreams"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2instanceconnect"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecrpublic"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/aws/aws-sdk-go/service/elasticinference"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/aws/aws-sdk-go/service/elastictranscoder"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/aws/aws-sdk-go/service/emrcontainers"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/aws/aws-sdk-go/service/finspace"
	"github.com/aws/aws-sdk-go/service/finspacedata"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-sdk-go/service/fis"
	"github.com/aws/aws-sdk-go/service/fms"
	"github.com/aws/aws-sdk-go/service/forecastqueryservice"
	"github.com/aws/aws-sdk-go/service/forecastservice"
	"github.com/aws/aws-sdk-go/service/frauddetector"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/aws/aws-sdk-go/service/glacier"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/aws/aws-sdk-go/service/gluedatabrew"
	"github.com/aws/aws-sdk-go/service/greengrass"
	"github.com/aws/aws-sdk-go/service/greengrassv2"
	"github.com/aws/aws-sdk-go/service/groundstation"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/aws/aws-sdk-go/service/health"
	"github.com/aws/aws-sdk-go/service/healthlake"
	"github.com/aws/aws-sdk-go/service/honeycode"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/identitystore"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/aws/aws-sdk-go/service/inspector"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/aws/aws-sdk-go/service/iot1clickdevicesservice"
	"github.com/aws/aws-sdk-go/service/iot1clickprojects"
	"github.com/aws/aws-sdk-go/service/iotanalytics"
	"github.com/aws/aws-sdk-go/service/iotdataplane"
	"github.com/aws/aws-sdk-go/service/iotdeviceadvisor"
	"github.com/aws/aws-sdk-go/service/iotevents"
	"github.com/aws/aws-sdk-go/service/ioteventsdata"
	"github.com/aws/aws-sdk-go/service/iotfleethub"
	"github.com/aws/aws-sdk-go/service/iotjobsdataplane"
	"github.com/aws/aws-sdk-go/service/iotsecuretunneling"
	"github.com/aws/aws-sdk-go/service/iotsitewise"
	"github.com/aws/aws-sdk-go/service/iotthingsgraph"
	"github.com/aws/aws-sdk-go/service/iotwireless"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	"github.com/aws/aws-sdk-go/service/kendra"
	"github.com/aws/aws-sdk-go/service/keyspaces"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
	"github.com/aws/aws-sdk-go/service/kinesisanalyticsv2"
	"github.com/aws/aws-sdk-go/service/kinesisvideo"
	"github.com/aws/aws-sdk-go/service/kinesisvideoarchivedmedia"
	"github.com/aws/aws-sdk-go/service/kinesisvideomedia"
	"github.com/aws/aws-sdk-go/service/kinesisvideosignalingchannels"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/aws/aws-sdk-go/service/lexmodelsv2"
	"github.com/aws/aws-sdk-go/service/lexruntimeservice"
	"github.com/aws/aws-sdk-go/service/lexruntimev2"
	"github.com/aws/aws-sdk-go/service/licensemanager"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/aws/aws-sdk-go/service/locationservice"
	"github.com/aws/aws-sdk-go/service/lookoutequipment"
	"github.com/aws/aws-sdk-go/service/lookoutforvision"
	"github.com/aws/aws-sdk-go/service/lookoutmetrics"
	"github.com/aws/aws-sdk-go/service/machinelearning"
	"github.com/aws/aws-sdk-go/service/macie"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/aws/aws-sdk-go/service/managedblockchain"
	"github.com/aws/aws-sdk-go/service/managedgrafana"
	"github.com/aws/aws-sdk-go/service/marketplacecatalog"
	"github.com/aws/aws-sdk-go/service/marketplacecommerceanalytics"
	"github.com/aws/aws-sdk-go/service/marketplaceentitlementservice"
	"github.com/aws/aws-sdk-go/service/marketplacemetering"
	"github.com/aws/aws-sdk-go/service/mediaconnect"
	"github.com/aws/aws-sdk-go/service/mediaconvert"
	"github.com/aws/aws-sdk-go/service/medialive"
	"github.com/aws/aws-sdk-go/service/mediapackage"
	"github.com/aws/aws-sdk-go/service/mediapackagevod"
	"github.com/aws/aws-sdk-go/service/mediastore"
	"github.com/aws/aws-sdk-go/service/mediastoredata"
	"github.com/aws/aws-sdk-go/service/mediatailor"
	"github.com/aws/aws-sdk-go/service/memorydb"
	"github.com/aws/aws-sdk-go/service/mgn"
	"github.com/aws/aws-sdk-go/service/migrationhub"
	"github.com/aws/aws-sdk-go/service/migrationhubconfig"
	"github.com/aws/aws-sdk-go/service/mobile"
	"github.com/aws/aws-sdk-go/service/mobileanalytics"
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/aws/aws-sdk-go/service/mturk"
	"github.com/aws/aws-sdk-go/service/mwaa"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/aws/aws-sdk-go/service/nimblestudio"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/aws/aws-sdk-go/service/opsworkscm"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/outposts"
	"github.com/aws/aws-sdk-go/service/personalize"
	"github.com/aws/aws-sdk-go/service/personalizeevents"
	"github.com/aws/aws-sdk-go/service/personalizeruntime"
	"github.com/aws/aws-sdk-go/service/pi"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/aws/aws-sdk-go/service/pinpointemail"
	"github.com/aws/aws-sdk-go/service/pinpointsmsvoice"
	"github.com/aws/aws-sdk-go/service/polly"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/aws/aws-sdk-go/service/proton"
	"github.com/aws/aws-sdk-go/service/qldb"
	"github.com/aws/aws-sdk-go/service/qldbsession"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rdsdataservice"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/aws/aws-sdk-go/service/redshiftdataapiservice"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/aws/aws-sdk-go/service/resourcegroups"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go/service/robomaker"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53recoverycontrolconfig"
	"github.com/aws/aws-sdk-go/service/route53recoveryreadiness"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/aws/aws-sdk-go/service/s3outposts"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/aws/aws-sdk-go/service/sagemakeredgemanager"
	"github.com/aws/aws-sdk-go/service/sagemakerfeaturestoreruntime"
	"github.com/aws/aws-sdk-go/service/sagemakerruntime"
	"github.com/aws/aws-sdk-go/service/savingsplans"
	"github.com/aws/aws-sdk-go/service/schemas"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/aws/aws-sdk-go/service/serverlessapplicationrepository"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/aws/aws-sdk-go/service/servicequotas"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/sesv2"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/aws/aws-sdk-go/service/shield"
	"github.com/aws/aws-sdk-go/service/signer"
	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/aws/aws-sdk-go/service/sms"
	"github.com/aws/aws-sdk-go/service/snowball"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssmcontacts"
	"github.com/aws/aws-sdk-go/service/ssmincidents"
	"github.com/aws/aws-sdk-go/service/sso"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/aws/aws-sdk-go/service/ssooidc"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/support"
	"github.com/aws/aws-sdk-go/service/swf"
	"github.com/aws/aws-sdk-go/service/synthetics"
	"github.com/aws/aws-sdk-go/service/textract"
	"github.com/aws/aws-sdk-go/service/timestreamquery"
	"github.com/aws/aws-sdk-go/service/timestreamwrite"
	"github.com/aws/aws-sdk-go/service/transcribeservice"
	"github.com/aws/aws-sdk-go/service/transcribestreamingservice"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/aws/aws-sdk-go/service/translate"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/aws/aws-sdk-go/service/wellarchitected"
	"github.com/aws/aws-sdk-go/service/workdocs"
	"github.com/aws/aws-sdk-go/service/worklink"
	"github.com/aws/aws-sdk-go/service/workmail"
	"github.com/aws/aws-sdk-go/service/workmailmessageflow"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/aws/aws-sdk-go/service/xray"
	awsbase "github.com/hashicorp/aws-sdk-go-base/v2"
	awsbasev1 "github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
	"github.com/hashicorp/terraform-provider-aws/version"
)

type Config struct {
	AccessKey                      string
	AllowedAccountIds              []string
	AssumeRole                     *awsbase.AssumeRole
	CustomCABundle                 string
	DefaultTagsConfig              *tftags.DefaultConfig
	EC2MetadataServiceEndpoint     string
	EC2MetadataServiceEndpointMode string
	Endpoints                      map[string]string
	ForbiddenAccountIds            []string
	HTTPProxy                      string
	IgnoreTagsConfig               *tftags.IgnoreConfig
	Insecure                       bool
	MaxRetries                     int
	Profile                        string
	Region                         string
	S3UsePathStyle                 bool
	SecretKey                      string
	SharedConfigFiles              []string
	SharedCredentialsFiles         []string
	SkipCredsValidation            bool
	SkipGetEC2Platforms            bool
	SkipMetadataApiCheck           bool
	SkipRegionValidation           bool
	SkipRequestingAccountId        bool
	STSRegion                      string
	TerraformVersion               string
	Token                          string
	UseDualStackEndpoint           bool
	UseFIPSEndpoint                bool
}

type AWSClient struct {
	AccessAnalyzerConn                *accessanalyzer.AccessAnalyzer
	AccountConn                       *account.Account
	AccountID                         string
	ACMConn                           *acm.ACM
	ACMPCAConn                        *acmpca.ACMPCA
	AlexaForBusinessConn              *alexaforbusiness.AlexaForBusiness
	AMPConn                           *prometheusservice.PrometheusService
	AmplifyBackendConn                *amplifybackend.AmplifyBackend
	AmplifyConn                       *amplify.Amplify
	APIGatewayConn                    *apigateway.APIGateway
	APIGatewayV2Conn                  *apigatewayv2.ApiGatewayV2
	AppAutoScalingConn                *applicationautoscaling.ApplicationAutoScaling
	AppConfigConn                     *appconfig.AppConfig
	AppFlowConn                       *appflow.Appflow
	AppIntegrationsConn               *appintegrationsservice.AppIntegrationsService
	ApplicationCostProfilerConn       *applicationcostprofiler.ApplicationCostProfiler
	ApplicationDiscoveryConn          *applicationdiscoveryservice.ApplicationDiscoveryService
	ApplicationInsightsConn           *applicationinsights.ApplicationInsights
	AppMeshConn                       *appmesh.AppMesh
	AppRegistryConn                   *appregistry.AppRegistry
	AppRunnerConn                     *apprunner.AppRunner
	AppStreamConn                     *appstream.AppStream
	AppSyncConn                       *appsync.AppSync
	AthenaConn                        *athena.Athena
	AuditManagerConn                  *auditmanager.AuditManager
	AugmentedAIRuntimeConn            *augmentedairuntime.AugmentedAIRuntime
	AutoScalingConn                   *autoscaling.AutoScaling
	AutoScalingPlansConn              *autoscalingplans.AutoScalingPlans
	BackupConn                        *backup.Backup
	BatchConn                         *batch.Batch
	BraketConn                        *braket.Braket
	BudgetsConn                       *budgets.Budgets
	ChimeConn                         *chime.Chime
	Cloud9Conn                        *cloud9.Cloud9
	CloudControlConn                  *cloudcontrolapi.CloudControlApi
	CloudDirectoryConn                *clouddirectory.CloudDirectory
	CloudFormationConn                *cloudformation.CloudFormation
	CloudFrontConn                    *cloudfront.CloudFront
	CloudHSMV2Conn                    *cloudhsmv2.CloudHSMV2
	CloudSearchConn                   *cloudsearch.CloudSearch
	CloudSearchDomainConn             *cloudsearchdomain.CloudSearchDomain
	CloudTrailConn                    *cloudtrail.CloudTrail
	CloudWatchConn                    *cloudwatch.CloudWatch
	CloudWatchLogsConn                *cloudwatchlogs.CloudWatchLogs
	CloudWatchRUMConn                 *cloudwatchrum.CloudWatchRUM
	CodeArtifactConn                  *codeartifact.CodeArtifact
	CodeBuildConn                     *codebuild.CodeBuild
	CodeCommitConn                    *codecommit.CodeCommit
	CodeDeployConn                    *codedeploy.CodeDeploy
	CodeGuruProfilerConn              *codeguruprofiler.CodeGuruProfiler
	CodeGuruReviewerConn              *codegurureviewer.CodeGuruReviewer
	CodePipelineConn                  *codepipeline.CodePipeline
	CodeStarConn                      *codestar.CodeStar
	CodeStarConnectionsConn           *codestarconnections.CodeStarConnections
	CodeStarNotificationsConn         *codestarnotifications.CodeStarNotifications
	CognitoIdentityConn               *cognitoidentity.CognitoIdentity
	CognitoIDPConn                    *cognitoidentityprovider.CognitoIdentityProvider
	CognitoSyncConn                   *cognitosync.CognitoSync
	ComprehendConn                    *comprehend.Comprehend
	ComprehendMedicalConn             *comprehendmedical.ComprehendMedical
	ConfigServiceConn                 *configservice.ConfigService
	ConnectConn                       *connect.Connect
	ConnectContactLensConn            *connectcontactlens.ConnectContactLens
	ConnectParticipantConn            *connectparticipant.ConnectParticipant
	CostExplorerConn                  *costexplorer.CostExplorer
	CURConn                           *costandusagereportservice.CostandUsageReportService
	DataExchangeConn                  *dataexchange.DataExchange
	DataPipelineConn                  *datapipeline.DataPipeline
	DataSyncConn                      *datasync.DataSync
	DAXConn                           *dax.DAX
	DefaultTagsConfig                 *tftags.DefaultConfig
	DetectiveConn                     *detective.Detective
	DeviceFarmConn                    *devicefarm.DeviceFarm
	DevOpsGuruConn                    *devopsguru.DevOpsGuru
	DirectConnectConn                 *directconnect.DirectConnect
	DLMConn                           *dlm.DLM
	DMSConn                           *databasemigrationservice.DatabaseMigrationService
	DNSSuffix                         string
	DocDBConn                         *docdb.DocDB
	DSConn                            *directoryservice.DirectoryService
	DynamoDBConn                      *dynamodb.DynamoDB
	DynamoDBStreamsConn               *dynamodbstreams.DynamoDBStreams
	EC2Conn                           *ec2.EC2
	EC2InstanceConnectConn            *ec2instanceconnect.EC2InstanceConnect
	ECRConn                           *ecr.ECR
	ECRPublicConn                     *ecrpublic.ECRPublic
	ECSConn                           *ecs.ECS
	EFSConn                           *efs.EFS
	EKSConn                           *eks.EKS
	ElastiCacheConn                   *elasticache.ElastiCache
	ElasticBeanstalkConn              *elasticbeanstalk.ElasticBeanstalk
	ElasticInferenceConn              *elasticinference.ElasticInference
	ElasticsearchConn                 *elasticsearch.ElasticsearchService
	ElasticTranscoderConn             *elastictranscoder.ElasticTranscoder
	ELBConn                           *elb.ELB
	ELBV2Conn                         *elbv2.ELBV2
	EMRConn                           *emr.EMR
	EMRContainersConn                 *emrcontainers.EMRContainers
	EventsConn                        *eventbridge.EventBridge
	FinSpaceConn                      *finspace.Finspace
	FinSpaceDataConn                  *finspacedata.FinSpaceData
	FirehoseConn                      *firehose.Firehose
	FISConn                           *fis.FIS
	FMSConn                           *fms.FMS
	ForecastConn                      *forecastservice.ForecastService
	ForecastQueryConn                 *forecastqueryservice.ForecastQueryService
	FraudDetectorConn                 *frauddetector.FraudDetector
	FSxConn                           *fsx.FSx
	GameLiftConn                      *gamelift.GameLift
	GlacierConn                       *glacier.Glacier
	GlobalAcceleratorConn             *globalaccelerator.GlobalAccelerator
	GlueConn                          *glue.Glue
	GlueDataBrewConn                  *gluedatabrew.GlueDataBrew
	GrafanaConn                       *managedgrafana.ManagedGrafana
	GreengrassConn                    *greengrass.Greengrass
	GreengrassV2Conn                  *greengrassv2.GreengrassV2
	GroundStationConn                 *groundstation.GroundStation
	GuardDutyConn                     *guardduty.GuardDuty
	HealthConn                        *health.Health
	HealthLakeConn                    *healthlake.HealthLake
	HoneycodeConn                     *honeycode.Honeycode
	IAMConn                           *iam.IAM
	IdentityStoreConn                 *identitystore.IdentityStore
	IgnoreTagsConfig                  *tftags.IgnoreConfig
	ImageBuilderConn                  *imagebuilder.Imagebuilder
	InspectorConn                     *inspector.Inspector
	IoT1ClickDevicesConn              *iot1clickdevicesservice.IoT1ClickDevicesService
	IoT1ClickProjectsConn             *iot1clickprojects.IoT1ClickProjects
	IoTAnalyticsConn                  *iotanalytics.IoTAnalytics
	IoTConn                           *iot.IoT
	IoTDataPlaneConn                  *iotdataplane.IoTDataPlane
	IoTDeviceAdvisorConn              *iotdeviceadvisor.IoTDeviceAdvisor
	IoTEventsConn                     *iotevents.IoTEvents
	IoTEventsDataConn                 *ioteventsdata.IoTEventsData
	IoTFleetHubConn                   *iotfleethub.IoTFleetHub
	IoTJobsDataPlaneConn              *iotjobsdataplane.IoTJobsDataPlane
	IoTSecureTunnelingConn            *iotsecuretunneling.IoTSecureTunneling
	IoTSiteWiseConn                   *iotsitewise.IoTSiteWise
	IoTThingsGraphConn                *iotthingsgraph.IoTThingsGraph
	IoTWirelessConn                   *iotwireless.IoTWireless
	KafkaConn                         *kafka.Kafka
	KafkaConnectConn                  *kafkaconnect.KafkaConnect
	KendraConn                        *kendra.Kendra
	KeyspacesConn                     *keyspaces.Keyspaces
	KinesisAnalyticsConn              *kinesisanalytics.KinesisAnalytics
	KinesisAnalyticsV2Conn            *kinesisanalyticsv2.KinesisAnalyticsV2
	KinesisConn                       *kinesis.Kinesis
	KinesisVideoArchivedMediaConn     *kinesisvideoarchivedmedia.KinesisVideoArchivedMedia
	KinesisVideoConn                  *kinesisvideo.KinesisVideo
	KinesisVideoMediaConn             *kinesisvideomedia.KinesisVideoMedia
	KinesisVideoSignalingChannelsConn *kinesisvideosignalingchannels.KinesisVideoSignalingChannels
	KMSConn                           *kms.KMS
	LakeFormationConn                 *lakeformation.LakeFormation
	LambdaConn                        *lambda.Lambda
	LexModelsConn                     *lexmodelbuildingservice.LexModelBuildingService
	LexModelsV2Conn                   *lexmodelsv2.LexModelsV2
	LexRuntimeConn                    *lexruntimeservice.LexRuntimeService
	LexRuntimeV2Conn                  *lexruntimev2.LexRuntimeV2
	LicenseManagerConn                *licensemanager.LicenseManager
	LightsailConn                     *lightsail.Lightsail
	LocationConn                      *locationservice.LocationService
	LookoutEquipmentConn              *lookoutequipment.LookoutEquipment
	LookoutForVisionConn              *lookoutforvision.LookoutForVision
	LookoutMetricsConn                *lookoutmetrics.LookoutMetrics
	MachineLearningConn               *machinelearning.MachineLearning
	Macie2Conn                        *macie2.Macie2
	MacieConn                         *macie.Macie
	ManagedBlockchainConn             *managedblockchain.ManagedBlockchain
	MarketplaceCatalogConn            *marketplacecatalog.MarketplaceCatalog
	MarketplaceCommerceAnalyticsConn  *marketplacecommerceanalytics.MarketplaceCommerceAnalytics
	MarketplaceEntitlementConn        *marketplaceentitlementservice.MarketplaceEntitlementService
	MarketplaceMeteringConn           *marketplacemetering.MarketplaceMetering
	MediaConnectConn                  *mediaconnect.MediaConnect
	MediaConvertAccountConn           *mediaconvert.MediaConvert
	MediaConvertConn                  *mediaconvert.MediaConvert
	MediaLiveConn                     *medialive.MediaLive
	MediaPackageConn                  *mediapackage.MediaPackage
	MediaPackageVODConn               *mediapackagevod.MediaPackageVod
	MediaStoreConn                    *mediastore.MediaStore
	MediaStoreDataConn                *mediastoredata.MediaStoreData
	MediaTailorConn                   *mediatailor.MediaTailor
	MemoryDBConn                      *memorydb.MemoryDB
	MgnConn                           *mgn.Mgn
	MigrationHubConfigConn            *migrationhubconfig.MigrationHubConfig
	MigrationHubConn                  *migrationhub.MigrationHub
	MobileAnalyticsConn               *mobileanalytics.MobileAnalytics
	MobileConn                        *mobile.Mobile
	MQConn                            *mq.MQ
	MTurkConn                         *mturk.MTurk
	MWAAConn                          *mwaa.MWAA
	NeptuneConn                       *neptune.Neptune
	NetworkFirewallConn               *networkfirewall.NetworkFirewall
	NetworkManagerConn                *networkmanager.NetworkManager
	NimbleStudioConn                  *nimblestudio.NimbleStudio
	OpsWorksCMConn                    *opsworkscm.OpsWorksCM
	OpsWorksConn                      *opsworks.OpsWorks
	OrganizationsConn                 *organizations.Organizations
	OutpostsConn                      *outposts.Outposts
	Partition                         string
	PersonalizeConn                   *personalize.Personalize
	PersonalizeEventsConn             *personalizeevents.PersonalizeEvents
	PersonalizeRuntimeConn            *personalizeruntime.PersonalizeRuntime
	PIConn                            *pi.PI
	PinpointConn                      *pinpoint.Pinpoint
	PinpointEmailConn                 *pinpointemail.PinpointEmail
	PinpointSMSVoiceConn              *pinpointsmsvoice.PinpointSMSVoice
	PollyConn                         *polly.Polly
	PricingConn                       *pricing.Pricing
	ProtonConn                        *proton.Proton
	QLDBConn                          *qldb.QLDB
	QLDBSessionConn                   *qldbsession.QLDBSession
	QuickSightConn                    *quicksight.QuickSight
	RAMConn                           *ram.RAM
	RDSConn                           *rds.RDS
	RDSDataConn                       *rdsdataservice.RDSDataService
	RedshiftConn                      *redshift.Redshift
	RedshiftDataConn                  *redshiftdataapiservice.RedshiftDataAPIService
	Region                            string
	RekognitionConn                   *rekognition.Rekognition
	ResourceGroupsConn                *resourcegroups.ResourceGroups
	ResourceGroupsTaggingAPIConn      *resourcegroupstaggingapi.ResourceGroupsTaggingAPI
	ReverseDNSPrefix                  string
	RoboMakerConn                     *robomaker.RoboMaker
	Route53Conn                       *route53.Route53
	Route53DomainsConn                *route53domains.Client
	Route53RecoveryControlConfigConn  *route53recoverycontrolconfig.Route53RecoveryControlConfig
	Route53RecoveryReadinessConn      *route53recoveryreadiness.Route53RecoveryReadiness
	Route53ResolverConn               *route53resolver.Route53Resolver
	S3Conn                            *s3.S3
	S3ConnURICleaningDisabled         *s3.S3
	S3ControlConn                     *s3control.S3Control
	S3OutpostsConn                    *s3outposts.S3Outposts
	SageMakerConn                     *sagemaker.SageMaker
	SageMakerEdgeManagerConn          *sagemakeredgemanager.SagemakerEdgeManager
	SageMakerFeatureStoreRuntimeConn  *sagemakerfeaturestoreruntime.SageMakerFeatureStoreRuntime
	SageMakerRuntimeConn              *sagemakerruntime.SageMakerRuntime
	SavingsPlansConn                  *savingsplans.SavingsPlans
	SchemasConn                       *schemas.Schemas
	SecretsManagerConn                *secretsmanager.SecretsManager
	SecurityHubConn                   *securityhub.SecurityHub
	ServerlessRepoConn                *serverlessapplicationrepository.ServerlessApplicationRepository
	ServiceCatalogConn                *servicecatalog.ServiceCatalog
	ServiceDiscoveryConn              *servicediscovery.ServiceDiscovery
	ServiceQuotasConn                 *servicequotas.ServiceQuotas
	SESConn                           *ses.SES
	Session                           *session.Session
	SESV2Conn                         *sesv2.SESV2
	SFNConn                           *sfn.SFN
	ShieldConn                        *shield.Shield
	SignerConn                        *signer.Signer
	SimpleDBConn                      *simpledb.SimpleDB
	SMSConn                           *sms.SMS
	SnowballConn                      *snowball.Snowball
	SNSConn                           *sns.SNS
	SQSConn                           *sqs.SQS
	SSMConn                           *ssm.SSM
	SSMContactsConn                   *ssmcontacts.SSMContacts
	SSMIncidentsConn                  *ssmincidents.SSMIncidents
	SSOAdminConn                      *ssoadmin.SSOAdmin
	SSOConn                           *sso.SSO
	SSOOIDCConn                       *ssooidc.SSOOIDC
	StorageGatewayConn                *storagegateway.StorageGateway
	STSConn                           *sts.STS
	SupportConn                       *support.Support
	SupportedPlatforms                []string
	SWFConn                           *swf.SWF
	SyntheticsConn                    *synthetics.Synthetics
	TerraformVersion                  string
	TextractConn                      *textract.Textract
	TimestreamQueryConn               *timestreamquery.TimestreamQuery
	TimestreamWriteConn               *timestreamwrite.TimestreamWrite
	TranscribeConn                    *transcribeservice.TranscribeService
	TranscribeStreamingConn           *transcribestreamingservice.TranscribeStreamingService
	TransferConn                      *transfer.Transfer
	TranslateConn                     *translate.Translate
	WAFConn                           *waf.WAF
	WAFRegionalConn                   *wafregional.WAFRegional
	WAFV2Conn                         *wafv2.WAFV2
	WellArchitectedConn               *wellarchitected.WellArchitected
	WorkDocsConn                      *workdocs.WorkDocs
	WorkLinkConn                      *worklink.WorkLink
	WorkMailConn                      *workmail.WorkMail
	WorkMailMessageFlowConn           *workmailmessageflow.WorkMailMessageFlow
	WorkSpacesConn                    *workspaces.WorkSpaces
	XRayConn                          *xray.XRay
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
func (c *Config) Client(ctx context.Context) (interface{}, diag.Diagnostics) {
	awsbaseConfig := awsbase.Config{
		AccessKey:               c.AccessKey,
		APNInfo:                 StdUserAgentProducts(c.TerraformVersion),
		CallerDocumentationURL:  "https://registry.terraform.io/providers/hashicorp/aws",
		CallerName:              "Terraform AWS Provider",
		IamEndpoint:             c.Endpoints[names.IAM],
		Insecure:                c.Insecure,
		HTTPProxy:               c.HTTPProxy,
		MaxRetries:              c.MaxRetries,
		Profile:                 c.Profile,
		Region:                  c.Region,
		SecretKey:               c.SecretKey,
		SkipCredsValidation:     c.SkipCredsValidation,
		SkipEC2MetadataApiCheck: c.SkipMetadataApiCheck,
		SkipRequestingAccountId: c.SkipRequestingAccountId,
		StsEndpoint:             c.Endpoints[names.STS],
		Token:                   c.Token,
		UseDualStackEndpoint:    c.UseDualStackEndpoint,
		UseFIPSEndpoint:         c.UseFIPSEndpoint,
	}

	if c.AssumeRole != nil && c.AssumeRole.RoleARN != "" {
		awsbaseConfig.AssumeRole = c.AssumeRole
	}

	if c.CustomCABundle != "" {
		awsbaseConfig.CustomCABundle = c.CustomCABundle
	}

	if c.EC2MetadataServiceEndpoint != "" {
		awsbaseConfig.EC2MetadataServiceEndpoint = c.EC2MetadataServiceEndpoint
		awsbaseConfig.EC2MetadataServiceEndpointMode = c.EC2MetadataServiceEndpointMode
	}

	if len(c.SharedConfigFiles) != 0 {
		awsbaseConfig.SharedConfigFiles = c.SharedConfigFiles
	}

	if len(c.SharedCredentialsFiles) != 0 {
		awsbaseConfig.SharedCredentialsFiles = c.SharedCredentialsFiles
	}

	if c.STSRegion != "" {
		awsbaseConfig.StsRegion = c.STSRegion
	}

	cfg, err := awsbase.GetAwsConfig(ctx, &awsbaseConfig)
	if err != nil {
		return nil, diag.Errorf("error configuring Terraform AWS Provider: %s", err)
	}

	if !c.SkipRegionValidation {
		if err := awsbase.ValidateRegion(cfg.Region); err != nil {
			return nil, diag.FromErr(err)
		}
	}
	c.Region = cfg.Region

	sess, err := awsbasev1.GetSession(&cfg, &awsbaseConfig)
	if err != nil {
		return nil, diag.Errorf("error creating AWS SDK v1 session: %s", err)
	}

	accountID, partition, err := awsbase.GetAwsAccountIDAndPartition(ctx, cfg, &awsbaseConfig)
	if err != nil {
		return nil, diag.Errorf("error retrieving account details: %s", err)
	}

	if accountID == "" {
		log.Println("[WARN] AWS account ID not found for provider. See https://www.terraform.io/docs/providers/aws/index.html#skip_requesting_account_id for implications.")
	}

	if len(c.ForbiddenAccountIds) > 0 {
		for _, forbiddenAccountID := range c.AllowedAccountIds {
			if accountID == forbiddenAccountID {
				return nil, diag.Errorf("AWS Account ID not allowed: %s", accountID)
			}
		}
	}
	if len(c.AllowedAccountIds) > 0 {
		found := false
		for _, allowedAccountID := range c.AllowedAccountIds {
			if accountID == allowedAccountID {
				found = true
				break
			}
		}
		if !found {
			return nil, diag.Errorf("AWS Account ID not allowed: %s", accountID)
		}
	}

	DNSSuffix := "amazonaws.com"
	if p, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), c.Region); ok {
		DNSSuffix = p.DNSSuffix()
	}

	client := &AWSClient{
		AccessAnalyzerConn:                accessanalyzer.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.AccessAnalyzer])})),
		AccountConn:                       account.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Account])})),
		AccountID:                         accountID,
		ACMConn:                           acm.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.ACM])})),
		ACMPCAConn:                        acmpca.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.ACMPCA])})),
		AlexaForBusinessConn:              alexaforbusiness.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.AlexaForBusiness])})),
		AMPConn:                           prometheusservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.AMP])})),
		AmplifyBackendConn:                amplifybackend.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.AmplifyBackend])})),
		AmplifyConn:                       amplify.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Amplify])})),
		APIGatewayConn:                    apigateway.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.APIGateway])})),
		APIGatewayV2Conn:                  apigatewayv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.APIGatewayV2])})),
		AppAutoScalingConn:                applicationautoscaling.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.AppAutoScaling])})),
		AppConfigConn:                     appconfig.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.AppConfig])})),
		AppFlowConn:                       appflow.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.AppFlow])})),
		AppIntegrationsConn:               appintegrationsservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.AppIntegrations])})),
		ApplicationCostProfilerConn:       applicationcostprofiler.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.ApplicationCostProfiler])})),
		ApplicationDiscoveryConn:          applicationdiscoveryservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.ApplicationDiscovery])})),
		ApplicationInsightsConn:           applicationinsights.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.ApplicationInsights])})),
		AppMeshConn:                       appmesh.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.AppMesh])})),
		AppRegistryConn:                   appregistry.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.AppRegistry])})),
		AppRunnerConn:                     apprunner.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.AppRunner])})),
		AppStreamConn:                     appstream.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.AppStream])})),
		AppSyncConn:                       appsync.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.AppSync])})),
		AthenaConn:                        athena.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Athena])})),
		AuditManagerConn:                  auditmanager.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.AuditManager])})),
		AugmentedAIRuntimeConn:            augmentedairuntime.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.AugmentedAIRuntime])})),
		AutoScalingConn:                   autoscaling.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.AutoScaling])})),
		AutoScalingPlansConn:              autoscalingplans.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.AutoScalingPlans])})),
		BackupConn:                        backup.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Backup])})),
		BatchConn:                         batch.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Batch])})),
		BraketConn:                        braket.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Braket])})),
		BudgetsConn:                       budgets.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Budgets])})),
		ChimeConn:                         chime.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Chime])})),
		Cloud9Conn:                        cloud9.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Cloud9])})),
		CloudControlConn:                  cloudcontrolapi.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.CloudControl])})),
		CloudDirectoryConn:                clouddirectory.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.CloudDirectory])})),
		CloudFormationConn:                cloudformation.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.CloudFormation])})),
		CloudFrontConn:                    cloudfront.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.CloudFront])})),
		CloudHSMV2Conn:                    cloudhsmv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.CloudHSMV2])})),
		CloudSearchConn:                   cloudsearch.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.CloudSearch])})),
		CloudSearchDomainConn:             cloudsearchdomain.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.CloudSearchDomain])})),
		CloudTrailConn:                    cloudtrail.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.CloudTrail])})),
		CloudWatchConn:                    cloudwatch.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.CloudWatch])})),
		CloudWatchLogsConn:                cloudwatchlogs.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.CloudWatchLogs])})),
		CloudWatchRUMConn:                 cloudwatchrum.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.CloudWatchRUM])})),
		CodeArtifactConn:                  codeartifact.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.CodeArtifact])})),
		CodeBuildConn:                     codebuild.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.CodeBuild])})),
		CodeCommitConn:                    codecommit.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.CodeCommit])})),
		CodeDeployConn:                    codedeploy.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.CodeDeploy])})),
		CodeGuruProfilerConn:              codeguruprofiler.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.CodeGuruProfiler])})),
		CodeGuruReviewerConn:              codegurureviewer.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.CodeGuruReviewer])})),
		CodePipelineConn:                  codepipeline.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.CodePipeline])})),
		CodeStarConn:                      codestar.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.CodeStar])})),
		CodeStarConnectionsConn:           codestarconnections.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.CodeStarConnections])})),
		CodeStarNotificationsConn:         codestarnotifications.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.CodeStarNotifications])})),
		CognitoIdentityConn:               cognitoidentity.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.CognitoIdentity])})),
		CognitoIDPConn:                    cognitoidentityprovider.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.CognitoIDP])})),
		CognitoSyncConn:                   cognitosync.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.CognitoSync])})),
		ComprehendConn:                    comprehend.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Comprehend])})),
		ComprehendMedicalConn:             comprehendmedical.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.ComprehendMedical])})),
		ConfigServiceConn:                 configservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.ConfigService])})),
		ConnectConn:                       connect.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Connect])})),
		ConnectContactLensConn:            connectcontactlens.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.ConnectContactLens])})),
		ConnectParticipantConn:            connectparticipant.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.ConnectParticipant])})),
		CostExplorerConn:                  costexplorer.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.CostExplorer])})),
		CURConn:                           costandusagereportservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.CUR])})),
		DataExchangeConn:                  dataexchange.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.DataExchange])})),
		DataPipelineConn:                  datapipeline.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.DataPipeline])})),
		DataSyncConn:                      datasync.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.DataSync])})),
		DAXConn:                           dax.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.DAX])})),
		DefaultTagsConfig:                 c.DefaultTagsConfig,
		DetectiveConn:                     detective.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Detective])})),
		DeviceFarmConn:                    devicefarm.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.DeviceFarm])})),
		DevOpsGuruConn:                    devopsguru.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.DevOpsGuru])})),
		DirectConnectConn:                 directconnect.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.DirectConnect])})),
		DLMConn:                           dlm.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.DLM])})),
		DMSConn:                           databasemigrationservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.DMS])})),
		DNSSuffix:                         DNSSuffix,
		DocDBConn:                         docdb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.DocDB])})),
		DSConn:                            directoryservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.DS])})),
		DynamoDBConn:                      dynamodb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.DynamoDB])})),
		DynamoDBStreamsConn:               dynamodbstreams.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.DynamoDBStreams])})),
		EC2Conn:                           ec2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.EC2])})),
		EC2InstanceConnectConn:            ec2instanceconnect.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.EC2InstanceConnect])})),
		ECRConn:                           ecr.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.ECR])})),
		ECRPublicConn:                     ecrpublic.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.ECRPublic])})),
		ECSConn:                           ecs.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.ECS])})),
		EFSConn:                           efs.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.EFS])})),
		EKSConn:                           eks.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.EKS])})),
		ElastiCacheConn:                   elasticache.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.ElastiCache])})),
		ElasticBeanstalkConn:              elasticbeanstalk.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.ElasticBeanstalk])})),
		ElasticInferenceConn:              elasticinference.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.ElasticInference])})),
		ElasticsearchConn:                 elasticsearch.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Elasticsearch])})),
		ElasticTranscoderConn:             elastictranscoder.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.ElasticTranscoder])})),
		ELBConn:                           elb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.ELB])})),
		ELBV2Conn:                         elbv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.ELBV2])})),
		EMRConn:                           emr.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.EMR])})),
		EMRContainersConn:                 emrcontainers.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.EMRContainers])})),
		EventsConn:                        eventbridge.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Events])})),
		FinSpaceConn:                      finspace.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.FinSpace])})),
		FinSpaceDataConn:                  finspacedata.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.FinSpaceData])})),
		FirehoseConn:                      firehose.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Firehose])})),
		FISConn:                           fis.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.FIS])})),
		FMSConn:                           fms.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.FMS])})),
		ForecastConn:                      forecastservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Forecast])})),
		ForecastQueryConn:                 forecastqueryservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.ForecastQuery])})),
		FraudDetectorConn:                 frauddetector.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.FraudDetector])})),
		FSxConn:                           fsx.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.FSx])})),
		GameLiftConn:                      gamelift.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.GameLift])})),
		GlacierConn:                       glacier.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Glacier])})),
		GlueConn:                          glue.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Glue])})),
		GlueDataBrewConn:                  gluedatabrew.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.GlueDataBrew])})),
		GrafanaConn:                       managedgrafana.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Grafana])})),
		GreengrassConn:                    greengrass.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Greengrass])})),
		GreengrassV2Conn:                  greengrassv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.GreengrassV2])})),
		GroundStationConn:                 groundstation.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.GroundStation])})),
		GuardDutyConn:                     guardduty.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.GuardDuty])})),
		HealthConn:                        health.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Health])})),
		HealthLakeConn:                    healthlake.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.HealthLake])})),
		HoneycodeConn:                     honeycode.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Honeycode])})),
		IAMConn:                           iam.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.IAM])})),
		IdentityStoreConn:                 identitystore.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.IdentityStore])})),
		IgnoreTagsConfig:                  c.IgnoreTagsConfig,
		ImageBuilderConn:                  imagebuilder.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.ImageBuilder])})),
		InspectorConn:                     inspector.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Inspector])})),
		IoT1ClickDevicesConn:              iot1clickdevicesservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.IoT1ClickDevices])})),
		IoT1ClickProjectsConn:             iot1clickprojects.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.IoT1ClickProjects])})),
		IoTAnalyticsConn:                  iotanalytics.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.IoTAnalytics])})),
		IoTConn:                           iot.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.IoT])})),
		IoTDataPlaneConn:                  iotdataplane.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.IoTDataPlane])})),
		IoTDeviceAdvisorConn:              iotdeviceadvisor.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.IoTDeviceAdvisor])})),
		IoTEventsConn:                     iotevents.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.IoTEvents])})),
		IoTEventsDataConn:                 ioteventsdata.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.IoTEventsData])})),
		IoTFleetHubConn:                   iotfleethub.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.IoTFleetHub])})),
		IoTJobsDataPlaneConn:              iotjobsdataplane.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.IoTJobsDataPlane])})),
		IoTSecureTunnelingConn:            iotsecuretunneling.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.IoTSecureTunneling])})),
		IoTSiteWiseConn:                   iotsitewise.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.IoTSiteWise])})),
		IoTThingsGraphConn:                iotthingsgraph.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.IoTThingsGraph])})),
		IoTWirelessConn:                   iotwireless.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.IoTWireless])})),
		KafkaConn:                         kafka.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Kafka])})),
		KafkaConnectConn:                  kafkaconnect.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.KafkaConnect])})),
		KendraConn:                        kendra.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Kendra])})),
		KeyspacesConn:                     keyspaces.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Keyspaces])})),
		KinesisAnalyticsConn:              kinesisanalytics.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.KinesisAnalytics])})),
		KinesisAnalyticsV2Conn:            kinesisanalyticsv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.KinesisAnalyticsV2])})),
		KinesisConn:                       kinesis.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Kinesis])})),
		KinesisVideoArchivedMediaConn:     kinesisvideoarchivedmedia.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.KinesisVideoArchivedMedia])})),
		KinesisVideoConn:                  kinesisvideo.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.KinesisVideo])})),
		KinesisVideoMediaConn:             kinesisvideomedia.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.KinesisVideoMedia])})),
		KinesisVideoSignalingChannelsConn: kinesisvideosignalingchannels.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.KinesisVideoSignalingChannels])})),
		KMSConn:                           kms.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.KMS])})),
		LakeFormationConn:                 lakeformation.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.LakeFormation])})),
		LambdaConn:                        lambda.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Lambda])})),
		LexModelsConn:                     lexmodelbuildingservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.LexModels])})),
		LexModelsV2Conn:                   lexmodelsv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.LexModelsV2])})),
		LexRuntimeConn:                    lexruntimeservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.LexRuntime])})),
		LexRuntimeV2Conn:                  lexruntimev2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.LexRuntimeV2])})),
		LicenseManagerConn:                licensemanager.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.LicenseManager])})),
		LightsailConn:                     lightsail.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Lightsail])})),
		LocationConn:                      locationservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Location])})),
		LookoutEquipmentConn:              lookoutequipment.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.LookoutEquipment])})),
		LookoutForVisionConn:              lookoutforvision.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.LookoutForVision])})),
		LookoutMetricsConn:                lookoutmetrics.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.LookoutMetrics])})),
		MachineLearningConn:               machinelearning.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.MachineLearning])})),
		Macie2Conn:                        macie2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Macie2])})),
		MacieConn:                         macie.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Macie])})),
		ManagedBlockchainConn:             managedblockchain.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.ManagedBlockchain])})),
		MarketplaceCatalogConn:            marketplacecatalog.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.MarketplaceCatalog])})),
		MarketplaceCommerceAnalyticsConn:  marketplacecommerceanalytics.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.MarketplaceCommerceAnalytics])})),
		MarketplaceEntitlementConn:        marketplaceentitlementservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.MarketplaceEntitlement])})),
		MarketplaceMeteringConn:           marketplacemetering.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.MarketplaceMetering])})),
		MediaConnectConn:                  mediaconnect.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.MediaConnect])})),
		MediaConvertConn:                  mediaconvert.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.MediaConvert])})),
		MediaLiveConn:                     medialive.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.MediaLive])})),
		MediaPackageConn:                  mediapackage.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.MediaPackage])})),
		MediaPackageVODConn:               mediapackagevod.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.MediaPackageVOD])})),
		MediaStoreConn:                    mediastore.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.MediaStore])})),
		MediaStoreDataConn:                mediastoredata.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.MediaStoreData])})),
		MediaTailorConn:                   mediatailor.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.MediaTailor])})),
		MemoryDBConn:                      memorydb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.MemoryDB])})),
		MgnConn:                           mgn.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Mgn])})),
		MigrationHubConfigConn:            migrationhubconfig.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.MigrationHubConfig])})),
		MigrationHubConn:                  migrationhub.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.MigrationHub])})),
		MobileAnalyticsConn:               mobileanalytics.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.MobileAnalytics])})),
		MobileConn:                        mobile.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Mobile])})),
		MQConn:                            mq.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.MQ])})),
		MTurkConn:                         mturk.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.MTurk])})),
		MWAAConn:                          mwaa.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.MWAA])})),
		NeptuneConn:                       neptune.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Neptune])})),
		NetworkFirewallConn:               networkfirewall.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.NetworkFirewall])})),
		NetworkManagerConn:                networkmanager.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.NetworkManager])})),
		NimbleStudioConn:                  nimblestudio.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.NimbleStudio])})),
		OpsWorksCMConn:                    opsworkscm.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.OpsWorksCM])})),
		OpsWorksConn:                      opsworks.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.OpsWorks])})),
		OrganizationsConn:                 organizations.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Organizations])})),
		OutpostsConn:                      outposts.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Outposts])})),
		Partition:                         partition,
		PersonalizeConn:                   personalize.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Personalize])})),
		PersonalizeEventsConn:             personalizeevents.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.PersonalizeEvents])})),
		PersonalizeRuntimeConn:            personalizeruntime.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.PersonalizeRuntime])})),
		PIConn:                            pi.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.PI])})),
		PinpointConn:                      pinpoint.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Pinpoint])})),
		PinpointEmailConn:                 pinpointemail.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.PinpointEmail])})),
		PinpointSMSVoiceConn:              pinpointsmsvoice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.PinpointSMSVoice])})),
		PollyConn:                         polly.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Polly])})),
		PricingConn:                       pricing.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Pricing])})),
		ProtonConn:                        proton.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Proton])})),
		QLDBConn:                          qldb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.QLDB])})),
		QLDBSessionConn:                   qldbsession.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.QLDBSession])})),
		QuickSightConn:                    quicksight.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.QuickSight])})),
		RAMConn:                           ram.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.RAM])})),
		RDSConn:                           rds.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.RDS])})),
		RDSDataConn:                       rdsdataservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.RDSData])})),
		RedshiftConn:                      redshift.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Redshift])})),
		RedshiftDataConn:                  redshiftdataapiservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.RedshiftData])})),
		Region:                            c.Region,
		RekognitionConn:                   rekognition.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Rekognition])})),
		ResourceGroupsConn:                resourcegroups.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.ResourceGroups])})),
		ResourceGroupsTaggingAPIConn:      resourcegroupstaggingapi.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.ResourceGroupsTaggingAPI])})),
		ReverseDNSPrefix:                  ReverseDNS(DNSSuffix),
		RoboMakerConn:                     robomaker.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.RoboMaker])})),
		Route53DomainsConn: route53domains.NewFromConfig(cfg, func(o *route53domains.Options) {
			if endpoint := c.Endpoints[names.Route53Domains]; endpoint != "" {
				o.EndpointResolver = route53domains.EndpointResolverFromURL(endpoint)
			} else if partition == endpoints.AwsPartitionID {
				// Route 53 Domains is only available in AWS Commercial us-east-1 Region.
				o.Region = endpoints.UsEast1RegionID
			}
		}),
		Route53RecoveryControlConfigConn: route53recoverycontrolconfig.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Route53RecoveryControlConfig])})),
		Route53RecoveryReadinessConn:     route53recoveryreadiness.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Route53RecoveryReadiness])})),
		Route53ResolverConn:              route53resolver.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Route53Resolver])})),
		S3ControlConn:                    s3control.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.S3Control])})),
		S3OutpostsConn:                   s3outposts.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.S3Outposts])})),
		SageMakerConn:                    sagemaker.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.SageMaker])})),
		SageMakerEdgeManagerConn:         sagemakeredgemanager.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.SageMakerEdgeManager])})),
		SageMakerFeatureStoreRuntimeConn: sagemakerfeaturestoreruntime.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.SageMakerFeatureStoreRuntime])})),
		SageMakerRuntimeConn:             sagemakerruntime.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.SageMakerRuntime])})),
		SavingsPlansConn:                 savingsplans.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.SavingsPlans])})),
		SchemasConn:                      schemas.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Schemas])})),
		SecretsManagerConn:               secretsmanager.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.SecretsManager])})),
		SecurityHubConn:                  securityhub.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.SecurityHub])})),
		ServerlessRepoConn:               serverlessapplicationrepository.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.ServerlessRepo])})),
		ServiceCatalogConn:               servicecatalog.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.ServiceCatalog])})),
		ServiceDiscoveryConn:             servicediscovery.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.ServiceDiscovery])})),
		ServiceQuotasConn:                servicequotas.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.ServiceQuotas])})),
		SESConn:                          ses.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.SES])})),
		SESV2Conn:                        sesv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.SESV2])})),
		Session:                          sess,
		SFNConn:                          sfn.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.SFN])})),
		SignerConn:                       signer.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Signer])})),
		SimpleDBConn:                     simpledb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.SimpleDB])})),
		SMSConn:                          sms.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.SMS])})),
		SnowballConn:                     snowball.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Snowball])})),
		SNSConn:                          sns.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.SNS])})),
		SQSConn:                          sqs.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.SQS])})),
		SSMConn:                          ssm.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.SSM])})),
		SSMContactsConn:                  ssmcontacts.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.SSMContacts])})),
		SSMIncidentsConn:                 ssmincidents.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.SSMIncidents])})),
		SSOAdminConn:                     ssoadmin.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.SSOAdmin])})),
		SSOConn:                          sso.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.SSO])})),
		SSOOIDCConn:                      ssooidc.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.SSOOIDC])})),
		StorageGatewayConn:               storagegateway.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.StorageGateway])})),
		SupportConn:                      support.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Support])})),
		SWFConn:                          swf.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.SWF])})),
		SyntheticsConn:                   synthetics.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Synthetics])})),
		TerraformVersion:                 c.TerraformVersion,
		TextractConn:                     textract.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Textract])})),
		TimestreamQueryConn:              timestreamquery.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.TimestreamQuery])})),
		TimestreamWriteConn:              timestreamwrite.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.TimestreamWrite])})),
		TranscribeConn:                   transcribeservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Transcribe])})),
		TranscribeStreamingConn:          transcribestreamingservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.TranscribeStreaming])})),
		TransferConn:                     transfer.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Transfer])})),
		TranslateConn:                    translate.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.Translate])})),
		WAFConn:                          waf.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.WAF])})),
		WAFRegionalConn:                  wafregional.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.WAFRegional])})),
		WAFV2Conn:                        wafv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.WAFV2])})),
		WellArchitectedConn:              wellarchitected.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.WellArchitected])})),
		WorkDocsConn:                     workdocs.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.WorkDocs])})),
		WorkLinkConn:                     worklink.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.WorkLink])})),
		WorkMailConn:                     workmail.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.WorkMail])})),
		WorkMailMessageFlowConn:          workmailmessageflow.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.WorkMailMessageFlow])})),
		WorkSpacesConn:                   workspaces.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.WorkSpaces])})),
		XRayConn:                         xray.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[names.XRay])})),
	}

	// sts
	stsConfig := &aws.Config{
		Endpoint: aws.String(c.Endpoints[names.STS]),
	}

	if c.STSRegion != "" {
		stsConfig.Region = aws.String(c.STSRegion)
	}

	client.STSConn = sts.New(sess.Copy(stsConfig))

	// "Global" services that require customizations
	globalAcceleratorConfig := &aws.Config{
		Endpoint: aws.String(c.Endpoints[names.GlobalAccelerator]),
	}
	route53Config := &aws.Config{
		Endpoint: aws.String(c.Endpoints[names.Route53]),
	}
	route53RecoveryControlConfigConfig := &aws.Config{
		Endpoint: aws.String(c.Endpoints[names.Route53RecoveryControlConfig]),
	}
	route53RecoveryReadinessConfig := &aws.Config{
		Endpoint: aws.String(c.Endpoints[names.Route53RecoveryReadiness]),
	}
	shieldConfig := &aws.Config{
		Endpoint: aws.String(c.Endpoints[names.Shield]),
	}

	// Services that require multiple client configurations
	s3Config := &aws.Config{
		Endpoint:         aws.String(c.Endpoints[names.S3]),
		S3ForcePathStyle: aws.Bool(c.S3UsePathStyle),
	}

	client.S3Conn = s3.New(sess.Copy(s3Config))

	s3Config.DisableRestProtocolURICleaning = aws.Bool(true)
	client.S3ConnURICleaningDisabled = s3.New(sess.Copy(s3Config))

	// Force "global" services to correct regions
	switch partition {
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
	client.AppAutoScalingConn.Handlers.Retry.PushBack(func(r *request.Request) {
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

	client.ConfigServiceConn.Handlers.Retry.PushBack(func(r *request.Request) {
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
		switch err := r.Error; r.Operation.Name {
		case "AttachVpnGateway", "DetachVpnGateway":
			if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "This call cannot be completed because there are pending VPNs or Virtual Interfaces") {
				r.Retryable = aws.Bool(true)
			}

		case "CreateClientVpnEndpoint":
			if tfawserr.ErrMessageContains(err, "OperationNotPermitted", "Endpoint cannot be created while another endpoint is being created") {
				r.Retryable = aws.Bool(true)
			}

		case "CreateClientVpnRoute", "DeleteClientVpnRoute":
			if tfawserr.ErrMessageContains(err, "ConcurrentMutationLimitExceeded", "Cannot initiate another change for this endpoint at this time") {
				r.Retryable = aws.Bool(true)
			}

		case "CreateVpnConnection":
			if tfawserr.ErrMessageContains(err, "VpnConnectionLimitExceeded", "maximum number of mutating objects has been reached") {
				r.Retryable = aws.Bool(true)
			}

		case "CreateVpnGateway":
			if tfawserr.ErrMessageContains(err, "VpnGatewayLimitExceeded", "maximum number of mutating objects has been reached") {
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

	client.S3Conn.Handlers.Retry.PushBack(func(r *request.Request) {
		if tfawserr.ErrMessageContains(r.Error, "OperationAborted", "A conflicting conditional operation is currently in progress against this resource. Please try again.") {
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

func StdUserAgentProducts(terraformVersion string) *awsbase.APNInfo {
	return &awsbase.APNInfo{
		PartnerName: "HashiCorp",
		Products: []awsbase.UserAgentProduct{
			{Name: "Terraform", Version: terraformVersion, Comment: "+https://www.terraform.io"},
			{Name: "terraform-provider-aws", Version: version.ProviderVersion, Comment: "+https://registry.terraform.io/providers/hashicorp/aws"},
		},
	}
}

func NewSessionForRegion(cfg *aws.Config, region, terraformVersion string) (*session.Session, error) {
	session, err := session.NewSession(cfg)

	if err != nil {
		return nil, err
	}

	apnInfo := StdUserAgentProducts(terraformVersion)

	awsbasev1.SetSessionUserAgent(session, apnInfo, awsbase.UserAgentProducts{})

	return session.Copy(&aws.Config{Region: aws.String(region)}), nil
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
