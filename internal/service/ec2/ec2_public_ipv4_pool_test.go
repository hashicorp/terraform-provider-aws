// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2PublicIPv4Pool_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var publicipv4pool ec2.DescribePublicIpv4PoolsOutput
	resourceName := "aws_ec2_public_ipv4_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPublicIPv4PoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPublicIPv4PoolConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicIPv4PoolExists(ctx, resourceName, &publicipv4pool),
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

func TestAccEC2PublicIPv4Pool_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var publicipv4pool ec2.DescribePublicIpv4PoolsOutput
	resourceName := "aws_ec2_public_ipv4_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPublicIPv4PoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPublicIPv4PoolConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicIPv4PoolExists(ctx, resourceName, &publicipv4pool),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfec2.ResourcePublicIPv4Pool, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPublicIPv4PoolDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_public_ipv4_pool" {
				continue
			}

			_, err := conn.DescribePublicIpv4Pools(ctx, &ec2.DescribePublicIpv4PoolsInput{
				PoolIds: []string{rs.Primary.ID},
			})

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNamePublicIPv4Pool, rs.Primary.ID, err)
			}

			return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNamePublicIPv4Pool, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckPublicIPv4PoolExists(ctx context.Context, name string, publicipv4pool *ec2.DescribePublicIpv4PoolsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNamePublicIPv4Pool, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNamePublicIPv4Pool, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)
		resp, err := conn.DescribePublicIpv4Pools(ctx, &ec2.DescribePublicIpv4PoolsInput{
			PoolIds: []string{rs.Primary.ID},
		})

		if err != nil {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNamePublicIPv4Pool, rs.Primary.ID, err)
		}

		*publicipv4pool = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribePublicIpv4PoolsInput{}
	_, err := conn.DescribePublicIpv4Pools(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPublicIPv4PoolConfig_basic() string {
	return fmt.Sprintf(`
resource "aws_ec2_public_ipv4_pool" "test" {}
`)
}
