// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCEndpointPrivateDNS_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var endpoint awstypes.VpcEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_endpoint_private_dns.test"
	endpointResourceName := "aws_vpc_endpoint.test"

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
				Config: testAccVPCEndpointPrivateDNSConfig_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, endpointResourceName, &endpoint),
					testAccCheckVPCEndpointPrivateDNSEnabled(ctx, endpointResourceName),
					resource.TestCheckResourceAttrPair(endpointResourceName, names.AttrID, resourceName, names.AttrVPCEndpointID),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccVPCEndpointPrivateDNSImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrVPCEndpointID,
			},
		},
	})
}
func TestAccVPCEndpointPrivateDNS_disabled(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var endpoint awstypes.VpcEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_endpoint_private_dns.test"
	endpointResourceName := "aws_vpc_endpoint.test"

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
				Config: testAccVPCEndpointPrivateDNSConfig_disabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, endpointResourceName, &endpoint),
					testAccCheckVPCEndpointPrivateDNSDisabled(ctx, endpointResourceName),
					resource.TestCheckResourceAttrPair(endpointResourceName, names.AttrID, resourceName, names.AttrVPCEndpointID),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", acctest.CtFalse),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccVPCEndpointPrivateDNSImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrVPCEndpointID,
			},
		},
	})
}

func TestAccVPCEndpointPrivateDNS_disappears_Endpoint(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var endpoint awstypes.VpcEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	endpointResourceName := "aws_vpc_endpoint.test"

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
				Config: testAccVPCEndpointPrivateDNSConfig_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, endpointResourceName, &endpoint),
					testAccCheckVPCEndpointPrivateDNSEnabled(ctx, endpointResourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPCEndpoint(), endpointResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCEndpointPrivateDNS_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var endpoint awstypes.VpcEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_endpoint_private_dns.test"
	endpointResourceName := "aws_vpc_endpoint.test"

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
				Config: testAccVPCEndpointPrivateDNSConfig_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, endpointResourceName, &endpoint),
					testAccCheckVPCEndpointPrivateDNSEnabled(ctx, endpointResourceName),
					resource.TestCheckResourceAttrPair(endpointResourceName, names.AttrID, resourceName, names.AttrVPCEndpointID),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccVPCEndpointPrivateDNSConfig_basic(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, endpointResourceName, &endpoint),
					testAccCheckVPCEndpointPrivateDNSDisabled(ctx, endpointResourceName),
					resource.TestCheckResourceAttrPair(endpointResourceName, names.AttrID, resourceName, names.AttrVPCEndpointID),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccVPCEndpointPrivateDNSConfig_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, endpointResourceName, &endpoint),
					testAccCheckVPCEndpointPrivateDNSEnabled(ctx, endpointResourceName),
					resource.TestCheckResourceAttrPair(endpointResourceName, names.AttrID, resourceName, names.AttrVPCEndpointID),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", acctest.CtTrue),
				),
			},
		},
	})
}

// testAccCheckVPCEndpointPrivateDNSEnabled verifies private DNS is enabled for a given VPC endpoint
func testAccCheckVPCEndpointPrivateDNSEnabled(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindVPCEndpointByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if aws.ToBool(output.PrivateDnsEnabled) {
			return nil
		}

		return errors.New("private DNS not enabled")
	}
}

// testAccCheckVPCEndpointPrivateDNSDisabled verifies private DNS is not enabled for a given VPC endpoint
func testAccCheckVPCEndpointPrivateDNSDisabled(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindVPCEndpointByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if aws.ToBool(output.PrivateDnsEnabled) {
			return errors.New("private DNS enabled")
		}

		return nil
	}
}

func testAccVPCEndpointPrivateDNSImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes[names.AttrVPCEndpointID], nil
	}
}

func testAccVPCEndpointPrivateDNSConfig_basic(rName string, enabled bool) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  vpc_id            = aws_vpc.test.id
  service_name      = "com.amazonaws.${data.aws_region.current.name}.ec2"
  vpc_endpoint_type = "Interface"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_private_dns" "test" {
  vpc_endpoint_id     = aws_vpc_endpoint.test.id
  private_dns_enabled = %[2]t
}
`, rName, enabled)
}

func testAccVPCEndpointPrivateDNSConfig_disabled(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  vpc_id            = aws_vpc.test.id
  service_name      = "com.amazonaws.${data.aws_region.current.name}.ec2"
  vpc_endpoint_type = "Interface"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_private_dns" "test" {
  vpc_endpoint_id     = aws_vpc_endpoint.test.id
  private_dns_enabled = false
}
`, rName)
}
