package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSSSMParameter_importBasic(t *testing.T) {
	resourceName := "aws_ssm_parameter.foo"
	randName := acctest.RandString(5)
	randValue := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMParameterBasicConfig(randName, "String", randValue),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
		},
	})
}
