// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCEndpointServicePrivateDNSVerification_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("tfacctest") // 32 character limit
	domainName := acctest.RandomDomainName()
	resourceName := "aws_vpc_endpoint_service_private_dns_verification.test"
	endpointServiceResourceName := "aws_vpc_endpoint_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServicePrivateDNSVerificationConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "service_id", endpointServiceResourceName, names.AttrID),
				),
			},
		},
	})
}

func TestAccVPCEndpointServicePrivateDNSVerification_waitForVerification(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("tfacctest") // 32 character limit
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServicePrivateDNSVerificationConfig_waitForVerification(rName, domainName),
				// Expect an error as private DNS setup and verification is not
				// included in this configuration. This test simply verifies the
				// create waiter functions as expected.
				ExpectError: regexache.MustCompile("waiting for VPC Endpoint Service Private DNS Verification"),
			},
		},
	})
}

func testAccVPCEndpointServicePrivateDNSVerificationConfig_base(rName, domainName string, count int) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_lb" "test" {
  count = %[3]d

  load_balancer_type = "network"
  name               = "%[1]s-${count.index}"

  subnets = aws_subnet.test[*].id

  internal                   = true
  idle_timeout               = 60
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  network_load_balancer_arns = aws_lb.test[*].arn
  private_dns_name           = %[2]q
}
`, rName, domainName, count))
}

func testAccVPCEndpointServicePrivateDNSVerificationConfig_basic(rName, domainName string) string {
	return acctest.ConfigCompose(
		testAccVPCEndpointServicePrivateDNSVerificationConfig_base(rName, domainName, 1),
		`
resource "aws_vpc_endpoint_service_private_dns_verification" "test" {
  service_id = aws_vpc_endpoint_service.test.id
}
`)
}

func testAccVPCEndpointServicePrivateDNSVerificationConfig_waitForVerification(rName, domainName string) string {
	return acctest.ConfigCompose(
		testAccVPCEndpointServicePrivateDNSVerificationConfig_base(rName, domainName, 1),
		`
resource "aws_vpc_endpoint_service_private_dns_verification" "test" {
  service_id = aws_vpc_endpoint_service.test.id

  wait_for_verification = true
  timeouts {
    create = "1m"
  }
}
`)
}
