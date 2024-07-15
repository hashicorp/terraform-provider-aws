// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"time"

	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
)

const (
	iamPropagationTimeout   = 2 * time.Minute
	elbv2PropagationTimeout = 5 * time.Minute // nosemgrep:ci.elbv2-in-const-name, ci.elbv2-in-var-name
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
	loadBalancerAttributeClientKeepAliveSeconds                          = "client_keep_alive.seconds"
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

// See https://docs.aws.amazon.com/elasticloadbalancing/latest/APIReference/API_TargetGroupAttribute.html#API_TargetGroupAttribute_Contents.
const (
	// The following attributes are supported by all load balancers:
	targetGroupAttributeDeregistrationDelayTimeoutSeconds = "deregistration_delay.timeout_seconds"
	targetGroupAttributeStickinessEnabled                 = "stickiness.enabled"
	targetGroupAttributeStickinessType                    = "stickiness.type"

	// The following attributes are supported by Application Load Balancers and Network Load Balancers:
	targetGroupAttributeLoadBalancingCrossZoneEnabled                                         = "load_balancing.cross_zone.enabled"
	targetGroupAttributeTargetGroupHealthDNSFailoverMinimumHealthyTargetsCount                = "target_group_health.dns_failover.minimum_healthy_targets.count"
	targetGroupAttributeTargetGroupHealthDNSFailoverMinimumHealthyTargetsPercentage           = "target_group_health.dns_failover.minimum_healthy_targets.percentage"
	targetGroupAttributeTargetGroupHealthUnhealthyStateRoutingMinimumHealthyTargetsCount      = "target_group_health.unhealthy_state_routing.minimum_healthy_targets.count"
	targetGroupAttributeTargetGroupHealthUnhealthyStateRoutingMinimumHealthyTargetsPercentage = "target_group_health.unhealthy_state_routing.minimum_healthy_targets.percentage"

	// The following attributes are supported only if the load balancer is an Application Load Balancer and the target is an instance or an IP address:
	targetGroupAttributeLoadBalancingAlgorithmType              = "load_balancing.algorithm.type"
	targetGroupAttributeLoadBalancingAlgorithmAnomalyMitigation = "load_balancing.algorithm.anomaly_mitigation"
	targetGroupAttributeSlowStartDurationSeconds                = "slow_start.duration_seconds"
	targetGroupAttributeStickinessAppCookieCookieName           = "stickiness.app_cookie.cookie_name"
	targetGroupAttributeStickinessAppCookieDurationSeconds      = "stickiness.app_cookie.duration_seconds"
	targetGroupAttributeStickinessLBCookieDurationSeconds       = "stickiness.lb_cookie.duration_seconds"

	// The following attribute is supported only if the load balancer is an Application Load Balancer and the target is a Lambda function:
	targetGroupAttributeLambdaMultiValueHeadersEnabled = "lambda.multi_value_headers.enabled"

	// The following attributes are supported only by Network Load Balancers:
	targetGroupAttributeDeregistrationDelayConnectionTerminationEnabled        = "deregistration_delay.connection_termination.enabled"
	targetGroupAttributePreserveClientIPEnabled                                = "preserve_client_ip.enabled"
	targetGroupAttributeProxyProtocolV2Enabled                                 = "proxy_protocol_v2.enabled"
	targetGroupAttributeTargetHealthStateUnhealthyConnectionTerminationEnabled = "target_health_state.unhealthy.connection_termination.enabled"

	// The following attributes are supported only by Gateway Load Balancers:
	targetGroupAttributeTargetFailoverOnDeregistration = "target_failover.on_deregistration"
	targetGroupAttributeTargetFailoverOnUnhealthy      = "target_failover.on_unhealthy"
)

const (
	loadBalancingAlgorithmTypeRoundRobin               = "round_robin"
	loadBalancingAlgorithmTypeLeastOutstandingRequests = "least_outstanding_requests"
	loadBalancingAlgorithmTypeWeightedRandom           = "weighted_random"
)

func loadBalancingAlgorithmType_Values() []string {
	return []string{
		loadBalancingAlgorithmTypeRoundRobin,
		loadBalancingAlgorithmTypeLeastOutstandingRequests,
		loadBalancingAlgorithmTypeWeightedRandom,
	}
}

const (
	loadBalancingAnomalyMitigationOn  = "on"
	loadBalancingAnomalyMitigationOff = "off"
)

func loadBalancingAnomalyMitigationType_Values() []string {
	return []string{
		loadBalancingAnomalyMitigationOn,
		loadBalancingAnomalyMitigationOff,
	}
}

const (
	loadBalancingCrossZoneEnabledTrue                         = "true"
	loadBalancingCrossZoneEnabledFalse                        = "false"
	loadBalancingCrossZoneEnabledUseLoadBalancerConfiguration = "use_load_balancer_configuration"
)

func loadBalancingCrossZoneEnabled_Values() []string {
	return []string{
		loadBalancingCrossZoneEnabledTrue,
		loadBalancingCrossZoneEnabledFalse,
		loadBalancingCrossZoneEnabledUseLoadBalancerConfiguration,
	}
}

const (
	stickinessTypeLBCookie            = "lb_cookie"               // Only for ALBs
	stickinessTypeAppCookie           = "app_cookie"              // Only for ALBs
	stickinessTypeSourceIP            = "source_ip"               // Only for NLBs
	stickinessTypeSourceIPDestIP      = "source_ip_dest_ip"       // Only for GWLBs
	stickinessTypeSourceIPDestIPProto = "source_ip_dest_ip_proto" // Only for GWLBs
)

func stickinessType_Values() []string {
	return []string{
		stickinessTypeLBCookie,
		stickinessTypeAppCookie,
		stickinessTypeSourceIP,
		stickinessTypeSourceIPDestIP,
		stickinessTypeSourceIPDestIPProto,
	}
}

const (
	targetFailoverRebalance   = "rebalance"
	targetFailoverNoRebalance = "no_rebalance"
)

func targetFailover_Values() []string {
	return []string{
		targetFailoverRebalance,
		targetFailoverNoRebalance,
	}
}

const (
	healthCheckPortTrafficPort = "traffic-port"
)

func healthCheckProtocolEnumValues() []string {
	return enum.Slice[awstypes.ProtocolEnum](
		awstypes.ProtocolEnumHttp,
		awstypes.ProtocolEnumHttps,
		awstypes.ProtocolEnumTcp,
	)
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
