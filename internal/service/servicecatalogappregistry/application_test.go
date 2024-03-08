// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalogappregistry_test

// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.

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
	// using the services/servicecatalogappregistry/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// types.<Type Name>.
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	servicecatalogappregistry_sdkv2 "github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	servicecatalogappregistry "github.com/hashicorp/terraform-provider-aws/internal/service/servicecatalogappregistry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// TIP: ==== ACCEPTANCE TESTS ====
// This is an example of a basic acceptance test. This should test as much of
// standard functionality of the resource as possible, and test importing, if
// applicable. We prefix its name with "TestAcc", the service, and the
// resource name.
//
// Acceptance test access AWS and cost money to run.
func TestAccServiceCatalogAppRegistryApplication_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_servicecatalog_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogAppRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "servicecatalog", regexache.MustCompile(`/applications/+.`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccServiceCatalogAppRegistryApplication_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_servicecatalog_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogAppRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, servicecatalogappregistry.ResourceApplication, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckApplicationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogAppRegistryClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicecatalog_application" {
				continue
			}

			_, err := conn.GetApplication(ctx, &servicecatalogappregistry_sdkv2.GetApplicationInput{
				Application: aws.String(rs.Primary.Attributes["name"]),
			})

			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingDestroyed, servicecatalogappregistry.ResNameApplication, rs.Primary.ID, err)
			}

			return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingDestroyed, servicecatalogappregistry.ResNameApplication, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckApplicationExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingExistence, servicecatalogappregistry.ResNameApplication, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingExistence, servicecatalogappregistry.ResNameApplication, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogAppRegistryClient(ctx)

		_, err := servicecatalogappregistry.FindApplicationByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccApplicationConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_application" "test" {
  name = %[1]q
}
`, name)
}
