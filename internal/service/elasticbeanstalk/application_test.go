// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticbeanstalk_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelasticbeanstalk "github.com/hashicorp/terraform-provider-aws/internal/service/elasticbeanstalk"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElasticBeanstalkApplication_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.ApplicationDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "appversion_lifecycle.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccElasticBeanstalkApplication_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.ApplicationDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &app),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelasticbeanstalk.ResourceApplication(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccElasticBeanstalkApplication_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.ApplicationDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccApplicationConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccApplicationConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkApplication_description(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.ApplicationDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_description(rName, "description 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "appversion_lifecycle.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description 1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccApplicationConfig_description(rName, "description 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "appversion_lifecycle.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description 2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkApplication_appVersionLifecycle(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.ApplicationDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_maxAge(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "appversion_lifecycle.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "appversion_lifecycle.0.service_role", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "appversion_lifecycle.0.max_age_in_days", "90"),
					resource.TestCheckResourceAttr(resourceName, "appversion_lifecycle.0.max_count", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "appversion_lifecycle.0.delete_source_from_s3", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccApplicationConfig_maxCount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "appversion_lifecycle.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "appversion_lifecycle.0.service_role", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "appversion_lifecycle.0.max_age_in_days", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "appversion_lifecycle.0.max_count", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "appversion_lifecycle.0.delete_source_from_s3", acctest.CtFalse),
				),
			},
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "appversion_lifecycle.#", acctest.Ct0),
				),
			},
			{
				Config: testAccApplicationConfig_maxAge(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "appversion_lifecycle.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "appversion_lifecycle.0.service_role", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "appversion_lifecycle.0.max_age_in_days", "90"),
					resource.TestCheckResourceAttr(resourceName, "appversion_lifecycle.0.max_count", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "appversion_lifecycle.0.delete_source_from_s3", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccCheckApplicationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticBeanstalkClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_elastic_beanstalk_application" {
				continue
			}

			_, err := tfelasticbeanstalk.FindApplicationByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Elastic Beanstalk Application %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckApplicationExists(ctx context.Context, n string, v *awstypes.ApplicationDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Elastic Beanstalk Application ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticBeanstalkClient(ctx)

		output, err := tfelasticbeanstalk.FindApplicationByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccApplicationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "test" {
  name = %[1]q
}
`, rName)
}

func testAccApplicationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccApplicationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccApplicationConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "test" {
  name        = %[1]q
  description = %[2]q
}
`, rName, description)
}

func testAccApplicationConfig_baseServiceRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticbeanstalk.amazonaws.com"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "elasticbeanstalk"
        }
      }
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Sid": "AllowOperations",
    "Effect": "Allow",
    "Action": ["iam:PassRole"],
    "Resource": ["*"]
  }]
}
EOF
}
`, rName)
}

func testAccApplicationConfig_maxAge(rName string) string {
	return acctest.ConfigCompose(testAccApplicationConfig_baseServiceRole(rName), fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "test" {
  name = %[1]q

  appversion_lifecycle {
    service_role          = aws_iam_role.test.arn
    max_age_in_days       = 90
    delete_source_from_s3 = true
  }
}
`, rName))
}

func testAccApplicationConfig_maxCount(rName string) string {
	return acctest.ConfigCompose(testAccApplicationConfig_baseServiceRole(rName), fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "test" {
  name = %[1]q

  appversion_lifecycle {
    service_role          = aws_iam_role.test.arn
    max_count             = 10
    delete_source_from_s3 = false
  }
}
`, rName))
}
