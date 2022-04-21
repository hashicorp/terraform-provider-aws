package conns

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/route53domains"
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
	"github.com/aws/aws-sdk-go/service/cloudwatchevidently"
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
	"github.com/aws/aws-sdk-go/service/opensearchservice"
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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

type AWSClient struct {
	AccountID                 string
	DefaultTagsConfig         *tftags.DefaultConfig
	DNSSuffix                 string
	IgnoreTagsConfig          *tftags.IgnoreConfig
	MediaConvertAccountConn   *mediaconvert.MediaConvert
	Partition                 string
	Region                    string
	ReverseDNSPrefix          string
	S3ConnURICleaningDisabled *s3.S3
	Session                   *session.Session
	SupportedPlatforms        []string
	TerraformVersion          string

	AccessAnalyzerConn                *accessanalyzer.AccessAnalyzer
	AccountConn                       *account.Account
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
	CloudWatchEvidentlyConn           *cloudwatchevidently.CloudWatchEvidently
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
	DetectiveConn                     *detective.Detective
	DeviceFarmConn                    *devicefarm.DeviceFarm
	DevOpsGuruConn                    *devopsguru.DevOpsGuru
	DirectConnectConn                 *directconnect.DirectConnect
	DLMConn                           *dlm.DLM
	DMSConn                           *databasemigrationservice.DatabaseMigrationService
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
	OpenSearchConn                    *opensearchservice.OpenSearchService
	OpsWorksCMConn                    *opsworkscm.OpsWorksCM
	OpsWorksConn                      *opsworks.OpsWorks
	OrganizationsConn                 *organizations.Organizations
	OutpostsConn                      *outposts.Outposts
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
	RekognitionConn                   *rekognition.Rekognition
	ResourceGroupsConn                *resourcegroups.ResourceGroups
	ResourceGroupsTaggingAPIConn      *resourcegroupstaggingapi.ResourceGroupsTaggingAPI
	RoboMakerConn                     *robomaker.RoboMaker
	Route53Conn                       *route53.Route53
	Route53DomainsConn                *route53domains.Client
	Route53RecoveryControlConfigConn  *route53recoverycontrolconfig.Route53RecoveryControlConfig
	Route53RecoveryReadinessConn      *route53recoveryreadiness.Route53RecoveryReadiness
	Route53ResolverConn               *route53resolver.Route53Resolver
	S3Conn                            *s3.S3
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
	SWFConn                           *swf.SWF
	SyntheticsConn                    *synthetics.Synthetics
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
