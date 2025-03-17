// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package names provides constants for AWS service names that are used as keys
// for the endpoints slice in internal/conns/conns.go. The package also exposes
// access to data found in the data/names_data.hcl file, which provides additional
// service-related name information.
//
// Consumers of the names package include the conns package
// (internal/conn/conns.go), the provider package
// (internal/provider/provider.go), generators, and the skaff tool.
//
// It is very important that information in the data/names_data.hcl be exactly
// correct because the Terrform AWS Provider relies on the information to
// function correctly.
package names

import (
	"fmt"
	"log"
	"slices"

	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-provider-aws/names/data"
)

// Endpoint constants defined by the AWS SDK v1 but not defined in the AWS SDK v2.
const (
	ACMPCAEndpointID                       = "acm-pca"
	AMPEndpointID                          = "aps"
	APIGatewayID                           = "apigateway"
	APIGatewayV2EndpointID                 = "apigateway"
	AccessAnalyzerEndpointID               = "access-analyzer"
	AmplifyEndpointID                      = "amplify"
	AppConfigEndpointID                    = "appconfig"
	AppFabricEndpointID                    = "appfabric"
	AppIntegrationsEndpointID              = "app-integrations"
	AppMeshEndpointID                      = "appmesh"
	AppStreamEndpointID                    = "appstream2"
	AppSyncEndpointID                      = "appsync"
	ApplicationAutoscalingEndpointID       = "application-autoscaling"
	ApplicationInsightsEndpointID          = "applicationinsights"
	AthenaEndpointID                       = "athena"
	AuditManagerEndpointID                 = "auditmanager"
	AutoScalingPlansEndpointID             = "autoscaling-plans"
	BCMDataExportsEndpointID               = "bcm-data-exports"
	BackupEndpointID                       = "backup"
	BatchEndpointID                        = "batch"
	BedrockAgentEndpointID                 = "bedrockagent"
	BedrockEndpointID                      = "bedrock"
	BudgetsEndpointID                      = "budgets"
	ChimeEndpointID                        = "chime"
	ChimeSDKMediaPipelinesEndpointID       = "media-pipelines-chime"
	ChimeSDKVoiceEndpointID                = "voice-chime"
	Cloud9EndpointID                       = "cloud9"
	CloudControlEndpointID                 = "cloudcontrol"
	CloudFormationEndpointID               = "cloudformation"
	CloudFrontEndpointID                   = "cloudfront"
	CloudSearchEndpointID                  = "cloudsearch"
	CloudWatchEndpointID                   = "monitoring"
	CodeArtifactEndpointID                 = "codeartifact"
	CodeGuruReviewerEndpointID             = "codeguru-reviewer"
	CodeStarConnectionsEndpointID          = "codestar-connections"
	CognitoIdentityEndpointID              = "cognito-identity"
	ComprehendEndpointID                   = "comprehend"
	ComputeOptimizerEndpointID             = "compute-optimizer"
	ConfigServiceEndpointID                = "config"
	ConnectEndpointID                      = "connect"
	DataExchangeEndpointID                 = "dataexchange"
	DataPipelineEndpointID                 = "datapipeline"
	DataZoneEndpointID                     = "datazone"
	DetectiveEndpointID                    = "api.detective"
	DeviceFarmEndpointID                   = "devicefarm"
	DevOpsGuruEndpointID                   = "devops-guru"
	DirectConnectEndpointID                = "directconnect"
	DLMEndpointID                          = "dlm"
	ECREndpointID                          = "api.ecr"
	ECSEndpointID                          = "ecs"
	EFSEndpointID                          = "elasticfilesystem"
	EKSEndpointID                          = "eks"
	ELBEndpointID                          = "elasticloadbalancing"
	EMREndpointID                          = "elasticmapreduce"
	ElasticsearchEndpointID                = "es"
	ElasticTranscoderEndpointID            = "elastictranscoder"
	ElastiCacheEndpointID                  = "elasticache"
	EventsEndpointID                       = "events"
	EvidentlyEndpointID                    = "evidently"
	FMSEndpointID                          = "fms"
	FSxEndpointID                          = "fsx"
	GameLiftEndpointID                     = "gamelift"
	GrafanaEndpointID                      = "grafana"
	GlueEndpointID                         = "glue"
	IVSEndpointID                          = "ivs"
	IVSChatEndpointID                      = "ivschat"
	IdentityStoreEndpointID                = "identitystore"
	ImageBuilderEndpointID                 = "imagebuilder"
	Inspector2EndpointID                   = "inspector2"
	InternetMonitorEndpointID              = "internetmonitor"
	KMSEndpointID                          = "kms"
	KafkaConnectEndpointID                 = "kafkaconnect"
	KendraEndpointID                       = "kendra"
	KinesisVideoEndpointID                 = "kinesisvideo"
	LambdaEndpointID                       = "lambda"
	LexModelBuildingServiceEndpointID      = "models.lex"
	LexV2ModelsEndpointID                  = "models-v2-lex"
	LocationEndpointID                     = "location"
	M2EndpointID                           = "m2"
	MQEndpointID                           = "mq"
	Macie2EndpointID                       = "macie2"
	MediaConvertEndpointID                 = "mediaconvert"
	MediaLiveEndpointID                    = "medialive"
	ObservabilityAccessManagerEndpointID   = "oam"
	OpenSearchIngestionEndpointID          = "osis"
	OpenSearchServerlessEndpointID         = "aoss"
	PaymentCryptographyEndpointID          = "paymentcryptography"
	PipesEndpointID                        = "pipes"
	PollyEndpointID                        = "polly"
	QLDBEndpointID                         = "qldb"
	QuickSightEndpointID                   = "quicksight"
	RUMEndpointID                          = "rum"
	RedshiftEndpointID                     = "redshift"
	RedshiftServerlessEndpointID           = "redshift-serverless"
	RekognitionEndpointID                  = "rekognition"
	ResourceExplorer2EndpointID            = "resource-explorer-2"
	RolesAnywhereEndpointID                = "rolesanywhere"
	Route53DomainsEndpointID               = "route53domains"
	Route53RecoveryControlConfigEndpointID = "route53-recovery-control-config"
	ServiceCatalogEndpointID               = "servicecatalog"
	SSMEndpointID                          = "ssm"
	SSMIncidentsEndpointID                 = "ssm-incidents"
	SSMQuickSetupEndpointID                = "ssm-quicksetup"
	SSOAdminEndpointID                     = "sso"
	STSEndpointID                          = "sts"
	SchedulerEndpointID                    = "scheduler"
	SchemasEndpointID                      = "schemas"
	ServiceCatalogAppRegistryEndpointID    = "servicecatalog-appregistry"
	ServiceDiscoveryEndpointID             = "servicediscovery"
	ServiceQuotasEndpointID                = "servicequotas"
	SESEndpointID                          = "email"
	ShieldEndpointID                       = "shield"
	TranscribeEndpointID                   = "transcribe"
	TransferEndpointID                     = "transfer"
	VPCLatticeEndpointID                   = "vpc-lattice"
	VerifiedPermissionsEndpointID          = "verifiedpermissions"
	WAFEndpointID                          = "waf"
	WAFRegionalEndpointID                  = "waf-regional"
)

// PartitionForRegion returns the partition for the given Region.
// Returns the empty partition if the Region is empty.
// Returns the standard partition if no known partition includes the Region.
func PartitionForRegion(region string) endpoints.Partition {
	if region == "" {
		return endpoints.Partition{}
	}

	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), region); ok {
		return partition
	}

	return PartitionForRegion(endpoints.UsEast1RegionID)
}

// Type ServiceDatum corresponds closely to attributes and blocks in `data/names_data.hcl` and are
// described in detail in README.md.
type serviceDatum struct {
	aliases           []string
	brand             string
	humanFriendly     string
	providerNameUpper string
}

// serviceData key is the AWS provider service package
var serviceData map[string]serviceDatum

func init() {
	serviceData = make(map[string]serviceDatum)

	// Data from names_data.hcl
	if err := readHCLIntoServiceData(); err != nil {
		log.Fatalf("reading HCL into service data: %s", err)
	}
}

func readHCLIntoServiceData() error {
	// names_data.hcl is dynamically embedded so changes, additions should be made
	// there also

	d, err := data.ReadAllServiceData()
	if err != nil {
		return fmt.Errorf("reading HCL into service data: %w", err)
	}

	for _, l := range d {
		if l.Exclude() {
			continue
		}

		if l.NotImplemented() && !l.EndpointOnly() {
			continue
		}

		p := l.ProviderPackage()

		sd := serviceDatum{
			brand:             l.Brand(),
			humanFriendly:     l.HumanFriendly(),
			providerNameUpper: l.ProviderNameUpper(),
		}

		a := []string{p}

		if len(l.Aliases()) > 0 {
			a = append(a, l.Aliases()...)
		}

		sd.aliases = a

		serviceData[p] = sd
	}

	return nil
}

func ProviderPackageForAlias(serviceAlias string) (string, error) {
	for k, v := range serviceData {
		if slices.Contains(v.aliases, serviceAlias) {
			return k, nil
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
		keys = append(keys, v.aliases...)
	}

	return keys
}

func ProviderNameUpper(service string) (string, error) {
	if v, ok := serviceData[service]; ok {
		return v.providerNameUpper, nil
	}

	return "", fmt.Errorf("no service data found for %s", service)
}

func FullHumanFriendly(service string) (string, error) {
	if v, ok := serviceData[service]; ok {
		if v.brand == "" {
			return v.humanFriendly, nil
		}

		return fmt.Sprintf("%s %s", v.brand, v.humanFriendly), nil
	}

	if s, err := ProviderPackageForAlias(service); err == nil {
		return FullHumanFriendly(s)
	}

	return "", fmt.Errorf("no service data found for %s", service)
}

func HumanFriendly(service string) (string, error) {
	if v, ok := serviceData[service]; ok {
		return v.humanFriendly, nil
	}

	if s, err := ProviderPackageForAlias(service); err == nil {
		return HumanFriendly(s)
	}

	return "", fmt.Errorf("no service data found for %s", service)
}
