// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2SerialConsoleAccess_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Resource": {
			acctest.CtBasic: testAccEC2SerialConsoleAccess_basic,
			"Identity":      testAccEC2SerialConsoleAccess_identitySerial,
		},
		"DataSource": {
			acctest.CtBasic: testAccEC2SerialConsoleAccessDataSource_basic,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccEC2SerialConsoleAccess_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_serial_console_access.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSerialConsoleAccessDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSerialConsoleAccessConfig_basic(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSerialConsoleAccess(ctx, resourceName, false),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSerialConsoleAccessConfig_basic(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSerialConsoleAccess(ctx, resourceName, true),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
		},
	})
}

func testAccCheckSerialConsoleAccessDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindSerialConsoleAccessStatus(ctx, conn)

		if retry.NotFound(err) {
			return nil
		}

		if err != nil {
			return err
		}

		if aws.ToBool(output.SerialConsoleAccessEnabled) != false {
			return fmt.Errorf("EC2 Serial Console Access not disabled on resource removal")
		}

		return nil
	}
}

func testAccCheckSerialConsoleAccess(ctx context.Context, n string, enabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindSerialConsoleAccessStatus(ctx, conn)

		if err != nil {
			return err
		}

		if aws.ToBool(output.SerialConsoleAccessEnabled) != enabled {
			return fmt.Errorf("EC2 Serial Console Access is not in expected state (%t)", enabled)
		}

		return nil
	}
}

func testAccSerialConsoleAccessConfig_basic(enabled bool) string {
	return fmt.Sprintf(`
resource "aws_ec2_serial_console_access" "test" {
  enabled = %[1]t
}
`, enabled)
}
