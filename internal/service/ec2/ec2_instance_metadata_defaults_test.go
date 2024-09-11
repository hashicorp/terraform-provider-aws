// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2InstanceMetadataDefaults_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccInstanceMetadataDefaults_basic,
		acctest.CtDisappears: testAccInstanceMetadataDefaults_disappears,
		"empty":              testAccInstanceMetadataDefaults_empty,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccInstanceMetadataDefaults_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_instance_metadata_defaults.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceMetadataDefaultsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceMetadataDefaultsConfig_full,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceMetadataDefaultsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "http_endpoint", names.AttrEnabled),
					resource.TestCheckResourceAttr(resourceName, "http_put_response_hop_limit", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "http_tokens", "required"),
					resource.TestCheckResourceAttr(resourceName, "instance_metadata_tags", "disabled"),
				),
			},
			{
				Config: testAccInstanceMetadataDefaultsConfig_partial,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceMetadataDefaultsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "http_endpoint", "no-preference"),
					resource.TestCheckResourceAttr(resourceName, "http_put_response_hop_limit", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "http_tokens", "required"),
					resource.TestCheckResourceAttr(resourceName, "instance_metadata_tags", "no-preference"),
				),
			},
		},
	})
}

func testAccInstanceMetadataDefaults_empty(t *testing.T) {
	ctx := acctest.Context(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceMetadataDefaultsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccInstanceMetadataDefaultsConfig_empty,
				ExpectError: regexache.MustCompile(`At least one of these attributes must be configured`),
			},
		},
	})
}

func testAccInstanceMetadataDefaults_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_instance_metadata_defaults.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceMetadataDefaultsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceMetadataDefaultsConfig_full,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceMetadataDefaultsExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfec2.ResourceInstanceMetadataDefaults, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckInstanceMetadataDefaultsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_instance_metadata_defaults" {
				continue
			}

			output, err := tfec2.FindInstanceMetadataDefaults(ctx, conn)

			if tfresource.NotFound(err) || err == nil && itypes.IsZero(output) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Instance Metadata Defaults %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckInstanceMetadataDefaultsExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		_, err := tfec2.FindInstanceMetadataDefaults(ctx, conn)

		return err
	}
}

const (
	testAccInstanceMetadataDefaultsConfig_empty = `
resource "aws_ec2_instance_metadata_defaults" "test" {}
`

	testAccInstanceMetadataDefaultsConfig_full = `
resource "aws_ec2_instance_metadata_defaults" "test" {
  http_tokens                 = "required" # non-default
  instance_metadata_tags      = "disabled"
  http_endpoint               = "enabled"
  http_put_response_hop_limit = 1
}
`
	testAccInstanceMetadataDefaultsConfig_partial = `
resource "aws_ec2_instance_metadata_defaults" "test" {
  http_tokens                 = "required" # non-default
  http_put_response_hop_limit = 2          # non-default
}
`
)
