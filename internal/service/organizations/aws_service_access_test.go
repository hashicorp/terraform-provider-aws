// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package organizations_test

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the service/organizations/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
	"context"
	"errors"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"

	// TIP: You will often need to import the package that this test file lives
	// in. Since it is in the "test" context, it must import the package to use
	// any normal context constants, variables, or functions.
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
)

// TIP: File Structure. The basic outline for all test files should be as
// follows. Improve this resource's maintainability by following this
// outline.
//
// 1. Package declaration (add "_test" since this is a test file)
// 2. Imports
// 3. Unit tests
// 4. Basic test
// 5. Disappears test
// 6. All the other tests
// 7. Helper functions (exists, destroy, check, etc.)
// 8. Functions that return Terraform configurations

// TIP: ==== ACCEPTANCE TESTS ====
// This is an example of a basic acceptance test. This should test as much of
// standard functionality of the resource as possible, and test importing, if
// applicable. We prefix its name with "TestAcc", the service, and the
// resource name.
//
// Acceptance tests access AWS and cost money to run.
func TestAccOrganizationsAwsServiceAccess_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var awsserviceaccess awstypes.EnabledServicePrincipal
	servicePrincipal := "tagpolicies.tag.amazonaws.com"
	resourceName := "aws_organizations_aws_service_access.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAwsServiceAccessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsServiceAccessConfig_basic(servicePrincipal),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsServiceAccessExists(ctx, t, resourceName, &awsserviceaccess),
					resource.TestCheckResourceAttr(resourceName, "service_principal", servicePrincipal),
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

func TestAccOrganizationsAwsServiceAccess_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var awsserviceaccess awstypes.EnabledServicePrincipal
	servicePrincipal := "tagpolicies.tag.amazonaws.com"
	resourceName := "aws_organizations_aws_service_access.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAwsServiceAccessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsServiceAccessConfig_basic(servicePrincipal),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsServiceAccessExists(ctx, t, resourceName, &awsserviceaccess),
					// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
					// but expects a new resource factory function as the third argument. To expose this
					// private function to the testing package, you may need to add a line like the following
					// to exports_test.go:
					//
					//   var ResourceAwsServiceAccess = newAwsServiceAccessResource
					acctest.CheckFrameworkResourceDisappears(ctx, t, tforganizations.ResourceAwsServiceAccess, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckAwsServiceAccessDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).OrganizationsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_organizations_aws_service_access" {
				continue
			}

			_, err := tforganizations.FindAwsServiceAccessByServicePrincipal(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Organizations, create.ErrActionCheckingDestroyed, tforganizations.ResNameAwsServiceAccess, rs.Primary.ID, err)
			}

			return create.Error(names.Organizations, create.ErrActionCheckingDestroyed, tforganizations.ResNameAwsServiceAccess, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAwsServiceAccessExists(ctx context.Context, t *testing.T, name string, awsserviceaccess *awstypes.EnabledServicePrincipal) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Organizations, create.ErrActionCheckingExistence, tforganizations.ResNameAwsServiceAccess, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Organizations, create.ErrActionCheckingExistence, tforganizations.ResNameAwsServiceAccess, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).OrganizationsClient(ctx)

		resp, err := tforganizations.FindAwsServiceAccessByServicePrincipal(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.Organizations, create.ErrActionCheckingExistence, tforganizations.ResNameAwsServiceAccess, rs.Primary.ID, err)
		}

		*awsserviceaccess = *resp

		return nil
	}
}

func testAccAwsServiceAccessConfig_basic(servicePrincipal string) string {
	return fmt.Sprintf(`
resource "aws_organizations_aws_service_access" "test" {
  service_principal             = %[1]q
}
`, servicePrincipal)
}
