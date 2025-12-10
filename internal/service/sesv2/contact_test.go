// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccContact_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var contact sesv2.GetContactOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_contact.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContactConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, resourceName, &contact),
					resource.TestCheckResourceAttr(resourceName, "contact_list_name", rName),
					resource.TestCheckResourceAttr(resourceName, "email_address", "test@test.com"),
					resource.TestCheckResourceAttr(resourceName, "unsubscribe_all", "false"),
					resource.TestCheckResourceAttr(resourceName, "topic_preferences.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccContactConfig_basic_replaced(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, resourceName, &contact),
					resource.TestCheckResourceAttr(resourceName, "contact_list_name", rName),
					resource.TestCheckResourceAttr(resourceName, "email_address", "test1@test.com"),
					resource.TestCheckResourceAttr(resourceName, "unsubscribe_all", "false"),
					resource.TestCheckResourceAttr(resourceName, "topic_preferences.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
			{
				Config: testAccContactConfig_basic_modified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, resourceName, &contact),
					resource.TestCheckResourceAttr(resourceName, "contact_list_name", rName),
					resource.TestCheckResourceAttr(resourceName, "email_address", "test1@test.com"),
					resource.TestCheckResourceAttr(resourceName, "unsubscribe_all", "true"),
					resource.TestCheckResourceAttr(resourceName, "topic_preferences.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccContactConfig_advanced(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, resourceName, &contact),
					resource.TestCheckResourceAttr(resourceName, "contact_list_name", rName),
					resource.TestCheckResourceAttr(resourceName, "email_address", "test1@test.com"),
					resource.TestCheckResourceAttr(resourceName, "unsubscribe_all", "true"),
					resource.TestCheckResourceAttr(resourceName, "topic_preferences.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "topic_preferences.0.subscription_status", "OPT_IN"),
					resource.TestCheckResourceAttr(resourceName, "topic_preferences.0.topic_name", "test1"),
					resource.TestCheckResourceAttr(resourceName, "topic_preferences.1.subscription_status", "OPT_OUT"),
					resource.TestCheckResourceAttr(resourceName, "topic_preferences.1.topic_name", "test3"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
			{
				Config: testAccContactConfig_advanced_replaced(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, resourceName, &contact),
					resource.TestCheckResourceAttr(resourceName, "contact_list_name", rName),
					resource.TestCheckResourceAttr(resourceName, "email_address", "test1@test.com"),
					resource.TestCheckResourceAttr(resourceName, "unsubscribe_all", "true"),
					resource.TestCheckResourceAttr(resourceName, "topic_preferences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "topic_preferences.0.subscription_status", "OPT_IN"),
					resource.TestCheckResourceAttr(resourceName, "topic_preferences.0.topic_name", "test1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
			{
				Config: testAccContactConfig_advanced_modified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, resourceName, &contact),
					resource.TestCheckResourceAttr(resourceName, "contact_list_name", rName),
					resource.TestCheckResourceAttr(resourceName, "email_address", "test1@test.com"),
					resource.TestCheckResourceAttr(resourceName, "unsubscribe_all", "false"),
					resource.TestCheckResourceAttr(resourceName, "topic_preferences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "topic_preferences.0.subscription_status", "OPT_IN"),
					resource.TestCheckResourceAttr(resourceName, "topic_preferences.0.topic_name", "test1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func testAccContact_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var contact sesv2.GetContactOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_contact.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContactConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContactExists(ctx, resourceName, &contact),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfsesv2.ResourceContact, resourceName),
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

func testAccCheckContactDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sesv2_contact" {
				continue
			}

			contactList := rs.Primary.Attributes["contact_list_name"]
			email := rs.Primary.Attributes["email_address"]

			_, err := tfsesv2.FindContact(ctx, conn, contactList, email)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.SESV2, create.ErrActionCheckingDestroyed, tfsesv2.ResNameContact, fmt.Sprintf("%s %s", contactList, email), err)
			}

			return create.Error(names.SESV2, create.ErrActionCheckingDestroyed, tfsesv2.ResNameContact, fmt.Sprintf("%s %s", contactList, email), errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckContactExists(ctx context.Context, name string, contact *sesv2.GetContactOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameContact, name, errors.New("not found"))
		}

		contactList := rs.Primary.Attributes["contact_list_name"]
		if contactList == "" {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameContact, name, errors.New("not set"))
		}
		email := rs.Primary.Attributes["email_address"]
		if email == "" {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameContact, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Client(ctx)

		resp, err := tfsesv2.FindContact(ctx, conn, contactList, email)
		if err != nil {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameContact, fmt.Sprintf("%s %s", contactList, email), err)
		}

		*contact = *resp

		return nil
	}
}

func testAccContactConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_contact_list" "test" {
  contact_list_name = %[1]q

  topic {
    default_subscription_status = "OPT_OUT"
    description                 = "Test 1"
    display_name                = "Test 1"
    topic_name                  = "test1"
  }
  topic {
    default_subscription_status = "OPT_OUT"
    description                 = "Test 2"
    display_name                = "Test 2"
    topic_name                  = "test2"
  }
  topic {
    default_subscription_status = "OPT_IN"
    description                 = "Test 3"
    display_name                = "Test 3"
    topic_name                  = "test3"
  }
}
`, rName)
}

func testAccContactConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccContactConfig_base(rName),
		`resource "aws_sesv2_contact" "test" {
		  contact_list_name = aws_sesv2_contact_list.test.contact_list_name
		  email_address     = "test@test.com"
		}`,
	)
}

func testAccContactConfig_basic_replaced(rName string) string {
	return acctest.ConfigCompose(
		testAccContactConfig_base(rName),
		`resource "aws_sesv2_contact" "test" {
		  contact_list_name = aws_sesv2_contact_list.test.contact_list_name
		  email_address     = "test1@test.com"
		}`,
	)
}

func testAccContactConfig_basic_modified(rName string) string {
	return acctest.ConfigCompose(
		testAccContactConfig_base(rName),
		`resource "aws_sesv2_contact" "test" {
		  contact_list_name = aws_sesv2_contact_list.test.contact_list_name
		  email_address     = "test1@test.com"
		  unsubscribe_all   = true
		}`,
	)
}

func testAccContactConfig_advanced(rName string) string {
	return acctest.ConfigCompose(
		testAccContactConfig_base(rName),
		`resource "aws_sesv2_contact" "test" {
		  contact_list_name = aws_sesv2_contact_list.test.contact_list_name
		  email_address     = "test1@test.com"
		  unsubscribe_all   = true
		
		  topic_preferences {
			subscription_status = "OPT_IN"
			topic_name          = "test1"
		  }
		  topic_preferences {
			subscription_status = "OPT_OUT"
			topic_name          = "test3"
		  }
		}`,
	)
}

func testAccContactConfig_advanced_replaced(rName string) string {
	return acctest.ConfigCompose(
		testAccContactConfig_base(rName),
		`resource "aws_sesv2_contact" "test" {
		  contact_list_name = aws_sesv2_contact_list.test.contact_list_name
		  email_address     = "test1@test.com"
		  unsubscribe_all   = true
		
		  topic_preferences {
			subscription_status = "OPT_IN"
			topic_name          = "test1"
		  }
		}`,
	)
}

func testAccContactConfig_advanced_modified(rName string) string {
	return acctest.ConfigCompose(
		testAccContactConfig_base(rName),
		`resource "aws_sesv2_contact" "test" {
		  contact_list_name = aws_sesv2_contact_list.test.contact_list_name
		  email_address     = "test1@test.com"
		  unsubscribe_all   = false
		
		  topic_preferences {
			subscription_status = "OPT_IN"
			topic_name          = "test1"
		  }
		}`,
	)
}
