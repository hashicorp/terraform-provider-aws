package iam_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccIAMInstanceProfileDataSource_basic(t *testing.T) {
	resourceName := "data.aws_iam_instance_profile.test"

	roleName := fmt.Sprintf("tf-acc-ds-instance-profile-role-%d", sdkacctest.RandInt())
	profileName := fmt.Sprintf("tf-acc-ds-instance-profile-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceProfileDataSourceConfig(roleName, profileName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", "aws_iam_instance_profile.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "path", "/testpath/"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "role_id", "aws_iam_role.test", "unique_id"),
					resource.TestCheckResourceAttr(resourceName, "role_name", roleName),
				),
			},
		},
	})
}

func testAccInstanceProfileDataSourceConfig(roleName, profileName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = "%s"
  assume_role_policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":[\"ec2.amazonaws.com\"]},\"Action\":[\"sts:AssumeRole\"]}]}"
}

resource "aws_iam_instance_profile" "test" {
  name = "%s"
  role = aws_iam_role.test.name
  path = "/testpath/"
}

data "aws_iam_instance_profile" "test" {
  name = aws_iam_instance_profile.test.name
}
`, roleName, profileName)
}
