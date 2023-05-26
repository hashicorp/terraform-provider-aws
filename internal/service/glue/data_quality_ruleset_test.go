package glue_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccGlueDataQualityRuleset_basic(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ruleset := "Rules = [Completeness \"colA\" between 0.4 and 0.8]"
	resourceName := "aws_glue_data_quality_ruleset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, glue.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityRulesetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataQualityRulesetConfig_basic(rName, ruleset),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataQualityRulesetExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("dataQualityRuleset/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "created_on"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_on"),
					resource.TestCheckResourceAttr(resourceName, "ruleset", ruleset),
					resource.TestCheckResourceAttr(resourceName, "target_table.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func testAccCheckDataQualityRulesetExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn()

		resp, err := tfglue.FindDataQualityRulesetByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("No Glue Data Quality Ruleset Found")
		}

		if aws.StringValue(resp.Name) != rs.Primary.ID {
			return fmt.Errorf("Glue Data Quality Ruleset Mismatch - existing: %q, state: %q",
				aws.StringValue(resp.Name), rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDataQualityRulesetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_glue_data_quality_ruleset" {
				continue
			}

			_, err := tfglue.FindDataQualityRulesetByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}
			return fmt.Errorf("Glue Data Quality Ruleset %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDataQualityRulesetConfig_basic(rName, ruleset string) string {
	return fmt.Sprintf(`
resource "aws_glue_data_quality_ruleset" "test" {
  name    = %[1]q
  ruleset = %[2]q
}
`, rName, ruleset)
}
