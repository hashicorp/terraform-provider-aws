package medialive_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/medialive"
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

func TestAccMediaLiveInput_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var input medialive.DescribeInputOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_medialive_input.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.MediaLiveEndpointID, t)
			testAccInputsPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInputDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInputConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInputExists(resourceName, &input),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccMediaLiveInput_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var input medialive.DescribeInputOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_medialive_input.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.MediaLiveEndpointID, t)
			testAccInputsPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInputDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInputConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInputExists(resourceName, &input),
					acctest.CheckResourceDisappears(acctest.Provider, tfmedialive.ResourceInput(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckInputDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MediaLiveConn
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_medialive_input" {
			continue
		}

		_, err := tfmedialive.FindInputByID(ctx, conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return create.Error(names.MediaLive, create.ErrActionCheckingDestroyed, tfmedialive.ResNameInput, rs.Primary.ID, err)
		}
	}

	return nil
}

func testAccCheckInputExists(name string, input *medialive.DescribeInputOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.MediaLive, create.ErrActionCheckingExistence, tfmedialive.ResNameInput, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.MediaLive, create.ErrActionCheckingExistence, tfmedialive.ResNameInput, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaLiveConn
		ctx := context.Background()
		resp, err := tfmedialive.FindInputByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.MediaLive, create.ErrActionCheckingExistence, tfmedialive.ResNameInput, rs.Primary.ID, err)
		}

		*input = *resp

		return nil
	}
}

func testAccInputsPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MediaLiveConn
	ctx := context.Background()

	input := &medialive.ListInputsInput{}
	_, err := conn.ListInputs(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccInputBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "medialive.amazonaws.com"
        }
      },
    ]
  })

  tags = {
    Name = %[1]q
  }
}

resource "aws_medialive_input_security_group" "test" {
  whitelist_rules {
    cidr = "10.0.0.8/32"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccInputConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccInputBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_medialive_input" "test" {
  name                  = %[1]q
  input_security_groups = [aws_medialive_input_security_group.test.id]
  type                  = "UDP_PUSH"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}
