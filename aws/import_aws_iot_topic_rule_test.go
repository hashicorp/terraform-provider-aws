package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSIoTTopicRule_importbasic(t *testing.T) {
	resourceName := "aws_iot_topic_rule.rule"
	rName := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_basic(rName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
