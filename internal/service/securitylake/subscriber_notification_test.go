package securitylake_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSecurityLakeSubscriberNotification_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	// var subscribernotification securitylake.CreateSubscriberNotificationOutput
	resourceName := "aws_securitylake_subscriber_notification.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// acctest.PreCheckPartitionHasService(t, names.SecurityLakeEndpointID)
			// testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		// CheckDestroy:             testAccCheckSubscriberNotificationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriberNotificationConfig_basic(),
				Check:  resource.ComposeTestCheckFunc(
				// testAccCheckSubscriberNotificationExists(ctx, resourceName, &subscribernotification),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configuration"},
			},
		},
	})
}

// func TestAccSecurityLakeSubscriberNotification_disappears(t *testing.T) {
// 	ctx := acctest.Context(t)
// 	if testing.Short() {
// 		t.Skip("skipping long-running test in short mode")
// 	}

// 	var subscribernotification securitylake.DescribeSubscriberNotificationResponse
// 	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
// 	resourceName := "aws_securitylake_subscriber_notification.test"

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck: func() {
// 			acctest.PreCheck(ctx, t)
// 			acctest.PreCheckPartitionHasService(t, names.SecurityLakeEndpointID)
// 			testAccPreCheck(t)
// 		},
// 		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeEndpointID),
// 		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
// 		CheckDestroy:             testAccCheckSubscriberNotificationDestroy(ctx),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccSubscriberNotificationConfig_basic(rName, testAccSubscriberNotificationVersionNewer),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckSubscriberNotificationExists(ctx, resourceName, &subscribernotification),
// 					// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
// 					// but expects a new resource factory function as the third argument. To expose this
// 					// private function to the testing package, you may need to add a line like the following
// 					// to exports_test.go:
// 					//
// 					//   var ResourceSubscriberNotification = newResourceSubscriberNotification
// 					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfsecuritylake.ResourceSubscriberNotification, resourceName),
// 				),
// 				ExpectNonEmptyPlan: true,
// 			},
// 		},
// 	})
// }

func testAccSubscriberNotificationConfig_basic() string {
	return `

resource "aws_securitylake_subscriber_notification" "test" {
	subscriber_id = "5ac67093-c109-486b-b12b-0cd1903cda30"

	configuration {
		sqs_notification_configuration {
			enable = true
		}
	}
}
`
}
