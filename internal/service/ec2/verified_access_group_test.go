package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVerifiedAccessAccessGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	//var verifiedaccessaccessgroup ec2.DescribeVerifiedAccessGroupsOutput

	rName := "test"
	resourceName := "aws_verifiedaccess_access_group." + rName
	description := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		//CheckDestroy:             testAccCheckVerifiedAccessAccessGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessGroupConfig_basic(rName, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", description),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

// func testAccCheckVerifiedAccessAccessGroupDestroy(ctx context.Context) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

// 		for _, rs := range s.RootModule().Resources {
// 			if rs.Type != "aws_verifiedaccess_access_group" {
// 				continue
// 			}

// 			_, err := tfec2.findVerifiedAccessGroupByID(ctx, conn, rs.Primary.ID)

// 			if tfresource.NotFound(err) {
// 				continue
// 			}

// 			if err != nil {
// 				return err
// 			}

// 			return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNameVerifiedAccessGroup, rs.Primary.ID, errors.New("not destroyed"))
// 		}
// 		return nil
// 	}
// }

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeVerifiedAccessGroupsInput{}
	_, err := conn.DescribeVerifiedAccessGroups(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
func testAccVerifiedAccessGroupConfig_basic(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_verifiedaccess_access_group" %[1]q {
  description = %[2]q
  verified_access_instance_id = "vai-014ba03ba2d7c6a6f"
}
`, rName, description)
}
