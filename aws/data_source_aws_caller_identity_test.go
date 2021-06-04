package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/atest"
)

func TestAccAWSCallerIdentity_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { atest.PreCheck(t) },
		ErrorCheck: atest.ErrorCheck(t, sts.EndpointsID),
		Providers:  atest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsCallerIdentityConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					atest.CheckAwsCallerIdentityAccountId("data.aws_caller_identity.current"),
				),
			},
		},
	})
}

const testAccCheckAwsCallerIdentityConfig_basic = `
data "aws_caller_identity" "current" {}
`
