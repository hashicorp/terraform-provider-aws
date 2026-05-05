// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appstream_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppStreamStack_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var stackOutput awstypes.Stack
	resourceName := "aws_appstream_stack.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, t, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, "access_endpoints.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.settings_group", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, ""),
					resource.TestCheckResourceAttr(resourceName, "embed_host_domains.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "feedback_url", ""),
					resource.TestCheckResourceAttr(resourceName, "redirect_url", ""),
					resource.TestCheckResourceAttr(resourceName, "storage_connectors.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "user_settings.#", "8"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "0"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, t, resourceName, &stackOutput),
					acctest.CheckSDKResourceDisappears(ctx, t, tfappstream.ResourceStack(), resourceName),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	description := "Description of a test"
	descriptionUpdated := "Updated Description of a test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_complete(rName, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, t, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "embed_host_domains.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "embed_host_domains.*", "example.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "embed_host_domains.*", "subdomain.example.com"),
					resource.TestCheckResourceAttr(resourceName, "access_endpoints.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.settings_group", "SettingsGroup"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, ""),
					resource.TestCheckResourceAttr(resourceName, "feedback_url", ""),
					resource.TestCheckResourceAttr(resourceName, "redirect_url", ""),
					resource.TestCheckResourceAttr(resourceName, "storage_connectors.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_settings.#", "8"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "0"),
				),
			},
			{
				Config: testAccStackConfig_complete(rName, descriptionUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, t, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionUpdated),
					resource.TestCheckResourceAttr(resourceName, "embed_host_domains.#", "2"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	settingsGroup := "group"
	settingsGroupUpdated := "group-updated"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_applicationSettings(rName, true, settingsGroup),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, t, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "application_settings.#", "1"),
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
					testAccCheckStackExists(ctx, t, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "application_settings.#", "1"),
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
					testAccCheckStackExists(ctx, t, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "application_settings.#", "1"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	settingsGroup := "group"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_applicationSettings(rName, true, settingsGroup),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, t, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "application_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.settings_group", settingsGroup),
				),
			},
			{
				Config: testAccStackConfig_applicationSettingsRemoved(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, t, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "application_settings.#", "1"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_applicationSettingsDisabled(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, t, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "application_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.settings_group", ""),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccStackConfig_applicationSettingsRemoved(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccAppStreamStack_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var stackOutput awstypes.Stack
	resourceName := "aws_appstream_stack.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	description := "Description of a test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_tags1(rName, description, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, t, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStackConfig_tags2(rName, description, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, t, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccStackConfig_tags1(rName, description, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, t, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccAppStreamStack_streamingExperienceSettings_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var stackOutput awstypes.Stack
	resourceName := "aws_appstream_stack.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	preferredProtocol := "TCP"
	newPreferredProtocol := "UDP"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_streamingExperienceSettings(rName, preferredProtocol),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, t, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "streaming_experience_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "streaming_experience_settings.0.preferred_protocol", "TCP"),
				),
			},
			{
				Config: testAccStackConfig_streamingExperienceSettings(rName, newPreferredProtocol),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, t, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "streaming_experience_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "streaming_experience_settings.0.preferred_protocol", "UDP"),
				),
			},
		},
	})
}

func testAccCheckStackDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AppStreamClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appstream_stack" {
				continue
			}

			_, err := tfappstream.FindStackByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AppStream Stack %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckStackExists(ctx context.Context, t *testing.T, n string, v *awstypes.Stack) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AppStreamClient(ctx)

		output, err := tfappstream.FindStackByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

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
    action     = "AUTO_TIME_ZONE_REDIRECTION"
    permission = "DISABLED"
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

func testAccStackConfig_tags1(name, description, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_appstream_stack" "test" {
  name        = %[1]q
  description = %[2]q

  storage_connectors {
    connector_type = "HOMEFOLDERS"
  }

  user_settings {
    action     = "AUTO_TIME_ZONE_REDIRECTION"
    permission = "DISABLED"
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
    %[3]q = %[4]q
  }
}
`, name, description, tagKey1, tagValue1)
}

func testAccStackConfig_tags2(name, description, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_appstream_stack" "test" {
  name        = %[1]q
  description = %[2]q

  storage_connectors {
    connector_type = "HOMEFOLDERS"
  }

  user_settings {
    action     = "AUTO_TIME_ZONE_REDIRECTION"
    permission = "DISABLED"
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
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, name, description, tagKey1, tagValue1, tagKey2, tagValue2)
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
