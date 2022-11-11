package controltower_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/controltower"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfct "github.com/hashicorp/terraform-provider-aws/internal/service/controltower"
)

func TestAccControlTowerControl_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Control": {
			"basic":      testAccControl_basic,
			"disappears": testAccControl_disappears,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}

func testAccControl_basic(t *testing.T) {
	var control controltower.EnabledControlSummary
	resourceName := "aws_controltower_control.test"
	controlName := "AWS-GR_EC2_VOLUME_INUSE_CHECK"
	ouName := "Security"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationManagementAccount(t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, controltower.EndpointsID),
		CheckDestroy:             testAccCheckControlDestroy,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccControlConfig_basic(controlName, ouName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckControlExists(resourceName, &control),
					resource.TestCheckResourceAttrSet(resourceName, "control_identifier"),
				),
			},
		},
	})
}

func testAccControl_disappears(t *testing.T) {
	var control controltower.EnabledControlSummary
	resourceName := "aws_controltower_control.test"
	controlName := "AWS-GR_EC2_VOLUME_INUSE_CHECK"
	ouName := "Security"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationManagementAccount(t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, controltower.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckControlDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccControlConfig_basic(controlName, ouName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckControlExists(resourceName, &control),
					acctest.CheckResourceDisappears(acctest.Provider, tfct.ResourceControl(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckControlExists(n string, control *controltower.EnabledControlSummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ControlTower Control ID is set")
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
				*control = *c
				return nil
			}
		}

		return fmt.Errorf("Expected Control Tower Control to be created, but wasn't found")
	}
}

func testAccCheckControlDestroy(s *terraform.State) error {
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
