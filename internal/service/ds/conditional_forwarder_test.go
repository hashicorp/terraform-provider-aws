// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfds "github.com/hashicorp/terraform-provider-aws/internal/service/ds"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDSConditionalForwarder_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_directory_service_conditional_forwarder.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()
	ip1, ip2, ip3 := "8.8.8.8", "1.1.1.1", "8.8.4.4"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckDirectoryService(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConditionalForwarderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConditionalForwarderConfig_basic(rName, domainName, ip1, ip2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConditionalForwarderExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "dns_ips.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "dns_ips.0", ip1),
					resource.TestCheckResourceAttr(resourceName, "dns_ips.1", ip2),
				),
			},
			{
				Config: testAccConditionalForwarderConfig_basic(rName, domainName, ip1, ip3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConditionalForwarderExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "dns_ips.#", "2"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()
	ip1, ip2 := "8.8.8.8", "1.1.1.1"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckDirectoryService(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConditionalForwarderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConditionalForwarderConfig_basic(rName, domainName, ip1, ip2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConditionalForwarderExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfds.ResourceConditionalForwarder(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckConditionalForwarderDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_directory_service_conditional_forwarder" {
				continue
			}

			_, err := tfds.FindConditionalForwarderByTwoPartKey(ctx, conn, rs.Primary.Attributes["directory_id"], rs.Primary.Attributes["remote_domain_name"])

			if retry.NotFound(err) {
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

func testAccCheckConditionalForwarderExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DSClient(ctx)

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
