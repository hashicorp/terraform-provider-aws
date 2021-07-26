package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSEksClustersDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceResourceName := "data.aws_eks_clusters.clusters"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		ErrorCheck:   testAccErrorCheck(t, eks.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksClustersDataSourceConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceResourceName, "clusters.#", "1"),
					resource.TestCheckResourceAttr(dataSourceResourceName, "clusters.0", rName),
				),
			},
		},
	})
}

func TestAccAWSEksClustersDataSource_empty(t *testing.T) {
	dataSourceResourceName := "data.aws_eks_clusters.clusters"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		ErrorCheck:   testAccErrorCheck(t, eks.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksClustersDataSourceConfig_empty(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceResourceName, "clusters.#", "0"),
				),
			},
		},
	})
}

func testAccAWSEksClustersDataSourceConfig_Basic(rName1 string) string {
	return composeConfig(
		testAccAWSEksClusterConfig_Required(rName1), `
data "aws_eks_clusters" "clusters" {
	depends_on = [aws_eks_cluster.test]
}
`)
}

func testAccAWSEksClustersDataSourceConfig_empty() string {
	return `
data "aws_eks_clusters" "clusters" {}
`
}
