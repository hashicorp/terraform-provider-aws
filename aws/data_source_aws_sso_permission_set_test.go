package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfawsresource"
)

func TestAccDataSourceAwsSsoPermissionSetBasic(t *testing.T) {
	datasourceName := "data.aws_sso_permission_set.test"
	rName := acctest.RandomWithPrefix("tf-sso-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSSSOInstance(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsSsoPermissionSetConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "managed_policy_arns.#", "1"),
					tfawsresource.TestCheckTypeSetElemAttr(datasourceName, "managed_policy_arns.*", "arn:aws:iam::aws:policy/ReadOnlyAccess"),
					resource.TestCheckResourceAttr(datasourceName, "name", fmt.Sprintf("%s", rName)),
					resource.TestCheckResourceAttr(datasourceName, "description", "testing"),
					resource.TestCheckResourceAttr(datasourceName, "session_duration", "PT1H"),
					resource.TestCheckResourceAttr(datasourceName, "relay_state", "https://console.aws.amazon.com/console/home"),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsSsoPermissionSetByTags(t *testing.T) {
	datasourceName := "data.aws_sso_permission_set.test"
	rName := acctest.RandomWithPrefix("tf-sso-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSSSOInstance(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsSsoPermissionSetConfigByTags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "managed_policy_arns.#", "1"),
					tfawsresource.TestCheckTypeSetElemAttr(datasourceName, "managed_policy_arns.*", "arn:aws:iam::aws:policy/ReadOnlyAccess"),
					resource.TestCheckResourceAttr(datasourceName, "name", fmt.Sprintf("%s", rName)),
					resource.TestCheckResourceAttr(datasourceName, "description", "testing"),
					resource.TestCheckResourceAttr(datasourceName, "session_duration", "PT1H"),
					resource.TestCheckResourceAttr(datasourceName, "relay_state", "https://console.aws.amazon.com/console/home"),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "3"),
				),
			},
		},
	})
}

func testAccDataSourceAwsSsoPermissionSetConfigBasic(rName string) string {
	return fmt.Sprintf(`
data "aws_sso_instance" "selected" {}

resource "aws_sso_permission_set" "test" {
	name                = "%s"
	description         = "testing"
	instance_arn        = data.aws_sso_instance.selected.arn
	session_duration    = "PT1H"
	relay_state         = "https://console.aws.amazon.com/console/home"
	managed_policy_arns = ["arn:aws:iam::aws:policy/ReadOnlyAccess"]
}

data "aws_sso_permission_set" "test" {
	instance_arn = data.aws_sso_instance.selected.arn
	name         = aws_sso_permission_set.test.name
}
`, rName)
}

func testAccDataSourceAwsSsoPermissionSetConfigByTags(rName string) string {
	return fmt.Sprintf(`
data "aws_sso_instance" "selected" {}

resource "aws_sso_permission_set" "test" {
	name                = "%s"
	description         = "testing"
	instance_arn        = data.aws_sso_instance.selected.arn
	session_duration    = "PT1H"
	relay_state         = "https://console.aws.amazon.com/console/home"
	managed_policy_arns = ["arn:aws:iam::aws:policy/ReadOnlyAccess"]

	tags = {
		Key1 = "Value1"
		Key2 = "Value2"
		Key3 = "Value3"
	}
}

data "aws_sso_permission_set" "test" {
	instance_arn = data.aws_sso_instance.selected.arn
	name         = aws_sso_permission_set.test.name

	tags = {
	Key1 = "Value1"
	Key2 = "Value2"
	Key3 = "Value3"
	}
}
`, rName)
}
