package s3control_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3control"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3control "github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccS3ControlStorageLensConfiguration_basic(t *testing.T) {
	var v s3control.StorageLensConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_storage_lens_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3control.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStorageLensConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageLensConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStorageLensConfigurationExists(resourceName, &v),
					acctest.CheckResourceAttrAccountID(resourceName, "account_id"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "s3", fmt.Sprintf("storage-lens/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.activity_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.activity_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.account_level.0.bucket_level.0.prefix_level.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "storage_lens_configuration.0.exclude.#", "0"),
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

func TestAccS3ControlStorageLensConfiguration_disappears(t *testing.T) {
	var v s3control.StorageLensConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_storage_lens_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3control.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStorageLensConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageLensConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageLensConfigurationExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3control.ResourceStorageLensConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3ControlStorageLensConfiguration_tags(t *testing.T) {
	var v s3control.StorageLensConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_storage_lens_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3control.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStorageLensConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageLensConfigurationConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageLensConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStorageLensConfigurationConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageLensConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccStorageLensConfigurationConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageLensConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckStorageLensConfigurationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3control_object_lambda_access_point" {
			continue
		}

		accountID, configID, err := tfs3control.StorageLensConfigurationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfs3control.FindStorageLensConfigurationByAccountIDAndConfigID(context.Background(), conn, accountID, configID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("S3 Storage Lens Configuration %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckStorageLensConfigurationExists(n string, v *s3control.StorageLensConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Storage Lens Configuration ID is set")
		}

		accountID, configID, err := tfs3control.StorageLensConfigurationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlConn

		output, err := tfs3control.FindStorageLensConfigurationByAccountIDAndConfigID(context.Background(), conn, accountID, configID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccStorageLensConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3control_storage_lens_configuration" "test" {
  config_id = %[1]q

  storage_lens_configuration {
    enabled = true

    account_level {
      bucket_level {}
    }
  }
}
`, rName)
}

func testAccStorageLensConfigurationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_s3control_storage_lens_configuration" "test" {
  config_id = %[1]q

  storage_lens_configuration {
    enabled = true

    account_level {
      bucket_level {}
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccStorageLensConfigurationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_s3control_storage_lens_configuration" "test" {
  config_id = %[1]q

  storage_lens_configuration {
    enabled = true

    account_level {
      bucket_level {}
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
