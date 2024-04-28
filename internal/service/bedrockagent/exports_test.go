// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

// Exports for use in tests only.
var (
	ResourceAgent            = newAgentResource
	ResourceAgentActionGroup = newAgentActionGroupResource
	ResourceAgentAlias       = newAgentAliasResource
	ResourceKnowledgeBase    = newKnowledgeBaseResource
	ResourceDataSource       = newDataSourceResource

	FindAgentActionGroupByThreePartKey = findAgentActionGroupByThreePartKey
	FindAgentAliasByTwoPartKey         = findAgentAliasByTwoPartKey
	FindAgentByID                      = findAgentByID
	FindKnowledgeBaseByID              = findKnowledgeBaseByID
	FindDataSourceByID                 = findDataSourceByID
)
