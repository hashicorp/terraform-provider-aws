// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2_test

import (
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

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
