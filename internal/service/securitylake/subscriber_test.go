// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitylake_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/securitylake"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSubscriber_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_subscriber_test"
	// var subscriber types.SubscriberResource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomLogSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriberConfig_basic(rName),
				Check:  resource.ComposeAggregateTestCheckFunc(),
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

// func TestAccSecurityLakeSubscriber_disappears(t *testing.T) {
// 	ctx := acctest.Context(t)
// 	if testing.Short() {
// 		t.Skip("skipping long-running test in short mode")
// 	}

// 	var Subscriber securitylake.DescribeSubscriberResponse
// 	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
// 	resourceName := "aws_securitylake_securitylake_subscriber.test"

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck: func() {
// 			acctest.PreCheck(ctx, t)
// 			acctest.PreCheckPartitionHasService(t, names.SecurityLakeEndpointID)
// 			testAccPreCheck(t)
// 		},
// 		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeEndpointID),
// 		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
// 		CheckDestroy:             testAccCheckSubscriberDestroy(ctx),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccSubscriberConfig_basic(rName, testAccSubscriberVersionNewer),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckSubscriberExists(ctx, resourceName, &Subscriber),
// 					// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
// 					// but expects a new resource factory function as the third argument. To expose this
// 					// private function to the testing package, you may need to add a line like the following
// 					// to exports_test.go:
// 					//
// 					//   var ResourceSubscriber = newResourceSubscriber
// 					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfsecuritylake.ResourceSubscriber, resourceName),
// 				),
// 				ExpectNonEmptyPlan: true,
// 			},
// 		},
// 	})
// }

// func testAccCheckSubscriberDestroy(ctx context.Context) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityLakeClient(ctx)

// 		for _, rs := range s.RootModule().Resources {
// 			if rs.Type != "aws_securitylake_securitylake_subscriber" {
// 				continue
// 			}

// 			input := &securitylake.DescribeSubscriberInput{
// 				SubscriberId: aws.String(rs.Primary.ID),
// 			}
// 			_, err := conn.DescribeSubscriber(ctx, &securitylake.DescribeSubscriberInput{
// 				SubscriberId: aws.String(rs.Primary.ID),
// 			})
// 			if errs.IsA[*types.ResourceNotFoundException](err) {
// 				return nil
// 			}
// 			if err != nil {
// 				return nil
// 			}

// 			return create.Error(names.SecurityLake, create.ErrActionCheckingDestroyed, tfsecuritylake.ResNameSubscriber, rs.Primary.ID, errors.New("not destroyed"))
// 		}

// 		return nil
// 	}
// }

// func testAccCheckSubscriberExists(ctx context.Context, name string, Subscriber *securitylake.DescribeSubscriberResponse) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		rs, ok := s.RootModule().Resources[name]
// 		if !ok {
// 			return create.Error(names.SecurityLake, create.ErrActionCheckingExistence, tfsecuritylake.ResNameSubscriber, name, errors.New("not found"))
// 		}

// 		if rs.Primary.ID == "" {
// 			return create.Error(names.SecurityLake, create.ErrActionCheckingExistence, tfsecuritylake.ResNameSubscriber, name, errors.New("not set"))
// 		}

// 		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityLakeClient(ctx)
// 		resp, err := conn.DescribeSubscriber(ctx, &securitylake.DescribeSubscriberInput{
// 			SubscriberId: aws.String(rs.Primary.ID),
// 		})

// 		if err != nil {
// 			return create.Error(names.SecurityLake, create.ErrActionCheckingExistence, tfsecuritylake.ResNameSubscriber, rs.Primary.ID, err)
// 		}

// 		*Subscriber = *resp

// 		return nil
// 	}
// }

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityLakeClient(ctx)

	input := &securitylake.ListSubscribersInput{}
	_, err := conn.ListSubscribers(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

// func testAccCheckSubscriberNotRecreated(before, after *securitylake.DescribeSubscriberResponse) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		if before, after := aws.ToString(before.SubscriberId), aws.ToString(after.SubscriberId); before != after {
// 			return create.Error(names.SecurityLake, create.ErrActionCheckingNotRecreated, tfsecuritylake.ResNameSubscriber, aws.ToString(before.SubscriberId), errors.New("recreated"))
// 		}

// 		return nil
// 	}
// }

func testAccSubscriberConfig_basic(rName string) string {
	return fmt.Sprintf(`
	
data "aws_caller_identity" "current" {}
resource "aws_securitylake_subscriber" "test" {
	sources {
		aws_log_source_resource {
			source_name = "ROUTE53"
			source_version = "1"
		}
	}
	subscriber_identity {
		external_id = "test-external"
		principal   = data.aws_caller_identity.current.account_id
	}
	subscriber_name = %[1]q
}
`, rName)
}
