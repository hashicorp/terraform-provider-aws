// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfds "github.com/hashicorp/terraform-provider-aws/internal/service/ds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDSSharedDirectory_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SharedDirectory
	resourceName := "aws_directory_service_shared_directory.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckSharedDirectoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSharedDirectoryConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSharedDirectoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "method", "HANDSHAKE"),
					resource.TestCheckResourceAttr(resourceName, "notes", "test"),
					resource.TestCheckResourceAttrSet(resourceName, "shared_directory_id"),
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

func TestAccDSSharedDirectory_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SharedDirectory
	resourceName := "aws_directory_service_shared_directory.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckSharedDirectoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSharedDirectoryConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSharedDirectoryExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfds.ResourceSharedDirectory(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSharedDirectoryExists(ctx context.Context, n string, v *awstypes.SharedDirectory) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DSClient(ctx)

		output, err := tfds.FindSharedDirectoryByTwoPartKey(ctx, conn, rs.Primary.Attributes["directory_id"], rs.Primary.Attributes["shared_directory_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckSharedDirectoryDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_directory_service_shared_directory" {
				continue
			}

			_, err := tfds.FindSharedDirectoryByTwoPartKey(ctx, conn, rs.Primary.Attributes["directory_id"], rs.Primary.Attributes["shared_directory_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Directory Service Shared Directory %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccSharedDirectoryConfig_basic(rName, domain string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		testAccDirectoryConfig_microsoftStandard(rName, domain),
		`
resource "aws_directory_service_shared_directory" "test" {
  directory_id = aws_directory_service_directory.test.id
  notes        = "test"

  target {
    id = data.aws_caller_identity.receiver.account_id
  }
}

data "aws_caller_identity" "receiver" {
  provider = "awsalternate"
}
`)
}
