// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appintegrations_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appintegrations"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfappintegrations "github.com/hashicorp/terraform-provider-aws/internal/service/appintegrations"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppIntegrationsEventIntegration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var eventIntegration appintegrations.GetEventIntegrationOutput

	rName := acctest.RandomWithPrefix(t, "resource-test-terraform")
	originalDescription := "original description"
	updatedDescription := "updated description"
	resourceName := "aws_appintegrations_event_integration.test"

	key := "EVENT_BRIDGE_PARTNER_EVENT_SOURCE_NAME"
	var sourceName string
	sourceName = os.Getenv(key)
	if sourceName == "" {
		sourceName = "aws.partner/examplepartner.com"
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppIntegrationsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppIntegrationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventIntegrationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEventIntegrationConfig_basic(rName, originalDescription, sourceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventIntegrationExists(ctx, t, resourceName, &eventIntegration),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "app-integrations", "event-integration/{name}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, originalDescription),
					resource.TestCheckResourceAttr(resourceName, "eventbridge_bus", "default"),
					resource.TestCheckResourceAttr(resourceName, "event_filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_filter.0.source", sourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEventIntegrationConfig_basic(rName, updatedDescription, sourceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventIntegrationExists(ctx, t, resourceName, &eventIntegration),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "app-integrations", "event-integration/{name}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, updatedDescription),
					resource.TestCheckResourceAttr(resourceName, "eventbridge_bus", "default"),
					resource.TestCheckResourceAttr(resourceName, "event_filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_filter.0.source", sourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func TestAccAppIntegrationsEventIntegration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var eventIntegration appintegrations.GetEventIntegrationOutput

	rName := acctest.RandomWithPrefix(t, "resource-test-terraform")
	resourceName := "aws_appintegrations_event_integration.test"

	key := "EVENT_BRIDGE_PARTNER_EVENT_SOURCE_NAME"
	var sourceName string
	sourceName = os.Getenv(key)
	if sourceName == "" {
		sourceName = "aws.partner/examplepartner.com"
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppIntegrationsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppIntegrationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventIntegrationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEventIntegrationConfig_basic(rName, acctest.CtDisappears, sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventIntegrationExists(ctx, t, resourceName, &eventIntegration),
					acctest.CheckSDKResourceDisappears(ctx, t, tfappintegrations.ResourceEventIntegration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEventIntegrationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AppIntegrationsClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appintegrations_event_integration" {
				continue
			}

			input := &appintegrations.GetEventIntegrationInput{
				Name: aws.String(rs.Primary.ID),
			}

			resp, err := conn.GetEventIntegration(ctx, input)

			if err == nil {
				if aws.ToString(resp.Name) == rs.Primary.ID {
					return fmt.Errorf("Event Integration '%s' was not deleted properly", rs.Primary.ID)
				}
			}
		}

		return nil
	}
}

func testAccCheckEventIntegrationExists(ctx context.Context, t *testing.T, name string, eventIntegration *appintegrations.GetEventIntegrationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.ProviderMeta(ctx, t).AppIntegrationsClient(ctx)
		input := &appintegrations.GetEventIntegrationInput{
			Name: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetEventIntegration(ctx, input)

		if err != nil {
			return err
		}

		*eventIntegration = *resp

		return nil
	}
}

func testAccEventIntegrationConfig_basic(rName, label, sourceName string) string {
	return fmt.Sprintf(`
resource "aws_appintegrations_event_integration" "test" {
  name            = %[1]q
  description     = %[2]q
  eventbridge_bus = "default"

  event_filter {
    source = %[3]q
  }
}
`, rName, label, sourceName)
}
