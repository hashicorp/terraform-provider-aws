package names

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/route53domains"
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
	CloudWatchRUM                 = "cloudwatchrum"
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
	Grafana                       = "grafana"
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
	Keyspaces                     = "keyspaces"
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

// These "should" be defined by the AWS Go SDK v2, but currently aren't.
const (
	Route53DomainsEndpointID  = "route53domains"
	Route53DomainsServiceName = "route53domains"
)

type ServiceDatum struct {
	AWSClientName     string
	AWSServiceName    string
	AWSEndpointsID    string
	AWSServiceID      string
	ProviderNameUpper string
	HCLKeys           []string
	EnvVar            string
	DeprecatedEnvVar  string
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
	serviceData[CloudWatchRUM] = &ServiceDatum{AWSClientName: "CloudWatchRUM", AWSServiceName: cloudwatchrum.ServiceName, AWSEndpointsID: cloudwatchrum.EndpointsID, AWSServiceID: cloudwatchrum.ServiceID, ProviderNameUpper: "CloudWatchRUM", HCLKeys: []string{"cloudwatchrum"}}
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
	serviceData[DynamoDB] = &ServiceDatum{AWSClientName: "DynamoDB", AWSServiceName: dynamodb.ServiceName, AWSEndpointsID: dynamodb.EndpointsID, AWSServiceID: dynamodb.ServiceID, ProviderNameUpper: "DynamoDB", HCLKeys: []string{"dynamodb"}, EnvVar: "TF_AWS_DYNAMODB_ENDPOINT", DeprecatedEnvVar: "AWS_DYNAMODB_ENDPOINT"}
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
	serviceData[Events] = &ServiceDatum{AWSClientName: "EventBridge", AWSServiceName: eventbridge.ServiceName, AWSEndpointsID: eventbridge.EndpointsID, AWSServiceID: eventbridge.ServiceID, ProviderNameUpper: "Events", HCLKeys: []string{"eventbridge", "cloudwatchevents", "events"}}
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
	serviceData[Grafana] = &ServiceDatum{AWSClientName: "Grafana", AWSServiceName: managedgrafana.ServiceName, AWSEndpointsID: managedgrafana.EndpointsID, AWSServiceID: managedgrafana.ServiceID, ProviderNameUpper: "Grafana", HCLKeys: []string{"grafana", "managedgrafana", "amg"}}
	serviceData[Greengrass] = &ServiceDatum{AWSClientName: "Greengrass", AWSServiceName: greengrass.ServiceName, AWSEndpointsID: greengrass.EndpointsID, AWSServiceID: greengrass.ServiceID, ProviderNameUpper: "Greengrass", HCLKeys: []string{"greengrass"}}
	serviceData[GreengrassV2] = &ServiceDatum{AWSClientName: "GreengrassV2", AWSServiceName: greengrassv2.ServiceName, AWSEndpointsID: greengrassv2.EndpointsID, AWSServiceID: greengrassv2.ServiceID, ProviderNameUpper: "GreengrassV2", HCLKeys: []string{"greengrassv2"}}
	serviceData[GroundStation] = &ServiceDatum{AWSClientName: "GroundStation", AWSServiceName: groundstation.ServiceName, AWSEndpointsID: groundstation.EndpointsID, AWSServiceID: groundstation.ServiceID, ProviderNameUpper: "GroundStation", HCLKeys: []string{"groundstation"}}
	serviceData[GuardDuty] = &ServiceDatum{AWSClientName: "GuardDuty", AWSServiceName: guardduty.ServiceName, AWSEndpointsID: guardduty.EndpointsID, AWSServiceID: guardduty.ServiceID, ProviderNameUpper: "GuardDuty", HCLKeys: []string{"guardduty"}}
	serviceData[Health] = &ServiceDatum{AWSClientName: "Health", AWSServiceName: health.ServiceName, AWSEndpointsID: health.EndpointsID, AWSServiceID: health.ServiceID, ProviderNameUpper: "Health", HCLKeys: []string{"health"}}
	serviceData[HealthLake] = &ServiceDatum{AWSClientName: "HealthLake", AWSServiceName: healthlake.ServiceName, AWSEndpointsID: healthlake.EndpointsID, AWSServiceID: healthlake.ServiceID, ProviderNameUpper: "HealthLake", HCLKeys: []string{"healthlake"}}
	serviceData[Honeycode] = &ServiceDatum{AWSClientName: "Honeycode", AWSServiceName: honeycode.ServiceName, AWSEndpointsID: honeycode.EndpointsID, AWSServiceID: honeycode.ServiceID, ProviderNameUpper: "Honeycode", HCLKeys: []string{"honeycode"}}
	serviceData[IAM] = &ServiceDatum{AWSClientName: "IAM", AWSServiceName: iam.ServiceName, AWSEndpointsID: iam.EndpointsID, AWSServiceID: iam.ServiceID, ProviderNameUpper: "IAM", HCLKeys: []string{"iam"}, EnvVar: "TF_AWS_IAM_ENDPOINT", DeprecatedEnvVar: "AWS_IAM_ENDPOINT"}
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
	serviceData[Keyspaces] = &ServiceDatum{AWSClientName: "Keyspaces", AWSServiceName: keyspaces.ServiceName, AWSEndpointsID: keyspaces.EndpointsID, AWSServiceID: keyspaces.ServiceID, ProviderNameUpper: "Keyspaces", HCLKeys: []string{"keyspaces"}}
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
	serviceData[Route53Domains] = &ServiceDatum{AWSClientName: "Route53Domains", AWSServiceName: Route53DomainsServiceName, AWSEndpointsID: Route53DomainsEndpointID, AWSServiceID: route53domains.ServiceID, ProviderNameUpper: "Route53Domains", HCLKeys: []string{"route53domains"}}
	serviceData[Route53RecoveryControlConfig] = &ServiceDatum{AWSClientName: "Route53RecoveryControlConfig", AWSServiceName: route53recoverycontrolconfig.ServiceName, AWSEndpointsID: route53recoverycontrolconfig.EndpointsID, AWSServiceID: route53recoverycontrolconfig.ServiceID, ProviderNameUpper: "Route53RecoveryControlConfig", HCLKeys: []string{"route53recoverycontrolconfig"}}
	serviceData[Route53RecoveryReadiness] = &ServiceDatum{AWSClientName: "Route53RecoveryReadiness", AWSServiceName: route53recoveryreadiness.ServiceName, AWSEndpointsID: route53recoveryreadiness.EndpointsID, AWSServiceID: route53recoveryreadiness.ServiceID, ProviderNameUpper: "Route53RecoveryReadiness", HCLKeys: []string{"route53recoveryreadiness"}}
	serviceData[Route53Resolver] = &ServiceDatum{AWSClientName: "Route53Resolver", AWSServiceName: route53resolver.ServiceName, AWSEndpointsID: route53resolver.EndpointsID, AWSServiceID: route53resolver.ServiceID, ProviderNameUpper: "Route53Resolver", HCLKeys: []string{"route53resolver"}}
	serviceData[S3] = &ServiceDatum{AWSClientName: "S3", AWSServiceName: s3.ServiceName, AWSEndpointsID: s3.EndpointsID, AWSServiceID: s3.ServiceID, ProviderNameUpper: "S3", HCLKeys: []string{"s3"}, EnvVar: "TF_AWS_S3_ENDPOINT", DeprecatedEnvVar: "AWS_S3_ENDPOINT"}
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
	serviceData[STS] = &ServiceDatum{AWSClientName: "STS", AWSServiceName: sts.ServiceName, AWSEndpointsID: sts.EndpointsID, AWSServiceID: sts.ServiceID, ProviderNameUpper: "STS", HCLKeys: []string{"sts"}, EnvVar: "TF_AWS_STS_ENDPOINT", DeprecatedEnvVar: "AWS_STS_ENDPOINT"}
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

func ServiceDeprecatedEnvVar(key string) string {
	if v, ok := serviceData[key]; ok {
		return v.DeprecatedEnvVar
	}

	return ""
}

func ServiceEnvVar(key string) string {
	if v, ok := serviceData[key]; ok {
		return v.EnvVar
	}

	return ""
}

func AWSServiceName(key string) (string, error) {
	if v, ok := serviceData[key]; ok {
		return v.AWSServiceName, nil
	}

	return "", fmt.Errorf("no service data found for %s", key)
}
