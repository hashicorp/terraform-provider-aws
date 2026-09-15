// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2ConfigurationSetDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set.test"
	dataSourceName := "data.aws_sesv2_configuration_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, dataSourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_set_name", dataSourceName, "configuration_set_name"),
					resource.TestCheckResourceAttrPair(resourceName, "delivery_options.#", dataSourceName, "delivery_options.#"),
					resource.TestCheckResourceAttrPair(resourceName, "delivery_options.0.max_delivery_seconds", dataSourceName, "delivery_options.0.max_delivery_seconds"),
					resource.TestCheckResourceAttrPair(resourceName, "delivery_options.0.sending_pool_name", dataSourceName, "delivery_options.0.sending_pool_name"),
					resource.TestCheckResourceAttrPair(resourceName, "delivery_options.0.tls_policy", dataSourceName, "delivery_options.0.tls_policy"),
					resource.TestCheckResourceAttrPair(resourceName, "reputation_options.#", dataSourceName, "reputation_options.#"),
					resource.TestCheckResourceAttrPair(resourceName, "reputation_options.0.last_fresh_start", dataSourceName, "reputation_options.0.last_fresh_start"),
					resource.TestCheckResourceAttrPair(resourceName, "reputation_options.0.reputation_metrics_enabled", dataSourceName, "reputation_options.0.reputation_metrics_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "sending_options.#", dataSourceName, "sending_options.#"),
					resource.TestCheckResourceAttrPair(resourceName, "sending_options.0.sending_enabled", dataSourceName, "sending_options.0.sending_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "suppression_options.#", dataSourceName, "suppression_options.#"),
					resource.TestCheckResourceAttrPair(resourceName, "suppression_options.0.suppressed_reasons", dataSourceName, "suppression_options.0.suppressed_reasons"),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(resourceName, "vdm_options.#", dataSourceName, "vdm_options.#"),
					resource.TestCheckResourceAttrPair(resourceName, "vdm_options.0.dashboard_options.#", dataSourceName, "vdm_options.0.dashboard_options.#"),
					resource.TestCheckResourceAttrPair(resourceName, "vdm_options.0.dashboard_options.0.engagement_metrics", dataSourceName, "vdm_options.0.dashboard_options.0.engagement_metrics"),
					resource.TestCheckResourceAttrPair(resourceName, "vdm_options.0.guardian_options.#", dataSourceName, "vdm_options.0.guardian_options.#"),
					resource.TestCheckResourceAttrPair(resourceName, "vdm_options.0.guardian_options.0.optimized_shared_delivery", dataSourceName, "vdm_options.0.guardian_options.0.optimized_shared_delivery"),
				),
			},
		},
	})
}

func testAccConfigurationSetDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = %[1]q

  delivery_options {
    max_delivery_seconds = 300
    tls_policy           = "REQUIRE"
  }

  reputation_options {
    reputation_metrics_enabled = true
  }

  sending_options {
    sending_enabled = true
  }

  suppression_options {
    suppressed_reasons = ["BOUNCE"]
  }

  vdm_options {
    dashboard_options {
      engagement_metrics = "ENABLED"
    }

    guardian_options {
      optimized_shared_delivery = "ENABLED"
    }
  }
}

data "aws_sesv2_configuration_set" "test" {
  configuration_set_name = aws_sesv2_configuration_set.test.configuration_set_name
}
`, rName)
}
