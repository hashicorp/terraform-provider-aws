package conns

import (
	"fmt"
	"log"
	"strings"

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
	"github.com/aws/aws-sdk-go/service/route53domains"
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
	awsbase "github.com/hashicorp/aws-sdk-go-base"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/version"
)

const (
	AccessAnalyzer                = "accessanalyzer"
	Account                       = "account"
	ACM                           = "acm"
	ACMPCA                        = "acmpca"
	AlexaForBusiness              = "alexaforbusiness"
	AMP                           = "amp"
	Amplify                       = "amplify"
	AmplifyBackend                = "amplifybackend"
	APIGateway                    = "apigateway"
	APIGatewayV2                  = "apigatewayv2"
	AppAutoScaling                = "appautoscaling"
	AppConfig                     = "appconfig"
	AppFlow                       = "appflow"
	AppIntegrations               = "appintegrations"
	ApplicationCostProfiler       = "applicationcostprofiler"
	ApplicationDiscovery          = "applicationdiscovery"
	ApplicationInsights           = "applicationinsights"
	AppMesh                       = "appmesh"
	AppRegistry                   = "appregistry"
	AppRunner                     = "apprunner"
	AppStream                     = "appstream"
	AppSync                       = "appsync"
	Athena                        = "athena"
	AuditManager                  = "auditmanager"
	AugmentedAIRuntime            = "augmentedairuntime"
	AutoScaling                   = "autoscaling"
	AutoScalingPlans              = "autoscalingplans"
	Backup                        = "backup"
	Batch                         = "batch"
	Braket                        = "braket"
	Budgets                       = "budgets"
	Chime                         = "chime"
	Cloud9                        = "cloud9"
	CloudControl                  = "cloudcontrol"
	CloudDirectory                = "clouddirectory"
	CloudFormation                = "cloudformation"
	CloudFront                    = "cloudfront"
	CloudHSMV2                    = "cloudhsmv2"
	CloudSearch                   = "cloudsearch"
	CloudSearchDomain             = "cloudsearchdomain"
	CloudTrail                    = "cloudtrail"
	CloudWatch                    = "cloudwatch"
	CloudWatchLogs                = "cloudwatchlogs"
	CodeArtifact                  = "codeartifact"
	CodeBuild                     = "codebuild"
	CodeCommit                    = "codecommit"
	CodeDeploy                    = "codedeploy"
	CodeGuruProfiler              = "codeguruprofiler"
	CodeGuruReviewer              = "codegurureviewer"
	CodePipeline                  = "codepipeline"
	CodeStar                      = "codestar"
	CodeStarConnections           = "codestarconnections"
	CodeStarNotifications         = "codestarnotifications"
	CognitoIdentity               = "cognitoidentity"
	CognitoIDP                    = "cognitoidp"
	CognitoSync                   = "cognitosync"
	Comprehend                    = "comprehend"
	ComprehendMedical             = "comprehendmedical"
	ComputeOptimizer              = "computeoptimizer"
	ConfigService                 = "configservice"
	Connect                       = "connect"
	ConnectContactLens            = "connectcontactlens"
	ConnectParticipant            = "connectparticipant"
	CostExplorer                  = "costexplorer"
	CUR                           = "cur"
	CustomerProfiles              = "customerprofiles"
	DataExchange                  = "dataexchange"
	DataPipeline                  = "datapipeline"
	DataSync                      = "datasync"
	DAX                           = "dax"
	Detective                     = "detective"
	DeviceFarm                    = "devicefarm"
	DevOpsGuru                    = "devopsguru"
	DirectConnect                 = "directconnect"
	DLM                           = "dlm"
	DMS                           = "dms"
	DocDB                         = "docdb"
	DS                            = "ds"
	DynamoDB                      = "dynamodb"
	DynamoDBStreams               = "dynamodbstreams"
	EC2                           = "ec2"
	EC2InstanceConnect            = "ec2instanceconnect"
	ECR                           = "ecr"
	ECRPublic                     = "ecrpublic"
	ECS                           = "ecs"
	EFS                           = "efs"
	EKS                           = "eks"
	ElastiCache                   = "elasticache"
	ElasticBeanstalk              = "elasticbeanstalk"
	ElasticInference              = "elasticinference"
	Elasticsearch                 = "elasticsearch"
	ElasticTranscoder             = "elastictranscoder"
	ELB                           = "elb"
	ELBV2                         = "elbv2"
	EMR                           = "emr"
	EMRContainers                 = "emrcontainers"
	Events                        = "events"
	FinSpace                      = "finspace"
	FinSpaceData                  = "finspacedata"
	Firehose                      = "firehose"
	FIS                           = "fis"
	FMS                           = "fms"
	Forecast                      = "forecast"
	ForecastQuery                 = "forecastquery"
	FraudDetector                 = "frauddetector"
	FSx                           = "fsx"
	GameLift                      = "gamelift"
	Glacier                       = "glacier"
	GlobalAccelerator             = "globalaccelerator"
	Glue                          = "glue"
	GlueDataBrew                  = "gluedatabrew"
	Greengrass                    = "greengrass"
	GreengrassV2                  = "greengrassv2"
	GroundStation                 = "groundstation"
	GuardDuty                     = "guardduty"
	Health                        = "health"
	HealthLake                    = "healthlake"
	Honeycode                     = "honeycode"
	IAM                           = "iam"
	IdentityStore                 = "identitystore"
	ImageBuilder                  = "imagebuilder"
	Inspector                     = "inspector"
	IoT                           = "iot"
	IoT1ClickDevices              = "iot1clickdevices"
	IoT1ClickProjects             = "iot1clickprojects"
	IoTAnalytics                  = "iotanalytics"
	IoTDataPlane                  = "iotdataplane"
	IoTDeviceAdvisor              = "iotdeviceadvisor"
	IoTEvents                     = "iotevents"
	IoTEventsData                 = "ioteventsdata"
	IoTFleetHub                   = "iotfleethub"
	IoTJobsDataPlane              = "iotjobsdataplane"
	IoTSecureTunneling            = "iotsecuretunneling"
	IoTSiteWise                   = "iotsitewise"
	IoTThingsGraph                = "iotthingsgraph"
	IoTWireless                   = "iotwireless"
	IVS                           = "ivs"
	Kafka                         = "kafka"
	KafkaConnect                  = "kafkaconnect"
	Kendra                        = "kendra"
	Kinesis                       = "kinesis"
	KinesisAnalytics              = "kinesisanalytics"
	KinesisAnalyticsV2            = "kinesisanalyticsv2"
	KinesisVideo                  = "kinesisvideo"
	KinesisVideoArchivedMedia     = "kinesisvideoarchivedmedia"
	KinesisVideoMedia             = "kinesisvideomedia"
	KinesisVideoSignalingChannels = "kinesisvideosignalingchannels"
	KMS                           = "kms"
	LakeFormation                 = "lakeformation"
	Lambda                        = "lambda"
	LexModels                     = "lexmodels"
	LexModelsV2                   = "lexmodelsv2"
	LexRuntime                    = "lexruntime"
	LexRuntimeV2                  = "lexruntimev2"
	LicenseManager                = "licensemanager"
	Lightsail                     = "lightsail"
	Location                      = "location"
	LookoutEquipment              = "lookoutequipment"
	LookoutForVision              = "lookoutforvision"
	LookoutMetrics                = "lookoutmetrics"
	MachineLearning               = "machinelearning"
	Macie                         = "macie"
	Macie2                        = "macie2"
	ManagedBlockchain             = "managedblockchain"
	MarketplaceCatalog            = "marketplacecatalog"
	MarketplaceCommerceAnalytics  = "marketplacecommerceanalytics"
	MarketplaceEntitlement        = "marketplaceentitlement"
	MarketplaceMetering           = "marketplacemetering"
	MediaConnect                  = "mediaconnect"
	MediaConvert                  = "mediaconvert"
	MediaLive                     = "medialive"
	MediaPackage                  = "mediapackage"
	MediaPackageVOD               = "mediapackagevod"
	MediaStore                    = "mediastore"
	MediaStoreData                = "mediastoredata"
	MediaTailor                   = "mediatailor"
	MemoryDB                      = "memorydb"
	Mgn                           = "mgn"
	MigrationHub                  = "migrationhub"
	MigrationHubConfig            = "migrationhubconfig"
	Mobile                        = "mobile"
	MobileAnalytics               = "mobileanalytics"
	MQ                            = "mq"
	MTurk                         = "mturk"
	MWAA                          = "mwaa"
	Neptune                       = "neptune"
	NetworkFirewall               = "networkfirewall"
	NetworkManager                = "networkmanager"
	NimbleStudio                  = "nimblestudio"
	OpsWorks                      = "opsworks"
	OpsWorksCM                    = "opsworkscm"
	Organizations                 = "organizations"
	Outposts                      = "outposts"
	Personalize                   = "personalize"
	PersonalizeEvents             = "personalizeevents"
	PersonalizeRuntime            = "personalizeruntime"
	PI                            = "pi"
	Pinpoint                      = "pinpoint"
	PinpointEmail                 = "pinpointemail"
	PinpointSMSVoice              = "pinpointsmsvoice"
	Polly                         = "polly"
	Pricing                       = "pricing"
	Proton                        = "proton"
	QLDB                          = "qldb"
	QLDBSession                   = "qldbsession"
	QuickSight                    = "quicksight"
	RAM                           = "ram"
	RDS                           = "rds"
	RDSData                       = "rdsdata"
	Redshift                      = "redshift"
	RedshiftData                  = "redshiftdata"
	Rekognition                   = "rekognition"
	ResourceGroups                = "resourcegroups"
	ResourceGroupsTaggingAPI      = "resourcegroupstaggingapi"
	RoboMaker                     = "robomaker"
	Route53                       = "route53"
	Route53Domains                = "route53domains"
	Route53RecoveryControlConfig  = "route53recoverycontrolconfig"
	Route53RecoveryReadiness      = "route53recoveryreadiness"
	Route53Resolver               = "route53resolver"
	S3                            = "s3"
	S3Control                     = "s3control"
	S3Outposts                    = "s3outposts"
	SageMaker                     = "sagemaker"
	SageMakerEdgeManager          = "sagemakeredgemanager"
	SageMakerFeatureStoreRuntime  = "sagemakerfeaturestoreruntime"
	SageMakerRuntime              = "sagemakerruntime"
	SavingsPlans                  = "savingsplans"
	Schemas                       = "schemas"
	SecretsManager                = "secretsmanager"
	SecurityHub                   = "securityhub"
	ServerlessRepo                = "serverlessrepo"
	ServiceCatalog                = "servicecatalog"
	ServiceDiscovery              = "servicediscovery"
	ServiceQuotas                 = "servicequotas"
	SES                           = "ses"
	SESV2                         = "sesv2"
	SFN                           = "sfn"
	Shield                        = "shield"
	Signer                        = "signer"
	SimpleDB                      = "simpledb"
	SMS                           = "sms"
	Snowball                      = "snowball"
	SNS                           = "sns"
	SQS                           = "sqs"
	SSM                           = "ssm"
	SSMContacts                   = "ssmcontacts"
	SSMIncidents                  = "ssmincidents"
	SSO                           = "sso"
	SSOAdmin                      = "ssoadmin"
	SSOOIDC                       = "ssooidc"
	StorageGateway                = "storagegateway"
	STS                           = "sts"
	Support                       = "support"
	SWF                           = "swf"
	Synthetics                    = "synthetics"
	Textract                      = "textract"
	TimestreamQuery               = "timestreamquery"
	TimestreamWrite               = "timestreamwrite"
	Transcribe                    = "transcribe"
	TranscribeStreaming           = "transcribestreaming"
	Transfer                      = "transfer"
	Translate                     = "translate"
	WAF                           = "waf"
	WAFRegional                   = "wafregional"
	WAFV2                         = "wafv2"
	WellArchitected               = "wellarchitected"
	WorkDocs                      = "workdocs"
	WorkLink                      = "worklink"
	WorkMail                      = "workmail"
	WorkMailMessageFlow           = "workmailmessageflow"
	WorkSpaces                    = "workspaces"
	XRay                          = "xray"
)

type ServiceDatum struct {
	AWSClientName     string
	AWSServiceName    string
	AWSEndpointsID    string
	AWSServiceID      string
	ProviderNameUpper string
	HCLKeys           []string
}

var serviceData map[string]*ServiceDatum

func init() {
	serviceData = make(map[string]*ServiceDatum)

	serviceData[AccessAnalyzer] = &ServiceDatum{AWSClientName: "AccessAnalyzer", AWSServiceName: accessanalyzer.ServiceName, AWSEndpointsID: accessanalyzer.EndpointsID, AWSServiceID: accessanalyzer.ServiceID, ProviderNameUpper: "AccessAnalyzer", HCLKeys: []string{"accessanalyzer"}}
	serviceData[Account] = &ServiceDatum{AWSClientName: "Account", AWSServiceName: account.ServiceName, AWSEndpointsID: account.EndpointsID, AWSServiceID: account.ServiceID, ProviderNameUpper: "Account", HCLKeys: []string{"account"}}
	serviceData[ACM] = &ServiceDatum{AWSClientName: "ACM", AWSServiceName: acm.ServiceName, AWSEndpointsID: acm.EndpointsID, AWSServiceID: acm.ServiceID, ProviderNameUpper: "ACM", HCLKeys: []string{"acm"}}
	serviceData[ACMPCA] = &ServiceDatum{AWSClientName: "ACMPCA", AWSServiceName: acmpca.ServiceName, AWSEndpointsID: acmpca.EndpointsID, AWSServiceID: acmpca.ServiceID, ProviderNameUpper: "ACMPCA", HCLKeys: []string{"acmpca"}}
	serviceData[AlexaForBusiness] = &ServiceDatum{AWSClientName: "AlexaForBusiness", AWSServiceName: alexaforbusiness.ServiceName, AWSEndpointsID: alexaforbusiness.EndpointsID, AWSServiceID: alexaforbusiness.ServiceID, ProviderNameUpper: "AlexaForBusiness", HCLKeys: []string{"alexaforbusiness"}}
	serviceData[AMP] = &ServiceDatum{AWSClientName: "PrometheusService", AWSServiceName: prometheusservice.ServiceName, AWSEndpointsID: prometheusservice.EndpointsID, AWSServiceID: prometheusservice.ServiceID, ProviderNameUpper: "AMP", HCLKeys: []string{"amp", "prometheus", "prometheusservice"}}
	serviceData[Amplify] = &ServiceDatum{AWSClientName: "Amplify", AWSServiceName: amplify.ServiceName, AWSEndpointsID: amplify.EndpointsID, AWSServiceID: amplify.ServiceID, ProviderNameUpper: "Amplify", HCLKeys: []string{"amplify"}}
	serviceData[AmplifyBackend] = &ServiceDatum{AWSClientName: "AmplifyBackend", AWSServiceName: amplifybackend.ServiceName, AWSEndpointsID: amplifybackend.EndpointsID, AWSServiceID: amplifybackend.ServiceID, ProviderNameUpper: "AmplifyBackend", HCLKeys: []string{"amplifybackend"}}
	serviceData[APIGateway] = &ServiceDatum{AWSClientName: "APIGateway", AWSServiceName: apigateway.ServiceName, AWSEndpointsID: apigateway.EndpointsID, AWSServiceID: apigateway.ServiceID, ProviderNameUpper: "APIGateway", HCLKeys: []string{"apigateway"}}
	serviceData[APIGatewayV2] = &ServiceDatum{AWSClientName: "APIGatewayV2", AWSServiceName: apigatewayv2.ServiceName, AWSEndpointsID: apigatewayv2.EndpointsID, AWSServiceID: apigatewayv2.ServiceID, ProviderNameUpper: "APIGatewayV2", HCLKeys: []string{"apigatewayv2"}}
	serviceData[AppAutoScaling] = &ServiceDatum{AWSClientName: "ApplicationAutoScaling", AWSServiceName: applicationautoscaling.ServiceName, AWSEndpointsID: applicationautoscaling.EndpointsID, AWSServiceID: applicationautoscaling.ServiceID, ProviderNameUpper: "AppAutoScaling", HCLKeys: []string{"appautoscaling", "applicationautoscaling"}}
	serviceData[AppConfig] = &ServiceDatum{AWSClientName: "AppConfig", AWSServiceName: appconfig.ServiceName, AWSEndpointsID: appconfig.EndpointsID, AWSServiceID: appconfig.ServiceID, ProviderNameUpper: "AppConfig", HCLKeys: []string{"appconfig"}}
	serviceData[AppFlow] = &ServiceDatum{AWSClientName: "Appflow", AWSServiceName: appflow.ServiceName, AWSEndpointsID: appflow.EndpointsID, AWSServiceID: appflow.ServiceID, ProviderNameUpper: "AppFlow", HCLKeys: []string{"appflow"}}
	serviceData[AppIntegrations] = &ServiceDatum{AWSClientName: "AppIntegrationsService", AWSServiceName: appintegrationsservice.ServiceName, AWSEndpointsID: appintegrationsservice.EndpointsID, AWSServiceID: appintegrationsservice.ServiceID, ProviderNameUpper: "AppIntegrations", HCLKeys: []string{"appintegrations", "appintegrationsservice"}}
	serviceData[ApplicationCostProfiler] = &ServiceDatum{AWSClientName: "ApplicationCostProfiler", AWSServiceName: applicationcostprofiler.ServiceName, AWSEndpointsID: applicationcostprofiler.EndpointsID, AWSServiceID: applicationcostprofiler.ServiceID, ProviderNameUpper: "ApplicationCostProfiler", HCLKeys: []string{"applicationcostprofiler"}}
	serviceData[ApplicationDiscovery] = &ServiceDatum{AWSClientName: "ApplicationDiscoveryService", AWSServiceName: applicationdiscoveryservice.ServiceName, AWSEndpointsID: applicationdiscoveryservice.EndpointsID, AWSServiceID: applicationdiscoveryservice.ServiceID, ProviderNameUpper: "ApplicationDiscovery", HCLKeys: []string{"applicationdiscovery", "applicationdiscoveryservice"}}
	serviceData[ApplicationInsights] = &ServiceDatum{AWSClientName: "ApplicationInsights", AWSServiceName: applicationinsights.ServiceName, AWSEndpointsID: applicationinsights.EndpointsID, AWSServiceID: applicationinsights.ServiceID, ProviderNameUpper: "ApplicationInsights", HCLKeys: []string{"applicationinsights"}}
	serviceData[AppMesh] = &ServiceDatum{AWSClientName: "AppMesh", AWSServiceName: appmesh.ServiceName, AWSEndpointsID: appmesh.EndpointsID, AWSServiceID: appmesh.ServiceID, ProviderNameUpper: "AppMesh", HCLKeys: []string{"appmesh"}}
	serviceData[AppRegistry] = &ServiceDatum{AWSClientName: "AppRegistry", AWSServiceName: appregistry.ServiceName, AWSEndpointsID: appregistry.EndpointsID, AWSServiceID: appregistry.ServiceID, ProviderNameUpper: "AppRegistry", HCLKeys: []string{"appregistry"}}
	serviceData[AppRunner] = &ServiceDatum{AWSClientName: "AppRunner", AWSServiceName: apprunner.ServiceName, AWSEndpointsID: apprunner.EndpointsID, AWSServiceID: apprunner.ServiceID, ProviderNameUpper: "AppRunner", HCLKeys: []string{"apprunner"}}
	serviceData[AppStream] = &ServiceDatum{AWSClientName: "AppStream", AWSServiceName: appstream.ServiceName, AWSEndpointsID: appstream.EndpointsID, AWSServiceID: appstream.ServiceID, ProviderNameUpper: "AppStream", HCLKeys: []string{"appstream"}}
	serviceData[AppSync] = &ServiceDatum{AWSClientName: "AppSync", AWSServiceName: appsync.ServiceName, AWSEndpointsID: appsync.EndpointsID, AWSServiceID: appsync.ServiceID, ProviderNameUpper: "AppSync", HCLKeys: []string{"appsync"}}
	serviceData[Athena] = &ServiceDatum{AWSClientName: "Athena", AWSServiceName: athena.ServiceName, AWSEndpointsID: athena.EndpointsID, AWSServiceID: athena.ServiceID, ProviderNameUpper: "Athena", HCLKeys: []string{"athena"}}
	serviceData[AuditManager] = &ServiceDatum{AWSClientName: "AuditManager", AWSServiceName: auditmanager.ServiceName, AWSEndpointsID: auditmanager.EndpointsID, AWSServiceID: auditmanager.ServiceID, ProviderNameUpper: "AuditManager", HCLKeys: []string{"auditmanager"}}
	serviceData[AugmentedAIRuntime] = &ServiceDatum{AWSClientName: "AugmentedAIRuntime", AWSServiceName: augmentedairuntime.ServiceName, AWSEndpointsID: augmentedairuntime.EndpointsID, AWSServiceID: augmentedairuntime.ServiceID, ProviderNameUpper: "AugmentedAIRuntime", HCLKeys: []string{"augmentedairuntime"}}
	serviceData[AutoScaling] = &ServiceDatum{AWSClientName: "AutoScaling", AWSServiceName: autoscaling.ServiceName, AWSEndpointsID: autoscaling.EndpointsID, AWSServiceID: autoscaling.ServiceID, ProviderNameUpper: "AutoScaling", HCLKeys: []string{"autoscaling"}}
	serviceData[AutoScalingPlans] = &ServiceDatum{AWSClientName: "AutoScalingPlans", AWSServiceName: autoscalingplans.ServiceName, AWSEndpointsID: autoscalingplans.EndpointsID, AWSServiceID: autoscalingplans.ServiceID, ProviderNameUpper: "AutoScalingPlans", HCLKeys: []string{"autoscalingplans"}}
	serviceData[Backup] = &ServiceDatum{AWSClientName: "Backup", AWSServiceName: backup.ServiceName, AWSEndpointsID: backup.EndpointsID, AWSServiceID: backup.ServiceID, ProviderNameUpper: "Backup", HCLKeys: []string{"backup"}}
	serviceData[Batch] = &ServiceDatum{AWSClientName: "Batch", AWSServiceName: batch.ServiceName, AWSEndpointsID: batch.EndpointsID, AWSServiceID: batch.ServiceID, ProviderNameUpper: "Batch", HCLKeys: []string{"batch"}}
	serviceData[Braket] = &ServiceDatum{AWSClientName: "Braket", AWSServiceName: braket.ServiceName, AWSEndpointsID: braket.EndpointsID, AWSServiceID: braket.ServiceID, ProviderNameUpper: "Braket", HCLKeys: []string{"braket"}}
	serviceData[Budgets] = &ServiceDatum{AWSClientName: "Budgets", AWSServiceName: budgets.ServiceName, AWSEndpointsID: budgets.EndpointsID, AWSServiceID: budgets.ServiceID, ProviderNameUpper: "Budgets", HCLKeys: []string{"budgets"}}
	serviceData[Chime] = &ServiceDatum{AWSClientName: "Chime", AWSServiceName: chime.ServiceName, AWSEndpointsID: chime.EndpointsID, AWSServiceID: chime.ServiceID, ProviderNameUpper: "Chime", HCLKeys: []string{"chime"}}
	serviceData[Cloud9] = &ServiceDatum{AWSClientName: "Cloud9", AWSServiceName: cloud9.ServiceName, AWSEndpointsID: cloud9.EndpointsID, AWSServiceID: cloud9.ServiceID, ProviderNameUpper: "Cloud9", HCLKeys: []string{"cloud9"}}
	serviceData[CloudControl] = &ServiceDatum{AWSClientName: "CloudControlApi", AWSServiceName: cloudcontrolapi.ServiceName, AWSEndpointsID: cloudcontrolapi.EndpointsID, AWSServiceID: cloudcontrolapi.ServiceID, ProviderNameUpper: "CloudControl", HCLKeys: []string{"cloudcontrolapi", "cloudcontrol"}}
	serviceData[CloudDirectory] = &ServiceDatum{AWSClientName: "CloudDirectory", AWSServiceName: clouddirectory.ServiceName, AWSEndpointsID: clouddirectory.EndpointsID, AWSServiceID: clouddirectory.ServiceID, ProviderNameUpper: "CloudDirectory", HCLKeys: []string{"clouddirectory"}}
	serviceData[CloudFormation] = &ServiceDatum{AWSClientName: "CloudFormation", AWSServiceName: cloudformation.ServiceName, AWSEndpointsID: cloudformation.EndpointsID, AWSServiceID: cloudformation.ServiceID, ProviderNameUpper: "CloudFormation", HCLKeys: []string{"cloudformation"}}
	serviceData[CloudFront] = &ServiceDatum{AWSClientName: "CloudFront", AWSServiceName: cloudfront.ServiceName, AWSEndpointsID: cloudfront.EndpointsID, AWSServiceID: cloudfront.ServiceID, ProviderNameUpper: "CloudFront", HCLKeys: []string{"cloudfront"}}
	serviceData[CloudHSMV2] = &ServiceDatum{AWSClientName: "CloudHSMV2", AWSServiceName: cloudhsmv2.ServiceName, AWSEndpointsID: cloudhsmv2.EndpointsID, AWSServiceID: cloudhsmv2.ServiceID, ProviderNameUpper: "CloudHSMV2", HCLKeys: []string{"cloudhsm", "cloudhsmv2"}}
	serviceData[CloudSearch] = &ServiceDatum{AWSClientName: "CloudSearch", AWSServiceName: cloudsearch.ServiceName, AWSEndpointsID: cloudsearch.EndpointsID, AWSServiceID: cloudsearch.ServiceID, ProviderNameUpper: "CloudSearch", HCLKeys: []string{"cloudsearch"}}
	serviceData[CloudSearchDomain] = &ServiceDatum{AWSClientName: "CloudSearchDomain", AWSServiceName: cloudsearchdomain.ServiceName, AWSEndpointsID: cloudsearchdomain.EndpointsID, AWSServiceID: cloudsearchdomain.ServiceID, ProviderNameUpper: "CloudSearchDomain", HCLKeys: []string{"cloudsearchdomain"}}
	serviceData[CloudTrail] = &ServiceDatum{AWSClientName: "CloudTrail", AWSServiceName: cloudtrail.ServiceName, AWSEndpointsID: cloudtrail.EndpointsID, AWSServiceID: cloudtrail.ServiceID, ProviderNameUpper: "CloudTrail", HCLKeys: []string{"cloudtrail"}}
	serviceData[CloudWatch] = &ServiceDatum{AWSClientName: "CloudWatch", AWSServiceName: cloudwatch.ServiceName, AWSEndpointsID: cloudwatch.EndpointsID, AWSServiceID: cloudwatch.ServiceID, ProviderNameUpper: "CloudWatch", HCLKeys: []string{"cloudwatch"}}
	serviceData[CloudWatchLogs] = &ServiceDatum{AWSClientName: "CloudWatchLogs", AWSServiceName: cloudwatchlogs.ServiceName, AWSEndpointsID: cloudwatchlogs.EndpointsID, AWSServiceID: cloudwatchlogs.ServiceID, ProviderNameUpper: "CloudWatchLogs", HCLKeys: []string{"cloudwatchlogs"}}
	serviceData[CodeArtifact] = &ServiceDatum{AWSClientName: "CodeArtifact", AWSServiceName: codeartifact.ServiceName, AWSEndpointsID: codeartifact.EndpointsID, AWSServiceID: codeartifact.ServiceID, ProviderNameUpper: "CodeArtifact", HCLKeys: []string{"codeartifact"}}
	serviceData[CodeBuild] = &ServiceDatum{AWSClientName: "CodeBuild", AWSServiceName: codebuild.ServiceName, AWSEndpointsID: codebuild.EndpointsID, AWSServiceID: codebuild.ServiceID, ProviderNameUpper: "CodeBuild", HCLKeys: []string{"codebuild"}}
	serviceData[CodeCommit] = &ServiceDatum{AWSClientName: "CodeCommit", AWSServiceName: codecommit.ServiceName, AWSEndpointsID: codecommit.EndpointsID, AWSServiceID: codecommit.ServiceID, ProviderNameUpper: "CodeCommit", HCLKeys: []string{"codecommit"}}
	serviceData[CodeDeploy] = &ServiceDatum{AWSClientName: "CodeDeploy", AWSServiceName: codedeploy.ServiceName, AWSEndpointsID: codedeploy.EndpointsID, AWSServiceID: codedeploy.ServiceID, ProviderNameUpper: "CodeDeploy", HCLKeys: []string{"codedeploy"}}
	serviceData[CodeGuruProfiler] = &ServiceDatum{AWSClientName: "CodeGuruProfiler", AWSServiceName: codeguruprofiler.ServiceName, AWSEndpointsID: codeguruprofiler.EndpointsID, AWSServiceID: codeguruprofiler.ServiceID, ProviderNameUpper: "CodeGuruProfiler", HCLKeys: []string{"codeguruprofiler"}}
	serviceData[CodeGuruReviewer] = &ServiceDatum{AWSClientName: "CodeGuruReviewer", AWSServiceName: codegurureviewer.ServiceName, AWSEndpointsID: codegurureviewer.EndpointsID, AWSServiceID: codegurureviewer.ServiceID, ProviderNameUpper: "CodeGuruReviewer", HCLKeys: []string{"codegurureviewer"}}
	serviceData[CodePipeline] = &ServiceDatum{AWSClientName: "CodePipeline", AWSServiceName: codepipeline.ServiceName, AWSEndpointsID: codepipeline.EndpointsID, AWSServiceID: codepipeline.ServiceID, ProviderNameUpper: "CodePipeline", HCLKeys: []string{"codepipeline"}}
	serviceData[CodeStar] = &ServiceDatum{AWSClientName: "CodeStar", AWSServiceName: codestar.ServiceName, AWSEndpointsID: codestar.EndpointsID, AWSServiceID: codestar.ServiceID, ProviderNameUpper: "CodeStar", HCLKeys: []string{"codestar"}}
	serviceData[CodeStarConnections] = &ServiceDatum{AWSClientName: "CodeStarConnections", AWSServiceName: codestarconnections.ServiceName, AWSEndpointsID: codestarconnections.EndpointsID, AWSServiceID: codestarconnections.ServiceID, ProviderNameUpper: "CodeStarConnections", HCLKeys: []string{"codestarconnections"}}
	serviceData[CodeStarNotifications] = &ServiceDatum{AWSClientName: "CodeStarNotifications", AWSServiceName: codestarnotifications.ServiceName, AWSEndpointsID: codestarnotifications.EndpointsID, AWSServiceID: codestarnotifications.ServiceID, ProviderNameUpper: "CodeStarNotifications", HCLKeys: []string{"codestarnotifications"}}
	serviceData[CognitoIdentity] = &ServiceDatum{AWSClientName: "CognitoIdentity", AWSServiceName: cognitoidentity.ServiceName, AWSEndpointsID: cognitoidentity.EndpointsID, AWSServiceID: cognitoidentity.ServiceID, ProviderNameUpper: "CognitoIdentity", HCLKeys: []string{"cognitoidentity"}}
	serviceData[CognitoIDP] = &ServiceDatum{AWSClientName: "CognitoIdentityProvider", AWSServiceName: cognitoidentityprovider.ServiceName, AWSEndpointsID: cognitoidentityprovider.EndpointsID, AWSServiceID: cognitoidentityprovider.ServiceID, ProviderNameUpper: "CognitoIDP", HCLKeys: []string{"cognitoidp", "cognitoidentityprovider"}}
	serviceData[CognitoSync] = &ServiceDatum{AWSClientName: "CognitoSync", AWSServiceName: cognitosync.ServiceName, AWSEndpointsID: cognitosync.EndpointsID, AWSServiceID: cognitosync.ServiceID, ProviderNameUpper: "CognitoSync", HCLKeys: []string{"cognitosync"}}
	serviceData[Comprehend] = &ServiceDatum{AWSClientName: "Comprehend", AWSServiceName: comprehend.ServiceName, AWSEndpointsID: comprehend.EndpointsID, AWSServiceID: comprehend.ServiceID, ProviderNameUpper: "Comprehend", HCLKeys: []string{"comprehend"}}
	serviceData[ComprehendMedical] = &ServiceDatum{AWSClientName: "ComprehendMedical", AWSServiceName: comprehendmedical.ServiceName, AWSEndpointsID: comprehendmedical.EndpointsID, AWSServiceID: comprehendmedical.ServiceID, ProviderNameUpper: "ComprehendMedical", HCLKeys: []string{"comprehendmedical"}}
	serviceData[ConfigService] = &ServiceDatum{AWSClientName: "ConfigService", AWSServiceName: configservice.ServiceName, AWSEndpointsID: configservice.EndpointsID, AWSServiceID: configservice.ServiceID, ProviderNameUpper: "ConfigService", HCLKeys: []string{"configservice", "config"}}
	serviceData[Connect] = &ServiceDatum{AWSClientName: "Connect", AWSServiceName: connect.ServiceName, AWSEndpointsID: connect.EndpointsID, AWSServiceID: connect.ServiceID, ProviderNameUpper: "Connect", HCLKeys: []string{"connect"}}
	serviceData[ConnectContactLens] = &ServiceDatum{AWSClientName: "ConnectContactLens", AWSServiceName: connectcontactlens.ServiceName, AWSEndpointsID: connectcontactlens.EndpointsID, AWSServiceID: connectcontactlens.ServiceID, ProviderNameUpper: "ConnectContactLens", HCLKeys: []string{"connectcontactlens"}}
	serviceData[ConnectParticipant] = &ServiceDatum{AWSClientName: "ConnectParticipant", AWSServiceName: connectparticipant.ServiceName, AWSEndpointsID: connectparticipant.EndpointsID, AWSServiceID: connectparticipant.ServiceID, ProviderNameUpper: "ConnectParticipant", HCLKeys: []string{"connectparticipant"}}
	serviceData[CostExplorer] = &ServiceDatum{AWSClientName: "CostExplorer", AWSServiceName: costexplorer.ServiceName, AWSEndpointsID: costexplorer.EndpointsID, AWSServiceID: costexplorer.ServiceID, ProviderNameUpper: "CostExplorer", HCLKeys: []string{"costexplorer"}}
	serviceData[CUR] = &ServiceDatum{AWSClientName: "CostandUsageReportService", AWSServiceName: costandusagereportservice.ServiceName, AWSEndpointsID: costandusagereportservice.EndpointsID, AWSServiceID: costandusagereportservice.ServiceID, ProviderNameUpper: "CUR", HCLKeys: []string{"cur", "costandusagereportservice"}}
	serviceData[DataExchange] = &ServiceDatum{AWSClientName: "DataExchange", AWSServiceName: dataexchange.ServiceName, AWSEndpointsID: dataexchange.EndpointsID, AWSServiceID: dataexchange.ServiceID, ProviderNameUpper: "DataExchange", HCLKeys: []string{"dataexchange"}}
	serviceData[DataPipeline] = &ServiceDatum{AWSClientName: "DataPipeline", AWSServiceName: datapipeline.ServiceName, AWSEndpointsID: datapipeline.EndpointsID, AWSServiceID: datapipeline.ServiceID, ProviderNameUpper: "DataPipeline", HCLKeys: []string{"datapipeline"}}
	serviceData[DataSync] = &ServiceDatum{AWSClientName: "DataSync", AWSServiceName: datasync.ServiceName, AWSEndpointsID: datasync.EndpointsID, AWSServiceID: datasync.ServiceID, ProviderNameUpper: "DataSync", HCLKeys: []string{"datasync"}}
	serviceData[DAX] = &ServiceDatum{AWSClientName: "DAX", AWSServiceName: dax.ServiceName, AWSEndpointsID: dax.EndpointsID, AWSServiceID: dax.ServiceID, ProviderNameUpper: "DAX", HCLKeys: []string{"dax"}}
	serviceData[Detective] = &ServiceDatum{AWSClientName: "Detective", AWSServiceName: detective.ServiceName, AWSEndpointsID: detective.EndpointsID, AWSServiceID: detective.ServiceID, ProviderNameUpper: "Detective", HCLKeys: []string{"detective"}}
	serviceData[DeviceFarm] = &ServiceDatum{AWSClientName: "DeviceFarm", AWSServiceName: devicefarm.ServiceName, AWSEndpointsID: devicefarm.EndpointsID, AWSServiceID: devicefarm.ServiceID, ProviderNameUpper: "DeviceFarm", HCLKeys: []string{"devicefarm"}}
	serviceData[DevOpsGuru] = &ServiceDatum{AWSClientName: "DevOpsGuru", AWSServiceName: devopsguru.ServiceName, AWSEndpointsID: devopsguru.EndpointsID, AWSServiceID: devopsguru.ServiceID, ProviderNameUpper: "DevOpsGuru", HCLKeys: []string{"devopsguru"}}
	serviceData[DirectConnect] = &ServiceDatum{AWSClientName: "DirectConnect", AWSServiceName: directconnect.ServiceName, AWSEndpointsID: directconnect.EndpointsID, AWSServiceID: directconnect.ServiceID, ProviderNameUpper: "DirectConnect", HCLKeys: []string{"directconnect"}}
	serviceData[DLM] = &ServiceDatum{AWSClientName: "DLM", AWSServiceName: dlm.ServiceName, AWSEndpointsID: dlm.EndpointsID, AWSServiceID: dlm.ServiceID, ProviderNameUpper: "DLM", HCLKeys: []string{"dlm"}}
	serviceData[DMS] = &ServiceDatum{AWSClientName: "DatabaseMigrationService", AWSServiceName: databasemigrationservice.ServiceName, AWSEndpointsID: databasemigrationservice.EndpointsID, AWSServiceID: databasemigrationservice.ServiceID, ProviderNameUpper: "DMS", HCLKeys: []string{"dms", "databasemigration", "databasemigrationservice"}}
	serviceData[DocDB] = &ServiceDatum{AWSClientName: "DocDB", AWSServiceName: docdb.ServiceName, AWSEndpointsID: docdb.EndpointsID, AWSServiceID: docdb.ServiceID, ProviderNameUpper: "DocDB", HCLKeys: []string{"docdb"}}
	serviceData[DS] = &ServiceDatum{AWSClientName: "DirectoryService", AWSServiceName: directoryservice.ServiceName, AWSEndpointsID: directoryservice.EndpointsID, AWSServiceID: directoryservice.ServiceID, ProviderNameUpper: "DS", HCLKeys: []string{"ds"}}
	serviceData[DynamoDB] = &ServiceDatum{AWSClientName: "DynamoDB", AWSServiceName: dynamodb.ServiceName, AWSEndpointsID: dynamodb.EndpointsID, AWSServiceID: dynamodb.ServiceID, ProviderNameUpper: "DynamoDB", HCLKeys: []string{"dynamodb"}}
	serviceData[DynamoDBStreams] = &ServiceDatum{AWSClientName: "DynamoDBStreams", AWSServiceName: dynamodbstreams.ServiceName, AWSEndpointsID: dynamodbstreams.EndpointsID, AWSServiceID: dynamodbstreams.ServiceID, ProviderNameUpper: "DynamoDBStreams", HCLKeys: []string{"dynamodbstreams"}}
	serviceData[EC2] = &ServiceDatum{AWSClientName: "EC2", AWSServiceName: ec2.ServiceName, AWSEndpointsID: ec2.EndpointsID, AWSServiceID: ec2.ServiceID, ProviderNameUpper: "EC2", HCLKeys: []string{"ec2"}}
	serviceData[EC2InstanceConnect] = &ServiceDatum{AWSClientName: "EC2InstanceConnect", AWSServiceName: ec2instanceconnect.ServiceName, AWSEndpointsID: ec2instanceconnect.EndpointsID, AWSServiceID: ec2instanceconnect.ServiceID, ProviderNameUpper: "EC2InstanceConnect", HCLKeys: []string{"ec2instanceconnect"}}
	serviceData[ECR] = &ServiceDatum{AWSClientName: "ECR", AWSServiceName: ecr.ServiceName, AWSEndpointsID: ecr.EndpointsID, AWSServiceID: ecr.ServiceID, ProviderNameUpper: "ECR", HCLKeys: []string{"ecr"}}
	serviceData[ECRPublic] = &ServiceDatum{AWSClientName: "ECRPublic", AWSServiceName: ecrpublic.ServiceName, AWSEndpointsID: ecrpublic.EndpointsID, AWSServiceID: ecrpublic.ServiceID, ProviderNameUpper: "ECRPublic", HCLKeys: []string{"ecrpublic"}}
	serviceData[ECS] = &ServiceDatum{AWSClientName: "ECS", AWSServiceName: ecs.ServiceName, AWSEndpointsID: ecs.EndpointsID, AWSServiceID: ecs.ServiceID, ProviderNameUpper: "ECS", HCLKeys: []string{"ecs"}}
	serviceData[EFS] = &ServiceDatum{AWSClientName: "EFS", AWSServiceName: efs.ServiceName, AWSEndpointsID: efs.EndpointsID, AWSServiceID: efs.ServiceID, ProviderNameUpper: "EFS", HCLKeys: []string{"efs"}}
	serviceData[EKS] = &ServiceDatum{AWSClientName: "EKS", AWSServiceName: eks.ServiceName, AWSEndpointsID: eks.EndpointsID, AWSServiceID: eks.ServiceID, ProviderNameUpper: "EKS", HCLKeys: []string{"eks"}}
	serviceData[ElastiCache] = &ServiceDatum{AWSClientName: "ElastiCache", AWSServiceName: elasticache.ServiceName, AWSEndpointsID: elasticache.EndpointsID, AWSServiceID: elasticache.ServiceID, ProviderNameUpper: "ElastiCache", HCLKeys: []string{"elasticache"}}
	serviceData[ElasticBeanstalk] = &ServiceDatum{AWSClientName: "ElasticBeanstalk", AWSServiceName: elasticbeanstalk.ServiceName, AWSEndpointsID: elasticbeanstalk.EndpointsID, AWSServiceID: elasticbeanstalk.ServiceID, ProviderNameUpper: "ElasticBeanstalk", HCLKeys: []string{"elasticbeanstalk"}}
	serviceData[ElasticInference] = &ServiceDatum{AWSClientName: "ElasticInference", AWSServiceName: elasticinference.ServiceName, AWSEndpointsID: elasticinference.EndpointsID, AWSServiceID: elasticinference.ServiceID, ProviderNameUpper: "ElasticInference", HCLKeys: []string{"elasticinference"}}
	serviceData[Elasticsearch] = &ServiceDatum{AWSClientName: "ElasticsearchService", AWSServiceName: elasticsearch.ServiceName, AWSEndpointsID: elasticsearch.EndpointsID, AWSServiceID: elasticsearch.ServiceID, ProviderNameUpper: "Elasticsearch", HCLKeys: []string{"es", "elasticsearch", "elasticsearchservice"}}
	serviceData[ElasticTranscoder] = &ServiceDatum{AWSClientName: "ElasticTranscoder", AWSServiceName: elastictranscoder.ServiceName, AWSEndpointsID: elastictranscoder.EndpointsID, AWSServiceID: elastictranscoder.ServiceID, ProviderNameUpper: "ElasticTranscoder", HCLKeys: []string{"elastictranscoder"}}
	serviceData[ELB] = &ServiceDatum{AWSClientName: "ELB", AWSServiceName: elb.ServiceName, AWSEndpointsID: elb.EndpointsID, AWSServiceID: elb.ServiceID, ProviderNameUpper: "ELB", HCLKeys: []string{"elb"}}
	serviceData[ELBV2] = &ServiceDatum{AWSClientName: "ELBV2", AWSServiceName: elbv2.ServiceName, AWSEndpointsID: elbv2.EndpointsID, AWSServiceID: elbv2.ServiceID, ProviderNameUpper: "ELBV2", HCLKeys: []string{"elbv2"}}
	serviceData[EMR] = &ServiceDatum{AWSClientName: "EMR", AWSServiceName: emr.ServiceName, AWSEndpointsID: emr.EndpointsID, AWSServiceID: emr.ServiceID, ProviderNameUpper: "EMR", HCLKeys: []string{"emr"}}
	serviceData[EMRContainers] = &ServiceDatum{AWSClientName: "EMRContainers", AWSServiceName: emrcontainers.ServiceName, AWSEndpointsID: emrcontainers.EndpointsID, AWSServiceID: emrcontainers.ServiceID, ProviderNameUpper: "EMRContainers", HCLKeys: []string{"emrcontainers"}}
	serviceData[Events] = &ServiceDatum{AWSClientName: "EventBridge", AWSServiceName: eventbridge.ServiceName, AWSEndpointsID: eventbridge.EndpointsID, AWSServiceID: eventbridge.ServiceID, ProviderNameUpper: "Events", HCLKeys: []string{"cloudwatchevents", "eventbridge", "events"}}
	serviceData[FinSpace] = &ServiceDatum{AWSClientName: "Finspace", AWSServiceName: finspace.ServiceName, AWSEndpointsID: finspace.EndpointsID, AWSServiceID: finspace.ServiceID, ProviderNameUpper: "FinSpace", HCLKeys: []string{"finspace"}}
	serviceData[FinSpaceData] = &ServiceDatum{AWSClientName: "FinSpaceData", AWSServiceName: finspacedata.ServiceName, AWSEndpointsID: finspacedata.EndpointsID, AWSServiceID: finspacedata.ServiceID, ProviderNameUpper: "FinSpaceData", HCLKeys: []string{"finspacedata"}}
	serviceData[Firehose] = &ServiceDatum{AWSClientName: "Firehose", AWSServiceName: firehose.ServiceName, AWSEndpointsID: firehose.EndpointsID, AWSServiceID: firehose.ServiceID, ProviderNameUpper: "Firehose", HCLKeys: []string{"firehose"}}
	serviceData[FIS] = &ServiceDatum{AWSClientName: "FIS", AWSServiceName: fis.ServiceName, AWSEndpointsID: fis.EndpointsID, AWSServiceID: fis.ServiceID, ProviderNameUpper: "FIS", HCLKeys: []string{"fis"}}
	serviceData[FMS] = &ServiceDatum{AWSClientName: "FMS", AWSServiceName: fms.ServiceName, AWSEndpointsID: fms.EndpointsID, AWSServiceID: fms.ServiceID, ProviderNameUpper: "FMS", HCLKeys: []string{"fms"}}
	serviceData[Forecast] = &ServiceDatum{AWSClientName: "ForecastService", AWSServiceName: forecastservice.ServiceName, AWSEndpointsID: forecastservice.EndpointsID, AWSServiceID: forecastservice.ServiceID, ProviderNameUpper: "Forecast", HCLKeys: []string{"forecast", "forecastservice"}}
	serviceData[ForecastQuery] = &ServiceDatum{AWSClientName: "ForecastQueryService", AWSServiceName: forecastqueryservice.ServiceName, AWSEndpointsID: forecastqueryservice.EndpointsID, AWSServiceID: forecastqueryservice.ServiceID, ProviderNameUpper: "ForecastQuery", HCLKeys: []string{"forecastquery", "forecastqueryservice"}}
	serviceData[FraudDetector] = &ServiceDatum{AWSClientName: "FraudDetector", AWSServiceName: frauddetector.ServiceName, AWSEndpointsID: frauddetector.EndpointsID, AWSServiceID: frauddetector.ServiceID, ProviderNameUpper: "FraudDetector", HCLKeys: []string{"frauddetector"}}
	serviceData[FSx] = &ServiceDatum{AWSClientName: "FSx", AWSServiceName: fsx.ServiceName, AWSEndpointsID: fsx.EndpointsID, AWSServiceID: fsx.ServiceID, ProviderNameUpper: "FSx", HCLKeys: []string{"fsx"}}
	serviceData[GameLift] = &ServiceDatum{AWSClientName: "GameLift", AWSServiceName: gamelift.ServiceName, AWSEndpointsID: gamelift.EndpointsID, AWSServiceID: gamelift.ServiceID, ProviderNameUpper: "GameLift", HCLKeys: []string{"gamelift"}}
	serviceData[Glacier] = &ServiceDatum{AWSClientName: "Glacier", AWSServiceName: glacier.ServiceName, AWSEndpointsID: glacier.EndpointsID, AWSServiceID: glacier.ServiceID, ProviderNameUpper: "Glacier", HCLKeys: []string{"glacier"}}
	serviceData[GlobalAccelerator] = &ServiceDatum{AWSClientName: "GlobalAccelerator", AWSServiceName: globalaccelerator.ServiceName, AWSEndpointsID: globalaccelerator.EndpointsID, AWSServiceID: globalaccelerator.ServiceID, ProviderNameUpper: "GlobalAccelerator", HCLKeys: []string{"globalaccelerator"}}
	serviceData[Glue] = &ServiceDatum{AWSClientName: "Glue", AWSServiceName: glue.ServiceName, AWSEndpointsID: glue.EndpointsID, AWSServiceID: glue.ServiceID, ProviderNameUpper: "Glue", HCLKeys: []string{"glue"}}
	serviceData[GlueDataBrew] = &ServiceDatum{AWSClientName: "GlueDataBrew", AWSServiceName: gluedatabrew.ServiceName, AWSEndpointsID: gluedatabrew.EndpointsID, AWSServiceID: gluedatabrew.ServiceID, ProviderNameUpper: "GlueDataBrew", HCLKeys: []string{"gluedatabrew"}}
	serviceData[Greengrass] = &ServiceDatum{AWSClientName: "Greengrass", AWSServiceName: greengrass.ServiceName, AWSEndpointsID: greengrass.EndpointsID, AWSServiceID: greengrass.ServiceID, ProviderNameUpper: "Greengrass", HCLKeys: []string{"greengrass"}}
	serviceData[GreengrassV2] = &ServiceDatum{AWSClientName: "GreengrassV2", AWSServiceName: greengrassv2.ServiceName, AWSEndpointsID: greengrassv2.EndpointsID, AWSServiceID: greengrassv2.ServiceID, ProviderNameUpper: "GreengrassV2", HCLKeys: []string{"greengrassv2"}}
	serviceData[GroundStation] = &ServiceDatum{AWSClientName: "GroundStation", AWSServiceName: groundstation.ServiceName, AWSEndpointsID: groundstation.EndpointsID, AWSServiceID: groundstation.ServiceID, ProviderNameUpper: "GroundStation", HCLKeys: []string{"groundstation"}}
	serviceData[GuardDuty] = &ServiceDatum{AWSClientName: "GuardDuty", AWSServiceName: guardduty.ServiceName, AWSEndpointsID: guardduty.EndpointsID, AWSServiceID: guardduty.ServiceID, ProviderNameUpper: "GuardDuty", HCLKeys: []string{"guardduty"}}
	serviceData[Health] = &ServiceDatum{AWSClientName: "Health", AWSServiceName: health.ServiceName, AWSEndpointsID: health.EndpointsID, AWSServiceID: health.ServiceID, ProviderNameUpper: "Health", HCLKeys: []string{"health"}}
	serviceData[HealthLake] = &ServiceDatum{AWSClientName: "HealthLake", AWSServiceName: healthlake.ServiceName, AWSEndpointsID: healthlake.EndpointsID, AWSServiceID: healthlake.ServiceID, ProviderNameUpper: "HealthLake", HCLKeys: []string{"healthlake"}}
	serviceData[Honeycode] = &ServiceDatum{AWSClientName: "Honeycode", AWSServiceName: honeycode.ServiceName, AWSEndpointsID: honeycode.EndpointsID, AWSServiceID: honeycode.ServiceID, ProviderNameUpper: "Honeycode", HCLKeys: []string{"honeycode"}}
	serviceData[IAM] = &ServiceDatum{AWSClientName: "IAM", AWSServiceName: iam.ServiceName, AWSEndpointsID: iam.EndpointsID, AWSServiceID: iam.ServiceID, ProviderNameUpper: "IAM", HCLKeys: []string{"iam"}}
	serviceData[IdentityStore] = &ServiceDatum{AWSClientName: "IdentityStore", AWSServiceName: identitystore.ServiceName, AWSEndpointsID: identitystore.EndpointsID, AWSServiceID: identitystore.ServiceID, ProviderNameUpper: "IdentityStore", HCLKeys: []string{"identitystore"}}
	serviceData[ImageBuilder] = &ServiceDatum{AWSClientName: "ImageBuilder", AWSServiceName: imagebuilder.ServiceName, AWSEndpointsID: imagebuilder.EndpointsID, AWSServiceID: imagebuilder.ServiceID, ProviderNameUpper: "ImageBuilder", HCLKeys: []string{"imagebuilder"}}
	serviceData[Inspector] = &ServiceDatum{AWSClientName: "Inspector", AWSServiceName: inspector.ServiceName, AWSEndpointsID: inspector.EndpointsID, AWSServiceID: inspector.ServiceID, ProviderNameUpper: "Inspector", HCLKeys: []string{"inspector"}}
	serviceData[IoT] = &ServiceDatum{AWSClientName: "IoT", AWSServiceName: iot.ServiceName, AWSEndpointsID: iot.EndpointsID, AWSServiceID: iot.ServiceID, ProviderNameUpper: "IoT", HCLKeys: []string{"iot"}}
	serviceData[IoT1ClickDevices] = &ServiceDatum{AWSClientName: "IoT1ClickDevicesService", AWSServiceName: iot1clickdevicesservice.ServiceName, AWSEndpointsID: iot1clickdevicesservice.EndpointsID, AWSServiceID: iot1clickdevicesservice.ServiceID, ProviderNameUpper: "IoT1ClickDevices", HCLKeys: []string{"iot1clickdevices", "iot1clickdevicesservice"}}
	serviceData[IoT1ClickProjects] = &ServiceDatum{AWSClientName: "IoT1ClickProjects", AWSServiceName: iot1clickprojects.ServiceName, AWSEndpointsID: iot1clickprojects.EndpointsID, AWSServiceID: iot1clickprojects.ServiceID, ProviderNameUpper: "IoT1ClickProjects", HCLKeys: []string{"iot1clickprojects"}}
	serviceData[IoTAnalytics] = &ServiceDatum{AWSClientName: "IoTAnalytics", AWSServiceName: iotanalytics.ServiceName, AWSEndpointsID: iotanalytics.EndpointsID, AWSServiceID: iotanalytics.ServiceID, ProviderNameUpper: "IoTAnalytics", HCLKeys: []string{"iotanalytics"}}
	serviceData[IoTDataPlane] = &ServiceDatum{AWSClientName: "IoTDataPlane", AWSServiceName: iotdataplane.ServiceName, AWSEndpointsID: iotdataplane.EndpointsID, AWSServiceID: iotdataplane.ServiceID, ProviderNameUpper: "IoTDataPlane", HCLKeys: []string{"iotdataplane"}}
	serviceData[IoTDeviceAdvisor] = &ServiceDatum{AWSClientName: "IoTDeviceAdvisor", AWSServiceName: iotdeviceadvisor.ServiceName, AWSEndpointsID: iotdeviceadvisor.EndpointsID, AWSServiceID: iotdeviceadvisor.ServiceID, ProviderNameUpper: "IoTDeviceAdvisor", HCLKeys: []string{"iotdeviceadvisor"}}
	serviceData[IoTEvents] = &ServiceDatum{AWSClientName: "IoTEvents", AWSServiceName: iotevents.ServiceName, AWSEndpointsID: iotevents.EndpointsID, AWSServiceID: iotevents.ServiceID, ProviderNameUpper: "IoTEvents", HCLKeys: []string{"iotevents"}}
	serviceData[IoTEventsData] = &ServiceDatum{AWSClientName: "IoTEventsData", AWSServiceName: ioteventsdata.ServiceName, AWSEndpointsID: ioteventsdata.EndpointsID, AWSServiceID: ioteventsdata.ServiceID, ProviderNameUpper: "IoTEventsData", HCLKeys: []string{"ioteventsdata"}}
	serviceData[IoTFleetHub] = &ServiceDatum{AWSClientName: "IoTFleetHub", AWSServiceName: iotfleethub.ServiceName, AWSEndpointsID: iotfleethub.EndpointsID, AWSServiceID: iotfleethub.ServiceID, ProviderNameUpper: "IoTFleetHub", HCLKeys: []string{"iotfleethub"}}
	serviceData[IoTJobsDataPlane] = &ServiceDatum{AWSClientName: "IoTJobsDataPlane", AWSServiceName: iotjobsdataplane.ServiceName, AWSEndpointsID: iotjobsdataplane.EndpointsID, AWSServiceID: iotjobsdataplane.ServiceID, ProviderNameUpper: "IoTJobsDataPlane", HCLKeys: []string{"iotjobsdataplane"}}
	serviceData[IoTSecureTunneling] = &ServiceDatum{AWSClientName: "IoTSecureTunneling", AWSServiceName: iotsecuretunneling.ServiceName, AWSEndpointsID: iotsecuretunneling.EndpointsID, AWSServiceID: iotsecuretunneling.ServiceID, ProviderNameUpper: "IoTSecureTunneling", HCLKeys: []string{"iotsecuretunneling"}}
	serviceData[IoTSiteWise] = &ServiceDatum{AWSClientName: "IoTSiteWise", AWSServiceName: iotsitewise.ServiceName, AWSEndpointsID: iotsitewise.EndpointsID, AWSServiceID: iotsitewise.ServiceID, ProviderNameUpper: "IoTSiteWise", HCLKeys: []string{"iotsitewise"}}
	serviceData[IoTThingsGraph] = &ServiceDatum{AWSClientName: "IoTThingsGraph", AWSServiceName: iotthingsgraph.ServiceName, AWSEndpointsID: iotthingsgraph.EndpointsID, AWSServiceID: iotthingsgraph.ServiceID, ProviderNameUpper: "IoTThingsGraph", HCLKeys: []string{"iotthingsgraph"}}
	serviceData[IoTWireless] = &ServiceDatum{AWSClientName: "IoTWireless", AWSServiceName: iotwireless.ServiceName, AWSEndpointsID: iotwireless.EndpointsID, AWSServiceID: iotwireless.ServiceID, ProviderNameUpper: "IoTWireless", HCLKeys: []string{"iotwireless"}}
	serviceData[Kafka] = &ServiceDatum{AWSClientName: "Kafka", AWSServiceName: kafka.ServiceName, AWSEndpointsID: kafka.EndpointsID, AWSServiceID: kafka.ServiceID, ProviderNameUpper: "Kafka", HCLKeys: []string{"kafka"}}
	serviceData[KafkaConnect] = &ServiceDatum{AWSClientName: "KafkaConnect", AWSServiceName: kafkaconnect.ServiceName, AWSEndpointsID: kafkaconnect.EndpointsID, AWSServiceID: kafkaconnect.ServiceID, ProviderNameUpper: "KafkaConnect", HCLKeys: []string{"kafkaconnect"}}
	serviceData[Kendra] = &ServiceDatum{AWSClientName: "Kendra", AWSServiceName: kendra.ServiceName, AWSEndpointsID: kendra.EndpointsID, AWSServiceID: kendra.ServiceID, ProviderNameUpper: "Kendra", HCLKeys: []string{"kendra"}}
	serviceData[Kinesis] = &ServiceDatum{AWSClientName: "Kinesis", AWSServiceName: kinesis.ServiceName, AWSEndpointsID: kinesis.EndpointsID, AWSServiceID: kinesis.ServiceID, ProviderNameUpper: "Kinesis", HCLKeys: []string{"kinesis"}}
	serviceData[KinesisAnalytics] = &ServiceDatum{AWSClientName: "KinesisAnalytics", AWSServiceName: kinesisanalytics.ServiceName, AWSEndpointsID: kinesisanalytics.EndpointsID, AWSServiceID: kinesisanalytics.ServiceID, ProviderNameUpper: "KinesisAnalytics", HCLKeys: []string{"kinesisanalytics"}}
	serviceData[KinesisAnalyticsV2] = &ServiceDatum{AWSClientName: "KinesisAnalyticsV2", AWSServiceName: kinesisanalyticsv2.ServiceName, AWSEndpointsID: kinesisanalyticsv2.EndpointsID, AWSServiceID: kinesisanalyticsv2.ServiceID, ProviderNameUpper: "KinesisAnalyticsV2", HCLKeys: []string{"kinesisanalyticsv2"}}
	serviceData[KinesisVideo] = &ServiceDatum{AWSClientName: "KinesisVideo", AWSServiceName: kinesisvideo.ServiceName, AWSEndpointsID: kinesisvideo.EndpointsID, AWSServiceID: kinesisvideo.ServiceID, ProviderNameUpper: "KinesisVideo", HCLKeys: []string{"kinesisvideo"}}
	serviceData[KinesisVideoArchivedMedia] = &ServiceDatum{AWSClientName: "KinesisVideoArchivedMedia", AWSServiceName: kinesisvideoarchivedmedia.ServiceName, AWSEndpointsID: kinesisvideoarchivedmedia.EndpointsID, AWSServiceID: kinesisvideoarchivedmedia.ServiceID, ProviderNameUpper: "KinesisVideoArchivedMedia", HCLKeys: []string{"kinesisvideoarchivedmedia"}}
	serviceData[KinesisVideoMedia] = &ServiceDatum{AWSClientName: "KinesisVideoMedia", AWSServiceName: kinesisvideomedia.ServiceName, AWSEndpointsID: kinesisvideomedia.EndpointsID, AWSServiceID: kinesisvideomedia.ServiceID, ProviderNameUpper: "KinesisVideoMedia", HCLKeys: []string{"kinesisvideomedia"}}
	serviceData[KinesisVideoSignalingChannels] = &ServiceDatum{AWSClientName: "KinesisVideoSignalingChannels", AWSServiceName: kinesisvideosignalingchannels.ServiceName, AWSEndpointsID: kinesisvideosignalingchannels.EndpointsID, AWSServiceID: kinesisvideosignalingchannels.ServiceID, ProviderNameUpper: "KinesisVideoSignalingChannels", HCLKeys: []string{"kinesisvideosignalingchannels"}}
	serviceData[KMS] = &ServiceDatum{AWSClientName: "KMS", AWSServiceName: kms.ServiceName, AWSEndpointsID: kms.EndpointsID, AWSServiceID: kms.ServiceID, ProviderNameUpper: "KMS", HCLKeys: []string{"kms"}}
	serviceData[LakeFormation] = &ServiceDatum{AWSClientName: "LakeFormation", AWSServiceName: lakeformation.ServiceName, AWSEndpointsID: lakeformation.EndpointsID, AWSServiceID: lakeformation.ServiceID, ProviderNameUpper: "LakeFormation", HCLKeys: []string{"lakeformation"}}
	serviceData[Lambda] = &ServiceDatum{AWSClientName: "Lambda", AWSServiceName: lambda.ServiceName, AWSEndpointsID: lambda.EndpointsID, AWSServiceID: lambda.ServiceID, ProviderNameUpper: "Lambda", HCLKeys: []string{"lambda"}}
	serviceData[LexModels] = &ServiceDatum{AWSClientName: "LexModelBuildingService", AWSServiceName: lexmodelbuildingservice.ServiceName, AWSEndpointsID: lexmodelbuildingservice.EndpointsID, AWSServiceID: lexmodelbuildingservice.ServiceID, ProviderNameUpper: "LexModels", HCLKeys: []string{"lexmodels", "lexmodelbuilding", "lexmodelbuildingservice"}}
	serviceData[LexModelsV2] = &ServiceDatum{AWSClientName: "LexModelsV2", AWSServiceName: lexmodelsv2.ServiceName, AWSEndpointsID: lexmodelsv2.EndpointsID, AWSServiceID: lexmodelsv2.ServiceID, ProviderNameUpper: "LexModelsV2", HCLKeys: []string{"lexmodelsv2"}}
	serviceData[LexRuntime] = &ServiceDatum{AWSClientName: "LexRuntimeService", AWSServiceName: lexruntimeservice.ServiceName, AWSEndpointsID: lexruntimeservice.EndpointsID, AWSServiceID: lexruntimeservice.ServiceID, ProviderNameUpper: "LexRuntime", HCLKeys: []string{"lexruntime", "lexruntimeservice"}}
	serviceData[LexRuntimeV2] = &ServiceDatum{AWSClientName: "LexRuntimeV2", AWSServiceName: lexruntimev2.ServiceName, AWSEndpointsID: lexruntimev2.EndpointsID, AWSServiceID: lexruntimev2.ServiceID, ProviderNameUpper: "LexRuntimeV2", HCLKeys: []string{"lexruntimev2"}}
	serviceData[LicenseManager] = &ServiceDatum{AWSClientName: "LicenseManager", AWSServiceName: licensemanager.ServiceName, AWSEndpointsID: licensemanager.EndpointsID, AWSServiceID: licensemanager.ServiceID, ProviderNameUpper: "LicenseManager", HCLKeys: []string{"licensemanager"}}
	serviceData[Lightsail] = &ServiceDatum{AWSClientName: "Lightsail", AWSServiceName: lightsail.ServiceName, AWSEndpointsID: lightsail.EndpointsID, AWSServiceID: lightsail.ServiceID, ProviderNameUpper: "Lightsail", HCLKeys: []string{"lightsail"}}
	serviceData[Location] = &ServiceDatum{AWSClientName: "LocationService", AWSServiceName: locationservice.ServiceName, AWSEndpointsID: locationservice.EndpointsID, AWSServiceID: locationservice.ServiceID, ProviderNameUpper: "Location", HCLKeys: []string{"location"}}
	serviceData[LookoutEquipment] = &ServiceDatum{AWSClientName: "LookoutEquipment", AWSServiceName: lookoutequipment.ServiceName, AWSEndpointsID: lookoutequipment.EndpointsID, AWSServiceID: lookoutequipment.ServiceID, ProviderNameUpper: "LookoutEquipment", HCLKeys: []string{"lookoutequipment"}}
	serviceData[LookoutForVision] = &ServiceDatum{AWSClientName: "LookoutForVision", AWSServiceName: lookoutforvision.ServiceName, AWSEndpointsID: lookoutforvision.EndpointsID, AWSServiceID: lookoutforvision.ServiceID, ProviderNameUpper: "LookoutForVision", HCLKeys: []string{"lookoutforvision"}}
	serviceData[LookoutMetrics] = &ServiceDatum{AWSClientName: "LookoutMetrics", AWSServiceName: lookoutmetrics.ServiceName, AWSEndpointsID: lookoutmetrics.EndpointsID, AWSServiceID: lookoutmetrics.ServiceID, ProviderNameUpper: "LookoutMetrics", HCLKeys: []string{"lookoutmetrics"}}
	serviceData[MachineLearning] = &ServiceDatum{AWSClientName: "MachineLearning", AWSServiceName: machinelearning.ServiceName, AWSEndpointsID: machinelearning.EndpointsID, AWSServiceID: machinelearning.ServiceID, ProviderNameUpper: "MachineLearning", HCLKeys: []string{"machinelearning"}}
	serviceData[Macie] = &ServiceDatum{AWSClientName: "Macie", AWSServiceName: macie.ServiceName, AWSEndpointsID: macie.EndpointsID, AWSServiceID: macie.ServiceID, ProviderNameUpper: "Macie", HCLKeys: []string{"macie"}}
	serviceData[Macie2] = &ServiceDatum{AWSClientName: "Macie2", AWSServiceName: macie2.ServiceName, AWSEndpointsID: macie2.EndpointsID, AWSServiceID: macie2.ServiceID, ProviderNameUpper: "Macie2", HCLKeys: []string{"macie2"}}
	serviceData[ManagedBlockchain] = &ServiceDatum{AWSClientName: "ManagedBlockchain", AWSServiceName: managedblockchain.ServiceName, AWSEndpointsID: managedblockchain.EndpointsID, AWSServiceID: managedblockchain.ServiceID, ProviderNameUpper: "ManagedBlockchain", HCLKeys: []string{"managedblockchain"}}
	serviceData[MarketplaceCatalog] = &ServiceDatum{AWSClientName: "MarketplaceCatalog", AWSServiceName: marketplacecatalog.ServiceName, AWSEndpointsID: marketplacecatalog.EndpointsID, AWSServiceID: marketplacecatalog.ServiceID, ProviderNameUpper: "MarketplaceCatalog", HCLKeys: []string{"marketplacecatalog"}}
	serviceData[MarketplaceCommerceAnalytics] = &ServiceDatum{AWSClientName: "MarketplaceCommerceAnalytics", AWSServiceName: marketplacecommerceanalytics.ServiceName, AWSEndpointsID: marketplacecommerceanalytics.EndpointsID, AWSServiceID: marketplacecommerceanalytics.ServiceID, ProviderNameUpper: "MarketplaceCommerceAnalytics", HCLKeys: []string{"marketplacecommerceanalytics"}}
	serviceData[MarketplaceEntitlement] = &ServiceDatum{AWSClientName: "MarketplaceEntitlementService", AWSServiceName: marketplaceentitlementservice.ServiceName, AWSEndpointsID: marketplaceentitlementservice.EndpointsID, AWSServiceID: marketplaceentitlementservice.ServiceID, ProviderNameUpper: "MarketplaceEntitlement", HCLKeys: []string{"marketplaceentitlement", "marketplaceentitlementservice"}}
	serviceData[MarketplaceMetering] = &ServiceDatum{AWSClientName: "MarketplaceMetering", AWSServiceName: marketplacemetering.ServiceName, AWSEndpointsID: marketplacemetering.EndpointsID, AWSServiceID: marketplacemetering.ServiceID, ProviderNameUpper: "MarketplaceMetering", HCLKeys: []string{"marketplacemetering"}}
	serviceData[MediaConnect] = &ServiceDatum{AWSClientName: "MediaConnect", AWSServiceName: mediaconnect.ServiceName, AWSEndpointsID: mediaconnect.EndpointsID, AWSServiceID: mediaconnect.ServiceID, ProviderNameUpper: "MediaConnect", HCLKeys: []string{"mediaconnect"}}
	serviceData[MediaConvert] = &ServiceDatum{AWSClientName: "MediaConvert", AWSServiceName: mediaconvert.ServiceName, AWSEndpointsID: mediaconvert.EndpointsID, AWSServiceID: mediaconvert.ServiceID, ProviderNameUpper: "MediaConvert", HCLKeys: []string{"mediaconvert"}}
	serviceData[MediaLive] = &ServiceDatum{AWSClientName: "MediaLive", AWSServiceName: medialive.ServiceName, AWSEndpointsID: medialive.EndpointsID, AWSServiceID: medialive.ServiceID, ProviderNameUpper: "MediaLive", HCLKeys: []string{"medialive"}}
	serviceData[MediaPackage] = &ServiceDatum{AWSClientName: "MediaPackage", AWSServiceName: mediapackage.ServiceName, AWSEndpointsID: mediapackage.EndpointsID, AWSServiceID: mediapackage.ServiceID, ProviderNameUpper: "MediaPackage", HCLKeys: []string{"mediapackage"}}
	serviceData[MediaPackageVOD] = &ServiceDatum{AWSClientName: "MediaPackageVOD", AWSServiceName: mediapackagevod.ServiceName, AWSEndpointsID: mediapackagevod.EndpointsID, AWSServiceID: mediapackagevod.ServiceID, ProviderNameUpper: "MediaPackageVOD", HCLKeys: []string{"mediapackagevod"}}
	serviceData[MediaStore] = &ServiceDatum{AWSClientName: "MediaStore", AWSServiceName: mediastore.ServiceName, AWSEndpointsID: mediastore.EndpointsID, AWSServiceID: mediastore.ServiceID, ProviderNameUpper: "MediaStore", HCLKeys: []string{"mediastore"}}
	serviceData[MediaStoreData] = &ServiceDatum{AWSClientName: "MediaStoreData", AWSServiceName: mediastoredata.ServiceName, AWSEndpointsID: mediastoredata.EndpointsID, AWSServiceID: mediastoredata.ServiceID, ProviderNameUpper: "MediaStoreData", HCLKeys: []string{"mediastoredata"}}
	serviceData[MediaTailor] = &ServiceDatum{AWSClientName: "MediaTailor", AWSServiceName: mediatailor.ServiceName, AWSEndpointsID: mediatailor.EndpointsID, AWSServiceID: mediatailor.ServiceID, ProviderNameUpper: "MediaTailor", HCLKeys: []string{"mediatailor"}}
	serviceData[MemoryDB] = &ServiceDatum{AWSClientName: "MemoryDB", AWSServiceName: memorydb.ServiceName, AWSEndpointsID: memorydb.EndpointsID, AWSServiceID: memorydb.ServiceID, ProviderNameUpper: "MemoryDB", HCLKeys: []string{"memorydb"}}
	serviceData[Mgn] = &ServiceDatum{AWSClientName: "Mgn", AWSServiceName: mgn.ServiceName, AWSEndpointsID: mgn.EndpointsID, AWSServiceID: mgn.ServiceID, ProviderNameUpper: "Mgn", HCLKeys: []string{"mgn"}}
	serviceData[MigrationHub] = &ServiceDatum{AWSClientName: "MigrationHub", AWSServiceName: migrationhub.ServiceName, AWSEndpointsID: migrationhub.EndpointsID, AWSServiceID: migrationhub.ServiceID, ProviderNameUpper: "MigrationHub", HCLKeys: []string{"migrationhub"}}
	serviceData[MigrationHubConfig] = &ServiceDatum{AWSClientName: "MigrationHubConfig", AWSServiceName: migrationhubconfig.ServiceName, AWSEndpointsID: migrationhubconfig.EndpointsID, AWSServiceID: migrationhubconfig.ServiceID, ProviderNameUpper: "MigrationHubConfig", HCLKeys: []string{"migrationhubconfig"}}
	serviceData[Mobile] = &ServiceDatum{AWSClientName: "Mobile", AWSServiceName: mobile.ServiceName, AWSEndpointsID: mobile.EndpointsID, AWSServiceID: mobile.ServiceID, ProviderNameUpper: "Mobile", HCLKeys: []string{"mobile"}}
	serviceData[MobileAnalytics] = &ServiceDatum{AWSClientName: "MobileAnalytics", AWSServiceName: mobileanalytics.ServiceName, AWSEndpointsID: mobileanalytics.EndpointsID, AWSServiceID: mobileanalytics.ServiceID, ProviderNameUpper: "MobileAnalytics", HCLKeys: []string{"mobileanalytics"}}
	serviceData[MQ] = &ServiceDatum{AWSClientName: "MQ", AWSServiceName: mq.ServiceName, AWSEndpointsID: mq.EndpointsID, AWSServiceID: mq.ServiceID, ProviderNameUpper: "MQ", HCLKeys: []string{"mq"}}
	serviceData[MTurk] = &ServiceDatum{AWSClientName: "MTurk", AWSServiceName: mturk.ServiceName, AWSEndpointsID: mturk.EndpointsID, AWSServiceID: mturk.ServiceID, ProviderNameUpper: "MTurk", HCLKeys: []string{"mturk"}}
	serviceData[MWAA] = &ServiceDatum{AWSClientName: "MWAA", AWSServiceName: mwaa.ServiceName, AWSEndpointsID: mwaa.EndpointsID, AWSServiceID: mwaa.ServiceID, ProviderNameUpper: "MWAA", HCLKeys: []string{"mwaa"}}
	serviceData[Neptune] = &ServiceDatum{AWSClientName: "Neptune", AWSServiceName: neptune.ServiceName, AWSEndpointsID: neptune.EndpointsID, AWSServiceID: neptune.ServiceID, ProviderNameUpper: "Neptune", HCLKeys: []string{"neptune"}}
	serviceData[NetworkFirewall] = &ServiceDatum{AWSClientName: "NetworkFirewall", AWSServiceName: networkfirewall.ServiceName, AWSEndpointsID: networkfirewall.EndpointsID, AWSServiceID: networkfirewall.ServiceID, ProviderNameUpper: "NetworkFirewall", HCLKeys: []string{"networkfirewall"}}
	serviceData[NetworkManager] = &ServiceDatum{AWSClientName: "NetworkManager", AWSServiceName: networkmanager.ServiceName, AWSEndpointsID: networkmanager.EndpointsID, AWSServiceID: networkmanager.ServiceID, ProviderNameUpper: "NetworkManager", HCLKeys: []string{"networkmanager"}}
	serviceData[NimbleStudio] = &ServiceDatum{AWSClientName: "NimbleStudio", AWSServiceName: nimblestudio.ServiceName, AWSEndpointsID: nimblestudio.EndpointsID, AWSServiceID: nimblestudio.ServiceID, ProviderNameUpper: "NimbleStudio", HCLKeys: []string{"nimblestudio"}}
	serviceData[OpsWorks] = &ServiceDatum{AWSClientName: "OpsWorks", AWSServiceName: opsworks.ServiceName, AWSEndpointsID: opsworks.EndpointsID, AWSServiceID: opsworks.ServiceID, ProviderNameUpper: "OpsWorks", HCLKeys: []string{"opsworks"}}
	serviceData[OpsWorksCM] = &ServiceDatum{AWSClientName: "OpsWorksCM", AWSServiceName: opsworkscm.ServiceName, AWSEndpointsID: opsworkscm.EndpointsID, AWSServiceID: opsworkscm.ServiceID, ProviderNameUpper: "OpsWorksCM", HCLKeys: []string{"opsworkscm"}}
	serviceData[Organizations] = &ServiceDatum{AWSClientName: "Organizations", AWSServiceName: organizations.ServiceName, AWSEndpointsID: organizations.EndpointsID, AWSServiceID: organizations.ServiceID, ProviderNameUpper: "Organizations", HCLKeys: []string{"organizations"}}
	serviceData[Outposts] = &ServiceDatum{AWSClientName: "Outposts", AWSServiceName: outposts.ServiceName, AWSEndpointsID: outposts.EndpointsID, AWSServiceID: outposts.ServiceID, ProviderNameUpper: "Outposts", HCLKeys: []string{"outposts"}}
	serviceData[Personalize] = &ServiceDatum{AWSClientName: "Personalize", AWSServiceName: personalize.ServiceName, AWSEndpointsID: personalize.EndpointsID, AWSServiceID: personalize.ServiceID, ProviderNameUpper: "Personalize", HCLKeys: []string{"personalize"}}
	serviceData[PersonalizeEvents] = &ServiceDatum{AWSClientName: "PersonalizeEvents", AWSServiceName: personalizeevents.ServiceName, AWSEndpointsID: personalizeevents.EndpointsID, AWSServiceID: personalizeevents.ServiceID, ProviderNameUpper: "PersonalizeEvents", HCLKeys: []string{"personalizeevents"}}
	serviceData[PersonalizeRuntime] = &ServiceDatum{AWSClientName: "PersonalizeRuntime", AWSServiceName: personalizeruntime.ServiceName, AWSEndpointsID: personalizeruntime.EndpointsID, AWSServiceID: personalizeruntime.ServiceID, ProviderNameUpper: "PersonalizeRuntime", HCLKeys: []string{"personalizeruntime"}}
	serviceData[PI] = &ServiceDatum{AWSClientName: "PI", AWSServiceName: pi.ServiceName, AWSEndpointsID: pi.EndpointsID, AWSServiceID: pi.ServiceID, ProviderNameUpper: "PI", HCLKeys: []string{"pi"}}
	serviceData[Pinpoint] = &ServiceDatum{AWSClientName: "Pinpoint", AWSServiceName: pinpoint.ServiceName, AWSEndpointsID: pinpoint.EndpointsID, AWSServiceID: pinpoint.ServiceID, ProviderNameUpper: "Pinpoint", HCLKeys: []string{"pinpoint"}}
	serviceData[PinpointEmail] = &ServiceDatum{AWSClientName: "PinpointEmail", AWSServiceName: pinpointemail.ServiceName, AWSEndpointsID: pinpointemail.EndpointsID, AWSServiceID: pinpointemail.ServiceID, ProviderNameUpper: "PinpointEmail", HCLKeys: []string{"pinpointemail"}}
	serviceData[PinpointSMSVoice] = &ServiceDatum{AWSClientName: "PinpointSMSVoice", AWSServiceName: pinpointsmsvoice.ServiceName, AWSEndpointsID: pinpointsmsvoice.EndpointsID, AWSServiceID: pinpointsmsvoice.ServiceID, ProviderNameUpper: "PinpointSMSVoice", HCLKeys: []string{"pinpointsmsvoice"}}
	serviceData[Polly] = &ServiceDatum{AWSClientName: "Polly", AWSServiceName: polly.ServiceName, AWSEndpointsID: polly.EndpointsID, AWSServiceID: polly.ServiceID, ProviderNameUpper: "Polly", HCLKeys: []string{"polly"}}
	serviceData[Pricing] = &ServiceDatum{AWSClientName: "Pricing", AWSServiceName: pricing.ServiceName, AWSEndpointsID: pricing.EndpointsID, AWSServiceID: pricing.ServiceID, ProviderNameUpper: "Pricing", HCLKeys: []string{"pricing"}}
	serviceData[Proton] = &ServiceDatum{AWSClientName: "Proton", AWSServiceName: proton.ServiceName, AWSEndpointsID: proton.EndpointsID, AWSServiceID: proton.ServiceID, ProviderNameUpper: "Proton", HCLKeys: []string{"proton"}}
	serviceData[QLDB] = &ServiceDatum{AWSClientName: "QLDB", AWSServiceName: qldb.ServiceName, AWSEndpointsID: qldb.EndpointsID, AWSServiceID: qldb.ServiceID, ProviderNameUpper: "QLDB", HCLKeys: []string{"qldb"}}
	serviceData[QLDBSession] = &ServiceDatum{AWSClientName: "QLDBSession", AWSServiceName: qldbsession.ServiceName, AWSEndpointsID: qldbsession.EndpointsID, AWSServiceID: qldbsession.ServiceID, ProviderNameUpper: "QLDBSession", HCLKeys: []string{"qldbsession"}}
	serviceData[QuickSight] = &ServiceDatum{AWSClientName: "QuickSight", AWSServiceName: quicksight.ServiceName, AWSEndpointsID: quicksight.EndpointsID, AWSServiceID: quicksight.ServiceID, ProviderNameUpper: "QuickSight", HCLKeys: []string{"quicksight"}}
	serviceData[RAM] = &ServiceDatum{AWSClientName: "RAM", AWSServiceName: ram.ServiceName, AWSEndpointsID: ram.EndpointsID, AWSServiceID: ram.ServiceID, ProviderNameUpper: "RAM", HCLKeys: []string{"ram"}}
	serviceData[RDS] = &ServiceDatum{AWSClientName: "RDS", AWSServiceName: rds.ServiceName, AWSEndpointsID: rds.EndpointsID, AWSServiceID: rds.ServiceID, ProviderNameUpper: "RDS", HCLKeys: []string{"rds"}}
	serviceData[RDSData] = &ServiceDatum{AWSClientName: "RDSDataService", AWSServiceName: rdsdataservice.ServiceName, AWSEndpointsID: rdsdataservice.EndpointsID, AWSServiceID: rdsdataservice.ServiceID, ProviderNameUpper: "RDSData", HCLKeys: []string{"rdsdata", "rdsdataservice"}}
	serviceData[Redshift] = &ServiceDatum{AWSClientName: "Redshift", AWSServiceName: redshift.ServiceName, AWSEndpointsID: redshift.EndpointsID, AWSServiceID: redshift.ServiceID, ProviderNameUpper: "Redshift", HCLKeys: []string{"redshift"}}
	serviceData[RedshiftData] = &ServiceDatum{AWSClientName: "RedshiftData", AWSServiceName: redshiftdataapiservice.ServiceName, AWSEndpointsID: redshiftdataapiservice.EndpointsID, AWSServiceID: redshiftdataapiservice.ServiceID, ProviderNameUpper: "RedshiftData", HCLKeys: []string{"redshiftdata"}}
	serviceData[Rekognition] = &ServiceDatum{AWSClientName: "Rekognition", AWSServiceName: rekognition.ServiceName, AWSEndpointsID: rekognition.EndpointsID, AWSServiceID: rekognition.ServiceID, ProviderNameUpper: "Rekognition", HCLKeys: []string{"rekognition"}}
	serviceData[ResourceGroups] = &ServiceDatum{AWSClientName: "ResourceGroups", AWSServiceName: resourcegroups.ServiceName, AWSEndpointsID: resourcegroups.EndpointsID, AWSServiceID: resourcegroups.ServiceID, ProviderNameUpper: "ResourceGroups", HCLKeys: []string{"resourcegroups"}}
	serviceData[ResourceGroupsTaggingAPI] = &ServiceDatum{AWSClientName: "ResourceGroupsTaggingAPI", AWSServiceName: resourcegroupstaggingapi.ServiceName, AWSEndpointsID: resourcegroupstaggingapi.EndpointsID, AWSServiceID: resourcegroupstaggingapi.ServiceID, ProviderNameUpper: "ResourceGroupsTaggingAPI", HCLKeys: []string{"resourcegroupstaggingapi", "resourcegroupstagging"}}
	serviceData[RoboMaker] = &ServiceDatum{AWSClientName: "RoboMaker", AWSServiceName: robomaker.ServiceName, AWSEndpointsID: robomaker.EndpointsID, AWSServiceID: robomaker.ServiceID, ProviderNameUpper: "RoboMaker", HCLKeys: []string{"robomaker"}}
	serviceData[Route53] = &ServiceDatum{AWSClientName: "Route53", AWSServiceName: route53.ServiceName, AWSEndpointsID: route53.EndpointsID, AWSServiceID: route53.ServiceID, ProviderNameUpper: "Route53", HCLKeys: []string{"route53"}}
	serviceData[Route53Domains] = &ServiceDatum{AWSClientName: "Route53Domains", AWSServiceName: route53domains.ServiceName, AWSEndpointsID: route53domains.EndpointsID, AWSServiceID: route53domains.ServiceID, ProviderNameUpper: "Route53Domains", HCLKeys: []string{"route53domains"}}
	serviceData[Route53RecoveryControlConfig] = &ServiceDatum{AWSClientName: "Route53RecoveryControlConfig", AWSServiceName: route53recoverycontrolconfig.ServiceName, AWSEndpointsID: route53recoverycontrolconfig.EndpointsID, AWSServiceID: route53recoverycontrolconfig.ServiceID, ProviderNameUpper: "Route53RecoveryControlConfig", HCLKeys: []string{"route53recoverycontrolconfig"}}
	serviceData[Route53RecoveryReadiness] = &ServiceDatum{AWSClientName: "Route53RecoveryReadiness", AWSServiceName: route53recoveryreadiness.ServiceName, AWSEndpointsID: route53recoveryreadiness.EndpointsID, AWSServiceID: route53recoveryreadiness.ServiceID, ProviderNameUpper: "Route53RecoveryReadiness", HCLKeys: []string{"route53recoveryreadiness"}}
	serviceData[Route53Resolver] = &ServiceDatum{AWSClientName: "Route53Resolver", AWSServiceName: route53resolver.ServiceName, AWSEndpointsID: route53resolver.EndpointsID, AWSServiceID: route53resolver.ServiceID, ProviderNameUpper: "Route53Resolver", HCLKeys: []string{"route53resolver"}}
	serviceData[S3] = &ServiceDatum{AWSClientName: "S3", AWSServiceName: s3.ServiceName, AWSEndpointsID: s3.EndpointsID, AWSServiceID: s3.ServiceID, ProviderNameUpper: "S3", HCLKeys: []string{"s3"}}
	serviceData[S3Control] = &ServiceDatum{AWSClientName: "S3Control", AWSServiceName: s3control.ServiceName, AWSEndpointsID: s3control.EndpointsID, AWSServiceID: s3control.ServiceID, ProviderNameUpper: "S3Control", HCLKeys: []string{"s3control"}}
	serviceData[S3Outposts] = &ServiceDatum{AWSClientName: "S3Outposts", AWSServiceName: s3outposts.ServiceName, AWSEndpointsID: s3outposts.EndpointsID, AWSServiceID: s3outposts.ServiceID, ProviderNameUpper: "S3Outposts", HCLKeys: []string{"s3outposts"}}
	serviceData[SageMaker] = &ServiceDatum{AWSClientName: "SageMaker", AWSServiceName: sagemaker.ServiceName, AWSEndpointsID: sagemaker.EndpointsID, AWSServiceID: sagemaker.ServiceID, ProviderNameUpper: "SageMaker", HCLKeys: []string{"sagemaker"}}
	serviceData[SageMakerEdgeManager] = &ServiceDatum{AWSClientName: "SagemakerEdgeManager", AWSServiceName: sagemakeredgemanager.ServiceName, AWSEndpointsID: sagemakeredgemanager.EndpointsID, AWSServiceID: sagemakeredgemanager.ServiceID, ProviderNameUpper: "SageMakerEdgeManager", HCLKeys: []string{"sagemakeredgemanager"}}
	serviceData[SageMakerFeatureStoreRuntime] = &ServiceDatum{AWSClientName: "SageMakerFeatureStoreRuntime", AWSServiceName: sagemakerfeaturestoreruntime.ServiceName, AWSEndpointsID: sagemakerfeaturestoreruntime.EndpointsID, AWSServiceID: sagemakerfeaturestoreruntime.ServiceID, ProviderNameUpper: "SageMakerFeatureStoreRuntime", HCLKeys: []string{"sagemakerfeaturestoreruntime"}}
	serviceData[SageMakerRuntime] = &ServiceDatum{AWSClientName: "SageMakerRuntime", AWSServiceName: sagemakerruntime.ServiceName, AWSEndpointsID: sagemakerruntime.EndpointsID, AWSServiceID: sagemakerruntime.ServiceID, ProviderNameUpper: "SageMakerRuntime", HCLKeys: []string{"sagemakerruntime"}}
	serviceData[SavingsPlans] = &ServiceDatum{AWSClientName: "SavingsPlans", AWSServiceName: savingsplans.ServiceName, AWSEndpointsID: savingsplans.EndpointsID, AWSServiceID: savingsplans.ServiceID, ProviderNameUpper: "SavingsPlans", HCLKeys: []string{"savingsplans"}}
	serviceData[Schemas] = &ServiceDatum{AWSClientName: "Schemas", AWSServiceName: schemas.ServiceName, AWSEndpointsID: schemas.EndpointsID, AWSServiceID: schemas.ServiceID, ProviderNameUpper: "Schemas", HCLKeys: []string{"schemas"}}
	serviceData[SecretsManager] = &ServiceDatum{AWSClientName: "SecretsManager", AWSServiceName: secretsmanager.ServiceName, AWSEndpointsID: secretsmanager.EndpointsID, AWSServiceID: secretsmanager.ServiceID, ProviderNameUpper: "SecretsManager", HCLKeys: []string{"secretsmanager"}}
	serviceData[SecurityHub] = &ServiceDatum{AWSClientName: "SecurityHub", AWSServiceName: securityhub.ServiceName, AWSEndpointsID: securityhub.EndpointsID, AWSServiceID: securityhub.ServiceID, ProviderNameUpper: "SecurityHub", HCLKeys: []string{"securityhub"}}
	serviceData[ServerlessRepo] = &ServiceDatum{AWSClientName: "ServerlessApplicationRepository", AWSServiceName: serverlessapplicationrepository.ServiceName, AWSEndpointsID: serverlessapplicationrepository.EndpointsID, AWSServiceID: serverlessapplicationrepository.ServiceID, ProviderNameUpper: "ServerlessRepo", HCLKeys: []string{"serverlessrepo", "serverlessapprepo", "serverlessapplicationrepository"}}
	serviceData[ServiceCatalog] = &ServiceDatum{AWSClientName: "ServiceCatalog", AWSServiceName: servicecatalog.ServiceName, AWSEndpointsID: servicecatalog.EndpointsID, AWSServiceID: servicecatalog.ServiceID, ProviderNameUpper: "ServiceCatalog", HCLKeys: []string{"servicecatalog"}}
	serviceData[ServiceDiscovery] = &ServiceDatum{AWSClientName: "ServiceDiscovery", AWSServiceName: servicediscovery.ServiceName, AWSEndpointsID: servicediscovery.EndpointsID, AWSServiceID: servicediscovery.ServiceID, ProviderNameUpper: "ServiceDiscovery", HCLKeys: []string{"servicediscovery"}}
	serviceData[ServiceQuotas] = &ServiceDatum{AWSClientName: "ServiceQuotas", AWSServiceName: servicequotas.ServiceName, AWSEndpointsID: servicequotas.EndpointsID, AWSServiceID: servicequotas.ServiceID, ProviderNameUpper: "ServiceQuotas", HCLKeys: []string{"servicequotas"}}
	serviceData[SES] = &ServiceDatum{AWSClientName: "SES", AWSServiceName: ses.ServiceName, AWSEndpointsID: ses.EndpointsID, AWSServiceID: ses.ServiceID, ProviderNameUpper: "SES", HCLKeys: []string{"ses"}}
	serviceData[SESV2] = &ServiceDatum{AWSClientName: "SESV2", AWSServiceName: sesv2.ServiceName, AWSEndpointsID: sesv2.EndpointsID, AWSServiceID: sesv2.ServiceID, ProviderNameUpper: "SESV2", HCLKeys: []string{"sesv2"}}
	serviceData[SFN] = &ServiceDatum{AWSClientName: "SFN", AWSServiceName: sfn.ServiceName, AWSEndpointsID: sfn.EndpointsID, AWSServiceID: sfn.ServiceID, ProviderNameUpper: "SFN", HCLKeys: []string{"stepfunctions", "sfn"}}
	serviceData[Shield] = &ServiceDatum{AWSClientName: "Shield", AWSServiceName: shield.ServiceName, AWSEndpointsID: shield.EndpointsID, AWSServiceID: shield.ServiceID, ProviderNameUpper: "Shield", HCLKeys: []string{"shield"}}
	serviceData[Signer] = &ServiceDatum{AWSClientName: "Signer", AWSServiceName: signer.ServiceName, AWSEndpointsID: signer.EndpointsID, AWSServiceID: signer.ServiceID, ProviderNameUpper: "Signer", HCLKeys: []string{"signer"}}
	serviceData[SimpleDB] = &ServiceDatum{AWSClientName: "SimpleDB", AWSServiceName: simpledb.ServiceName, AWSEndpointsID: simpledb.EndpointsID, AWSServiceID: simpledb.ServiceID, ProviderNameUpper: "SimpleDB", HCLKeys: []string{"sdb", "simpledb"}}
	serviceData[SMS] = &ServiceDatum{AWSClientName: "SMS", AWSServiceName: sms.ServiceName, AWSEndpointsID: sms.EndpointsID, AWSServiceID: sms.ServiceID, ProviderNameUpper: "SMS", HCLKeys: []string{"sms"}}
	serviceData[Snowball] = &ServiceDatum{AWSClientName: "Snowball", AWSServiceName: snowball.ServiceName, AWSEndpointsID: snowball.EndpointsID, AWSServiceID: snowball.ServiceID, ProviderNameUpper: "Snowball", HCLKeys: []string{"snowball"}}
	serviceData[SNS] = &ServiceDatum{AWSClientName: "SNS", AWSServiceName: sns.ServiceName, AWSEndpointsID: sns.EndpointsID, AWSServiceID: sns.ServiceID, ProviderNameUpper: "SNS", HCLKeys: []string{"sns"}}
	serviceData[SQS] = &ServiceDatum{AWSClientName: "SQS", AWSServiceName: sqs.ServiceName, AWSEndpointsID: sqs.EndpointsID, AWSServiceID: sqs.ServiceID, ProviderNameUpper: "SQS", HCLKeys: []string{"sqs"}}
	serviceData[SSM] = &ServiceDatum{AWSClientName: "SSM", AWSServiceName: ssm.ServiceName, AWSEndpointsID: ssm.EndpointsID, AWSServiceID: ssm.ServiceID, ProviderNameUpper: "SSM", HCLKeys: []string{"ssm"}}
	serviceData[SSMContacts] = &ServiceDatum{AWSClientName: "SSMContacts", AWSServiceName: ssmcontacts.ServiceName, AWSEndpointsID: ssmcontacts.EndpointsID, AWSServiceID: ssmcontacts.ServiceID, ProviderNameUpper: "SSMContacts", HCLKeys: []string{"ssmcontacts"}}
	serviceData[SSMIncidents] = &ServiceDatum{AWSClientName: "SSMIncidents", AWSServiceName: ssmincidents.ServiceName, AWSEndpointsID: ssmincidents.EndpointsID, AWSServiceID: ssmincidents.ServiceID, ProviderNameUpper: "SSMIncidents", HCLKeys: []string{"ssmincidents"}}
	serviceData[SSO] = &ServiceDatum{AWSClientName: "SSO", AWSServiceName: sso.ServiceName, AWSEndpointsID: sso.EndpointsID, AWSServiceID: sso.ServiceID, ProviderNameUpper: "SSO", HCLKeys: []string{"sso"}}
	serviceData[SSOAdmin] = &ServiceDatum{AWSClientName: "SSOAdmin", AWSServiceName: ssoadmin.ServiceName, AWSEndpointsID: ssoadmin.EndpointsID, AWSServiceID: ssoadmin.ServiceID, ProviderNameUpper: "SSOAdmin", HCLKeys: []string{"ssoadmin"}}
	serviceData[SSOOIDC] = &ServiceDatum{AWSClientName: "SSOOIDC", AWSServiceName: ssooidc.ServiceName, AWSEndpointsID: ssooidc.EndpointsID, AWSServiceID: ssooidc.ServiceID, ProviderNameUpper: "SSOOIDC", HCLKeys: []string{"ssooidc"}}
	serviceData[StorageGateway] = &ServiceDatum{AWSClientName: "StorageGateway", AWSServiceName: storagegateway.ServiceName, AWSEndpointsID: storagegateway.EndpointsID, AWSServiceID: storagegateway.ServiceID, ProviderNameUpper: "StorageGateway", HCLKeys: []string{"storagegateway"}}
	serviceData[STS] = &ServiceDatum{AWSClientName: "STS", AWSServiceName: sts.ServiceName, AWSEndpointsID: sts.EndpointsID, AWSServiceID: sts.ServiceID, ProviderNameUpper: "STS", HCLKeys: []string{"sts"}}
	serviceData[Support] = &ServiceDatum{AWSClientName: "Support", AWSServiceName: support.ServiceName, AWSEndpointsID: support.EndpointsID, AWSServiceID: support.ServiceID, ProviderNameUpper: "Support", HCLKeys: []string{"support"}}
	serviceData[SWF] = &ServiceDatum{AWSClientName: "SWF", AWSServiceName: swf.ServiceName, AWSEndpointsID: swf.EndpointsID, AWSServiceID: swf.ServiceID, ProviderNameUpper: "SWF", HCLKeys: []string{"swf"}}
	serviceData[Synthetics] = &ServiceDatum{AWSClientName: "Synthetics", AWSServiceName: synthetics.ServiceName, AWSEndpointsID: synthetics.EndpointsID, AWSServiceID: synthetics.ServiceID, ProviderNameUpper: "Synthetics", HCLKeys: []string{"synthetics"}}
	serviceData[Textract] = &ServiceDatum{AWSClientName: "Textract", AWSServiceName: textract.ServiceName, AWSEndpointsID: textract.EndpointsID, AWSServiceID: textract.ServiceID, ProviderNameUpper: "Textract", HCLKeys: []string{"textract"}}
	serviceData[TimestreamQuery] = &ServiceDatum{AWSClientName: "TimestreamQuery", AWSServiceName: timestreamquery.ServiceName, AWSEndpointsID: timestreamquery.EndpointsID, AWSServiceID: timestreamquery.ServiceID, ProviderNameUpper: "TimestreamQuery", HCLKeys: []string{"timestreamquery"}}
	serviceData[TimestreamWrite] = &ServiceDatum{AWSClientName: "TimestreamWrite", AWSServiceName: timestreamwrite.ServiceName, AWSEndpointsID: timestreamwrite.EndpointsID, AWSServiceID: timestreamwrite.ServiceID, ProviderNameUpper: "TimestreamWrite", HCLKeys: []string{"timestreamwrite"}}
	serviceData[Transcribe] = &ServiceDatum{AWSClientName: "TranscribeService", AWSServiceName: transcribeservice.ServiceName, AWSEndpointsID: transcribeservice.EndpointsID, AWSServiceID: transcribeservice.ServiceID, ProviderNameUpper: "Transcribe", HCLKeys: []string{"transcribe", "transcribeservice"}}
	serviceData[TranscribeStreaming] = &ServiceDatum{AWSClientName: "TranscribeStreamingService", AWSServiceName: transcribestreamingservice.ServiceName, AWSEndpointsID: transcribestreamingservice.EndpointsID, AWSServiceID: transcribestreamingservice.ServiceID, ProviderNameUpper: "TranscribeStreaming", HCLKeys: []string{"transcribestreaming", "transcribestreamingservice"}}
	serviceData[Transfer] = &ServiceDatum{AWSClientName: "Transfer", AWSServiceName: transfer.ServiceName, AWSEndpointsID: transfer.EndpointsID, AWSServiceID: transfer.ServiceID, ProviderNameUpper: "Transfer", HCLKeys: []string{"transfer"}}
	serviceData[Translate] = &ServiceDatum{AWSClientName: "Translate", AWSServiceName: translate.ServiceName, AWSEndpointsID: translate.EndpointsID, AWSServiceID: translate.ServiceID, ProviderNameUpper: "Translate", HCLKeys: []string{"translate"}}
	serviceData[WAF] = &ServiceDatum{AWSClientName: "WAF", AWSServiceName: waf.ServiceName, AWSEndpointsID: waf.EndpointsID, AWSServiceID: waf.ServiceID, ProviderNameUpper: "WAF", HCLKeys: []string{"waf"}}
	serviceData[WAFRegional] = &ServiceDatum{AWSClientName: "WAFRegional", AWSServiceName: wafregional.ServiceName, AWSEndpointsID: wafregional.EndpointsID, AWSServiceID: wafregional.ServiceID, ProviderNameUpper: "WAFRegional", HCLKeys: []string{"wafregional"}}
	serviceData[WAFV2] = &ServiceDatum{AWSClientName: "WAFV2", AWSServiceName: wafv2.ServiceName, AWSEndpointsID: wafv2.EndpointsID, AWSServiceID: wafv2.ServiceID, ProviderNameUpper: "WAFV2", HCLKeys: []string{"wafv2"}}
	serviceData[WellArchitected] = &ServiceDatum{AWSClientName: "WellArchitected", AWSServiceName: wellarchitected.ServiceName, AWSEndpointsID: wellarchitected.EndpointsID, AWSServiceID: wellarchitected.ServiceID, ProviderNameUpper: "WellArchitected", HCLKeys: []string{"wellarchitected"}}
	serviceData[WorkDocs] = &ServiceDatum{AWSClientName: "WorkDocs", AWSServiceName: workdocs.ServiceName, AWSEndpointsID: workdocs.EndpointsID, AWSServiceID: workdocs.ServiceID, ProviderNameUpper: "WorkDocs", HCLKeys: []string{"workdocs"}}
	serviceData[WorkLink] = &ServiceDatum{AWSClientName: "WorkLink", AWSServiceName: worklink.ServiceName, AWSEndpointsID: worklink.EndpointsID, AWSServiceID: worklink.ServiceID, ProviderNameUpper: "WorkLink", HCLKeys: []string{"worklink"}}
	serviceData[WorkMail] = &ServiceDatum{AWSClientName: "WorkMail", AWSServiceName: workmail.ServiceName, AWSEndpointsID: workmail.EndpointsID, AWSServiceID: workmail.ServiceID, ProviderNameUpper: "WorkMail", HCLKeys: []string{"workmail"}}
	serviceData[WorkMailMessageFlow] = &ServiceDatum{AWSClientName: "WorkMailMessageFlow", AWSServiceName: workmailmessageflow.ServiceName, AWSEndpointsID: workmailmessageflow.EndpointsID, AWSServiceID: workmailmessageflow.ServiceID, ProviderNameUpper: "WorkMailMessageFlow", HCLKeys: []string{"workmailmessageflow"}}
	serviceData[WorkSpaces] = &ServiceDatum{AWSClientName: "WorkSpaces", AWSServiceName: workspaces.ServiceName, AWSEndpointsID: workspaces.EndpointsID, AWSServiceID: workspaces.ServiceID, ProviderNameUpper: "WorkSpaces", HCLKeys: []string{"workspaces"}}
	serviceData[XRay] = &ServiceDatum{AWSClientName: "XRay", AWSServiceName: xray.ServiceName, AWSEndpointsID: xray.EndpointsID, AWSServiceID: xray.ServiceID, ProviderNameUpper: "XRay", HCLKeys: []string{"xray"}}
}

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
	HTTPProxy         string

	SkipCredsValidation     bool
	SkipGetEC2Platforms     bool
	SkipRegionValidation    bool
	SkipRequestingAccountId bool
	SkipMetadataApiCheck    bool
	S3ForcePathStyle        bool

	TerraformVersion string
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
	Route53DomainsConn                *route53domains.Route53Domains
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
		IamEndpoint:                 c.Endpoints[IAM],
		Insecure:                    c.Insecure,
		HTTPProxy:                   c.HTTPProxy,
		MaxRetries:                  c.MaxRetries,
		Profile:                     c.Profile,
		Region:                      c.Region,
		SecretKey:                   c.SecretKey,
		SkipCredsValidation:         c.SkipCredsValidation,
		SkipMetadataApiCheck:        c.SkipMetadataApiCheck,
		SkipRequestingAccountId:     c.SkipRequestingAccountId,
		StsEndpoint:                 c.Endpoints[STS],
		Token:                       c.Token,
		UserAgentProducts:           StdUserAgentProducts(c.TerraformVersion),
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
		AccessAnalyzerConn:                accessanalyzer.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[AccessAnalyzer])})),
		AccountConn:                       account.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Account])})),
		AccountID:                         accountID,
		ACMConn:                           acm.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[ACM])})),
		ACMPCAConn:                        acmpca.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[ACMPCA])})),
		AlexaForBusinessConn:              alexaforbusiness.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[AlexaForBusiness])})),
		AMPConn:                           prometheusservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[AMP])})),
		AmplifyBackendConn:                amplifybackend.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[AmplifyBackend])})),
		AmplifyConn:                       amplify.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Amplify])})),
		APIGatewayConn:                    apigateway.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[APIGateway])})),
		APIGatewayV2Conn:                  apigatewayv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[APIGatewayV2])})),
		AppAutoScalingConn:                applicationautoscaling.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[AppAutoScaling])})),
		AppConfigConn:                     appconfig.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[AppConfig])})),
		AppFlowConn:                       appflow.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[AppFlow])})),
		AppIntegrationsConn:               appintegrationsservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[AppIntegrations])})),
		ApplicationCostProfilerConn:       applicationcostprofiler.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[ApplicationCostProfiler])})),
		ApplicationDiscoveryConn:          applicationdiscoveryservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[ApplicationDiscovery])})),
		ApplicationInsightsConn:           applicationinsights.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[ApplicationInsights])})),
		AppMeshConn:                       appmesh.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[AppMesh])})),
		AppRegistryConn:                   appregistry.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[AppRegistry])})),
		AppRunnerConn:                     apprunner.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[AppRunner])})),
		AppStreamConn:                     appstream.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[AppStream])})),
		AppSyncConn:                       appsync.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[AppSync])})),
		AthenaConn:                        athena.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Athena])})),
		AuditManagerConn:                  auditmanager.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[AuditManager])})),
		AugmentedAIRuntimeConn:            augmentedairuntime.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[AugmentedAIRuntime])})),
		AutoScalingConn:                   autoscaling.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[AutoScaling])})),
		AutoScalingPlansConn:              autoscalingplans.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[AutoScalingPlans])})),
		BackupConn:                        backup.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Backup])})),
		BatchConn:                         batch.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Batch])})),
		BraketConn:                        braket.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Braket])})),
		BudgetsConn:                       budgets.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Budgets])})),
		ChimeConn:                         chime.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Chime])})),
		Cloud9Conn:                        cloud9.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Cloud9])})),
		CloudControlConn:                  cloudcontrolapi.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[CloudControl])})),
		CloudDirectoryConn:                clouddirectory.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[CloudDirectory])})),
		CloudFormationConn:                cloudformation.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[CloudFormation])})),
		CloudFrontConn:                    cloudfront.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[CloudFront])})),
		CloudHSMV2Conn:                    cloudhsmv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[CloudHSMV2])})),
		CloudSearchConn:                   cloudsearch.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[CloudSearch])})),
		CloudSearchDomainConn:             cloudsearchdomain.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[CloudSearchDomain])})),
		CloudTrailConn:                    cloudtrail.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[CloudTrail])})),
		CloudWatchConn:                    cloudwatch.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[CloudWatch])})),
		CloudWatchLogsConn:                cloudwatchlogs.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[CloudWatchLogs])})),
		CodeArtifactConn:                  codeartifact.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[CodeArtifact])})),
		CodeBuildConn:                     codebuild.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[CodeBuild])})),
		CodeCommitConn:                    codecommit.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[CodeCommit])})),
		CodeDeployConn:                    codedeploy.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[CodeDeploy])})),
		CodeGuruProfilerConn:              codeguruprofiler.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[CodeGuruProfiler])})),
		CodeGuruReviewerConn:              codegurureviewer.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[CodeGuruReviewer])})),
		CodePipelineConn:                  codepipeline.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[CodePipeline])})),
		CodeStarConn:                      codestar.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[CodeStar])})),
		CodeStarConnectionsConn:           codestarconnections.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[CodeStarConnections])})),
		CodeStarNotificationsConn:         codestarnotifications.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[CodeStarNotifications])})),
		CognitoIdentityConn:               cognitoidentity.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[CognitoIdentity])})),
		CognitoIDPConn:                    cognitoidentityprovider.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[CognitoIDP])})),
		CognitoSyncConn:                   cognitosync.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[CognitoSync])})),
		ComprehendConn:                    comprehend.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Comprehend])})),
		ComprehendMedicalConn:             comprehendmedical.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[ComprehendMedical])})),
		ConfigServiceConn:                 configservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[ConfigService])})),
		ConnectConn:                       connect.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Connect])})),
		ConnectContactLensConn:            connectcontactlens.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[ConnectContactLens])})),
		ConnectParticipantConn:            connectparticipant.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[ConnectParticipant])})),
		CostExplorerConn:                  costexplorer.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[CostExplorer])})),
		CURConn:                           costandusagereportservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[CUR])})),
		DataExchangeConn:                  dataexchange.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[DataExchange])})),
		DataPipelineConn:                  datapipeline.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[DataPipeline])})),
		DataSyncConn:                      datasync.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[DataSync])})),
		DAXConn:                           dax.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[DAX])})),
		DefaultTagsConfig:                 c.DefaultTagsConfig,
		DetectiveConn:                     detective.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Detective])})),
		DeviceFarmConn:                    devicefarm.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[DeviceFarm])})),
		DevOpsGuruConn:                    devopsguru.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[DevOpsGuru])})),
		DirectConnectConn:                 directconnect.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[DirectConnect])})),
		DLMConn:                           dlm.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[DLM])})),
		DMSConn:                           databasemigrationservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[DMS])})),
		DNSSuffix:                         DNSSuffix,
		DocDBConn:                         docdb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[DocDB])})),
		DSConn:                            directoryservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[DS])})),
		DynamoDBConn:                      dynamodb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[DynamoDB])})),
		DynamoDBStreamsConn:               dynamodbstreams.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[DynamoDBStreams])})),
		EC2Conn:                           ec2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[EC2])})),
		EC2InstanceConnectConn:            ec2instanceconnect.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[EC2InstanceConnect])})),
		ECRConn:                           ecr.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[ECR])})),
		ECRPublicConn:                     ecrpublic.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[ECRPublic])})),
		ECSConn:                           ecs.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[ECS])})),
		EFSConn:                           efs.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[EFS])})),
		EKSConn:                           eks.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[EKS])})),
		ElastiCacheConn:                   elasticache.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[ElastiCache])})),
		ElasticBeanstalkConn:              elasticbeanstalk.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[ElasticBeanstalk])})),
		ElasticInferenceConn:              elasticinference.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[ElasticInference])})),
		ElasticsearchConn:                 elasticsearch.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Elasticsearch])})),
		ElasticTranscoderConn:             elastictranscoder.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[ElasticTranscoder])})),
		ELBConn:                           elb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[ELB])})),
		ELBV2Conn:                         elbv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[ELBV2])})),
		EMRConn:                           emr.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[EMR])})),
		EMRContainersConn:                 emrcontainers.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[EMRContainers])})),
		EventsConn:                        eventbridge.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Events])})),
		FinSpaceConn:                      finspace.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[FinSpace])})),
		FinSpaceDataConn:                  finspacedata.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[FinSpaceData])})),
		FirehoseConn:                      firehose.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Firehose])})),
		FISConn:                           fis.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[FIS])})),
		FMSConn:                           fms.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[FMS])})),
		ForecastConn:                      forecastservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Forecast])})),
		ForecastQueryConn:                 forecastqueryservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[ForecastQuery])})),
		FraudDetectorConn:                 frauddetector.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[FraudDetector])})),
		FSxConn:                           fsx.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[FSx])})),
		GameLiftConn:                      gamelift.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[GameLift])})),
		GlacierConn:                       glacier.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Glacier])})),
		GlueConn:                          glue.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Glue])})),
		GlueDataBrewConn:                  gluedatabrew.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[GlueDataBrew])})),
		GreengrassConn:                    greengrass.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Greengrass])})),
		GreengrassV2Conn:                  greengrassv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[GreengrassV2])})),
		GroundStationConn:                 groundstation.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[GroundStation])})),
		GuardDutyConn:                     guardduty.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[GuardDuty])})),
		HealthConn:                        health.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Health])})),
		HealthLakeConn:                    healthlake.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[HealthLake])})),
		HoneycodeConn:                     honeycode.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Honeycode])})),
		IAMConn:                           iam.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[IAM])})),
		IdentityStoreConn:                 identitystore.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[IdentityStore])})),
		IgnoreTagsConfig:                  c.IgnoreTagsConfig,
		ImageBuilderConn:                  imagebuilder.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[ImageBuilder])})),
		InspectorConn:                     inspector.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Inspector])})),
		IoT1ClickDevicesConn:              iot1clickdevicesservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[IoT1ClickDevices])})),
		IoT1ClickProjectsConn:             iot1clickprojects.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[IoT1ClickProjects])})),
		IoTAnalyticsConn:                  iotanalytics.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[IoTAnalytics])})),
		IoTConn:                           iot.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[IoT])})),
		IoTDataPlaneConn:                  iotdataplane.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[IoTDataPlane])})),
		IoTDeviceAdvisorConn:              iotdeviceadvisor.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[IoTDeviceAdvisor])})),
		IoTEventsConn:                     iotevents.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[IoTEvents])})),
		IoTEventsDataConn:                 ioteventsdata.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[IoTEventsData])})),
		IoTFleetHubConn:                   iotfleethub.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[IoTFleetHub])})),
		IoTJobsDataPlaneConn:              iotjobsdataplane.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[IoTJobsDataPlane])})),
		IoTSecureTunnelingConn:            iotsecuretunneling.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[IoTSecureTunneling])})),
		IoTSiteWiseConn:                   iotsitewise.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[IoTSiteWise])})),
		IoTThingsGraphConn:                iotthingsgraph.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[IoTThingsGraph])})),
		IoTWirelessConn:                   iotwireless.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[IoTWireless])})),
		KafkaConn:                         kafka.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Kafka])})),
		KafkaConnectConn:                  kafkaconnect.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[KafkaConnect])})),
		KendraConn:                        kendra.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Kendra])})),
		KinesisAnalyticsConn:              kinesisanalytics.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[KinesisAnalytics])})),
		KinesisAnalyticsV2Conn:            kinesisanalyticsv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[KinesisAnalyticsV2])})),
		KinesisConn:                       kinesis.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Kinesis])})),
		KinesisVideoArchivedMediaConn:     kinesisvideoarchivedmedia.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[KinesisVideoArchivedMedia])})),
		KinesisVideoConn:                  kinesisvideo.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[KinesisVideo])})),
		KinesisVideoMediaConn:             kinesisvideomedia.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[KinesisVideoMedia])})),
		KinesisVideoSignalingChannelsConn: kinesisvideosignalingchannels.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[KinesisVideoSignalingChannels])})),
		KMSConn:                           kms.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[KMS])})),
		LakeFormationConn:                 lakeformation.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[LakeFormation])})),
		LambdaConn:                        lambda.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Lambda])})),
		LexModelsConn:                     lexmodelbuildingservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[LexModels])})),
		LexModelsV2Conn:                   lexmodelsv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[LexModelsV2])})),
		LexRuntimeConn:                    lexruntimeservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[LexRuntime])})),
		LexRuntimeV2Conn:                  lexruntimev2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[LexRuntimeV2])})),
		LicenseManagerConn:                licensemanager.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[LicenseManager])})),
		LightsailConn:                     lightsail.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Lightsail])})),
		LocationConn:                      locationservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Location])})),
		LookoutEquipmentConn:              lookoutequipment.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[LookoutEquipment])})),
		LookoutForVisionConn:              lookoutforvision.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[LookoutForVision])})),
		LookoutMetricsConn:                lookoutmetrics.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[LookoutMetrics])})),
		MachineLearningConn:               machinelearning.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[MachineLearning])})),
		Macie2Conn:                        macie2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Macie2])})),
		MacieConn:                         macie.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Macie])})),
		ManagedBlockchainConn:             managedblockchain.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[ManagedBlockchain])})),
		MarketplaceCatalogConn:            marketplacecatalog.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[MarketplaceCatalog])})),
		MarketplaceCommerceAnalyticsConn:  marketplacecommerceanalytics.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[MarketplaceCommerceAnalytics])})),
		MarketplaceEntitlementConn:        marketplaceentitlementservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[MarketplaceEntitlement])})),
		MarketplaceMeteringConn:           marketplacemetering.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[MarketplaceMetering])})),
		MediaConnectConn:                  mediaconnect.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[MediaConnect])})),
		MediaConvertConn:                  mediaconvert.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[MediaConvert])})),
		MediaLiveConn:                     medialive.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[MediaLive])})),
		MediaPackageConn:                  mediapackage.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[MediaPackage])})),
		MediaPackageVODConn:               mediapackagevod.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[MediaPackageVOD])})),
		MediaStoreConn:                    mediastore.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[MediaStore])})),
		MediaStoreDataConn:                mediastoredata.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[MediaStoreData])})),
		MediaTailorConn:                   mediatailor.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[MediaTailor])})),
		MemoryDBConn:                      memorydb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[MemoryDB])})),
		MgnConn:                           mgn.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Mgn])})),
		MigrationHubConfigConn:            migrationhubconfig.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[MigrationHubConfig])})),
		MigrationHubConn:                  migrationhub.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[MigrationHub])})),
		MobileAnalyticsConn:               mobileanalytics.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[MobileAnalytics])})),
		MobileConn:                        mobile.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Mobile])})),
		MQConn:                            mq.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[MQ])})),
		MTurkConn:                         mturk.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[MTurk])})),
		MWAAConn:                          mwaa.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[MWAA])})),
		NeptuneConn:                       neptune.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Neptune])})),
		NetworkFirewallConn:               networkfirewall.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[NetworkFirewall])})),
		NetworkManagerConn:                networkmanager.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[NetworkManager])})),
		NimbleStudioConn:                  nimblestudio.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[NimbleStudio])})),
		OpsWorksCMConn:                    opsworkscm.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[OpsWorksCM])})),
		OpsWorksConn:                      opsworks.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[OpsWorks])})),
		OrganizationsConn:                 organizations.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Organizations])})),
		OutpostsConn:                      outposts.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Outposts])})),
		Partition:                         Partition,
		PersonalizeConn:                   personalize.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Personalize])})),
		PersonalizeEventsConn:             personalizeevents.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[PersonalizeEvents])})),
		PersonalizeRuntimeConn:            personalizeruntime.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[PersonalizeRuntime])})),
		PIConn:                            pi.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[PI])})),
		PinpointConn:                      pinpoint.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Pinpoint])})),
		PinpointEmailConn:                 pinpointemail.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[PinpointEmail])})),
		PinpointSMSVoiceConn:              pinpointsmsvoice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[PinpointSMSVoice])})),
		PollyConn:                         polly.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Polly])})),
		PricingConn:                       pricing.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Pricing])})),
		ProtonConn:                        proton.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Proton])})),
		QLDBConn:                          qldb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[QLDB])})),
		QLDBSessionConn:                   qldbsession.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[QLDBSession])})),
		QuickSightConn:                    quicksight.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[QuickSight])})),
		RAMConn:                           ram.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[RAM])})),
		RDSConn:                           rds.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[RDS])})),
		RDSDataConn:                       rdsdataservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[RDSData])})),
		RedshiftConn:                      redshift.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Redshift])})),
		RedshiftDataConn:                  redshiftdataapiservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[RedshiftData])})),
		Region:                            c.Region,
		RekognitionConn:                   rekognition.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Rekognition])})),
		ResourceGroupsConn:                resourcegroups.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[ResourceGroups])})),
		ResourceGroupsTaggingAPIConn:      resourcegroupstaggingapi.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[ResourceGroupsTaggingAPI])})),
		ReverseDNSPrefix:                  ReverseDNS(DNSSuffix),
		RoboMakerConn:                     robomaker.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[RoboMaker])})),
		Route53DomainsConn:                route53domains.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Route53Domains])})),
		Route53RecoveryControlConfigConn:  route53recoverycontrolconfig.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Route53RecoveryControlConfig])})),
		Route53RecoveryReadinessConn:      route53recoveryreadiness.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Route53RecoveryReadiness])})),
		Route53ResolverConn:               route53resolver.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Route53Resolver])})),
		S3ControlConn:                     s3control.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[S3Control])})),
		S3OutpostsConn:                    s3outposts.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[S3Outposts])})),
		SageMakerConn:                     sagemaker.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[SageMaker])})),
		SageMakerEdgeManagerConn:          sagemakeredgemanager.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[SageMakerEdgeManager])})),
		SageMakerFeatureStoreRuntimeConn:  sagemakerfeaturestoreruntime.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[SageMakerFeatureStoreRuntime])})),
		SageMakerRuntimeConn:              sagemakerruntime.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[SageMakerRuntime])})),
		SavingsPlansConn:                  savingsplans.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[SavingsPlans])})),
		SchemasConn:                       schemas.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Schemas])})),
		SecretsManagerConn:                secretsmanager.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[SecretsManager])})),
		SecurityHubConn:                   securityhub.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[SecurityHub])})),
		ServerlessRepoConn:                serverlessapplicationrepository.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[ServerlessRepo])})),
		ServiceCatalogConn:                servicecatalog.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[ServiceCatalog])})),
		ServiceDiscoveryConn:              servicediscovery.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[ServiceDiscovery])})),
		ServiceQuotasConn:                 servicequotas.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[ServiceQuotas])})),
		SESConn:                           ses.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[SES])})),
		SESV2Conn:                         sesv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[SESV2])})),
		SFNConn:                           sfn.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[SFN])})),
		SignerConn:                        signer.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Signer])})),
		SimpleDBConn:                      simpledb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[SimpleDB])})),
		SMSConn:                           sms.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[SMS])})),
		SnowballConn:                      snowball.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Snowball])})),
		SNSConn:                           sns.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[SNS])})),
		SQSConn:                           sqs.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[SQS])})),
		SSMConn:                           ssm.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[SSM])})),
		SSMContactsConn:                   ssmcontacts.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[SSMContacts])})),
		SSMIncidentsConn:                  ssmincidents.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[SSMIncidents])})),
		SSOAdminConn:                      ssoadmin.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[SSOAdmin])})),
		SSOConn:                           sso.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[SSO])})),
		SSOOIDCConn:                       ssooidc.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[SSOOIDC])})),
		StorageGatewayConn:                storagegateway.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[StorageGateway])})),
		STSConn:                           sts.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[STS])})),
		SupportConn:                       support.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Support])})),
		SWFConn:                           swf.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[SWF])})),
		SyntheticsConn:                    synthetics.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Synthetics])})),
		TerraformVersion:                  c.TerraformVersion,
		TextractConn:                      textract.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Textract])})),
		TimestreamQueryConn:               timestreamquery.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[TimestreamQuery])})),
		TimestreamWriteConn:               timestreamwrite.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[TimestreamWrite])})),
		TranscribeConn:                    transcribeservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Transcribe])})),
		TranscribeStreamingConn:           transcribestreamingservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[TranscribeStreaming])})),
		TransferConn:                      transfer.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Transfer])})),
		TranslateConn:                     translate.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[Translate])})),
		WAFConn:                           waf.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[WAF])})),
		WAFRegionalConn:                   wafregional.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[WAFRegional])})),
		WAFV2Conn:                         wafv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[WAFV2])})),
		WellArchitectedConn:               wellarchitected.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[WellArchitected])})),
		WorkDocsConn:                      workdocs.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[WorkDocs])})),
		WorkLinkConn:                      worklink.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[WorkLink])})),
		WorkMailConn:                      workmail.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[WorkMail])})),
		WorkMailMessageFlowConn:           workmailmessageflow.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[WorkMailMessageFlow])})),
		WorkSpacesConn:                    workspaces.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[WorkSpaces])})),
		XRayConn:                          xray.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[XRay])})),
	}

	// "Global" services that require customizations
	globalAcceleratorConfig := &aws.Config{
		Endpoint: aws.String(c.Endpoints[GlobalAccelerator]),
	}
	route53Config := &aws.Config{
		Endpoint: aws.String(c.Endpoints[Route53]),
	}
	route53RecoveryControlConfigConfig := &aws.Config{
		Endpoint: aws.String(c.Endpoints[Route53RecoveryControlConfig]),
	}
	route53RecoveryReadinessConfig := &aws.Config{
		Endpoint: aws.String(c.Endpoints[Route53RecoveryReadiness]),
	}
	shieldConfig := &aws.Config{
		Endpoint: aws.String(c.Endpoints[Shield]),
	}

	// Services that require multiple client configurations
	s3Config := &aws.Config{
		Endpoint:         aws.String(c.Endpoints[S3]),
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

func StdUserAgentProducts(terraformVersion string) []*awsbase.UserAgentProduct {
	return []*awsbase.UserAgentProduct{
		{Name: "APN", Version: "1.0"},
		{Name: "HashiCorp", Version: "1.0"},
		{Name: "Terraform", Version: terraformVersion, Extra: []string{"+https://www.terraform.io"}},
		{Name: "terraform-provider-aws", Version: version.ProviderVersion, Extra: []string{"+https://registry.terraform.io/providers/hashicorp/aws"}},
	}
}

func NewSessionForRegion(cfg *aws.Config, region, terraformVersion string) (*session.Session, error) {
	session, err := session.NewSession(cfg)

	if err != nil {
		return nil, err
	}

	userAgentProducts := StdUserAgentProducts(terraformVersion)
	// Copied from github.com/hashicorp/aws-sdk-go-base@v1.0.0/session.go:
	for i := len(userAgentProducts) - 1; i >= 0; i-- {
		product := userAgentProducts[i]
		session.Handlers.Build.PushFront(request.MakeAddToUserAgentHandler(product.Name, product.Version, product.Extra...))
	}

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

func ServiceForHCLKey(s string) (string, error) {
	for k, v := range serviceData {
		for _, hclKey := range v.HCLKeys {
			if s == hclKey {
				return k, nil
			}
		}
	}

	return "", fmt.Errorf("unable to find service for HCL key %s", s)
}

func ServiceKeys() []string {
	keys := make([]string, len(serviceData))

	i := 0
	for k := range serviceData {
		keys[i] = k
		i++
	}

	return keys
}

func HCLKeys() []string {
	keys := make([]string, 0)

	for _, v := range serviceData {
		keys = append(keys, v.HCLKeys...)
	}

	return keys
}

func ServiceProviderNameUpper(key string) (string, error) {
	if v, ok := serviceData[key]; ok {
		return v.ProviderNameUpper, nil
	}

	return "", fmt.Errorf("no service data found for %s", key)
}
