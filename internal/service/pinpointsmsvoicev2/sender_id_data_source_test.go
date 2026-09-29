// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccPinpointSMSVoiceV2SenderIDDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	senderID := testAccRandomSenderID(t)
	isoCountryCode := "GB"
	dataSourceName := "data.aws_pinpointsmsvoicev2_sender_id.test"
	resourceName := "aws_pinpointsmsvoicev2_sender_id.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSenderID(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSenderIDDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSenderIDDataSourceConfig_basic(senderID, isoCountryCode),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "deletion_protection_enabled", resourceName, "deletion_protection_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "iso_country_code", resourceName, "iso_country_code"),
					resource.TestCheckResourceAttrPair(dataSourceName, "message_types", resourceName, "message_types"),
					resource.TestCheckResourceAttrPair(dataSourceName, "monthly_leasing_price", resourceName, "monthly_leasing_price"),
					resource.TestCheckResourceAttrPair(dataSourceName, "registered", resourceName, "registered"),
					resource.TestCheckResourceAttrPair(dataSourceName, "registration_id", resourceName, "registration_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "sender_id", resourceName, "sender_id"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
		},
	})
}

func testAccSenderIDDataSourceConfig_basic(senderID, isoCountryCode string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_sender_id" "test" {
  sender_id        = %[1]q
  iso_country_code = %[2]q
}

data "aws_pinpointsmsvoicev2_sender_id" "test" {
  sender_id        = aws_pinpointsmsvoicev2_sender_id.test.sender_id
  iso_country_code = aws_pinpointsmsvoicev2_sender_id.test.iso_country_code
}
`, senderID, isoCountryCode)
}
