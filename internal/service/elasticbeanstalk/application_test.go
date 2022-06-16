package elasticbeanstalk_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccElasticBeanstalkApplication_BeanstalkApp_basic(t *testing.T) {
	var app elasticbeanstalk.ApplicationDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_application.tftest"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticbeanstalk.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName, &app),
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

func TestAccElasticBeanstalkApplication_BeanstalkApp_appVersionLifecycle(t *testing.T) {
	var app elasticbeanstalk.ApplicationDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticbeanstalk.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists("aws_elastic_beanstalk_application.tftest", &app),
					resource.TestCheckNoResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.service_role"),
					resource.TestCheckNoResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.max_age_in_days"),
					resource.TestCheckNoResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.max_count"),
					resource.TestCheckNoResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.delete_source_from_s3"),
				),
			},
			{
				Config: testAccApplicationConfig_maxAge(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists("aws_elastic_beanstalk_application.tftest", &app),
					resource.TestCheckResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.#", "1"),
					resource.TestCheckResourceAttrPair(
						"aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.service_role",
						"aws_iam_role.beanstalk_service", "arn"),
					resource.TestCheckResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.max_age_in_days", "90"),
					resource.TestCheckResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.max_count", "0"),
					resource.TestCheckResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.delete_source_from_s3", "true"),
				),
			},
			{
				Config: testAccApplicationConfig_maxCount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists("aws_elastic_beanstalk_application.tftest", &app),
					resource.TestCheckResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.#", "1"),
					resource.TestCheckResourceAttrPair(
						"aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.service_role",
						"aws_iam_role.beanstalk_service", "arn"),
					resource.TestCheckResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.max_age_in_days", "0"),
					resource.TestCheckResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.max_count", "10"),
					resource.TestCheckResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.delete_source_from_s3", "false"),
				),
			},
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists("aws_elastic_beanstalk_application.tftest", &app),
					resource.TestCheckNoResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.service_role"),
					resource.TestCheckNoResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.max_age_in_days"),
					resource.TestCheckNoResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.max_count"),
					resource.TestCheckNoResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.delete_source_from_s3"),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkApplication_BeanstalkApp_tags(t *testing.T) {
	var app elasticbeanstalk.ApplicationDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_application.tftest"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticbeanstalk.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_tags2(rName, "test1", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.firstTag", "test1"),
					resource.TestCheckResourceAttr(resourceName, "tags.secondTag", "test2"),
				),
			},
			{
				Config: testAccApplicationConfig_tags2(rName, "updateTest1", "updateTest2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.firstTag", "updateTest1"),
					resource.TestCheckResourceAttr(resourceName, "tags.secondTag", "updateTest2"),
				),
			},
			{
				Config: testAccApplicationConfig_tags(rName, "updateTest1", "updateTest2", "addTest3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.firstTag", "updateTest1"),
					resource.TestCheckResourceAttr(resourceName, "tags.secondTag", "updateTest2"),
					resource.TestCheckResourceAttr(resourceName, "tags.thirdTag", "addTest3"),
				),
			},
			{
				Config: testAccApplicationConfig_tags2(rName, "updateTest1", "updateTest2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.firstTag", "updateTest1"),
					resource.TestCheckResourceAttr(resourceName, "tags.secondTag", "updateTest2"),
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

func testAccCheckApplicationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticBeanstalkConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elastic_beanstalk_application" {
			continue
		}

		// Try to find the application
		DescribeBeanstalkAppOpts := &elasticbeanstalk.DescribeApplicationsInput{
			ApplicationNames: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeApplications(DescribeBeanstalkAppOpts)
		if err == nil {
			if len(resp.Applications) > 0 {
				return fmt.Errorf("Elastic Beanstalk Application still exists.")
			}
			return nil
		}

		if !tfawserr.ErrCodeEquals(err, "InvalidBeanstalkAppID.NotFound") {
			return err
		}
	}

	return nil
}

func testAccCheckApplicationExists(n string, app *elasticbeanstalk.ApplicationDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Elastic Beanstalk app ID is not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticBeanstalkConn
		DescribeBeanstalkAppOpts := &elasticbeanstalk.DescribeApplicationsInput{
			ApplicationNames: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeApplications(DescribeBeanstalkAppOpts)
		if err != nil {
			return err
		}
		if len(resp.Applications) == 0 {
			return fmt.Errorf("Elastic Beanstalk Application not found.")
		}

		*app = *resp.Applications[0]

		return nil
	}
}

func testAccApplicationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "tftest" {
  name        = "%s"
  description = "tf-test-desc"
}
`, rName)
}

func testAccApplicationConfig_serviceRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "beanstalk_service" {
  name = "%[1]s"

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

resource "aws_iam_role_policy" "beanstalk_service" {
  name = "%[1]s"
  role = aws_iam_role.beanstalk_service.id

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AllowOperations",
            "Effect": "Allow",
            "Action": [
                "iam:PassRole"
            ],
            "Resource": [
                "*"
            ]
        }
    ]
}
EOF
}
`, rName)
}

func testAccApplicationConfig_maxAge(rName string) string {
	return testAccApplicationConfig_serviceRole(rName) + fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "tftest" {
  name        = "%s"
  description = "tf-test-desc"

  appversion_lifecycle {
    service_role          = aws_iam_role.beanstalk_service.arn
    max_age_in_days       = 90
    delete_source_from_s3 = true
  }
}
`, rName)
}

func testAccApplicationConfig_maxCount(rName string) string {
	return testAccApplicationConfig_serviceRole(rName) + fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "tftest" {
  name        = "%s"
  description = "tf-test-desc"

  appversion_lifecycle {
    service_role          = aws_iam_role.beanstalk_service.arn
    max_count             = 10
    delete_source_from_s3 = false
  }
}
`, rName)
}

func testAccApplicationConfig_tags2(rName, tag1, tag2 string) string {
	return fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "tftest" {
  name        = "%s"
  description = "tf-test-desc"

  tags = {
    firstTag  = "%s"
    secondTag = "%s"
  }
}
`, rName, tag1, tag2)
}

func testAccApplicationConfig_tags(rName, tag1, tag2, tag3 string) string {
	return fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "tftest" {
  name        = "%s"
  description = "tf-test-desc"

  tags = {
    firstTag  = "%s"
    secondTag = "%s"
    thirdTag  = "%s"
  }
}
`, rName, tag1, tag2, tag3)
}
