// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/s3control/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3control "github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3ControlBucketLifecycleConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_ruleID(rName, "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3control_bucket.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#":      acctest.Ct1,
						"expiration.0.days": "365",
						names.AttrID:        "test",
						names.AttrStatus:    string(types.ExpirationStatusEnabled),
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

func TestAccS3ControlBucketLifecycleConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_ruleID(rName, "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3control.ResourceBucketLifecycleConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3ControlBucketLifecycleConfiguration_RuleAbortIncompleteMultipartUpload_daysAfterInitiation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_ruleAbortIncompleteMultipartUploadDaysAfterInitiation(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"abort_incomplete_multipart_upload.#":                       acctest.Ct1,
						"abort_incomplete_multipart_upload.0.days_after_initiation": acctest.Ct1,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_ruleAbortIncompleteMultipartUploadDaysAfterInitiation(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"abort_incomplete_multipart_upload.#":                       acctest.Ct1,
						"abort_incomplete_multipart_upload.0.days_after_initiation": acctest.Ct2,
					}),
				),
			},
		},
	})
}

func TestAccS3ControlBucketLifecycleConfiguration_RuleExpiration_date(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_bucket_lifecycle_configuration.test"
	date1 := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	date2 := time.Now().AddDate(0, 0, 2).Format("2006-01-02")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_ruleExpirationDate(rName, date1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#":      acctest.Ct1,
						"expiration.0.date": date1,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_ruleExpirationDate(rName, date2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#":      acctest.Ct1,
						"expiration.0.date": date2,
					}),
				),
			},
		},
	})
}

func TestAccS3ControlBucketLifecycleConfiguration_RuleExpiration_days(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_ruleExpirationDays(rName, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#":      acctest.Ct1,
						"expiration.0.days": "7",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_ruleExpirationDays(rName, 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#":      acctest.Ct1,
						"expiration.0.days": "30",
					}),
				),
			},
		},
	})
}

func TestAccS3ControlBucketLifecycleConfiguration_RuleExpiration_expiredObjectDeleteMarker(t *testing.T) {
	acctest.Skip(t, "S3 on Outposts does not error or save it in the API when receiving this parameter")
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_ruleExpirationExpiredObjectDeleteMarker(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#": acctest.Ct1,
						"expiration.0.expired_object_delete_marker": acctest.CtTrue,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_ruleExpirationExpiredObjectDeleteMarker(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#": acctest.Ct1,
						"expiration.0.expired_object_delete_marker": acctest.CtFalse,
					}),
				),
			},
		},
	})
}

func TestAccS3ControlBucketLifecycleConfiguration_RuleFilter_prefix(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_ruleFilterPrefix(rName, "test1/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"filter.#":        acctest.Ct1,
						"filter.0.prefix": "test1/",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_ruleFilterPrefix(rName, "test2/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"filter.#":        acctest.Ct1,
						"filter.0.prefix": "test2/",
					}),
				),
			},
		},
	})
}

func TestAccS3ControlBucketLifecycleConfiguration_RuleFilter_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_ruleFilterTags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"filter.#":           acctest.Ct1,
						"filter.0.tags.%":    acctest.Ct1,
						"filter.0.tags.key1": acctest.CtValue1,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// There is currently an API model or AWS Go SDK bug where LifecycleFilter.And.Tags
			// does not get populated from the XML response. Reference:
			// https://github.com/aws/aws-sdk-go/issues/3591
			// {
			// 	Config: testAccBucketLifecycleConfigurationConfig_ruleFilterTags2(rName, "key1", "value1updated", "key2", "value2"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		testAccCheckBucketLifecycleConfigurationExists(resourceName),
			// 		resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
			// 		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
			// 			"filter.#":           "1",
			// 			"filter.0.tags.%":    "2",
			// 			"filter.0.tags.key1": "value1updated",
			// 			"filter.0.tags.key2": "value2",
			// 		}),
			// 	),
			// },
			{
				Config: testAccBucketLifecycleConfigurationConfig_ruleFilterTags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"filter.#":           acctest.Ct1,
						"filter.0.tags.%":    acctest.Ct1,
						"filter.0.tags.key2": acctest.CtValue2,
					}),
				),
			},
		},
	})
}

func TestAccS3ControlBucketLifecycleConfiguration_Rule_id(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_ruleID(rName, "test1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID: "test1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_ruleID(rName, "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrID: "test2",
					}),
				),
			},
		},
	})
}

func TestAccS3ControlBucketLifecycleConfiguration_Rule_status(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLifecycleConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationConfig_ruleStatus(rName, string(types.ExpirationStatusDisabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrStatus: string(types.ExpirationStatusDisabled),
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfigurationConfig_ruleStatus(rName, string(types.ExpirationStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrStatus: string(types.ExpirationStatusEnabled),
					}),
				),
			},
		},
	})
}

func testAccCheckBucketLifecycleConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3control_bucket_lifecycle_configuration" {
				continue
			}

			parsedArn, err := arn.Parse(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfs3control.FindBucketLifecycleConfigurationByTwoPartKey(ctx, conn, parsedArn.AccountID, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Control Bucket Lifecycle Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketLifecycleConfigurationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

		parsedArn, err := arn.Parse(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = tfs3control.FindBucketLifecycleConfigurationByTwoPartKey(ctx, conn, parsedArn.AccountID, rs.Primary.ID)

		return err
	}
}

func testAccBucketLifecycleConfigurationConfig_ruleAbortIncompleteMultipartUploadDaysAfterInitiation(rName string, daysAfterInitiation int) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_s3control_bucket" "test" {
  bucket     = %[1]q
  outpost_id = data.aws_outposts_outpost.test.id
}

resource "aws_s3control_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3control_bucket.test.arn

  rule {
    abort_incomplete_multipart_upload {
      days_after_initiation = %[2]d
    }

    expiration {
      days = 365
    }

    id = "test"
  }
}
`, rName, daysAfterInitiation)
}

func testAccBucketLifecycleConfigurationConfig_ruleExpirationDate(rName string, date string) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_s3control_bucket" "test" {
  bucket     = %[1]q
  outpost_id = data.aws_outposts_outpost.test.id
}

resource "aws_s3control_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3control_bucket.test.arn

  rule {
    expiration {
      date = %[2]q
    }

    id = "test"
  }
}
`, rName, date)
}

func testAccBucketLifecycleConfigurationConfig_ruleExpirationDays(rName string, days int) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_s3control_bucket" "test" {
  bucket     = %[1]q
  outpost_id = data.aws_outposts_outpost.test.id
}

resource "aws_s3control_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3control_bucket.test.arn

  rule {
    expiration {
      days = %[2]d
    }

    id = "test"
  }
}
`, rName, days)
}

func testAccBucketLifecycleConfigurationConfig_ruleExpirationExpiredObjectDeleteMarker(rName string, expiredObjectDeleteMarker bool) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_s3control_bucket" "test" {
  bucket     = %[1]q
  outpost_id = data.aws_outposts_outpost.test.id
}

resource "aws_s3control_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3control_bucket.test.arn

  rule {
    expiration {
      days                         = %[2]t ? null : 365
      expired_object_delete_marker = %[2]t
    }

    id = "test"
  }
}
`, rName, expiredObjectDeleteMarker)
}

func testAccBucketLifecycleConfigurationConfig_ruleFilterPrefix(rName, prefix string) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_s3control_bucket" "test" {
  bucket     = %[1]q
  outpost_id = data.aws_outposts_outpost.test.id
}

resource "aws_s3control_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3control_bucket.test.arn

  rule {
    expiration {
      days = 365
    }

    filter {
      prefix = %[2]q
    }

    id = "test"
  }
}
`, rName, prefix)
}

func testAccBucketLifecycleConfigurationConfig_ruleFilterTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_s3control_bucket" "test" {
  bucket     = %[1]q
  outpost_id = data.aws_outposts_outpost.test.id
}

resource "aws_s3control_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3control_bucket.test.arn

  rule {
    expiration {
      days = 365
    }

    filter {
      tags = {
        %[2]q = %[3]q
      }
    }

    id = "test"
  }
}
`, rName, tagKey1, tagValue1)
}

// See TestAccS3ControlBucketLifecycleConfiguration_RuleFilter_tags note about XML handling bug.
// func testAccBucketLifecycleConfigurationConfig_ruleFilterTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
// 	return fmt.Sprintf(`
// data "aws_outposts_outposts" "test" {}

// data "aws_outposts_outpost" "test" {
//   id = tolist(data.aws_outposts_outposts.test.ids)[0]
// }

// resource "aws_s3control_bucket" "test" {
//   bucket     = %[1]q
//   outpost_id = data.aws_outposts_outpost.test.id
// }

// resource "aws_s3control_bucket_lifecycle_configuration" "test" {
//   bucket = aws_s3control_bucket.test.arn

//   rule {
//     expiration {
//       days = 365
//     }

//     filter {
//       tags = {
//         %[2]q = %[3]q
//         %[4]q = %[5]q
//       }
//     }

//     id = "test"
//   }
// }
// `, rName, tagKey1, tagValue1, tagKey2, tagValue2)
// }

func testAccBucketLifecycleConfigurationConfig_ruleID(rName, id string) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_s3control_bucket" "test" {
  bucket     = %[1]q
  outpost_id = data.aws_outposts_outpost.test.id
}

resource "aws_s3control_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3control_bucket.test.arn

  rule {
    expiration {
      days = 365
    }

    id = %[2]q
  }
}
`, rName, id)
}

func testAccBucketLifecycleConfigurationConfig_ruleStatus(rName, status string) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_s3control_bucket" "test" {
  bucket     = %[1]q
  outpost_id = data.aws_outposts_outpost.test.id
}

resource "aws_s3control_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3control_bucket.test.arn

  rule {
    expiration {
      days = 365
    }

    id     = "test"
    status = %[2]q
  }
}
`, rName, status)
}
