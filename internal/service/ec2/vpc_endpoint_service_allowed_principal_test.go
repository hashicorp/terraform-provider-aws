// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCEndpointServiceAllowedPrincipal_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("tfacctest")

	resourceName := "aws_vpc_endpoint_service_allowed_principal.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointServiceAllowedPrincipalDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceAllowedPrincipalConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointServiceAllowedPrincipalExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, names.AttrID, regexache.MustCompile(`^vpce-svc-perm-\w{17}$`)),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_endpoint_service_id", "aws_vpc_endpoint_service.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "principal_arn", "data.aws_iam_session_context.current", "issuer_arn"),
				),
			},
		},
	})
}

func TestAccVPCEndpointServiceAllowedPrincipal_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("tfacctest")

	resourceName := "aws_vpc_endpoint_service_allowed_principal.test"
	serviceResourceName := "aws_vpc_endpoint_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointServiceAllowedPrincipalDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceAllowedPrincipalConfig_Multiple(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointServiceAllowedPrincipalExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, names.AttrID, regexache.MustCompile(`^vpce-svc-perm-\w{17}$`)),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_endpoint_service_id", "aws_vpc_endpoint_service.test", names.AttrID),
					resource.TestCheckResourceAttr(serviceResourceName, "allowed_principals.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "principal_arn", "data.aws_iam_session_context.current", "issuer_arn"),
				),
			},
		},
	})
}

func TestAccVPCEndpointServiceAllowedPrincipal_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("tfacctest")

	resourceName := "aws_vpc_endpoint_service_allowed_principal.test"
	tagResourceName := "aws_ec2_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointServiceAllowedPrincipalDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointServiceAllowedPrincipalConfig_tag(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointServiceAllowedPrincipalExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(tagResourceName, names.AttrResourceID, resourceName, names.AttrID),
					resource.TestCheckResourceAttr(tagResourceName, names.AttrKey, "Name"),
					resource.TestCheckResourceAttr(tagResourceName, names.AttrValue, rName),
				),
			},
		},
	})
}

func TestAccVPCEndpointServiceAllowedPrincipal_migrateID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("tfacctest")

	resourceName := "aws_vpc_endpoint_service_allowed_principal.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy: testAccCheckVPCEndpointServiceAllowedPrincipalDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "4.63.0",
					},
				},
				Config: testAccVPCEndpointServiceAllowedPrincipalConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointServiceAllowedPrincipalExists(ctx, resourceName),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccVPCEndpointServiceAllowedPrincipalConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointServiceAllowedPrincipalExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, names.AttrID, regexache.MustCompile(`^vpce-svc-perm-\w{17}$`)),
				),
			},
		},
	})
}

// Verify that the resource returns an ID usable for creating an `aws_ec2_tag`
func TestAccVPCEndpointServiceAllowedPrincipal_migrateAndTag(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("tfacctest")

	resourceName := "aws_vpc_endpoint_service_allowed_principal.test"
	tagResourceName := "aws_ec2_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy: testAccCheckVPCEndpointServiceAllowedPrincipalDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "4.63.0",
					},
				},
				Config: testAccVPCEndpointServiceAllowedPrincipalConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointServiceAllowedPrincipalExists(ctx, resourceName),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccVPCEndpointServiceAllowedPrincipalConfig_tag(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointServiceAllowedPrincipalExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, names.AttrID, regexache.MustCompile(`^vpce-svc-perm-\w{17}$`)),
					resource.TestCheckResourceAttrPair(tagResourceName, names.AttrResourceID, resourceName, names.AttrID),
					resource.TestCheckResourceAttr(tagResourceName, names.AttrKey, "Name"),
					resource.TestCheckResourceAttr(tagResourceName, names.AttrValue, rName),
				),
			},
		},
	})
}

func testAccCheckVPCEndpointServiceAllowedPrincipalDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_endpoint_service_allowed_principal" {
				continue
			}

			_, err := tfec2.FindVPCEndpointServicePermission(ctx, conn, rs.Primary.Attributes["vpc_endpoint_service_id"], rs.Primary.Attributes["principal_arn"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 VPC Endpoint Service Allowed Principal %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVPCEndpointServiceAllowedPrincipalExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 VPC Endpoint Service Allowed Principal ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		_, err := tfec2.FindVPCEndpointServicePermission(ctx, conn, rs.Primary.Attributes["vpc_endpoint_service_id"], rs.Primary.Attributes["principal_arn"])

		return err
	}
}

func testAccVPCEndpointServiceAllowedPrincipalConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVPCEndpointServiceConfig_baseNetworkLoadBalancer(rName, 1), fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  network_load_balancer_arns = aws_lb.test[*].arn

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_service_allowed_principal" "test" {
  vpc_endpoint_service_id = aws_vpc_endpoint_service.test.id

  principal_arn = data.aws_iam_session_context.current.issuer_arn
}
`, rName))
}

func testAccVPCEndpointServiceAllowedPrincipalConfig_Multiple(rName string) string {
	return acctest.ConfigCompose(testAccVPCEndpointServiceConfig_baseNetworkLoadBalancer(rName, 1), fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  network_load_balancer_arns = aws_lb.test[*].arn
  allowed_principals         = ["arn:${data.aws_partition.current.partition}:iam::123456789012:root"]

  tags = {
    Name = %[1]q
  }

  lifecycle {
    ignore_changes = [
      allowed_principals
    ]
  }
}

resource "aws_vpc_endpoint_service_allowed_principal" "test" {
  vpc_endpoint_service_id = aws_vpc_endpoint_service.test.id

  principal_arn = data.aws_iam_session_context.current.issuer_arn
}
`, rName))
}

func testAccVPCEndpointServiceAllowedPrincipalConfig_tag(rName string) string {
	return acctest.ConfigCompose(testAccVPCEndpointServiceAllowedPrincipalConfig_basic(rName), fmt.Sprintf(`
resource "aws_ec2_tag" "test" {
  resource_id = aws_vpc_endpoint_service_allowed_principal.test.id

  key   = "Name"
  value = %[1]q
}
`, rName))
}
