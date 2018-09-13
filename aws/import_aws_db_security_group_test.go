package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSDBSecurityGroup_importBasic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	rName := fmt.Sprintf("tf-acc-%s", acctest.RandString(5))
	resourceName := "aws_db_security_group.bar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBSecurityGroupConfig(rName),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
