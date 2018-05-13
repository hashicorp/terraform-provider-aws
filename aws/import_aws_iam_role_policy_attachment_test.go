package aws

import (
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"testing"
)

func TestAccAwsIamRolePolicyAttachmentImport(t *testing.T) {
	resourceName := "aws_iam_role_policy_attachment.test-attach"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRolePolicyAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRolePolicyAttachConfig(rInt),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     "test-role",
				ImportStateVerify: true,
			},
		},
	})
}
