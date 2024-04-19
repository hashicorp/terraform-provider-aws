// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

// Exports for use in tests only.
var (
	ResourceAgent      = newAgentResource
	ResourceAgentAlias = newAgentAliasResource

	FindAgentAliasByID = findAgentAliasByID
	FindAgentByID      = findAgentByID
)
