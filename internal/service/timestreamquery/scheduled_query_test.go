// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestreamquery_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/timestreamquery"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"

	tftimestreamquery "github.com/hashicorp/terraform-provider-aws/internal/service/timestreamquery"
)

func TestAccTimestreamQueryScheduledQuery_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var scheduledquery timestreamquery.DescribeScheduledQueryResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreamquery_scheduled_query.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TimestreamQueryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamQueryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledQueryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledQueryConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledQueryExists(ctx, resourceName, &scheduledquery),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					// TIP: If the ARN can be partially or completely determined by the parameters passed, e.g. it contains the
					// value of `rName`, either include the values in the regex or check for an exact match using `acctest.CheckResourceAttrRegionalARN`
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "timestreamquery", regexache.MustCompile(`scheduledquery:.+$`)),
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

func TestAccTimestreamQueryScheduledQuery_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var scheduledquery timestreamquery.DescribeScheduledQueryResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreamquery_scheduled_query.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TimestreamQueryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamQueryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledQueryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledQueryConfig_basic(rName, testAccScheduledQueryVersionNewer),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledQueryExists(ctx, resourceName, &scheduledquery),
					// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
					// but expects a new resource factory function as the third argument. To expose this
					// private function to the testing package, you may need to add a line like the following
					// to exports_test.go:
					//
					//   var ResourceScheduledQuery = newResourceScheduledQuery
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tftimestreamquery.ResourceScheduledQuery, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckScheduledQueryDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).TimestreamQueryClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_timestreamquery_scheduled_query" {
				continue
			}

			// TIP: ==== FINDERS ====
			// The find function should be exported. Since it won't be used outside of the package, it can be exported
			// in the `exports_test.go` file.
			_, err := tftimestreamquery.FindScheduledQueryByARN(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.TimestreamQuery, create.ErrActionCheckingDestroyed, tftimestreamquery.ResNameScheduledQuery, rs.Primary.ID, err)
			}

			return create.Error(names.TimestreamQuery, create.ErrActionCheckingDestroyed, tftimestreamquery.ResNameScheduledQuery, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckScheduledQueryExists(ctx context.Context, name string, scheduledquery *timestreamquery.DescribeScheduledQueryResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.TimestreamQuery, create.ErrActionCheckingExistence, tftimestreamquery.ResNameScheduledQuery, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.TimestreamQuery, create.ErrActionCheckingExistence, tftimestreamquery.ResNameScheduledQuery, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TimestreamQueryClient(ctx)

		resp, err := tftimestreamquery.FindScheduledQueryByARN(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.TimestreamQuery, create.ErrActionCheckingExistence, tftimestreamquery.ResNameScheduledQuery, rs.Primary.ID, err)
		}

		*scheduledquery = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).TimestreamQueryClient(ctx)

	input := &timestreamquery.ListScheduledQueriesInput{}

	_, err := conn.ListScheduledQueries(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckScheduledQueryNotRecreated(before, after *timestreamquery.DescribeScheduledQueryResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.ScheduledQueryId), aws.ToString(after.ScheduledQueryId); before != after {
			return create.Error(names.TimestreamQuery, create.ErrActionCheckingNotRecreated, tftimestreamquery.ResNameScheduledQuery, aws.ToString(before.ScheduledQueryId), errors.New("recreated"))
		}

		return nil
	}
}

func testAccScheduledQueryConfig_basic(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_timestreamquery_scheduled_query" "test" {
  scheduled_query_name             = %[1]q
  engine_type             = "ActiveTimestreamQuery"
  engine_version          = %[2]q
  host_instance_type      = "timestreamquery.t2.micro"
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
