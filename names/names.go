// Package names provides constants for AWS service names that are used as keys
// for the endpoints slice in internal/conns/conns.go. The package also exposes
// access to data found in the names_data.csv file, which provides additional
// service-related name information.
//
// Consumers of the names package include the conns package
// (internal/conn/conns.go), the provider package
// (internal/provider/provider.go), generators, and the skaff tool.
//
// It is very important that information in the names_data.csv be exactly
// correct because the Terrform AWS Provider relies on the information to
// function correctly.

package names

import (
	_ "embed"
	"encoding/csv"
	"fmt"
	"log"
	"strings"
)

const (
	// The value should exactly correspond to the TF AWS provider service
	// package name (if it exists) and the providerPackageBoth column in
	// names_data.csv. See README.md.
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

// This "should" be defined by the AWS Go SDK v2, but currently isn't.
const (
	Route53DomainsEndpointID = "route53domains"
)

// Type ServiceDatum corresponds closely to columns in `names_data.csv` and are
// described in detail in README.md.
type ServiceDatum struct {
	Aliases           []string
	Brand             string
	DeprecatedEnvVar  string
	EnvVar            string
	GoV1ClientName    string
	GoV1Package       string
	GoV2Package       string
	HumanFriendly     string
	ProviderNameUpper string
}

// serviceData key is the AWS provider service package
var serviceData map[string]*ServiceDatum

func init() {
	serviceData = make(map[string]*ServiceDatum)

	// Data from names_data.csv
	if err := readCSVIntoServiceData(); err != nil {
		log.Fatalf("reading CSV into service data: %s", err)
	}
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

//go:embed names_data.csv
var namesData string

func readCSVIntoServiceData() error {
	// names_data.csv is dynamically embedded so changes, additions should be made
	// there also

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

		serviceData[l[providerPackageBoth]] = &ServiceDatum{
			Brand:             l[brand],
			DeprecatedEnvVar:  l[deprecatedEnvVar],
			EnvVar:            l[envVar],
			GoV1ClientName:    l[goV1ClientName],
			GoV1Package:       l[goV1Package],
			GoV2Package:       l[goV2Package],
			HumanFriendly:     l[humanFriendly],
			ProviderNameUpper: l[providerNameUpper],
		}

		a := []string{l[providerPackageBoth]}

		if l[aliases] != "" {
			a = append(a, strings.Split(l[aliases], ";")...)
		}

		serviceData[l[providerPackageBoth]].Aliases = a
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

	if s, err := ProviderPackageForAlias(service); err == nil {
		return FullHumanFriendly(s)
	}

	return "", fmt.Errorf("no service data found for %s", service)
}
