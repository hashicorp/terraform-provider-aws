// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker

type (
	AlgorithmResourceModel                = algorithmResourceModel
	AlgorithmValidationSpecificationModel = algorithmValidationSpecificationModel
	AlgorithmValidationProfileModel       = algorithmValidationProfileModel
	TrainingJobDefinitionModel            = trainingJobDefinitionModel
	ChannelModel                          = channelModel
	DataSourceModel                       = dataSourceModel
	ShuffleConfigModel                    = shuffleConfigModel
	OutputDataConfigModel                 = outputDataConfigModel
	ResourceConfigModel                   = resourceConfigModel
	InstanceGroupModel                    = instanceGroupModel
	InstancePlacementConfigModel          = instancePlacementConfigModel
	StoppingConditionModel                = stoppingConditionModel
	TransformJobDefinitionModel           = transformJobDefinitionModel

	TrainingJobAlgorithmSpecificationModel = trainingJobAlgorithmSpecificationModel
	TrainingJobMetricDefinitionModel       = trainingJobMetricDefinitionModel
	TrainingJobModelPackageConfigModel     = trainingJobModelPackageConfigModel
	TrainingJobTrainingImageConfigModel    = trainingJobTrainingImageConfigModel
	TrainingJobServerlessJobConfigModel    = trainingJobServerlessJobConfigModel
	TrainingJobStoppingConditionModel      = trainingJobStoppingConditionModel
	TrainingJobVPCConfigModel              = trainingJobVPCConfigModel

	HyperParameterTrainingJobDefinitionModel   = hyperParameterTrainingJobDefinitionModel
	HyperParameterAlgorithmSpecificationModel  = algorithmSpecificationModel
	HyperParameterMetricDefinitionModel        = hyperParameterTuningMetricDefinitionModel
	CheckpointConfigModel                      = checkpointConfigModel
	ParameterRangesModel                       = parameterRangesModel
	HyperParameterInputDataConfigModel         = inputDataConfigModel
	HyperParameterDataSourceModel              = hyperParameterTuningDataSourceModel
	HyperParameterFileSystemDataSourceModel    = hyperParameterTuningFileSystemDataSourceModel
	HyperParameterHubAccessConfigModel         = hyperParameterTuningHubAccessConfigModel
	HyperParameterModelAccessConfigModel       = hyperParameterTuningModelAccessConfigModel
	HyperParameterS3DataSourceModel            = s3DataSourceModel
	HyperParameterShuffleConfigModel           = hyperParameterTuningShuffleConfigModel
	HyperParameterOutputDataConfigModel        = hyperParameterTuningOutputDataConfigModel
	HyperParameterTrainingResourceConfigModel  = trainingResourceConfigModel
	HyperParameterTuningResourceConfigModel    = hyperParameterTuningResourceConfigModel
	HyperParameterInstanceConfigModel          = hyperParameterTuningInstanceConfigModel
	HyperParameterInstanceGroupModel           = hyperParameterTuningInstanceGroupModel
	HyperParameterInstancePlacementConfigModel = hyperParameterTuningInstancePlacementConfigModel
	RetryStrategyModel                         = retryStrategyModel
	HyperParameterStoppingConditionModel       = hyperParameterTuningStoppingConditionModel
	TuningObjectiveModel                       = tuningObjectiveModel
	HyperParameterTuningJobVPCConfigModel      = hyperParameterTuningJobVPCConfigModel
)

// Exports for use in tests only.
var (
	ResourceApp                                    = resourceApp
	ResourceAppImageConfig                         = resourceAppImageConfig
	ResourceAlgorithm                              = newAlgorithmResource
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
	ResourceHyperParameterTuningJob                = newHyperParameterTuningJobResource
	ResourceImage                                  = resourceImage
	ResourceLabelingJob                            = newLabelingJobResource
	ResourceImageVersion                           = resourceImageVersion
	ResourceMlflowApp                              = newMlflowAppResource
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
	ResourceTrainingJob                            = newResourceTrainingJob
	ResourceUserProfile                            = resourceUserProfile
	ResourceWorkforce                              = resourceWorkforce
	ResourceWorkteam                               = resourceWorkteam

	FindAppByName                             = findAppByName
	FindAlgorithmByName                       = findAlgorithmByName
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
	FindHyperParameterTuningJobByName         = findHyperParameterTuningJobByName
	FindHumanTaskUIByName                     = findHumanTaskUIByName
	FindImageByName                           = findImageByName
	FindImageVersionByTwoPartKey              = findImageVersionByTwoPartKey
	FindLabelingJobByName                     = findLabelingJobByName
	FindMlflowAppByARN                        = findMlflowAppByARN
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
	FindTrainingJobByName                     = findTrainingJobByName
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
	PreserveAlgorithmValidationSpecification       = preserveAlgorithmValidationSpecification

	NormalizeAlgoSpecMetricDefinitions = normalizeAlgoSpecMetricDefinitions
	NormalizeStoppingCondition         = normalizeStoppingCondition
	ServerlessJobConfigEqualityFunc    = serverlessJobConfigEqualityFunc

	RestoreTrainingJobDefinitionNoFlattenFields = restoreTrainingJobDefinitionNoFlattenFields

	ValidName   = validName
	ValidPrefix = validPrefix
)
