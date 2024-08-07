// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppStreamStack_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var stackOutput awstypes.Stack
	resourceName := "aws_appstream_stack.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, "access_endpoints.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.settings_group", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, ""),
					resource.TestCheckResourceAttr(resourceName, "embed_host_domains.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "feedback_url", ""),
					resource.TestCheckResourceAttr(resourceName, "redirect_url", ""),
					resource.TestCheckResourceAttr(resourceName, "storage_connectors.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "user_settings.#", "7"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct0),
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

func TestAccAppStreamStack_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var stackOutput awstypes.Stack
	resourceName := "aws_appstream_stack.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stackOutput),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappstream.ResourceStack(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppStreamStack_complete(t *testing.T) {
	ctx := acctest.Context(t)
	var stackOutput awstypes.Stack
	resourceName := "aws_appstream_stack.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "Description of a test"
	descriptionUpdated := "Updated Description of a test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_complete(rName, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "embed_host_domains.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "embed_host_domains.*", "example.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "embed_host_domains.*", "subdomain.example.com"),
					resource.TestCheckResourceAttr(resourceName, "access_endpoints.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.settings_group", "SettingsGroup"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, ""),
					resource.TestCheckResourceAttr(resourceName, "feedback_url", ""),
					resource.TestCheckResourceAttr(resourceName, "redirect_url", ""),
					resource.TestCheckResourceAttr(resourceName, "storage_connectors.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_settings.#", "7"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccStackConfig_complete(rName, descriptionUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionUpdated),
					resource.TestCheckResourceAttr(resourceName, "embed_host_domains.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "embed_host_domains.*", "example.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "embed_host_domains.*", "subdomain.example.com"),
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

func TestAccAppStreamStack_applicationSettings_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var stackOutput awstypes.Stack
	resourceName := "aws_appstream_stack.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	settingsGroup := "group"
	settingsGroupUpdated := "group-updated"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_applicationSettings(rName, true, settingsGroup),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "application_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.settings_group", settingsGroup),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStackConfig_applicationSettings(rName, true, settingsGroupUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "application_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.settings_group", settingsGroupUpdated),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStackConfig_applicationSettings(rName, false, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "application_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.settings_group", ""),
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

func TestAccAppStreamStack_applicationSettings_removeFromEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var stackOutput awstypes.Stack
	resourceName := "aws_appstream_stack.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	settingsGroup := "group"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_applicationSettings(rName, true, settingsGroup),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "application_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.settings_group", settingsGroup),
				),
			},
			{
				Config: testAccStackConfig_applicationSettingsRemoved(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "application_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.settings_group", ""),
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

func TestAccAppStreamStack_applicationSettings_removeFromDisabled(t *testing.T) {
	ctx := acctest.Context(t)
	var stackOutput awstypes.Stack
	resourceName := "aws_appstream_stack.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_applicationSettingsDisabled(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "application_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.settings_group", ""),
				),
			},
			{
				Config:   testAccStackConfig_applicationSettingsRemoved(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccAppStreamStack_withTags(t *testing.T) {
	ctx := acctest.Context(t)
	var stackOutput awstypes.Stack
	resourceName := "aws_appstream_stack.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "Description of a test"
	descriptionUpdated := "Updated Description of a test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_complete(rName, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
				),
			},
			{
				Config: testAccStackConfig_tags(rName, descriptionUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionUpdated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", names.AttrValue),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key", names.AttrValue),
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

func TestAccAppStreamStack_streamingExperienceSettings_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var stackOutput awstypes.Stack
	resourceName := "aws_appstream_stack.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	preferredProtocol := "TCP"
	newPreferredProtocol := "UDP"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_streamingExperienceSettings(rName, preferredProtocol),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "streaming_experience_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "streaming_experience_settings.0.preferred_protocol", "TCP"),
				),
			},
			{
				Config: testAccStackConfig_streamingExperienceSettings(rName, newPreferredProtocol),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "streaming_experience_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "streaming_experience_settings.0.preferred_protocol", "UDP"),
				),
			},
		},
	})
}

func testAccCheckStackExists(ctx context.Context, n string, v *awstypes.Stack) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Appstream Stack ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamClient(ctx)

		output, err := tfappstream.FindStackByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckStackDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appstream_stack" {
				continue
			}

			_, err := tfappstream.FindStackByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Appstream Stack %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccStackConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_appstream_stack" "test" {
  name = %[1]q
}
`, name)
}

func testAccStackConfig_complete(name, description string) string {
	return fmt.Sprintf(`
resource "aws_appstream_stack" "test" {
  name        = %[1]q
  description = %[2]q

  embed_host_domains = ["example.com", "subdomain.example.com"]

  storage_connectors {
    connector_type = "HOMEFOLDERS"
  }

  user_settings {
    action     = "CLIPBOARD_COPY_FROM_LOCAL_DEVICE"
    permission = "ENABLED"
  }
  user_settings {
    action     = "CLIPBOARD_COPY_TO_LOCAL_DEVICE"
    permission = "ENABLED"
  }
  user_settings {
    action     = "DOMAIN_PASSWORD_SIGNIN"
    permission = "ENABLED"
  }
  user_settings {
    action     = "DOMAIN_SMART_CARD_SIGNIN"
    permission = "DISABLED"
  }
  user_settings {
    action     = "FILE_DOWNLOAD"
    permission = "ENABLED"
  }
  user_settings {
    action     = "FILE_UPLOAD"
    permission = "ENABLED"
  }
  user_settings {
    action     = "PRINTING_TO_LOCAL_DEVICE"
    permission = "ENABLED"
  }

  application_settings {
    enabled        = true
    settings_group = "SettingsGroup"
  }
}
`, name, description)
}

func testAccStackConfig_applicationSettings(name string, enabled bool, settingsGroupName string) string {
	return fmt.Sprintf(`
resource "aws_appstream_stack" "test" {
  name = %[1]q

  application_settings {
    enabled        = %[2]t
    settings_group = %[3]q
  }
}
`, name, enabled, settingsGroupName)
}

func testAccStackConfig_applicationSettingsDisabled(name string) string {
	return fmt.Sprintf(`
resource "aws_appstream_stack" "test" {
  name = %[1]q

  application_settings {
    enabled = false
  }
}
`, name)
}

func testAccStackConfig_applicationSettingsRemoved(name string) string {
	return fmt.Sprintf(`
resource "aws_appstream_stack" "test" {
  name = %[1]q
}
`, name)
}

func testAccStackConfig_tags(name, description string) string {
	return fmt.Sprintf(`
resource "aws_appstream_stack" "test" {
  name        = %[1]q
  description = %[2]q

  storage_connectors {
    connector_type = "HOMEFOLDERS"
  }

  user_settings {
    action     = "CLIPBOARD_COPY_FROM_LOCAL_DEVICE"
    permission = "ENABLED"
  }
  user_settings {
    action     = "CLIPBOARD_COPY_TO_LOCAL_DEVICE"
    permission = "ENABLED"
  }
  user_settings {
    action     = "DOMAIN_PASSWORD_SIGNIN"
    permission = "ENABLED"
  }
  user_settings {
    action     = "DOMAIN_SMART_CARD_SIGNIN"
    permission = "DISABLED"
  }
  user_settings {
    action     = "FILE_DOWNLOAD"
    permission = "ENABLED"
  }
  user_settings {
    action     = "FILE_UPLOAD"
    permission = "ENABLED"
  }
  user_settings {
    action     = "PRINTING_TO_LOCAL_DEVICE"
    permission = "ENABLED"
  }

  application_settings {
    enabled        = true
    settings_group = "SettingsGroup"
  }

  tags = {
    Key = "value"
  }
}
`, name, description)
}

func testAccStackConfig_streamingExperienceSettings(name, preferredProtocol string) string {
	return fmt.Sprintf(`
resource "aws_appstream_stack" "test" {
  name = %[1]q

  streaming_experience_settings {
    preferred_protocol = %[2]q
  }
}
`, name, preferredProtocol)
}
