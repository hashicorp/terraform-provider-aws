// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

// Exports for use in tests only.
var (
	ResourceGateway       = newResourceGateway
	ResourceGatewayTarget = newResourceGatewayTarget

	FindGatewayByID       = findGatewayByID
	FindGatewayTargetByID = findGatewayTargetByID
)
