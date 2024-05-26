// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3BucketReplicationConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	iamRoleResourceName := "aws_iam_role.test"
	dstBucketResourceName := "aws_s3_bucket.destination"
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_s3_bucket_replication_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationConfig_basic(rName, string(types.StorageClassStandard)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRole, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:                  "foobar",
						names.AttrPrefix:              "foo",
						names.AttrStatus:              string(types.ReplicationRuleStatusEnabled),
						"destination.#":               acctest.Ct1,
						"destination.0.storage_class": string(types.StorageClassStandard),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, names.AttrARN),
				),
			},
			{
				Config: testAccBucketReplicationConfigurationConfig_basic(rName, string(types.StorageClassGlacier)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRole, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:                  "foobar",
						names.AttrPrefix:              "foo",
						names.AttrStatus:              string(types.ReplicationRuleStatusEnabled),
						"destination.#":               acctest.Ct1,
						"destination.0.storage_class": string(types.StorageClassGlacier),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, names.AttrARN),
				),
			},
			{
				Config: testAccBucketReplicationConfigurationConfig_sseKMSEncryptedObjects(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRole, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:     "foobar",
						names.AttrPrefix: "foo",
						names.AttrStatus: string(types.ReplicationRuleStatusEnabled),
						"destination.#":  acctest.Ct1,
						"destination.0.encryption_configuration.#":                       acctest.Ct1,
						"destination.0.storage_class":                                    string(types.StorageClassStandard),
						"source_selection_criteria.#":                                    acctest.Ct1,
						"source_selection_criteria.0.sse_kms_encrypted_objects.#":        acctest.Ct1,
						"source_selection_criteria.0.sse_kms_encrypted_objects.0.status": string(types.SseKmsEncryptedObjectsStatusEnabled),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.encryption_configuration.0.replica_kms_key_id", kmsKeyResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccS3BucketReplicationConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3_bucket_replication_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationConfig_basic(rName, string(types.StorageClassStandard)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3.ResourceBucketReplicationConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketReplicationConfiguration_multipleDestinationsEmptyFilter(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_replication_configuration.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationConfig_multipleDestinationsEmptyFilter(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:                  "rule1",
						names.AttrPriority:            acctest.Ct1,
						names.AttrStatus:              string(types.ReplicationRuleStatusEnabled),
						"filter.#":                    acctest.Ct1,
						"filter.0.prefix":             "",
						"destination.#":               acctest.Ct1,
						"destination.0.storage_class": string(types.StorageClassStandard),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:                  "rule2",
						names.AttrPriority:            acctest.Ct2,
						names.AttrStatus:              string(types.ReplicationRuleStatusEnabled),
						"filter.#":                    acctest.Ct1,
						"filter.0.prefix":             "",
						"destination.#":               acctest.Ct1,
						"destination.0.storage_class": string(types.StorageClassStandardIa),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:                  "rule3",
						names.AttrPriority:            acctest.Ct3,
						names.AttrStatus:              string(types.ReplicationRuleStatusDisabled),
						"filter.#":                    acctest.Ct1,
						"filter.0.prefix":             "",
						"destination.#":               acctest.Ct1,
						"destination.0.storage_class": string(types.StorageClassOnezoneIa),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_replication_configuration.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationConfig_multipleDestinationsNonEmptyFilter(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:                  "rule1",
						names.AttrPriority:            acctest.Ct1,
						names.AttrStatus:              string(types.ReplicationRuleStatusEnabled),
						"filter.#":                    acctest.Ct1,
						"filter.0.prefix":             "prefix1",
						"destination.#":               acctest.Ct1,
						"destination.0.storage_class": string(types.StorageClassStandard),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:                  "rule2",
						names.AttrPriority:            acctest.Ct2,
						names.AttrStatus:              string(types.ReplicationRuleStatusEnabled),
						"filter.#":                    acctest.Ct1,
						"filter.0.tag.#":              acctest.Ct1,
						"filter.0.tag.0.key":          "Key2",
						"filter.0.tag.0.value":        "Value2",
						"destination.#":               acctest.Ct1,
						"destination.0.storage_class": string(types.StorageClassStandardIa),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:                  "rule3",
						names.AttrPriority:            acctest.Ct3,
						names.AttrStatus:              string(types.ReplicationRuleStatusDisabled),
						"filter.#":                    acctest.Ct1,
						"filter.0.and.#":              acctest.Ct1,
						"filter.0.and.0.prefix":       "prefix3",
						"filter.0.and.0.tags.%":       acctest.Ct1,
						"filter.0.and.0.tags.Key3":    "Value3",
						"destination.#":               acctest.Ct1,
						"destination.0.storage_class": string(types.StorageClassOnezoneIa),
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
	ctx := acctest.Context(t)

	// This tests 2 destinations since GovCloud and possibly other non-standard partitions allow a max of 2
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_replication_configuration.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationConfig_multipleDestinationsTwoDestination(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:                  "rule1",
						names.AttrPriority:            acctest.Ct1,
						names.AttrStatus:              string(types.ReplicationRuleStatusEnabled),
						"filter.#":                    acctest.Ct1,
						"filter.0.prefix":             "prefix1",
						"destination.#":               acctest.Ct1,
						"destination.0.storage_class": string(types.StorageClassStandard),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:                  "rule2",
						names.AttrPriority:            acctest.Ct2,
						names.AttrStatus:              string(types.ReplicationRuleStatusEnabled),
						"filter.#":                    acctest.Ct1,
						"filter.0.prefix":             "prefix1",
						"destination.#":               acctest.Ct1,
						"destination.0.storage_class": string(types.StorageClassStandardIa),
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
	ctx := acctest.Context(t)
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
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationConfig_accessControlTranslation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRole, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:     "foobar",
						names.AttrPrefix: "foo",
						names.AttrStatus: string(types.ReplicationRuleStatusEnabled),
						"destination.#":  acctest.Ct1,
						"destination.0.access_control_translation.#":       acctest.Ct1,
						"destination.0.access_control_translation.0.owner": string(types.OwnerOverrideDestination),
						"destination.0.storage_class":                      string(types.StorageClassStandard),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.account", callerIdentityDataSourceName, names.AttrAccountID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, names.AttrARN),
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
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRole, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:     "foobar",
						names.AttrPrefix: "foo",
						names.AttrStatus: string(types.ReplicationRuleStatusEnabled),
						"destination.#":  acctest.Ct1,
						"destination.0.access_control_translation.#":                     acctest.Ct1,
						"destination.0.access_control_translation.0.owner":               string(types.OwnerOverrideDestination),
						"destination.0.encryption_configuration.#":                       acctest.Ct1,
						"source_selection_criteria.#":                                    acctest.Ct1,
						"source_selection_criteria.0.sse_kms_encrypted_objects.#":        acctest.Ct1,
						"source_selection_criteria.0.sse_kms_encrypted_objects.0.status": string(types.SseKmsEncryptedObjectsStatusEnabled),
						"destination.0.storage_class":                                    string(types.StorageClassStandard),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.account", callerIdentityDataSourceName, names.AttrAccountID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.encryption_configuration.0.replica_kms_key_id", kmsKeyResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	callerIdentityDataSourceName := "data.aws_caller_identity.current"
	dstBucketResourceName := "aws_s3_bucket.destination"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_s3_bucket_replication_configuration.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationConfig_rulesDestination(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRole, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:                  "foobar",
						names.AttrPrefix:              "foo",
						names.AttrStatus:              string(types.ReplicationRuleStatusEnabled),
						"destination.#":               acctest.Ct1,
						"destination.0.storage_class": string(types.StorageClassStandard),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.account", callerIdentityDataSourceName, names.AttrAccountID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketReplicationConfigurationConfig_accessControlTranslation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRole, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:     "foobar",
						names.AttrPrefix: "foo",
						names.AttrStatus: string(types.ReplicationRuleStatusEnabled),
						"destination.#":  acctest.Ct1,
						"destination.0.access_control_translation.#":       acctest.Ct1,
						"destination.0.access_control_translation.0.owner": string(types.OwnerOverrideDestination),
						"destination.0.storage_class":                      string(types.StorageClassStandard),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.account", callerIdentityDataSourceName, names.AttrAccountID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dstBucketResourceName := "aws_s3_bucket.destination"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_s3_bucket_replication_configuration.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationConfig_rtc(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRole, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:                                        "foobar",
						"filter.#":                                          acctest.Ct1,
						"filter.0.prefix":                                   "foo",
						names.AttrStatus:                                    string(types.ReplicationRuleStatusEnabled),
						"delete_marker_replication.#":                       acctest.Ct1,
						"delete_marker_replication.0.status":                string(types.DeleteMarkerReplicationStatusEnabled),
						"destination.#":                                     acctest.Ct1,
						"destination.0.replication_time.#":                  acctest.Ct1,
						"destination.0.replication_time.0.status":           string(types.ReplicationTimeStatusEnabled),
						"destination.0.replication_time.0.time.#":           acctest.Ct1,
						"destination.0.replication_time.0.time.0.minutes":   "15",
						"destination.0.metrics.#":                           acctest.Ct1,
						"destination.0.metrics.0.status":                    string(types.MetricsStatusEnabled),
						"destination.0.metrics.0.event_threshold.#":         acctest.Ct1,
						"destination.0.metrics.0.event_threshold.0.minutes": "15",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dstBucketResourceName := "aws_s3_bucket.destination"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_s3_bucket_replication_configuration.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationConfig_replicaMods(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRole, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:                         "foobar",
						"filter.#":                           acctest.Ct1,
						"filter.0.prefix":                    "foo",
						"delete_marker_replication.#":        acctest.Ct1,
						"delete_marker_replication.0.status": string(types.DeleteMarkerReplicationStatusEnabled),
						"source_selection_criteria.#":        acctest.Ct1,
						"source_selection_criteria.0.replica_modifications.#":        acctest.Ct1,
						"source_selection_criteria.0.replica_modifications.0.status": string(types.ReplicaModificationsStatusEnabled),
						names.AttrStatus: string(types.ReplicationRuleStatusEnabled),
						"destination.#":  acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	resourceName := "aws_s3_bucket_replication_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dstBucketResourceName := "aws_s3_bucket.destination"
	iamRoleResourceName := "aws_iam_role.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationConfig_prefixNoID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRole, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "rule.0.id"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.prefix", "foo"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.status", string(types.ReplicationRuleStatusEnabled)),
					resource.TestCheckResourceAttr(resourceName, "rule.0.destination.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "rule.0.destination.0.bucket", dstBucketResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dstBucketResourceName := "aws_s3_bucket.destination"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_s3_bucket_replication_configuration.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationConfig_noStorageClass(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRole, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:     "foobar",
						names.AttrPrefix: "foo",
						names.AttrStatus: string(types.ReplicationRuleStatusEnabled),
						"destination.#":  acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dstBucketResourceName := "aws_s3_bucket.destination"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_s3_bucket_replication_configuration.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationConfig_v2NoTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRole, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:                         "foobar",
						"filter.#":                           acctest.Ct1,
						"filter.0.prefix":                    "foo",
						"delete_marker_replication.#":        acctest.Ct1,
						"delete_marker_replication.0.status": string(types.DeleteMarkerReplicationStatusEnabled),
						names.AttrStatus:                     string(types.ReplicationRuleStatusEnabled),
						"destination.#":                      acctest.Ct1,
						"destination.0.storage_class":        string(types.StorageClassStandard),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameDestination := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dstBucketResourceName := "aws_s3_bucket.destination"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_s3_bucket_replication_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketReplicationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationConfig_schemaV2SameRegion(rName, rNameDestination),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRole, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:                         "testid",
						"filter.#":                           acctest.Ct1,
						"filter.0.prefix":                    "testprefix",
						"delete_marker_replication.#":        acctest.Ct1,
						"delete_marker_replication.0.status": string(types.DeleteMarkerReplicationStatusEnabled),
						names.AttrStatus:                     string(types.ReplicationRuleStatusEnabled),
						"destination.#":                      acctest.Ct1,
						"destination.0.storage_class":        string(types.StorageClassStandard),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	resourceName := "aws_s3_bucket_replication_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationConfig_schemaV2DestinationMetricsStatusOnly(rName, string(types.StorageClassStandard)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"destination.#":                             acctest.Ct1,
						"destination.0.metrics.#":                   acctest.Ct1,
						"destination.0.metrics.0.status":            string(types.MetricsStatusEnabled),
						"destination.0.metrics.0.event_threshold.#": acctest.Ct0,
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

	ctx := acctest.Context(t)
	resourceName := "aws_s3_bucket_replication_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameDestination := sdkacctest.RandomWithPrefix("tf-acc-test")
	dstBucketResourceName := "aws_s3_bucket.destination"
	iamRoleResourceName := "aws_iam_role.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationConfig_existingObject(rName, rNameDestination),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRole, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:                           "testid",
						"filter.#":                             acctest.Ct1,
						"filter.0.prefix":                      "testprefix",
						"delete_marker_replication.#":          acctest.Ct1,
						"delete_marker_replication.0.status":   string(types.DeleteMarkerReplicationStatusEnabled),
						"existing_object_replication.#":        acctest.Ct1,
						"existing_object_replication.0.status": string(types.ExistingObjectReplicationStatusEnabled),
						names.AttrStatus:                       string(types.ReplicationRuleStatusEnabled),
						"destination.#":                        acctest.Ct1,
						"destination.0.storage_class":          string(types.StorageClassStandard),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	resourceName := "aws_s3_bucket_replication_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dstBucketResourceName := "aws_s3_bucket.destination"
	iamRoleResourceName := "aws_iam_role.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationConfig_filterEmptyBlock(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRole, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:                         "foobar",
						"delete_marker_replication.#":        acctest.Ct1,
						"delete_marker_replication.0.status": string(types.DeleteMarkerReplicationStatusDisabled),
						"filter.#":                           acctest.Ct1,
						names.AttrStatus:                     string(types.ReplicationRuleStatusEnabled),
						"destination.#":                      acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	resourceName := "aws_s3_bucket_replication_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dstBucketResourceName := "aws_s3_bucket.destination"
	iamRoleResourceName := "aws_iam_role.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationConfig_filterEmptyPrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRole, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:                         "foobar",
						"delete_marker_replication.#":        acctest.Ct1,
						"delete_marker_replication.0.status": string(types.DeleteMarkerReplicationStatusDisabled),
						"filter.#":                           acctest.Ct1,
						"filter.0.prefix":                    "",
						names.AttrStatus:                     string(types.ReplicationRuleStatusEnabled),
						"destination.#":                      acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	resourceName := "aws_s3_bucket_replication_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dstBucketResourceName := "aws_s3_bucket.destination"
	iamRoleResourceName := "aws_iam_role.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationConfig_filterTag(rName, "testkey", "testvalue"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRole, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:                         "foobar",
						"delete_marker_replication.#":        acctest.Ct1,
						"delete_marker_replication.0.status": string(types.DeleteMarkerReplicationStatusDisabled),
						"filter.#":                           acctest.Ct1,
						"filter.0.tag.#":                     acctest.Ct1,
						"filter.0.tag.0.key":                 "testkey",
						"filter.0.tag.0.value":               "testvalue",
						names.AttrStatus:                     string(types.ReplicationRuleStatusEnabled),
						"destination.#":                      acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	resourceName := "aws_s3_bucket_replication_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dstBucketResourceName := "aws_s3_bucket.destination"
	iamRoleResourceName := "aws_iam_role.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationConfig_filterAndOperatorPrefixAndTags(rName, "testkey1", "testvalue1", "testkey2", "testvalue2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRole, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:                         "foobar",
						"delete_marker_replication.#":        acctest.Ct1,
						"delete_marker_replication.0.status": string(types.DeleteMarkerReplicationStatusDisabled),
						"filter.#":                           acctest.Ct1,
						"filter.0.and.#":                     acctest.Ct1,
						"filter.0.and.0.prefix":              "foo",
						"filter.0.and.0.tags.%":              acctest.Ct2,
						"filter.0.and.0.tags.testkey1":       "testvalue1",
						"filter.0.and.0.tags.testkey2":       "testvalue2",
						names.AttrStatus:                     string(types.ReplicationRuleStatusEnabled),
						"destination.#":                      acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketReplicationConfigurationConfig_filterAndOperatorTags(rName, "testkey1", "testvalue1", "testkey2", "testvalue2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRole, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID:                         "foobar",
						"delete_marker_replication.#":        acctest.Ct1,
						"delete_marker_replication.0.status": string(types.DeleteMarkerReplicationStatusDisabled),
						"filter.#":                           acctest.Ct1,
						"filter.0.and.#":                     acctest.Ct1,
						"filter.0.and.0.tags.%":              acctest.Ct2,
						"filter.0.and.0.tags.testkey1":       "testvalue1",
						"filter.0.and.0.tags.testkey2":       "testvalue2",
						names.AttrStatus:                     string(types.ReplicationRuleStatusEnabled),
						"destination.#":                      acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.destination.0.bucket", dstBucketResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	resourceName := "aws_s3_bucket_replication_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dstBucketResourceName := "aws_s3_bucket.destination"
	iamRoleResourceName := "aws_iam_role.test"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationConfig_filterNoID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRole, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "rule.0.id"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.status", string(types.ReplicationRuleStatusEnabled)),
					resource.TestCheckResourceAttr(resourceName, "rule.0.delete_marker_replication.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.destination.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "rule.0.destination.0.bucket", dstBucketResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	resourceName := "aws_s3_bucket_replication_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationConfig_noPrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_replication_configuration.test"
	bucketResourceName := "aws_s3_bucket.source"
	region := acctest.Region()

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_replicationV2PrefixAndTags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExistsWithProvider(ctx, bucketResourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(bucketResourceName, "replication_configuration.0.rules.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(bucketResourceName, "replication_configuration.0.rules.*", map[string]string{
						"filter.#":        acctest.Ct1,
						"filter.0.prefix": "foo",
						"filter.0.tags.%": acctest.Ct2,
					}),
				),
			},
			{
				Config: testAccBucketReplicationConfigurationConfig_migrateNoChange(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.0.and.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.0.and.0.prefix", "foo"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.0.and.0.tags.%", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccS3BucketReplicationConfiguration_migrate_withChange(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_replication_configuration.test"
	bucketResourceName := "aws_s3_bucket.source"
	region := acctest.Region()

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_replicationV2PrefixAndTags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExistsWithProvider(ctx, bucketResourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(bucketResourceName, "replication_configuration.0.rules.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(bucketResourceName, "replication_configuration.0.rules.*", map[string]string{
						"filter.#":        acctest.Ct1,
						"filter.0.prefix": "foo",
						"filter.0.tags.%": acctest.Ct2,
					}),
				),
			},
			{
				Config: testAccBucketReplicationConfigurationConfig_migrateChange(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.0.prefix", "bar"),
				),
			},
		},
	})
}

func TestAccS3BucketReplicationConfiguration_directoryBucket(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketReplicationConfigurationConfig_directoryBucket(rName, string(types.StorageClassStandard)),
				ExpectError: regexache.MustCompile(`directory buckets are not supported`),
			},
		},
	})
}

// testAccCheckBucketReplicationConfigurationDestroy is the equivalent of the "WithProvider"
// version, but for use with "same region" tests requiring only one provider.
func testAccCheckBucketReplicationConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_bucket_replication_configuration" {
				continue
			}

			_, err := tfs3.FindReplicationConfiguration(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Bucket Replication Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketReplicationConfigurationDestroyWithProvider(ctx context.Context) acctest.TestCheckWithProviderFunc {
	return func(s *terraform.State, provider *schema.Provider) error {
		conn := provider.Meta().(*conns.AWSClient).S3Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_bucket_replication_configuration" {
				continue
			}

			_, err := tfs3.FindReplicationConfiguration(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Bucket Replication Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketReplicationConfigurationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		_, err := tfs3.FindReplicationConfiguration(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccBucketReplicationConfigurationConfig_base(rName string) string {
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

func testAccBucketReplicationConfigurationConfig_basic(rName, storageClass string) string {
	return acctest.ConfigCompose(testAccBucketReplicationConfigurationConfig_base(rName), fmt.Sprintf(`
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
}`, storageClass))
}

func testAccBucketReplicationConfigurationConfig_prefixNoID(rName string) string {
	return acctest.ConfigCompose(testAccBucketReplicationConfigurationConfig_base(rName), `
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

func testAccBucketReplicationConfigurationConfig_filterNoID(rName string) string {
	return acctest.ConfigCompose(testAccBucketReplicationConfigurationConfig_base(rName), `
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

func testAccBucketReplicationConfigurationConfig_rtc(rName string) string {
	return acctest.ConfigCompose(testAccBucketReplicationConfigurationConfig_base(rName), `
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

func testAccBucketReplicationConfigurationConfig_replicaMods(rName string) string {
	return acctest.ConfigCompose(testAccBucketReplicationConfigurationConfig_base(rName), `
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
}`)
}

func testAccBucketReplicationConfigurationConfig_multipleDestinationsEmptyFilter(rName string) string {
	return acctest.ConfigCompose(testAccBucketReplicationConfigurationConfig_base(rName), fmt.Sprintf(`
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

func testAccBucketReplicationConfigurationConfig_multipleDestinationsNonEmptyFilter(rName string) string {
	return acctest.ConfigCompose(testAccBucketReplicationConfigurationConfig_base(rName), fmt.Sprintf(`
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

func testAccBucketReplicationConfigurationConfig_multipleDestinationsTwoDestination(rName string) string {
	return acctest.ConfigCompose(testAccBucketReplicationConfigurationConfig_base(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccBucketReplicationConfigurationConfig_base(rName), `
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
}`)
}

func testAccBucketReplicationConfigurationConfig_accessControlTranslation(rName string) string {
	return acctest.ConfigCompose(testAccBucketReplicationConfigurationConfig_base(rName), `
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
}`)
}

func testAccBucketReplicationConfigurationConfig_rulesDestination(rName string) string {
	return acctest.ConfigCompose(testAccBucketReplicationConfigurationConfig_base(rName), `
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
}`)
}

func testAccBucketReplicationConfigurationConfig_sseKMSEncryptedObjectsAndAccessControlTranslation(rName string) string {
	return acctest.ConfigCompose(testAccBucketReplicationConfigurationConfig_base(rName), `
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
}`)
}

func testAccBucketReplicationConfigurationConfig_noStorageClass(rName string) string {
	return acctest.ConfigCompose(testAccBucketReplicationConfigurationConfig_base(rName), `
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
}`)
}

func testAccBucketReplicationConfigurationConfig_v2NoTags(rName string) string {
	return acctest.ConfigCompose(testAccBucketReplicationConfigurationConfig_base(rName), `
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
}`)
}

func testAccBucketReplicationConfigurationConfig_schemaV2SameRegion(rName, rNameDestination string) string {
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

func testAccBucketReplicationConfigurationConfig_existingObject(rName, rNameDestination string) string {
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

func testAccBucketReplicationConfigurationConfig_filterEmptyBlock(rName string) string {
	return acctest.ConfigCompose(testAccBucketReplicationConfigurationConfig_base(rName), `
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

func testAccBucketReplicationConfigurationConfig_filterEmptyPrefix(rName string) string {
	return acctest.ConfigCompose(testAccBucketReplicationConfigurationConfig_base(rName), `
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
}`)
}

func testAccBucketReplicationConfigurationConfig_filterTag(rName, key, value string) string {
	return acctest.ConfigCompose(testAccBucketReplicationConfigurationConfig_base(rName), fmt.Sprintf(`
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

func testAccBucketReplicationConfigurationConfig_filterAndOperatorTags(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(testAccBucketReplicationConfigurationConfig_base(rName), fmt.Sprintf(`
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

func testAccBucketReplicationConfigurationConfig_filterAndOperatorPrefixAndTags(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(testAccBucketReplicationConfigurationConfig_base(rName), fmt.Sprintf(`
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

func testAccBucketReplicationConfigurationConfig_schemaV2DestinationMetricsStatusOnly(rName, storageClass string) string {
	return acctest.ConfigCompose(testAccBucketReplicationConfigurationConfig_base(rName), fmt.Sprintf(`
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
}`, storageClass))
}

func testAccBucketReplicationConfigurationConfig_noPrefix(rName string) string {
	return acctest.ConfigCompose(testAccBucketReplicationConfigurationConfig_base(rName), `
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

func testAccBucketReplicationConfigurationConfig_migrationBase(rName string) string {
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

func testAccBucketReplicationConfigurationConfig_migrateNoChange(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), testAccBucketReplicationConfigurationConfig_migrationBase(rName), `
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
}`)
}

func testAccBucketReplicationConfigurationConfig_migrateChange(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), testAccBucketReplicationConfigurationConfig_migrationBase(rName), `
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
}`)
}

func testAccBucketReplicationConfigurationConfig_directoryBucket(rName, storageClass string) string {
	return acctest.ConfigCompose(testAccBucketReplicationConfigurationConfig_base(rName), testAccDirectoryBucketConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket
  location {
    name = local.location_name
  }
}
resource "aws_s3_bucket_replication_configuration" "test" {
  depends_on = [
    aws_s3_bucket_versioning.source,
    aws_s3_bucket_versioning.destination
  ]
  bucket = aws_s3_directory_bucket.test.bucket
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
}`, storageClass))
}
