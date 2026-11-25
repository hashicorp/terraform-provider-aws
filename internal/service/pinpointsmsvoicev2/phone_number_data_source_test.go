// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccPinpointSMSVoiceV2PhoneNumberDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_pinpointsmsvoicev2_phone_number.test"
	resourceName := "aws_pinpointsmsvoicev2_phone_number.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPhoneNumber(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPhoneNumberDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPhoneNumberDataSourceConfig_basic(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New(names.AttrARN), resourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("created_timestamp"), knownvalue.NotNull()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("deletion_protection_enabled"), resourceName, tfjsonpath.New("deletion_protection_enabled"), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("iso_country_code"), resourceName, tfjsonpath.New("iso_country_code"), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("message_type"), resourceName, tfjsonpath.New("message_type"), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("monthly_leasing_price"), resourceName, tfjsonpath.New("monthly_leasing_price"), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("number_capabilities"), resourceName, tfjsonpath.New("number_capabilities"), compare.ValuesSame()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("number_type"), knownvalue.StringExact("SIMULATOR")),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("opt_out_list_name"), resourceName, tfjsonpath.New("opt_out_list_name"), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("phone_number"), resourceName, tfjsonpath.New("phone_number"), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("self_managed_opt_outs_enabled"), resourceName, tfjsonpath.New("self_managed_opt_outs_enabled"), compare.ValuesSame()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrStatus), knownvalue.StringExact("ACTIVE")),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("two_way_channel_arn"), resourceName, tfjsonpath.New("two_way_channel_arn"), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("two_way_channel_role"), resourceName, tfjsonpath.New("two_way_channel_role"), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("two_way_channel_enabled"), resourceName, tfjsonpath.New("two_way_channel_enabled"), compare.ValuesSame()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
		},
	})
}

// A phone number associated with a pool reports the pool's ID. The unassociated
// case in the basic test leaves pool_id null, so both paths are covered.
func TestAccPinpointSMSVoiceV2PhoneNumberDataSource_pool(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_pinpointsmsvoicev2_phone_number.test"
	poolResourceName := "aws_pinpointsmsvoicev2_pool.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPhoneNumber(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPhoneNumberDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPhoneNumberDataSourceConfig_pool,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("pool_id"), poolResourceName, tfjsonpath.New(names.AttrID), compare.ValuesSame()),
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2PhoneNumberDataSource_nonExistent(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPhoneNumber(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPhoneNumberDataSourceConfig_nonExistent,
				ExpectError: regexache.MustCompile(`couldn't find resource`),
			},
		},
	})
}

func testAccPhoneNumberDataSourceConfig_basic() string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]

  tags = {
    %[1]q = %[2]q
  }
}

data "aws_pinpointsmsvoicev2_phone_number" "test" {
  id = aws_pinpointsmsvoicev2_phone_number.test.id
}
`, acctest.CtKey1, acctest.CtValue1)
}

const testAccPhoneNumberDataSourceConfig_pool = `
resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  force_disassociate  = true
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}

resource "aws_pinpointsmsvoicev2_pool" "test" {
  iso_country_code       = "US"
  message_type           = "TRANSACTIONAL"
  origination_identities = [aws_pinpointsmsvoicev2_phone_number.test.arn]
}

data "aws_pinpointsmsvoicev2_phone_number" "test" {
  id = aws_pinpointsmsvoicev2_phone_number.test.id

  # The pool association is what populates pool_id, so the read must happen
  # after the pool exists.
  depends_on = [aws_pinpointsmsvoicev2_pool.test]
}
`

const testAccPhoneNumberDataSourceConfig_nonExistent = `
data "aws_pinpointsmsvoicev2_phone_number" "test" {
  id = "phone-00000000000000000000000000000000"
}
`
