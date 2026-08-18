// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mailmanager_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/service/mailmanager"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfmailmanager "github.com/hashicorp/terraform-provider-aws/internal/service/mailmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMailManagerRelay_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_relay.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccRelayPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRelayDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRelayConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRelayExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ses", regexache.MustCompile(`mailmanager-smtp-relay/.+`)),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_timestamp"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_modified_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "server_name", "smtp.example.com"),
					resource.TestCheckResourceAttr(resourceName, "server_port", "25"),
					resource.TestCheckResourceAttr(resourceName, "authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication.0.no_authentication.#", "1"),
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

func TestAccMailManagerRelay_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_relay.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccRelayPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRelayDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRelayConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRelayExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfmailmanager.ResourceRelay, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccMailManagerRelay_update(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_relay.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccRelayPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRelayDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRelayConfig_update(rName, "smtp.example.com", 25),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRelayExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "server_name", "smtp.example.com"),
					resource.TestCheckResourceAttr(resourceName, "server_port", "25"),
				),
			},
			{
				Config: testAccRelayConfig_update(rNameUpdated, "smtp2.example.com", 587),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRelayExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "server_name", "smtp2.example.com"),
					resource.TestCheckResourceAttr(resourceName, "server_port", "587"),
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

func TestAccMailManagerRelay_Authentication_secretARN(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_relay.test"
	password := "Abcd1234!"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccRelayPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRelayDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Relay/authentication/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"password":      config.StringVariable(password),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRelayExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication.0.no_authentication.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "authentication.0.secret_arn", "aws_secretsmanager_secret_version.test", names.AttrARN),
				),
			},
		},
	})
}

func testAccCheckRelayDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).MailManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_mailmanager_relay" {
				continue
			}
			_, err := tfmailmanager.FindRelayByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return smarterr.NewError(err)
			}
			return smarterr.NewError(errors.New("not destroyed"))
		}
		return nil
	}
}

func testAccCheckRelayExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return smarterr.NewError(errors.New("not found"))
		}
		if rs.Primary.ID == "" {
			return smarterr.NewError(errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).MailManagerClient(ctx)

		_, err := tfmailmanager.FindRelayByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return smarterr.NewError(err)
		}
		return nil
	}
}

func testAccRelayPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).MailManagerClient(ctx)

	var input mailmanager.ListRelaysInput
	_, err := conn.ListRelays(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccRelayConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_mailmanager_relay" "test" {
  name        = %[1]q
  server_name = "smtp.example.com"
  server_port = 25

  authentication {
    no_authentication {}
  }
}
`, rName)
}

func testAccRelayConfig_update(rName, serverName string, serverPort int) string {
	return fmt.Sprintf(`
resource "aws_mailmanager_relay" "test" {
  name        = %[1]q
  server_name = %[2]q
  server_port = %[3]d

  authentication {
    no_authentication {}
  }
}
`, rName, serverName, serverPort)
}
