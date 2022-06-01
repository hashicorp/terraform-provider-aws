package redshift_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/redshift"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccRedshiftUsageLimit_basic(t *testing.T) {
	resourceName := "aws_redshift_usage_limit.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUsageLimitDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUsageLimitConfig_basic(rName, 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsageLimitExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "feature_type", "concurrency-scaling"),
					resource.TestCheckResourceAttr(resourceName, "limit_type", "time"),
					resource.TestCheckResourceAttr(resourceName, "amount", "60"),
					resource.TestCheckResourceAttr(resourceName, "breach_action", "log"),
					resource.TestCheckResourceAttr(resourceName, "period", "monthly"),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_identifier", "aws_redshift_cluster.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUsageLimitConfig_basic(rName, 120),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsageLimitExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "feature_type", "concurrency-scaling"),
					resource.TestCheckResourceAttr(resourceName, "limit_type", "time"),
					resource.TestCheckResourceAttr(resourceName, "amount", "120"),
					resource.TestCheckResourceAttr(resourceName, "breach_action", "log"),
					resource.TestCheckResourceAttr(resourceName, "period", "monthly"),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_identifier", "aws_redshift_cluster.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccRedshiftUsageLimit_tags(t *testing.T) {
	resourceName := "aws_redshift_usage_limit.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUsageLimitDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUsageLimitConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsageLimitExists(resourceName),
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
				Config: testAccUsageLimitConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccUsageLimitConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsageLimitExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccRedshiftUsageLimit_disappears(t *testing.T) {
	resourceName := "aws_redshift_usage_limit.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUsageLimitDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUsageLimitConfig_basic(rName, 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsageLimitExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfredshift.ResourceUsageLimit(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckUsageLimitDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_redshift_usage_limit" {
			continue
		}
		_, err := tfredshift.FindUsageLimitByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Redshift Usage Limit %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckUsageLimitExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Snapshot Copy Grant ID (UsageLimitName) is not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

		_, err := tfredshift.FindUsageLimitByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccUsageLimitConfig_basic(rName string, amount int) string {
	return acctest.ConfigCompose(testAccClusterConfig_basic(rName), fmt.Sprintf(`
resource "aws_redshift_usage_limit" "test" {
  cluster_identifier = aws_redshift_cluster.test.id
  feature_type       = "concurrency-scaling"
  limit_type         = "time"
  amount             = %[1]d
}
`, amount))
}

func testAccUsageLimitConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccClusterConfig_basic(rName), fmt.Sprintf(`
resource "aws_redshift_usage_limit" "test" {
  cluster_identifier = aws_redshift_cluster.test.id
  feature_type       = "concurrency-scaling"
  limit_type         = "time"
  amount             = 60

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccUsageLimitConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccClusterConfig_basic(rName), fmt.Sprintf(`
resource "aws_redshift_usage_limit" "test" {
  cluster_identifier = aws_redshift_cluster.test.id
  feature_type       = "concurrency-scaling"
  limit_type         = "time"
  amount             = 60

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}
