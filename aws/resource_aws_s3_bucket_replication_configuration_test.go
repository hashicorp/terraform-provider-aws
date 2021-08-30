package aws

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSS3BucketReplicationConfig_basic(t *testing.T) {
	rInt := acctest.RandInt()
	partition := testAccGetPartition()
	iamRoleResourceName := "aws_iam_role.role"
	resourceName := "aws_s3_bucket_replication_configuration.replication"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 2)
		},
		ErrorCheck:        testAccErrorCheck(t, s3.EndpointsID),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckAWSS3BucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketReplicationConfig(rInt, "STANDARD"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					testAccCheckAWSS3BucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.StorageClassStandard),
								},
								Prefix: aws.String("foo"),
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
							},
						},
					),
				),
			},
			{
				Config: testAccAWSS3BucketReplicationConfig(rInt, "GLACIER"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					testAccCheckAWSS3BucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.StorageClassGlacier),
								},
								Prefix: aws.String("foo"),
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
							},
						},
					),
				),
			},
			{
				Config: testAccAWSS3BucketReplicationConfigWithSseKmsEncryptedObjects(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					testAccCheckAWSS3BucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
									EncryptionConfiguration: &s3.EncryptionConfiguration{
										ReplicaKmsKeyID: aws.String("${aws_kms_key.replica.arn}"),
									},
								},
								Prefix: aws.String("foo"),
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
								SourceSelectionCriteria: &s3.SourceSelectionCriteria{
									SseKmsEncryptedObjects: &s3.SseKmsEncryptedObjects{
										Status: aws.String(s3.SseKmsEncryptedObjectsStatusEnabled),
									},
								},
							},
						},
					),
				),
			},
		},
	})
}

func TestAccAWSS3BucketReplicationConfig_multipleDestinationsEmptyFilter(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_s3_bucket_replication_configuration.replication"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 2)
		},
		ErrorCheck:        testAccErrorCheckSkipS3(t),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckAWSS3BucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketReplicationConfigWithMultipleDestinationsEmptyFilter(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "rules.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rules.*", map[string]string{
						"id":                          "rule1",
						"priority":                    "1",
						"status":                      "Enabled",
						"filter.#":                    "1",
						"filter.0.prefix":             "",
						"destination.#":               "1",
						"destination.0.storage_class": "STANDARD",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rules.*", map[string]string{
						"id":                          "rule2",
						"priority":                    "2",
						"status":                      "Enabled",
						"filter.#":                    "1",
						"filter.0.prefix":             "",
						"destination.#":               "1",
						"destination.0.storage_class": "STANDARD_IA",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rules.*", map[string]string{
						"id":                          "rule3",
						"priority":                    "3",
						"status":                      "Disabled",
						"filter.#":                    "1",
						"filter.0.prefix":             "",
						"destination.#":               "1",
						"destination.0.storage_class": "ONEZONE_IA",
					}),
				),
			},
			{
				Config:                  testAccAWSS3BucketReplicationConfigWithMultipleDestinationsEmptyFilter(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccAWSS3BucketReplicationConfig_multipleDestinationsNonEmptyFilter(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_s3_bucket_replication_configuration.replication"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 2)
		},
		ErrorCheck:        testAccErrorCheckSkipS3(t),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckAWSS3BucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketReplicationConfigWithMultipleDestinationsNonEmptyFilter(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "rules.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rules.*", map[string]string{
						"id":                          "rule1",
						"priority":                    "1",
						"status":                      "Enabled",
						"filter.#":                    "1",
						"filter.0.prefix":             "prefix1",
						"destination.#":               "1",
						"destination.0.storage_class": "STANDARD",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rules.*", map[string]string{
						"id":                          "rule2",
						"priority":                    "2",
						"status":                      "Enabled",
						"filter.#":                    "1",
						"filter.0.tags.%":             "1",
						"filter.0.tags.Key2":          "Value2",
						"destination.#":               "1",
						"destination.0.storage_class": "STANDARD_IA",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rules.*", map[string]string{
						"id":                          "rule3",
						"priority":                    "3",
						"status":                      "Disabled",
						"filter.#":                    "1",
						"filter.0.prefix":             "prefix3",
						"filter.0.tags.%":             "1",
						"filter.0.tags.Key3":          "Value3",
						"destination.#":               "1",
						"destination.0.storage_class": "ONEZONE_IA",
					}),
				),
			},
			{
				Config:                  testAccAWSS3BucketReplicationConfigWithMultipleDestinationsNonEmptyFilter(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccAWSS3BucketReplicationConfig_twoDestination(t *testing.T) {
	// This tests 2 destinations since GovCloud and possibly other non-standard partitions allow a max of 2
	rInt := acctest.RandInt()
	resourceName := "aws_s3_bucket_replication_configuration.replication"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 2)
		},
		ErrorCheck:        testAccErrorCheckSkipS3(t),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckAWSS3BucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketReplicationConfigWithMultipleDestinationsTwoDestination(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "rules.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rules.*", map[string]string{
						"id":                          "rule1",
						"priority":                    "1",
						"status":                      "Enabled",
						"filter.#":                    "1",
						"filter.0.prefix":             "prefix1",
						"destination.#":               "1",
						"destination.0.storage_class": "STANDARD",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rules.*", map[string]string{
						"id":                          "rule2",
						"priority":                    "2",
						"status":                      "Enabled",
						"filter.#":                    "1",
						"filter.0.tags.%":             "1",
						"filter.0.tags.Key2":          "Value2",
						"destination.#":               "1",
						"destination.0.storage_class": "STANDARD_IA",
					}),
				),
			},
			{
				Config:                  testAccAWSS3BucketReplicationConfigWithMultipleDestinationsTwoDestination(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccAWSS3BucketReplicationConfig_configurationRuleDestinationAccessControlTranslation(t *testing.T) {
	rInt := acctest.RandInt()
	partition := testAccGetPartition()
	iamRoleResourceName := "aws_iam_role.role"
	resourceName := "aws_s3_bucket_replication_configuration.replication"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 2)
		},
		ErrorCheck:        testAccErrorCheck(t, s3.EndpointsID),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckAWSS3BucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketReplicationConfigWithAccessControlTranslation(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					testAccCheckAWSS3BucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Account:      aws.String("${data.aws_caller_identity.current.account_id}"),
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
									AccessControlTranslation: &s3.AccessControlTranslation{
										Owner: aws.String("Destination"),
									},
								},
								Prefix: aws.String("foo"),
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
							},
						},
					),
				),
			},
			{
				Config:                  testAccAWSS3BucketReplicationConfigWithAccessControlTranslation(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl", "versioning"},
			},
			{
				Config: testAccAWSS3BucketReplicationConfigWithSseKmsEncryptedObjectsAndAccessControlTranslation(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					testAccCheckAWSS3BucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Account:      aws.String("${data.aws_caller_identity.current.account_id}"),
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
									EncryptionConfiguration: &s3.EncryptionConfiguration{
										ReplicaKmsKeyID: aws.String("${aws_kms_key.replica.arn}"),
									},
									AccessControlTranslation: &s3.AccessControlTranslation{
										Owner: aws.String("Destination"),
									},
								},
								Prefix: aws.String("foo"),
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
								SourceSelectionCriteria: &s3.SourceSelectionCriteria{
									SseKmsEncryptedObjects: &s3.SseKmsEncryptedObjects{
										Status: aws.String(s3.SseKmsEncryptedObjectsStatusEnabled),
									},
								},
							},
						},
					),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/12480
func TestAccAWSS3BucketReplicationConfig_configurationRuleDestinationAddAccessControlTranslation(t *testing.T) {
	rInt := acctest.RandInt()
	partition := testAccGetPartition()
	iamRoleResourceName := "aws_iam_role.role"
	resourceName := "aws_s3_bucket_replication_configuration.replication"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 2)
		},
		ErrorCheck:        testAccErrorCheck(t, s3.EndpointsID),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckAWSS3BucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketReplicationConfigConfigurationRulesDestination(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					testAccCheckAWSS3BucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Account:      aws.String("${data.aws_caller_identity.current.account_id}"),
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
								},
								Prefix: aws.String("foo"),
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
							},
						},
					),
				),
			},
			{
				Config:                  testAccAWSS3BucketReplicationConfigWithAccessControlTranslation(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl", "versioning"},
			},
			{
				Config: testAccAWSS3BucketReplicationConfigWithAccessControlTranslation(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					testAccCheckAWSS3BucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Account:      aws.String("${data.aws_caller_identity.current.account_id}"),
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
									AccessControlTranslation: &s3.AccessControlTranslation{
										Owner: aws.String("Destination"),
									},
								},
								Prefix: aws.String("foo"),
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
							},
						},
					),
				),
			},
		},
	})
}

// StorageClass issue: https://github.com/hashicorp/terraform/issues/10909
func TestAccAWSS3BucketReplicationConfig_withoutStorageClass(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_s3_bucket_replication_configuration.replication"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 2)
		},
		ErrorCheck:        testAccErrorCheck(t, s3.EndpointsID),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckAWSS3BucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketReplicationConfigWithoutStorageClass(rInt),
				Check:  resource.ComposeTestCheckFunc(),
			},
			{
				Config:                  testAccAWSS3BucketReplicationConfigWithoutStorageClass(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccAWSS3BucketReplicationConfig_schemaV2(t *testing.T) {
	rInt := acctest.RandInt()
	partition := testAccGetPartition()
	iamRoleResourceName := "aws_iam_role.role"
	resourceName := "aws_s3_bucket_replication_configuration.replication"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 2)
		},
		ErrorCheck:        testAccErrorCheck(t, s3.EndpointsID),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckAWSS3BucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketReplicationConfigWithV2ConfigurationDeleteMarkerReplicationDisabled(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					testAccCheckAWSS3BucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
								},
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
								Filter: &s3.ReplicationRuleFilter{
									Prefix: aws.String("foo"),
								},
								Priority: aws.Int64(0),
								DeleteMarkerReplication: &s3.DeleteMarkerReplication{
									Status: aws.String(s3.DeleteMarkerReplicationStatusDisabled),
								},
							},
						},
					),
				),
			},
			{
				Config: testAccAWSS3BucketReplicationConfigWithV2ConfigurationNoTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					testAccCheckAWSS3BucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
								},
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
								Filter: &s3.ReplicationRuleFilter{
									Prefix: aws.String("foo"),
								},
								Priority: aws.Int64(0),
								DeleteMarkerReplication: &s3.DeleteMarkerReplication{
									Status: aws.String(s3.DeleteMarkerReplicationStatusEnabled),
								},
							},
						},
					),
				),
			},
			{
				Config:                  testAccAWSS3BucketReplicationConfigWithV2ConfigurationNoTags(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccAWSS3BucketReplicationConfigWithV2ConfigurationOnlyOneTag(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					testAccCheckAWSS3BucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
								},
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
								Filter: &s3.ReplicationRuleFilter{
									And: &s3.ReplicationRuleAndOperator{
										Prefix: aws.String(""),
										Tags: []*s3.Tag{
											{
												Key:   aws.String("ReplicateMe"),
												Value: aws.String("Yes"),
											},
										},
									},
								},
								Priority: aws.Int64(42),
								DeleteMarkerReplication: &s3.DeleteMarkerReplication{
									Status: aws.String(s3.DeleteMarkerReplicationStatusDisabled),
								},
							},
						},
					),
				),
			},
			{
				Config: testAccAWSS3BucketReplicationConfigWithV2ConfigurationPrefixAndTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					testAccCheckAWSS3BucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
								},
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
								Filter: &s3.ReplicationRuleFilter{
									And: &s3.ReplicationRuleAndOperator{
										Prefix: aws.String("foo"),
										Tags: []*s3.Tag{
											{
												Key:   aws.String("ReplicateMe"),
												Value: aws.String("Yes"),
											},
											{
												Key:   aws.String("AnotherTag"),
												Value: aws.String("OK"),
											},
										},
									},
								},
								Priority: aws.Int64(41),
								DeleteMarkerReplication: &s3.DeleteMarkerReplication{
									Status: aws.String(s3.DeleteMarkerReplicationStatusDisabled),
								},
							},
						},
					),
				),
			},
			{
				Config: testAccAWSS3BucketReplicationConfigWithV2ConfigurationMultipleTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					testAccCheckAWSS3BucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
								},
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
								Filter: &s3.ReplicationRuleFilter{
									And: &s3.ReplicationRuleAndOperator{
										Prefix: aws.String(""),
										Tags: []*s3.Tag{
											{
												Key:   aws.String("ReplicateMe"),
												Value: aws.String("Yes"),
											},
											{
												Key:   aws.String("AnotherTag"),
												Value: aws.String("OK"),
											},
											{
												Key:   aws.String("Foo"),
												Value: aws.String("Bar"),
											},
										},
									},
								},
								Priority: aws.Int64(0),
								DeleteMarkerReplication: &s3.DeleteMarkerReplication{
									Status: aws.String(s3.DeleteMarkerReplicationStatusDisabled),
								},
							},
						},
					),
				),
			},
		},
	})
}

func TestAccAWSS3BucketReplicationConfig_schemaV2SameRegion(t *testing.T) {
	resourceName := "aws_s3_bucket_replication_configuration.replication"
	rInt := acctest.RandInt()
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameDestination := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, s3.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketReplicationConfig_schemaV2SameRegion(rName, rNameDestination, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrGlobalARN(resourceName, "role", "iam", fmt.Sprintf("role/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					testAccCheckAWSS3BucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("testid"),
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::%s", testAccGetPartition(), rNameDestination)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
								},
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
								Filter: &s3.ReplicationRuleFilter{
									Prefix: aws.String("testprefix"),
								},
								Priority: aws.Int64(0),
								DeleteMarkerReplication: &s3.DeleteMarkerReplication{
									Status: aws.String(s3.DeleteMarkerReplicationStatusEnabled),
								},
							},
						},
					),
				),
			},
			{
				Config:            testAccAWSS3BucketReplicationConfig_schemaV2SameRegion(rName, rNameDestination, rInt),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl"},
			},
		},
	})
}

const isExistingObjectReplicationBlocked = true

func TestAccAWSS3BucketReplicationConfig_existingObjectReplication(t *testing.T) {
	if isExistingObjectReplicationBlocked {
		/*  https://aws.amazon.com/blogs/storage/replicating-existing-objects-between-s3-buckets/
		    A request to AWS Technical Support needs to be made in order to allow ExistingObjectReplication.
			Once that request is approved, this can be unblocked for testing. */
		return
	}
	resourceName := "aws_s3_bucket_replication_configuration.replication"
	rInt := acctest.RandInt()
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameDestination := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, s3.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketReplicationConfig_existingObjectReplication(rName, rNameDestination, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrGlobalARN(resourceName, "role", "iam", fmt.Sprintf("role/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					testAccCheckAWSS3BucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("testid"),
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::%s", testAccGetPartition(), rNameDestination)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
								},
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
								Filter: &s3.ReplicationRuleFilter{
									Prefix: aws.String("testprefix"),
								},
								Priority: aws.Int64(0),
								DeleteMarkerReplication: &s3.DeleteMarkerReplication{
									Status: aws.String(s3.DeleteMarkerReplicationStatusEnabled),
								},
								ExistingObjectReplication: &s3.ExistingObjectReplication{
									Status: aws.String(s3.ExistingObjectReplicationStatusEnabled),
								},
							},
						},
					),
				),
			},
			{
				Config:            testAccAWSS3BucketReplicationConfig_existingObjectReplication(rName, rNameDestination, rInt),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl"},
			},
		},
	})
}

func testAccCheckAWSS3BucketReplicationRules(n string, rules []*s3.ReplicationRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		for _, rule := range rules {
			if dest := rule.Destination; dest != nil {
				if account := dest.Account; account != nil && strings.HasPrefix(aws.StringValue(dest.Account), "${") {
					resourceReference := strings.Replace(aws.StringValue(dest.Account), "${", "", 1)
					resourceReference = strings.Replace(resourceReference, "}", "", 1)
					resourceReferenceParts := strings.Split(resourceReference, ".")
					resourceAttribute := resourceReferenceParts[len(resourceReferenceParts)-1]
					resourceName := strings.Join(resourceReferenceParts[:len(resourceReferenceParts)-1], ".")
					value := s.RootModule().Resources[resourceName].Primary.Attributes[resourceAttribute]
					dest.Account = aws.String(value)
				}
				if ec := dest.EncryptionConfiguration; ec != nil {
					if ec.ReplicaKmsKeyID != nil {
						key_arn := s.RootModule().Resources["aws_kms_key.replica"].Primary.Attributes["arn"]
						ec.ReplicaKmsKeyID = aws.String(strings.Replace(*ec.ReplicaKmsKeyID, "${aws_kms_key.replica.arn}", key_arn, -1))
					}
				}
			}
			// Sort filter tags by key.
			if filter := rule.Filter; filter != nil {
				if and := filter.And; and != nil {
					if tags := and.Tags; tags != nil {
						sort.Slice(tags, func(i, j int) bool { return *tags[i].Key < *tags[j].Key })
					}
				}
			}
		}

		conn := testAccProvider.Meta().(*AWSClient).s3conn
		out, err := conn.GetBucketReplication(&s3.GetBucketReplicationInput{
			Bucket: aws.String(rs.Primary.ID),
		})
		if err != nil {
			if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") {
				return fmt.Errorf("S3 bucket not found")
			}
			if rules == nil {
				return nil
			}
			return fmt.Errorf("GetReplicationConfiguration error: %v", err)
		}

		for _, rule := range out.ReplicationConfiguration.Rules {
			// Sort filter tags by key.
			if filter := rule.Filter; filter != nil {
				if and := filter.And; and != nil {
					if tags := and.Tags; tags != nil {
						sort.Slice(tags, func(i, j int) bool { return *tags[i].Key < *tags[j].Key })
					}
				}
			}
		}
		if !reflect.DeepEqual(out.ReplicationConfiguration.Rules, rules) {
			return fmt.Errorf("bad replication rules, expected: %v, got %v", rules, out.ReplicationConfiguration.Rules)
		}

		return nil
	}
}

func testAccAWSS3BucketReplicationConfigBasic(randInt int) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "role" {
  name = "tf-iam-role-replication-%[1]d"

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
  bucket   = "tf-test-bucket-destination-%[1]d"

  versioning {
    enabled = true
  }

  lifecycle {
	  ignore_changes = [replication_configuration]
  }
}

resource "aws_s3_bucket" "source" {
  bucket   = "tf-test-bucket-source-%[1]d"

  versioning {
    enabled = true
  }

  lifecycle {
	  ignore_changes = [replication_configuration]
  }
} `, randInt)
}

func testAccAWSS3BucketReplicationConfig(randInt int, storageClass string) string {
	return testAccAWSS3BucketReplicationConfigBasic(randInt) + fmt.Sprintf(`
resource "aws_s3_bucket_replication_configuration" "replication" {
  bucket = aws_s3_bucket.source.id
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "%[1]s"
      }
    }
} `, storageClass)
}

func testAccAWSS3BucketReplicationConfigWithMultipleDestinationsEmptyFilter(randInt int) string {
	return composeConfig(
		testAccAWSS3BucketReplicationConfigBasic(randInt),
		fmt.Sprintf(`
resource "aws_s3_bucket" "destination2" {
  provider = "awsalternate"
  bucket   = "tf-test-bucket-destination2-%[1]d"

  versioning {
    enabled = true
  }
  lifecycle {
	  ignore_changes = [replication_configuration]
  }
}

resource "aws_s3_bucket" "destination3" {
  provider = "awsalternate"
  bucket   = "tf-test-bucket-destination3-%[1]d"

  versioning {
    enabled = true
  }
  lifecycle {
	  ignore_changes = [replication_configuration]
  }
}

resource "aws_s3_bucket_replication_configuration" "replication" {
    bucket = aws_s3_bucket.source.id
    role = aws_iam_role.role.arn

    rules {
      id       = "rule1"
      priority = 1
      status   = "Enabled"

      filter {}

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }

    rules {
      id       = "rule2"
      priority = 2
      status   = "Enabled"

      filter {}

      destination {
        bucket        = aws_s3_bucket.destination2.arn
        storage_class = "STANDARD_IA"
      }
    }

    rules {
      id       = "rule3"
      priority = 3
      status   = "Disabled"

      filter {}

      destination {
        bucket        = aws_s3_bucket.destination3.arn
        storage_class = "ONEZONE_IA"
      }
    }
  
} `, randInt))
}

func testAccAWSS3BucketReplicationConfigWithMultipleDestinationsNonEmptyFilter(randInt int) string {
	return composeConfig(
		testAccAWSS3BucketReplicationConfigBasic(randInt),
		fmt.Sprintf(`
resource "aws_s3_bucket" "destination2" {
  provider = "awsalternate"
  bucket   = "tf-test-bucket-destination2-%[1]d"

  versioning {
    enabled = true
  }
  lifecycle {
	  ignore_changes = [replication_configuration]
  }
}

resource "aws_s3_bucket" "destination3" {
  provider = "awsalternate"
  bucket   = "tf-test-bucket-destination3-%[1]d"

  versioning {
    enabled = true
  }
  lifecycle {
	  ignore_changes = [replication_configuration]
  }
}

resource "aws_s3_bucket_replication_configuration" "replication" {
  bucket = aws_s3_bucket.source.id
    role = aws_iam_role.role.arn

    rules {
      id       = "rule1"
      priority = 1
      status   = "Enabled"

      filter {
        prefix = "prefix1"
      }

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }

    rules {
      id       = "rule2"
      priority = 2
      status   = "Enabled"

      filter {
        tags = {
          Key2 = "Value2"
        }
      }

      destination {
        bucket        = aws_s3_bucket.destination2.arn
        storage_class = "STANDARD_IA"
      }
    }

    rules {
      id       = "rule3"
      priority = 3
      status   = "Disabled"

      filter {
        prefix = "prefix3"

        tags = {
          Key3 = "Value3"
        }
      }

      destination {
        bucket        = aws_s3_bucket.destination3.arn
        storage_class = "ONEZONE_IA"
      }
    }
} `, randInt))
}

func testAccAWSS3BucketReplicationConfigWithMultipleDestinationsTwoDestination(randInt int) string {
	return composeConfig(
		testAccAWSS3BucketReplicationConfigBasic(randInt),
		fmt.Sprintf(`
resource "aws_s3_bucket" "destination2" {
  provider = "awsalternate"
  bucket   = "tf-test-bucket-destination2-%[1]d"

  versioning {
    enabled = true
  }
  lifecycle {
	  ignore_changes = [replication_configuration]
  }
}

resource "aws_s3_bucket_replication_configuration" "replication" {
  bucket = aws_s3_bucket.source.id
    role = aws_iam_role.role.arn

    rules {
      id       = "rule1"
      priority = 1
      status   = "Enabled"

      filter {
        prefix = "prefix1"
      }

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }

    rules {
      id       = "rule2"
      priority = 2
      status   = "Enabled"

      filter {
        tags = {
          Key2 = "Value2"
        }
      }

      destination {
        bucket        = aws_s3_bucket.destination2.arn
        storage_class = "STANDARD_IA"
      }
    }
} `, randInt))
}

func testAccAWSS3BucketReplicationConfigWithSseKmsEncryptedObjects(randInt int) string {
	return testAccAWSS3BucketReplicationConfigBasic(randInt) + `
resource "aws_kms_key" "replica" {
  provider                = "awsalternate"
  description             = "TF Acceptance Test S3 repl KMS key"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket_replication_configuration" "replication" {
  bucket = aws_s3_bucket.source.id
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        bucket             = aws_s3_bucket.destination.arn
        storage_class      = "STANDARD"
        replica_kms_key_id = aws_kms_key.replica.arn
      }

      source_selection_criteria {
        sse_kms_encrypted_objects {
          enabled = true
        }
      }
    }
} `
}

func testAccAWSS3BucketReplicationConfigWithAccessControlTranslation(randInt int) string {
	return testAccAWSS3BucketReplicationConfigBasic(randInt) + `
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket_replication_configuration" "replication" {
  bucket = aws_s3_bucket.source.id
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        account_id    = data.aws_caller_identity.current.account_id
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"

        access_control_translation {
          owner = "Destination"
        }
      }
    }
} `
}

func testAccAWSS3BucketReplicationConfigConfigurationRulesDestination(randInt int) string {
	return testAccAWSS3BucketReplicationConfigBasic(randInt) + `
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket_replication_configuration" "replication" {
  bucket = aws_s3_bucket.source.id
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        account_id    = data.aws_caller_identity.current.account_id
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
} `
}

func testAccAWSS3BucketReplicationConfigWithSseKmsEncryptedObjectsAndAccessControlTranslation(randInt int) string {
	return testAccAWSS3BucketReplicationConfigBasic(randInt) + `
data "aws_caller_identity" "current" {}

resource "aws_kms_key" "replica" {
  provider                = "awsalternate"
  description             = "TF Acceptance Test S3 repl KMS key"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket_replication_configuration" "replication" {
  bucket = aws_s3_bucket.source.id
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        account_id         = data.aws_caller_identity.current.account_id
        bucket             = aws_s3_bucket.destination.arn
        storage_class      = "STANDARD"
        replica_kms_key_id = aws_kms_key.replica.arn

        access_control_translation {
          owner = "Destination"
        }
      }

      source_selection_criteria {
        sse_kms_encrypted_objects {
          enabled = true
        }
      }
    }
} `
}

func testAccAWSS3BucketReplicationConfigWithoutStorageClass(randInt int) string {
	return testAccAWSS3BucketReplicationConfigBasic(randInt) + `
resource "aws_s3_bucket_replication_configuration" "replication" {
  bucket = aws_s3_bucket.source.id
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        bucket = aws_s3_bucket.destination.arn
      }
    }
} `
}

func testAccAWSS3BucketReplicationConfigWithoutPrefix(randInt int) string {
	return testAccAWSS3BucketReplicationConfigBasic(randInt) + `
resource "aws_s3_bucket_replication_configuration" "replication" {
  bucket = aws_s3_bucket.source.id
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      status = "Enabled"

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
} `
}

func testAccAWSS3BucketReplicationConfigWithV2ConfigurationDeleteMarkerReplicationDisabled(randInt int) string {
	return testAccAWSS3BucketReplicationConfigBasic(randInt) + `
resource "aws_s3_bucket_replication_configuration" "replication" {
  bucket = aws_s3_bucket.source.id
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      status = "Enabled"

      filter {
        prefix = "foo"
      }

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
} `
}

func testAccAWSS3BucketReplicationConfigWithV2ConfigurationNoTags(randInt int) string {
	return testAccAWSS3BucketReplicationConfigBasic(randInt) + `
resource "aws_s3_bucket_replication_configuration" "replication" {
  bucket = aws_s3_bucket.source.id
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      status = "Enabled"

      filter {
        prefix = "foo"
      }

      delete_marker_replication_status = "Enabled"

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
} `
}

func testAccAWSS3BucketReplicationConfigWithV2ConfigurationOnlyOneTag(randInt int) string {
	return testAccAWSS3BucketReplicationConfigBasic(randInt) + `
resource "aws_s3_bucket_replication_configuration" "replication" {
  bucket = aws_s3_bucket.source.id
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      status = "Enabled"

      priority = 42

      filter {
        tags = {
          ReplicateMe = "Yes"
        }
      }

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
} `
}

func testAccAWSS3BucketReplicationConfigWithV2ConfigurationPrefixAndTags(randInt int) string {
	return testAccAWSS3BucketReplicationConfigBasic(randInt) + `
resource "aws_s3_bucket_replication_configuration" "replication" {
  bucket = aws_s3_bucket.source.id
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      status = "Enabled"

      priority = 41

      filter {
        prefix = "foo"

        tags = {
          AnotherTag  = "OK"
          ReplicateMe = "Yes"
        }
      }

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
} `
}

func testAccAWSS3BucketReplicationConfigWithV2ConfigurationMultipleTags(randInt int) string {
	return testAccAWSS3BucketReplicationConfigBasic(randInt) + `
resource "aws_s3_bucket_replication_configuration" "replication" {
  bucket = aws_s3_bucket.source.id
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      status = "Enabled"

      filter {
        tags = {
          AnotherTag  = "OK"
          Foo         = "Bar"
          ReplicateMe = "Yes"
        }
      }

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
} `
}

func testAccAWSS3BucketReplicationConfig_schemaV2SameRegion(rName, rNameDestination string, rInt int) string {
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
  bucket   = %[2]q

  versioning {
    enabled = true
  }

  lifecycle {
	  ignore_changes = [replication_configuration]
  }
}

resource "aws_s3_bucket" "source" {
  bucket   = "tf-test-bucket-source-%[3]d"
  acl = "private"

  versioning {
    enabled = true
  }

  lifecycle {
	  ignore_changes = [replication_configuration]
  }
}
resource "aws_s3_bucket_replication_configuration" "replication" {
    bucket = aws_s3_bucket.source.id
    role = aws_iam_role.test.arn

    rules {
      id     = "testid"
      status = "Enabled"

      filter {
        prefix = "testprefix"
      }

      delete_marker_replication_status = "Enabled"

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
} `, rName, rNameDestination, rInt)
}

func testAccAWSS3BucketReplicationConfig_existingObjectReplication(rName, rNameDestination string, rInt int) string {
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
  bucket   = %[2]q

  versioning {
    enabled = true
  }

  lifecycle {
	  ignore_changes = [replication_configuration]
  }
}

resource "aws_s3_bucket" "source" {
  bucket   = "tf-test-bucket-source-%[3]d"
  acl = "private"

  versioning {
    enabled = true
  }

  lifecycle {
	  ignore_changes = [replication_configuration]
  }
}

resource "aws_s3_bucket_replication_configuration" "replication" {
    bucket = aws_s3_bucket.source.id
    role = aws_iam_role.test.arn

    rules {
      id     = "testid"
      status = "Enabled"

      filter {
        prefix = "testprefix"
      }

	  existing_object_replication {
		  status = "Enabled"
	  }

      delete_marker_replication_status = "Enabled"

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
} `, rName, rNameDestination, rInt)
}
