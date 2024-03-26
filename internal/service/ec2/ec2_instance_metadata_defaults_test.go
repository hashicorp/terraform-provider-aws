// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccInstanceMetadataDefaultsConfig_basic() string {
	return fmt.Sprintf(`
resource "aws_ec2_instance_metadata_defaults" "test" {
  http_tokens                 = "required" # non-default
  instance_metadata_tags      = "disabled"
  http_endpoint               = "enabled"
  http_put_response_hop_limit = 1
}
`)
}

func testAccInstanceMetadataDefaultsConfig_partial() string {
	return fmt.Sprintf(`
resource "aws_ec2_instance_metadata_defaults" "test-partial" {
  http_tokens                 = "required" # non-default
  http_put_response_hop_limit = 2          # non-default
}
`)
}

func TestAccEC2InstanceMetadataDefaults_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_ec2_instance_metadata_defaults.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, "ec2")
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceMetadataDefaultsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceMetadataDefaultsConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
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

func TestAccEC2InstanceMetadataDefaults_partial(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_ec2_instance_metadata_defaults.test-partial"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, "ec2")
			testAccPreCheck(ctx, t)
		},
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
		result, err := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx).GetInstanceMetadataDefaults(ctx, &ec2.GetInstanceMetadataDefaultsInput{})
		if err != nil {
			return fmt.Errorf("unable to describe instance metadata defaults: %w", err)
		}
		if string(result.AccountLevel.HttpEndpoint) != "" || string(result.AccountLevel.HttpTokens) != "" || result.AccountLevel.HttpPutResponseHopLimit != nil || string(result.AccountLevel.InstanceMetadataTags) != "" {
			return errors.New("expected instance metadata defaults to be reset")
		}
		return nil
	}
}

func testAccCheckInstanceMetadataDefaultsExists(ctx context.Context, name string, expectations *awstypes.InstanceMetadataDefaultsResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameEC2InstanceMetadataDefaults, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameEC2InstanceMetadataDefaults, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)
		resp, err := conn.GetInstanceMetadataDefaults(ctx, &ec2.GetInstanceMetadataDefaultsInput{})

		if err != nil {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameEC2InstanceMetadataDefaults, rs.Primary.ID, err)
		}

		// check assertions
		if resp.AccountLevel.HttpTokens != expectations.HttpTokens {
			return fmt.Errorf("expected HttpTokens to be '%s', got '%s'", expectations.HttpTokens, resp.AccountLevel.HttpTokens)
		}
		if *resp.AccountLevel.HttpPutResponseHopLimit != *expectations.HttpPutResponseHopLimit {
			return fmt.Errorf("expected HttpPutResponseHopLimit to be '%d', got '%d'", *expectations.HttpPutResponseHopLimit, *resp.AccountLevel.HttpPutResponseHopLimit)
		}
		if resp.AccountLevel.HttpEndpoint != expectations.HttpEndpoint {
			return fmt.Errorf("expected HttpEndpoint to be '%s', got '%s'", expectations.HttpEndpoint, resp.AccountLevel.HttpEndpoint)
		}
		if resp.AccountLevel.InstanceMetadataTags != expectations.InstanceMetadataTags {
			return fmt.Errorf("expected InstanceMetadataTags to be '%s', got '%s'", expectations.InstanceMetadataTags, resp.AccountLevel.InstanceMetadataTags)
		}

		return nil
	}
}
func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	_, err := conn.GetInstanceMetadataDefaults(ctx, &ec2.GetInstanceMetadataDefaultsInput{})

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
