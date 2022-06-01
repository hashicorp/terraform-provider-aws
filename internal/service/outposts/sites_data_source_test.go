package outposts_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/outposts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccOutpostsSitesDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_outposts_sites.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSites(t) },
		ErrorCheck:        acctest.ErrorCheck(t, outposts.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
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

func testAccPreCheckSites(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).OutpostsConn

	input := &outposts.ListSitesInput{}

	output, err := conn.ListSites(input)

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
