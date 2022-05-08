package iam_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
)

func TestAccIAMInstanceProfile_basic(t *testing.T) {
	var conf iam.GetInstanceProfileOutput
	resourceName := "aws_iam_instance_profile.test"
	rName := sdkacctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceProfileConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceProfileExists(resourceName, &conf),
					acctest.CheckResourceAttrGlobalARN(resourceName, "arn", "iam", fmt.Sprintf("instance-profile/test-%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "role", "aws_iam_role.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccIAMInstanceProfile_withoutRole(t *testing.T) {
	var conf iam.GetInstanceProfileOutput
	resourceName := "aws_iam_instance_profile.test"
	rName := sdkacctest.RandString(5)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceProfileWithoutRoleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceProfileExists(resourceName, &conf),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccIAMInstanceProfile_tags(t *testing.T) {
	var conf iam.GetInstanceProfileOutput
	resourceName := "aws_iam_instance_profile.test"
	rName := sdkacctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceProfileTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceProfileExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
			{
				Config: testAccInstanceProfileTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceProfileExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccInstanceProfileTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceProfileExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccIAMInstanceProfile_namePrefix(t *testing.T) {
	var conf iam.GetInstanceProfileOutput
	rName := sdkacctest.RandString(5)
	resourceName := "aws_iam_instance_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceProfilePrefixNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceProfileExists(resourceName, &conf),
					testAccCheckInstanceProfileGeneratedNamePrefix(
						resourceName, "test-"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccIAMInstanceProfile_disappears(t *testing.T) {
	var conf iam.GetInstanceProfileOutput
	resourceName := "aws_iam_instance_profile.test"
	rName := sdkacctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceProfileConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceProfileExists(resourceName, &conf),
					acctest.CheckResourceDisappears(acctest.Provider, tfiam.ResourceInstanceProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMInstanceProfile_Disappears_role(t *testing.T) {
	var conf iam.GetInstanceProfileOutput
	resourceName := "aws_iam_instance_profile.test"
	rName := sdkacctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceProfileConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceProfileExists(resourceName, &conf),
					acctest.CheckResourceDisappears(acctest.Provider, tfiam.ResourceRole(), "aws_iam_role.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckInstanceProfileGeneratedNamePrefix(resource, prefix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Resource not found")
		}
		name, ok := r.Primary.Attributes["name"]
		if !ok {
			return fmt.Errorf("Name attr not found: %#v", r.Primary.Attributes)
		}
		if !strings.HasPrefix(name, prefix) {
			return fmt.Errorf("Name: %q, does not have prefix: %q", name, prefix)
		}
		return nil
	}
}

func testAccCheckInstanceProfileDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_instance_profile" {
			continue
		}

		// Try to get role
		_, err := conn.GetInstanceProfile(&iam.GetInstanceProfileInput{
			InstanceProfileName: aws.String(rs.Primary.ID),
		})
		if err == nil {
			return fmt.Errorf("still exist.")
		}

		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			continue
		}

		return err
	}

	return nil
}

func testAccCheckInstanceProfileExists(n string, res *iam.GetInstanceProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Instance Profile name is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

		resp, err := conn.GetInstanceProfile(&iam.GetInstanceProfileInput{
			InstanceProfileName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*res = *resp

		return nil
	}
}

func testAccInstanceProfileBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = "test-%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}
`, rName)
}

func testAccInstanceProfileConfig(rName string) string {
	return testAccInstanceProfileBaseConfig(rName) + fmt.Sprintf(`
resource "aws_iam_instance_profile" "test" {
  name = "test-%[1]s"
  role = aws_iam_role.test.name
}
`, rName)
}

func testAccInstanceProfileWithoutRoleConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_instance_profile" "test" {
  name = "test-%s"
}
`, rName)
}

func testAccInstanceProfilePrefixNameConfig(rName string) string {
	return testAccInstanceProfileBaseConfig(rName) + `
resource "aws_iam_instance_profile" "test" {
  name_prefix = "test-"
  role        = aws_iam_role.test.name
}
`
}

func testAccInstanceProfileTags1Config(rName, tagKey1, tagValue1 string) string {
	return testAccInstanceProfileBaseConfig(rName) + fmt.Sprintf(`
resource "aws_iam_instance_profile" "test" {
  name = "test-%[1]s"
  role = aws_iam_role.test.name

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccInstanceProfileTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccInstanceProfileBaseConfig(rName) + fmt.Sprintf(`
resource "aws_iam_instance_profile" "test" {
  name = "test-%[1]s"
  role = aws_iam_role.test.name

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
