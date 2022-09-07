package medialive_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/medialive"
	"github.com/aws/aws-sdk-go/aws"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfmedialive "github.com/hashicorp/terraform-provider-aws/internal/service/medialive"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccMultiplexProgram_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var multiplexprogram medialive.DescribeMultiplexProgramOutput
	rName := fmt.Sprintf("tf_acc_%s", sdkacctest.RandString(8))
	resourceName := "aws_medialive_multiplex_program.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.MediaLiveEndpointID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiplexProgramDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMultiplexProgramConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiplexProgramExists(resourceName, &multiplexprogram),
					resource.TestCheckResourceAttr(resourceName, "program_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "multiplex_id"),
					resource.TestCheckResourceAttr(resourceName, "multiplex_program_settings.0.program_number", "1"),
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

func testAccMultiplexProgram_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var multiplexprogram medialive.DescribeMultiplexProgramOutput
	rName := fmt.Sprintf("tf_acc_%s", sdkacctest.RandString(8))
	resourceName := "aws_medialive_multiplex_program.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.MediaLiveEndpointID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiplexProgramDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMultiplexProgramConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiplexProgramExists(resourceName, &multiplexprogram),
					testAccCheckMultiplexProgramDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckMultiplexProgramDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID missing: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaLiveConn
		ctx := context.Background()
		programName, multiplexId, err := tfmedialive.ParseMultiplexProgramID(rs.Primary.ID)

		if err != nil {
			return create.Error(names.MediaLive, create.ErrActionCheckingExistence, tfmedialive.ResNameMultiplexProgram, rs.Primary.ID, err)
		}

		_, err = conn.DeleteMultiplexProgram(ctx, &medialive.DeleteMultiplexProgramInput{
			MultiplexId: aws.String(multiplexId),
			ProgramName: aws.String(programName),
		})

		if err != nil {
			return create.Error(names.MediaLive, "checking disappears", tfmedialive.ResNameMultiplexProgram, rs.Primary.ID, err)
		}

		return nil
	}
}
func testAccCheckMultiplexProgramDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MediaLiveConn
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_medialive_multiplex_program" {
			continue
		}

		attributes := rs.Primary.Attributes

		_, err := tfmedialive.FindMultipleProgramByID(ctx, conn, attributes["multiplex_id"], attributes["program_name"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return create.Error(names.MediaLive, create.ErrActionCheckingDestroyed, tfmedialive.ResNameMultiplexProgram, rs.Primary.ID, err)
		}
	}

	return nil
}

func testAccCheckMultiplexProgramExists(name string, multiplexprogram *medialive.DescribeMultiplexProgramOutput) resource.TestCheckFunc {
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaLiveConn
		ctx := context.Background()
		resp, err := tfmedialive.FindMultipleProgramByID(ctx, conn, multiplexId, programName)

		if err != nil {
			return create.Error(names.MediaLive, create.ErrActionCheckingExistence, tfmedialive.ResNameMultiplexProgram, rs.Primary.ID, err)
		}

		*multiplexprogram = *resp

		return nil
	}
}

func testAccMultiplexProgramBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "test" {
  state = "available"
}

resource "aws_medialive_multiplex" "test" {
  name               = %[1]q
  availability_zones = [data.aws_availability_zones.test.names[0], data.aws_availability_zones.test.names[1]]

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
`, rName)
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
