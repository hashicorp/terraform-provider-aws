package ssoadmin_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfssoadmin "github.com/hashicorp/terraform-provider-aws/internal/service/ssoadmin"
)

func TestAccSSOAdminAccessControlAttributes_basic(t *testing.T) {
	resourceName := "aws_ssoadmin_instance_access_control_attributes.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckInstances(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessControlAttributesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceAccessControlAttributesConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessControlAttributesExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "attribute.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "status", "ENABLED"),
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
func TestAccSSOAdminAccessControlAttributes_multiple(t *testing.T) {
	resourceName := "aws_ssoadmin_instance_access_control_attributes.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckInstances(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessControlAttributesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceAccessControlAttributesConfig_multiple(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessControlAttributesExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "attribute.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "status", "ENABLED"),
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

func TestAccSSOAdminAccessControlAttributes_update(t *testing.T) {
	resourceName := "aws_ssoadmin_instance_access_control_attributes.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckInstances(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessControlAttributesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceAccessControlAttributesConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessControlAttributesExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceAccessControlAttributesConfig_update(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessControlAttributesExists(resourceName),
				),
			},
		},
	})
}

func TestAccSSOAdminAccessControlAttributes_disappears(t *testing.T) {
	resourceName := "aws_ssoadmin_permission_set_inline_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckInstances(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionSetInlinePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceAccessControlAttributesConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionSetInlinePolicyExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfssoadmin.ResourcePermissionSetInlinePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAccessControlAttributesDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminConn()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssoadmin_instance_access_control_attributes" {
			continue
		}

		input := &ssoadmin.DescribeInstanceAccessControlAttributeConfigurationInput{
			InstanceArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeInstanceAccessControlAttributeConfiguration(input)
		if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if output == nil {
			continue
		}

		// SSO API returns empty string when removed from Permission Set
		if len(output.InstanceAccessControlAttributeConfiguration.AccessControlAttributes) == 0 {
			continue
		}

		return fmt.Errorf("Access control attributes for SSO Instance (%s) still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAccessControlAttributesExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		input := &ssoadmin.DescribeInstanceAccessControlAttributeConfigurationInput{
			InstanceArn: aws.String(rs.Primary.ID),
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminConn()
		output, err := conn.DescribeInstanceAccessControlAttributeConfiguration(input)
		if err != nil {
			return err
		}

		if output == nil || len(output.InstanceAccessControlAttributeConfiguration.AccessControlAttributes) == 0 {
			return fmt.Errorf("Access Control Attributes for SSO Instance Set (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccInstanceAccessControlAttributesConfig_basic() string {
	return `
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_instance_access_control_attributes" "test" {
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  attribute {
    key = "name"
    value {
      source = ["$${path:name.givenName}"]
    }
  }
}
`
}
func testAccInstanceAccessControlAttributesConfig_multiple() string {
	return `
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_instance_access_control_attributes" "test" {
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  attribute {
    key = "name"
    value {
      source = ["$${path:name.givenName}"]
    }
  }
  attribute {
    key = "last"
    value {
      source = ["$${path:name.familyName}"]
    }
  }
}
`
}

func testAccInstanceAccessControlAttributesConfig_update() string {
	return `
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_instance_access_control_attributes" "test" {
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  attribute {
    key = "name"
    value {
      source = ["$${path:name.familyName}"]
    }
  }
}
`
}
