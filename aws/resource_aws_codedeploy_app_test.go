package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/codedeploy"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCodeDeployApp_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeDeployApp,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_codedeploy_app.foo", "compute_platform", "Server"),
					testAccCheckAWSCodeDeployAppExists("aws_codedeploy_app.foo"),
				),
			},
			{
				Config: testAccAWSCodeDeployAppModified,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployAppExists("aws_codedeploy_app.foo"),
				),
			},
		},
	})
}

func TestAccAWSCodeDeployApp_computePlatform(t *testing.T) {
	rName := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeDeployAppConfigComputePlatform(rName, "Server"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployAppExists("aws_codedeploy_app.foo"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_app.foo", "compute_platform", "Server"),
				),
			},
			{
				Config: testAccAWSCodeDeployAppConfigComputePlatform(rName, "Lambda"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployAppExists("aws_codedeploy_app.foo"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_app.foo", "compute_platform", "Lambda"),
				),
			},
		},
	})
}

func testAccCheckAWSCodeDeployAppDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).codedeployconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codedeploy_app" {
			continue
		}

		_, err := conn.GetApplication(&codedeploy.GetApplicationInput{
			ApplicationName: aws.String(rs.Primary.Attributes["name"]),
		})

		if err != nil {
			// Verify the error is what we want
			if ae, ok := err.(awserr.Error); ok && ae.Code() == "ApplicationDoesNotExistException" {
				continue
			}
			return err
		}

		return fmt.Errorf("still exists")
	}

	return nil
}

func testAccCheckAWSCodeDeployAppExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccAWSCodeDeployAppConfigComputePlatform(rName string, value string) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_app" "foo" {
	name = "test-codedeploy-app-%s"
	compute_platform = "%s"
}`, rName, value)
}

var testAccAWSCodeDeployApp = `
resource "aws_codedeploy_app" "foo" {
	name = "foo"
}`

var testAccAWSCodeDeployAppModified = `
resource "aws_codedeploy_app" "foo" {
	name = "bar"
}`
