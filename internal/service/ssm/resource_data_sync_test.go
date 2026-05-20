// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssm_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfssm "github.com/hashicorp/terraform-provider-aws/internal/service/ssm"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSMResourceDataSync_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssm_resource_data_sync.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceDataSyncDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDataSyncConfig_basic(rName, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceDataSyncExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("s3_destination"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrBucketName:       knownvalue.StringExact(bucketName),
							names.AttrKMSKeyARN:        knownvalue.StringExact(""),
							names.AttrPrefix:           knownvalue.StringExact(""),
							names.AttrRegion:           knownvalue.StringExact(acctest.Region()),
							"sync_format":              tfknownvalue.StringExact(awstypes.ResourceDataSyncS3FormatJsonSerde),
							"destination_data_sharing": knownvalue.ListExact([]knownvalue.Check{}),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSSMResourceDataSync_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssm_resource_data_sync.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceDataSyncDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDataSyncConfig_basic(rName, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceDataSyncExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfssm.ResourceResourceDataSync(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSMResourceDataSync_Update_s3DestinationPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssm_resource_data_sync.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceDataSyncDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDataSyncConfig_basic(rName, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceDataSyncExists(ctx, t, resourceName),
				),
			},
			{
				Config: testAccResourceDataSyncConfig_update_s3DestinationPrefix(rName, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceDataSyncExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("s3_destination"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrBucketName:       knownvalue.StringExact(bucketName),
							names.AttrKMSKeyARN:        knownvalue.StringExact(""),
							names.AttrPrefix:           knownvalue.StringExact("test-"),
							names.AttrRegion:           knownvalue.StringExact(acctest.Region()),
							"sync_format":              tfknownvalue.StringExact(awstypes.ResourceDataSyncS3FormatJsonSerde),
							"destination_data_sharing": knownvalue.ListExact([]knownvalue.Check{}),
						}),
					})),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSSMResourceDataSync_DestinationDataSharing_sameAccount(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssm_resource_data_sync.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceDataSyncDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDataSyncConfig_destinationDataSharing_sameAccount(rName, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceDataSyncExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("s3_destination"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrBucketName: knownvalue.StringExact(bucketName),
							names.AttrKMSKeyARN:  knownvalue.StringExact(""),
							names.AttrPrefix:     knownvalue.StringExact(""),
							names.AttrRegion:     knownvalue.StringExact(acctest.Region()),
							"sync_format":        tfknownvalue.StringExact(awstypes.ResourceDataSyncS3FormatJsonSerde),
							"destination_data_sharing": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"destination_data_sharing_type": knownvalue.StringExact("Organization"),
								}),
							}),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"s3_destination.0.destination_data_sharing",
				},
			},
		},
	})
}

func TestAccSSMResourceDataSync_DestinationDataSharing_add(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssm_resource_data_sync.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceDataSyncDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDataSyncConfig_destinationDataSharing_add_setup(rName, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceDataSyncExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("s3_destination"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"destination_data_sharing": knownvalue.ListExact([]knownvalue.Check{}),
						}),
					})),
				},
			},
			{
				Config: testAccResourceDataSyncConfig_destinationDataSharing_add(rName, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceDataSyncExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

func TestAccSSMResourceDataSync_DestinationDataSharing_remove(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssm_resource_data_sync.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceDataSyncDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDataSyncConfig_destinationDataSharing_remove_setup(rName, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceDataSyncExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("s3_destination"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"destination_data_sharing": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"destination_data_sharing_type": knownvalue.StringExact("Organization"),
								}),
							}),
						}),
					})),
				},
			},
			{
				Config: testAccResourceDataSyncConfig_destinationDataSharing_remove(rName, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceDataSyncExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

func testAccCheckResourceDataSyncDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SSMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssm_resource_data_sync" {
				continue
			}

			_, err := tfssm.FindResourceDataSyncByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SSM Resource Data Sync %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckResourceDataSyncExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SSMClient(ctx)

		_, err := tfssm.FindResourceDataSyncByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccResourceDataSyncConfig_basic(rName, bucketName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_resource_data_sync" "test" {
  name = %[1]q

  s3_destination {
    bucket_name = aws_s3_bucket.test.bucket
    region      = aws_s3_bucket.test.region
  }
}

resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

data "aws_partition" "current" {}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.bucket

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "SSMBucketPermissionsCheck",
      "Effect": "Allow",
      "Principal": {
        "Service": "ssm.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "s3:GetBucketAcl",
      "Resource": "${aws_s3_bucket.test.arn}"
    },
    {
      "Sid": " SSMBucketDelivery",
      "Effect": "Allow",
      "Principal": {
        "Service": "ssm.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "s3:PutObject",
      "Resource": [
        "${aws_s3_bucket.test.arn}/*"
      ],
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "bucket-owner-full-control"
        }
      }
    }
  ]
}
      EOF
}
`, rName, bucketName)
}

func testAccResourceDataSyncConfig_update_s3DestinationPrefix(rName, bucketName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_resource_data_sync" "test" {
  name = %[1]q

  s3_destination {
    bucket_name = aws_s3_bucket.test.bucket
    region      = aws_s3_bucket.test.region
    prefix      = "test-"
  }
}

resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

data "aws_partition" "current" {}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.bucket

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "SSMBucketPermissionsCheck",
      "Effect": "Allow",
      "Principal": {
        "Service": "ssm.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "s3:GetBucketAcl",
      "Resource": "${aws_s3_bucket.test.arn}"
    },
    {
      "Sid": " SSMBucketDelivery",
      "Effect": "Allow",
      "Principal": {
        "Service": "ssm.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "s3:PutObject",
      "Resource": [
        "${aws_s3_bucket.test.arn}/*"
      ],
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "bucket-owner-full-control"
        }
      }
    }
  ]
}
      EOF
}
`, rName, bucketName)
}

func testAccResourceDataSyncConfig_destinationDataSharing_sameAccount(rName, bucketName string) string {
	return acctest.ConfigCompose(
		testAccResourceDataSyncConfig_destinationDataSharingBucketPolicy(),
		fmt.Sprintf(`
resource "aws_ssm_resource_data_sync" "test" {
  name = %[1]q

  s3_destination {
    bucket_name = aws_s3_bucket.test.bucket
    region      = aws_s3_bucket.test.region

    destination_data_sharing {
      destination_data_sharing_type = "Organization"
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}
`, rName, bucketName))
}

func testAccResourceDataSyncConfig_destinationDataSharing_add_setup(rName, bucketName string) string {
	return acctest.ConfigCompose(
		testAccResourceDataSyncConfig_destinationDataSharingBucketPolicy(),
		fmt.Sprintf(`
resource "aws_ssm_resource_data_sync" "test" {
  name = %[1]q

  s3_destination {
    bucket_name = aws_s3_bucket.test.bucket
    region      = aws_s3_bucket.test.region
  }
}

resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}
`, rName, bucketName))
}

func testAccResourceDataSyncConfig_destinationDataSharing_add(rName, bucketName string) string {
	return acctest.ConfigCompose(
		testAccResourceDataSyncConfig_destinationDataSharingBucketPolicy(),
		fmt.Sprintf(`
resource "aws_ssm_resource_data_sync" "test" {
  name = %[1]q

  s3_destination {
    bucket_name = aws_s3_bucket.test.bucket
    region      = aws_s3_bucket.test.region

    destination_data_sharing {
      destination_data_sharing_type = "Organization"
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}
`, rName, bucketName))
}

func testAccResourceDataSyncConfig_destinationDataSharing_remove_setup(rName, bucketName string) string {
	return acctest.ConfigCompose(
		testAccResourceDataSyncConfig_destinationDataSharingBucketPolicy(),
		fmt.Sprintf(`
resource "aws_ssm_resource_data_sync" "test" {
  name = %[1]q

  s3_destination {
    bucket_name = aws_s3_bucket.test.bucket
    region      = aws_s3_bucket.test.region

    destination_data_sharing {
      destination_data_sharing_type = "Organization"
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}
`, rName, bucketName))
}

func testAccResourceDataSyncConfig_destinationDataSharing_remove(rName, bucketName string) string {
	return acctest.ConfigCompose(
		testAccResourceDataSyncConfig_destinationDataSharingBucketPolicy(),
		fmt.Sprintf(`
resource "aws_ssm_resource_data_sync" "test" {
  name = %[1]q

  s3_destination {
    bucket_name = aws_s3_bucket.test.bucket
    region      = aws_s3_bucket.test.region
  }
}

resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}
`, rName, bucketName))
}

func testAccResourceDataSyncConfig_destinationDataSharingBucketPolicy() string {
	return `
resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.bucket

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "SSMBucketPermissionsCheck",
      "Effect": "Allow",
      "Principal": {
        "Service": "ssm.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "s3:GetBucketAcl",
      "Resource": "${aws_s3_bucket.test.arn}"
    },
    {
      "Sid": " SSMBucketDelivery",
      "Effect": "Allow",
      "Principal": {
        "Service": "ssm.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "s3:PutObject",
      "Resource": [
        "${aws_s3_bucket.test.arn}/*/accountid=*/*"
      ],
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "bucket-owner-full-control"
        }
      }
    },
    {
      "Sid": " SSMBucketDeliveryTagging",
      "Effect": "Allow",
      "Principal": {
        "Service": "ssm.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "s3:PutObjectTagging",
      "Resource": [
        "${aws_s3_bucket.test.arn}/*/accountid=*/*"
      ]
     }
  ]
}
      EOF
}

data "aws_partition" "current" {}
`
}
