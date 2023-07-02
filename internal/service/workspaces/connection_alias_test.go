package workspaces_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfworkspaces "github.com/hashicorp/terraform-provider-aws/internal/service/workspaces"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWorkSpacesConnectionAlias_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var connectionalias workspaces.ConnectionAlias
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_workspaces_connection_alias.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, workspaces.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, workspaces.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionAliasConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionAliasExists(ctx, resourceName, &connectionalias),
					resource.TestCheckResourceAttr(resourceName, "connec", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "workspaces", regexp.MustCompile(`connectionalias:+.`)),
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

func TestAccWorkSpacesConnectionAlias_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var connectionalias workspaces.ConnectionAlias
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_workspaces_connection_alias.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, workspaces.EndpointsID)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, workspaces.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionAliasConfig_basic(rName, testAccConnectionAliasVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionAliasExists(ctx, resourceName, &connectionalias),
					// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
					// but expects a new resource factory function as the third argument. To expose this
					// private function to the testing package, you may need to add a line like the following
					// to exports_test.go:
					//
					//   var ResourceConnectionAlias = newResourceConnectionAlias
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfworkspaces.ResourceConnectionAlias, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckConnectionAliasDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkSpacesConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workspaces_connection_alias" {
				continue
			}

			_, err := tfworkspaces.FindConnectionAliasByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				return nil
			}

			if err != nil {
				return nil
			}

			return create.Error(names.WorkSpaces, create.ErrActionCheckingDestroyed, tfworkspaces.ResNameConnectionAlias, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckConnectionAliasExists(ctx context.Context, name string, connectionalias *workspaces.ConnectionAlias) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.WorkSpaces, create.ErrActionCheckingExistence, tfworkspaces.ResNameConnectionAlias, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.WorkSpaces, create.ErrActionCheckingExistence, tfworkspaces.ResNameConnectionAlias, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkSpacesConn(ctx)
		out, err := tfworkspaces.FindConnectionAliasByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.WorkSpaces, create.ErrActionCheckingExistence, tfworkspaces.ResNameConnectionAlias, rs.Primary.ID, err)
		}

		*connectionalias = *out

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).WorkSpacesConn(ctx)

	input := &workspaces.DescribeConnectionAliasesInput{}
	_, err := conn.DescribeConnectionAliasesWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckConnectionAliasNotRecreated(before, after *workspaces.ConnectionAlias) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.AliasId), aws.StringValue(after.AliasId); before != after {
			return create.Error(names.WorkSpaces, create.ErrActionCheckingNotRecreated, tfworkspaces.ResNameConnectionAlias, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccConnectionAliasConfig_basic(rName, domain, workspaceName string) string {
	return acctest.ConfigCompose(
		testAccWorkspaceConfig_Prerequisites(rName, domain),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_workspaces_connection_alias" "test" {
  connection_alias_name             = %[1]q
  engine_type             = "ActiveWorkSpaces"
  engine_version          = %[2]q
  host_instance_type      = "workspaces.t2.micro"
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
`, rName, domain))
}
