// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codecommit_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccCodeCommitTrigger_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codecommit_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, codecommit.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTriggerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "trigger.#", "1"),
				),
			},
		},
	})
}

func testAccCheckTriggerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeCommitConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codecommit_trigger" {
				continue
			}

			_, err := conn.GetRepositoryTriggersWithContext(ctx, &codecommit.GetRepositoryTriggersInput{
				RepositoryName: aws.String(rs.Primary.ID),
			})

			if tfawserr.ErrCodeEquals(err, codecommit.ErrCodeRepositoryDoesNotExistException) {
				continue
			}

			if err == nil {
				return fmt.Errorf("Trigger still exists: %s", rs.Primary.ID)
			}
			return err
		}

		return nil
	}
}

func testAccCheckTriggerExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeCommitConn(ctx)
		out, err := conn.GetRepositoryTriggersWithContext(ctx, &codecommit.GetRepositoryTriggersInput{
			RepositoryName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if len(out.Triggers) == 0 {
			return fmt.Errorf("CodeCommit Trigger Failed: %q", out)
		}

		return nil
	}
}

func testAccTriggerConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_codecommit_repository" "test" {
  repository_name = %[1]q
}

resource "aws_codecommit_trigger" "test" {
  repository_name = aws_codecommit_repository.test.id

  trigger {
    name            = %[1]q
    events          = ["all"]
    destination_arn = aws_sns_topic.test.arn
  }
}
`, rName)
}
