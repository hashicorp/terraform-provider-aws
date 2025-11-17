// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

// Tests AutoFlex's Expand/Flatten of AWS API XML wrappers (Items/Quantity).

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// AWS types
type Distribution struct {
	ARN                           *string             // required
	DistributionConfig            *DistributionConfig // required
	DomainName                    *string             // required
	Id                            *string             // required
	InProgressInvalidationBatches *int32              // required
	LastModifiedTime              *time.Time          // required
	Status                        *string             // required
}

type DistributionConfig struct {
	Aliases                      *Aliases
	AnycastIpListId              *string
	CacheBehaviors               *CacheBehaviors
	CallerReference              *string // required
	Comment                      *string // required
	ConnectionMode               ConnectionMode
	ContinuousDeploymentPolicyId *string
	CustomErrorResponses         *CustomErrorResponses
	DefaultCacheBehavior         *DefaultCacheBehavior // required
	DefaultRootObject            *string
	Enabled                      *bool // required
	HttpVersion                  HttpVersion
	IsIPV6Enabled                *bool
	Logging                      *LoggingConfig
	OriginGroups                 *OriginGroups
	Origins                      *Origins // required
	PriceClass                   PriceClass
	Restrictions                 *Restrictions
	Staging                      *bool
	TenantConfig                 *TenantConfig
	ViewerCertificate            *ViewerCertificate
	WebACLId                     *string
}

type DefaultCacheBehavior struct {
	AllowedMethods             *AllowedMethods
	CachePolicyId              *string
	Compress                   *bool
	DefaultTTL                 *int64
	FieldLevelEncryptionId     *string
	ForwardedValues            *ForwardedValues
	FunctionAssociations       *FunctionAssociations
	GrpcConfig                 *GrpcConfig
	LambdaFunctionAssociations *LambdaFunctionAssociations
	MaxTTL                     *int64
	MinTTL                     *int64
	OriginRequestPolicyId      *string
	RealtimeLogConfigArn       *string
	ResponseHeadersPolicyId    *string
	SmoothStreaming            *bool
	TargetOriginId             *string // required
	TrustedKeyGroups           *TrustedKeyGroups
	TrustedSigners             *TrustedSigners
	ViewerProtocolPolicy       ViewerProtocolPolicy // required
}

type AllowedMethods struct {
	CachedMethods *CachedMethods
	Items         []Method // required
	Quantity      *int32   // required
}

type Method string

const (
	MethodGet     Method = "GET"
	MethodHead    Method = "HEAD"
	MethodPost    Method = "POST"
	MethodPut     Method = "PUT"
	MethodPatch   Method = "PATCH"
	MethodOptions Method = "OPTIONS"
	MethodDelete  Method = "DELETE"
)

// Values returns all possible values of Method
func (m Method) Values() []Method {
	return []Method{
		MethodGet,
		MethodHead,
		MethodPost,
		MethodPut,
		MethodPatch,
		MethodOptions,
		MethodDelete,
	}
}

type CachedMethods struct {
	Items    []Method // required
	Quantity *int32   // required
}

type ForwardedValues struct { // deprecated
	Cookies              *CookiePreference // required
	Headers              *Headers
	QueryString          *bool // required
	QueryStringCacheKeys *QueryStringCacheKeys
}

type CookiePreference struct { // deprecated
	Forward          ItemSelection // required
	WhitelistedNames *CookieNames
}

type ItemSelection string

const (
	ItemSelectionNone      ItemSelection = "none"
	ItemSelectionWhitelist ItemSelection = "whitelist"
	ItemSelectionAll       ItemSelection = "all"
)

type CookieNames struct {
	Quantity *int32 // required
	Items    []string
}

type Headers struct {
	Items    []string
	Quantity *int32 // required
}

type QueryStringCacheKeys struct {
	Items    []string
	Quantity *int32 // required
}

type GrpcConfig struct {
	Enabled *bool // required
}

type LambdaFunctionAssociations struct {
	Quantity *int32 // required
	Items    []LambdaFunctionAssociation
}

type LambdaFunctionAssociation struct {
	EventType         EventType // required
	IncludeBody       *bool
	LambdaFunctionARN *string // required
}

type EventType string

const (
	EventTypeViewerRequest  EventType = "viewer-request"
	EventTypeViewerResponse EventType = "viewer-response"
	EventTypeOriginRequest  EventType = "origin-request"
	EventTypeOriginResponse EventType = "origin-response"
)

// Values returns all possible values of EventType
func (e EventType) Values() []EventType {
	return []EventType{
		EventTypeViewerRequest,
		EventTypeViewerResponse,
		EventTypeOriginRequest,
		EventTypeOriginResponse,
	}
}

type TrustedKeyGroups struct {
	Enabled  *bool // required
	Items    []string
	Quantity *int32 // required
}

type TrustedSigners struct {
	Enabled  *bool  // required
	Quantity *int32 // required
	Items    []string
}

type ViewerProtocolPolicy string

const (
	ViewerProtocolPolicyAllowAll        ViewerProtocolPolicy = "allow-all"
	ViewerProtocolPolicyHttpsOnly       ViewerProtocolPolicy = "https-only"
	ViewerProtocolPolicyRedirectToHttps ViewerProtocolPolicy = "redirect-to-https"
)

// Values returns all possible values of ViewerProtocolPolicy
func (v ViewerProtocolPolicy) Values() []ViewerProtocolPolicy {
	return []ViewerProtocolPolicy{
		ViewerProtocolPolicyAllowAll,
		ViewerProtocolPolicyHttpsOnly,
		ViewerProtocolPolicyRedirectToHttps,
	}
}

type Origins struct {
	Items    []Origin // required
	Quantity *int32   // required
}

type Origin struct {
	ConnectionAttempts        *int32
	ConnectionTimeout         *int32
	CustomHeaders             *CustomHeaders
	CustomOriginConfig        *CustomOriginConfig
	DomainName                *string // required
	Id                        *string // required
	OriginAccessControlId     *string
	OriginPath                *string
	OriginShield              *OriginShield
	ResponseCompletionTimeout *int32
	S3OriginConfig            *S3OriginConfig
	VpcOriginConfig           *VpcOriginConfig
}

type CustomHeaders struct {
	Quantity *int32 // required
	Items    []OriginCustomHeader
}

type OriginCustomHeader struct {
	HeaderName  *string // required
	HeaderValue *string // required
}

type CustomOriginConfig struct {
	HTTPPort               *int32 // required
	HTTPSPort              *int32 // required
	IpAddressType          IpAddressType
	OriginKeepaliveTimeout *int32
	OriginProtocolPolicy   OriginProtocolPolicy // required
	OriginReadTimeout      *int32
	OriginSslProtocols     *OriginSslProtocols
}

type IpAddressType string

const (
	IpAddressTypeIpv4      IpAddressType = "ipv4"
	IpAddressTypeIpv6      IpAddressType = "ipv6"
	IpAddressTypeDualStack IpAddressType = "dualstack"
)

// Values returns all possible values of IpAddressType
func (i IpAddressType) Values() []IpAddressType {
	return []IpAddressType{
		IpAddressTypeIpv4,
		IpAddressTypeIpv6,
		IpAddressTypeDualStack,
	}
}

type OriginProtocolPolicy string

const (
	OriginProtocolPolicyHttpOnly    OriginProtocolPolicy = "http-only"
	OriginProtocolPolicyMatchViewer OriginProtocolPolicy = "match-viewer"
	OriginProtocolPolicyHttpsOnly   OriginProtocolPolicy = "https-only"
)

// Values returns all possible values of OriginProtocolPolicy
func (o OriginProtocolPolicy) Values() []OriginProtocolPolicy {
	return []OriginProtocolPolicy{
		OriginProtocolPolicyHttpOnly,
		OriginProtocolPolicyMatchViewer,
		OriginProtocolPolicyHttpsOnly,
	}
}

type OriginSslProtocols struct {
	Items    []SslProtocol // required
	Quantity *int32        // required
}

type SslProtocol string

const (
	SslProtocolSSLv3  SslProtocol = "SSLv3"
	SslProtocolTLSv1  SslProtocol = "TLSv1"
	SslProtocolTLSv11 SslProtocol = "TLSv1.1"
	SslProtocolTLSv12 SslProtocol = "TLSv1.2"
)

// Values returns all possible values of SslProtocol
func (s SslProtocol) Values() []SslProtocol {
	return []SslProtocol{
		SslProtocolSSLv3,
		SslProtocolTLSv1,
		SslProtocolTLSv11,
		SslProtocolTLSv12,
	}
}

type OriginShield struct {
	Enabled            *bool // required
	OriginShieldRegion *string
}

type S3OriginConfig struct {
	OriginAccessIdentity *string // required
	OriginReadTimeout    *int32
}

type VpcOriginConfig struct {
	OriginKeepaliveTimeout *int32
	OriginReadTimeout      *int32
	VpcOriginId            *string // required
}

type Aliases struct {
	Items    []string
	Quantity *int32 // required
}

type CacheBehaviors struct {
	Quantity *int32 // required
	Items    []CacheBehavior
}

type CacheBehavior struct {
	AllowedMethods             *AllowedMethods
	CachePolicyId              *string
	Compress                   *bool
	DefaultTTL                 *int64 // deprecated
	FieldLevelEncryptionId     *string
	ForwardedValues            *ForwardedValues // deprecated
	FunctionAssociations       *FunctionAssociations
	GrpcConfig                 *GrpcConfig
	LambdaFunctionAssociations *LambdaFunctionAssociations
	MaxTTL                     *int64 // deprecated
	MinTTL                     *int64 // deprecated
	OriginRequestPolicyId      *string
	PathPattern                *string // required
	RealtimeLogConfigArn       *string
	ResponseHeadersPolicyId    *string
	SmoothStreaming            *bool
	TargetOriginId             *string // required
	TrustedKeyGroups           *TrustedKeyGroups
	TrustedSigners             *TrustedSigners
	ViewerProtocolPolicy       ViewerProtocolPolicy // required
}

type FunctionAssociations struct {
	Items    []FunctionAssociation `json:"Items"`
	Quantity *int32                `json:"Quantity"`
}

type FunctionAssociation struct {
	FunctionArn *string   // required
	EventType   EventType // required
}

type ConnectionMode string

const (
	ConnectionModeDirect     ConnectionMode = "direct"
	ConnectionModeTenantOnly ConnectionMode = "tenant-only"
)

// Values returns all possible values of ConnectionMode
func (c ConnectionMode) Values() []ConnectionMode {
	return []ConnectionMode{
		ConnectionModeDirect,
		ConnectionModeTenantOnly,
	}
}

type FunctionAssociationTF struct {
	EventType   fwtypes.StringEnum[EventType] `tfsdk:"event_type"`
	FunctionArn types.String                  `tfsdk:"function_arn"`
}

type DistributionConfigTF struct {
	FunctionAssociations fwtypes.SetNestedObjectValueOf[FunctionAssociationTF] `tfsdk:"function_associations" autoflex:",xmlwrapper=items"`
}

type CustomErrorResponses struct {
	Quantity *int32 // required
	Items    []CustomErrorResponse
}

type CustomErrorResponse struct {
	ErrorCachingMinTTL *int64
	ErrorCode          *int32 // required
	ResponseCode       *string
	ResponsePagePath   *string
}

type HttpVersion string

const (
	HttpVersionHttp11    HttpVersion = "http1.1"
	HttpVersionHttp2     HttpVersion = "http2"
	HttpVersionHttp3     HttpVersion = "http3"
	HttpVersionHttp2and3 HttpVersion = "http2and3"
)

// Values returns all possible values of HttpVersion
func (h HttpVersion) Values() []HttpVersion {
	return []HttpVersion{
		HttpVersionHttp11,
		HttpVersionHttp2,
		HttpVersionHttp3,
		HttpVersionHttp2and3,
	}
}

type LoggingConfig struct {
	Bucket         *string
	Enabled        *bool
	IncludeCookies *bool
	Prefix         *string
}

type OriginGroups struct {
	Quantity *int32 // required
	Items    []OriginGroup
}

type OriginGroup struct {
	FailoverCriteria  *OriginGroupFailoverCriteria // required
	Id                *string                      // required
	Members           *OriginGroupMembers          // required
	SelectionCriteria OriginGroupSelectionCriteria
}

type OriginGroupFailoverCriteria struct {
	StatusCodes *StatusCodes // required
}

type StatusCodes struct {
	Items    []int32 // required
	Quantity *int32  // required
}

type OriginGroupMembers struct {
	Items    []OriginGroupMember // required
	Quantity *int32              // required
}

type OriginGroupMember struct {
	OriginId *string // required
}

type OriginGroupSelectionCriteria string

const (
	OriginGroupSelectionCriteriaDefault           OriginGroupSelectionCriteria = "default"
	OriginGroupSelectionCriteriaMediaQualityBased OriginGroupSelectionCriteria = "media-quality-based"
)

// Values returns all possible values of OriginGroupSelectionCriteria
func (o OriginGroupSelectionCriteria) Values() []OriginGroupSelectionCriteria {
	return []OriginGroupSelectionCriteria{
		OriginGroupSelectionCriteriaDefault,
		OriginGroupSelectionCriteriaMediaQualityBased,
	}
}

type PriceClass string

const (
	PriceClassPriceClass100 PriceClass = "PriceClass_100"
	PriceClassPriceClass200 PriceClass = "PriceClass_200"
	PriceClassPriceClassAll PriceClass = "PriceClass_All"
	PriceClassNone          PriceClass = "None"
)

// Values returns all possible values of PriceClass
func (p PriceClass) Values() []PriceClass {
	return []PriceClass{
		PriceClassPriceClass100,
		PriceClassPriceClass200,
		PriceClassPriceClassAll,
		PriceClassNone,
	}
}

type Restrictions struct {
	GeoRestriction *GeoRestriction // required
}

type GeoRestriction struct {
	Items           []string
	Quantity        *int32             // required
	RestrictionType GeoRestrictionType // required
}

type GeoRestrictionType string

const (
	GeoRestrictionTypeBlacklist GeoRestrictionType = "blacklist"
	GeoRestrictionTypeWhitelist GeoRestrictionType = "whitelist"
	GeoRestrictionTypeNone      GeoRestrictionType = "none"
)

// Values returns all possible values of GeoRestrictionType
func (g GeoRestrictionType) Values() []GeoRestrictionType {
	return []GeoRestrictionType{
		GeoRestrictionTypeBlacklist,
		GeoRestrictionTypeWhitelist,
		GeoRestrictionTypeNone,
	}
}

type TenantConfig struct {
	ParameterDefinitions []ParameterDefinition
}

type ParameterDefinition struct {
	Definition *ParameterDefinitionSchema // required
	Name       *string                    // required
}

type ParameterDefinitionSchema struct {
	StringSchema *StringSchemaConfig
}

type StringSchemaConfig struct {
	Comment      *string
	DefaultValue *string
	Required     *bool // required
}

type ViewerCertificate struct {
	ACMCertificateArn            *string
	Certificate                  *string           // deprecated
	CertificateSource            CertificateSource // deprecated
	CloudFrontDefaultCertificate *bool
	IAMCertificateId             *string
	MinimumProtocolVersion       MinimumProtocolVersion
	SSLSupportMethod             SSLSupportMethod
}

type CertificateSource string

const (
	CertificateSourceCloudfront CertificateSource = "cloudfront"
	CertificateSourceIam        CertificateSource = "iam"
	CertificateSourceAcm        CertificateSource = "acm"
)

// Values returns all possible values of CertificateSource
func (c CertificateSource) Values() []CertificateSource {
	return []CertificateSource{
		CertificateSourceCloudfront,
		CertificateSourceIam,
		CertificateSourceAcm,
	}
}

type MinimumProtocolVersion string

const (
	MinimumProtocolVersionSSLv3      MinimumProtocolVersion = "SSLv3"
	MinimumProtocolVersionTLSv1      MinimumProtocolVersion = "TLSv1"
	MinimumProtocolVersionTLSv12016  MinimumProtocolVersion = "TLSv1_2016"
	MinimumProtocolVersionTLSv112016 MinimumProtocolVersion = "TLSv1.1_2016"
	MinimumProtocolVersionTLSv122018 MinimumProtocolVersion = "TLSv1.2_2018"
	MinimumProtocolVersionTLSv122019 MinimumProtocolVersion = "TLSv1.2_2019"
	MinimumProtocolVersionTLSv122021 MinimumProtocolVersion = "TLSv1.2_2021"
	MinimumProtocolVersionTLSv132025 MinimumProtocolVersion = "TLSv1.3_2025"
)

// Values returns all possible values of MinimumProtocolVersion
func (m MinimumProtocolVersion) Values() []MinimumProtocolVersion {
	return []MinimumProtocolVersion{
		MinimumProtocolVersionSSLv3,
		MinimumProtocolVersionTLSv1,
		MinimumProtocolVersionTLSv12016,
		MinimumProtocolVersionTLSv112016,
		MinimumProtocolVersionTLSv122018,
		MinimumProtocolVersionTLSv122019,
		MinimumProtocolVersionTLSv122021,
		MinimumProtocolVersionTLSv132025,
	}
}

type SSLSupportMethod string

const (
	SSLSupportMethodSniOnly  SSLSupportMethod = "sni-only"
	SSLSupportMethodVip      SSLSupportMethod = "vip"
	SSLSupportMethodStaticIp SSLSupportMethod = "static-ip"
)

// Values returns all possible values of SSLSupportMethod
func (s SSLSupportMethod) Values() []SSLSupportMethod {
	return []SSLSupportMethod{
		SSLSupportMethodSniOnly,
		SSLSupportMethodVip,
		SSLSupportMethodStaticIp,
	}
}

// Wrapper types for testing
type testDistributionConfigModel struct {
	Origins *Origins
}

// Terraform models
type multiTenantDistributionResourceModel struct {
	ARN                           types.String                                             `tfsdk:"arn"`
	DistributionConfig            fwtypes.ListNestedObjectValueOf[distributionConfigModel] `tfsdk:"distribution_config" `
	DomainName                    types.String                                             `tfsdk:"domain_name"`
	ID                            types.String                                             `tfsdk:"id"`
	InProgressInvalidationBatches types.Int32                                              `tfsdk:"in_progress_invalidation_batches"`
	LastModifiedTime              types.String                                             `tfsdk:"last_modified_time"`
	Status                        types.String                                             `tfsdk:"status"`
}

type distributionConfigModel struct {
	CacheBehavior        fwtypes.ListNestedObjectValueOf[cacheBehaviorModel]        `tfsdk:"cache_behavior" autoflex:",wrapper=items"`
	CallerReference      types.String                                               `tfsdk:"caller_reference"`
	Comment              types.String                                               `tfsdk:"comment"`
	CustomErrorResponse  fwtypes.SetNestedObjectValueOf[customErrorResponseModel]   `tfsdk:"custom_error_response" autoflex:",wrapper=items"`
	DefaultCacheBehavior fwtypes.ListNestedObjectValueOf[defaultCacheBehaviorModel] `tfsdk:"default_cache_behavior"`
	DefaultRootObject    types.String                                               `tfsdk:"default_root_object"`
	Enabled              types.Bool                                                 `tfsdk:"enabled"`
	HTTPVersion          fwtypes.StringEnum[HttpVersion]                            `tfsdk:"http_version"`
	ID                   types.String                                               `tfsdk:"id"`
	Origin               fwtypes.SetNestedObjectValueOf[originModel]                `tfsdk:"origin" autoflex:",wrapper=items"`
	OriginGroup          fwtypes.SetNestedObjectValueOf[originGroupModel]           `tfsdk:"origin_group" autoflex:",wrapper=items"`
	Restrictions         fwtypes.ListNestedObjectValueOf[restrictionsModel]         `tfsdk:"restrictions"`
	TenantConfig         fwtypes.ListNestedObjectValueOf[tenantConfigModel]         `tfsdk:"tenant_config"`
	ViewerCertificate    fwtypes.ListNestedObjectValueOf[viewerCertificateModel]    `tfsdk:"viewer_certificate"`
	WebACLID             types.String                                               `tfsdk:"web_acl_id"`
}

type originModel struct {
	ConnectionAttempts        types.Int32                                              `tfsdk:"connection_attempts"`
	ConnectionTimeout         types.Int32                                              `tfsdk:"connection_timeout"`
	CustomHeader              fwtypes.SetNestedObjectValueOf[customHeaderModel]        `tfsdk:"custom_header"`
	CustomOriginConfig        fwtypes.ListNestedObjectValueOf[customOriginConfigModel] `tfsdk:"custom_origin_config"`
	DomainName                types.String                                             `tfsdk:"domain_name"`
	ID                        types.String                                             `tfsdk:"id"`
	OriginAccessControlID     types.String                                             `tfsdk:"origin_access_control_id"`
	OriginPath                types.String                                             `tfsdk:"origin_path"`
	OriginShield              fwtypes.ListNestedObjectValueOf[originShieldModel]       `tfsdk:"origin_shield"`
	ResponseCompletionTimeout types.Int32                                              `tfsdk:"response_completion_timeout"`
	S3OriginConfig            fwtypes.ListNestedObjectValueOf[s3OriginConfigModel]     `tfsdk:"s3_origin_config"`
}

type customHeaderModel struct {
	HeaderName  types.String `tfsdk:"name"`
	HeaderValue types.String `tfsdk:"value"`
}

type customOriginConfigModel struct {
	HTTPPort               types.Int32                                         `tfsdk:"http_port"`
	HTTPSPort              types.Int32                                         `tfsdk:"https_port"`
	IPAddressType          fwtypes.StringEnum[IpAddressType]                   `tfsdk:"ip_address_type"`
	OriginKeepaliveTimeout types.Int32                                         `tfsdk:"origin_keepalive_timeout"`
	OriginReadTimeout      types.Int32                                         `tfsdk:"origin_read_timeout"`
	OriginProtocolPolicy   fwtypes.StringEnum[OriginProtocolPolicy]            `tfsdk:"origin_protocol_policy"`
	OriginSSLProtocols     fwtypes.SetValueOf[fwtypes.StringEnum[SslProtocol]] `tfsdk:"origin_ssl_protocols"`
}

type originShieldModel struct {
	Enabled            types.Bool   `tfsdk:"enabled"`
	OriginShieldRegion types.String `tfsdk:"origin_shield_region"`
}

type s3OriginConfigModel struct {
	OriginAccessIdentity types.String `tfsdk:"origin_access_identity"`
}

type originGroupModel struct {
	FailoverCriteria fwtypes.ListNestedObjectValueOf[failoverCriteriaModel] `tfsdk:"failover_criteria"`
	Member           fwtypes.ListNestedObjectValueOf[memberModel]           `tfsdk:"member"`
	OriginID         types.String                                           `tfsdk:"origin_id"`
}

type failoverCriteriaModel struct {
	StatusCodes fwtypes.SetValueOf[types.Int64] `tfsdk:"status_codes"`
}

type memberModel struct {
	OriginID types.String `tfsdk:"origin_id"`
}

type defaultCacheBehaviorModel struct {
	AllowedMethods            fwtypes.SetValueOf[fwtypes.StringEnum[Method]]                 `tfsdk:"allowed_methods"`
	CachePolicyID             types.String                                                   `tfsdk:"cache_policy_id"`
	CachedMethods             fwtypes.SetValueOf[fwtypes.StringEnum[Method]]                 `tfsdk:"cached_methods"`
	Compress                  types.Bool                                                     `tfsdk:"compress"`
	FieldLevelEncryptionID    types.String                                                   `tfsdk:"field_level_encryption_id"`
	FunctionAssociation       fwtypes.SetNestedObjectValueOf[functionAssociationModel]       `tfsdk:"function_association"`
	LambdaFunctionAssociation fwtypes.SetNestedObjectValueOf[lambdaFunctionAssociationModel] `tfsdk:"lambda_function_association"`
	OriginRequestPolicyID     types.String                                                   `tfsdk:"origin_request_policy_id"`
	RealtimeLogConfigARN      types.String                                                   `tfsdk:"realtime_log_config_arn"`
	ResponseHeadersPolicyID   types.String                                                   `tfsdk:"response_headers_policy_id"`
	TargetOriginID            types.String                                                   `tfsdk:"target_origin_id"`
	TrustedKeyGroups          fwtypes.ListNestedObjectValueOf[trustedKeyGroupsModel]         `tfsdk:"trusted_key_groups"`
	ViewerProtocolPolicy      fwtypes.StringEnum[ViewerProtocolPolicy]                       `tfsdk:"viewer_protocol_policy"`
	// Note: SmoothStreaming and TrustedSigners removed - not supported for multi-tenant distributions
}

type cacheBehaviorModel struct {
	AllowedMethods            fwtypes.SetValueOf[fwtypes.StringEnum[Method]]                 `tfsdk:"allowed_methods"`
	CachePolicyID             types.String                                                   `tfsdk:"cache_policy_id"`
	CachedMethods             fwtypes.SetValueOf[fwtypes.StringEnum[Method]]                 `tfsdk:"cached_methods"`
	Compress                  types.Bool                                                     `tfsdk:"compress"`
	FieldLevelEncryptionID    types.String                                                   `tfsdk:"field_level_encryption_id"`
	FunctionAssociation       fwtypes.SetNestedObjectValueOf[functionAssociationModel]       `tfsdk:"function_association"`
	LambdaFunctionAssociation fwtypes.SetNestedObjectValueOf[lambdaFunctionAssociationModel] `tfsdk:"lambda_function_association"`
	OriginRequestPolicyID     types.String                                                   `tfsdk:"origin_request_policy_id"`
	PathPattern               types.String                                                   `tfsdk:"path_pattern"`
	RealtimeLogConfigARN      types.String                                                   `tfsdk:"realtime_log_config_arn"`
	ResponseHeadersPolicyID   types.String                                                   `tfsdk:"response_headers_policy_id"`
	TargetOriginID            types.String                                                   `tfsdk:"target_origin_id"`
	TrustedKeyGroups          fwtypes.ListNestedObjectValueOf[trustedKeyGroupsModel]         `tfsdk:"trusted_key_groups"`
	ViewerProtocolPolicy      fwtypes.StringEnum[ViewerProtocolPolicy]                       `tfsdk:"viewer_protocol_policy"`
	// Note: SmoothStreaming and TrustedSigners removed - not supported for multi-tenant distributions
}

type customErrorResponseModel struct {
	ErrorCachingMinTtl types.Int64  `tfsdk:"error_caching_min_ttl"`
	ErrorCode          types.Int64  `tfsdk:"error_code"`
	ResponseCode       types.Int64  `tfsdk:"response_code"`
	ResponsePagePath   types.String `tfsdk:"response_page_path"`
}

type restrictionsModel struct {
	GeoRestriction fwtypes.ListNestedObjectValueOf[geoRestrictionModel] `tfsdk:"geo_restriction"`
}

type geoRestrictionModel struct {
	Items           fwtypes.SetValueOf[types.String]       `tfsdk:"items"`
	RestrictionType fwtypes.StringEnum[GeoRestrictionType] `tfsdk:"restriction_type"`
}

type viewerCertificateModel struct {
	ACMCertificateARN            types.String                               `tfsdk:"acm_certificate_arn"`
	CloudfrontDefaultCertificate types.Bool                                 `tfsdk:"cloudfront_default_certificate"`
	MinimumProtocolVersion       fwtypes.StringEnum[MinimumProtocolVersion] `tfsdk:"minimum_protocol_version"`
	SSLSupportMethod             fwtypes.StringEnum[SSLSupportMethod]       `tfsdk:"ssl_support_method"`
}

type functionAssociationModel struct {
	EventType   fwtypes.StringEnum[EventType] `tfsdk:"event_type"`
	FunctionArn types.String                  `tfsdk:"function_arn"`
}

type lambdaFunctionAssociationModel struct {
	EventType   fwtypes.StringEnum[EventType] `tfsdk:"event_type"`
	IncludeBody types.Bool                    `tfsdk:"include_body"`
	LambdaARN   types.String                  `tfsdk:"lambda_arn"`
}

type tenantConfigModel struct {
	ParameterDefinition fwtypes.ListNestedObjectValueOf[parameterDefinitionModel] `tfsdk:"parameter_definition"`
}

type parameterDefinitionModel struct {
	Name       types.String                                                    `tfsdk:"name"`
	Definition fwtypes.ListNestedObjectValueOf[parameterDefinitionSchemaModel] `tfsdk:"definition"`
}

type parameterDefinitionSchemaModel struct {
	StringSchema fwtypes.ListNestedObjectValueOf[stringSchemaConfigModel] `tfsdk:"string_schema"`
}

type stringSchemaConfigModel struct {
	Required     types.Bool   `tfsdk:"required"`
	Comment      types.String `tfsdk:"comment"`
	DefaultValue types.String `tfsdk:"default_value"`
}

type trustedKeyGroupsModel struct {
	Items   fwtypes.ListValueOf[types.String] `tfsdk:"items"`
	Enabled types.Bool                        `tfsdk:"enabled"`
}

func TestFlattenXMLWrapperRealWorld(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"two origins": {
			Source: &testDistributionConfigModel{
				Origins: &Origins{
					Items: []Origin{
						{DomainName: aws.String("example.com"), Id: aws.String("origin1")},
						{DomainName: aws.String("cdn.example.com"), Id: aws.String("origin2")},
					},
					Quantity: aws.Int32(2),
				},
			},
			Target: &distributionConfigModel{},
			WantTarget: &distributionConfigModel{
				Origin: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []originModel{
					{
						DomainName:         types.StringValue("example.com"),
						ID:                 types.StringValue("origin1"),
						CustomHeader:       fwtypes.NewSetNestedObjectValueOfNull[customHeaderModel](ctx),
						CustomOriginConfig: fwtypes.NewListNestedObjectValueOfNull[customOriginConfigModel](ctx),
						OriginShield:       fwtypes.NewListNestedObjectValueOfNull[originShieldModel](ctx),
						S3OriginConfig:     fwtypes.NewListNestedObjectValueOfNull[s3OriginConfigModel](ctx),
					},
					{
						DomainName:         types.StringValue("cdn.example.com"),
						ID:                 types.StringValue("origin2"),
						CustomHeader:       fwtypes.NewSetNestedObjectValueOfNull[customHeaderModel](ctx),
						CustomOriginConfig: fwtypes.NewListNestedObjectValueOfNull[customOriginConfigModel](ctx),
						OriginShield:       fwtypes.NewListNestedObjectValueOfNull[originShieldModel](ctx),
						S3OriginConfig:     fwtypes.NewListNestedObjectValueOfNull[s3OriginConfigModel](ctx),
					},
				}),
			},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, GoldenLogs: true})
}

// Test XML wrapper expansion for direct struct (not pointer to struct)
type DirectXMLWrapper struct {
	Items    []string
	Quantity *int32
}

type DirectWrapperTF struct {
	Items fwtypes.SetValueOf[types.String] `tfsdk:"items" autoflex:",xmlwrapper=items"`
}

type DirectWrapperAWS struct {
	Items DirectXMLWrapper
}

func TestExpandXMLWrapperDirect(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"direct xml wrapper": {
			Source: DirectWrapperTF{
				Items: fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("item1"),
					types.StringValue("item2"),
				}),
			},
			Target: &DirectWrapperAWS{},
			WantTarget: &DirectWrapperAWS{
				Items: DirectXMLWrapper{
					Items:    []string{"item1", "item2"},
					Quantity: aws.Int32(2),
				},
			},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, GoldenLogs: true})
}

func TestIsXMLWrapperStruct(t *testing.T) {
	t.Parallel()

	type embedWithField struct {
		Count int64
	}
	type embedWithoutField struct{}

	testCases := []struct {
		name     string
		input    any
		expected bool
	}{
		{
			name:     "valid XML wrapper",
			input:    FunctionAssociations{},
			expected: true,
		},
		{
			name:     "valid XML wrapper with slice of strings",
			input:    DirectXMLWrapper{},
			expected: true,
		},
		{
			name: "valid XML wrapper with anonymous struct",
			input: struct {
				Items    []string
				Quantity *int32
			}{},
			expected: true,
		},
		{
			name:     "not a struct",
			input:    "string",
			expected: false,
		},
		{
			name:     "struct without Items field",
			input:    struct{ Quantity *int32 }{},
			expected: false,
		},
		{
			name:     "struct without Quantity field",
			input:    struct{ Items []string }{},
			expected: false,
		},
		{
			name: "struct with wrong Quantity type",
			input: struct {
				Items    []string
				Quantity int32
			}{},
			expected: false,
		},
		{
			name: "struct with Items not a slice",
			input: struct {
				Items    string
				Quantity *int32
			}{},
			expected: false,
		},
		{
			name: "struct with extra field",
			input: struct {
				Items    []string
				Quantity *int32
				Name     string
			}{},
			expected: false,
		},
		{
			name: "struct with anonymous embedWithField",
			input: struct {
				Items    []string
				Quantity *int32
				embedWithField
			}{},
			expected: true,
		},
		{
			name: "struct with anonymous embedWithoutField",
			input: struct {
				Items    []string
				Quantity *int32
				embedWithoutField
			}{},
			expected: true,
		},
		{
			name: "struct with private embedWithField",
			input: struct {
				Items    []string
				Quantity *int32
				private  embedWithField
			}{},
			expected: false,
		},
		{
			name: "struct with private embedWithoutField",
			input: struct {
				Items    []string
				Quantity *int32
				private  embedWithoutField
			}{},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := isXMLWrapperStruct(reflect.TypeOf(tc.input))
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for input %T", tc.expected, result, tc.input)
			}
		})
	}
}

// Mock AWS types with XML wrapper pattern (for flattening - AWS to TF)
type awsStatusCodesForFlatten struct {
	Items    []int32
	Quantity *int32
}

type awsHeadersForFlatten struct {
	Items    []string
	Quantity *int32
}

// TF model types with wrapper tags (for flattening - AWS to TF)
type tfStatusCodesModelForFlatten struct {
	StatusCodes fwtypes.SetValueOf[types.Int64] `tfsdk:"status_codes" autoflex:",xmlwrapper=items"`
}

type tfHeadersModelForFlatten struct {
	Headers fwtypes.ListValueOf[types.String] `tfsdk:"headers" autoflex:",xmlwrapper=items"`
}

func TestFlattenXMLWrapper(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"int32 slice to set": {
			Source: awsStatusCodesForFlatten{
				Items:    []int32{400, 404},
				Quantity: aws.Int32(2),
			},
			Target: &tfStatusCodesModelForFlatten{},
			WantTarget: &tfStatusCodesModelForFlatten{
				StatusCodes: fwtypes.NewSetValueOfMust[types.Int64](ctx, []attr.Value{
					types.Int64Value(400),
					types.Int64Value(404),
				}),
			},
		},
		"string slice to list": {
			Source: awsHeadersForFlatten{
				Items:    []string{"accept", "content-type"},
				Quantity: aws.Int32(2),
			},
			Target: &tfHeadersModelForFlatten{},
			WantTarget: &tfHeadersModelForFlatten{
				Headers: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("accept"),
					types.StringValue("content-type"),
				}),
			},
		},
		"complex type - function associations": {
			Source: FunctionAssociations{
				Items: []FunctionAssociation{
					{
						EventType:   "viewer-request",
						FunctionArn: aws.String("arn:aws:cloudfront::123456789012:function/example-function"),
					},
					{
						EventType:   "viewer-response",
						FunctionArn: aws.String("arn:aws:cloudfront::123456789012:function/another-function"),
					},
				},
				Quantity: aws.Int32(2),
			},
			Target: &DistributionConfigTF{},
			WantTarget: &DistributionConfigTF{
				FunctionAssociations: fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, []*FunctionAssociationTF{
					{
						EventType:   types.StringValue("viewer-request"),
						FunctionArn: types.StringValue("arn:aws:cloudfront::123456789012:function/example-function"),
					},
					{
						EventType:   types.StringValue("viewer-response"),
						FunctionArn: types.StringValue("arn:aws:cloudfront::123456789012:function/another-function"),
					},
				}),
			},
		},
		"empty slice to null set": {
			Source: awsStatusCodesForFlatten{
				Items:    nil,
				Quantity: aws.Int32(0),
			},
			Target: &tfStatusCodesModelForFlatten{},
			WantTarget: &tfStatusCodesModelForFlatten{
				StatusCodes: fwtypes.NewSetValueOfNull[types.Int64](ctx),
			},
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, GoldenLogs: true})
}

type FunctionAssociationsTF struct {
	Items    fwtypes.ListNestedObjectValueOf[FunctionAssociationTF] `tfsdk:"items"`
	Quantity types.Int64                                            `tfsdk:"quantity"`
}

type DistributionConfigTFNoXMLWrapper struct {
	FunctionAssociations fwtypes.ListNestedObjectValueOf[FunctionAssociationsTF] `tfsdk:"function_associations"`
}

func TestExpandNoXMLWrapper(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"valid function associations": {
			Source: DistributionConfigTFNoXMLWrapper{
				FunctionAssociations: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &FunctionAssociationsTF{
					Items: fwtypes.NewListNestedObjectValueOfSliceMust(
						ctx,
						[]*FunctionAssociationTF{
							{
								EventType:   types.StringValue("viewer-request"),
								FunctionArn: types.StringValue("arn:aws:cloudfront::123456789012:function/test-function-1"),
							},
							{
								EventType:   types.StringValue("viewer-response"),
								FunctionArn: types.StringValue("arn:aws:cloudfront::123456789012:function/test-function-2"),
							},
						},
					),
					Quantity: types.Int64Value(2),
				}),
			},
			Target: &DistributionConfigAWS{},
			WantTarget: &DistributionConfigAWS{
				FunctionAssociations: &FunctionAssociations{
					Quantity: aws.Int32(2),
					Items: []FunctionAssociation{
						{
							EventType:   "viewer-request",
							FunctionArn: aws.String("arn:aws:cloudfront::123456789012:function/test-function-1"),
						},
						{
							EventType:   "viewer-response",
							FunctionArn: aws.String("arn:aws:cloudfront::123456789012:function/test-function-2"),
						},
					},
				},
			},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, GoldenLogs: true})
}

func TestFlattenNoXMLWrapper(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"complex type - function associations": {
			Source: DistributionConfigAWS{
				FunctionAssociations: &FunctionAssociations{
					Quantity: aws.Int32(2),
					Items: []FunctionAssociation{
						{
							EventType:   "viewer-request",
							FunctionArn: aws.String("arn:aws:cloudfront::123456789012:function/test-function-1"),
						},
						{
							EventType:   "viewer-response",
							FunctionArn: aws.String("arn:aws:cloudfront::123456789012:function/test-function-2"),
						},
					},
				},
			},
			// 	Items: []FunctionAssociation{
			// 		{
			// 			EventType:   "viewer-request",
			// 			FunctionArn: aws.String("arn:aws:cloudfront::123456789012:function/example-function"),
			// 		},
			// 		{
			// 			EventType:   "viewer-response",
			// 			FunctionArn: aws.String("arn:aws:cloudfront::123456789012:function/another-function"),
			// 		},
			// 	},
			// 	Quantity: aws.Int32(2),
			// },
			Target: &DistributionConfigTFNoXMLWrapper{},
			WantTarget: &DistributionConfigTFNoXMLWrapper{
				FunctionAssociations: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &FunctionAssociationsTF{
					Items: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*FunctionAssociationTF{
						{
							EventType:   types.StringValue("viewer-request"),
							FunctionArn: types.StringValue("arn:aws:cloudfront::123456789012:function/test-function-1"),
						},
						{
							EventType:   types.StringValue("viewer-response"),
							FunctionArn: types.StringValue("arn:aws:cloudfront::123456789012:function/test-function-2"),
						},
					}),
					Quantity: types.Int64Value(2),
				}),
			},
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, GoldenLogs: true})
}
