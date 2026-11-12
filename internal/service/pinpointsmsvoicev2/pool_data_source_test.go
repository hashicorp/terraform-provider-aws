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

func TestAccPinpointSMSVoiceV2PoolDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_pinpointsmsvoicev2_pool.test"
	resourceName := "aws_pinpointsmsvoicev2_pool.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPool(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolDataSourceConfig_basic(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New(names.AttrARN), resourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("created_timestamp"), knownvalue.NotNull()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("deletion_protection_enabled"), resourceName, tfjsonpath.New("deletion_protection_enabled"), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("message_type"), resourceName, tfjsonpath.New("message_type"), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("opt_out_list_name"), resourceName, tfjsonpath.New("opt_out_list_name"), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("origination_identities"), resourceName, tfjsonpath.New("origination_identities"), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("self_managed_opt_outs_enabled"), resourceName, tfjsonpath.New("self_managed_opt_outs_enabled"), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("shared_routes_enabled"), resourceName, tfjsonpath.New("shared_routes_enabled"), compare.ValuesSame()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrStatus), knownvalue.StringExact("ACTIVE")),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("two_way_channel_arn"), resourceName, tfjsonpath.New("two_way_channel_arn"), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("two_way_channel_role"), resourceName, tfjsonpath.New("two_way_channel_role"), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("two_way_enabled"), resourceName, tfjsonpath.New("two_way_enabled"), compare.ValuesSame()),
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2PoolDataSource_associations(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_pinpointsmsvoicev2_pool.test"
	resourceName := "aws_pinpointsmsvoicev2_pool.test"
	optOutListResourceName := "aws_pinpointsmsvoicev2_opt_out_list.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPool(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolDataSourceConfig_associations(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("opt_out_list_name"), optOutListResourceName, tfjsonpath.New(names.AttrName), compare.ValuesSame()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("origination_identities"), knownvalue.SetSizeExact(2)),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("origination_identities"), resourceName, tfjsonpath.New("origination_identities"), compare.ValuesSame()),
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2PoolDataSource_nonExistent(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPool(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPoolDataSourceConfig_nonExistent(),
				ExpectError: regexache.MustCompile(`couldn't find resource`),
			},
		},
	})
}

func testAccPoolDataSourceConfig_basic() string {
	return `
data "aws_pinpointsmsvoicev2_pool" "test" {
  id = aws_pinpointsmsvoicev2_pool.test.id
}

resource "aws_pinpointsmsvoicev2_pool" "test" {
  iso_country_code       = "US"
  message_type           = "TRANSACTIONAL"
  origination_identities = [aws_pinpointsmsvoicev2_phone_number.test.arn]
}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  force_disassociate  = true
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}
`
}

func testAccPoolDataSourceConfig_associations(rName string) string {
	return fmt.Sprintf(`
data "aws_pinpointsmsvoicev2_pool" "test" {
  id = aws_pinpointsmsvoicev2_pool.test.id
}

resource "aws_pinpointsmsvoicev2_pool" "test" {
  iso_country_code  = "US"
  message_type      = "TRANSACTIONAL"
  opt_out_list_name = aws_pinpointsmsvoicev2_opt_out_list.test.name
  origination_identities = [
    aws_pinpointsmsvoicev2_phone_number.test[0].arn,
    aws_pinpointsmsvoicev2_phone_number.test[1].arn,
  ]
}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  count = 2

  force_disassociate  = true
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}

resource "aws_pinpointsmsvoicev2_opt_out_list" "test" {
  name = %[1]q
}
`, rName)
}

func testAccPoolDataSourceConfig_nonExistent() string {
	return `
data "aws_pinpointsmsvoicev2_pool" "test" {
  id = "pool-00000000000000000000000000000000"
}
`
}
