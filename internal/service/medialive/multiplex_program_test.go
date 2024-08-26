// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package medialive_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/medialive"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfmedialive "github.com/hashicorp/terraform-provider-aws/internal/service/medialive"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestParseMultiplexProgramIDUnitTest(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		Input       string
		ProgramName string
		MultiplexID string
		Error       bool
	}{
		{
			TestName:    "valid id",
			Input:       "program_name/multiplex_id",
			ProgramName: "program_name",
			MultiplexID: "multiplex_id",
			Error:       false,
		},
		{
			TestName:    "invalid id",
			Input:       "multiplex_id",
			ProgramName: "",
			MultiplexID: "",
			Error:       true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			pn, mid, err := tfmedialive.ParseMultiplexProgramID(testCase.Input)

			if err != nil && !testCase.Error {
				t.Errorf("got error (%s), expected no error", err)
			}

			if err == nil && testCase.Error {
				t.Errorf("got (%s, %s) and no error, expected error", pn, mid)
			}

			if pn != testCase.ProgramName {
				t.Errorf("got %s, expected %s", pn, testCase.ProgramName)
			}

			if pn != testCase.ProgramName {
				t.Errorf("got %s, expected %s", mid, testCase.MultiplexID)
			}
		})
	}
}

func testAccMultiplexProgram_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var multiplexprogram medialive.DescribeMultiplexProgramOutput
	rName := fmt.Sprintf("tf_acc_%s", sdkacctest.RandString(8))
	resourceName := "aws_medialive_multiplex_program.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiplexProgramDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiplexProgramConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiplexProgramExists(ctx, resourceName, &multiplexprogram),
					resource.TestCheckResourceAttr(resourceName, "program_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "multiplex_id"),
					resource.TestCheckResourceAttr(resourceName, "multiplex_program_settings.0.program_number", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "multiplex_program_settings.0.preferred_channel_pipeline", "CURRENTLY_ACTIVE"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"multiplex_id"},
			},
		},
	})
}

func testAccMultiplexProgram_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var multiplexprogram medialive.DescribeMultiplexProgramOutput
	rName := fmt.Sprintf("tf_acc_%s", sdkacctest.RandString(8))
	resourceName := "aws_medialive_multiplex_program.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiplexProgramDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiplexProgramConfig_update(rName, 100000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiplexProgramExists(ctx, resourceName, &multiplexprogram),
					resource.TestCheckResourceAttr(resourceName, "program_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "multiplex_id"),
					resource.TestCheckResourceAttr(resourceName, "multiplex_program_settings.0.program_number", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "multiplex_program_settings.0.preferred_channel_pipeline", "CURRENTLY_ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "multiplex_program_settings.0.video_settings.0.statmux_settings.0.minimum_bitrate", "100000"),
				),
			},
			{
				Config: testAccMultiplexProgramConfig_update(rName, 100001),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiplexProgramExists(ctx, resourceName, &multiplexprogram),
					resource.TestCheckResourceAttr(resourceName, "program_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "multiplex_id"),
					resource.TestCheckResourceAttr(resourceName, "multiplex_program_settings.0.program_number", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "multiplex_program_settings.0.preferred_channel_pipeline", "CURRENTLY_ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "multiplex_program_settings.0.video_settings.0.statmux_settings.0.minimum_bitrate", "100001"),
				),
			},
		},
	})
}

func testAccMultiplexProgram_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var multiplexprogram medialive.DescribeMultiplexProgramOutput
	rName := fmt.Sprintf("tf_acc_%s", sdkacctest.RandString(8))
	resourceName := "aws_medialive_multiplex_program.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiplexProgramDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiplexProgramConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiplexProgramExists(ctx, resourceName, &multiplexprogram),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfmedialive.ResourceMultiplexProgram, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckMultiplexProgramDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaLiveClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_medialive_multiplex_program" {
				continue
			}

			attributes := rs.Primary.Attributes

			_, err := tfmedialive.FindMultiplexProgramByID(ctx, conn, attributes["multiplex_id"], attributes["program_name"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return create.Error(names.MediaLive, create.ErrActionCheckingDestroyed, tfmedialive.ResNameMultiplexProgram, rs.Primary.ID, err)
			}
		}

		return nil
	}
}

func testAccCheckMultiplexProgramExists(ctx context.Context, name string, multiplexprogram *medialive.DescribeMultiplexProgramOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.MediaLive, create.ErrActionCheckingExistence, tfmedialive.ResNameMultiplexProgram, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.MediaLive, create.ErrActionCheckingExistence, tfmedialive.ResNameMultiplexProgram, name, errors.New("not set"))
		}

		programName, multiplexId, err := tfmedialive.ParseMultiplexProgramID(rs.Primary.ID)

		if err != nil {
			return create.Error(names.MediaLive, create.ErrActionCheckingExistence, tfmedialive.ResNameMultiplexProgram, rs.Primary.ID, err)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaLiveClient(ctx)

		resp, err := tfmedialive.FindMultiplexProgramByID(ctx, conn, multiplexId, programName)

		if err != nil {
			return create.Error(names.MediaLive, create.ErrActionCheckingExistence, tfmedialive.ResNameMultiplexProgram, rs.Primary.ID, err)
		}

		*multiplexprogram = *resp

		return nil
	}
}

func testAccMultiplexProgramBaseConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
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

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccMultiplexProgramConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccMultiplexProgramBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_medialive_multiplex_program" "test" {
  program_name = %[1]q
  multiplex_id = aws_medialive_multiplex.test.id

  multiplex_program_settings {
    program_number             = 1
    preferred_channel_pipeline = "CURRENTLY_ACTIVE"

    video_settings {
      constant_bitrate = 100000
    }
  }
}
`, rName))
}

func testAccMultiplexProgramConfig_update(rName string, minBitrate int) string {
	return acctest.ConfigCompose(
		testAccMultiplexProgramBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_medialive_multiplex_program" "test" {
  program_name = %[1]q
  multiplex_id = aws_medialive_multiplex.test.id

  multiplex_program_settings {
    program_number             = 1
    preferred_channel_pipeline = "CURRENTLY_ACTIVE"

    video_settings {
      statmux_settings {
        minimum_bitrate = %[2]d
      }
    }
  }
}
`, rName, minBitrate))
}
