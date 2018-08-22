package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSRoute53Record_importBasic(t *testing.T) {
	resourceName := "aws_route53_record.default"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53RecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53RecordConfig,
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", "weight"},
			},
		},
	})
}

func TestAccAWSRoute53Record_importUnderscored(t *testing.T) {
	resourceName := "aws_route53_record.underscore"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53RecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53RecordConfigUnderscoreInName,
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", "weight"},
			},
		},
	})
}
