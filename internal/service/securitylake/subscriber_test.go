// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitylake_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecuritylake "github.com/hashicorp/terraform-provider-aws/internal/service/securitylake"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccSubscriber_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_subscriber.test"
	var subscriber types.SubscriberResource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriberConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriberExists(ctx, resourceName, &subscriber),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccSubscriber_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var subscriber types.SubscriberResource
	resourceName := "aws_securitylake_subscriber.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriberConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriberExists(ctx, resourceName, &subscriber),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfsecuritylake.ResourceSubscriber, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSubscriber_customLogSource(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_subscriber.test"
	var subscriber types.SubscriberResource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriberCustomLogConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriberExists(ctx, resourceName, &subscriber),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccCheckSubscriberDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityLakeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securitylake_subscriber" {
				continue
			}

			_, err := tfsecuritylake.FindSubscriberByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Security Lake Subscriber %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSubscriberExists(ctx context.Context, n string, v *types.SubscriberResource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityLakeClient(ctx)

		output, err := tfsecuritylake.FindSubscriberByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccSubscriberConfig_basic(rName string) string {
	return fmt.Sprintf(`
	
resource "aws_securitylake_subscriber" "test" {
	subscriber_name = %[1]q
	sources {
	aws_log_source_resource {
			source_name    = "ROUTE53"
			source_version = "1.0"
		}
	}
	subscriber_identity {
		external_id = "windows-sysmon-test"
		principal   = "568227374639"
	}
}
`, rName)
}

func testAccSubscriberCustomLogConfig_basic(rName string) string {
	return fmt.Sprintf(`
	
resource "aws_securitylake_subscriber" "test" {
	subscriber_name = %[1]q
	// access_types 	= ["LAKEFORMATION"]
	subscriber_description = "Testing"
	sources {
		custom_log_source_resource {
			source_name    = "windows-sysmon"
			source_version = "1.0"
		}
	}
	subscriber_identity {
		external_id = "windows-sysmon-test"
		principal   = "568227374639"
	}
}
`, rName)
}
