// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directoryservicedata_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/smarterr"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservicedata/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdirectoryservicedata "github.com/hashicorp/terraform-provider-aws/internal/service/directoryservicedata"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDirectoryServiceDataGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName(t)
	samAccountName := fmt.Sprintf(
		"%s%s",
		acctest.ResourcePrefix,
		acctest.RandStringFromCharSet(t, 64-len(acctest.ResourcePrefix), "abcdefghijklmnopqrstuvwxyz0123456789"),
	)
	resourceName := "aws_directoryservicedata_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectoryServiceDataServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName, domainName, samAccountName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "directory_id", "aws_directory_service_directory.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "sam_account_name", samAccountName),
					resource.TestCheckResourceAttrSet(resourceName, "distinguished_name"),
					resource.TestCheckResourceAttrSet(resourceName, "group_scope"),
					resource.TestCheckResourceAttrSet(resourceName, "group_type"),
					resource.TestCheckResourceAttrSet(resourceName, "sid"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, flex.ResourceIdSeparator, "directory_id", "sam_account_name"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "directory_id",
			},
		},
	})
}

func TestAccDirectoryServiceDataGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName(t)
	samAccountName := fmt.Sprintf(
		"%s%s",
		acctest.ResourcePrefix,
		acctest.RandStringFromCharSet(t, 64-len(acctest.ResourcePrefix), "abcdefghijklmnopqrstuvwxyz0123456789"),
	)
	resourceName := "aws_directoryservicedata_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectoryServiceDataServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName, domainName, samAccountName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdirectoryservicedata.ResourceGroup, resourceName),
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

func TestAccDirectoryServiceDataGroup_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName(t)
	samAccountName := fmt.Sprintf(
		"%s%s",
		acctest.ResourcePrefix,
		acctest.RandStringFromCharSet(t, 64-len(acctest.ResourcePrefix), "abcdefghijklmnopqrstuvwxyz0123456789"),
	)
	resourceName := "aws_directoryservicedata_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectoryServiceDataServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_updated(rName, domainName, samAccountName, "Global", "Security"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "group_scope", "Global"),
					resource.TestCheckResourceAttr(resourceName, "group_type", "Security"),
				),
			},
			{
				Config: testAccGroupConfig_updated(rName, domainName, samAccountName, "Universal", "Security"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "group_scope", "Universal"),
					resource.TestCheckResourceAttr(resourceName, "group_type", "Security"),
				),
			},
			{
				Config: testAccGroupConfig_updated(rName, domainName, samAccountName, "Universal", "Distribution"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "group_scope", "Universal"),
					resource.TestCheckResourceAttr(resourceName, "group_type", "Distribution"),
				),
			},
			{
				Config: testAccGroupConfig_updated(rName, domainName, samAccountName, "DomainLocal", "Security"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "group_scope", "DomainLocal"),
					resource.TestCheckResourceAttr(resourceName, "group_type", "Security"),
				),
			},
			{
				Config: testAccGroupConfig_updated(rName, domainName, samAccountName, "Universal", ""),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "group_scope", "Universal"),
					resource.TestCheckResourceAttr(resourceName, "group_type", "Security"),
				),
			},
			{
				Config: testAccGroupConfig_updated(rName, domainName, samAccountName, "", "Distribution"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "group_scope", "Universal"),
					resource.TestCheckResourceAttr(resourceName, "group_type", "Distribution"),
				),
			},
		},
	})
}

func testAccCheckGroupDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DirectoryServiceDataClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_directoryservicedata_group" {
				continue
			}

			_, err := tfdirectoryservicedata.FindGroupByTwoPartKey(ctx, conn, rs.Primary.Attributes["directory_id"], rs.Primary.Attributes["sam_account_name"])
			if retry.NotFound(err) || errs.IsA[*awstypes.AccessDeniedException](err) {
				continue
			}
			if err != nil {
				return smarterr.NewError(err)
			}

			return smarterr.NewError(errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckGroupExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return smarterr.NewError(errors.New("not found"))
		}

		conn := acctest.ProviderMeta(ctx, t).DirectoryServiceDataClient(ctx)
		_, err := tfdirectoryservicedata.FindGroupByTwoPartKey(ctx, conn, rs.Primary.Attributes["directory_id"], rs.Primary.Attributes["sam_account_name"])

		if err != nil {
			return smarterr.NewError(err)
		}
		return nil
	}
}

func testAccGroupConfig_basic(rName, domainName, samAccountName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  name                         = %[1]q
  password                     = "SuperSecretPassw0rd"
  type                         = "MicrosoftAD"
  edition                      = "Standard"
  enable_directory_data_access = true

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_directoryservicedata_group" "test" {
  directory_id     = aws_directory_service_directory.test.id
  sam_account_name = %[2]q
}
`, domainName, samAccountName),
	)
}

func testAccGroupConfig_updated(rName, domainName, samAccountName, groupScope, groupType string) string {
	groupScopeConfig, groupTypeConfig := "", ""
	if groupScope != "" {
		groupScopeConfig = fmt.Sprintf("group_scope = %q", groupScope)
	}
	if groupType != "" {
		groupTypeConfig = fmt.Sprintf("group_type = %q", groupType)
	}

	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  name                         = %[1]q
  password                     = "SuperSecretPassw0rd"
  type                         = "MicrosoftAD"
  edition                      = "Standard"
  enable_directory_data_access = true

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_directoryservicedata_group" "test" {
  directory_id     = aws_directory_service_directory.test.id
  sam_account_name = %[2]q
  %[3]s
  %[4]s
}
		`, domainName, samAccountName, groupScopeConfig, groupTypeConfig),
	)
}
