package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSRoute53DelegationSet_importBasic(t *testing.T) {
	resourceName := "aws_route53_delegation_set.test"
	refName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53DelegationSetConfig(refName),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"reference_name"},
			},
		},
	})
}
