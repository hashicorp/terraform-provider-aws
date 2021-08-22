package aws

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go/service/greengrass"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsIotGreengrassServiceRole_basic(t *testing.T) {
	resourceName := "aws_iot_greengrass_service_role.test"
	roleResourceName := "aws_iam_role.greengrass_service_role"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, greengrass.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsIotGreengrassServiceRole_Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotGreengrassServiceRoleConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsIotGreengrassServiceRole_Exists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", roleResourceName, "arn"),
				),
			},
		},
	})
}

func testAccCheckAwsIotGreengrassServiceRole_Destroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).greengrassconn
	input := &greengrass.GetServiceRoleForAccountInput{}

	_, err := conn.GetServiceRoleForAccount(input)

	if err != nil {
		if tfawserr.ErrStatusCodeEquals(err, http.StatusNotFound) {
			//No greengrass service role is set for this account
			return nil
		}
		return err
	}
	return errors.New("greengrass service role was not reset")
}

func testAccCheckAwsIotGreengrassServiceRole_Exists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccAWSIotGreengrassServiceRoleConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "greengrass_service_role" {
  name               = "greengrass_service_role_test_%[1]d"
  assume_role_policy = <<-EOT
  {
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Principal": {
          "Service": "greengrass.amazonaws.com"
        },
        "Action": "sts:AssumeRole"
      }
    ]
  }
  EOT
}

resource "aws_iot_greengrass_service_role" "test" {
  role_arn = aws_iam_role.greengrass_service_role.arn
}
`, rInt)
}
