package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccAWSS3BucketReplicationConfig_1basic(t *testing.T) {
	rInt := acctest.RandInt()
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
				),
			},
		},
	})
}

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
	alternateRegion := testAccGetAlternateRegion()
	region := testAccGetRegion()
	resourceName := "aws_s3_bucket.bucket"

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
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination2", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination3", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						"id":                          "rule1",
						"priority":                    "1",
						"status":                      "Enabled",
						"filter.#":                    "1",
						"filter.0.prefix":             "",
						"destination.#":               "1",
						"destination.0.storage_class": "STANDARD",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						"id":                          "rule2",
						"priority":                    "2",
						"status":                      "Enabled",
						"filter.#":                    "1",
						"filter.0.prefix":             "",
						"destination.#":               "1",
						"destination.0.storage_class": "STANDARD_IA",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
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
	alternateRegion := testAccGetAlternateRegion()
	region := testAccGetRegion()
	resourceName := "aws_s3_bucket.bucket"

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
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination2", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination3", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						"id":                          "rule1",
						"priority":                    "1",
						"status":                      "Enabled",
						"filter.#":                    "1",
						"filter.0.prefix":             "prefix1",
						"destination.#":               "1",
						"destination.0.storage_class": "STANDARD",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						"id":                          "rule2",
						"priority":                    "2",
						"status":                      "Enabled",
						"filter.#":                    "1",
						"filter.0.tags.%":             "1",
						"filter.0.tags.Key2":          "Value2",
						"destination.#":               "1",
						"destination.0.storage_class": "STANDARD_IA",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
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
	alternateRegion := testAccGetAlternateRegion()
	region := testAccGetRegion()
	resourceName := "aws_s3_bucket.bucket"

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
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination2", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						"id":                          "rule1",
						"priority":                    "1",
						"status":                      "Enabled",
						"filter.#":                    "1",
						"filter.0.prefix":             "prefix1",
						"destination.#":               "1",
						"destination.0.storage_class": "STANDARD",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
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
	region := testAccGetRegion()
	partition := testAccGetPartition()
	iamRoleResourceName := "aws_iam_role.role"
	resourceName := "aws_s3_bucket.bucket"

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
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
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
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
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
	region := testAccGetRegion()
	partition := testAccGetPartition()
	iamRoleResourceName := "aws_iam_role.role"
	resourceName := "aws_s3_bucket.bucket"

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
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
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
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
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
	alternateRegion := testAccGetAlternateRegion()
	region := testAccGetRegion()
	resourceName := "aws_s3_bucket.bucket"

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
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
				),
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

func TestAccAWSS3BucketReplicationConfig_expectVersioningValidationError(t *testing.T) {
	rInt := acctest.RandInt()

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
				Config:      testAccAWSS3BucketReplicationConfigNoVersioning(rInt),
				ExpectError: regexp.MustCompile(`versioning must be enabled to allow S3 bucket replication`),
			},
		},
	})
}

// Prefix issue: https://github.com/hashicorp/terraform-provider-aws/issues/6340
func TestAccAWSS3BucketReplicationConfig_withoutPrefix(t *testing.T) {
	rInt := acctest.RandInt()
	alternateRegion := testAccGetAlternateRegion()
	region := testAccGetRegion()
	resourceName := "aws_s3_bucket.bucket"

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
				Config: testAccAWSS3BucketReplicationConfigWithoutPrefix(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
				),
			},
			{
				Config:                  testAccAWSS3BucketReplicationConfigWithoutPrefix(rInt),
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
	alternateRegion := testAccGetAlternateRegion()
	region := testAccGetRegion()
	partition := testAccGetPartition()
	iamRoleResourceName := "aws_iam_role.role"
	resourceName := "aws_s3_bucket.bucket"

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
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
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
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
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
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
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
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
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
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
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
	resourceName := "aws_s3_bucket.bucket"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationResourceName := "aws_s3_bucket.destination"
	rNameDestination := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, s3.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigSameRegionReplicationWithV2ConfigurationNoTags(rName, rNameDestination),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					testAccCheckResourceAttrGlobalARN(resourceName, "replication_configuration.0.role", "iam", fmt.Sprintf("role/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketExists(destinationResourceName),
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
				Config:            testAccAWSS3BucketConfigSameRegionReplicationWithV2ConfigurationNoTags(rName, rNameDestination),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl"},
			},
		},
	})
}

func testAccAWSS3BucketReplicationConfigBasic(randInt int) string {
	return testAccMultipleRegionProviderConfig(2) + fmt.Sprintf(`
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
}

resource "aws_s3_bucket" "source" {
  bucket   = "tf-test-bucket-source-%[1]d"

  versioning {
    enabled = true
  }
}
`, randInt)
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
}
`, storageClass)
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
}

resource "aws_s3_bucket" "destination3" {
  provider = "awsalternate"
  bucket   = "tf-test-bucket-destination3-%[1]d"

  versioning {
    enabled = true
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
  }
}
`, randInt))
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
}

resource "aws_s3_bucket" "destination3" {
  provider = "awsalternate"
  bucket   = "tf-test-bucket-destination3-%[1]d"

  versioning {
    enabled = true
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
}
`, randInt))
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
}
`, randInt))
}

func testAccAWSS3BucketReplicationConfigWithSseKmsEncryptedObjects(randInt int) string {
	return testAccAWSS3BucketReplicationConfigBasic(randInt) + fmt.Sprintf(`
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
}
`)
}

func testAccAWSS3BucketReplicationConfigWithAccessControlTranslation(randInt int) string {
	return testAccAWSS3BucketReplicationConfigBasic(randInt) + fmt.Sprintf(`
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
}
`)
}

func testAccAWSS3BucketReplicationConfigConfigurationRulesDestination(randInt int) string {
	return testAccAWSS3BucketReplicationConfigBasic(randInt) + fmt.Sprintf(`
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
}
`)
}

func testAccAWSS3BucketReplicationConfigWithSseKmsEncryptedObjectsAndAccessControlTranslation(randInt int) string {
	return testAccAWSS3BucketReplicationConfigBasic(randInt) + fmt.Sprintf(`
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
}
`)
}

func testAccAWSS3BucketReplicationConfigWithoutStorageClass(randInt int) string {
	return testAccAWSS3BucketReplicationConfigBasic(randInt) + fmt.Sprintf(`
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
}
`)
}

func testAccAWSS3BucketReplicationConfigWithoutPrefix(randInt int) string {
	return testAccAWSS3BucketReplicationConfigBasic(randInt) + fmt.Sprintf(`
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
}
`)
}

func testAccAWSS3BucketReplicationConfigNoVersioning(randInt int) string {
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
        storage_class = "STANDARD"
      }
    }
}
`)
}

func testAccAWSS3BucketReplicationConfigSameRegionReplicationWithV2ConfigurationNoTags(rName, rNameDestination string) string {
	return composeConfig(testAccAWSS3BucketReplicationConfig_iamPolicy(rName), fmt.Sprintf(`
resource "aws_s3_bucket_replication_configuration" "replication" {
  bucket = %[1]q
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
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
  }
}

resource "aws_s3_bucket_replication_configuration" "destination" {
  bucket = %[2]q

  versioning {
    enabled = true
  }
}
`, rName, rNameDestination))
}

func testAccAWSS3BucketReplicationConfigWithV2ConfigurationDeleteMarkerReplicationDisabled(randInt int) string {
	return testAccAWSS3BucketReplicationConfigBasic(randInt) + fmt.Sprintf(`
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
}
`)
}

func testAccAWSS3BucketReplicationConfigWithV2ConfigurationNoTags(randInt int) string {
	return testAccAWSS3BucketReplicationConfigBasic(randInt) + fmt.Sprintf(`
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
}
`)
}

func testAccAWSS3BucketReplicationConfigWithV2ConfigurationOnlyOneTag(randInt int) string {
	return testAccAWSS3BucketReplicationConfigBasic(randInt) + fmt.Sprintf(`
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
}
`)
}

func testAccAWSS3BucketReplicationConfigWithV2ConfigurationPrefixAndTags(randInt int) string {
	return testAccAWSS3BucketReplicationConfigBasic(randInt) + fmt.Sprintf(`
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
}
`)
}

func testAccAWSS3BucketReplicationConfigWithV2ConfigurationMultipleTags(randInt int) string {
	return testAccAWSS3BucketReplicationConfigBasic(randInt) + fmt.Sprintf(`
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
}
`)
}
