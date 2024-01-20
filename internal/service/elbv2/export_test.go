// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

// Exports for use in tests only.
var (
	AwsSdkId                      = awsSdkId // nosemgrep:ci.aws-in-var-name
	FindListenerByARN             = findListenerByARN
	HealthCheckProtocolEnumValues = healthCheckProtocolEnumValues
	ProtocolVersionEnumValues     = protocolVersionEnumValues
)
