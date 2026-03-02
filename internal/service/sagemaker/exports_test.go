// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker

// Exports for use in tests only.
var (
	ResourceApp                                    = resourceApp
	ResourceAppImageConfig                         = resourceAppImageConfig
	ResourceCodeRepository                         = resourceCodeRepository
	ResourceDataQualityJobDefinition               = resourceDataQualityJobDefinition
	ResourceDevice                                 = resourceDevice
	ResourceDeviceFleet                            = resourceDeviceFleet
	ResourceDomain                                 = resourceDomain
	ResourceEndpoint                               = resourceEndpoint
	ResourceEndpointConfiguration                  = resourceEndpointConfiguration
	ResourceFeatureGroup                           = resourceFeatureGroup
	ResourceFlowDefinition                         = resourceFlowDefinition
	ResourceHub                                    = resourceHub
	ResourceHumanTaskUI                            = resourceHumanTaskUI
	ResourceImage                                  = resourceImage
	ResourceLabelingJob                            = newLabelingJobResource
	ResourceImageVersion                           = resourceImageVersion
	ResourceMlflowTrackingServer                   = resourceMlflowTrackingServer
	ResourceModel                                  = resourceModel
	ResourceModelCard                              = newModelCardResource
	ResourceModelPackageGroup                      = resourceModelPackageGroup
	ResourceModelPackageGroupPolicy                = resourceModelPackageGroupPolicy
	ResourceMonitoringSchedule                     = resourceMonitoringSchedule
	ResourceNotebookInstance                       = resourceNotebookInstance
	ResourceNotebookInstanceLifeCycleConfiguration = resourceNotebookInstanceLifeCycleConfiguration
	ResourcePipeline                               = resourcePipeline
	ResourceProject                                = resourceProject
	ResourceSpace                                  = resourceSpace
	ResourceStudioLifecycleConfig                  = resourceStudioLifecycleConfig
	ResourceUserProfile                            = resourceUserProfile
	ResourceWorkforce                              = resourceWorkforce
	ResourceWorkteam                               = resourceWorkteam

	FindAppByName                             = findAppByName
	FindAppImageConfigByName                  = findAppImageConfigByName
	FindCodeRepositoryByName                  = findCodeRepositoryByName
	FindDataQualityJobDefinitionByName        = findDataQualityJobDefinitionByName
	FindDeviceByName                          = findDeviceByName
	FindDeviceFleetByName                     = findDeviceFleetByName
	FindDomainByName                          = findDomainByName
	FindEndpointByName                        = findEndpointByName
	FindEndpointConfigByName                  = findEndpointConfigByName
	FindFeatureGroupByName                    = findFeatureGroupByName
	FindFlowDefinitionByName                  = findFlowDefinitionByName
	FindHubByName                             = findHubByName
	FindHumanTaskUIByName                     = findHumanTaskUIByName
	FindImageByName                           = findImageByName
	FindImageVersionByTwoPartKey              = findImageVersionByTwoPartKey
	FindLabelingJobByName                     = findLabelingJobByName
	FindMlflowTrackingServerByName            = findMlflowTrackingServerByName
	FindModelByName                           = findModelByName
	FindModelCardByName                       = findModelCardByName
	FindModelCardExportJobByARN               = findModelCardExportJobByARN
	FindModelPackageGroupByName               = findModelPackageGroupByName
	FindModelPackageGroupPolicyByName         = findModelPackageGroupPolicyByName
	FindMonitoringScheduleByName              = findMonitoringScheduleByName
	FindNotebookInstanceByName                = findNotebookInstanceByName
	FindNotebookInstanceLifecycleConfigByName = findNotebookInstanceLifecycleConfigByName
	FindPipelineByName                        = findPipelineByName
	FindProjectByName                         = findProjectByName
	FindServicecatalogPortfolioStatus         = findServicecatalogPortfolioStatus
	FindSpaceByName                           = findSpaceByName
	FindStudioLifecycleConfigByName           = findStudioLifecycleConfigByName
	FindUserProfileByName                     = findUserProfileByName
	FindWorkforceByName                       = findWorkforceByName
	FindWorkteamByName                        = findWorkteamByName

	DecodeAppID                                    = decodeAppID
	DecodeDeviceId                                 = decodeDeviceId
	ImageVersionFromARN                            = imageVersionFromARN
	PrebuiltECRImageCreatePath                     = prebuiltECRImageCreatePath
	PrebuiltECRImageIDByRegion_factorMachines      = prebuiltECRImageIDByRegion_factorMachines
	PrebuiltECRImageIDByRegion_XGBoost             = prebuiltECRImageIDByRegion_XGBoost
	PrebuiltECRImageIDByRegion_clarify             = prebuiltECRImageIDByRegion_clarify
	PrebuiltECRImageIDByRegion_dataWrangler        = prebuiltECRImageIDByRegion_dataWrangler
	PrebuiltECRImageIDByRegion_debugger            = prebuiltECRImageIDByRegion_debugger
	PrebuiltECRImageIDByRegion_deepLearning        = prebuiltECRImageIDByRegion_deepLearning
	PrebuiltECRImageIDByRegion_inferentiaNeo       = prebuiltECRImageIDByRegion_inferentiaNeo
	PrebuiltECRImageIDByRegion_SageMakerBasePython = prebuiltECRImageIDByRegion_SageMakerBasePython // nosemgrep:ci.sagemaker-in-var-name
	PrebuiltECRImageIDByRegion_SageMakerCustom     = prebuiltECRImageIDByRegion_SageMakerCustom     // nosemgrep:ci.sagemaker-in-var-name
	PrebuiltECRImageIDByRegion_SageMakerRL         = prebuiltECRImageIDByRegion_SageMakerRL         // nosemgrep:ci.sagemaker-in-var-name
	PrebuiltECRImageIDByRegion_spark               = prebuiltECRImageIDByRegion_spark

	ValidName   = validName
	ValidPrefix = validPrefix
)
