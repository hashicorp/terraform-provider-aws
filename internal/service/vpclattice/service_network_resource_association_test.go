// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"
	tfvpclattice "github.com/hashicorp/terraform-provider-aws/internal/service/vpclattice"
)

func TestServiceNetworkResourceAssociationExampleUnitTest(t *testing.T) {
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
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()
			got, err := tfvpclattice.FunctionFromResource(testCase.Input)

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

func TestAccVPCLatticeServiceNetworkResourceAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var servicenetworkresourceassociation vpclattice.DescribeServiceNetworkResourceAssociationResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_service_network_resource_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceNetworkResourceAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkResourceAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkResourceAssociationExists(ctx, resourceName, &servicenetworkresourceassociation),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "vpclattice", regexache.MustCompile(`servicenetworkresourceassociation:+.`)),
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

func TestAccVPCLatticeServiceNetworkResourceAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var servicenetworkresourceassociation vpclattice.DescribeServiceNetworkResourceAssociationResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_service_network_resource_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceNetworkResourceAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkResourceAssociationConfig_basic(rName, testAccServiceNetworkResourceAssociationVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkResourceAssociationExists(ctx, resourceName, &servicenetworkresourceassociation),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfvpclattice.ResourceServiceNetworkResourceAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckServiceNetworkResourceAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpclattice_service_network_resource_association" {
				continue
			}

			
			_, err := tfvpclattice.FindServiceNetworkResourceAssociationByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
			        return create.Error(names.VPCLattice, create.ErrActionCheckingDestroyed, tfvpclattice.ResNameServiceNetworkResourceAssociation, rs.Primary.ID, err)
			}

			return create.Error(names.VPCLattice, create.ErrActionCheckingDestroyed, tfvpclattice.ResNameServiceNetworkResourceAssociation, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckServiceNetworkResourceAssociationExists(ctx context.Context, name string, servicenetworkresourceassociation *vpclattice.DescribeServiceNetworkResourceAssociationResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameServiceNetworkResourceAssociation, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameServiceNetworkResourceAssociation, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)

		resp, err := tfvpclattice.FindServiceNetworkResourceAssociationByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameServiceNetworkResourceAssociation, rs.Primary.ID, err)
		}

		*servicenetworkresourceassociation = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)

	input := &vpclattice.ListServiceNetworkResourceAssociationsInput{}

	_, err := conn.ListServiceNetworkResourceAssociations(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckServiceNetworkResourceAssociationNotRecreated(before, after *vpclattice.DescribeServiceNetworkResourceAssociationResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.ServiceNetworkResourceAssociationId), aws.ToString(after.ServiceNetworkResourceAssociationId); before != after {
			return create.Error(names.VPCLattice, create.ErrActionCheckingNotRecreated, tfvpclattice.ResNameServiceNetworkResourceAssociation, aws.ToString(before.ServiceNetworkResourceAssociationId), errors.New("recreated"))
		}

		return nil
	}
}

func testAccServiceNetworkResourceAssociationConfig_basic(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_vpclattice_service_network_resource_association" "test" {
  service_network_resource_association_name             = %[1]q
  engine_type             = "ActiveVPCLattice"
  engine_version          = %[2]q
  host_instance_type      = "vpclattice.t2.micro"
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
