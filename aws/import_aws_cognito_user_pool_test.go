package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSCognitoUserPool_importBasic(t *testing.T) {
	resourceName := "aws_cognito_user_pool.pool"
	name := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchDashboardDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_basic(name),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
