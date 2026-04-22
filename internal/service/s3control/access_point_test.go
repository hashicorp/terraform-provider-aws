// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfs3control "github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3ControlAccessPoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3control.GetAccessPointOutput
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	accessPointName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_access_point.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointConfig_basic(bucketName, accessPointName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPointExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
					// https://docs.aws.amazon.com/AmazonS3/latest/userguide/access-points-alias.html:
					resource.TestMatchResourceAttr(resourceName, names.AttrAlias, regexache.MustCompile(`^.*-s3alias$`)),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "s3", fmt.Sprintf("accesspoint/%s", accessPointName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucket, bucketName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "bucket_account_id"),
					acctest.MatchResourceAttrRegionalHostname(resourceName, names.AttrDomainName, "s3-accesspoint", regexache.MustCompile(fmt.Sprintf("^%s-\\d{12}", accessPointName))),
					resource.TestCheckResourceAttr(resourceName, "endpoints.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "has_public_access_policy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, accessPointName),
					resource.TestCheckResourceAttr(resourceName, "network_origin", "Internet"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPolicy, ""),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_policy", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.ignore_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.restrict_public_buckets", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "0"),
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

func TestAccS3ControlAccessPoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3control.GetAccessPointOutput
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	accessPointName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_access_point.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointConfig_basic(bucketName, accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfs3control.ResourceAccessPoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccS3ControlAccessPoint_Bucket_arn(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3control.GetAccessPointOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_access_point.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointConfig_bucketARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "s3-outposts", fmt.Sprintf("outpost/[^/]+/accesspoint/%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3control_bucket.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalHostname(resourceName, names.AttrDomainName, "s3-accesspoint", regexache.MustCompile(fmt.Sprintf("^%s-\\d{12}", rName))),
					resource.TestCheckResourceAttr(resourceName, "has_public_access_policy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "network_origin", "Vpc"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPolicy, ""),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_policy", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.ignore_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.restrict_public_buckets", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_configuration.0.vpc_id", "aws_vpc.test", names.AttrID),
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

func TestAccS3ControlAccessPoint_policy(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3control.GetAccessPointOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_access_point.test"

	expectedPolicyText1 := func() string {
		return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:%[1]s:iam::%[3]s:root"
      },
      "Action": "s3:GetObjectTagging",
      "Resource": [
        "arn:%[1]s:s3:%[2]s:%[3]s:accesspoint/%[4]s/object/*"
      ]
    }
  ]
}`, acctest.Partition(), acctest.Region(), acctest.AccountID(ctx), rName)
	}
	expectedPolicyText2 := func() string {
		return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:%[1]s:iam::%[3]s:root"
      },
      "Action": [
        "s3:GetObjectLegalHold",
        "s3:GetObjectRetention"
      ],
      "Resource": [
        "arn:%[1]s:s3:%[2]s:%[3]s:accesspoint/%[4]s/object/*"
      ]
    }
  ]
}`, acctest.Partition(), acctest.Region(), acctest.AccountID(ctx), rName)
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(ctx, t, resourceName, &v),
					testAccCheckAccessPointHasPolicy(ctx, t, resourceName, expectedPolicyText1),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "s3", fmt.Sprintf("accesspoint/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucket, rName),
					resource.TestCheckResourceAttr(resourceName, "has_public_access_policy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "network_origin", "Internet"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_policy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.ignore_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.restrict_public_buckets", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccessPointConfig_policyUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(ctx, t, resourceName, &v),
					testAccCheckAccessPointHasPolicy(ctx, t, resourceName, expectedPolicyText2),
				),
			},
			{
				Config: testAccAccessPointConfig_noPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrPolicy, ""),
				),
			},
		},
	})
}

func TestAccS3ControlAccessPoint_publicAccessBlock(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3control.GetAccessPointOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_access_point.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointConfig_publicBlock(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "s3", fmt.Sprintf("accesspoint/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucket, rName),
					resource.TestCheckResourceAttr(resourceName, "has_public_access_policy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "network_origin", "Internet"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPolicy, ""),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_acls", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_policy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.ignore_public_acls", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.restrict_public_buckets", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "0"),
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

func TestAccS3ControlAccessPoint_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3control.GetAccessPointOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_access_point.test"
	vpcResourceName := "aws_vpc.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointConfig_vpc(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "s3", fmt.Sprintf("accesspoint/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucket, rName),
					resource.TestCheckResourceAttr(resourceName, "has_public_access_policy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "network_origin", "VPC"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPolicy, ""),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_policy", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.ignore_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.restrict_public_buckets", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_configuration.0.vpc_id", vpcResourceName, names.AttrID),
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

func TestAccS3ControlAccessPoint_DirectoryBucket_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3control.GetAccessPointOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_access_point.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointConfig_directoryBucketBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
					// https://docs.aws.amazon.com/AmazonS3/latest/userguide/access-points-alias.html:
					resource.TestMatchResourceAttr(resourceName, names.AttrAlias, regexache.MustCompile(`^.*--xa-s3$`)),
					resource.TestMatchResourceAttr(resourceName, names.AttrBucket, regexache.MustCompile(`^.*--x-s3$`)),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "bucket_account_id"),
					resource.TestCheckResourceAttr(resourceName, "has_public_access_policy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "network_origin", "Internet"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPolicy, ""),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_policy", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.ignore_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.restrict_public_buckets", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "0"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("s3express", regexache.MustCompile(`accesspoint/.+--xa-s3`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringRegexp(tfs3control.AccessPointForDirectoryBucketNameRegex)),
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

func TestAccS3ControlAccessPoint_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3control.GetAccessPointOutput
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	accessPointName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_access_point.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointConfig_tags1(bucketName, accessPointName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccessPointConfig_tags2(bucketName, accessPointName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAccessPointConfig_tags1(bucketName, accessPointName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccS3ControlAccessPoint_DirectoryBucket_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3control.GetAccessPointOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_access_point.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointConfig_directoryBucketTags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccessPointConfig_directoryBucketTags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccAccessPointConfig_directoryBucketTags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func testAccCheckAccessPointDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).S3ControlClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_access_point" {
				continue
			}

			accountID, name, err := tfs3control.AccessPointParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfs3control.FindAccessPointByTwoPartKey(ctx, conn, accountID, name)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Access Point %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAccessPointExists(ctx context.Context, t *testing.T, n string, v *s3control.GetAccessPointOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		accountID, name, err := tfs3control.AccessPointParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.ProviderMeta(ctx, t).S3ControlClient(ctx)

		output, err := tfs3control.FindAccessPointByTwoPartKey(ctx, conn, accountID, name)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAccessPointHasPolicy(ctx context.Context, t *testing.T, n string, fn func() string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		accountID, name, err := tfs3control.AccessPointParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.ProviderMeta(ctx, t).S3ControlClient(ctx)

		actualPolicyText, _, err := tfs3control.FindAccessPointPolicyAndStatusByTwoPartKey(ctx, conn, accountID, name)

		if err != nil {
			return err
		}

		expectedPolicyText := fn()

		equivalent, err := awspolicy.PoliciesAreEquivalent(actualPolicyText, expectedPolicyText)
		if err != nil {
			return fmt.Errorf("Error testing policy equivalence: %w", err)
		}
		if !equivalent {
			return fmt.Errorf("Non-equivalent policy error:\n\nexpected: %s\n\n     got: %s\n",
				expectedPolicyText, actualPolicyText)
		}

		return nil
	}
}

func testAccAccessPointConfig_basic(bucketName, accessPointName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_access_point" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[2]q
}
`, bucketName, accessPointName)
}

func testAccAccessPointConfig_bucketARN(rName string) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_s3control_bucket" "test" {
  bucket     = %[1]q
  outpost_id = data.aws_outposts_outpost.test.id
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_s3_access_point" "test" {
  bucket = aws_s3control_bucket.test.arn
  name   = %[1]q

  vpc_configuration {
    vpc_id = aws_vpc.test.id
  }
}
`, rName)
}

func testAccAccessPointConfig_policy(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_access_point" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q
  policy = data.aws_iam_policy_document.test.json

  public_access_block_configuration {
    block_public_acls       = true
    block_public_policy     = false
    ignore_public_acls      = true
    restrict_public_buckets = false
  }
}

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    effect = "Allow"

    actions = [
      "s3:GetObjectTagging",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:accesspoint/%[1]s/object/*",
    ]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }
}
`, rName)
}

func testAccAccessPointConfig_policyUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_access_point" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q
  policy = data.aws_iam_policy_document.test.json

  public_access_block_configuration {
    block_public_acls       = true
    block_public_policy     = false
    ignore_public_acls      = true
    restrict_public_buckets = false
  }
}

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    effect = "Allow"

    actions = [
      "s3:GetObjectLegalHold",
      "s3:GetObjectRetention"
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:accesspoint/%[1]s/object/*",
    ]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }
}
`, rName)
}

func testAccAccessPointConfig_noPolicy(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_access_point" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  public_access_block_configuration {
    block_public_acls       = true
    block_public_policy     = false
    ignore_public_acls      = true
    restrict_public_buckets = false
  }

  policy = "{}"
}
`, rName)
}

func testAccAccessPointConfig_publicBlock(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_access_point" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  public_access_block_configuration {
    block_public_acls       = false
    block_public_policy     = false
    ignore_public_acls      = false
    restrict_public_buckets = false
  }
}
`, rName)
}

func testAccAccessPointConfig_vpc(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_access_point" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  vpc_configuration {
    vpc_id = aws_vpc.test.id
  }
}
`, rName)
}

func testAccAccessPointConfig_baseDirectoryBucket(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInExclude("use2-az2", "use1-az1", "use1-az2", "use1-az3", "usw2-az2", "aps1-az3", "apne1-az2", "euw1-az2"),
		fmt.Sprintf(`
locals {
  location_name = data.aws_availability_zones.available.zone_ids[0]
  bucket        = "%[1]s--${local.location_name}--x-s3"
}

resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket
  location {
    name = local.location_name
  }

  force_destroy = true
}
`, rName))
}

func testAccAccessPointConfig_directoryBucketBasic(rName string) string {
	return acctest.ConfigCompose(testAccAccessPointConfig_baseDirectoryBucket(rName), fmt.Sprintf(`
resource "aws_s3_access_point" "test" {
  bucket = aws_s3_directory_bucket.test.bucket
  name   = "%[1]s--${local.location_name}--xa-s3"
}
`, rName))
}

func testAccAccessPointConfig_tags1(bucketName, accessPointName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_access_point" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[2]q

  tags = {
    %[3]q = %[4]q
  }
}
`, bucketName, accessPointName, tagKey1, tagValue1)
}

func testAccAccessPointConfig_tags2(bucketName, accessPointName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_access_point" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[2]q

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, bucketName, accessPointName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAccessPointConfig_directoryBucketTags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccAccessPointConfig_baseDirectoryBucket(rName), fmt.Sprintf(`
resource "aws_s3_access_point" "test" {
  bucket = aws_s3_directory_bucket.test.bucket
  name   = "%[1]s--${local.location_name}--xa-s3"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccAccessPointConfig_directoryBucketTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccAccessPointConfig_baseDirectoryBucket(rName), fmt.Sprintf(`
resource "aws_s3_access_point" "test" {
  bucket = aws_s3_directory_bucket.test.bucket
  name   = "%[1]s--${local.location_name}--xa-s3"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
