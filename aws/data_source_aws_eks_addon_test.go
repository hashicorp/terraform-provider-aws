package aws

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/eks"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAWSEksAddonDataSource_basic(t *testing.T) {
	var addon eks.Addon
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	dataSourceResourceName := "data.aws_eks_addon.test"
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"
	ctx := context.TODO()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckAWSEks(t); testAccPreCheckAWSEksAddon(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAWSEksAddonDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksAddonDataSourceConfig_Basic(rName, addonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, dataSourceResourceName, &addon),
					acctest.MatchResourceAttrRegionalARN(dataSourceResourceName, "arn", "eks", regexp.MustCompile(fmt.Sprintf("addon/%s/%s/.+$", rName, addonName))),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "addon_version", dataSourceResourceName, "addon_version"),
					resource.TestCheckResourceAttrPair(resourceName, "service_account_role_arn", dataSourceResourceName, "service_account_role_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_at", dataSourceResourceName, "created_at"),
					resource.TestCheckResourceAttrPair(resourceName, "modified_at", dataSourceResourceName, "modified_at"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceResourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccAWSEksAddonDataSourceConfig_Basic(rName, addonName string) string {
	return acctest.ConfigCompose(testAccAWSEksAddonConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_addon" "test" {
  addon_name   = %[2]q
  cluster_name = aws_eks_cluster.test.name
}

data "aws_eks_addon" "test" {
  addon_name   = %[2]q
  cluster_name = aws_eks_cluster.test.name

  depends_on = [
    aws_eks_addon.test,
    aws_eks_cluster.test,
  ]
}
`, rName, addonName))
}
