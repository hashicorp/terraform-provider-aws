package quicksight_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/fis/types"
	"github.com/aws/aws-sdk-go/service/quicksight"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQuickSightFolder_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var folder quicksight.Folder
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_quicksight_folder.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, quicksight.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, quicksight.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFolderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFolderConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFolderExists(ctx, resourceName, &folder),
					resource.TestCheckResourceAttr(resourceName, "folder_id", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "quicksight", regexp.MustCompile(`folder/+.`)),
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

func TestAccQuickSightFolder_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var folder quicksight.Folder
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_quicksight_folder.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, quicksight.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, quicksight.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFolderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFolderConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFolderExists(ctx, resourceName, &folder),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfquicksight.ResourceFolder, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFolderDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_folder" {
				continue
			}

			_, err := tfquicksight.FindFolderByID(ctx, conn, rs.Primary.ID)
			if err != nil {
				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.QuickSight, create.ErrActionCheckingDestroyed, tfquicksight.ResNameFolder, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckFolderExists(ctx context.Context, name string, folder *quicksight.Folder) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.QuickSight, create.ErrActionCheckingExistence, tfquicksight.ResNameFolder, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.QuickSight, create.ErrActionCheckingExistence, tfquicksight.ResNameFolder, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn()
		output, err := tfquicksight.FindFolderByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.QuickSight, create.ErrActionCheckingExistence, tfquicksight.ResNameFolder, rs.Primary.ID, err)
		}

		*folder = *output

		return nil
	}
}

func testAccFolderConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_quicksight_folder" "test" {
  folder_id = %[1]q
  name      = %[1]q
}
`, rName)
}
