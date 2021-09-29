package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSEksClustersDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceResourceName := "data.aws_eks_clusters.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		ErrorCheck:   testAccErrorCheck(t, eks.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksClustersDataSourceConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrGreaterThanValue(dataSourceResourceName, "names.#", "0"),
				),
			},
		},
	})
}

func testAccAWSEksClustersDataSourceConfig_Basic(rName string) string {
	return composeConfig(
		testAccAWSEksClusterConfig_Required(rName), `
data "aws_eks_clusters" "test" {
  depends_on = [aws_eks_cluster.test]
}
`)
}
