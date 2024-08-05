// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appintegrations_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appintegrations"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappintegrations "github.com/hashicorp/terraform-provider-aws/internal/service/appintegrations"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppIntegrationsEventIntegration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var eventIntegration appintegrations.GetEventIntegrationOutput

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	originalDescription := "original description"
	updatedDescription := "updated description"
	resourceName := "aws_appintegrations_event_integration.test"

	key := "EVENT_BRIDGE_PARTNER_EVENT_SOURCE_NAME"
	var sourceName string
	sourceName = os.Getenv(key)
	if sourceName == "" {
		sourceName = "aws.partner/examplepartner.com"
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppIntegrationsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppIntegrationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventIntegrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventIntegrationConfig_basic(rName, originalDescription, sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventIntegrationExists(ctx, resourceName, &eventIntegration),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, originalDescription),
					resource.TestCheckResourceAttr(resourceName, "eventbridge_bus", "default"),
					resource.TestCheckResourceAttr(resourceName, "event_filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_filter.0.source", sourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Event Integration"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEventIntegrationConfig_basic(rName, updatedDescription, sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventIntegrationExists(ctx, resourceName, &eventIntegration),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, updatedDescription),
					resource.TestCheckResourceAttr(resourceName, "eventbridge_bus", "default"),
					resource.TestCheckResourceAttr(resourceName, "event_filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_filter.0.source", sourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Event Integration"),
				),
			},
		},
	})
}

func TestAccAppIntegrationsEventIntegration_updateTags(t *testing.T) {
	ctx := acctest.Context(t)
	var eventIntegration appintegrations.GetEventIntegrationOutput

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	description := "example description"
	resourceName := "aws_appintegrations_event_integration.test"

	key := "EVENT_BRIDGE_PARTNER_EVENT_SOURCE_NAME"
	var sourceName string
	sourceName = os.Getenv(key)
	if sourceName == "" {
		sourceName = "aws.partner/examplepartner.com"
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppIntegrationsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppIntegrationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventIntegrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventIntegrationConfig_basic(rName, description, sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventIntegrationExists(ctx, resourceName, &eventIntegration),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "eventbridge_bus", "default"),
					resource.TestCheckResourceAttr(resourceName, "event_filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_filter.0.source", sourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Event Integration"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEventIntegrationConfig_tags(rName, description, sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventIntegrationExists(ctx, resourceName, &eventIntegration),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "eventbridge_bus", "default"),
					resource.TestCheckResourceAttr(resourceName, "event_filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_filter.0.source", sourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Event Integration"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEventIntegrationConfig_tagsUpdated(rName, description, sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventIntegrationExists(ctx, resourceName, &eventIntegration),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "eventbridge_bus", "default"),
					resource.TestCheckResourceAttr(resourceName, "event_filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_filter.0.source", sourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Event Integration"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
				),
			},
		},
	})
}

func TestAccAppIntegrationsEventIntegration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var eventIntegration appintegrations.GetEventIntegrationOutput

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_appintegrations_event_integration.test"

	key := "EVENT_BRIDGE_PARTNER_EVENT_SOURCE_NAME"
	var sourceName string
	sourceName = os.Getenv(key)
	if sourceName == "" {
		sourceName = "aws.partner/examplepartner.com"
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppIntegrationsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppIntegrationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventIntegrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventIntegrationConfig_basic(rName, acctest.CtDisappears, sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventIntegrationExists(ctx, resourceName, &eventIntegration),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappintegrations.ResourceEventIntegration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEventIntegrationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppIntegrationsClient(ctx)
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

func testAccCheckEventIntegrationExists(ctx context.Context, name string, eventIntegration *appintegrations.GetEventIntegrationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppIntegrationsClient(ctx)
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

  tags = {
    "Name" = "Test Event Integration"
  }
}
`, rName, label, sourceName)
}

func testAccEventIntegrationConfig_tags(rName, label, sourceName string) string {
	return fmt.Sprintf(`
resource "aws_appintegrations_event_integration" "test" {
  name            = %[1]q
  description     = %[2]q
  eventbridge_bus = "default"

  event_filter {
    source = %[3]q
  }

  tags = {
    "Name" = "Test Event Integration"
    "Key2" = "Value2a"
  }
}
`, rName, label, sourceName)
}

func testAccEventIntegrationConfig_tagsUpdated(rName, label, sourceName string) string {
	return fmt.Sprintf(`
resource "aws_appintegrations_event_integration" "test" {
  name            = %[1]q
  description     = %[2]q
  eventbridge_bus = "default"

  event_filter {
    source = %[3]q
  }

  tags = {
    "Name" = "Test Event Integration"
    "Key2" = "Value2b"
    "Key3" = "Value3"
  }
}
`, rName, label, sourceName)
}
