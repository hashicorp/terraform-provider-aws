package emrcontainers_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/eks"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEMRContainersVirtualClusterDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	dataSourceResourceName := "data.aws_emrcontainers_virtual_cluster.test"
	resourceName := "aws_emrcontainers_virtual_cluster.test"
	testExternalProviders := map[string]resource.ExternalProvider{
		"kubernetes": {
			Source:            "hashicorp/kubernetes",
			VersionConstraint: "~> 2.3",
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckIAMServiceLinkedRole(t, "/aws-service-role/emr-containers.amazonaws.com")
		},
		ErrorCheck:        acctest.ErrorCheck(t, eks.EndpointsID),
		Providers:         acctest.Providers,
		ExternalProviders: testExternalProviders,
		CheckDestroy:      testAccCheckVirtualClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualClusterDataSourceConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceResourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceResourceName, "container_provider.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "container_provider.0.id", dataSourceResourceName, "container_provider.0.id"),
					resource.TestCheckResourceAttr(dataSourceResourceName, "container_provider.0.info.#", "1"),
					resource.TestCheckResourceAttr(dataSourceResourceName, "container_provider.0.info.0.eks_info.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "container_provider.0.info.0.eks_info.0.namespace", dataSourceResourceName, "container_provider.0.info.0.eks_info.0.namespace"),
					resource.TestCheckResourceAttrPair(resourceName, "container_provider.0.type", dataSourceResourceName, "container_provider.0.type"),
					resource.TestCheckResourceAttrPair(resourceName, "created_at", dataSourceResourceName, "created_at"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceResourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "state", dataSourceResourceName, "state"),
				),
			},
		},
	})
}

func testAccVirtualClusterDataSourceConfig_Basic(rName string) string {
	return acctest.ConfigCompose(testAccVirtualClusterBasicConfig(rName), `
data "aws_emrcontainers_virtual_cluster" "test" {
  virtual_cluster_id = aws_emrcontainers_virtual_cluster.test.id
}
`)
}
