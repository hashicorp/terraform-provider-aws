package neptune_test

import (
	//"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfneptune "github.com/hashicorp/terraform-provider-aws/internal/service/neptune"
)

func TestAccNeptuneClusterEndpoint_basic(t *testing.T) {
	var dbCluster neptune.DBClusterEndpoint
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_neptune_cluster_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterEndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterEndpointExists(resourceName, &dbCluster),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(`cluster-endpoint:.+`)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "READER"),
					resource.TestCheckResourceAttr(resourceName, "cluster_endpoint_identifier", rName),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_identifier", "aws_neptune_cluster.test", "cluster_identifier"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "static_members.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "excluded_members.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNeptuneClusterEndpoint_tags(t *testing.T) {
	if acctest.Partition() == "aws-us-gov" {
		t.Skip("Neptune Cluster Endpoint tags are not supported in GovCloud partition")
	}

	var v neptune.DBClusterEndpoint
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_neptune_cluster_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterEndpointConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterEndpointConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccClusterEndpointConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccNeptuneClusterEndpoint_disappears(t *testing.T) {
	var dbCluster neptune.DBClusterEndpoint
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_neptune_cluster_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterEndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterEndpointExists(resourceName, &dbCluster),
					acctest.CheckResourceDisappears(acctest.Provider, tfneptune.ResourceClusterEndpoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccNeptuneClusterEndpoint_Disappears_cluster(t *testing.T) {
	var dbCluster neptune.DBClusterEndpoint
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_neptune_cluster_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterEndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterEndpointExists(resourceName, &dbCluster),
					acctest.CheckResourceDisappears(acctest.Provider, tfneptune.ResourceCluster(), "aws_neptune_cluster.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckClusterEndpointDestroy(s *terraform.State) error {
	return testAccCheckClusterEndpointDestroyWithProvider(s, acctest.Provider)
}

func testAccCheckClusterEndpointDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*conns.AWSClient).NeptuneConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_neptune_cluster_endpoint" {
			continue
		}

		_, err := tfneptune.FindEndpointByID(conn, rs.Primary.ID)
		// Return nil if the cluster is already destroyed
		if err != nil {
			if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBClusterNotFoundFault) {
				return nil
			}
		}

		return err
	}

	return nil
}

func testAccCheckClusterEndpointExists(n string, v *neptune.DBClusterEndpoint) resource.TestCheckFunc {
	return testAccCheckClusterEndpointExistsWithProvider(n, v, func() *schema.Provider { return acctest.Provider })
}

func testAccCheckClusterEndpointExistsWithProvider(n string, v *neptune.DBClusterEndpoint, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Neptune Instance ID is set")
		}

		provider := providerF()
		conn := provider.Meta().(*conns.AWSClient).NeptuneConn
		resp, err := tfneptune.FindEndpointByID(conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Neptune Cluster Endpoint (%s) not found: %w", rs.Primary.ID, err)
		}

		*v = *resp

		return nil
	}
}

func testAccClusterEndpointBaseConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
locals {
  availability_zone_names = slice(data.aws_availability_zones.available.names, 0, min(3, length(data.aws_availability_zones.available.names)))
}

resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  availability_zones                   = local.availability_zone_names
  engine                               = "neptune"
  neptune_cluster_parameter_group_name = "default.neptune1"
  skip_final_snapshot                  = true
}
`, rName))
}

func testAccClusterEndpointConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccClusterEndpointBaseConfig(rName), fmt.Sprintf(`
resource "aws_neptune_cluster_endpoint" "test" {
  cluster_identifier          = aws_neptune_cluster.test.cluster_identifier
  cluster_endpoint_identifier = %[1]q
  endpoint_type               = "READER"
}
`, rName))
}

func testAccClusterEndpointConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccClusterEndpointBaseConfig(rName), fmt.Sprintf(`
resource "aws_neptune_cluster_endpoint" "test" {
  cluster_identifier          = aws_neptune_cluster.test.cluster_identifier
  cluster_endpoint_identifier = %[1]q
  endpoint_type               = "READER"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccClusterEndpointConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccClusterEndpointBaseConfig(rName), fmt.Sprintf(`
resource "aws_neptune_cluster_endpoint" "test" {
  cluster_identifier          = aws_neptune_cluster.test.cluster_identifier
  cluster_endpoint_identifier = %[1]q
  endpoint_type               = "READER"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
