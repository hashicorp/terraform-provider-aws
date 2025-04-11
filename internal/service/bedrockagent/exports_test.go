// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

// Exports for use in tests only.
var (
	ResourceAgent                         = newAgentResource
	ResourceAgentActionGroup              = newAgentActionGroupResource
	ResourceAgentAlias                    = newAgentAliasResource
	ResourceAgentCollaborator             = newAgentCollaboratorResource
	ResourceAgentKnowledgeBaseAssociation = newAgentKnowledgeBaseAssociationResource
	ResourceDataSource                    = newDataSourceResource
	ResourceKnowledgeBase                 = newKnowledgeBaseResource
	ResourcePrompt                        = newResourcePrompt

	FindAgentByID                                  = findAgentByID
	FindAgentActionGroupByThreePartKey             = findAgentActionGroupByThreePartKey
	FindAgentAliasByTwoPartKey                     = findAgentAliasByTwoPartKey
	FindAgentCollaboratorByThreePartKey            = findAgentCollaboratorByThreePartKey
	FindAgentKnowledgeBaseAssociationByThreePartID = findAgentKnowledgeBaseAssociationByThreePartKey
	FindDataSourceByTwoPartKey                     = findDataSourceByTwoPartKey
	FindKnowledgeBaseByID                          = findKnowledgeBaseByID
	FindPromptByID                                 = findPromptByID
)
