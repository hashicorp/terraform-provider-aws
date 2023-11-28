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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
					// resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_instance.test", "id"),
					// resource.TestCheckResourceAttr(resourceName, "configurations.#", "1"),
					// resource.TestCheckResourceAttr(resourceName, "targets.0.key", "InstanceIds"),
					// resource.TestCheckResourceAttr(resourceName, "targets.0.values.#", "1"),
					// resource.TestCheckResourceAttrPair(resourceName, "targets.0.values.0", "aws_instance.test", "id"),
					// resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					// resource.TestCheckResourceAttr(resourceName, "document_version", "$DEFAULT"),
					// resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccSecurityLakeDataLake_disappears(t *testing.T) {
	ctx := acctest.Context(t)
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
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfsecuritylake.ResourceDataLake, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDataLakeDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityLakeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securitylake_data_lake" {
				continue
			}

			_, err := tfsecuritylake.FindDataLakeByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
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

		*datalake = *resp

		return nil
	}
}

func testAccDataLakeConfig_basic() string {
	return fmt.Sprintf(`
resource "aws_securitylake_data_lake" "test" {
  meta_store_manager_role_arn = "arn:aws:iam::12345:role/service-role/AmazonSecurityLakeMetaStoreManager"

  configurations {
    region = "eu-west-1"
    encryption_configuration {
      kms_key_id = "S3_MANAGED_KEY"
    }

    lifecycle_configuration {
      transitions {
        days          = 31
        storage_class = "STANDARD_IA"
      }
      expiration {
        days = 300
      }
    }
	replication_configuration {
		role_arn = "arn:aws:iam::123454:role/service-role/AmazonSecurityLakeS3ReplicationRole"
		regions  = ["ap-south-1"]
	}
  }
}
`)
}
