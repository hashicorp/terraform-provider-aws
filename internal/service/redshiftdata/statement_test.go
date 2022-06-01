package redshiftdata_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/redshiftdataapiservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshiftdata "github.com/hashicorp/terraform-provider-aws/internal/service/redshiftdata"
)

func TestAccRedshiftDataStatement_basic(t *testing.T) {
	var v redshiftdataapiservice.DescribeStatementOutput
	resourceName := "aws_redshiftdata_statement.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshiftdataapiservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccStatementConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStatementExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_identifier", "aws_redshift_cluster.test", "cluster_identifier"),
					resource.TestCheckResourceAttr(resourceName, "sql", "CREATE GROUP group_name;"),
					resource.TestCheckResourceAttr(resourceName, "parameters.#", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"database", "db_user"},
			},
		},
	})
}

func testAccCheckStatementExists(n string, v *redshiftdataapiservice.DescribeStatementOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Redshift Data Statement ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftDataConn

		output, err := tfredshiftdata.FindStatementByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccStatementConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInExclude("usw2-az2"), fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier                  = %[1]q
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
}

resource "aws_redshiftdata_statement" "test" {
  cluster_identifier = aws_redshift_cluster.test.cluster_identifier
  database           = aws_redshift_cluster.test.database_name
  db_user            = aws_redshift_cluster.test.master_username
  sql                = "CREATE GROUP group_name;"
}
`, rName))
}
