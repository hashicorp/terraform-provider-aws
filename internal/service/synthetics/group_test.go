package synthetics_test

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsynthetics "github.com/hashicorp/terraform-provider-aws/internal/service/synthetics"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/synthetics"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccSyntheticsGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var group synthetics.Group
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, synthetics.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", synthetics.ServiceName, regexp.MustCompile(`group:.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccSyntheticsGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var group synthetics.Group
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, synthetics.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsynthetics.ResourceGroup(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsynthetics.ResourceGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckGroupExists(ctx context.Context, name string, group *synthetics.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Synthetics Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SyntheticsConn()
		output, err := tfsynthetics.FindGroupByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*group = *output

		return nil
	}
}

func testAccCheckGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SyntheticsConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_synthetics_group" {
				continue
			}

			_, err := tfsynthetics.FindGroupByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("synthetics Group %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_synthetics_group" "test" {
  name             		  = %[1]q
}
`, rName)
}
