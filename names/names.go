package names

import (
	_ "embed"
	"encoding/csv"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/route53domains"
	"github.com/aws/aws-sdk-go/service/accessanalyzer"
	"github.com/aws/aws-sdk-go/service/account"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/aws/aws-sdk-go/service/alexaforbusiness"
	"github.com/aws/aws-sdk-go/service/amplify"
	"github.com/aws/aws-sdk-go/service/amplifybackend"
	"github.com/aws/aws-sdk-go/service/amplifyuibuilder"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/aws/aws-sdk-go/service/appconfigdata"
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
	"github.com/aws/aws-sdk-go/service/backupgateway"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/aws/aws-sdk-go/service/billingconductor"
	"github.com/aws/aws-sdk-go/service/braket"
	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/aws/aws-sdk-go/service/chime"
	"github.com/aws/aws-sdk-go/service/chimesdkidentity"
	"github.com/aws/aws-sdk-go/service/chimesdkmeetings"
	"github.com/aws/aws-sdk-go/service/chimesdkmessaging"
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
	"github.com/aws/aws-sdk-go/service/computeoptimizer"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/aws/aws-sdk-go/service/connectcontactlens"
	"github.com/aws/aws-sdk-go/service/connectparticipant"
	"github.com/aws/aws-sdk-go/service/connectwisdomservice"
	"github.com/aws/aws-sdk-go/service/costandusagereportservice"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/aws/aws-sdk-go/service/customerprofiles"
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
	"github.com/aws/aws-sdk-go/service/drs"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodbstreams"
	"github.com/aws/aws-sdk-go/service/ebs"
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
	"github.com/aws/aws-sdk-go/service/elasticsearchservice"
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
	"github.com/aws/aws-sdk-go/service/inspector2"
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
	"github.com/aws/aws-sdk-go/service/iottwinmaker"
	"github.com/aws/aws-sdk-go/service/iotwireless"
	"github.com/aws/aws-sdk-go/service/ivs"
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
	"github.com/aws/aws-sdk-go/service/migrationhubrefactorspaces"
	"github.com/aws/aws-sdk-go/service/migrationhubstrategyrecommendations"
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
	"github.com/aws/aws-sdk-go/service/panorama"
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
	"github.com/aws/aws-sdk-go/service/recyclebin"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/aws/aws-sdk-go/service/redshiftdataapiservice"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/aws/aws-sdk-go/service/resiliencehub"
	"github.com/aws/aws-sdk-go/service/resourcegroups"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go/service/robomaker"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53recoverycluster"
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
	"github.com/aws/aws-sdk-go/service/snowdevicemanagement"
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
	"github.com/aws/aws-sdk-go/service/voiceid"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/aws/aws-sdk-go/service/wellarchitected"
	"github.com/aws/aws-sdk-go/service/workdocs"
	"github.com/aws/aws-sdk-go/service/worklink"
	"github.com/aws/aws-sdk-go/service/workmail"
	"github.com/aws/aws-sdk-go/service/workmailmessageflow"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/aws/aws-sdk-go/service/workspacesweb"
	"github.com/aws/aws-sdk-go/service/xray"
)

const (
	// The value should exactly correspond to the TF AWS provider service
	// package name and providerPackageBoth column in CSV:
	// * serviceData is populated using value as key
	// * generators may use this value as package
	//
	// E.g., if internal/service/amp, then -> AMP = "amp"
	//
	// The service package name is the service identifier. Use this doc to
	// determine it:
	// https://github.com/hashicorp/terraform-provider-aws/blob/main/docs/contributing/naming.md#service-identifier
	AccessAnalyzer               = "accessanalyzer"
	Account                      = "account"
	ACM                          = "acm"
	ACMPCA                       = "acmpca"
	AlexaForBusiness             = "alexaforbusiness"
	AMP                          = "amp"
	Amplify                      = "amplify"
	AmplifyBackend               = "amplifybackend"
	AmplifyUIBuilder             = "amplifyuibuilder"
	APIGateway                   = "apigateway"
	APIGatewayManagementAPI      = "apigatewaymanagementapi"
	APIGatewayV2                 = "apigatewayv2"
	AppAutoScaling               = "appautoscaling"
	AppConfig                    = "appconfig"
	AppConfigData                = "appconfigdata"
	AppFlow                      = "appflow"
	AppIntegrations              = "appintegrations"
	ApplicationCostProfiler      = "applicationcostprofiler"
	ApplicationInsights          = "applicationinsights"
	AppMesh                      = "appmesh"
	AppRunner                    = "apprunner"
	AppStream                    = "appstream"
	AppSync                      = "appsync"
	Athena                       = "athena"
	AuditManager                 = "auditmanager"
	AutoScaling                  = "autoscaling"
	AutoScalingPlans             = "autoscalingplans"
	Backup                       = "backup"
	BackupGateway                = "backupgateway"
	Batch                        = "batch"
	BillingConductor             = "billingconductor"
	Braket                       = "braket"
	Budgets                      = "budgets"
	CE                           = "ce"
	Chime                        = "chime"
	ChimeSDKIdentity             = "chimesdkidentity"
	ChimeSDKMeetings             = "chimesdkmeetings"
	ChimeSDKMessaging            = "chimesdkmessaging"
	Cloud9                       = "cloud9"
	CloudControl                 = "cloudcontrol"
	CloudDirectory               = "clouddirectory"
	CloudFormation               = "cloudformation"
	CloudFront                   = "cloudfront"
	CloudHSMV2                   = "cloudhsmv2"
	CloudSearch                  = "cloudsearch"
	CloudSearchDomain            = "cloudsearchdomain"
	CloudTrail                   = "cloudtrail"
	CloudWatch                   = "cloudwatch"
	CloudWatchLogs               = "cloudwatchlogs"
	CodeArtifact                 = "codeartifact"
	CodeBuild                    = "codebuild"
	CodeCommit                   = "codecommit"
	CodeDeploy                   = "codedeploy"
	CodeGuruProfiler             = "codeguruprofiler"
	CodeGuruReviewer             = "codegurureviewer"
	CodePipeline                 = "codepipeline"
	CodeStar                     = "codestar"
	CodeStarConnections          = "codestarconnections"
	CodeStarNotifications        = "codestarnotifications"
	CognitoIdentity              = "cognitoidentity"
	CognitoIDP                   = "cognitoidp"
	CognitoSync                  = "cognitosync"
	Comprehend                   = "comprehend"
	ComprehendMedical            = "comprehendmedical"
	ComputeOptimizer             = "computeoptimizer"
	ConfigService                = "configservice"
	Connect                      = "connect"
	ConnectContactLens           = "connectcontactlens"
	ConnectParticipant           = "connectparticipant"
	CUR                          = "cur"
	CustomerProfiles             = "customerprofiles"
	DataBrew                     = "databrew"
	DataExchange                 = "dataexchange"
	DataPipeline                 = "datapipeline"
	DataSync                     = "datasync"
	DAX                          = "dax"
	Detective                    = "detective"
	DeviceFarm                   = "devicefarm"
	DevOpsGuru                   = "devopsguru"
	DirectConnect                = "directconnect"
	Discovery                    = "discovery"
	DLM                          = "dlm"
	DMS                          = "dms"
	DocDB                        = "docdb"
	DRS                          = "drs"
	DS                           = "ds"
	DynamoDB                     = "dynamodb"
	DynamoDBStreams              = "dynamodbstreams"
	EBS                          = "ebs"
	EC2                          = "ec2"
	EC2InstanceConnect           = "ec2instanceconnect"
	ECR                          = "ecr"
	ECRPublic                    = "ecrpublic"
	ECS                          = "ecs"
	EFS                          = "efs"
	EKS                          = "eks"
	ElastiCache                  = "elasticache"
	ElasticBeanstalk             = "elasticbeanstalk"
	ElasticInference             = "elasticinference"
	Elasticsearch                = "elasticsearch"
	ElasticTranscoder            = "elastictranscoder"
	ELB                          = "elb"
	ELBV2                        = "elbv2"
	EMR                          = "emr"
	EMRContainers                = "emrcontainers"
	Events                       = "events"
	Evidently                    = "evidently"
	FinSpace                     = "finspace"
	FinSpaceData                 = "finspacedata"
	Firehose                     = "firehose"
	FIS                          = "fis"
	FMS                          = "fms"
	Forecast                     = "forecast"
	ForecastQuery                = "forecastquery"
	FraudDetector                = "frauddetector"
	FSx                          = "fsx"
	GameLift                     = "gamelift"
	Glacier                      = "glacier"
	GlobalAccelerator            = "globalaccelerator"
	Glue                         = "glue"
	Grafana                      = "grafana"
	Greengrass                   = "greengrass"
	GreengrassV2                 = "greengrassv2"
	GroundStation                = "groundstation"
	GuardDuty                    = "guardduty"
	Health                       = "health"
	HealthLake                   = "healthlake"
	Honeycode                    = "honeycode"
	IAM                          = "iam"
	IdentityStore                = "identitystore"
	ImageBuilder                 = "imagebuilder"
	Inspector                    = "inspector"
	Inspector2                   = "inspector2"
	IoT                          = "iot"
	IoT1ClickDevices             = "iot1clickdevices"
	IoT1ClickProjects            = "iot1clickprojects"
	IoTAnalytics                 = "iotanalytics"
	IoTData                      = "iotdata"
	IoTDeviceAdvisor             = "iotdeviceadvisor"
	IoTEvents                    = "iotevents"
	IoTEventsData                = "ioteventsdata"
	IoTFleetHub                  = "iotfleethub"
	IoTJobsData                  = "iotjobsdata"
	IoTSecureTunneling           = "iotsecuretunneling"
	IoTSiteWise                  = "iotsitewise"
	IoTThingsGraph               = "iotthingsgraph"
	IoTTwinMaker                 = "iottwinmaker"
	IoTWireless                  = "iotwireless"
	IVS                          = "ivs"
	Kafka                        = "kafka"
	KafkaConnect                 = "kafkaconnect"
	Kendra                       = "kendra"
	Keyspaces                    = "keyspaces"
	Kinesis                      = "kinesis"
	KinesisAnalytics             = "kinesisanalytics"
	KinesisAnalyticsV2           = "kinesisanalyticsv2"
	KinesisVideo                 = "kinesisvideo"
	KinesisVideoArchivedMedia    = "kinesisvideoarchivedmedia"
	KinesisVideoMedia            = "kinesisvideomedia"
	KinesisVideoSignaling        = "kinesisvideosignaling"
	KMS                          = "kms"
	LakeFormation                = "lakeformation"
	Lambda                       = "lambda"
	LexModels                    = "lexmodels"
	LexModelsV2                  = "lexmodelsv2"
	LexRuntime                   = "lexruntime"
	LexRuntimeV2                 = "lexruntimev2"
	LicenseManager               = "licensemanager"
	Lightsail                    = "lightsail"
	Location                     = "location"
	LookoutEquipment             = "lookoutequipment"
	LookoutMetrics               = "lookoutmetrics"
	LookoutVision                = "lookoutvision"
	MachineLearning              = "machinelearning"
	Macie                        = "macie"
	Macie2                       = "macie2"
	ManagedBlockchain            = "managedblockchain"
	MarketplaceCatalog           = "marketplacecatalog"
	MarketplaceCommerceAnalytics = "marketplacecommerceanalytics"
	MarketplaceEntitlement       = "marketplaceentitlement"
	MarketplaceMetering          = "marketplacemetering"
	MediaConnect                 = "mediaconnect"
	MediaConvert                 = "mediaconvert"
	MediaLive                    = "medialive"
	MediaPackage                 = "mediapackage"
	MediaPackageVOD              = "mediapackagevod"
	MediaStore                   = "mediastore"
	MediaStoreData               = "mediastoredata"
	MediaTailor                  = "mediatailor"
	MemoryDB                     = "memorydb"
	MgH                          = "mgh"
	Mgn                          = "mgn"
	MigrationHubConfig           = "migrationhubconfig"
	MigrationHubRefactorSpaces   = "migrationhubrefactorspaces"
	MigrationHubStrategy         = "migrationhubstrategy"
	Mobile                       = "mobile"
	MobileAnalytics              = "mobileanalytics"
	MQ                           = "mq"
	MTurk                        = "mturk"
	MWAA                         = "mwaa"
	Neptune                      = "neptune"
	NetworkFirewall              = "networkfirewall"
	NetworkManager               = "networkmanager"
	Nimble                       = "nimble"
	OpenSearch                   = "opensearch"
	OpsWorks                     = "opsworks"
	OpsWorksCM                   = "opsworkscm"
	Organizations                = "organizations"
	Outposts                     = "outposts"
	Panorama                     = "panorama"
	Personalize                  = "personalize"
	PersonalizeEvents            = "personalizeevents"
	PersonalizeRuntime           = "personalizeruntime"
	PI                           = "pi"
	Pinpoint                     = "pinpoint"
	PinpointEmail                = "pinpointemail"
	PinpointSMSVoice             = "pinpointsmsvoice"
	Polly                        = "polly"
	Pricing                      = "pricing"
	Proton                       = "proton"
	QLDB                         = "qldb"
	QLDBSession                  = "qldbsession"
	QuickSight                   = "quicksight"
	RAM                          = "ram"
	RBin                         = "rbin"
	RDS                          = "rds"
	RDSData                      = "rdsdata"
	Redshift                     = "redshift"
	RedshiftData                 = "redshiftdata"
	Rekognition                  = "rekognition"
	ResilienceHub                = "resiliencehub"
	ResourceGroups               = "resourcegroups"
	ResourceGroupsTaggingAPI     = "resourcegroupstaggingapi"
	RoboMaker                    = "robomaker"
	Route53                      = "route53"
	Route53Domains               = "route53domains"
	Route53RecoveryCluster       = "route53recoverycluster"
	Route53RecoveryControlConfig = "route53recoverycontrolconfig"
	Route53RecoveryReadiness     = "route53recoveryreadiness"
	Route53Resolver              = "route53resolver"
	RUM                          = "rum"
	S3                           = "s3"
	S3Control                    = "s3control"
	S3Outposts                   = "s3outposts"
	SageMaker                    = "sagemaker"
	SageMakerA2IRuntime          = "sagemakera2iruntime"
	SageMakerEdge                = "sagemakeredge"
	SageMakerFeatureStoreRuntime = "sagemakerfeaturestoreruntime"
	SageMakerRuntime             = "sagemakerruntime"
	SavingsPlans                 = "savingsplans"
	Schemas                      = "schemas"
	SecretsManager               = "secretsmanager"
	SecurityHub                  = "securityhub"
	ServerlessRepo               = "serverlessrepo"
	ServiceCatalog               = "servicecatalog"
	ServiceCatalogAppRegistry    = "servicecatalogappregistry"
	ServiceDiscovery             = "servicediscovery"
	ServiceQuotas                = "servicequotas"
	SES                          = "ses"
	SESV2                        = "sesv2"
	SFN                          = "sfn"
	Shield                       = "shield"
	Signer                       = "signer"
	SimpleDB                     = "simpledb"
	SMS                          = "sms"
	Snowball                     = "snowball"
	SnowDeviceManagement         = "snowdevicemanagement"
	SNS                          = "sns"
	SQS                          = "sqs"
	SSM                          = "ssm"
	SSMContacts                  = "ssmcontacts"
	SSMIncidents                 = "ssmincidents"
	SSO                          = "sso"
	SSOAdmin                     = "ssoadmin"
	SSOOIDC                      = "ssooidc"
	StorageGateway               = "storagegateway"
	STS                          = "sts"
	Support                      = "support"
	SWF                          = "swf"
	Synthetics                   = "synthetics"
	Textract                     = "textract"
	TimestreamQuery              = "timestreamquery"
	TimestreamWrite              = "timestreamwrite"
	Transcribe                   = "transcribe"
	TranscribeStreaming          = "transcribestreaming"
	Transfer                     = "transfer"
	Translate                    = "translate"
	VoiceID                      = "voiceid"
	WAF                          = "waf"
	WAFRegional                  = "wafregional"
	WAFV2                        = "wafv2"
	WellArchitected              = "wellarchitected"
	Wisdom                       = "wisdom"
	WorkDocs                     = "workdocs"
	WorkLink                     = "worklink"
	WorkMail                     = "workmail"
	WorkMailMessageFlow          = "workmailmessageflow"
	WorkSpaces                   = "workspaces"
	WorkSpacesWeb                = "workspacesweb"
	XRay                         = "xray"
)

// These "should" be defined by the AWS Go SDK v2, but currently aren't.
const (
	Route53DomainsEndpointID  = "route53domains"
	Route53DomainsServiceName = "route53domains"
)

type ServiceDatum struct {
	Aliases           []string // from csv - *
	Brand             string   // from csv
	DeprecatedEnvVar  string   // from csv - *
	EndpointsID       string   // from AWS const
	EnvVar            string   // from csv - *
	GoV1ClientName    string   // from csv - * Client name in AWS Go SDK v1
	GoV1Package       string   // from csv - * AWS Go SDK v1 package name
	GoV2Package       string   // from csv - * AWS Go SDK v2 package name
	HumanFriendly     string   // from csv
	ProviderNameUpper string   // from csv - * Conn name in internal/conns/conns.go
	ServiceID         string   // from AWS const
	ServiceName       string   // from AWS const

	// * These values must be exactly correct because they are used by code,
	//   generators, and scaffolding tool.
}

// serviceData key is the AWS provider service package
var serviceData map[string]*ServiceDatum

func init() {
	serviceData = make(map[string]*ServiceDatum)

	// Service data is built in 2 steps:
	// 1. Constants copied from AWS Go SDK v1
	populateServiceDataInitial()

	// 2. Data from names_data.csv
	if err := readCSVIntoServiceData(); err != nil {
		log.Fatalf("reading CSV into service data: %s", err)
	}
}

func populateServiceDataInitial() {
	// Constants copied from AWS Go SDK v1
	serviceData[AccessAnalyzer] = &ServiceDatum{EndpointsID: accessanalyzer.EndpointsID, ServiceID: accessanalyzer.ServiceID, ServiceName: accessanalyzer.ServiceName}
	serviceData[Account] = &ServiceDatum{EndpointsID: account.EndpointsID, ServiceID: account.ServiceID, ServiceName: account.ServiceName}
	serviceData[ACM] = &ServiceDatum{EndpointsID: acm.EndpointsID, ServiceID: acm.ServiceID, ServiceName: acm.ServiceName}
	serviceData[ACMPCA] = &ServiceDatum{EndpointsID: acmpca.EndpointsID, ServiceID: acmpca.ServiceID, ServiceName: acmpca.ServiceName}
	serviceData[AlexaForBusiness] = &ServiceDatum{EndpointsID: alexaforbusiness.EndpointsID, ServiceID: alexaforbusiness.ServiceID, ServiceName: alexaforbusiness.ServiceName}
	serviceData[AMP] = &ServiceDatum{EndpointsID: prometheusservice.EndpointsID, ServiceID: prometheusservice.ServiceID, ServiceName: prometheusservice.ServiceName}
	serviceData[Amplify] = &ServiceDatum{EndpointsID: amplify.EndpointsID, ServiceID: amplify.ServiceID, ServiceName: amplify.ServiceName}
	serviceData[AmplifyBackend] = &ServiceDatum{EndpointsID: amplifybackend.EndpointsID, ServiceID: amplifybackend.ServiceID, ServiceName: amplifybackend.ServiceName}
	serviceData[AmplifyUIBuilder] = &ServiceDatum{EndpointsID: amplifyuibuilder.EndpointsID, ServiceID: amplifyuibuilder.ServiceID, ServiceName: amplifyuibuilder.ServiceName}
	serviceData[APIGateway] = &ServiceDatum{EndpointsID: apigateway.EndpointsID, ServiceID: apigateway.ServiceID, ServiceName: apigateway.ServiceName}
	serviceData[APIGatewayManagementAPI] = &ServiceDatum{EndpointsID: apigatewaymanagementapi.EndpointsID, ServiceID: apigatewaymanagementapi.ServiceID, ServiceName: apigatewaymanagementapi.ServiceName}
	serviceData[APIGatewayV2] = &ServiceDatum{EndpointsID: apigatewayv2.EndpointsID, ServiceID: apigatewayv2.ServiceID, ServiceName: apigatewayv2.ServiceName}
	serviceData[AppAutoScaling] = &ServiceDatum{EndpointsID: applicationautoscaling.EndpointsID, ServiceID: applicationautoscaling.ServiceID, ServiceName: applicationautoscaling.ServiceName}
	serviceData[AppConfig] = &ServiceDatum{EndpointsID: appconfig.EndpointsID, ServiceID: appconfig.ServiceID, ServiceName: appconfig.ServiceName}
	serviceData[AppConfigData] = &ServiceDatum{EndpointsID: appconfigdata.EndpointsID, ServiceID: appconfigdata.ServiceID, ServiceName: appconfigdata.ServiceName}
	serviceData[AppFlow] = &ServiceDatum{EndpointsID: appflow.EndpointsID, ServiceID: appflow.ServiceID, ServiceName: appflow.ServiceName}
	serviceData[AppIntegrations] = &ServiceDatum{EndpointsID: appintegrationsservice.EndpointsID, ServiceID: appintegrationsservice.ServiceID, ServiceName: appintegrationsservice.ServiceName}
	serviceData[ApplicationCostProfiler] = &ServiceDatum{EndpointsID: applicationcostprofiler.EndpointsID, ServiceID: applicationcostprofiler.ServiceID, ServiceName: applicationcostprofiler.ServiceName}
	serviceData[ApplicationInsights] = &ServiceDatum{EndpointsID: applicationinsights.EndpointsID, ServiceID: applicationinsights.ServiceID, ServiceName: applicationinsights.ServiceName}
	serviceData[AppMesh] = &ServiceDatum{EndpointsID: appmesh.EndpointsID, ServiceID: appmesh.ServiceID, ServiceName: appmesh.ServiceName}
	serviceData[AppRunner] = &ServiceDatum{EndpointsID: apprunner.EndpointsID, ServiceID: apprunner.ServiceID, ServiceName: apprunner.ServiceName}
	serviceData[AppStream] = &ServiceDatum{EndpointsID: appstream.EndpointsID, ServiceID: appstream.ServiceID, ServiceName: appstream.ServiceName}
	serviceData[AppSync] = &ServiceDatum{EndpointsID: appsync.EndpointsID, ServiceID: appsync.ServiceID, ServiceName: appsync.ServiceName}
	serviceData[Athena] = &ServiceDatum{EndpointsID: athena.EndpointsID, ServiceID: athena.ServiceID, ServiceName: athena.ServiceName}
	serviceData[AuditManager] = &ServiceDatum{EndpointsID: auditmanager.EndpointsID, ServiceID: auditmanager.ServiceID, ServiceName: auditmanager.ServiceName}
	serviceData[AutoScaling] = &ServiceDatum{EndpointsID: autoscaling.EndpointsID, ServiceID: autoscaling.ServiceID, ServiceName: autoscaling.ServiceName}
	serviceData[AutoScalingPlans] = &ServiceDatum{EndpointsID: autoscalingplans.EndpointsID, ServiceID: autoscalingplans.ServiceID, ServiceName: autoscalingplans.ServiceName}
	serviceData[Backup] = &ServiceDatum{EndpointsID: backup.EndpointsID, ServiceID: backup.ServiceID, ServiceName: backup.ServiceName}
	serviceData[BackupGateway] = &ServiceDatum{EndpointsID: backupgateway.EndpointsID, ServiceID: backupgateway.ServiceID, ServiceName: backupgateway.ServiceName}
	serviceData[Batch] = &ServiceDatum{EndpointsID: batch.EndpointsID, ServiceID: batch.ServiceID, ServiceName: batch.ServiceName}
	serviceData[BillingConductor] = &ServiceDatum{EndpointsID: billingconductor.EndpointsID, ServiceID: billingconductor.ServiceID, ServiceName: billingconductor.ServiceName}
	serviceData[Braket] = &ServiceDatum{EndpointsID: braket.EndpointsID, ServiceID: braket.ServiceID, ServiceName: braket.ServiceName}
	serviceData[Budgets] = &ServiceDatum{EndpointsID: budgets.EndpointsID, ServiceID: budgets.ServiceID, ServiceName: budgets.ServiceName}
	serviceData[CE] = &ServiceDatum{EndpointsID: costexplorer.EndpointsID, ServiceID: costexplorer.ServiceID, ServiceName: costexplorer.ServiceName}
	serviceData[Chime] = &ServiceDatum{EndpointsID: chime.EndpointsID, ServiceID: chime.ServiceID, ServiceName: chime.ServiceName}
	serviceData[ChimeSDKIdentity] = &ServiceDatum{EndpointsID: chimesdkidentity.EndpointsID, ServiceID: chimesdkidentity.ServiceID, ServiceName: chimesdkidentity.ServiceName}
	serviceData[ChimeSDKMeetings] = &ServiceDatum{EndpointsID: chimesdkmeetings.EndpointsID, ServiceID: chimesdkmeetings.ServiceID, ServiceName: chimesdkmeetings.ServiceName}
	serviceData[ChimeSDKMessaging] = &ServiceDatum{EndpointsID: chimesdkmessaging.EndpointsID, ServiceID: chimesdkmessaging.ServiceID, ServiceName: chimesdkmessaging.ServiceName}
	serviceData[Cloud9] = &ServiceDatum{EndpointsID: cloud9.EndpointsID, ServiceID: cloud9.ServiceID, ServiceName: cloud9.ServiceName}
	serviceData[CloudControl] = &ServiceDatum{EndpointsID: cloudcontrolapi.EndpointsID, ServiceID: cloudcontrolapi.ServiceID, ServiceName: cloudcontrolapi.ServiceName}
	serviceData[CloudDirectory] = &ServiceDatum{EndpointsID: clouddirectory.EndpointsID, ServiceID: clouddirectory.ServiceID, ServiceName: clouddirectory.ServiceName}
	serviceData[CloudFormation] = &ServiceDatum{EndpointsID: cloudformation.EndpointsID, ServiceID: cloudformation.ServiceID, ServiceName: cloudformation.ServiceName}
	serviceData[CloudFront] = &ServiceDatum{EndpointsID: cloudfront.EndpointsID, ServiceID: cloudfront.ServiceID, ServiceName: cloudfront.ServiceName}
	serviceData[CloudHSMV2] = &ServiceDatum{EndpointsID: cloudhsmv2.EndpointsID, ServiceID: cloudhsmv2.ServiceID, ServiceName: cloudhsmv2.ServiceName}
	serviceData[CloudSearch] = &ServiceDatum{EndpointsID: cloudsearch.EndpointsID, ServiceID: cloudsearch.ServiceID, ServiceName: cloudsearch.ServiceName}
	serviceData[CloudSearchDomain] = &ServiceDatum{EndpointsID: cloudsearchdomain.EndpointsID, ServiceID: cloudsearchdomain.ServiceID, ServiceName: cloudsearchdomain.ServiceName}
	serviceData[CloudTrail] = &ServiceDatum{EndpointsID: cloudtrail.EndpointsID, ServiceID: cloudtrail.ServiceID, ServiceName: cloudtrail.ServiceName}
	serviceData[CloudWatch] = &ServiceDatum{EndpointsID: cloudwatch.EndpointsID, ServiceID: cloudwatch.ServiceID, ServiceName: cloudwatch.ServiceName}
	serviceData[CloudWatchLogs] = &ServiceDatum{EndpointsID: cloudwatchlogs.EndpointsID, ServiceID: cloudwatchlogs.ServiceID, ServiceName: cloudwatchlogs.ServiceName}
	serviceData[CodeArtifact] = &ServiceDatum{EndpointsID: codeartifact.EndpointsID, ServiceID: codeartifact.ServiceID, ServiceName: codeartifact.ServiceName}
	serviceData[CodeBuild] = &ServiceDatum{EndpointsID: codebuild.EndpointsID, ServiceID: codebuild.ServiceID, ServiceName: codebuild.ServiceName}
	serviceData[CodeCommit] = &ServiceDatum{EndpointsID: codecommit.EndpointsID, ServiceID: codecommit.ServiceID, ServiceName: codecommit.ServiceName}
	serviceData[CodeDeploy] = &ServiceDatum{EndpointsID: codedeploy.EndpointsID, ServiceID: codedeploy.ServiceID, ServiceName: codedeploy.ServiceName}
	serviceData[CodeGuruProfiler] = &ServiceDatum{EndpointsID: codeguruprofiler.EndpointsID, ServiceID: codeguruprofiler.ServiceID, ServiceName: codeguruprofiler.ServiceName}
	serviceData[CodeGuruReviewer] = &ServiceDatum{EndpointsID: codegurureviewer.EndpointsID, ServiceID: codegurureviewer.ServiceID, ServiceName: codegurureviewer.ServiceName}
	serviceData[CodePipeline] = &ServiceDatum{EndpointsID: codepipeline.EndpointsID, ServiceID: codepipeline.ServiceID, ServiceName: codepipeline.ServiceName}
	serviceData[CodeStar] = &ServiceDatum{EndpointsID: codestar.EndpointsID, ServiceID: codestar.ServiceID, ServiceName: codestar.ServiceName}
	serviceData[CodeStarConnections] = &ServiceDatum{EndpointsID: codestarconnections.EndpointsID, ServiceID: codestarconnections.ServiceID, ServiceName: codestarconnections.ServiceName}
	serviceData[CodeStarNotifications] = &ServiceDatum{EndpointsID: codestarnotifications.EndpointsID, ServiceID: codestarnotifications.ServiceID, ServiceName: codestarnotifications.ServiceName}
	serviceData[CognitoIdentity] = &ServiceDatum{EndpointsID: cognitoidentity.EndpointsID, ServiceID: cognitoidentity.ServiceID, ServiceName: cognitoidentity.ServiceName}
	serviceData[CognitoIDP] = &ServiceDatum{EndpointsID: cognitoidentityprovider.EndpointsID, ServiceID: cognitoidentityprovider.ServiceID, ServiceName: cognitoidentityprovider.ServiceName}
	serviceData[CognitoSync] = &ServiceDatum{EndpointsID: cognitosync.EndpointsID, ServiceID: cognitosync.ServiceID, ServiceName: cognitosync.ServiceName}
	serviceData[Comprehend] = &ServiceDatum{EndpointsID: comprehend.EndpointsID, ServiceID: comprehend.ServiceID, ServiceName: comprehend.ServiceName}
	serviceData[ComprehendMedical] = &ServiceDatum{EndpointsID: comprehendmedical.EndpointsID, ServiceID: comprehendmedical.ServiceID, ServiceName: comprehendmedical.ServiceName}
	serviceData[ComputeOptimizer] = &ServiceDatum{EndpointsID: computeoptimizer.EndpointsID, ServiceID: computeoptimizer.ServiceID, ServiceName: computeoptimizer.ServiceName}
	serviceData[ConfigService] = &ServiceDatum{EndpointsID: configservice.EndpointsID, ServiceID: configservice.ServiceID, ServiceName: configservice.ServiceName}
	serviceData[Connect] = &ServiceDatum{EndpointsID: connect.EndpointsID, ServiceID: connect.ServiceID, ServiceName: connect.ServiceName}
	serviceData[ConnectContactLens] = &ServiceDatum{EndpointsID: connectcontactlens.EndpointsID, ServiceID: connectcontactlens.ServiceID, ServiceName: connectcontactlens.ServiceName}
	serviceData[ConnectParticipant] = &ServiceDatum{EndpointsID: connectparticipant.EndpointsID, ServiceID: connectparticipant.ServiceID, ServiceName: connectparticipant.ServiceName}
	serviceData[CUR] = &ServiceDatum{EndpointsID: costandusagereportservice.EndpointsID, ServiceID: costandusagereportservice.ServiceID, ServiceName: costandusagereportservice.ServiceName}
	serviceData[CustomerProfiles] = &ServiceDatum{EndpointsID: customerprofiles.EndpointsID, ServiceID: customerprofiles.ServiceID, ServiceName: customerprofiles.ServiceName}
	serviceData[DataBrew] = &ServiceDatum{EndpointsID: gluedatabrew.EndpointsID, ServiceID: gluedatabrew.ServiceID, ServiceName: gluedatabrew.ServiceName}
	serviceData[DataExchange] = &ServiceDatum{EndpointsID: dataexchange.EndpointsID, ServiceID: dataexchange.ServiceID, ServiceName: dataexchange.ServiceName}
	serviceData[DataPipeline] = &ServiceDatum{EndpointsID: datapipeline.EndpointsID, ServiceID: datapipeline.ServiceID, ServiceName: datapipeline.ServiceName}
	serviceData[DataSync] = &ServiceDatum{EndpointsID: datasync.EndpointsID, ServiceID: datasync.ServiceID, ServiceName: datasync.ServiceName}
	serviceData[DAX] = &ServiceDatum{EndpointsID: dax.EndpointsID, ServiceID: dax.ServiceID, ServiceName: dax.ServiceName}
	serviceData[Detective] = &ServiceDatum{EndpointsID: detective.EndpointsID, ServiceID: detective.ServiceID, ServiceName: detective.ServiceName}
	serviceData[DeviceFarm] = &ServiceDatum{EndpointsID: devicefarm.EndpointsID, ServiceID: devicefarm.ServiceID, ServiceName: devicefarm.ServiceName}
	serviceData[DevOpsGuru] = &ServiceDatum{EndpointsID: devopsguru.EndpointsID, ServiceID: devopsguru.ServiceID, ServiceName: devopsguru.ServiceName}
	serviceData[DirectConnect] = &ServiceDatum{EndpointsID: directconnect.EndpointsID, ServiceID: directconnect.ServiceID, ServiceName: directconnect.ServiceName}
	serviceData[Discovery] = &ServiceDatum{EndpointsID: applicationdiscoveryservice.EndpointsID, ServiceID: applicationdiscoveryservice.ServiceID, ServiceName: applicationdiscoveryservice.ServiceName}
	serviceData[DLM] = &ServiceDatum{EndpointsID: dlm.EndpointsID, ServiceID: dlm.ServiceID, ServiceName: dlm.ServiceName}
	serviceData[DMS] = &ServiceDatum{EndpointsID: databasemigrationservice.EndpointsID, ServiceID: databasemigrationservice.ServiceID, ServiceName: databasemigrationservice.ServiceName}
	serviceData[DocDB] = &ServiceDatum{EndpointsID: docdb.EndpointsID, ServiceID: docdb.ServiceID, ServiceName: docdb.ServiceName}
	serviceData[DRS] = &ServiceDatum{EndpointsID: drs.EndpointsID, ServiceID: drs.ServiceID, ServiceName: drs.ServiceName}
	serviceData[DS] = &ServiceDatum{EndpointsID: directoryservice.EndpointsID, ServiceID: directoryservice.ServiceID, ServiceName: directoryservice.ServiceName}
	serviceData[DynamoDB] = &ServiceDatum{EndpointsID: dynamodb.EndpointsID, ServiceID: dynamodb.ServiceID, ServiceName: dynamodb.ServiceName}
	serviceData[DynamoDBStreams] = &ServiceDatum{EndpointsID: dynamodbstreams.EndpointsID, ServiceID: dynamodbstreams.ServiceID, ServiceName: dynamodbstreams.ServiceName}
	serviceData[EBS] = &ServiceDatum{EndpointsID: ebs.EndpointsID, ServiceID: ebs.ServiceID, ServiceName: ebs.ServiceName}
	serviceData[EC2] = &ServiceDatum{EndpointsID: ec2.EndpointsID, ServiceID: ec2.ServiceID, ServiceName: ec2.ServiceName}
	serviceData[EC2InstanceConnect] = &ServiceDatum{EndpointsID: ec2instanceconnect.EndpointsID, ServiceID: ec2instanceconnect.ServiceID, ServiceName: ec2instanceconnect.ServiceName}
	serviceData[ECR] = &ServiceDatum{EndpointsID: ecr.EndpointsID, ServiceID: ecr.ServiceID, ServiceName: ecr.ServiceName}
	serviceData[ECRPublic] = &ServiceDatum{EndpointsID: ecrpublic.EndpointsID, ServiceID: ecrpublic.ServiceID, ServiceName: ecrpublic.ServiceName}
	serviceData[ECS] = &ServiceDatum{EndpointsID: ecs.EndpointsID, ServiceID: ecs.ServiceID, ServiceName: ecs.ServiceName}
	serviceData[EFS] = &ServiceDatum{EndpointsID: efs.EndpointsID, ServiceID: efs.ServiceID, ServiceName: efs.ServiceName}
	serviceData[EKS] = &ServiceDatum{EndpointsID: eks.EndpointsID, ServiceID: eks.ServiceID, ServiceName: eks.ServiceName}
	serviceData[ElastiCache] = &ServiceDatum{EndpointsID: elasticache.EndpointsID, ServiceID: elasticache.ServiceID, ServiceName: elasticache.ServiceName}
	serviceData[ElasticBeanstalk] = &ServiceDatum{EndpointsID: elasticbeanstalk.EndpointsID, ServiceID: elasticbeanstalk.ServiceID, ServiceName: elasticbeanstalk.ServiceName}
	serviceData[ElasticInference] = &ServiceDatum{EndpointsID: elasticinference.EndpointsID, ServiceID: elasticinference.ServiceID, ServiceName: elasticinference.ServiceName}
	serviceData[Elasticsearch] = &ServiceDatum{EndpointsID: elasticsearchservice.EndpointsID, ServiceID: elasticsearchservice.ServiceID, ServiceName: elasticsearchservice.ServiceName}
	serviceData[ElasticTranscoder] = &ServiceDatum{EndpointsID: elastictranscoder.EndpointsID, ServiceID: elastictranscoder.ServiceID, ServiceName: elastictranscoder.ServiceName}
	serviceData[ELB] = &ServiceDatum{EndpointsID: elb.EndpointsID, ServiceID: elb.ServiceID, ServiceName: elb.ServiceName}
	serviceData[ELBV2] = &ServiceDatum{EndpointsID: elbv2.EndpointsID, ServiceID: elbv2.ServiceID, ServiceName: elbv2.ServiceName}
	serviceData[EMR] = &ServiceDatum{EndpointsID: emr.EndpointsID, ServiceID: emr.ServiceID, ServiceName: emr.ServiceName}
	serviceData[EMRContainers] = &ServiceDatum{EndpointsID: emrcontainers.EndpointsID, ServiceID: emrcontainers.ServiceID, ServiceName: emrcontainers.ServiceName}
	serviceData[Events] = &ServiceDatum{EndpointsID: eventbridge.EndpointsID, ServiceID: eventbridge.ServiceID, ServiceName: eventbridge.ServiceName}
	serviceData[Evidently] = &ServiceDatum{EndpointsID: cloudwatchevidently.EndpointsID, ServiceID: cloudwatchevidently.ServiceID, ServiceName: cloudwatchevidently.ServiceName}
	serviceData[FinSpace] = &ServiceDatum{EndpointsID: finspace.EndpointsID, ServiceID: finspace.ServiceID, ServiceName: finspace.ServiceName}
	serviceData[FinSpaceData] = &ServiceDatum{EndpointsID: finspacedata.EndpointsID, ServiceID: finspacedata.ServiceID, ServiceName: finspacedata.ServiceName}
	serviceData[Firehose] = &ServiceDatum{EndpointsID: firehose.EndpointsID, ServiceID: firehose.ServiceID, ServiceName: firehose.ServiceName}
	serviceData[FIS] = &ServiceDatum{EndpointsID: fis.EndpointsID, ServiceID: fis.ServiceID, ServiceName: fis.ServiceName}
	serviceData[FMS] = &ServiceDatum{EndpointsID: fms.EndpointsID, ServiceID: fms.ServiceID, ServiceName: fms.ServiceName}
	serviceData[Forecast] = &ServiceDatum{EndpointsID: forecastservice.EndpointsID, ServiceID: forecastservice.ServiceID, ServiceName: forecastservice.ServiceName}
	serviceData[ForecastQuery] = &ServiceDatum{EndpointsID: forecastqueryservice.EndpointsID, ServiceID: forecastqueryservice.ServiceID, ServiceName: forecastqueryservice.ServiceName}
	serviceData[FraudDetector] = &ServiceDatum{EndpointsID: frauddetector.EndpointsID, ServiceID: frauddetector.ServiceID, ServiceName: frauddetector.ServiceName}
	serviceData[FSx] = &ServiceDatum{EndpointsID: fsx.EndpointsID, ServiceID: fsx.ServiceID, ServiceName: fsx.ServiceName}
	serviceData[GameLift] = &ServiceDatum{EndpointsID: gamelift.EndpointsID, ServiceID: gamelift.ServiceID, ServiceName: gamelift.ServiceName}
	serviceData[Glacier] = &ServiceDatum{EndpointsID: glacier.EndpointsID, ServiceID: glacier.ServiceID, ServiceName: glacier.ServiceName}
	serviceData[GlobalAccelerator] = &ServiceDatum{EndpointsID: globalaccelerator.EndpointsID, ServiceID: globalaccelerator.ServiceID, ServiceName: globalaccelerator.ServiceName}
	serviceData[Glue] = &ServiceDatum{EndpointsID: glue.EndpointsID, ServiceID: glue.ServiceID, ServiceName: glue.ServiceName}
	serviceData[Grafana] = &ServiceDatum{EndpointsID: managedgrafana.EndpointsID, ServiceID: managedgrafana.ServiceID, ServiceName: managedgrafana.ServiceName}
	serviceData[Greengrass] = &ServiceDatum{EndpointsID: greengrass.EndpointsID, ServiceID: greengrass.ServiceID, ServiceName: greengrass.ServiceName}
	serviceData[GreengrassV2] = &ServiceDatum{EndpointsID: greengrassv2.EndpointsID, ServiceID: greengrassv2.ServiceID, ServiceName: greengrassv2.ServiceName}
	serviceData[GroundStation] = &ServiceDatum{EndpointsID: groundstation.EndpointsID, ServiceID: groundstation.ServiceID, ServiceName: groundstation.ServiceName}
	serviceData[GuardDuty] = &ServiceDatum{EndpointsID: guardduty.EndpointsID, ServiceID: guardduty.ServiceID, ServiceName: guardduty.ServiceName}
	serviceData[Health] = &ServiceDatum{EndpointsID: health.EndpointsID, ServiceID: health.ServiceID, ServiceName: health.ServiceName}
	serviceData[HealthLake] = &ServiceDatum{EndpointsID: healthlake.EndpointsID, ServiceID: healthlake.ServiceID, ServiceName: healthlake.ServiceName}
	serviceData[Honeycode] = &ServiceDatum{EndpointsID: honeycode.EndpointsID, ServiceID: honeycode.ServiceID, ServiceName: honeycode.ServiceName}
	serviceData[IAM] = &ServiceDatum{EndpointsID: iam.EndpointsID, ServiceID: iam.ServiceID, ServiceName: iam.ServiceName}
	serviceData[IdentityStore] = &ServiceDatum{EndpointsID: identitystore.EndpointsID, ServiceID: identitystore.ServiceID, ServiceName: identitystore.ServiceName}
	serviceData[ImageBuilder] = &ServiceDatum{EndpointsID: imagebuilder.EndpointsID, ServiceID: imagebuilder.ServiceID, ServiceName: imagebuilder.ServiceName}
	serviceData[Inspector] = &ServiceDatum{EndpointsID: inspector.EndpointsID, ServiceID: inspector.ServiceID, ServiceName: inspector.ServiceName}
	serviceData[Inspector2] = &ServiceDatum{EndpointsID: inspector2.EndpointsID, ServiceID: inspector2.ServiceID, ServiceName: inspector2.ServiceName}
	serviceData[IoT] = &ServiceDatum{EndpointsID: iot.EndpointsID, ServiceID: iot.ServiceID, ServiceName: iot.ServiceName}
	serviceData[IoT1ClickDevices] = &ServiceDatum{EndpointsID: iot1clickdevicesservice.EndpointsID, ServiceID: iot1clickdevicesservice.ServiceID, ServiceName: iot1clickdevicesservice.ServiceName}
	serviceData[IoT1ClickProjects] = &ServiceDatum{EndpointsID: iot1clickprojects.EndpointsID, ServiceID: iot1clickprojects.ServiceID, ServiceName: iot1clickprojects.ServiceName}
	serviceData[IoTAnalytics] = &ServiceDatum{EndpointsID: iotanalytics.EndpointsID, ServiceID: iotanalytics.ServiceID, ServiceName: iotanalytics.ServiceName}
	serviceData[IoTData] = &ServiceDatum{EndpointsID: iotdataplane.EndpointsID, ServiceID: iotdataplane.ServiceID, ServiceName: iotdataplane.ServiceName}
	serviceData[IoTDeviceAdvisor] = &ServiceDatum{EndpointsID: iotdeviceadvisor.EndpointsID, ServiceID: iotdeviceadvisor.ServiceID, ServiceName: iotdeviceadvisor.ServiceName}
	serviceData[IoTEvents] = &ServiceDatum{EndpointsID: iotevents.EndpointsID, ServiceID: iotevents.ServiceID, ServiceName: iotevents.ServiceName}
	serviceData[IoTEventsData] = &ServiceDatum{EndpointsID: ioteventsdata.EndpointsID, ServiceID: ioteventsdata.ServiceID, ServiceName: ioteventsdata.ServiceName}
	serviceData[IoTFleetHub] = &ServiceDatum{EndpointsID: iotfleethub.EndpointsID, ServiceID: iotfleethub.ServiceID, ServiceName: iotfleethub.ServiceName}
	serviceData[IoTJobsData] = &ServiceDatum{EndpointsID: iotjobsdataplane.EndpointsID, ServiceID: iotjobsdataplane.ServiceID, ServiceName: iotjobsdataplane.ServiceName}
	serviceData[IoTSecureTunneling] = &ServiceDatum{EndpointsID: iotsecuretunneling.EndpointsID, ServiceID: iotsecuretunneling.ServiceID, ServiceName: iotsecuretunneling.ServiceName}
	serviceData[IoTSiteWise] = &ServiceDatum{EndpointsID: iotsitewise.EndpointsID, ServiceID: iotsitewise.ServiceID, ServiceName: iotsitewise.ServiceName}
	serviceData[IoTThingsGraph] = &ServiceDatum{EndpointsID: iotthingsgraph.EndpointsID, ServiceID: iotthingsgraph.ServiceID, ServiceName: iotthingsgraph.ServiceName}
	serviceData[IoTTwinMaker] = &ServiceDatum{EndpointsID: iottwinmaker.EndpointsID, ServiceID: iottwinmaker.ServiceID, ServiceName: iottwinmaker.ServiceName}
	serviceData[IoTWireless] = &ServiceDatum{EndpointsID: iotwireless.EndpointsID, ServiceID: iotwireless.ServiceID, ServiceName: iotwireless.ServiceName}
	serviceData[IVS] = &ServiceDatum{EndpointsID: ivs.EndpointsID, ServiceID: ivs.ServiceID, ServiceName: ivs.ServiceName}
	serviceData[Kafka] = &ServiceDatum{EndpointsID: kafka.EndpointsID, ServiceID: kafka.ServiceID, ServiceName: kafka.ServiceName}
	serviceData[KafkaConnect] = &ServiceDatum{EndpointsID: kafkaconnect.EndpointsID, ServiceID: kafkaconnect.ServiceID, ServiceName: kafkaconnect.ServiceName}
	serviceData[Kendra] = &ServiceDatum{EndpointsID: kendra.EndpointsID, ServiceID: kendra.ServiceID, ServiceName: kendra.ServiceName}
	serviceData[Keyspaces] = &ServiceDatum{EndpointsID: keyspaces.EndpointsID, ServiceID: keyspaces.ServiceID, ServiceName: keyspaces.ServiceName}
	serviceData[Kinesis] = &ServiceDatum{EndpointsID: kinesis.EndpointsID, ServiceID: kinesis.ServiceID, ServiceName: kinesis.ServiceName}
	serviceData[KinesisAnalytics] = &ServiceDatum{EndpointsID: kinesisanalytics.EndpointsID, ServiceID: kinesisanalytics.ServiceID, ServiceName: kinesisanalytics.ServiceName}
	serviceData[KinesisAnalyticsV2] = &ServiceDatum{EndpointsID: kinesisanalyticsv2.EndpointsID, ServiceID: kinesisanalyticsv2.ServiceID, ServiceName: kinesisanalyticsv2.ServiceName}
	serviceData[KinesisVideo] = &ServiceDatum{EndpointsID: kinesisvideo.EndpointsID, ServiceID: kinesisvideo.ServiceID, ServiceName: kinesisvideo.ServiceName}
	serviceData[KinesisVideoArchivedMedia] = &ServiceDatum{EndpointsID: kinesisvideoarchivedmedia.EndpointsID, ServiceID: kinesisvideoarchivedmedia.ServiceID, ServiceName: kinesisvideoarchivedmedia.ServiceName}
	serviceData[KinesisVideoMedia] = &ServiceDatum{EndpointsID: kinesisvideomedia.EndpointsID, ServiceID: kinesisvideomedia.ServiceID, ServiceName: kinesisvideomedia.ServiceName}
	serviceData[KinesisVideoSignaling] = &ServiceDatum{EndpointsID: kinesisvideosignalingchannels.EndpointsID, ServiceID: kinesisvideosignalingchannels.ServiceID, ServiceName: kinesisvideosignalingchannels.ServiceName}
	serviceData[KMS] = &ServiceDatum{EndpointsID: kms.EndpointsID, ServiceID: kms.ServiceID, ServiceName: kms.ServiceName}
	serviceData[LakeFormation] = &ServiceDatum{EndpointsID: lakeformation.EndpointsID, ServiceID: lakeformation.ServiceID, ServiceName: lakeformation.ServiceName}
	serviceData[Lambda] = &ServiceDatum{EndpointsID: lambda.EndpointsID, ServiceID: lambda.ServiceID, ServiceName: lambda.ServiceName}
	serviceData[LexModels] = &ServiceDatum{EndpointsID: lexmodelbuildingservice.EndpointsID, ServiceID: lexmodelbuildingservice.ServiceID, ServiceName: lexmodelbuildingservice.ServiceName}
	serviceData[LexModelsV2] = &ServiceDatum{EndpointsID: lexmodelsv2.EndpointsID, ServiceID: lexmodelsv2.ServiceID, ServiceName: lexmodelsv2.ServiceName}
	serviceData[LexRuntime] = &ServiceDatum{EndpointsID: lexruntimeservice.EndpointsID, ServiceID: lexruntimeservice.ServiceID, ServiceName: lexruntimeservice.ServiceName}
	serviceData[LexRuntimeV2] = &ServiceDatum{EndpointsID: lexruntimev2.EndpointsID, ServiceID: lexruntimev2.ServiceID, ServiceName: lexruntimev2.ServiceName}
	serviceData[LicenseManager] = &ServiceDatum{EndpointsID: licensemanager.EndpointsID, ServiceID: licensemanager.ServiceID, ServiceName: licensemanager.ServiceName}
	serviceData[Lightsail] = &ServiceDatum{EndpointsID: lightsail.EndpointsID, ServiceID: lightsail.ServiceID, ServiceName: lightsail.ServiceName}
	serviceData[Location] = &ServiceDatum{EndpointsID: locationservice.EndpointsID, ServiceID: locationservice.ServiceID, ServiceName: locationservice.ServiceName}
	serviceData[LookoutEquipment] = &ServiceDatum{EndpointsID: lookoutequipment.EndpointsID, ServiceID: lookoutequipment.ServiceID, ServiceName: lookoutequipment.ServiceName}
	serviceData[LookoutMetrics] = &ServiceDatum{EndpointsID: lookoutmetrics.EndpointsID, ServiceID: lookoutmetrics.ServiceID, ServiceName: lookoutmetrics.ServiceName}
	serviceData[LookoutVision] = &ServiceDatum{EndpointsID: lookoutforvision.EndpointsID, ServiceID: lookoutforvision.ServiceID, ServiceName: lookoutforvision.ServiceName}
	serviceData[MachineLearning] = &ServiceDatum{EndpointsID: machinelearning.EndpointsID, ServiceID: machinelearning.ServiceID, ServiceName: machinelearning.ServiceName}
	serviceData[Macie] = &ServiceDatum{EndpointsID: macie.EndpointsID, ServiceID: macie.ServiceID, ServiceName: macie.ServiceName}
	serviceData[Macie2] = &ServiceDatum{EndpointsID: macie2.EndpointsID, ServiceID: macie2.ServiceID, ServiceName: macie2.ServiceName}
	serviceData[ManagedBlockchain] = &ServiceDatum{EndpointsID: managedblockchain.EndpointsID, ServiceID: managedblockchain.ServiceID, ServiceName: managedblockchain.ServiceName}
	serviceData[MarketplaceCatalog] = &ServiceDatum{EndpointsID: marketplacecatalog.EndpointsID, ServiceID: marketplacecatalog.ServiceID, ServiceName: marketplacecatalog.ServiceName}
	serviceData[MarketplaceCommerceAnalytics] = &ServiceDatum{EndpointsID: marketplacecommerceanalytics.EndpointsID, ServiceID: marketplacecommerceanalytics.ServiceID, ServiceName: marketplacecommerceanalytics.ServiceName}
	serviceData[MarketplaceEntitlement] = &ServiceDatum{EndpointsID: marketplaceentitlementservice.EndpointsID, ServiceID: marketplaceentitlementservice.ServiceID, ServiceName: marketplaceentitlementservice.ServiceName}
	serviceData[MarketplaceMetering] = &ServiceDatum{EndpointsID: marketplacemetering.EndpointsID, ServiceID: marketplacemetering.ServiceID, ServiceName: marketplacemetering.ServiceName}
	serviceData[MediaConnect] = &ServiceDatum{EndpointsID: mediaconnect.EndpointsID, ServiceID: mediaconnect.ServiceID, ServiceName: mediaconnect.ServiceName}
	serviceData[MediaConvert] = &ServiceDatum{EndpointsID: mediaconvert.EndpointsID, ServiceID: mediaconvert.ServiceID, ServiceName: mediaconvert.ServiceName}
	serviceData[MediaLive] = &ServiceDatum{EndpointsID: medialive.EndpointsID, ServiceID: medialive.ServiceID, ServiceName: medialive.ServiceName}
	serviceData[MediaPackage] = &ServiceDatum{EndpointsID: mediapackage.EndpointsID, ServiceID: mediapackage.ServiceID, ServiceName: mediapackage.ServiceName}
	serviceData[MediaPackageVOD] = &ServiceDatum{EndpointsID: mediapackagevod.EndpointsID, ServiceID: mediapackagevod.ServiceID, ServiceName: mediapackagevod.ServiceName}
	serviceData[MediaStore] = &ServiceDatum{EndpointsID: mediastore.EndpointsID, ServiceID: mediastore.ServiceID, ServiceName: mediastore.ServiceName}
	serviceData[MediaStoreData] = &ServiceDatum{EndpointsID: mediastoredata.EndpointsID, ServiceID: mediastoredata.ServiceID, ServiceName: mediastoredata.ServiceName}
	serviceData[MediaTailor] = &ServiceDatum{EndpointsID: mediatailor.EndpointsID, ServiceID: mediatailor.ServiceID, ServiceName: mediatailor.ServiceName}
	serviceData[MemoryDB] = &ServiceDatum{EndpointsID: memorydb.EndpointsID, ServiceID: memorydb.ServiceID, ServiceName: memorydb.ServiceName}
	serviceData[MgH] = &ServiceDatum{EndpointsID: migrationhub.EndpointsID, ServiceID: migrationhub.ServiceID, ServiceName: migrationhub.ServiceName}
	serviceData[Mgn] = &ServiceDatum{EndpointsID: mgn.EndpointsID, ServiceID: mgn.ServiceID, ServiceName: mgn.ServiceName}
	serviceData[MigrationHubConfig] = &ServiceDatum{EndpointsID: migrationhubconfig.EndpointsID, ServiceID: migrationhubconfig.ServiceID, ServiceName: migrationhubconfig.ServiceName}
	serviceData[MigrationHubRefactorSpaces] = &ServiceDatum{EndpointsID: migrationhubrefactorspaces.EndpointsID, ServiceID: migrationhubrefactorspaces.ServiceID, ServiceName: migrationhubrefactorspaces.ServiceName}
	serviceData[MigrationHubStrategy] = &ServiceDatum{EndpointsID: migrationhubstrategyrecommendations.EndpointsID, ServiceID: migrationhubstrategyrecommendations.ServiceID, ServiceName: migrationhubstrategyrecommendations.ServiceName}
	serviceData[Mobile] = &ServiceDatum{EndpointsID: mobile.EndpointsID, ServiceID: mobile.ServiceID, ServiceName: mobile.ServiceName}
	serviceData[MobileAnalytics] = &ServiceDatum{EndpointsID: mobileanalytics.EndpointsID, ServiceID: mobileanalytics.ServiceID, ServiceName: mobileanalytics.ServiceName}
	serviceData[MQ] = &ServiceDatum{EndpointsID: mq.EndpointsID, ServiceID: mq.ServiceID, ServiceName: mq.ServiceName}
	serviceData[MTurk] = &ServiceDatum{EndpointsID: mturk.EndpointsID, ServiceID: mturk.ServiceID, ServiceName: mturk.ServiceName}
	serviceData[MWAA] = &ServiceDatum{EndpointsID: mwaa.EndpointsID, ServiceID: mwaa.ServiceID, ServiceName: mwaa.ServiceName}
	serviceData[Neptune] = &ServiceDatum{EndpointsID: neptune.EndpointsID, ServiceID: neptune.ServiceID, ServiceName: neptune.ServiceName}
	serviceData[NetworkFirewall] = &ServiceDatum{EndpointsID: networkfirewall.EndpointsID, ServiceID: networkfirewall.ServiceID, ServiceName: networkfirewall.ServiceName}
	serviceData[NetworkManager] = &ServiceDatum{EndpointsID: networkmanager.EndpointsID, ServiceID: networkmanager.ServiceID, ServiceName: networkmanager.ServiceName}
	serviceData[Nimble] = &ServiceDatum{EndpointsID: nimblestudio.EndpointsID, ServiceID: nimblestudio.ServiceID, ServiceName: nimblestudio.ServiceName}
	serviceData[OpenSearch] = &ServiceDatum{EndpointsID: opensearchservice.EndpointsID, ServiceID: opensearchservice.ServiceID, ServiceName: opensearchservice.ServiceName}
	serviceData[OpsWorks] = &ServiceDatum{EndpointsID: opsworks.EndpointsID, ServiceID: opsworks.ServiceID, ServiceName: opsworks.ServiceName}
	serviceData[OpsWorksCM] = &ServiceDatum{EndpointsID: opsworkscm.EndpointsID, ServiceID: opsworkscm.ServiceID, ServiceName: opsworkscm.ServiceName}
	serviceData[Organizations] = &ServiceDatum{EndpointsID: organizations.EndpointsID, ServiceID: organizations.ServiceID, ServiceName: organizations.ServiceName}
	serviceData[Outposts] = &ServiceDatum{EndpointsID: outposts.EndpointsID, ServiceID: outposts.ServiceID, ServiceName: outposts.ServiceName}
	serviceData[Panorama] = &ServiceDatum{EndpointsID: panorama.EndpointsID, ServiceID: panorama.ServiceID, ServiceName: panorama.ServiceName}
	serviceData[Personalize] = &ServiceDatum{EndpointsID: personalize.EndpointsID, ServiceID: personalize.ServiceID, ServiceName: personalize.ServiceName}
	serviceData[PersonalizeEvents] = &ServiceDatum{EndpointsID: personalizeevents.EndpointsID, ServiceID: personalizeevents.ServiceID, ServiceName: personalizeevents.ServiceName}
	serviceData[PersonalizeRuntime] = &ServiceDatum{EndpointsID: personalizeruntime.EndpointsID, ServiceID: personalizeruntime.ServiceID, ServiceName: personalizeruntime.ServiceName}
	serviceData[PI] = &ServiceDatum{EndpointsID: pi.EndpointsID, ServiceID: pi.ServiceID, ServiceName: pi.ServiceName}
	serviceData[Pinpoint] = &ServiceDatum{EndpointsID: pinpoint.EndpointsID, ServiceID: pinpoint.ServiceID, ServiceName: pinpoint.ServiceName}
	serviceData[PinpointEmail] = &ServiceDatum{EndpointsID: pinpointemail.EndpointsID, ServiceID: pinpointemail.ServiceID, ServiceName: pinpointemail.ServiceName}
	serviceData[PinpointSMSVoice] = &ServiceDatum{EndpointsID: pinpointsmsvoice.EndpointsID, ServiceID: pinpointsmsvoice.ServiceID, ServiceName: pinpointsmsvoice.ServiceName}
	serviceData[Polly] = &ServiceDatum{EndpointsID: polly.EndpointsID, ServiceID: polly.ServiceID, ServiceName: polly.ServiceName}
	serviceData[Pricing] = &ServiceDatum{EndpointsID: pricing.EndpointsID, ServiceID: pricing.ServiceID, ServiceName: pricing.ServiceName}
	serviceData[Proton] = &ServiceDatum{EndpointsID: proton.EndpointsID, ServiceID: proton.ServiceID, ServiceName: proton.ServiceName}
	serviceData[QLDB] = &ServiceDatum{EndpointsID: qldb.EndpointsID, ServiceID: qldb.ServiceID, ServiceName: qldb.ServiceName}
	serviceData[QLDBSession] = &ServiceDatum{EndpointsID: qldbsession.EndpointsID, ServiceID: qldbsession.ServiceID, ServiceName: qldbsession.ServiceName}
	serviceData[QuickSight] = &ServiceDatum{EndpointsID: quicksight.EndpointsID, ServiceID: quicksight.ServiceID, ServiceName: quicksight.ServiceName}
	serviceData[RAM] = &ServiceDatum{EndpointsID: ram.EndpointsID, ServiceID: ram.ServiceID, ServiceName: ram.ServiceName}
	serviceData[RBin] = &ServiceDatum{EndpointsID: recyclebin.EndpointsID, ServiceID: recyclebin.ServiceID, ServiceName: recyclebin.ServiceName}
	serviceData[RDS] = &ServiceDatum{EndpointsID: rds.EndpointsID, ServiceID: rds.ServiceID, ServiceName: rds.ServiceName}
	serviceData[RDSData] = &ServiceDatum{EndpointsID: rdsdataservice.EndpointsID, ServiceID: rdsdataservice.ServiceID, ServiceName: rdsdataservice.ServiceName}
	serviceData[Redshift] = &ServiceDatum{EndpointsID: redshift.EndpointsID, ServiceID: redshift.ServiceID, ServiceName: redshift.ServiceName}
	serviceData[RedshiftData] = &ServiceDatum{EndpointsID: redshiftdataapiservice.EndpointsID, ServiceID: redshiftdataapiservice.ServiceID, ServiceName: redshiftdataapiservice.ServiceName}
	serviceData[Rekognition] = &ServiceDatum{EndpointsID: rekognition.EndpointsID, ServiceID: rekognition.ServiceID, ServiceName: rekognition.ServiceName}
	serviceData[ResilienceHub] = &ServiceDatum{EndpointsID: resiliencehub.EndpointsID, ServiceID: resiliencehub.ServiceID, ServiceName: resiliencehub.ServiceName}
	serviceData[ResourceGroups] = &ServiceDatum{EndpointsID: resourcegroups.EndpointsID, ServiceID: resourcegroups.ServiceID, ServiceName: resourcegroups.ServiceName}
	serviceData[ResourceGroupsTaggingAPI] = &ServiceDatum{EndpointsID: resourcegroupstaggingapi.EndpointsID, ServiceID: resourcegroupstaggingapi.ServiceID, ServiceName: resourcegroupstaggingapi.ServiceName}
	serviceData[RoboMaker] = &ServiceDatum{EndpointsID: robomaker.EndpointsID, ServiceID: robomaker.ServiceID, ServiceName: robomaker.ServiceName}
	serviceData[Route53] = &ServiceDatum{EndpointsID: route53.EndpointsID, ServiceID: route53.ServiceID, ServiceName: route53.ServiceName}
	serviceData[Route53Domains] = &ServiceDatum{EndpointsID: Route53DomainsEndpointID, ServiceID: route53domains.ServiceID, ServiceName: Route53DomainsServiceName}
	serviceData[Route53RecoveryCluster] = &ServiceDatum{EndpointsID: route53recoverycluster.EndpointsID, ServiceID: route53recoverycluster.ServiceID, ServiceName: route53recoverycluster.ServiceName}
	serviceData[Route53RecoveryControlConfig] = &ServiceDatum{EndpointsID: route53recoverycontrolconfig.EndpointsID, ServiceID: route53recoverycontrolconfig.ServiceID, ServiceName: route53recoverycontrolconfig.ServiceName}
	serviceData[Route53RecoveryReadiness] = &ServiceDatum{EndpointsID: route53recoveryreadiness.EndpointsID, ServiceID: route53recoveryreadiness.ServiceID, ServiceName: route53recoveryreadiness.ServiceName}
	serviceData[Route53Resolver] = &ServiceDatum{EndpointsID: route53resolver.EndpointsID, ServiceID: route53resolver.ServiceID, ServiceName: route53resolver.ServiceName}
	serviceData[RUM] = &ServiceDatum{EndpointsID: cloudwatchrum.EndpointsID, ServiceID: cloudwatchrum.ServiceID, ServiceName: cloudwatchrum.ServiceName}
	serviceData[S3] = &ServiceDatum{EndpointsID: s3.EndpointsID, ServiceID: s3.ServiceID, ServiceName: s3.ServiceName}
	serviceData[S3Control] = &ServiceDatum{EndpointsID: s3control.EndpointsID, ServiceID: s3control.ServiceID, ServiceName: s3control.ServiceName}
	serviceData[S3Outposts] = &ServiceDatum{EndpointsID: s3outposts.EndpointsID, ServiceID: s3outposts.ServiceID, ServiceName: s3outposts.ServiceName}
	serviceData[SageMaker] = &ServiceDatum{EndpointsID: sagemaker.EndpointsID, ServiceID: sagemaker.ServiceID, ServiceName: sagemaker.ServiceName}
	serviceData[SageMakerA2IRuntime] = &ServiceDatum{EndpointsID: augmentedairuntime.EndpointsID, ServiceID: augmentedairuntime.ServiceID, ServiceName: augmentedairuntime.ServiceName}
	serviceData[SageMakerEdge] = &ServiceDatum{EndpointsID: sagemakeredgemanager.EndpointsID, ServiceID: sagemakeredgemanager.ServiceID, ServiceName: sagemakeredgemanager.ServiceName}
	serviceData[SageMakerFeatureStoreRuntime] = &ServiceDatum{EndpointsID: sagemakerfeaturestoreruntime.EndpointsID, ServiceID: sagemakerfeaturestoreruntime.ServiceID, ServiceName: sagemakerfeaturestoreruntime.ServiceName}
	serviceData[SageMakerRuntime] = &ServiceDatum{EndpointsID: sagemakerruntime.EndpointsID, ServiceID: sagemakerruntime.ServiceID, ServiceName: sagemakerruntime.ServiceName}
	serviceData[SavingsPlans] = &ServiceDatum{EndpointsID: savingsplans.EndpointsID, ServiceID: savingsplans.ServiceID, ServiceName: savingsplans.ServiceName}
	serviceData[Schemas] = &ServiceDatum{EndpointsID: schemas.EndpointsID, ServiceID: schemas.ServiceID, ServiceName: schemas.ServiceName}
	serviceData[SecretsManager] = &ServiceDatum{EndpointsID: secretsmanager.EndpointsID, ServiceID: secretsmanager.ServiceID, ServiceName: secretsmanager.ServiceName}
	serviceData[SecurityHub] = &ServiceDatum{EndpointsID: securityhub.EndpointsID, ServiceID: securityhub.ServiceID, ServiceName: securityhub.ServiceName}
	serviceData[ServerlessRepo] = &ServiceDatum{EndpointsID: serverlessapplicationrepository.EndpointsID, ServiceID: serverlessapplicationrepository.ServiceID, ServiceName: serverlessapplicationrepository.ServiceName}
	serviceData[ServiceCatalog] = &ServiceDatum{EndpointsID: servicecatalog.EndpointsID, ServiceID: servicecatalog.ServiceID, ServiceName: servicecatalog.ServiceName}
	serviceData[ServiceCatalogAppRegistry] = &ServiceDatum{EndpointsID: appregistry.EndpointsID, ServiceID: appregistry.ServiceID, ServiceName: appregistry.ServiceName}
	serviceData[ServiceDiscovery] = &ServiceDatum{EndpointsID: servicediscovery.EndpointsID, ServiceID: servicediscovery.ServiceID, ServiceName: servicediscovery.ServiceName}
	serviceData[ServiceQuotas] = &ServiceDatum{EndpointsID: servicequotas.EndpointsID, ServiceID: servicequotas.ServiceID, ServiceName: servicequotas.ServiceName}
	serviceData[SES] = &ServiceDatum{EndpointsID: ses.EndpointsID, ServiceID: ses.ServiceID, ServiceName: ses.ServiceName}
	serviceData[SESV2] = &ServiceDatum{EndpointsID: sesv2.EndpointsID, ServiceID: sesv2.ServiceID, ServiceName: sesv2.ServiceName}
	serviceData[SFN] = &ServiceDatum{EndpointsID: sfn.EndpointsID, ServiceID: sfn.ServiceID, ServiceName: sfn.ServiceName}
	serviceData[Shield] = &ServiceDatum{EndpointsID: shield.EndpointsID, ServiceID: shield.ServiceID, ServiceName: shield.ServiceName}
	serviceData[Signer] = &ServiceDatum{EndpointsID: signer.EndpointsID, ServiceID: signer.ServiceID, ServiceName: signer.ServiceName}
	serviceData[SimpleDB] = &ServiceDatum{EndpointsID: simpledb.EndpointsID, ServiceID: simpledb.ServiceID, ServiceName: simpledb.ServiceName}
	serviceData[SMS] = &ServiceDatum{EndpointsID: sms.EndpointsID, ServiceID: sms.ServiceID, ServiceName: sms.ServiceName}
	serviceData[Snowball] = &ServiceDatum{EndpointsID: snowball.EndpointsID, ServiceID: snowball.ServiceID, ServiceName: snowball.ServiceName}
	serviceData[SnowDeviceManagement] = &ServiceDatum{EndpointsID: snowdevicemanagement.EndpointsID, ServiceID: snowdevicemanagement.ServiceID, ServiceName: snowdevicemanagement.ServiceName}
	serviceData[SNS] = &ServiceDatum{EndpointsID: sns.EndpointsID, ServiceID: sns.ServiceID, ServiceName: sns.ServiceName}
	serviceData[SQS] = &ServiceDatum{EndpointsID: sqs.EndpointsID, ServiceID: sqs.ServiceID, ServiceName: sqs.ServiceName}
	serviceData[SSM] = &ServiceDatum{EndpointsID: ssm.EndpointsID, ServiceID: ssm.ServiceID, ServiceName: ssm.ServiceName}
	serviceData[SSMContacts] = &ServiceDatum{EndpointsID: ssmcontacts.EndpointsID, ServiceID: ssmcontacts.ServiceID, ServiceName: ssmcontacts.ServiceName}
	serviceData[SSMIncidents] = &ServiceDatum{EndpointsID: ssmincidents.EndpointsID, ServiceID: ssmincidents.ServiceID, ServiceName: ssmincidents.ServiceName}
	serviceData[SSO] = &ServiceDatum{EndpointsID: sso.EndpointsID, ServiceID: sso.ServiceID, ServiceName: sso.ServiceName}
	serviceData[SSOAdmin] = &ServiceDatum{EndpointsID: ssoadmin.EndpointsID, ServiceID: ssoadmin.ServiceID, ServiceName: ssoadmin.ServiceName}
	serviceData[SSOOIDC] = &ServiceDatum{EndpointsID: ssooidc.EndpointsID, ServiceID: ssooidc.ServiceID, ServiceName: ssooidc.ServiceName}
	serviceData[StorageGateway] = &ServiceDatum{EndpointsID: storagegateway.EndpointsID, ServiceID: storagegateway.ServiceID, ServiceName: storagegateway.ServiceName}
	serviceData[STS] = &ServiceDatum{EndpointsID: sts.EndpointsID, ServiceID: sts.ServiceID, ServiceName: sts.ServiceName}
	serviceData[Support] = &ServiceDatum{EndpointsID: support.EndpointsID, ServiceID: support.ServiceID, ServiceName: support.ServiceName}
	serviceData[SWF] = &ServiceDatum{EndpointsID: swf.EndpointsID, ServiceID: swf.ServiceID, ServiceName: swf.ServiceName}
	serviceData[Synthetics] = &ServiceDatum{EndpointsID: synthetics.EndpointsID, ServiceID: synthetics.ServiceID, ServiceName: synthetics.ServiceName}
	serviceData[Textract] = &ServiceDatum{EndpointsID: textract.EndpointsID, ServiceID: textract.ServiceID, ServiceName: textract.ServiceName}
	serviceData[TimestreamQuery] = &ServiceDatum{EndpointsID: timestreamquery.EndpointsID, ServiceID: timestreamquery.ServiceID, ServiceName: timestreamquery.ServiceName}
	serviceData[TimestreamWrite] = &ServiceDatum{EndpointsID: timestreamwrite.EndpointsID, ServiceID: timestreamwrite.ServiceID, ServiceName: timestreamwrite.ServiceName}
	serviceData[Transcribe] = &ServiceDatum{EndpointsID: transcribeservice.EndpointsID, ServiceID: transcribeservice.ServiceID, ServiceName: transcribeservice.ServiceName}
	serviceData[TranscribeStreaming] = &ServiceDatum{EndpointsID: transcribestreamingservice.EndpointsID, ServiceID: transcribestreamingservice.ServiceID, ServiceName: transcribestreamingservice.ServiceName}
	serviceData[Transfer] = &ServiceDatum{EndpointsID: transfer.EndpointsID, ServiceID: transfer.ServiceID, ServiceName: transfer.ServiceName}
	serviceData[Translate] = &ServiceDatum{EndpointsID: translate.EndpointsID, ServiceID: translate.ServiceID, ServiceName: translate.ServiceName}
	serviceData[VoiceID] = &ServiceDatum{EndpointsID: voiceid.EndpointsID, ServiceID: voiceid.ServiceID, ServiceName: voiceid.ServiceName}
	serviceData[WAF] = &ServiceDatum{EndpointsID: waf.EndpointsID, ServiceID: waf.ServiceID, ServiceName: waf.ServiceName}
	serviceData[WAFRegional] = &ServiceDatum{EndpointsID: wafregional.EndpointsID, ServiceID: wafregional.ServiceID, ServiceName: wafregional.ServiceName}
	serviceData[WAFV2] = &ServiceDatum{EndpointsID: wafv2.EndpointsID, ServiceID: wafv2.ServiceID, ServiceName: wafv2.ServiceName}
	serviceData[WellArchitected] = &ServiceDatum{EndpointsID: wellarchitected.EndpointsID, ServiceID: wellarchitected.ServiceID, ServiceName: wellarchitected.ServiceName}
	serviceData[Wisdom] = &ServiceDatum{EndpointsID: connectwisdomservice.EndpointsID, ServiceID: connectwisdomservice.ServiceID, ServiceName: connectwisdomservice.ServiceName}
	serviceData[WorkDocs] = &ServiceDatum{EndpointsID: workdocs.EndpointsID, ServiceID: workdocs.ServiceID, ServiceName: workdocs.ServiceName}
	serviceData[WorkLink] = &ServiceDatum{EndpointsID: worklink.EndpointsID, ServiceID: worklink.ServiceID, ServiceName: worklink.ServiceName}
	serviceData[WorkMail] = &ServiceDatum{EndpointsID: workmail.EndpointsID, ServiceID: workmail.ServiceID, ServiceName: workmail.ServiceName}
	serviceData[WorkMailMessageFlow] = &ServiceDatum{EndpointsID: workmailmessageflow.EndpointsID, ServiceID: workmailmessageflow.ServiceID, ServiceName: workmailmessageflow.ServiceName}
	serviceData[WorkSpaces] = &ServiceDatum{EndpointsID: workspaces.EndpointsID, ServiceID: workspaces.ServiceID, ServiceName: workspaces.ServiceName}
	serviceData[WorkSpacesWeb] = &ServiceDatum{EndpointsID: workspacesweb.EndpointsID, ServiceID: workspacesweb.ServiceID, ServiceName: workspacesweb.ServiceName}
	serviceData[XRay] = &ServiceDatum{EndpointsID: xray.EndpointsID, ServiceID: xray.ServiceID, ServiceName: xray.ServiceName}
}

const (
	// column indices of CSV
	//awsCLIV2Command         = 0
	//awsCLIV2CommandNoDashes = 1
	//providerPackageActual   = 4
	//providerPackageCorrect  = 5
	//note                    = 12
	goV1Package         = 2
	goV2Package         = 3
	providerPackageBoth = 6
	providerNameUpper   = 7
	aliases             = 8
	goV1ClientName      = 9
	humanFriendly       = 10
	brand               = 11
	exclude             = 13
	deprecatedEnvVar    = 14
	envVar              = 15
)

func readCSVIntoServiceData() error {
	// names_data.csv is dynamically embedded so changes, additions should be made
	// there also

	//go:embed names_data.csv
	var namesData string
	r := csv.NewReader(strings.NewReader(namesData))

	d, err := r.ReadAll()
	if err != nil {
		return fmt.Errorf("reading CSV into service data: %w", err)
	}

	for i, l := range d {
		if i < 0 { // omit header line
			continue
		}

		if len(l) < 16 {
			return fmt.Errorf("CSV file format, expected 16+ columns, got: %d", len(l))
		}

		if l[exclude] != "" || l[providerPackageBoth] == "" {
			continue
		}

		serviceData[l[providerPackageBoth]].GoV1Package = l[goV1Package]
		serviceData[l[providerPackageBoth]].GoV2Package = l[goV2Package]
		serviceData[l[providerPackageBoth]].Brand = l[brand]
		serviceData[l[providerPackageBoth]].DeprecatedEnvVar = l[deprecatedEnvVar]
		serviceData[l[providerPackageBoth]].EnvVar = l[envVar]
		serviceData[l[providerPackageBoth]].GoV1ClientName = l[goV1ClientName]
		serviceData[l[providerPackageBoth]].HumanFriendly = l[humanFriendly]
		serviceData[l[providerPackageBoth]].ProviderNameUpper = l[providerNameUpper]

		if l[aliases] != "" {
			serviceData[l[providerPackageBoth]].Aliases = strings.Split(l[aliases], ";")
		}
	}

	return nil
}

func ProviderPackageForAlias(serviceAlias string) (string, error) {
	for k, v := range serviceData {
		for _, hclKey := range v.Aliases {
			if serviceAlias == hclKey {
				return k, nil
			}
		}
	}

	return "", fmt.Errorf("unable to find service for service alias %s", serviceAlias)
}

func ProviderPackages() []string {
	keys := make([]string, len(serviceData))

	i := 0
	for k := range serviceData {
		keys[i] = k
		i++
	}

	return keys
}

func Aliases() []string {
	keys := make([]string, 0)

	for _, v := range serviceData {
		keys = append(keys, v.Aliases...)
	}

	return keys
}

func ProviderNameUpper(service string) (string, error) {
	if v, ok := serviceData[service]; ok {
		return v.ProviderNameUpper, nil
	}

	return "", fmt.Errorf("no service data found for %s", service)
}

func DeprecatedEnvVar(service string) string {
	if v, ok := serviceData[service]; ok {
		return v.DeprecatedEnvVar
	}

	return ""
}

func EnvVar(service string) string {
	if v, ok := serviceData[service]; ok {
		return v.EnvVar
	}

	return ""
}

func FullHumanFriendly(service string) (string, error) {
	if v, ok := serviceData[service]; ok {
		if v.Brand == "" {
			return v.HumanFriendly, nil
		}

		return fmt.Sprintf("%s %s", v.Brand, v.HumanFriendly), nil
	}

	return "", fmt.Errorf("no service data found for %s", service)
}
