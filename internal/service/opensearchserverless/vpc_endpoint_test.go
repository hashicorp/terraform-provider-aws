// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfopensearchserverless "github.com/hashicorp/terraform-provider-aws/internal/service/opensearchserverless"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchServerlessVPCEndpoint_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	ctx := acctest.Context(t)
	var vpcendpoint types.VpcEndpointDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_vpc_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckVPCEndpoint(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, t, resourceName, &vpcendpoint),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
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

func TestAccOpenSearchServerlessVPCEndpoint_securityGroups(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	ctx := acctest.Context(t)
	var vpcendpoint1, vpcendpoint2, vpcendpoint3 types.VpcEndpointDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_vpc_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckVPCEndpoint(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, t, resourceName, &vpcendpoint1),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
				),
			},
			{
				Config: testAccVPCEndpointConfig_multiple_securityGroups(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, t, resourceName, &vpcendpoint2),
					testAccCheckVPCEndpointNotRecreated(&vpcendpoint1, &vpcendpoint2),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
				),
			},
			{
				Config: testAccVPCEndpointConfig_single_securityGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, t, resourceName, &vpcendpoint3),
					testAccCheckVPCEndpointNotRecreated(&vpcendpoint1, &vpcendpoint3),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccOpenSearchServerlessVPCEndpoint_update(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	ctx := acctest.Context(t)
	var vpcendpoint1, vpcendpoint2, vpcendpoint3 types.VpcEndpointDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_vpc_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckVPCEndpoint(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, t, resourceName, &vpcendpoint1),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
				),
			},
			{
				Config: testAccVPCEndpointConfig_multiple_subnets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, t, resourceName, &vpcendpoint2),
					testAccCheckVPCEndpointNotRecreated(&vpcendpoint1, &vpcendpoint2),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
				),
			},
			{
				Config: testAccVPCEndpointConfig_multiple_securityGroups(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, t, resourceName, &vpcendpoint3),
					testAccCheckVPCEndpointNotRecreated(&vpcendpoint2, &vpcendpoint3),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
				),
			},
			{
				Config: testAccVPCEndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, t, resourceName, &vpcendpoint3),
					testAccCheckVPCEndpointNotRecreated(&vpcendpoint2, &vpcendpoint3),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
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

func TestAccOpenSearchServerlessVPCEndpoint_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	ctx := acctest.Context(t)
	var vpcendpoint types.VpcEndpointDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_vpc_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckVPCEndpoint(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, t, resourceName, &vpcendpoint),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfopensearchserverless.ResourceVPCEndpoint, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVPCEndpointDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).OpenSearchServerlessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_opensearchserverless_vpc_endpoint" {
				continue
			}

			_, err := tfopensearchserverless.FindVPCEndpointByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("OpenSearch Serverless VPC Endpoint %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVPCEndpointExists(ctx context.Context, t *testing.T, n string, v *types.VpcEndpointDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).OpenSearchServerlessClient(ctx)

		output, err := tfopensearchserverless.FindVPCEndpointByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheckVPCEndpoint(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).OpenSearchServerlessClient(ctx)

	input := &opensearchserverless.ListVpcEndpointsInput{}
	_, err := conn.ListVpcEndpoints(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckVPCEndpointNotRecreated(before, after *types.VpcEndpointDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.Id), aws.ToString(after.Id); before != after {
			return fmt.Errorf("OpenSearch Serverless VPC Endpoint %s recreated", before)
		}

		return nil
	}
}

func testAccVPCEndpointConfig_networkingBase(rName string, subnetCount int) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = %[2]d

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)

  tags = {
    Name = %[1]q
  }
}
`, rName, subnetCount),
	)
}

func testAccVPCEndpointConfig_securityGroupBase(rName string, sgCount int) string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  count  = %[2]d
  name   = "%[1]s-${count.index}"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, sgCount),
	)
}

func testAccVPCEndpointConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccVPCEndpointConfig_networkingBase(rName, 2),
		testAccVPCEndpointConfig_securityGroupBase(rName, 2),
		fmt.Sprintf(`
resource "aws_opensearchserverless_vpc_endpoint" "test" {
  name               = %[1]q
  subnet_ids         = [aws_subnet.test[0].id]
  vpc_id             = aws_vpc.test.id
  security_group_ids = [aws_security_group.test[0].id]
}
`, rName))
}

func testAccVPCEndpointConfig_multiple_subnets(rName string) string {
	return acctest.ConfigCompose(
		testAccVPCEndpointConfig_networkingBase(rName, 2),
		testAccVPCEndpointConfig_securityGroupBase(rName, 1),
		fmt.Sprintf(`
resource "aws_opensearchserverless_vpc_endpoint" "test" {
  name               = %[1]q
  subnet_ids         = aws_subnet.test[*].id
  vpc_id             = aws_vpc.test.id
  security_group_ids = [aws_security_group.test[0].id]
}
`, rName))
}

func testAccVPCEndpointConfig_multiple_securityGroups(rName string) string {
	return acctest.ConfigCompose(
		testAccVPCEndpointConfig_networkingBase(rName, 2),
		testAccVPCEndpointConfig_securityGroupBase(rName, 2),
		fmt.Sprintf(`
resource "aws_opensearchserverless_vpc_endpoint" "test" {
  name               = %[1]q
  subnet_ids         = aws_subnet.test[*].id
  vpc_id             = aws_vpc.test.id
  security_group_ids = aws_security_group.test[*].id
}
`, rName))
}

func testAccVPCEndpointConfig_single_securityGroup(rName string) string {
	return acctest.ConfigCompose(
		testAccVPCEndpointConfig_networkingBase(rName, 2),
		testAccVPCEndpointConfig_securityGroupBase(rName, 2),
		fmt.Sprintf(`
resource "aws_opensearchserverless_vpc_endpoint" "test" {
  name               = %[1]q
  subnet_ids         = aws_subnet.test[*].id
  vpc_id             = aws_vpc.test.id
  security_group_ids = [aws_security_group.test[0].id]
}
`, rName))
}
