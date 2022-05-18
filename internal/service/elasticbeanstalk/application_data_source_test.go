package elasticbeanstalk_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfeks "github.com/hashicorp/terraform-provider-aws/internal/service/eks"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccElasticBeanstalkApplicationDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))
	dataSourceResourceName := "data.aws_elastic_beanstalk_application.test"
	resourceName := "aws_elastic_beanstalk_application.tftest"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticbeanstalk.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEKSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationDataSourceConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceResourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceResourceName, "description"),
					resource.TestCheckResourceAttr(dataSourceResourceName, "appversion_lifecycle.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "appversion_lifecycle.0.service_role", dataSourceResourceName, "appversion_lifecycle.0.service_role"),
					resource.TestCheckResourceAttrPair(resourceName, "appversion_lifecycle.0.max_age_in_days", dataSourceResourceName, "appversion_lifecycle.0.max_age_in_days"),
					resource.TestCheckResourceAttrPair(resourceName, "appversion_lifecycle.0.delete_source_from_s3", dataSourceResourceName, "appversion_lifecycle.0.delete_source_from_s3"),
				),
			},
		},
	})
}

func testAccApplicationDataSourceConfig_Basic(rName string) string {
	return fmt.Sprintf(`
%s

data "aws_elastic_beanstalk_application" "test" {
  name = aws_elastic_beanstalk_application.tftest.name
}
`, testAccBeanstalkAppConfigWithMaxAge(rName))
}

func testAccCheckEKSClusterDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_eks_cluster" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EKSConn

		_, err := tfeks.FindClusterByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EKS Cluster %s still exists", rs.Primary.ID)
	}

	return nil
}
