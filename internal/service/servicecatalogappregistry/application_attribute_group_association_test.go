// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalogappregistry_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"
	tfservicecatalogappregistry "github.com/hashicorp/terraform-provider-aws/internal/service/servicecatalogappregistry"
)

func TestApplicationAttributeGroupAssociationExampleUnitTest(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Input    string
		Expected string
		Error    bool
	}{
		{
			TestName: "empty",
			Input:    "",
			Expected: "",
			Error:    true,
		},
		{
			TestName: "descriptive name",
			Input:    "some input",
			Expected: "some output",
			Error:    false,
		},
		{
			TestName: "another descriptive name",
			Input:    "more input",
			Expected: "more output",
			Error:    false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()
			got, err := tfservicecatalogappregistry.FunctionFromResource(testCase.Input)

			if err != nil && !testCase.Error {
				t.Errorf("got error (%s), expected no error", err)
			}

			if err == nil && testCase.Error {
				t.Errorf("got (%s) and no error, expected error", got)
			}

			if got != testCase.Expected {
				t.Errorf("got %s, expected %s", got, testCase.Expected)
			}
		})
	}
}

func TestAccServiceCatalogAppRegistryApplicationAttributeGroupAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var applicationattributegroupassociation servicecatalogappregistry.DescribeApplicationAttributeGroupAssociationResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_servicecatalogappregistry_application_attribute_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceCatalogAppRegistryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogAppRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationAttributeGroupAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationAttributeGroupAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationAttributeGroupAssociationExists(ctx, resourceName, &applicationattributegroupassociation),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "servicecatalogappregistry", regexache.MustCompile(`applicationattributegroupassociation:+.`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccServiceCatalogAppRegistryApplicationAttributeGroupAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var applicationattributegroupassociation servicecatalogappregistry.DescribeApplicationAttributeGroupAssociationResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_servicecatalogappregistry_application_attribute_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceCatalogAppRegistryEndpointID)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogAppRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationAttributeGroupAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationAttributeGroupAssociationConfig_basic(rName, testAccApplicationAttributeGroupAssociationVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationAttributeGroupAssociationExists(ctx, resourceName, &applicationattributegroupassociation),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfservicecatalogappregistry.ResourceApplicationAttributeGroupAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckApplicationAttributeGroupAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogAppRegistryClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicecatalogappregistry_application_attribute_group_association" {
				continue
			}

			input := &servicecatalogappregistry.DescribeApplicationAttributeGroupAssociationInput{
				ApplicationAttributeGroupAssociationId: aws.String(rs.Primary.ID),
			}
			_, err := conn.DescribeApplicationAttributeGroupAssociation(ctx, &servicecatalogappregistry.DescribeApplicationAttributeGroupAssociationInput{
				ApplicationAttributeGroupAssociationId: aws.String(rs.Primary.ID),
			})
			if errs.IsA[*types.ResourceNotFoundException](err){
				return nil
			}
			if err != nil {
			        return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingDestroyed, tfservicecatalogappregistry.ResNameApplicationAttributeGroupAssociation, rs.Primary.ID, err)
			}

			return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingDestroyed, tfservicecatalogappregistry.ResNameApplicationAttributeGroupAssociation, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckApplicationAttributeGroupAssociationExists(ctx context.Context, name string, applicationattributegroupassociation *servicecatalogappregistry.DescribeApplicationAttributeGroupAssociationResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingExistence, tfservicecatalogappregistry.ResNameApplicationAttributeGroupAssociation, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingExistence, tfservicecatalogappregistry.ResNameApplicationAttributeGroupAssociation, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogAppRegistryClient(ctx)
		resp, err := conn.DescribeApplicationAttributeGroupAssociation(ctx, &servicecatalogappregistry.DescribeApplicationAttributeGroupAssociationInput{
			ApplicationAttributeGroupAssociationId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingExistence, tfservicecatalogappregistry.ResNameApplicationAttributeGroupAssociation, rs.Primary.ID, err)
		}

		*applicationattributegroupassociation = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogAppRegistryClient(ctx)

	input := &servicecatalogappregistry.ListApplicationAttributeGroupAssociationsInput{}
	_, err := conn.ListApplicationAttributeGroupAssociations(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckApplicationAttributeGroupAssociationNotRecreated(before, after *servicecatalogappregistry.DescribeApplicationAttributeGroupAssociationResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.ApplicationAttributeGroupAssociationId), aws.ToString(after.ApplicationAttributeGroupAssociationId); before != after {
			return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingNotRecreated, tfservicecatalogappregistry.ResNameApplicationAttributeGroupAssociation, aws.ToString(before.ApplicationAttributeGroupAssociationId), errors.New("recreated"))
		}

		return nil
	}
}

func testAccApplicationAttributeGroupAssociationConfig_basic(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_servicecatalogappregistry_application_attribute_group_association" "test" {
  application_attribute_group_association_name             = %[1]q
  engine_type             = "ActiveServiceCatalogAppRegistry"
  engine_version          = %[2]q
  host_instance_type      = "servicecatalogappregistry.t2.micro"
  security_groups         = [aws_security_group.test.id]
  authentication_strategy = "simple"
  storage_type            = "efs"

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }
}
`, rName, version)
}
