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

func TestAccAWSShieldDrtRoleAssociation(t *testing.T) {
	iamResourceName := "aws_iam_role.shield_drt"
	shieldResourceName := "aws_shield_drt_role_association.shield_drt"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

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
				Config: testAccAWSShieldDrtRoleAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSShieldDrtRoleAssociated(iamResourceName, shieldResourceName),
				),
			},
		},
	})
}

func testAccAWSShieldDrtRoleAssociationConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "shield_drt" {
  name = "%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": [
          "drt.shield.amazonaws.com"
        ]
       }
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "shield_drt" {
  role       = "${aws_iam_role.shield_drt.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSShieldDRTAccessPolicy"
}

resource "aws_shield_drt_role_association" "shield_drt" {
  role_arn = "${aws_iam_role.shield_drt.arn}"
}
`, rName)
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

func testAccCheckAWSShieldDrtRoleAssociated(iamName, drtName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		role, ok := s.RootModule().Resources[iamName]
		if !ok {
			return fmt.Errorf("Not found: %s", iamName)
		}

		_, ok = s.RootModule().Resources[drtName]
		if !ok {
			return fmt.Errorf("Not found: %s", drtName)
		}

		conn := testAccProvider.Meta().(*AWSClient).shieldconn

		input := &shield.DescribeDRTAccessInput{}

		resp, err := conn.DescribeDRTAccess(input)
		if isAWSErr(err, shield.ErrCodeResourceNotFoundException, "") {
			return fmt.Errorf("AWS Shield DRT Role not associated")
		}

		if err != nil {
			return err
		}

		if aws.StringValue(resp.RoleArn) != role.Primary.Attributes["arn"] {
			return fmt.Errorf("AWS Shield DRT Role incorrectly associated")
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
