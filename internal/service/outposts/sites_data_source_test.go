// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package outposts_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/outposts"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOutpostsSitesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_outposts_sites.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSites(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OutpostsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSitesDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSitesAttributes(dataSourceName),
				),
			},
		},
	})
}

func testAccCheckSitesAttributes(dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", dataSourceName)
		}

		if v := rs.Primary.Attributes["ids.#"]; v == "0" {
			return fmt.Errorf("expected at least one ids result, got none")
		}

		return nil
	}
}

func testAccPreCheckSites(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).OutpostsClient(ctx)

	input := &outposts.ListSitesInput{}

	output, err := conn.ListSites(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	// Ensure there is at least one Site
	if output == nil || len(output.Sites) == 0 {
		t.Skip("skipping since no Sites Outpost found")
	}
}

func testAccSitesDataSourceConfig_basic() string {
	return `
data "aws_outposts_sites" "test" {}
`
}
