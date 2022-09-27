package controltower_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/controltower"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccControlTowerControl_basic(t *testing.T) {
	resourceName := "aws_controltower_control.test"
	controlName := "AWS-GR_EC2_VOLUME_INUSE_CHECK"
	ouName := "Security"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckControlTowerDeployed(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, controltower.EndpointsID),
		CheckDestroy:             testAccCheckControlDisable,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccControlConfig_basic(controlName, ouName),
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrGlobalARN(resourceName, "target_identifier", "organizations", regexp.MustCompile(`ou/.+`)),
				),
			},
		},
	})
}

func testAccControlConfig_basic(controlName string, ouName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_partition" "current" {}

data "aws_organizations_organization" "test" {}

data "aws_organizations_organizational_units" "test" {
  parent_id = data.aws_organizations_organization.test.roots[0].id
}

resource "aws_controltower_control" "test" {
  control_identifier = "arn:${data.aws_partition.current.partition}:controltower:${data.aws_region.current.name}::control/%[1]s"
  target_identifier = [
    for x in data.aws_organizations_organizational_units.test.children :
    x.arn if x.name == "%[2]s"
  ][0]
}
`, controlName, ouName)
}

func testAccCheckControlDisable(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_controltower_control" {
			continue
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).ControlTowerConn
		input := &controltower.ListEnabledControlsInput{
			TargetIdentifier: &rs.Primary.ID,
		}

		output, err := conn.ListEnabledControls(input)

		if err != nil {
			return err
		}

		for _, c := range output.EnabledControls {
			if *c.ControlIdentifier == rs.Primary.Attributes["control_identifier"] {
				return fmt.Errorf("ControlTower Control still enabled: %s", rs.Primary.ID)
			}
		}
	}

	return nil
}
