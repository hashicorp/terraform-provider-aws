// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/apprunner"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapprunner "github.com/hashicorp/terraform-provider-aws/internal/service/apprunner"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppRunnerVPCConnector_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_vpc_connector.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckVPCConnector(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCConnectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConnectorConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCConnectorExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "apprunner", regexache.MustCompile(fmt.Sprintf(`vpcconnector/%s/1/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "vpc_connector_name", rName),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccAppRunnerVPCConnector_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_vpc_connector.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckVPCConnector(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCConnectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConnectorConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCConnectorExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapprunner.ResourceVPCConnector(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVPCConnectorDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_apprunner_vpc_connector" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).AppRunnerClient(ctx)

			_, err := tfapprunner.FindVPCConnectorByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("App Runner VPC Connector %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVPCConnectorExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppRunnerClient(ctx)

		_, err := tfapprunner.FindVPCConnectorByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccVPCConnectorConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCConnectorConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVPCConnectorConfig_base(rName), fmt.Sprintf(`
resource "aws_apprunner_vpc_connector" "test" {
  vpc_connector_name = %[1]q
  subnets            = aws_subnet.test[*].id
  security_groups    = [aws_security_group.test.id]
}
`, rName))
}

func testAccPreCheckVPCConnector(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppRunnerClient(ctx)

	input := &apprunner.ListVpcConnectorsInput{}

	_, err := conn.ListVpcConnectors(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
