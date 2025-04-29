// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

// Exports for use in tests only.
var (
	ResourceCatalogDatabase               = resourceCatalogDatabase
	ResourceCatalogTable                  = resourceCatalogTable
	ResourceCatalogTableOptimizer         = newCatalogTableOptimizerResource
	ResourceClassifier                    = resourceClassifier
	ResourceConnection                    = resourceConnection
	ResourceCrawler                       = resourceCrawler
	ResourceDataCatalogEncryptionSettings = resourceDataCatalogEncryptionSettings
	ResourceDataQualityRuleset            = resourceDataQualityRuleset
	ResourceDevEndpoint                   = resourceDevEndpoint
	ResourceJob                           = resourceJob
	ResourceMLTransform                   = resourceMLTransform
	ResourcePartition                     = resourcePartition
	ResourcePartitionIndex                = resourcePartitionIndex
	ResourceRegistry                      = resourceRegistry
	ResourceResourcePolicy                = resourceResourcePolicy
	ResourceSchema                        = resourceSchema
	ResourceSecurityConfiguration         = resourceSecurityConfiguration
	ResourceTrigger                       = resourceTrigger
	ResourceUserDefinedFunction           = resourceUserDefinedFunction
	ResourceWorkflow                      = resourceWorkflow

	FindCatalogTableOptimizer    = findCatalogTableOptimizer
	FindClassifierByName         = findClassifierByName
	FindConnectionByTwoPartKey   = findConnectionByTwoPartKey
	FindCrawlerByName            = findCrawlerByName
	FindDatabaseByName           = findDatabaseByName
	FindDataQualityRulesetByName = findDataQualityRulesetByName
	FindDevEndpointByName        = findDevEndpointByName
	FindJobByName                = findJobByName
	FindPartitionByValues        = findPartitionByValues
	FindPartitionIndexByName     = findPartitionIndexByName
	FindRegistryByID             = findRegistryByID
	FindSchemaByID               = findSchemaByID
	FindTableByName              = findTableByName
	FindTriggerByName            = findTriggerByName
)
