package inspector_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/inspector"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccInspectorRulesPackagesDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, inspector.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSInspectorRulesPackagesConfig,
				Check:  resource.TestCheckResourceAttrSet("data.aws_inspector_rules_packages.test", "arns.#"),
			},
		},
	})
}

const testAccCheckAWSInspectorRulesPackagesConfig = `
data "aws_inspector_rules_packages" "test" {}
`
