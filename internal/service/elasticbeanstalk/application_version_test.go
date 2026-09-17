// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elasticbeanstalk_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfelasticbeanstalk "github.com/hashicorp/terraform-provider-aws/internal/service/elasticbeanstalk"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElasticBeanstalkApplicationVersion_BeanstalkApp_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var appVersion awstypes.ApplicationVersionDescription

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationVersionConfig_basic(acctest.RandInt(t)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationVersionExists(ctx, t, "aws_elastic_beanstalk_application_version.default", &appVersion),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkApplicationVersion_BeanstalkApp_duplicateLabels(t *testing.T) {
	ctx := acctest.Context(t)
	var firstAppVersion awstypes.ApplicationVersionDescription
	var secondAppVersion awstypes.ApplicationVersionDescription

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationVersionConfig_duplicateLabel(acctest.RandInt(t)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationVersionExists(ctx, t, "aws_elastic_beanstalk_application_version.first", &firstAppVersion),
					testAccCheckApplicationVersionExists(ctx, t, "aws_elastic_beanstalk_application_version.second", &secondAppVersion),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkApplicationVersion_BeanstalkApp_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var appVersion awstypes.ApplicationVersionDescription
	resourceName := "aws_elastic_beanstalk_application_version.default"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationVersionConfig_tags(acctest.RandInt(t), "test1", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationVersionExists(ctx, t, resourceName, &appVersion),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.firstTag", "test1"),
					resource.TestCheckResourceAttr(resourceName, "tags.secondTag", "test2"),
				),
			},
			{
				Config: testAccApplicationVersionConfig_tags(acctest.RandInt(t), "updateTest1", "updateTest2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationVersionExists(ctx, t, resourceName, &appVersion),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.firstTag", "updateTest1"),
					resource.TestCheckResourceAttr(resourceName, "tags.secondTag", "updateTest2"),
				),
			},
			{
				Config: testAccApplicationVersionConfig_addTags(acctest.RandInt(t), "updateTest1", "updateTest2", "addTest3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationVersionExists(ctx, t, resourceName, &appVersion),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.firstTag", "updateTest1"),
					resource.TestCheckResourceAttr(resourceName, "tags.secondTag", "updateTest2"),
					resource.TestCheckResourceAttr(resourceName, "tags.thirdTag", "addTest3"),
				),
			},
			{
				Config: testAccApplicationVersionConfig_tags(acctest.RandInt(t), "updateTest1", "updateTest2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationVersionExists(ctx, t, resourceName, &appVersion),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.firstTag", "updateTest1"),
					resource.TestCheckResourceAttr(resourceName, "tags.secondTag", "updateTest2"),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkApplicationVersion_BeanstalkApp_process(t *testing.T) {
	ctx := acctest.Context(t)
	var appVersion awstypes.ApplicationVersionDescription
	resourceName := "aws_elastic_beanstalk_application_version.default"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationVersionConfig_process(acctest.RandInt(t), acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationVersionExists(ctx, t, resourceName, &appVersion),
					testAccCheckApplicationVersionMatchStatus(&appVersion, awstypes.ApplicationVersionStatusProcessed),
				),
			},
			{
				Config: testAccApplicationVersionConfig_process(acctest.RandInt(t), acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationVersionExists(ctx, t, resourceName, &appVersion),
					testAccCheckApplicationVersionMatchStatus(&appVersion, awstypes.ApplicationVersionStatusUnprocessed),
				),
			},
		},
	})
}

func testAccCheckApplicationVersionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ElasticBeanstalkClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_elastic_beanstalk_application_version" {
				continue
			}

			_, err := tfelasticbeanstalk.FindApplicationVersionByTwoPartKey(ctx, conn, rs.Primary.Attributes["application"], rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Elastic Beanstalk Application Version %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckApplicationVersionExists(ctx context.Context, t *testing.T, n string, v *awstypes.ApplicationVersionDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ElasticBeanstalkClient(ctx)

		output, err := tfelasticbeanstalk.FindApplicationVersionByTwoPartKey(ctx, conn, rs.Primary.Attributes["application"], rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckApplicationVersionMatchStatus(v *awstypes.ApplicationVersionDescription, status awstypes.ApplicationVersionStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !strings.EqualFold(string(v.Status), string(status)) {
			return fmt.Errorf("Elastic Beanstalk Application Version status %s does not match to expected status %s", v.Status, status)
		}

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
  bucket      = aws_s3_object.default.bucket
  key         = aws_s3_object.default.key
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
  bucket      = aws_s3_object.default.bucket
  key         = aws_s3_object.default.key
}

resource "aws_elastic_beanstalk_application" "second" {
  name        = "tf-test-name-%d-second"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_application_version" "second" {
  application = aws_elastic_beanstalk_application.second.name
  name        = "tf-test-version-label-%d"
  bucket      = aws_s3_object.default.bucket
  key         = aws_s3_object.default.key
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
  bucket      = aws_s3_object.default.bucket
  key         = aws_s3_object.default.key

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
  bucket      = aws_s3_object.default.bucket
  key         = aws_s3_object.default.key

  tags = {
    firstTag  = "%[2]s"
    secondTag = "%[3]s"
    thirdTag  = "%[4]s"
  }
}
`, randInt, tag1, tag2, tag3)
}

func testAccApplicationVersionConfig_process(randInt int, process string) string {
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
  bucket      = aws_s3_object.default.bucket
  key         = aws_s3_object.default.key
  process     = %s
}
`, randInt, randInt, randInt, process)
}
