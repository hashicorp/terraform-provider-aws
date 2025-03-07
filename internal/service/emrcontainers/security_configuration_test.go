// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emrcontainers_test

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
	// using the services/emrcontainers/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// types.<Type Name>.
	"context"
	"errors"
	"fmt"
	awstypes "github.com/aws/aws-sdk-go-v2/service/emrcontainers/types"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/emrcontainers"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"

	// TIP: You will often need to import the package that this test file lives
	// in. Since it is in the "test" context, it must import the package to use
	// any normal context constants, variables, or functions.
	tfemrcontainers "github.com/hashicorp/terraform-provider-aws/internal/service/emrcontainers"
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
// Acceptance test access AWS and cost money to run.
func TestAccEMRContainersSecurityConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var securityconfiguration awstypes.SecurityConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emrcontainers_security_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRContainersServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityConfigurationExists(ctx, resourceName, &securityconfiguration),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrID),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrID,
			},
		},
	})
}

// Disappear test will make sense in the future when (if) delete operation is implemented by AWS

//func TestAccEMRContainersSecurityConfiguration_disappears(t *testing.T) {
//	ctx := acctest.Context(t)
//
//	var securityconfiguration awstypes.SecurityConfiguration
//	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
//	resourceName := "aws_emrcontainers_security_configuration.test"
//
//	resource.ParallelTest(t, resource.TestCase{
//		PreCheck: func() {
//			acctest.PreCheck(ctx, t)
//			testAccPreCheck(ctx, t)
//		},
//		ErrorCheck:               acctest.ErrorCheck(t, names.EMRContainersServiceID),
//		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
//		Steps: []resource.TestStep{
//			{
//				Config: testAccSecurityConfigurationConfig_basic(rName),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					testAccCheckSecurityConfigurationExists(ctx, resourceName, &securityconfiguration),
//					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfemrcontainers.ResourceSecurityConfiguration, resourceName),
//				),
//				ExpectNonEmptyPlan: true,
//			},
//		},
//	})
//}

func testAccCheckSecurityConfigurationExists(ctx context.Context, name string, securityconfiguration *awstypes.SecurityConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EMRContainers, create.ErrActionCheckingExistence, tfemrcontainers.ResNameSecurityConfiguration, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.EMRContainers, create.ErrActionCheckingExistence, tfemrcontainers.ResNameSecurityConfiguration, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EMRContainersClient(ctx)

		resp, err := tfemrcontainers.FindSecurityConfigurationByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.EMRContainers, create.ErrActionCheckingExistence, tfemrcontainers.ResNameSecurityConfiguration, rs.Primary.ID, err)
		}

		*securityconfiguration = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EMRContainersClient(ctx)

	input := &emrcontainers.ListSecurityConfigurationsInput{}

	_, err := conn.ListSecurityConfigurations(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccSecurityConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`

resource "aws_emrcontainers_security_configuration" "test" {
  name = %[1]q

  security_configuration_data {
    authorization_configuration {
      lake_formation_configuration {
		authorized_session_tag_value = "EMR on EKS Engine"
		query_engine_role_arn = "arn:aws:iam::123456789012:role/query-engine-role"
		secure_namespace_info {
		  cluster_id = "test"
		  namespace = "default"
		}
	  }
	  encryption_configuration {
	    in_transit_encryption_configuration {
		  tls_certificate_configuration {
		    certificate_provider_type = "PEM"
		    private_certificate_secret_arn = "arn:aws:secretsmanager:us-west-2:123456789012:secret:tls/certificate/private"
		    public_certificate_secret_arn = "arn:aws:secretsmanager:us-west-2:123456789012:secret:tls/certificate/public"
		  }
	    }
	  }
	}
  }

  tags = {
    Environment = "test"
  }
}
`, rName)
}
