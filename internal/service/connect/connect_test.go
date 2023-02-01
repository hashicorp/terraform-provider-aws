package connect_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

// Serialized acceptance tests due to Connect account limits (max 2 parallel tests)
func TestAccConnect_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"BotAssociation": {
			"basic":            testAccBotAssociation_basic,
			"disappears":       testAccBotAssociation_disappears,
			"dataSource_basic": testAccBotAssociationDataSource_basic,
		},
		"ContactFlow": {
			"basic":           testAccContactFlow_basic,
			"disappears":      testAccContactFlow_disappears,
			"filename":        testAccContactFlow_filename,
			"dataSource_id":   testAccContactFlowDataSource_contactFlowID,
			"dataSource_name": testAccContactFlowDataSource_name,
		},
		"ContactFlowModule": {
			"basic":           testAccContactFlowModule_basic,
			"disappears":      testAccContactFlowModule_disappears,
			"filename":        testAccContactFlowModule_filename,
			"dataSource_id":   testAccContactFlowModuleDataSource_contactFlowModuleID,
			"dataSource_name": testAccContactFlowModuleDataSource_name,
		},
		"HoursOfOperation": {
			"basic":           testAccHoursOfOperation_basic,
			"disappears":      testAccHoursOfOperation_disappears,
			"tags":            testAccHoursOfOperation_updateTags,
			"config":          testAccHoursOfOperation_updateConfig,
			"dataSource_id":   testAccHoursOfOperationDataSource_hoursOfOperationID,
			"dataSource_name": testAccHoursOfOperationDataSource_name,
		},
		"Instance": {
			"basic":            testAccInstance_basic,
			"directory":        testAccInstance_directory,
			"saml":             testAccInstance_saml,
			"dataSource_basic": testAccInstanceDataSource_basic,
		},
		"InstanceStorageConfig": {
			"basic":                                     testAccInstanceStorageConfig_basic,
			"disappears":                                testAccInstanceStorageConfig_disappears,
			"KinesisFirehoseConfig_FirehoseARN":         testAccInstanceStorageConfig_KinesisFirehoseConfig_FirehoseARN,
			"KinesisStreamConfig_StreamARN":             testAccInstanceStorageConfig_KinesisStreamConfig_StreamARN,
			"KinesisVideoStreamConfig_EncryptionConfig": testAccInstanceStorageConfig_KinesisVideoStreamConfig_EncryptionConfig,
			"KinesisVideoStreamConfig_Prefix":           testAccInstanceStorageConfig_KinesisVideoStreamConfig_Prefix,
			"KinesisVideoStreamConfig_Retention":        testAccInstanceStorageConfig_KinesisVideoStreamConfig_Retention,
			"S3Config_BucketName":                       testAccInstanceStorageConfig_S3Config_BucketName,
			"S3Config_BucketPrefix":                     testAccInstanceStorageConfig_S3Config_BucketPrefix,
			"S3Config_EncryptionConfig":                 testAccInstanceStorageConfig_S3Config_EncryptionConfig,
			"dataSource_KinesisFirehoseConfig":          testAccInstanceStorageConfigDataSource_KinesisFirehoseConfig,
			"dataSource_KinesisStreamConfig":            testAccInstanceStorageConfigDataSource_KinesisStreamConfig,
			"dataSource_KinesisVideoStreamConfig":       testAccInstanceStorageConfigDataSource_KinesisVideoStreamConfig,
			"dataSource_S3Config":                       testAccInstanceStorageConfigDataSource_S3Config,
		},
		"LambdaFunctionAssociation": {
			"basic":            testAccLambdaFunctionAssociation_basic,
			"disappears":       testAccLambdaFunctionAssociation_disappears,
			"dataSource_basic": testAccLambdaFunctionAssociationDataSource_basic,
		},
		"PhoneNumber": {
			"basic":       testAccPhoneNumber_basic,
			"disappears":  testAccPhoneNumber_disappears,
			"tags":        testAccPhoneNumber_tags,
			"description": testAccPhoneNumber_description,
			"prefix":      testAccPhoneNumber_prefix,
			"targetARN":   testAccPhoneNumber_targetARN,
		},
		"Prompt": {
			"dataSource_name": testAccPromptDataSource_name,
		},
		"Queue": {
			"basic":                testAccQueue_basic,
			"disappears":           testAccQueue_disappears,
			"tags":                 testAccQueue_updateTags,
			"hoursOfOperationId":   testAccQueue_updateHoursOfOperationId,
			"maxContacts":          testAccQueue_updateMaxContacts,
			"outboundCallerConfig": testAccQueue_updateOutboundCallerConfig,
			"status":               testAccQueue_updateStatus,
			"quickConnectIds":      testAccQueue_updateQuickConnectIds,
			"dataSource_id":        testAccQueueDataSource_queueID,
			"dataSource_name":      testAccQueueDataSource_name,
		},
		"QuickConnect": {
			"basic":           testAccQuickConnect_phoneNumber,
			"disappears":      testAccQuickConnect_disappears,
			"tags":            testAccQuickConnect_updateTags,
			"dataSource_id":   testAccQuickConnectDataSource_id,
			"dataSource_name": testAccQuickConnectDataSource_name,
		},
		"RoutingProfile": {
			"basic":                testAccRoutingProfile_basic,
			"disappears":           testAccRoutingProfile_disappears,
			"tags":                 testAccRoutingProfile_updateTags,
			"concurrency":          testAccRoutingProfile_updateConcurrency,
			"defaultOutboundQueue": testAccRoutingProfile_updateDefaultOutboundQueue,
			"queues":               testAccRoutingProfile_updateQueues,
			"dataSource_id":        testAccRoutingProfileDataSource_routingProfileID,
			"dataSource_name":      testAccRoutingProfileDataSource_name,
		},
		"SecurityProfile": {
			"basic":           testAccSecurityProfile_basic,
			"disappears":      testAccSecurityProfile_disappears,
			"tags":            testAccSecurityProfile_updateTags,
			"permissions":     testAccSecurityProfile_updatePermissions,
			"dataSource_id":   testAccSecurityProfileDataSource_securityProfileID,
			"dataSource_name": testAccSecurityProfileDataSource_name,
		},
		"User": {
			"basic":              testAccUser_basic,
			"disappears":         testAccUser_disappears,
			"tags":               testAccUser_updateTags,
			"hierarchyGroupId":   testAccUser_updateHierarchyGroupId,
			"identityInfo":       testAccUser_updateIdentityInfo,
			"phoneConfig":        testAccUser_updatePhoneConfig,
			"routingProfileId":   testAccUser_updateRoutingProfileId,
			"securityProfileIds": testAccUser_updateSecurityProfileIds,
		},
		"UserHierarchyGroup": {
			"basic":           testAccUserHierarchyGroup_basic,
			"disappears":      testAccUserHierarchyGroup_disappears,
			"updateTags":      testAccUserHierarchyGroup_updateTags,
			"parentGroupId":   testAccUserHierarchyGroup_parentGroupId,
			"dataSource_id":   testAccUserHierarchyGroupDataSource_hierarchyGroupID,
			"dataSource_name": testAccUserHierarchyGroupDataSource_name,
		},
		"UserHierarchyStructure": {
			"basic":         testAccUserHierarchyStructure_basic,
			"disappears":    testAccUserHierarchyStructure_disappears,
			"dataSource_id": testAccUserHierarchyStructureDataSource_instanceID,
		},
		"Vocabulary": {
			"basic":      testAccVocabulary_basic,
			"disappears": testAccVocabulary_disappears,
			"tags":       testAccVocabulary_updateTags,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
