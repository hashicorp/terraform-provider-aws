// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2InstanceMetadataDefaults_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		"basic":  testAccInstanceMetadataDefaults_basic,
		"update": testAccInstanceMetadataDefaults_update,
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
				Config: testAccInstanceMetadataDefaultsConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceMetadataDefaultsExists(ctx, resourceName, &awstypes.InstanceMetadataDefaultsResponse{
						HttpTokens:              awstypes.HttpTokensState(awstypes.MetadataDefaultHttpTokensStateRequired),
						HttpPutResponseHopLimit: aws.Int32(1),
						HttpEndpoint:            awstypes.InstanceMetadataEndpointStateEnabled,
						InstanceMetadataTags:    awstypes.InstanceMetadataTagsStateDisabled,
					}),
					resource.TestCheckResourceAttr(resourceName, "http_tokens", "required"),
					resource.TestCheckResourceAttr(resourceName, "instance_metadata_tags", "disabled"),
					resource.TestCheckResourceAttr(resourceName, "http_endpoint", "enabled"),
					resource.TestCheckResourceAttr(resourceName, "http_put_response_hop_limit", "1"),
				),
			},
		},
	})
}

func testAccInstanceMetadataDefaults_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_instance_metadata_defaults.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceMetadataDefaultsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceMetadataDefaultsConfig_partial(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceMetadataDefaultsExists(ctx, resourceName, &awstypes.InstanceMetadataDefaultsResponse{
						HttpTokens:              awstypes.HttpTokensState(awstypes.MetadataDefaultHttpTokensStateRequired),
						HttpPutResponseHopLimit: aws.Int32(2),
						HttpEndpoint:            "",
						InstanceMetadataTags:    "",
					}),
					resource.TestCheckResourceAttr(resourceName, "http_tokens", "required"),
					resource.TestCheckResourceAttr(resourceName, "http_put_response_hop_limit", "2"),
				),
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

			_, err := tfec2.FindInstanceMetadataDefaults(ctx, conn)

			if tfresource.NotFound(err) {
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

func testAccCheckInstanceMetadataDefaultsExists(ctx context.Context, n string, v *awstypes.InstanceMetadataDefaultsResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindInstanceMetadataDefaults(ctx, conn)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccInstanceMetadataDefaultsConfig_basic() string {
	return `
resource "aws_ec2_instance_metadata_defaults" "test" {
  http_tokens                 = "required" # non-default
  instance_metadata_tags      = "disabled"
  http_endpoint               = "enabled"
  http_put_response_hop_limit = 1
}
`
}

func testAccInstanceMetadataDefaultsConfig_partial() string {
	return `
resource "aws_ec2_instance_metadata_defaults" "test" {
  http_tokens                 = "required" # non-default
  http_put_response_hop_limit = 2          # non-default
}
`
}
