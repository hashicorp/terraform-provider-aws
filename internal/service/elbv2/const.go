// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"time"

	"github.com/aws/aws-sdk-go/service/elbv2"
)

const (
	propagationTimeout = 2 * time.Minute
)

const (
	errCodeValidationError = "ValidationError"

	tagsOnCreationErrMessage = "cannot specify tags on creation"
)

// See https://docs.aws.amazon.com/elasticloadbalancing/latest/APIReference/API_LoadBalancerAttribute.html#API_LoadBalancerAttribute_Contents.
const (
	// The following attributes are supported by all load balancers:
	loadBalancerAttributeDeletionProtectionEnabled     = "deletion_protection.enabled"
	loadBalancerAttributeLoadBalancingCrossZoneEnabled = "load_balancing.cross_zone.enabled"

	// The following attributes are supported by both Application Load Balancers and Network Load Balancers:
	loadBalancerAttributeAccessLogsS3Enabled   = "access_logs.s3.enabled"
	loadBalancerAttributeAccessLogsS3Bucket    = "access_logs.s3.bucket"
	loadBalancerAttributeAccessLogsS3Prefix    = "access_logs.s3.prefix"
	loadBalancerAttributeIPv6DenyAllIGWTraffic = "ipv6.deny_all_igw_traffic"

	// The following attributes are supported by only Application Load Balancers:
	loadBalancerAttributeIdleTimeoutTimeoutSeconds                       = "idle_timeout.timeout_seconds"
	loadBalancerAttributeConnectionLogsS3Enabled                         = "connection_logs.s3.enabled"
	loadBalancerAttributeConnectionLogsS3Bucket                          = "connection_logs.s3.bucket"
	loadBalancerAttributeConnectionLogsS3Prefix                          = "connection_logs.s3.prefix"
	loadBalancerAttributeRoutingHTTPDesyncMitigationMode                 = "routing.http.desync_mitigation_mode"
	loadBalancerAttributeRoutingHTTPDropInvalidHeaderFieldsEnabled       = "routing.http.drop_invalid_header_fields.enabled"
	loadBalancerAttributeRoutingHTTPPreserveHostHeaderEnabled            = "routing.http.preserve_host_header.enabled"
	loadBalancerAttributeRoutingHTTPXAmznTLSVersionAndCipherSuiteEnabled = "routing.http.x_amzn_tls_version_and_cipher_suite.enabled"
	loadBalancerAttributeRoutingHTTPXFFClientPortEnabled                 = "routing.http.xff_client_port.enabled"
	loadBalancerAttributeRoutingHTTPXFFHeaderProcessingMode              = "routing.http.xff_header_processing.mode"
	loadBalancerAttributeRoutingHTTP2Enabled                             = "routing.http2.enabled"
	loadBalancerAttributeWAFFailOpenEnabled                              = "waf.fail_open.enabled"

	// The following attributes are supported by only Network Load Balancers:
	loadBalancerAttributeDNSRecordClientRoutingPolicy = "dns_record.client_routing_policy"
)

const (
	httpDesyncMitigationModeMonitor   = "monitor"
	httpDesyncMitigationModeDefensive = "defensive"
	httpDesyncMitigationModeStrictest = "strictest"
)

func httpDesyncMitigationMode_Values() []string {
	return []string{
		httpDesyncMitigationModeMonitor,
		httpDesyncMitigationModeDefensive,
		httpDesyncMitigationModeStrictest,
	}
}

const (
	dnsRecordClientRoutingPolicyAvailabilityZoneAffinity        = "availability_zone_affinity"
	dnsRecordClientRoutingPolicyPartialAvailabilityZoneAffinity = "partial_availability_zone_affinity"
	dnsRecordClientRoutingPolicyAnyAvailabilityZone             = "any_availability_zone"
)

func dnsRecordClientRoutingPolicy_Values() []string {
	return []string{
		dnsRecordClientRoutingPolicyAvailabilityZoneAffinity,
		dnsRecordClientRoutingPolicyPartialAvailabilityZoneAffinity,
		dnsRecordClientRoutingPolicyAnyAvailabilityZone,
	}
}

const (
	httpXFFHeaderProcessingModeAppend   = "append"
	httpXFFHeaderProcessingModePreserve = "preserve"
	httpXFFHeaderProcessingModeRemove   = "remove"
)

func httpXFFHeaderProcessingMode_Values() []string {
	return []string{
		httpXFFHeaderProcessingModeAppend,
		httpXFFHeaderProcessingModePreserve,
		httpXFFHeaderProcessingModeRemove,
	}
}

const (
	healthCheckPortTrafficPort = "traffic-port"
)

func healthCheckProtocolEnumValues() []string {
	return []string{
		elbv2.ProtocolEnumHttp,
		elbv2.ProtocolEnumHttps,
		elbv2.ProtocolEnumTcp,
	}
}

const (
	protocolVersionGRPC  = "GRPC"
	protocolVersionHTTP1 = "HTTP1"
	protocolVersionHTTP2 = "HTTP2"
)

func protocolVersionEnumValues() []string {
	return []string{
		protocolVersionGRPC,
		protocolVersionHTTP1,
		protocolVersionHTTP2,
	}
}
