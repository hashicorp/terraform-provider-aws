// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks_test

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfopsworks "github.com/hashicorp/terraform-provider-aws/internal/service/opsworks"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccCheckLayerExists(ctx context.Context, n string, v *opsworks.Layer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No OpsWorks Layer ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OpsWorksConn(ctx)

		output, err := tfopsworks.FindLayerByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckLayerDestroy(ctx context.Context, resourceType string, s *terraform.State) error { // nosemgrep:ci.semgrep.acctest.naming.destroy-check-signature
	conn := acctest.Provider.Meta().(*conns.AWSClient).OpsWorksConn(ctx)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != resourceType {
			continue
		}

		_, err := tfopsworks.FindLayerByID(ctx, conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("OpsWorks Layer %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccLayerConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccStackConfig_basic(rName), fmt.Sprintf(`
resource "aws_security_group" "test" {
  count = 2

  name   = "%[1]s-${count.index}"
  vpc_id = aws_vpc.test.id

  ingress {
    from_port   = 8
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccLayerConfig_baseAlternateRegion(rName string) string {
	return acctest.ConfigCompose(testAccStackConfig_baseVPCAlternateRegion(rName), fmt.Sprintf(`
resource "aws_opsworks_stack" "test" {
  name                         = %[1]q
  region                       = %[2]q
  service_role_arn             = aws_iam_role.opsworks_service.arn
  default_instance_profile_arn = aws_iam_instance_profile.opsworks_instance.arn
  default_subnet_id            = aws_subnet.test[0].id
  vpc_id                       = aws_vpc.test.id
  use_opsworks_security_groups = false
}

resource "aws_security_group" "test" {
  count = 2

  provider = "awsalternate"

  name   = "%[1]s-${count.index}"
  vpc_id = aws_vpc.test.id

  ingress {
    from_port   = 8
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, acctest.AlternateRegion()))
}
