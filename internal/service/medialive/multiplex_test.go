// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package medialive_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/medialive"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfmedialive "github.com/hashicorp/terraform-provider-aws/internal/service/medialive"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccMultiplex_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var multiplex medialive.DescribeMultiplexOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_medialive_multiplex.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
			testAccMultiplexesPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiplexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiplexConfig_basic(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiplexExists(ctx, t, resourceName, &multiplex),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "medialive", "multiplex:{id}"),
					resource.TestCheckResourceAttr(resourceName, "multiplex_settings.0.transport_stream_bitrate", "1000000"),
					resource.TestCheckResourceAttr(resourceName, "multiplex_settings.0.transport_stream_reserved_bitrate", "1"),
					resource.TestCheckResourceAttr(resourceName, "multiplex_settings.0.transport_stream_id", "1"),
					resource.TestCheckResourceAttr(resourceName, "multiplex_settings.0.maximum_video_buffer_delay_milliseconds", "1000"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_multiplex"},
			},
		},
	})
}

func testAccMultiplex_start(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var multiplex medialive.DescribeMultiplexOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_medialive_multiplex.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
			testAccMultiplexesPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiplexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiplexConfig_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiplexExists(ctx, t, resourceName, &multiplex),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "medialive", "multiplex:{id}"),
				),
			},
			{
				Config: testAccMultiplexConfig_basic(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiplexExists(ctx, t, resourceName, &multiplex),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "medialive", "multiplex:{id}"),
				),
			},
		},
	})
}

func testAccMultiplex_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var multiplex medialive.DescribeMultiplexOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_medialive_multiplex.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
			testAccMultiplexesPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiplexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiplexConfig_basic(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiplexExists(ctx, t, resourceName, &multiplex),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "medialive", "multiplex:{id}"),
					resource.TestCheckResourceAttr(resourceName, "multiplex_settings.0.transport_stream_bitrate", "1000000"),
					resource.TestCheckResourceAttr(resourceName, "multiplex_settings.0.transport_stream_reserved_bitrate", "1"),
					resource.TestCheckResourceAttr(resourceName, "multiplex_settings.0.transport_stream_id", "1"),
					resource.TestCheckResourceAttr(resourceName, "multiplex_settings.0.maximum_video_buffer_delay_milliseconds", "1000"),
				),
			},
			{
				Config: testAccMultiplexConfig_update(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiplexExists(ctx, t, resourceName, &multiplex),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "medialive", "multiplex:{id}"),
					resource.TestCheckResourceAttr(resourceName, "multiplex_settings.0.transport_stream_bitrate", "1000001"),
					resource.TestCheckResourceAttr(resourceName, "multiplex_settings.0.transport_stream_reserved_bitrate", "1"),
					resource.TestCheckResourceAttr(resourceName, "multiplex_settings.0.transport_stream_id", "2"),
					resource.TestCheckResourceAttr(resourceName, "multiplex_settings.0.maximum_video_buffer_delay_milliseconds", "1000"),
				),
			},
		},
	})
}

func testAccMultiplex_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var multiplex medialive.DescribeMultiplexOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_medialive_multiplex.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
			testAccMultiplexesPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiplexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiplexConfig_basic(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiplexExists(ctx, t, resourceName, &multiplex),
					acctest.CheckSDKResourceDisappears(ctx, t, tfmedialive.ResourceMultiplex(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckMultiplexDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).MediaLiveClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_medialive_multiplex" {
				continue
			}

			_, err := tfmedialive.FindInputSecurityGroupByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return create.Error(names.MediaLive, create.ErrActionCheckingDestroyed, tfmedialive.ResNameMultiplex, rs.Primary.ID, err)
			}
		}

		return nil
	}
}

func testAccCheckMultiplexExists(ctx context.Context, t *testing.T, name string, multiplex *medialive.DescribeMultiplexOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.MediaLive, create.ErrActionCheckingExistence, tfmedialive.ResNameMultiplex, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.MediaLive, create.ErrActionCheckingExistence, tfmedialive.ResNameMultiplex, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).MediaLiveClient(ctx)

		resp, err := tfmedialive.FindMultiplexByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.MediaLive, create.ErrActionCheckingExistence, tfmedialive.ResNameMultiplex, rs.Primary.ID, err)
		}

		*multiplex = *resp

		return nil
	}
}

func testAccMultiplexesPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).MediaLiveClient(ctx)

	input := &medialive.ListMultiplexesInput{}
	_, err := conn.ListMultiplexes(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccMultiplexConfig_basic(rName string, start bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInExclude("usw2-las1-az1"),
		fmt.Sprintf(`
resource "aws_medialive_multiplex" "test" {
  name               = %[1]q
  availability_zones = [data.aws_availability_zones.available.names[0], data.aws_availability_zones.available.names[1]]

  multiplex_settings {
    transport_stream_bitrate                = 1000000
    transport_stream_id                     = 1
    transport_stream_reserved_bitrate       = 1
    maximum_video_buffer_delay_milliseconds = 1000
  }

  start_multiplex = %[2]t

  tags = {
    Name = %[1]q
  }
}
`, rName, start))
}

func testAccMultiplexConfig_update(rName string, start bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInExclude("usw2-las1-az1"),
		fmt.Sprintf(`
resource "aws_medialive_multiplex" "test" {
  name               = %[1]q
  availability_zones = [data.aws_availability_zones.available.names[0], data.aws_availability_zones.available.names[1]]

  multiplex_settings {
    transport_stream_bitrate                = 1000001
    transport_stream_id                     = 2
    transport_stream_reserved_bitrate       = 1
    maximum_video_buffer_delay_milliseconds = 1000
  }

  start_multiplex = %[2]t

  tags = {
    Name = %[1]q
  }
}
`, rName, start))
}
