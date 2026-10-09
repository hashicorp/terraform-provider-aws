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

func TestAccPinpointSMSVoiceV2OptOutListDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var optOutList awstypes.OptOutListInformation
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_pinpointsmsvoicev2_opt_out_list.test"
	resourceName := "aws_pinpointsmsvoicev2_opt_out_list.test"
	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckOptOutList(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOptOutListDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOptOutListDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOptOutListExists(ctx, t, resourceName, &optOutList),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
		},
	})
}

func testAccOptOutListDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_opt_out_list" "test" {
  name = %[1]q
}

data "aws_pinpointsmsvoicev2_opt_out_list" "test" {
  name = aws_pinpointsmsvoicev2_opt_out_list.test.name
}
`, rName)
}
