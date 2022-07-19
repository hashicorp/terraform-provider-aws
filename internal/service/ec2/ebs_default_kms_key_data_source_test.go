package ec2_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2EBSDefaultKMSKeyDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSDefaultKMSKeyDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSDefaultKMSKey("data.aws_ebs_default_kms_key.current"),
				),
			},
		},
	})
}

const testAccEBSDefaultKMSKeyDataSourceConfig_basic = `
data "aws_ebs_default_kms_key" "current" {}
`
