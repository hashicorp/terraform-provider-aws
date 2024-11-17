// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"

	// TIP: You will often need to import the package that this test file lives
	// in. Since it is in the "test" context, it must import the package to use
	// any normal context constants, variables, or functions.
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
)

// TIP: ==== ACCEPTANCE TESTS ====
// This is an example of a basic acceptance test. This should test as much of
// standard functionality of the resource as possible, and test importing, if
// applicable. We prefix its name with "TestAcc", the service, and the
// resource name.
//
// Acceptance test access AWS and cost money to run.
func TestAccIAMOrganizationFeatures_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var organizationfeatures iam.ListOrganizationsFeaturesOutput
	resourceName := "aws_iam_organization_features.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationsAWSServiceAccess(ctx, t, "iam.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationFeaturesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationFeaturesConfig_basic([]string{"RootCredentialsManagement", "RootSessions"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationFeaturesExists(ctx, resourceName, &organizationfeatures),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccIAMOrganizationFeatures_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var organizationfeatures iam.ListOrganizationsFeaturesOutput
	resourceName := "aws_iam_organization_features.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationsAWSServiceAccess(ctx, t, "iam.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationFeaturesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationFeaturesConfig_basic([]string{"RootCredentialsManagement", "RootSessions"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationFeaturesExists(ctx, resourceName, &organizationfeatures),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckOrganizationFeaturesDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_organization_features" {
				continue
			}

			out, err := conn.ListOrganizationsFeatures(ctx, &iam.ListOrganizationsFeaturesInput{})
			if err != nil {
				return create.Error(names.IAM, create.ErrActionCheckingDestroyed, tfiam.ResNameOrganizationFeatures, rs.Primary.Attributes["organization_id"], err)
			}
			if len(out.EnabledFeatures) == 0 {
				return nil
			}

			return create.Error(names.IAM, create.ErrActionCheckingDestroyed, tfiam.ResNameOrganizationFeatures, rs.Primary.Attributes["organization_id"], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckOrganizationFeaturesExists(ctx context.Context, name string, organizationfeatures *iam.ListOrganizationsFeaturesOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameOrganizationFeatures, name, errors.New("not found"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)
		resp, err := conn.ListOrganizationsFeatures(ctx, &iam.ListOrganizationsFeaturesInput{})
		if err != nil {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameOrganizationFeatures, rs.Primary.Attributes["organization_id"], err)
		}

		*organizationfeatures = *resp

		return nil
	}
}

func testAccOrganizationFeaturesConfig_basic(features []string) string {
	return fmt.Sprintf(`
resource "aws_iam_organization_features" "test" {
  features = [%[1]s]
}
`, fmt.Sprintf(`"%s"`, strings.Join(features, `", "`)))
}
