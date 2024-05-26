// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticbeanstalk_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElasticBeanstalkApplicationVersion_BeanstalkApp_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var appVersion awstypes.ApplicationVersionDescription

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationVersionConfig_basic(sdkacctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationVersionExists(ctx, "aws_elastic_beanstalk_application_version.default", &appVersion),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkApplicationVersion_BeanstalkApp_duplicateLabels(t *testing.T) {
	ctx := acctest.Context(t)
	var firstAppVersion awstypes.ApplicationVersionDescription
	var secondAppVersion awstypes.ApplicationVersionDescription

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationVersionConfig_duplicateLabel(sdkacctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationVersionExists(ctx, "aws_elastic_beanstalk_application_version.first", &firstAppVersion),
					testAccCheckApplicationVersionExists(ctx, "aws_elastic_beanstalk_application_version.second", &secondAppVersion),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkApplicationVersion_BeanstalkApp_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var appVersion awstypes.ApplicationVersionDescription
	resourceName := "aws_elastic_beanstalk_application_version.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationVersionConfig_tags(sdkacctest.RandInt(), "test1", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationVersionExists(ctx, resourceName, &appVersion),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.firstTag", "test1"),
					resource.TestCheckResourceAttr(resourceName, "tags.secondTag", "test2"),
				),
			},
			{
				Config: testAccApplicationVersionConfig_tags(sdkacctest.RandInt(), "updateTest1", "updateTest2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationVersionExists(ctx, resourceName, &appVersion),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.firstTag", "updateTest1"),
					resource.TestCheckResourceAttr(resourceName, "tags.secondTag", "updateTest2"),
				),
			},
			{
				Config: testAccApplicationVersionConfig_addTags(sdkacctest.RandInt(), "updateTest1", "updateTest2", "addTest3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationVersionExists(ctx, resourceName, &appVersion),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.firstTag", "updateTest1"),
					resource.TestCheckResourceAttr(resourceName, "tags.secondTag", "updateTest2"),
					resource.TestCheckResourceAttr(resourceName, "tags.thirdTag", "addTest3"),
				),
			},
			{
				Config: testAccApplicationVersionConfig_tags(sdkacctest.RandInt(), "updateTest1", "updateTest2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationVersionExists(ctx, resourceName, &appVersion),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.firstTag", "updateTest1"),
					resource.TestCheckResourceAttr(resourceName, "tags.secondTag", "updateTest2"),
				),
			},
		},
	})
}

func testAccCheckApplicationVersionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticBeanstalkClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_elastic_beanstalk_application_version" {
				continue
			}

			describeApplicationVersionOpts := &elasticbeanstalk.DescribeApplicationVersionsInput{
				ApplicationName: aws.String(rs.Primary.Attributes["application"]),
				VersionLabels:   []string{rs.Primary.ID},
			}
			resp, err := conn.DescribeApplicationVersions(ctx, describeApplicationVersionOpts)
			if err == nil {
				if len(resp.ApplicationVersions) > 0 {
					return fmt.Errorf("Elastic Beanstalk Application Verson still exists.")
				}

				return nil
			}
			if !tfawserr.ErrCodeEquals(err, "InvalidParameterValue") {
				return err
			}
		}

		return nil
	}
}

func testAccCheckApplicationVersionExists(ctx context.Context, n string, app *awstypes.ApplicationVersionDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Elastic Beanstalk Application Version is not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticBeanstalkClient(ctx)
		describeApplicationVersionOpts := &elasticbeanstalk.DescribeApplicationVersionsInput{
			ApplicationName: aws.String(rs.Primary.Attributes["application"]),
			VersionLabels:   []string{rs.Primary.ID},
		}

		log.Printf("[DEBUG] Elastic Beanstalk Application Version TEST describe opts: %v", describeApplicationVersionOpts)

		resp, err := conn.DescribeApplicationVersions(ctx, describeApplicationVersionOpts)
		if err != nil {
			return err
		}
		if len(resp.ApplicationVersions) == 0 {
			return fmt.Errorf("Elastic Beanstalk Application Version not found.")
		}

		*app = resp.ApplicationVersions[0]

		return nil
	}
}

func testAccApplicationVersionConfig_basic(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "default" {
  bucket = "tftest.applicationversion.bucket-%d"
}

resource "aws_s3_object" "default" {
  bucket = aws_s3_bucket.default.id
  key    = "beanstalk/python-v1.zip"
  source = "test-fixtures/python-v1.zip"
}

resource "aws_elastic_beanstalk_application" "default" {
  name        = "tf-test-name-%d"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_application_version" "default" {
  application = aws_elastic_beanstalk_application.default.name
  name        = "tf-test-version-label-%d"
  bucket      = aws_s3_bucket.default.id
  key         = aws_s3_object.default.id
}
`, randInt, randInt, randInt)
}

func testAccApplicationVersionConfig_duplicateLabel(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "default" {
  bucket = "tftest.applicationversion.bucket-%d"
}

resource "aws_s3_object" "default" {
  bucket = aws_s3_bucket.default.id
  key    = "beanstalk/python-v1.zip"
  source = "test-fixtures/python-v1.zip"
}

resource "aws_elastic_beanstalk_application" "first" {
  name        = "tf-test-name-%d-first"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_application_version" "first" {
  application = aws_elastic_beanstalk_application.first.name
  name        = "tf-test-version-label-%d"
  bucket      = aws_s3_bucket.default.id
  key         = aws_s3_object.default.id
}

resource "aws_elastic_beanstalk_application" "second" {
  name        = "tf-test-name-%d-second"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_application_version" "second" {
  application = aws_elastic_beanstalk_application.second.name
  name        = "tf-test-version-label-%d"
  bucket      = aws_s3_bucket.default.id
  key         = aws_s3_object.default.id
}
`, randInt, randInt, randInt, randInt, randInt)
}

func testAccApplicationVersionConfig_tags(randInt int, tag1, tag2 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "default" {
  bucket = "tftest.applicationversion.bucket-%[1]d"
}

resource "aws_s3_object" "default" {
  bucket = aws_s3_bucket.default.id
  key    = "beanstalk/python-v1.zip"
  source = "test-fixtures/python-v1.zip"
}

resource "aws_elastic_beanstalk_application" "default" {
  name        = "tf-test-name-%[1]d"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_application_version" "default" {
  application = aws_elastic_beanstalk_application.default.name
  name        = "tf-test-version-label-%[1]d"
  bucket      = aws_s3_bucket.default.id
  key         = aws_s3_object.default.id

  tags = {
    firstTag  = "%[2]s"
    secondTag = "%[3]s"
  }
}
`, randInt, tag1, tag2)
}

func testAccApplicationVersionConfig_addTags(randInt int, tag1, tag2, tag3 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "default" {
  bucket = "tftest.applicationversion.bucket-%[1]d"
}

resource "aws_s3_object" "default" {
  bucket = aws_s3_bucket.default.id
  key    = "beanstalk/python-v1.zip"
  source = "test-fixtures/python-v1.zip"
}

resource "aws_elastic_beanstalk_application" "default" {
  name        = "tf-test-name-%[1]d"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_application_version" "default" {
  application = aws_elastic_beanstalk_application.default.name
  name        = "tf-test-version-label-%[1]d"
  bucket      = aws_s3_bucket.default.id
  key         = aws_s3_object.default.id

  tags = {
    firstTag  = "%[2]s"
    secondTag = "%[3]s"
    thirdTag  = "%[4]s"
  }
}
`, randInt, tag1, tag2, tag3)
}
