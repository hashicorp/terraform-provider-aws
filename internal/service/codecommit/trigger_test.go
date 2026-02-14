// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package codecommit_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcodecommit "github.com/hashicorp/terraform-provider-aws/internal/service/codecommit"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCodeCommitTrigger_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_codecommit_trigger.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeCommitServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTriggerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTriggerExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "trigger.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.branches.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.events.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.events.0", "all"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.name", rName),
				),
			},
		},
	})
}

func TestAccCodeCommitTrigger_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_codecommit_trigger.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeCommitServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTriggerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfcodecommit.ResourceTrigger(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCodeCommitTrigger_branches(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_codecommit_trigger.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeCommitServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTriggerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_branches(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.branches.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.branches.0", "main"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.branches.1", "develop"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.events.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.events.0", "updateReference"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.events.1", "createReference"),
					resource.TestCheckResourceAttr(resourceName, "trigger.0.name", rName),
				),
			},
		},
	})
}

func testAccCheckTriggerDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CodeCommitClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codecommit_trigger" {
				continue
			}

			_, err := tfcodecommit.FindRepositoryTriggersByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CodeCommit Trigger (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTriggerExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CodeCommitClient(ctx)

		_, err := tfcodecommit.FindRepositoryTriggersByName(ctx, conn, rs.Primary.ID)

		return err
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

func testAccTriggerConfig_branches(rName string) string {
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
    events          = ["updateReference", "createReference"]
    destination_arn = aws_sns_topic.test.arn
    branches        = ["main", "develop"]
  }
}
`, rName)
}
