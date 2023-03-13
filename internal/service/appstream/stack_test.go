package appstream_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
)

func TestAccAppStreamStack_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var stackOutput appstream.Stack
	resourceName := "aws_appstream_stack.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "access_endpoints.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.settings_group", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "display_name", ""),
					resource.TestCheckResourceAttr(resourceName, "embed_host_domains.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "feedback_url", ""),
					resource.TestCheckResourceAttr(resourceName, "redirect_url", ""),
					resource.TestCheckResourceAttr(resourceName, "storage_connectors.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "user_settings.#", "7"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
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
	var stackOutput appstream.Stack
	resourceName := "aws_appstream_stack.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, appstream.EndpointsID),
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
	var stackOutput appstream.Stack
	resourceName := "aws_appstream_stack.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "Description of a test"
	descriptionUpdated := "Updated Description of a test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_complete(rName, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "embed_host_domains.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "embed_host_domains.*", "example.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "embed_host_domains.*", "subdomain.example.com"),
					resource.TestCheckResourceAttr(resourceName, "access_endpoints.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.settings_group", "SettingsGroup"),
					resource.TestCheckResourceAttr(resourceName, "display_name", ""),
					resource.TestCheckResourceAttr(resourceName, "feedback_url", ""),
					resource.TestCheckResourceAttr(resourceName, "redirect_url", ""),
					resource.TestCheckResourceAttr(resourceName, "storage_connectors.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_settings.#", "7"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
				),
			},
			{
				Config: testAccStackConfig_complete(rName, descriptionUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionUpdated),
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
	var stackOutput appstream.Stack
	resourceName := "aws_appstream_stack.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	settingsGroup := "group"
	settingsGroupUpdated := "group-updated"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_applicationSettings(rName, true, settingsGroup),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "application_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.enabled", "true"),
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
					resource.TestCheckResourceAttr(resourceName, "application_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.enabled", "true"),
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
					resource.TestCheckResourceAttr(resourceName, "application_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.enabled", "false"),
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
	var stackOutput appstream.Stack
	resourceName := "aws_appstream_stack.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	settingsGroup := "group"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_applicationSettings(rName, true, settingsGroup),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "application_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.settings_group", settingsGroup),
				),
			},
			{
				Config: testAccStackConfig_applicationSettingsRemoved(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "application_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.enabled", "false"),
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
	var stackOutput appstream.Stack
	resourceName := "aws_appstream_stack.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_applicationSettingsDisabled(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "application_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.enabled", "false"),
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
	var stackOutput appstream.Stack
	resourceName := "aws_appstream_stack.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "Description of a test"
	descriptionUpdated := "Updated Description of a test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_complete(rName, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
				),
			},
			{
				Config: testAccStackConfig_tags(rName, descriptionUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionUpdated),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", "value"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key", "value"),
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

func testAccCheckStackExists(ctx context.Context, resourceName string, appStreamStack *appstream.Stack) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamConn()
		resp, err := conn.DescribeStacksWithContext(ctx, &appstream.DescribeStacksInput{Names: []*string{aws.String(rs.Primary.ID)}})

		if err != nil {
			return fmt.Errorf("problem checking for AppStream Stack existence: %w", err)
		}

		if resp == nil || len(resp.Stacks) == 0 {
			return fmt.Errorf("appstream stack %q does not exist", rs.Primary.ID)
		}

		*appStreamStack = *resp.Stacks[0]

		return nil
	}
}

func testAccCheckStackDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appstream_stack" {
				continue
			}

			resp, err := conn.DescribeStacksWithContext(ctx, &appstream.DescribeStacksInput{Names: []*string{aws.String(rs.Primary.ID)}})

			if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
				continue
			}

			if err != nil {
				return fmt.Errorf("problem while checking AppStream Stack was destroyed: %w", err)
			}

			if resp != nil && len(resp.Stacks) > 0 {
				return fmt.Errorf("appstream stack %q still exists", rs.Primary.ID)
			}
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
