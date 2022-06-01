package s3_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccS3BucketReplicationConfiguration_basic(t *testing.T) {
	iamRoleResourceName := "aws_iam_role.test"
	dstBucketResourceName := "aws_s3_bucket.destination"
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_s3_bucket_replication_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroy, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationBasic(rName, s3.StorageClassStandard),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":                          "foobar",
						"prefix":                      "foo",
						"status":                      s3.ReplicationRuleStatusEnabled,
						"destination.#":               "1",
						"destination.0.storage_class": s3.StorageClassStandard,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, "arn"),
				),
			},
			{
				Config: testAccBucketReplicationConfigurationBasic(rName, s3.StorageClassGlacier),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":                          "foobar",
						"prefix":                      "foo",
						"status":                      s3.ReplicationRuleStatusEnabled,
						"destination.#":               "1",
						"destination.0.storage_class": s3.StorageClassGlacier,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, "arn"),
				),
			},
			{
				Config: testAccBucketReplicationConfigurationConfig_sseKMSEncryptedObjects(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":            "foobar",
						"prefix":        "foo",
						"status":        s3.ReplicationRuleStatusEnabled,
						"destination.#": "1",
						"destination.0.encryption_configuration.#":                       "1",
						"destination.0.storage_class":                                    s3.StorageClassStandard,
						"source_selection_criteria.#":                                    "1",
						"source_selection_criteria.0.sse_kms_encrypted_objects.#":        "1",
						"source_selection_criteria.0.sse_kms_encrypted_objects.0.status": s3.SseKmsEncryptedObjectsStatusEnabled,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, "arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.encryption_configuration.0.replica_kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccS3BucketReplicationConfiguration_disappears(t *testing.T) {
	resourceName := "aws_s3_bucket_replication_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationBasic(rName, s3.StorageClassStandard),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3.ResourceBucketReplicationConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketReplicationConfiguration_multipleDestinationsEmptyFilter(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_replication_configuration.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroy, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationWithMultipleDestinationsEmptyFilter(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":                          "rule1",
						"priority":                    "1",
						"status":                      s3.ReplicationRuleStatusEnabled,
						"filter.#":                    "1",
						"filter.0.prefix":             "",
						"destination.#":               "1",
						"destination.0.storage_class": s3.StorageClassStandard,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":                          "rule2",
						"priority":                    "2",
						"status":                      s3.ReplicationRuleStatusEnabled,
						"filter.#":                    "1",
						"filter.0.prefix":             "",
						"destination.#":               "1",
						"destination.0.storage_class": s3.StorageClassStandardIa,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":                          "rule3",
						"priority":                    "3",
						"status":                      s3.ReplicationRuleStatusDisabled,
						"filter.#":                    "1",
						"filter.0.prefix":             "",
						"destination.#":               "1",
						"destination.0.storage_class": s3.StorageClassOnezoneIa,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketReplicationConfiguration_multipleDestinationsNonEmptyFilter(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_replication_configuration.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroy, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationWithMultipleDestinationsNonEmptyFilter(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":                          "rule1",
						"priority":                    "1",
						"status":                      s3.ReplicationRuleStatusEnabled,
						"filter.#":                    "1",
						"filter.0.prefix":             "prefix1",
						"destination.#":               "1",
						"destination.0.storage_class": s3.StorageClassStandard,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":                          "rule2",
						"priority":                    "2",
						"status":                      s3.ReplicationRuleStatusEnabled,
						"filter.#":                    "1",
						"filter.0.tag.#":              "1",
						"filter.0.tag.0.key":          "Key2",
						"filter.0.tag.0.value":        "Value2",
						"destination.#":               "1",
						"destination.0.storage_class": s3.StorageClassStandardIa,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":                          "rule3",
						"priority":                    "3",
						"status":                      s3.ReplicationRuleStatusDisabled,
						"filter.#":                    "1",
						"filter.0.and.#":              "1",
						"filter.0.and.0.prefix":       "prefix3",
						"filter.0.and.0.tags.%":       "1",
						"filter.0.and.0.tags.Key3":    "Value3",
						"destination.#":               "1",
						"destination.0.storage_class": s3.StorageClassOnezoneIa,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketReplicationConfiguration_twoDestination(t *testing.T) {
	// This tests 2 destinations since GovCloud and possibly other non-standard partitions allow a max of 2
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_replication_configuration.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroy, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationWithMultipleDestinationsTwoDestination(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":                          "rule1",
						"priority":                    "1",
						"status":                      s3.ReplicationRuleStatusEnabled,
						"filter.#":                    "1",
						"filter.0.prefix":             "prefix1",
						"destination.#":               "1",
						"destination.0.storage_class": s3.StorageClassStandard,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":                          "rule2",
						"priority":                    "2",
						"status":                      s3.ReplicationRuleStatusEnabled,
						"filter.#":                    "1",
						"filter.0.prefix":             "prefix1",
						"destination.#":               "1",
						"destination.0.storage_class": s3.StorageClassStandardIa,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketReplicationConfiguration_configurationRuleDestinationAccessControlTranslation(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	callerIdentityDataSourceName := "data.aws_caller_identity.current"
	iamRoleResourceName := "aws_iam_role.test"
	dstBucketResourceName := "aws_s3_bucket.destination"
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_s3_bucket_replication_configuration.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroy, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationWithAccessControlTranslation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":            "foobar",
						"prefix":        "foo",
						"status":        s3.ReplicationRuleStatusEnabled,
						"destination.#": "1",
						"destination.0.access_control_translation.#":       "1",
						"destination.0.access_control_translation.0.owner": s3.OwnerOverrideDestination,
						"destination.0.storage_class":                      s3.StorageClassStandard,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.account", callerIdentityDataSourceName, "account_id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketReplicationConfigurationConfig_sseKMSEncryptedObjectsAndAccessControlTranslation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":            "foobar",
						"prefix":        "foo",
						"status":        s3.ReplicationRuleStatusEnabled,
						"destination.#": "1",
						"destination.0.access_control_translation.#":                     "1",
						"destination.0.access_control_translation.0.owner":               s3.OwnerOverrideDestination,
						"destination.0.encryption_configuration.#":                       "1",
						"source_selection_criteria.#":                                    "1",
						"source_selection_criteria.0.sse_kms_encrypted_objects.#":        "1",
						"source_selection_criteria.0.sse_kms_encrypted_objects.0.status": s3.SseKmsEncryptedObjectsStatusEnabled,
						"destination.0.storage_class":                                    s3.StorageClassStandard,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.account", callerIdentityDataSourceName, "account_id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, "arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.encryption_configuration.0.replica_kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/12480
func TestAccS3BucketReplicationConfiguration_configurationRuleDestinationAddAccessControlTranslation(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	callerIdentityDataSourceName := "data.aws_caller_identity.current"
	dstBucketResourceName := "aws_s3_bucket.destination"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_s3_bucket_replication_configuration.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroy, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationRulesDestination(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":                          "foobar",
						"prefix":                      "foo",
						"status":                      s3.ReplicationRuleStatusEnabled,
						"destination.#":               "1",
						"destination.0.storage_class": s3.StorageClassStandard,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.account", callerIdentityDataSourceName, "account_id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketReplicationConfigurationWithAccessControlTranslation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":            "foobar",
						"prefix":        "foo",
						"status":        s3.ReplicationRuleStatusEnabled,
						"destination.#": "1",
						"destination.0.access_control_translation.#":       "1",
						"destination.0.access_control_translation.0.owner": s3.OwnerOverrideDestination,
						"destination.0.storage_class":                      s3.StorageClassStandard,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.account", callerIdentityDataSourceName, "account_id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketReplicationConfiguration_replicationTimeControl(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dstBucketResourceName := "aws_s3_bucket.destination"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_s3_bucket_replication_configuration.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroy, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationRTC(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":                                 "foobar",
						"filter.#":                           "1",
						"filter.0.prefix":                    "foo",
						"status":                             s3.ReplicationRuleStatusEnabled,
						"delete_marker_replication.#":        "1",
						"delete_marker_replication.0.status": s3.DeleteMarkerReplicationStatusEnabled,
						"destination.#":                      "1",
						"destination.0.replication_time.#":   "1",
						"destination.0.replication_time.0.status":           s3.ReplicationTimeStatusEnabled,
						"destination.0.replication_time.0.time.#":           "1",
						"destination.0.replication_time.0.time.0.minutes":   "15",
						"destination.0.metrics.#":                           "1",
						"destination.0.metrics.0.status":                    s3.MetricsStatusEnabled,
						"destination.0.metrics.0.event_threshold.#":         "1",
						"destination.0.metrics.0.event_threshold.0.minutes": "15",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketReplicationConfiguration_replicaModifications(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dstBucketResourceName := "aws_s3_bucket.destination"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_s3_bucket_replication_configuration.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroy, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationReplicaMods(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":                                 "foobar",
						"filter.#":                           "1",
						"filter.0.prefix":                    "foo",
						"delete_marker_replication.#":        "1",
						"delete_marker_replication.0.status": s3.DeleteMarkerReplicationStatusEnabled,
						"source_selection_criteria.#":        "1",
						"source_selection_criteria.0.replica_modifications.#":        "1",
						"source_selection_criteria.0.replica_modifications.0.status": s3.ReplicaModificationsStatusEnabled,
						"status":        s3.ReplicationRuleStatusEnabled,
						"destination.#": "1",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccS3BucketReplicationConfiguration_withoutId ensures a configuration with a Computed
// rule.id does not result in a non-empty plan
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/23690
func TestAccS3BucketReplicationConfiguration_withoutId(t *testing.T) {
	resourceName := "aws_s3_bucket_replication_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dstBucketResourceName := "aws_s3_bucket.destination"
	iamRoleResourceName := "aws_iam_role.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroy, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfiguration_prefix_withoutIdConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "rule.0.id"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.prefix", "foo"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.status", s3.ReplicationRuleStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, "rule.0.destination.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "rule.0.destination.0.bucket", dstBucketResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// StorageClass issue: https://github.com/hashicorp/terraform/issues/10909
func TestAccS3BucketReplicationConfiguration_withoutStorageClass(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dstBucketResourceName := "aws_s3_bucket.destination"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_s3_bucket_replication_configuration.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroy, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationWithoutStorageClass(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":            "foobar",
						"prefix":        "foo",
						"status":        s3.ReplicationRuleStatusEnabled,
						"destination.#": "1",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketReplicationConfiguration_schemaV2(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dstBucketResourceName := "aws_s3_bucket.destination"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_s3_bucket_replication_configuration.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroy, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationWithV2ConfigurationNoTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":                                 "foobar",
						"filter.#":                           "1",
						"filter.0.prefix":                    "foo",
						"delete_marker_replication.#":        "1",
						"delete_marker_replication.0.status": s3.DeleteMarkerReplicationStatusEnabled,
						"status":                             s3.ReplicationRuleStatusEnabled,
						"destination.#":                      "1",
						"destination.0.storage_class":        s3.StorageClassStandard,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketReplicationConfiguration_schemaV2SameRegion(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rNameDestination := sdkacctest.RandomWithPrefix("tf-acc-test")
	dstBucketResourceName := "aws_s3_bucket.destination"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_s3_bucket_replication_configuration.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroy, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfiguration_schemaV2SameRegion(rName, rNameDestination),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":                                 "testid",
						"filter.#":                           "1",
						"filter.0.prefix":                    "testprefix",
						"delete_marker_replication.#":        "1",
						"delete_marker_replication.0.status": s3.DeleteMarkerReplicationStatusEnabled,
						"status":                             s3.ReplicationRuleStatusEnabled,
						"destination.#":                      "1",
						"destination.0.storage_class":        s3.StorageClassStandard,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/21895
func TestAccS3BucketReplicationConfiguration_schemaV2DestinationMetrics(t *testing.T) {
	resourceName := "aws_s3_bucket_replication_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroy, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfiguration_schemaV2DestinationMetrics_statusOnly(rName, s3.StorageClassStandard),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"destination.#":                             "1",
						"destination.0.metrics.#":                   "1",
						"destination.0.metrics.0.status":            s3.MetricsStatusEnabled,
						"destination.0.metrics.0.event_threshold.#": "0",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketReplicationConfiguration_existingObjectReplication(t *testing.T) {
	t.Skipf("skipping test: AWS Technical Support request required to allow ExistingObjectReplication")

	resourceName := "aws_s3_bucket_replication_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameDestination := sdkacctest.RandomWithPrefix("tf-acc-test")
	dstBucketResourceName := "aws_s3_bucket.destination"
	iamRoleResourceName := "aws_iam_role.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroy, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfiguration_existingObjectReplication(rName, rNameDestination),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":                                   "testid",
						"filter.#":                             "1",
						"filter.0.prefix":                      "testprefix",
						"delete_marker_replication.#":          "1",
						"delete_marker_replication.0.status":   s3.DeleteMarkerReplicationStatusEnabled,
						"existing_object_replication.#":        "1",
						"existing_object_replication.0.status": s3.ExistingObjectReplicationStatusEnabled,
						"status":                               s3.ReplicationRuleStatusEnabled,
						"destination.#":                        "1",
						"destination.0.storage_class":          s3.StorageClassStandard,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/23487
func TestAccS3BucketReplicationConfiguration_filter_emptyConfigurationBlock(t *testing.T) {
	resourceName := "aws_s3_bucket_replication_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dstBucketResourceName := "aws_s3_bucket.destination"
	iamRoleResourceName := "aws_iam_role.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroy, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfiguration_filter_emptyConfigurationBlock(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":                                 "foobar",
						"delete_marker_replication.#":        "1",
						"delete_marker_replication.0.status": s3.DeleteMarkerReplicationStatusDisabled,
						"filter.#":                           "1",
						"status":                             s3.ReplicationRuleStatusEnabled,
						"destination.#":                      "1",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/23487
func TestAccS3BucketReplicationConfiguration_filter_emptyPrefix(t *testing.T) {
	resourceName := "aws_s3_bucket_replication_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dstBucketResourceName := "aws_s3_bucket.destination"
	iamRoleResourceName := "aws_iam_role.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroy, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfiguration_filter_emptyPrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":                                 "foobar",
						"delete_marker_replication.#":        "1",
						"delete_marker_replication.0.status": s3.DeleteMarkerReplicationStatusDisabled,
						"filter.#":                           "1",
						"filter.0.prefix":                    "",
						"status":                             s3.ReplicationRuleStatusEnabled,
						"destination.#":                      "1",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// The "rule" parameter as a TypeList will have a nil value
				// if prefix was specified as an empty string, which was used as workaround
				// when the parameter was a TypeSet
				ImportStateVerifyIgnore: []string{"rule.0.filter.0.prefix"},
			},
		},
	})
}

func TestAccS3BucketReplicationConfiguration_filter_tagFilter(t *testing.T) {
	resourceName := "aws_s3_bucket_replication_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dstBucketResourceName := "aws_s3_bucket.destination"
	iamRoleResourceName := "aws_iam_role.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroy, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfiguration_filter_tag(rName, "testkey", "testvalue"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":                                 "foobar",
						"delete_marker_replication.#":        "1",
						"delete_marker_replication.0.status": s3.DeleteMarkerReplicationStatusDisabled,
						"filter.#":                           "1",
						"filter.0.tag.#":                     "1",
						"filter.0.tag.0.key":                 "testkey",
						"filter.0.tag.0.value":               "testvalue",
						"status":                             s3.ReplicationRuleStatusEnabled,
						"destination.#":                      "1",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketReplicationConfiguration_filter_andOperator(t *testing.T) {
	resourceName := "aws_s3_bucket_replication_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dstBucketResourceName := "aws_s3_bucket.destination"
	iamRoleResourceName := "aws_iam_role.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroy, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfiguration_filter_andOperator_prefixAndTags(rName, "testkey1", "testvalue1", "testkey2", "testvalue2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":                                 "foobar",
						"delete_marker_replication.#":        "1",
						"delete_marker_replication.0.status": s3.DeleteMarkerReplicationStatusDisabled,
						"filter.#":                           "1",
						"filter.0.and.#":                     "1",
						"filter.0.and.0.prefix":              "foo",
						"filter.0.and.0.tags.%":              "2",
						"filter.0.and.0.tags.testkey1":       "testvalue1",
						"filter.0.and.0.tags.testkey2":       "testvalue2",
						"status":                             s3.ReplicationRuleStatusEnabled,
						"destination.#":                      "1",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketReplicationConfiguration_filter_andOperator_tags(rName, "testkey1", "testvalue1", "testkey2", "testvalue2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":                                 "foobar",
						"delete_marker_replication.#":        "1",
						"delete_marker_replication.0.status": s3.DeleteMarkerReplicationStatusDisabled,
						"filter.#":                           "1",
						"filter.0.and.#":                     "1",
						"filter.0.and.0.tags.%":              "2",
						"filter.0.and.0.tags.testkey1":       "testvalue1",
						"filter.0.and.0.tags.testkey2":       "testvalue2",
						"status":                             s3.ReplicationRuleStatusEnabled,
						"destination.#":                      "1",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccS3BucketReplicationConfiguration_filter_withoutId ensures a configuration with a Computed
// rule.id does not result in a non-empty plan.
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/23690
func TestAccS3BucketReplicationConfiguration_filter_withoutId(t *testing.T) {
	resourceName := "aws_s3_bucket_replication_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dstBucketResourceName := "aws_s3_bucket.destination"
	iamRoleResourceName := "aws_iam_role.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroy, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfiguration_filter_withoutIdConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "rule.0.id"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.status", s3.ReplicationRuleStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, "rule.0.delete_marker_replication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.destination.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "rule.0.destination.0.bucket", dstBucketResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/21961
func TestAccS3BucketReplicationConfiguration_withoutPrefix(t *testing.T) {
	resourceName := "aws_s3_bucket_replication_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroy, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationWithoutPrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketReplicationConfiguration_migrate_noChange(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_replication_configuration.test"
	bucketResourceName := "aws_s3_bucket.source"
	region := acctest.Region()

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroy, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_withReplicationV2_PrefixAndTags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExistsWithProvider(bucketResourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(bucketResourceName, "replication_configuration.0.rules.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(bucketResourceName, "replication_configuration.0.rules.*", map[string]string{
						"filter.#":        "1",
						"filter.0.prefix": "foo",
						"filter.0.tags.%": "2",
					}),
				),
			},
			{
				Config: testAccBucketReplicationConfiguration_Migrate_NoChangeConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.0.and.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.0.and.0.prefix", "foo"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.0.and.0.tags.%", "2"),
				),
			},
		},
	})
}

func TestAccS3BucketReplicationConfiguration_migrate_withChange(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_replication_configuration.test"
	bucketResourceName := "aws_s3_bucket.source"
	region := acctest.Region()

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroy, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_withReplicationV2_PrefixAndTags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExistsWithProvider(bucketResourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(bucketResourceName, "replication_configuration.0.rules.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(bucketResourceName, "replication_configuration.0.rules.*", map[string]string{
						"filter.#":        "1",
						"filter.0.prefix": "foo",
						"filter.0.tags.%": "2",
					}),
				),
			},
			{
				Config: testAccBucketReplicationConfiguration_Migrate_WithChangeConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.0.prefix", "bar"),
				),
			},
		},
	})
}

func testAccCheckBucketReplicationConfigurationDestroy(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*conns.AWSClient).S3Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket_replication_configuration" {
			continue
		}
		input := &s3.GetBucketReplicationInput{Bucket: aws.String(rs.Primary.ID)}

		output, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
			return conn.GetBucketReplication(input)
		}, s3.ErrCodeNoSuchBucket)

		if tfawserr.ErrCodeEquals(err, tfs3.ErrCodeReplicationConfigurationNotFound, s3.ErrCodeNoSuchBucket) {
			continue
		}

		if err != nil {
			return err
		}

		if replication, ok := output.(*s3.GetBucketReplicationOutput); ok && replication != nil && replication.ReplicationConfiguration != nil {
			return fmt.Errorf("S3 Replication Configuration for bucket (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckBucketReplicationConfigurationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		output, err := conn.GetBucketReplication(&s3.GetBucketReplicationInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if output == nil || output.ReplicationConfiguration == nil {
			return fmt.Errorf("S3 Bucket Replication Configuration for bucket (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccBucketReplicationConfigurationBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "s3.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_s3_bucket" "destination" {
  provider = "awsalternate"
  bucket   = "%[1]s-destination"
}

resource "aws_s3_bucket_versioning" "destination" {
  bucket = aws_s3_bucket.destination.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-source"
}

resource "aws_s3_bucket_versioning" "source" {
  bucket = aws_s3_bucket.source.id
  versioning_configuration {
    status = "Enabled"
  }
}
`, rName)
}

func testAccBucketReplicationConfigurationBasic(rName, storageClass string) string {
	return testAccBucketReplicationConfigurationBase(rName) + fmt.Sprintf(`
resource "aws_s3_bucket_replication_configuration" "test" {
  depends_on = [
    aws_s3_bucket_versioning.source,
    aws_s3_bucket_versioning.destination
  ]

  bucket = aws_s3_bucket.source.id
  role   = aws_iam_role.test.arn

  rule {
    id     = "foobar"
    prefix = "foo"
    status = "Enabled"

    destination {
      bucket        = aws_s3_bucket.destination.arn
      storage_class = %[1]q
    }
  }
}`, storageClass)
}

func testAccBucketReplicationConfiguration_prefix_withoutIdConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccBucketReplicationConfigurationBase(rName), `
resource "aws_s3_bucket_replication_configuration" "test" {
  depends_on = [
    aws_s3_bucket_versioning.source,
    aws_s3_bucket_versioning.destination
  ]

  bucket = aws_s3_bucket.source.id
  role   = aws_iam_role.test.arn

  rule {
    prefix = "foo"
    status = "Enabled"

    destination {
      bucket = aws_s3_bucket.destination.arn
    }
  }
}`)
}

func testAccBucketReplicationConfiguration_filter_withoutIdConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccBucketReplicationConfigurationBase(rName), `
resource "aws_s3_bucket_replication_configuration" "test" {
  depends_on = [
    aws_s3_bucket_versioning.source,
    aws_s3_bucket_versioning.destination
  ]

  bucket = aws_s3_bucket.source.id
  role   = aws_iam_role.test.arn

  rule {
    filter {}

    status = "Enabled"

    delete_marker_replication {
      status = "Disabled"
    }

    destination {
      bucket = aws_s3_bucket.destination.arn
    }
  }
}`)
}

func testAccBucketReplicationConfigurationRTC(rName string) string {
	return acctest.ConfigCompose(
		testAccBucketReplicationConfigurationBase(rName),
		`
resource "aws_s3_bucket_replication_configuration" "test" {
  depends_on = [
    aws_s3_bucket_versioning.source,
    aws_s3_bucket_versioning.destination
  ]

  bucket = aws_s3_bucket.source.id
  role   = aws_iam_role.test.arn

  rule {
    id = "foobar"
    filter {
      prefix = "foo"
    }
    status = "Enabled"
    delete_marker_replication {
      status = "Enabled"
    }
    destination {
      bucket = aws_s3_bucket.destination.arn
      replication_time {
        status = "Enabled"
        time {
          minutes = 15
        }
      }
      metrics {
        status = "Enabled"
        event_threshold {
          minutes = 15
        }
      }
    }
  }
}`)
}

func testAccBucketReplicationConfigurationReplicaMods(rName string) string {
	return testAccBucketReplicationConfigurationBase(rName) + `
resource "aws_s3_bucket_replication_configuration" "test" {
  depends_on = [
    aws_s3_bucket_versioning.source,
    aws_s3_bucket_versioning.destination
  ]

  bucket = aws_s3_bucket.source.id
  role   = aws_iam_role.test.arn

  rule {
    id = "foobar"
    filter {
      prefix = "foo"
    }
    source_selection_criteria {
      replica_modifications {
        status = "Enabled"
      }
    }
    delete_marker_replication {
      status = "Enabled"
    }

    status = "Enabled"
    destination {
      bucket = aws_s3_bucket.destination.arn
    }
  }
}`
}

func testAccBucketReplicationConfigurationWithMultipleDestinationsEmptyFilter(rName string) string {
	return acctest.ConfigCompose(
		testAccBucketReplicationConfigurationBase(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "destination2" {
  provider = "awsalternate"
  bucket   = "%[1]s-destination2"
}

resource "aws_s3_bucket_versioning" "destination2" {
  bucket = aws_s3_bucket.destination2.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket" "destination3" {
  provider = "awsalternate"
  bucket   = "%[1]s-destination3"
}

resource "aws_s3_bucket_versioning" "destination3" {
  bucket = aws_s3_bucket.destination3.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_replication_configuration" "test" {
  # Must have bucket versioning enabled first
  depends_on = [
    aws_s3_bucket_versioning.source,
    aws_s3_bucket_versioning.destination,
    aws_s3_bucket_versioning.destination2,
    aws_s3_bucket_versioning.destination3
  ]

  bucket = aws_s3_bucket.source.id
  role   = aws_iam_role.test.arn

  rule {
    id       = "rule1"
    priority = 1
    status   = "Enabled"

    filter {}

    delete_marker_replication {
      status = "Enabled"
    }

    destination {
      bucket        = aws_s3_bucket.destination.arn
      storage_class = "STANDARD"
    }
  }

  rule {
    id       = "rule2"
    priority = 2
    status   = "Enabled"

    filter {}

    delete_marker_replication {
      status = "Enabled"
    }

    destination {
      bucket        = aws_s3_bucket.destination2.arn
      storage_class = "STANDARD_IA"
    }
  }

  rule {
    id       = "rule3"
    priority = 3
    status   = "Disabled"

    filter {}

    delete_marker_replication {
      status = "Enabled"
    }

    destination {
      bucket        = aws_s3_bucket.destination3.arn
      storage_class = "ONEZONE_IA"
    }
  }
}`, rName))
}

func testAccBucketReplicationConfigurationWithMultipleDestinationsNonEmptyFilter(rName string) string {
	return acctest.ConfigCompose(
		testAccBucketReplicationConfigurationBase(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "destination2" {
  provider = "awsalternate"
  bucket   = "%[1]s-destination2"
}

resource "aws_s3_bucket_versioning" "destination2" {
  bucket = aws_s3_bucket.destination2.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket" "destination3" {
  provider = "awsalternate"
  bucket   = "%[1]s-destination3"
}

resource "aws_s3_bucket_versioning" "destination3" {
  bucket = aws_s3_bucket.destination3.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_replication_configuration" "test" {
  depends_on = [
    aws_s3_bucket_versioning.source,
    aws_s3_bucket_versioning.destination,
    aws_s3_bucket_versioning.destination2,
    aws_s3_bucket_versioning.destination3
  ]

  bucket = aws_s3_bucket.source.id
  role   = aws_iam_role.test.arn

  rule {
    id       = "rule1"
    priority = 1
    status   = "Enabled"

    filter {
      prefix = "prefix1"
    }

    delete_marker_replication {
      status = "Enabled"
    }

    destination {
      bucket        = aws_s3_bucket.destination.arn
      storage_class = "STANDARD"
    }
  }

  rule {
    id       = "rule2"
    priority = 2
    status   = "Enabled"

    filter {
      tag {
        key   = "Key2"
        value = "Value2"
      }
    }

    delete_marker_replication {
      status = "Disabled"
    }

    destination {
      bucket        = aws_s3_bucket.destination2.arn
      storage_class = "STANDARD_IA"
    }
  }

  rule {
    id       = "rule3"
    priority = 3
    status   = "Disabled"

    filter {
      and {
        prefix = "prefix3"
        tags = {
          Key3 = "Value3"
        }
      }
    }

    delete_marker_replication {
      status = "Disabled"
    }

    destination {
      bucket        = aws_s3_bucket.destination3.arn
      storage_class = "ONEZONE_IA"
    }
  }
}`, rName))
}

func testAccBucketReplicationConfigurationWithMultipleDestinationsTwoDestination(rName string) string {
	return acctest.ConfigCompose(
		testAccBucketReplicationConfigurationBase(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "destination2" {
  provider = "awsalternate"
  bucket   = "%[1]s-destination2"
}

resource "aws_s3_bucket_versioning" "destination2" {
  bucket = aws_s3_bucket.destination2.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_replication_configuration" "test" {
  depends_on = [
    aws_s3_bucket_versioning.source,
    aws_s3_bucket_versioning.destination,
    aws_s3_bucket_versioning.destination2
  ]

  bucket = aws_s3_bucket.source.id
  role   = aws_iam_role.test.arn

  rule {
    id       = "rule1"
    priority = 1
    status   = "Enabled"

    filter {
      prefix = "prefix1"
    }

    delete_marker_replication {
      status = "Enabled"
    }

    destination {
      bucket        = aws_s3_bucket.destination.arn
      storage_class = "STANDARD"
    }
  }

  rule {
    id       = "rule2"
    priority = 2
    status   = "Enabled"

    filter {
      prefix = "prefix1"
    }

    delete_marker_replication {
      status = "Enabled"
    }

    destination {
      bucket        = aws_s3_bucket.destination2.arn
      storage_class = "STANDARD_IA"
    }
  }
}`, rName))
}

func testAccBucketReplicationConfigurationConfig_sseKMSEncryptedObjects(rName string) string {
	return testAccBucketReplicationConfigurationBase(rName) + `
resource "aws_kms_key" "test" {
  provider                = "awsalternate"
  description             = "TF Acceptance Test S3 repl KMS key"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket_replication_configuration" "test" {
  depends_on = [
    aws_s3_bucket_versioning.source,
    aws_s3_bucket_versioning.destination
  ]

  bucket = aws_s3_bucket.source.id
  role   = aws_iam_role.test.arn

  rule {
    id     = "foobar"
    prefix = "foo"
    status = "Enabled"

    destination {
      bucket = aws_s3_bucket.destination.arn

      encryption_configuration {
        replica_kms_key_id = aws_kms_key.test.arn
      }

      storage_class = "STANDARD"
    }

    source_selection_criteria {
      sse_kms_encrypted_objects {
        status = "Enabled"
      }
    }
  }
}`
}

func testAccBucketReplicationConfigurationWithAccessControlTranslation(rName string) string {
	return testAccBucketReplicationConfigurationBase(rName) + `
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket_replication_configuration" "test" {
  depends_on = [
    aws_s3_bucket_versioning.source,
    aws_s3_bucket_versioning.destination
  ]

  bucket = aws_s3_bucket.source.id
  role   = aws_iam_role.test.arn

  rule {
    id     = "foobar"
    prefix = "foo"
    status = "Enabled"

    destination {
      account       = data.aws_caller_identity.current.account_id
      bucket        = aws_s3_bucket.destination.arn
      storage_class = "STANDARD"

      access_control_translation {
        owner = "Destination"
      }
    }
  }
}`
}

func testAccBucketReplicationConfigurationRulesDestination(rName string) string {
	return testAccBucketReplicationConfigurationBase(rName) + `
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket_replication_configuration" "test" {
  depends_on = [
    aws_s3_bucket_versioning.source,
    aws_s3_bucket_versioning.destination
  ]

  bucket = aws_s3_bucket.source.id
  role   = aws_iam_role.test.arn

  rule {
    id     = "foobar"
    prefix = "foo"
    status = "Enabled"

    destination {
      account       = data.aws_caller_identity.current.account_id
      bucket        = aws_s3_bucket.destination.arn
      storage_class = "STANDARD"
    }
  }
}`
}

func testAccBucketReplicationConfigurationConfig_sseKMSEncryptedObjectsAndAccessControlTranslation(rName string) string {
	return testAccBucketReplicationConfigurationBase(rName) + `
data "aws_caller_identity" "current" {}

resource "aws_kms_key" "test" {
  provider                = "awsalternate"
  description             = "TF Acceptance Test S3 repl KMS key"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket_replication_configuration" "test" {
  depends_on = [
    aws_s3_bucket_versioning.source,
    aws_s3_bucket_versioning.destination
  ]

  bucket = aws_s3_bucket.source.id
  role   = aws_iam_role.test.arn

  rule {
    id     = "foobar"
    prefix = "foo"
    status = "Enabled"

    destination {
      account       = data.aws_caller_identity.current.account_id
      bucket        = aws_s3_bucket.destination.arn
      storage_class = "STANDARD"
      encryption_configuration {
        replica_kms_key_id = aws_kms_key.test.arn
      }

      access_control_translation {
        owner = "Destination"
      }
    }

    source_selection_criteria {
      sse_kms_encrypted_objects {
        status = "Enabled"
      }
    }
  }
}`
}

func testAccBucketReplicationConfigurationWithoutStorageClass(rName string) string {
	return testAccBucketReplicationConfigurationBase(rName) + `
resource "aws_s3_bucket_replication_configuration" "test" {
  depends_on = [
    aws_s3_bucket_versioning.source,
    aws_s3_bucket_versioning.destination
  ]

  bucket = aws_s3_bucket.source.id
  role   = aws_iam_role.test.arn

  rule {
    id     = "foobar"
    prefix = "foo"
    status = "Enabled"

    destination {
      bucket = aws_s3_bucket.destination.arn
    }
  }
}`
}

func testAccBucketReplicationConfigurationWithV2ConfigurationNoTags(rName string) string {
	return testAccBucketReplicationConfigurationBase(rName) + `
resource "aws_s3_bucket_replication_configuration" "test" {
  depends_on = [
    aws_s3_bucket_versioning.source,
    aws_s3_bucket_versioning.destination
  ]

  bucket = aws_s3_bucket.source.id
  role   = aws_iam_role.test.arn

  rule {
    id     = "foobar"
    status = "Enabled"

    filter {
      prefix = "foo"
    }

    delete_marker_replication {
      status = "Enabled"
    }

    destination {
      bucket        = aws_s3_bucket.destination.arn
      storage_class = "STANDARD"
    }
  }
}`
}

func testAccBucketReplicationConfiguration_schemaV2SameRegion(rName, rNameDestination string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "s3.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_s3_bucket" "destination" {
  bucket = %[2]q
}

resource "aws_s3_bucket_versioning" "destination" {
  bucket = aws_s3_bucket.destination.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket" "source" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "source_acl" {
  bucket = aws_s3_bucket.source.id
  acl    = "private"
}

resource "aws_s3_bucket_versioning" "source" {
  bucket = aws_s3_bucket.source.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_replication_configuration" "test" {
  depends_on = [
    aws_s3_bucket_versioning.source,
    aws_s3_bucket_versioning.destination
  ]

  bucket = aws_s3_bucket.source.id
  role   = aws_iam_role.test.arn

  rule {
    id     = "testid"
    status = "Enabled"

    filter {
      prefix = "testprefix"
    }

    delete_marker_replication {
      status = "Enabled"
    }

    destination {
      bucket        = aws_s3_bucket.destination.arn
      storage_class = "STANDARD"
    }
  }
}`, rName, rNameDestination)
}

func testAccBucketReplicationConfiguration_existingObjectReplication(rName, rNameDestination string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "s3.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_s3_bucket" "destination" {
  bucket = %[2]q
}

resource "aws_s3_bucket_versioning" "destination" {
  bucket = aws_s3_bucket.destination.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket" "source" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "source_acl" {
  bucket = aws_s3_bucket.source.id
  acl    = "private"
}

resource "aws_s3_bucket_versioning" "source" {
  bucket = aws_s3_bucket.source.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_replication_configuration" "test" {
  depends_on = [
    aws_s3_bucket_versioning.source,
    aws_s3_bucket_versioning.destination
  ]

  bucket = aws_s3_bucket.source.id
  role   = aws_iam_role.test.arn

  rule {
    id     = "testid"
    status = "Enabled"

    filter {
      prefix = "testprefix"
    }

    existing_object_replication {
      status = "Enabled"
    }

    delete_marker_replication {
      status = "Enabled"
    }

    destination {
      bucket        = aws_s3_bucket.destination.arn
      storage_class = "STANDARD"
    }
  }
}
`, rName, rNameDestination)
}

func testAccBucketReplicationConfiguration_filter_emptyConfigurationBlock(rName string) string {
	return acctest.ConfigCompose(
		testAccBucketReplicationConfigurationBase(rName),
		`
resource "aws_s3_bucket_replication_configuration" "test" {
  depends_on = [aws_s3_bucket_versioning.source]

  bucket = aws_s3_bucket.source.id
  role   = aws_iam_role.test.arn

  rule {
    id = "foobar"

    delete_marker_replication {
      status = "Disabled"
    }

    filter {}

    status = "Enabled"

    destination {
      bucket = aws_s3_bucket.destination.arn
    }
  }
}`)
}

func testAccBucketReplicationConfiguration_filter_emptyPrefix(rName string) string {
	return acctest.ConfigCompose(
		testAccBucketReplicationConfigurationBase(rName), `
resource "aws_s3_bucket_replication_configuration" "test" {
  depends_on = [aws_s3_bucket_versioning.source]

  bucket = aws_s3_bucket.source.id
  role   = aws_iam_role.test.arn

  rule {
    id = "foobar"

    delete_marker_replication {
      status = "Disabled"
    }

    filter {
      prefix = ""
    }

    status = "Enabled"

    destination {
      bucket = aws_s3_bucket.destination.arn
    }
  }
}`,
	)
}

func testAccBucketReplicationConfiguration_filter_tag(rName, key, value string) string {
	return acctest.ConfigCompose(
		testAccBucketReplicationConfigurationBase(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket_replication_configuration" "test" {
  depends_on = [
    aws_s3_bucket_versioning.source,
    aws_s3_bucket_versioning.destination
  ]

  bucket = aws_s3_bucket.source.id
  role   = aws_iam_role.test.arn

  rule {
    id = "foobar"

    delete_marker_replication {
      status = "Disabled"
    }

    filter {
      tag {
        key   = %[1]q
        value = %[2]q
      }
    }

    status = "Enabled"

    destination {
      bucket = aws_s3_bucket.destination.arn
    }
  }
}`, key, value))
}

func testAccBucketReplicationConfiguration_filter_andOperator_tags(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(
		testAccBucketReplicationConfigurationBase(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket_replication_configuration" "test" {
  depends_on = [
    aws_s3_bucket_versioning.source,
    aws_s3_bucket_versioning.destination
  ]

  bucket = aws_s3_bucket.source.id
  role   = aws_iam_role.test.arn

  rule {
    id = "foobar"

    delete_marker_replication {
      status = "Disabled"
    }

    filter {
      and {
        tags = {
          %[1]q = %[2]q
          %[3]q = %[4]q
        }
      }
    }

    status = "Enabled"

    destination {
      bucket = aws_s3_bucket.destination.arn
    }
  }
}`, key1, value1, key2, value2))
}

func testAccBucketReplicationConfiguration_filter_andOperator_prefixAndTags(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(
		testAccBucketReplicationConfigurationBase(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket_replication_configuration" "test" {
  depends_on = [
    aws_s3_bucket_versioning.source,
    aws_s3_bucket_versioning.destination
  ]

  bucket = aws_s3_bucket.source.id
  role   = aws_iam_role.test.arn

  rule {
    id = "foobar"

    delete_marker_replication {
      status = "Disabled"
    }

    filter {
      and {
        prefix = "foo"
        tags = {
          %[1]q = %[2]q
          %[3]q = %[4]q
        }
      }
    }

    status = "Enabled"

    destination {
      bucket = aws_s3_bucket.destination.arn
    }
  }
}`, key1, value1, key2, value2))
}

func testAccBucketReplicationConfiguration_schemaV2DestinationMetrics_statusOnly(rName, storageClass string) string {
	return testAccBucketReplicationConfigurationBase(rName) + fmt.Sprintf(`
resource "aws_s3_bucket_replication_configuration" "test" {
  depends_on = [
    aws_s3_bucket_versioning.source,
    aws_s3_bucket_versioning.destination
  ]

  bucket = aws_s3_bucket.source.id
  role   = aws_iam_role.test.arn

  rule {
    id = "foobar"
    filter {
      prefix = "foo"
    }
    status = "Enabled"

    delete_marker_replication {
      status = "Enabled"
    }

    destination {
      bucket        = aws_s3_bucket.destination.arn
      storage_class = %[1]q

      metrics {
        status = "Enabled"
      }
    }
  }
}`, storageClass)
}

func testAccBucketReplicationConfigurationWithoutPrefix(rName string) string {
	return acctest.ConfigCompose(
		testAccBucketReplicationConfigurationBase(rName),
		`
resource "aws_s3_bucket_replication_configuration" "test" {
  depends_on = [
    aws_s3_bucket_versioning.source,
    aws_s3_bucket_versioning.destination
  ]

  bucket = aws_s3_bucket.source.id
  role   = aws_iam_role.test.arn

  rule {
    id     = "foobar"
    status = "Enabled"

    destination {
      bucket        = aws_s3_bucket.destination.arn
      storage_class = "STANDARD"
    }
  }
}`)
}

func testAccBucketReplicationConfigurationMigrationBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "role" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "s3.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-source"
}

resource "aws_s3_bucket_acl" "source" {
  bucket = aws_s3_bucket.source.id
  acl    = "private"
}

resource "aws_s3_bucket_versioning" "source" {
  bucket = aws_s3_bucket.source.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket" "destination" {
  provider = "awsalternate"
  bucket   = "%[1]s-destination"
}

resource "aws_s3_bucket_versioning" "destination" {
  provider = "awsalternate"
  bucket   = aws_s3_bucket.destination.id
  versioning_configuration {
    status = "Enabled"
  }
}
`, rName)
}

func testAccBucketReplicationConfiguration_Migrate_NoChangeConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		testAccBucketReplicationConfigurationMigrationBase(rName), `
resource "aws_s3_bucket_replication_configuration" "test" {
  depends_on = [
    aws_s3_bucket_versioning.source,
    aws_s3_bucket_versioning.destination
  ]

  bucket = aws_s3_bucket.source.id
  role   = aws_iam_role.role.arn

  rule {
    id     = "foobar"
    status = "Enabled"

    priority = 41

    delete_marker_replication {
      status = "Disabled"
    }

    destination {
      bucket        = aws_s3_bucket.destination.arn
      storage_class = "STANDARD"
    }

    filter {
      and {
        prefix = "foo"

        tags = {
          AnotherTag  = "OK"
          ReplicateMe = "Yes"
        }
      }
    }
  }
}`,
	)
}

func testAccBucketReplicationConfiguration_Migrate_WithChangeConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		testAccBucketReplicationConfigurationMigrationBase(rName), `
resource "aws_s3_bucket_replication_configuration" "test" {
  depends_on = [
    aws_s3_bucket_versioning.source,
    aws_s3_bucket_versioning.destination
  ]

  bucket = aws_s3_bucket.source.id
  role   = aws_iam_role.role.arn

  rule {
    id     = "foobar"
    status = "Enabled"

    priority = 41

    delete_marker_replication {
      status = "Disabled"
    }

    destination {
      bucket        = aws_s3_bucket.destination.arn
      storage_class = "STANDARD"
    }

    filter {
      prefix = "bar"
    }
  }
}`,
	)
}
