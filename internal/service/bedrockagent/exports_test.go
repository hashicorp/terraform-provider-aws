// Copyright IBM Corp. 2014, 2025
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
	ResourceFlow                          = newFlowResource
	ResourceKnowledgeBase                 = newKnowledgeBaseResource
	ResourcePrompt                        = newPromptResource

	FindAgentByID                                  = findAgentByID
	FindAgentActionGroupByThreePartKey             = findAgentActionGroupByThreePartKey
	FindAgentAliasByTwoPartKey                     = findAgentAliasByTwoPartKey
	FindAgentCollaboratorByThreePartKey            = findAgentCollaboratorByThreePartKey
	FindAgentKnowledgeBaseAssociationByThreePartID = findAgentKnowledgeBaseAssociationByThreePartKey
	FindDataSourceByTwoPartKey                     = findDataSourceByTwoPartKey
	FindFlowByID                                   = findFlowByID
	FindKnowledgeBaseByID                          = findKnowledgeBaseByID
	FindPromptByID                                 = findPromptByID
)
