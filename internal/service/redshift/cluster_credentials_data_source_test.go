package redshift_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/redshift"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccRedshiftClusterCredentialsDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_redshift_cluster_credentials.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterCredentialsDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_identifier", "aws_redshift_cluster.test", "cluster_identifier"),
					resource.TestCheckResourceAttrSet(dataSourceName, "db_password"),
					resource.TestCheckResourceAttrSet(dataSourceName, "expiration"),
				),
			},
		},
	})
}

func testAccClusterCredentialsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier = %[1]q

  database_name       = "testdb"
  master_username     = "foo"
  master_password     = "Password1"
  node_type           = "dc2.large"
  cluster_type        = "single-node"
  skip_final_snapshot = true
}

data "aws_redshift_cluster_credentials" "test" {
  cluster_identifier = aws_redshift_cluster.test.cluster_identifier
  db_user            = aws_redshift_cluster.test.master_username
}
`, rName)
}
