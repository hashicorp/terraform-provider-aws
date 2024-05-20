// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package names provides constants for AWS service names that are used as keys
// for the endpoints slice in internal/conns/conns.go. The package also exposes
// access to data found in the data/names_data.csv file, which provides additional
// service-related name information.
//
// Consumers of the names package include the conns package
// (internal/conn/conns.go), the provider package
// (internal/provider/provider.go), generators, and the skaff tool.
//
// It is very important that information in the data/names_data.csv be exactly
// correct because the Terrform AWS Provider relies on the information to
// function correctly.
package names

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/names/data"
)

// Endpoint constants defined by the AWS SDK v1 but not defined in the AWS SDK v2.
const (
	AccessAnalyzerEndpointID             = "access-analyzer"
	ACMPCAEndpointID                     = "acm-pca"
	AMPEndpointID                        = "aps"
	AppStreamEndpointID                  = "appstream2"
	ApplicationAutoscalingEndpointID     = "application-autoscaling"
	AppIntegrationsEndpointID            = "app-integrations"
	AppConfigEndpointID                  = "appconfig"
	AmplifyEndpointID                    = "amplify"
	APIGatewayID                         = "apigateway"
	APIGatewayV2EndpointID               = "apigateway"
	AthenaEndpointID                     = "athena"
	AuditManagerEndpointID               = "auditmanager"
	AutoScalingPlansEndpointID           = "autoscaling-plans"
	BatchEndpointID                      = "batch"
	BedrockAgentEndpointID               = "bedrockagent"
	BedrockEndpointID                    = "bedrock"
	BCMDataExportsEndpointID             = "bcm-data-exports"
	BudgetsEndpointID                    = "budgets"
	ChimeSDKMediaPipelinesEndpointID     = "media-pipelines-chime"
	ChimeSDKVoiceEndpointID              = "voice-chime"
	CloudFormationEndpointID             = "cloudformation"
	CloudFrontEndpointID                 = "cloudfront"
	CloudSearchEndpointID                = "cloudsearch"
	CloudWatchEndpointID                 = "monitoring"
	Cloud9EndpointID                     = "cloud9"
	CodeArtifactEndpointID               = "codeartifact"
	CodeGuruReviewerEndpointID           = "codeguru-reviewer"
	CodeStarConnectionsEndpointID        = "codestar-connections"
	CognitoIdentityEndpointID            = "cognito-identity"
	ComprehendEndpointID                 = "comprehend"
	ConfigServiceEndpointID              = "config"
	DevOpsGuruEndpointID                 = "devops-guru"
	ECREndpointID                        = "api.ecr"
	EKSEndpointID                        = "eks"
	EMREndpointID                        = "elasticmapreduce"
	EventsEndpointID                     = "events"
	EvidentlyEndpointID                  = "evidently"
	FMSEndpointID                        = "fms"
	IdentityStoreEndpointID              = "identitystore"
	Inspector2EndpointID                 = "inspector2"
	IVSChatEndpointID                    = "ivschat"
	KendraEndpointID                     = "kendra"
	KMSEndpointID                        = "kms"
	LambdaEndpointID                     = "lambda"
	LexV2ModelsEndpointID                = "models-v2-lex"
	M2EndpointID                         = "m2"
	MediaConvertEndpointID               = "mediaconvert"
	MediaLiveEndpointID                  = "medialive"
	MQEndpointID                         = "mq"
	ObservabilityAccessManagerEndpointID = "oam"
	OpenSearchServerlessEndpointID       = "aoss"
	OpenSearchIngestionEndpointID        = "osis"
	PipesEndpointID                      = "pipes"
	PollyEndpointID                      = "polly"
	QLDBEndpointID                       = "qldb"
	RedshiftServerlessEndpointID         = "redshift-serverless"
	RedshiftEndpointID                   = "redshift"
	RekognitionEndpointID                = "rekognition"
	ResourceExplorer2EndpointID          = "resource-explorer-2"
	RolesAnywhereEndpointID              = "rolesanywhere"
	Route53DomainsEndpointID             = "route53domains"
	SchedulerEndpointID                  = "scheduler"
	ServiceQuotasEndpointID              = "servicequotas"
	ServiceCatalogAppRegistryEndpointID  = "servicecatalog-appregistry"
	ShieldEndpointID                     = "shield"
	SSMEndpointID                        = "ssm"
	SSMIncidentsEndpointID               = "ssm-incidents"
	SSOAdminEndpointID                   = "sso"
	STSEndpointID                        = "sts"
	TranscribeEndpointID                 = "transcribe"
	VerifiedPermissionsEndpointID        = "verifiedpermissions"
	VPCLatticeEndpointID                 = "vpc-lattice"
	WAFEndpointID                        = "waf"
	WAFRegionalEndpointID                = "waf-regional"
)

// These should move to aws-sdk-go-base.
// See https://github.com/hashicorp/aws-sdk-go-base/issues/649.
const (
	ChinaPartitionID      = "aws-cn"     // AWS China partition.
	ISOPartitionID        = "aws-iso"    // AWS ISO (US) partition.
	ISOBPartitionID       = "aws-iso-b"  // AWS ISOB (US) partition.
	ISOEPartitionID       = "aws-iso-e"  // AWS ISOE (Europe) partition.
	ISOFPartitionID       = "aws-iso-f"  // AWS ISOF partition.
	StandardPartitionID   = "aws"        // AWS Standard partition.
	USGovCloudPartitionID = "aws-us-gov" // AWS GovCloud (US) partition.
)

const (
	// AWS Standard partition's regions.
	GlobalRegionID = "aws-global" // AWS Standard global region.

	AFSouth1RegionID     = "af-south-1"     // Africa (Cape Town).
	APEast1RegionID      = "ap-east-1"      // Asia Pacific (Hong Kong).
	APNortheast1RegionID = "ap-northeast-1" // Asia Pacific (Tokyo).
	APNortheast2RegionID = "ap-northeast-2" // Asia Pacific (Seoul).
	APNortheast3RegionID = "ap-northeast-3" // Asia Pacific (Osaka).
	APSouth1RegionID     = "ap-south-1"     // Asia Pacific (Mumbai).
	APSouth2RegionID     = "ap-south-2"     // Asia Pacific (Hyderabad).
	APSoutheast1RegionID = "ap-southeast-1" // Asia Pacific (Singapore).
	APSoutheast2RegionID = "ap-southeast-2" // Asia Pacific (Sydney).
	APSoutheast3RegionID = "ap-southeast-3" // Asia Pacific (Jakarta).
	APSoutheast4RegionID = "ap-southeast-4" // Asia Pacific (Melbourne).
	CACentral1RegionID   = "ca-central-1"   // Canada (Central).
	CAWest1RegionID      = "ca-west-1"      // Canada West (Calgary).
	EUCentral1RegionID   = "eu-central-1"   // Europe (Frankfurt).
	EUCentral2RegionID   = "eu-central-2"   // Europe (Zurich).
	EUNorth1RegionID     = "eu-north-1"     // Europe (Stockholm).
	EUSouth1RegionID     = "eu-south-1"     // Europe (Milan).
	EUSouth2RegionID     = "eu-south-2"     // Europe (Spain).
	EUWest1RegionID      = "eu-west-1"      // Europe (Ireland).
	EUWest2RegionID      = "eu-west-2"      // Europe (London).
	EUWest3RegionID      = "eu-west-3"      // Europe (Paris).
	ILCentral1RegionID   = "il-central-1"   // Israel (Tel Aviv).
	MECentral1RegionID   = "me-central-1"   // Middle East (UAE).
	MESouth1RegionID     = "me-south-1"     // Middle East (Bahrain).
	SAEast1RegionID      = "sa-east-1"      // South America (Sao Paulo).
	USEast1RegionID      = "us-east-1"      // US East (N. Virginia).
	USEast2RegionID      = "us-east-2"      // US East (Ohio).
	USWest1RegionID      = "us-west-1"      // US West (N. California).
	USWest2RegionID      = "us-west-2"      // US West (Oregon).

	// AWS China partition's regions.
	CNNorth1RegionID     = "cn-north-1"     // China (Beijing).
	CNNorthwest1RegionID = "cn-northwest-1" // China (Ningxia).

	// AWS GovCloud (US) partition's regions.
	USGovEast1RegionID = "us-gov-east-1" // AWS GovCloud (US-East).
	USGovWest1RegionID = "us-gov-west-1" // AWS GovCloud (US-West).

	// AWS ISO (US) partition's regions.
	USISOEast1RegionID = "us-iso-east-1" // US ISO East.
	USISOWest1RegionID = "us-iso-west-1" // US ISO WEST.

	// AWS ISOB (US) partition's regions.
	USISOBEast1RegionID = "us-isob-east-1" // US ISOB East (Ohio).
)

func DNSSuffixForPartition(partition string) string {
	switch partition {
	case "":
		return ""
	case ChinaPartitionID:
		return "amazonaws.com.cn"
	case ISOPartitionID:
		return "c2s.ic.gov"
	case ISOBPartitionID:
		return "sc2s.sgov.gov"
	case ISOEPartitionID:
		return "cloud.adc-e.uk"
	case ISOFPartitionID:
		return "csp.hci.ic.gov"
	default:
		return "amazonaws.com"
	}
}

func IsOptInRegion(region string) bool {
	switch region {
	case AFSouth1RegionID,
		APEast1RegionID, APSouth2RegionID,
		APSoutheast3RegionID, APSoutheast4RegionID,
		CAWest1RegionID,
		EUCentral2RegionID,
		EUSouth1RegionID, EUSouth2RegionID,
		ILCentral1RegionID,
		MECentral1RegionID,
		MESouth1RegionID:
		return true
	default:
		return false
	}
}

func PartitionForRegion(region string) string {
	switch region {
	case "":
		return ""
	case CNNorth1RegionID, CNNorthwest1RegionID:
		return ChinaPartitionID
	case USISOEast1RegionID, USISOWest1RegionID:
		return ISOPartitionID
	case USISOBEast1RegionID:
		return ISOBPartitionID
	case USGovEast1RegionID, USGovWest1RegionID:
		return USGovCloudPartitionID
	default:
		return StandardPartitionID
	}
}

// ReverseDNS switches a DNS hostname to reverse DNS and vice-versa.
func ReverseDNS(hostname string) string {
	parts := strings.Split(hostname, ".")

	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}

	return strings.Join(parts, ".")
}

// Type ServiceDatum corresponds closely to columns in `data/names_data.csv` and are
// described in detail in README.md.
type ServiceDatum struct {
	Aliases            []string
	AWSServiceEnvVar   string
	Brand              string
	ClientSDKV1        bool
	DeprecatedEnvVar   string
	EndpointOnly       bool
	GoV1ClientTypeName string
	GoV1Package        string
	GoV2Package        string
	HumanFriendly      string
	ProviderNameUpper  string
	SDKID              string
	TFAWSEnvVar        string
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

func readCSVIntoServiceData() error {
	// names_data.csv is dynamically embedded so changes, additions should be made
	// there also

	d, err := data.ReadAllServiceData()
	if err != nil {
		return fmt.Errorf("reading CSV into service data: %w", err)
	}

	for _, l := range d {
		if l.Exclude() {
			continue
		}

		if l.NotImplemented() && !l.EndpointOnly() {
			continue
		}

		p := l.ProviderPackage()

		serviceData[p] = &ServiceDatum{
			AWSServiceEnvVar:   l.AWSServiceEnvVar(),
			Brand:              l.Brand(),
			ClientSDKV1:        l.ClientSDKV1(),
			DeprecatedEnvVar:   l.DeprecatedEnvVar(),
			EndpointOnly:       l.EndpointOnly(),
			GoV1ClientTypeName: l.GoV1ClientTypeName(),
			GoV1Package:        l.GoV1Package(),
			GoV2Package:        l.GoV2Package(),
			HumanFriendly:      l.HumanFriendly(),
			ProviderNameUpper:  l.ProviderNameUpper(),
			SDKID:              l.SDKID(),
			TFAWSEnvVar:        l.TFAWSEnvVar(),
		}

		a := []string{p}

		if len(l.Aliases()) > 0 {
			a = append(a, l.Aliases()...)
		}

		serviceData[p].Aliases = a
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

type Endpoint struct {
	ProviderPackage string
	Aliases         []string
}

func Endpoints() []Endpoint {
	endpoints := make([]Endpoint, 0, len(serviceData))

	for k, v := range serviceData {
		ep := Endpoint{
			ProviderPackage: k,
		}
		if len(v.Aliases) > 1 {
			ep.Aliases = v.Aliases[1:]
		}
		endpoints = append(endpoints, ep)
	}

	return endpoints
}

type ServiceNameUpper struct {
	ProviderPackage   string
	ProviderNameUpper string
	SDKID             string
}

func ServiceNamesUpper() []ServiceNameUpper {
	serviceNames := make([]ServiceNameUpper, 0, len(serviceData))

	for k, v := range serviceData {
		sn := ServiceNameUpper{
			ProviderPackage:   k,
			ProviderNameUpper: v.ProviderNameUpper,
			SDKID:             v.SDKID,
		}
		serviceNames = append(serviceNames, sn)
	}

	return serviceNames
}

func ProviderNameUpper(service string) (string, error) {
	if v, ok := serviceData[service]; ok {
		return v.ProviderNameUpper, nil
	}

	return "", fmt.Errorf("no service data found for %s", service)
}

// Deprecated `AWS_<service>_ENDPOINT` envvar defined for some services
func DeprecatedEnvVar(service string) string {
	if v, ok := serviceData[service]; ok {
		return v.DeprecatedEnvVar
	}

	return ""
}

// Deprecated `TF_AWS_<service>_ENDPOINT` envvar defined for some services
func TFAWSEnvVar(service string) string {
	if v, ok := serviceData[service]; ok {
		return v.TFAWSEnvVar
	}

	return ""
}

// Standard service endpoint envvar defined by AWS
func AWSServiceEnvVar(service string) string {
	if v, ok := serviceData[service]; ok {
		return v.AWSServiceEnvVar
	}

	return ""
}

// Service SDK ID from AWS SDK for Go v2
func SDKID(service string) string {
	if v, ok := serviceData[service]; ok {
		return v.SDKID
	}

	return ""
}

func ClientSDKV1(service string) bool {
	if v, ok := serviceData[service]; ok {
		return v.ClientSDKV1
	}

	return false
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

func HumanFriendly(service string) (string, error) {
	if v, ok := serviceData[service]; ok {
		return v.HumanFriendly, nil
	}

	if s, err := ProviderPackageForAlias(service); err == nil {
		return HumanFriendly(s)
	}

	return "", fmt.Errorf("no service data found for %s", service)
}

func AWSGoPackage(providerPackage string, version int) (string, error) {
	switch version {
	case 1:
		return AWSGoV1Package(providerPackage)
	case 2:
		return AWSGoV2Package(providerPackage)
	default:
		return "", fmt.Errorf("unsupported AWS SDK Go version: %d", version)
	}
}

func AWSGoV1Package(providerPackage string) (string, error) {
	if v, ok := serviceData[providerPackage]; ok {
		return v.GoV1Package, nil
	}

	return "", fmt.Errorf("getting AWS SDK Go v1 package, %s not found", providerPackage)
}

func AWSGoV2Package(providerPackage string) (string, error) {
	if v, ok := serviceData[providerPackage]; ok {
		return v.GoV2Package, nil
	}

	return "", fmt.Errorf("getting AWS SDK Go v2 package, %s not found", providerPackage)
}

func AWSGoClientTypeName(providerPackage string, version int) (string, error) {
	switch version {
	case 1:
		return AWSGoV1ClientTypeName(providerPackage)
	case 2:
		return "Client", nil
	default:
		return "", fmt.Errorf("unsupported AWS SDK Go version: %d", version)
	}
}

func AWSGoV1ClientTypeName(providerPackage string) (string, error) {
	if v, ok := serviceData[providerPackage]; ok {
		return v.GoV1ClientTypeName, nil
	}

	return "", fmt.Errorf("getting AWS SDK Go v1 client type name, %s not found", providerPackage)
}
