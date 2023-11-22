package securitylake_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfsecuritylake "github.com/hashicorp/terraform-provider-aws/internal/service/securitylake"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSecurityLakeDataLake_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var datalake types.DataLakeResource
	// rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_securitylake_data_lake.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLake),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataLakeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataLakeConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataLakeExists(ctx, resourceName, &datalake),
					// resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					// resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					// resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
					// 	"console_access": "false",
					// 	"groups.#":       "0",
					// 	"username":       "Test",
					// 	"password":       "TestTest1234",
					// }),
					// acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "securitylake", regexp.MustCompile(`data-lake/:+.`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"meta_store_manager_role_arn"},
			},
		},
	})
}

// func TestAccSecurityLakeDataLake_disappears(t *testing.T) {
// 	ctx := acctest.Context(t)
// 	if testing.Short() {
// 		t.Skip("skipping long-running test in short mode")
// 	}

// 	var datalake securitylake.DescribeDataLakeResponse
// 	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
// 	resourceName := "aws_securitylake_data_lake.test"

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck: func() {
// 			acctest.PreCheck(ctx, t)
// 			acctest.PreCheckPartitionHasService(t, names.SecurityLakeEndpointID)
// 			testAccPreCheck(t)
// 		},
// 		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeEndpointID),
// 		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
// 		CheckDestroy:             testAccCheckDataLakeDestroy(ctx),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccDataLakeConfig_basic(rName, testAccDataLakeVersionNewer),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckDataLakeExists(ctx, resourceName, &datalake),
// 					// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
// 					// but expects a new resource factory function as the third argument. To expose this
// 					// private function to the testing package, you may need to add a line like the following
// 					// to exports_test.go:
// 					//
// 					//   var ResourceDataLake = newResourceDataLake
// 					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfsecuritylake.ResourceDataLake, resourceName),
// 				),
// 				ExpectNonEmptyPlan: true,
// 			},
// 		},
// 	})
// }

func testAccCheckDataLakeDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityLakeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securitylake_data_lake" {
				continue
			}

			_, err := tfsecuritylake.FindDataLakeByID(ctx, conn, rs.Primary.ID)
			if err != nil {
				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.SecurityLake, create.ErrActionCheckingDestroyed, tfsecuritylake.ResNameDataLake, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckDataLakeExists(ctx context.Context, name string, datalake *types.DataLakeResource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SecurityLake, create.ErrActionCheckingExistence, tfsecuritylake.ResNameDataLake, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SecurityLake, create.ErrActionCheckingExistence, tfsecuritylake.ResNameDataLake, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityLakeClient(ctx)
		resp, err := tfsecuritylake.FindDataLakeByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.SecurityLake, create.ErrActionCheckingExistence, tfsecuritylake.ResNameDataLake, rs.Primary.ID, err)
		}

		dl := &resp.DataLakes[0]

		*datalake = *dl

		return nil
	}
}

func testAccDataLakeConfig_basic() string {
	return fmt.Sprintf(`
	resource "aws_securitylake_data_lake" "test" {
		meta_store_manager_role_arn = "arn:aws:iam::182198062889:role/service-role/AmazonSecurityLakeMetaStoreManager"
	  
		configurations {
		  region = "eu-west-2"
		}
	  }
`)
}
