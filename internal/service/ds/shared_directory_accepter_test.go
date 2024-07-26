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
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDSSharedDirectoryAccepter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_directory_service_shared_directory_accepter.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccSharedDirectoryAccepterConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSharedDirectoryAccepterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "method", string(awstypes.ShareMethodHandshake)),
					resource.TestCheckResourceAttr(resourceName, "notes", "There were hints and allegations"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrOwnerAccountID, "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckResourceAttrSet(resourceName, "owner_directory_id"),
					resource.TestCheckResourceAttrSet(resourceName, "shared_directory_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"notes",
				},
			},
		},
	})
}

func testAccCheckSharedDirectoryAccepterExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DSClient(ctx)

		_, err := tfds.FindSharedDirectoryByTwoPartKey(ctx, conn, rs.Primary.Attributes["owner_directory_id"], rs.Primary.Attributes["shared_directory_id"])

		return err
	}
}

func testAccSharedDirectoryAccepterConfig_basic(rName, domain string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		testAccDirectoryConfig_microsoftStandard(rName, domain),
		`
data "aws_caller_identity" "current" {}

resource "aws_directory_service_shared_directory" "test" {
  directory_id = aws_directory_service_directory.test.id
  notes        = "There were hints and allegations"

  target {
    id = data.aws_caller_identity.consumer.account_id
  }
}

data "aws_caller_identity" "consumer" {
  provider = "awsalternate"
}

resource "aws_directory_service_shared_directory_accepter" "test" {
  provider = "awsalternate"

  shared_directory_id = aws_directory_service_shared_directory.test.shared_directory_id
}
`)
}
