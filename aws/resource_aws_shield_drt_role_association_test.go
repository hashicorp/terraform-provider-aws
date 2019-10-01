package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/shield"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSShieldDrtRoleAssociation_RoleArn(t *testing.T) {
	resourceName := "aws_shield_drt_role.acctest"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	roleArn := "arn:aws:iam::aws:policy/service-role/AWSShieldDRTAccessPolicy"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSShield(t)
			testAccPreCheckAWSShieldDrtRoleAssociation(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSShieldDrtRoleAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSShieldDrtRoleAssociationConfig(rName, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSShieldDrtRoleAssociated(resourceName, roleArn),
				),
			},
		},
	})
}

func testAccAWSShieldDrtRoleAssociationConfig(refName, roleArn string) string {
	return fmt.Sprintf(`
variable "name" {
  default = "%s"
}

resource "aws_shield_protection" "acctest" {
  name = "${var.name}"
}

resource "aws_shield_drt_role" "acctest" {
  role_arn   = "%s"
  depends_on = ["aws_shield_protection.acctest"]
}
`, refName, roleArn)
}

func testAccPreCheckAWSShieldDrtRoleAssociation(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).shieldconn

	input := &shield.DescribeDRTAccessInput{}

	resp, err := conn.DescribeDRTAccess(input)
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %v", err)
	}

	if resp.RoleArn != nil {
		t.Fatalf("A Shield DRT Role is already associated")
	}
}

func testAccCheckAWSShieldDrtRoleAssociated(name, roleArn string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).shieldconn

		input := &shield.DescribeDRTAccessInput{}

		resp, err := conn.DescribeDRTAccess(input)
		if isAWSErr(err, shield.ErrCodeResourceNotFoundException, "") {
			return fmt.Errorf("AWS Shield DRT Role %s not associated", roleArn)
		}

		if err != nil {
			return err
		}

		if aws.StringValue(resp.RoleArn) != roleArn {
			return fmt.Errorf("AWS Shield DRT Role %s not associated", roleArn)
		}

		return nil
	}
}

func testAccCheckAWSShieldDrtRoleAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).shieldconn

	input := &shield.DescribeDRTAccessInput{}

	resp, err := conn.DescribeDRTAccess(input)
	if err != nil {
		return err
	}

	if resp != nil && aws.StringValue(resp.RoleArn) != "" {
		return fmt.Errorf("The Shield DRT Role %v still associated", aws.StringValue(resp.RoleArn))
	}

	return nil
}
