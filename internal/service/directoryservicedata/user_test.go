// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directoryservicedata_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdirectoryservicedata "github.com/hashicorp/terraform-provider-aws/internal/service/directoryservicedata"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDirectoryServiceDataUser_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName(t)
	samAccountName := fmt.Sprintf(
		"user%s",
		acctest.RandStringFromCharSet(t, 16, "abcdefghijklmnopqrstuvwxyz0123456789"),
	)
	resourceName := "aws_directoryservicedata_user.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectoryServiceDataServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(rName, domainName, samAccountName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "directory_id", "aws_directory_service_directory.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "sam_account_name", samAccountName),
					resource.TestCheckResourceAttr(resourceName, "email_address", fmt.Sprintf("%s@example.com", samAccountName)),
					resource.TestCheckResourceAttr(resourceName, "given_name", "Test"),
					resource.TestCheckResourceAttr(resourceName, "surname", "User"),
					resource.TestCheckResourceAttrSet(resourceName, "distinguished_name"),
					resource.TestCheckResourceAttrSet(resourceName, "enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "sid"),
					resource.TestCheckResourceAttrSet(resourceName, "user_principal_name"),
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

func TestAccDirectoryServiceDataUser_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName(t)
	samAccountName := fmt.Sprintf(
		"u%s",
		acctest.RandStringFromCharSet(t, 16, "abcdefghijklmnopqrstuvwxyz0123456789"),
	)
	resourceName := "aws_directoryservicedata_user.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectoryServiceDataServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(rName, domainName, samAccountName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdirectoryservicedata.ResourceUser, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckUserDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DirectoryServiceDataClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_directoryservicedata_user" {
				continue
			}

			_, err := tfdirectoryservicedata.FindUserByTwoPartKey(ctx, conn, rs.Primary.Attributes["directory_id"], rs.Primary.Attributes["sam_account_name"])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.DirectoryServiceData, create.ErrActionCheckingDestroyed, tfdirectoryservicedata.ResNameUser, rs.Primary.ID, err)
			}

			return create.Error(names.DirectoryServiceData, create.ErrActionCheckingDestroyed, tfdirectoryservicedata.ResNameUser, rs.Primary.ID, errors.New("not destroyed"))
		}
		return nil
	}
}

func testAccCheckUserExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DirectoryServiceData, create.ErrActionCheckingExistence, tfdirectoryservicedata.ResNameUser, name, fmt.Errorf("not found"))
		}

		conn := acctest.ProviderMeta(ctx, t).DirectoryServiceDataClient(ctx)
		_, err := tfdirectoryservicedata.FindUserByTwoPartKey(ctx, conn, rs.Primary.Attributes["directory_id"], rs.Primary.Attributes["sam_account_name"])

		if err != nil {
			return create.Error(names.DirectoryServiceData, create.ErrActionCheckingExistence, tfdirectoryservicedata.ResNameUser, rs.Primary.ID, err)
		}
		return nil
	}
}

func testAccUserConfig_basic(rName, domainName, samAccountName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_directoryservicedata_user" "test" {
  directory_id     = aws_directory_service_directory.test.id
  sam_account_name = %[2]q
  email_address    = "%[2]s@example.com"
  given_name       = "Test"
  surname          = "User"
}
`, domainName, samAccountName),
	)
}
