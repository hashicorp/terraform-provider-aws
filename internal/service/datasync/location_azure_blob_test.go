// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/datasync"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdatasync "github.com/hashicorp/terraform-provider-aws/internal/service/datasync"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataSyncLocationAzureBlob_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v datasync.DescribeLocationAzureBlobOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_location_azure_blob.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationAzureBlobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationAzureBlobConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLocationAzureBlobExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_tier", "HOT"),
					resource.TestCheckResourceAttr(resourceName, "agent_arns.#", acctest.Ct1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "datasync", regexache.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "SAS"),
					resource.TestCheckResourceAttr(resourceName, "blob_type", "BLOCK"),
					resource.TestCheckResourceAttr(resourceName, "container_url", "https://myaccount.blob.core.windows.net/mycontainer"),
					resource.TestCheckResourceAttr(resourceName, "sas_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "sas_configuration.0.token"),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/myvdir1/myvdir2/"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestMatchResourceAttr(resourceName, names.AttrURI, regexache.MustCompile(`^azure-blob://.+/`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"sas_configuration"},
			},
		},
	})
}

func TestAccDataSyncLocationAzureBlob_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v datasync.DescribeLocationAzureBlobOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_location_azure_blob.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationAzureBlobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationAzureBlobConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationAzureBlobExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdatasync.ResourceLocationAzureBlob(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataSyncLocationAzureBlob_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v datasync.DescribeLocationAzureBlobOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_location_azure_blob.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationAzureBlobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationAzureBlobConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationAzureBlobExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"sas_configuration"},
			},
			{
				Config: testAccLocationAzureBlobConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationAzureBlobExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccLocationAzureBlobConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationAzureBlobExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
		},
	})
}

func TestAccDataSyncLocationAzureBlob_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v datasync.DescribeLocationAzureBlobOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_location_azure_blob.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationAzureBlobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationAzureBlobConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLocationAzureBlobExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_tier", "HOT"),
					resource.TestCheckResourceAttr(resourceName, "agent_arns.#", acctest.Ct1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "datasync", regexache.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "SAS"),
					resource.TestCheckResourceAttr(resourceName, "blob_type", "BLOCK"),
					resource.TestCheckResourceAttr(resourceName, "container_url", "https://myaccount.blob.core.windows.net/mycontainer"),
					resource.TestCheckResourceAttr(resourceName, "sas_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "sas_configuration.0.token"),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/myvdir1/myvdir2/"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestMatchResourceAttr(resourceName, names.AttrURI, regexache.MustCompile(`^azure-blob://.+/`)),
				),
			},
			{
				Config: testAccLocationAzureBlobConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLocationAzureBlobExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_tier", "COOL"),
					resource.TestCheckResourceAttr(resourceName, "agent_arns.#", acctest.Ct1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "datasync", regexache.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "SAS"),
					resource.TestCheckResourceAttr(resourceName, "blob_type", "BLOCK"),
					resource.TestCheckResourceAttr(resourceName, "container_url", "https://myaccount.blob.core.windows.net/mycontainer"),
					resource.TestCheckResourceAttr(resourceName, "sas_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "sas_configuration.0.token"),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestMatchResourceAttr(resourceName, names.AttrURI, regexache.MustCompile(`^azure-blob://.+/`)),
				),
			},
		},
	})
}

func testAccCheckLocationAzureBlobDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datasync_location_azure_blob" {
				continue
			}

			_, err := tfdatasync.FindLocationAzureBlobByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DataSync Location Microsoft Azure Blob Storage %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLocationAzureBlobExists(ctx context.Context, n string, v *datasync.DescribeLocationAzureBlobOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncClient(ctx)

		output, err := tfdatasync.FindLocationAzureBlobByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccLocationAzureBlobConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccAgentAgentConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_agent" "test" {
  ip_address = aws_instance.test.public_ip
  name       = %[1]q
}
`, rName))
}

func testAccLocationAzureBlobConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccLocationAzureBlobConfig_base(rName), `
resource "aws_datasync_location_azure_blob" "test" {
  agent_arns          = [aws_datasync_agent.test.arn]
  authentication_type = "SAS"
  container_url       = "https://myaccount.blob.core.windows.net/mycontainer"
  subdirectory        = "/myvdir1/myvdir2"

  sas_configuration {
    token = "sp=r&st=2023-12-20T14:54:52Z&se=2023-12-20T22:54:52Z&spr=https&sv=2021-06-08&sr=c&sig=aBBKDWQvyuVcTPH9EBp%%2FXTI9E%%2F%%2Fmq171%%2BZU178wcwqU%%3D"
  }
}
`)
}

func testAccLocationAzureBlobConfig_tags1(rName, key1, value1 string) string {
	return acctest.ConfigCompose(testAccLocationAzureBlobConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_location_azure_blob" "test" {
  agent_arns          = [aws_datasync_agent.test.arn]
  authentication_type = "SAS"
  container_url       = "https://myaccount.blob.core.windows.net/mycontainer"

  sas_configuration {
    token = "sp=r&st=2023-12-20T14:54:52Z&se=2023-12-20T22:54:52Z&spr=https&sv=2021-06-08&sr=c&sig=aBBKDWQvyuVcTPH9EBp%%2FXTI9E%%2F%%2Fmq171%%2BZU178wcwqU%%3D"
  }

  tags = {
    %[1]q = %[2]q
  }
}
`, key1, value1))
}

func testAccLocationAzureBlobConfig_tags2(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(testAccLocationAzureBlobConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_location_azure_blob" "test" {
  agent_arns          = [aws_datasync_agent.test.arn]
  authentication_type = "SAS"
  container_url       = "https://myaccount.blob.core.windows.net/mycontainer"

  sas_configuration {
    token = "sp=r&st=2023-12-20T14:54:52Z&se=2023-12-20T22:54:52Z&spr=https&sv=2021-06-08&sr=c&sig=aBBKDWQvyuVcTPH9EBp%%2FXTI9E%%2F%%2Fmq171%%2BZU178wcwqU%%3D"
  }

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, key1, value1, key2, value2))
}

func testAccLocationAzureBlobConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccLocationAzureBlobConfig_base(rName), `
resource "aws_datasync_location_azure_blob" "test" {
  access_tier         = "COOL"
  agent_arns          = [aws_datasync_agent.test.arn]
  authentication_type = "SAS"
  container_url       = "https://myaccount.blob.core.windows.net/mycontainer"
  subdirectory        = "/"

  sas_configuration {
    token = "sp=r&st=2023-12-20T14:54:52Z&se=2023-12-20T22:54:52Z&spr=https&sv=2021-06-08&sr=c&sig=aBBKDWQvyuVcTPH9EBp%%2FXTI9E%%2F%%2Fmq171%%2BZU178wcwqU%%3D"
  }
}
`)
}
