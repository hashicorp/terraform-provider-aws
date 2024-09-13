// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfds "github.com/hashicorp/terraform-provider-aws/internal/service/ds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDSConditionalForwarder_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_directory_service_conditional_forwarder.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()
	ip1, ip2, ip3 := "8.8.8.8", "1.1.1.1", "8.8.4.4"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckDirectoryService(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConditionalForwarderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConditionalForwarderConfig_basic(rName, domainName, ip1, ip2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConditionalForwarderExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "dns_ips.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "dns_ips.0", ip1),
					resource.TestCheckResourceAttr(resourceName, "dns_ips.1", ip2),
				),
			},
			{
				Config: testAccConditionalForwarderConfig_basic(rName, domainName, ip1, ip3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConditionalForwarderExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "dns_ips.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "dns_ips.0", ip1),
					resource.TestCheckResourceAttr(resourceName, "dns_ips.1", ip3),
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

func TestAccDSConditionalForwarder_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_directory_service_conditional_forwarder.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()
	ip1, ip2 := "8.8.8.8", "1.1.1.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckDirectoryService(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConditionalForwarderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConditionalForwarderConfig_basic(rName, domainName, ip1, ip2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConditionalForwarderExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfds.ResourceConditionalForwarder(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckConditionalForwarderDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_directory_service_conditional_forwarder" {
				continue
			}

			_, err := tfds.FindConditionalForwarderByTwoPartKey(ctx, conn, rs.Primary.Attributes["directory_id"], rs.Primary.Attributes["remote_domain_name"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Directory Service Conditional Forwarder %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckConditionalForwarderExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DSClient(ctx)

		_, err := tfds.FindConditionalForwarderByTwoPartKey(ctx, conn, rs.Primary.Attributes["directory_id"], rs.Primary.Attributes["remote_domain_name"])

		return err
	}
}

func testAccConditionalForwarderConfig_basic(rName, domain, ip1, ip2 string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_directory_service_conditional_forwarder" "test" {
  directory_id = aws_directory_service_directory.test.id

  remote_domain_name = "test.example.com"

  dns_ips = [
    %[3]q,
    %[4]q,
  ]
}

resource "aws_directory_service_directory" "test" {
  name     = %[2]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, domain, ip1, ip2))
}
