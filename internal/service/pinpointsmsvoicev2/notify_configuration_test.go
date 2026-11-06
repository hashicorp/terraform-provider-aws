// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfpinpointsmsvoicev2 "github.com/hashicorp/terraform-provider-aws/internal/service/pinpointsmsvoicev2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// testAccRandomNotifyConfigurationDisplayName returns a random display name (1-15 chars).
func testAccRandomNotifyConfigurationDisplayName(t *testing.T) string {
	t.Helper()

	return acctest.RandStringFromCharSet(t, 10, acctest.CharSetAlpha)
}

func TestAccPinpointSMSVoiceV2NotifyConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var notifyConfig awstypes.NotifyConfigurationInformation
	rName := testAccRandomNotifyConfigurationDisplayName(t)
	resourceName := "aws_pinpointsmsvoicev2_notify_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckNotifyConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotifyConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNotifyConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotifyConfigurationExists(ctx, t, resourceName, &notifyConfig),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_timestamp"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDisplayName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("use_case"), knownvalue.StringExact("CODE_VERIFICATION")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("enabled_channels"), knownvalue.SetExact([]knownvalue.Check{knownvalue.StringExact("SMS")})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("enabled_countries"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("deletion_protection_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2NotifyConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	var notifyConfig awstypes.NotifyConfigurationInformation
	rName := testAccRandomNotifyConfigurationDisplayName(t)
	resourceName := "aws_pinpointsmsvoicev2_notify_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckNotifyConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotifyConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNotifyConfigurationConfig_update(rName, `["SMS"]`, `["US"]`, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotifyConfigurationExists(ctx, t, resourceName, &notifyConfig),
					testAccCheckNotifyConfigurationActive(t, &notifyConfig),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("enabled_channels"), knownvalue.SetExact([]knownvalue.Check{knownvalue.StringExact("SMS")})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("enabled_countries"), knownvalue.SetExact([]knownvalue.Check{knownvalue.StringExact("US")})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("deletion_protection_enabled"), knownvalue.Bool(false)),
				},
			},
			{
				Config: testAccNotifyConfigurationConfig_update(rName, `["SMS", "VOICE"]`, `["US", "CA"]`, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotifyConfigurationExists(ctx, t, resourceName, &notifyConfig),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("enabled_channels"), knownvalue.SetExact([]knownvalue.Check{knownvalue.StringExact("SMS"), knownvalue.StringExact("VOICE")})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("enabled_countries"), knownvalue.SetExact([]knownvalue.Check{knownvalue.StringExact("US"), knownvalue.StringExact("CA")})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("deletion_protection_enabled"), knownvalue.Bool(true)),
				},
			},
			{
				Config: testAccNotifyConfigurationConfig_update(rName, `["SMS", "VOICE"]`, `["US", "CA"]`, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotifyConfigurationExists(ctx, t, resourceName, &notifyConfig),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("deletion_protection_enabled"), knownvalue.Bool(false)),
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2NotifyConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var notifyConfig awstypes.NotifyConfigurationInformation
	rName := testAccRandomNotifyConfigurationDisplayName(t)
	resourceName := "aws_pinpointsmsvoicev2_notify_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckNotifyConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotifyConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNotifyConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotifyConfigurationExists(ctx, t, resourceName, &notifyConfig),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfpinpointsmsvoicev2.ResourceNotifyConfiguration, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2NotifyConfigurationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var notifyConfig awstypes.NotifyConfigurationInformation
	rName := testAccRandomNotifyConfigurationDisplayName(t)
	resourceName := "aws_pinpointsmsvoicev2_notify_configuration.test"
	dataSourceName := "data.aws_pinpointsmsvoicev2_notify_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckNotifyConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotifyConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNotifyConfigurationConfig_dataSource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotifyConfigurationExists(ctx, t, resourceName, &notifyConfig),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrDisplayName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("use_case"), knownvalue.StringExact("CODE_VERIFICATION")),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("enabled_channels"), knownvalue.SetExact([]knownvalue.Check{knownvalue.StringExact("SMS")})),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("enabled_countries"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("deletion_protection_enabled"), knownvalue.Bool(false)),
				},
			},
		},
	})
}

func testAccCheckNotifyConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pinpointsmsvoicev2_notify_configuration" {
				continue
			}

			_, err := tfpinpointsmsvoicev2.FindNotifyConfigurationByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("End User Messaging SMS Notify Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckNotifyConfigurationExists(ctx context.Context, t *testing.T, n string, v *awstypes.NotifyConfigurationInformation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No End User Messaging SMS Notify Configuration ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)

		output, err := tfpinpointsmsvoicev2.FindNotifyConfigurationByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckNotifyConfigurationActive(t *testing.T, v *awstypes.NotifyConfigurationInformation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if v.Status != awstypes.NotifyConfigurationStatusActive {
			t.Skipf("skipping update test: notify configuration status is %s (reason: %s)", v.Status, aws.ToString(v.RejectionReason))
		}

		return nil
	}
}

func testAccPreCheckNotifyConfiguration(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)

	input := &pinpointsmsvoicev2.DescribeNotifyConfigurationsInput{}

	_, err := conn.DescribeNotifyConfigurations(ctx, input)
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccNotifyConfigurationConfig_basic(displayName string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_notify_configuration" "test" {
  display_name     = %[1]q
  use_case         = "CODE_VERIFICATION"
  enabled_channels = ["SMS"]
}
`, displayName)
}

func testAccNotifyConfigurationConfig_update(displayName, channels, countries string, deletionProtection bool) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_notify_configuration" "test" {
  display_name                = %[1]q
  use_case                    = "CODE_VERIFICATION"
  enabled_channels            = %[2]s
  enabled_countries           = %[3]s
  deletion_protection_enabled = %[4]t
}
`, displayName, channels, countries, deletionProtection)
}

func testAccNotifyConfigurationConfig_dataSource(displayName string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_notify_configuration" "test" {
  display_name     = %[1]q
  use_case         = "CODE_VERIFICATION"
  enabled_channels = ["SMS"]
}

data "aws_pinpointsmsvoicev2_notify_configuration" "test" {
  id = aws_pinpointsmsvoicev2_notify_configuration.test.id
}
`, displayName)
}
