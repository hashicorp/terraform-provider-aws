package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccNetworkInsightsAnalysis_basic(t *testing.T) {
	resourceName := "aws_ec2_network_insights_analysis.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNetworkInsightsAnalysisDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEC2NetworkInsightsAnalysisConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkInsightsAnalysisExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`network-insights-analysis/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "filter_in_arns.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "network_insights_path_id", "aws_ec2_network_insights_path.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "path_found", "true"),
					acctest.CheckResourceAttrRFC3339(resourceName, "start_date"),
					resource.TestCheckResourceAttr(resourceName, "status", "succeeded"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "wait_for_completion", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_completion"},
			},
		},
	})
}

func TestAccNetworkInsightsAnalysis_disappears(t *testing.T) {
	resourceName := "aws_ec2_network_insights_analysis.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNetworkInsightsAnalysisDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEC2NetworkInsightsAnalysisConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInsightsAnalysisExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceNetworkInsightsAnalysis(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccNetworkInsightsAnalysis_tags(t *testing.T) {
	resourceName := "aws_ec2_network_insights_analysis.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNetworkInsightsAnalysisDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEC2NetworkInsightsAnalysisConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInsightsAnalysisExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_completion"},
			},
			{
				Config: testAccEC2NetworkInsightsAnalysisConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInsightsAnalysisExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccEC2NetworkInsightsAnalysisConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInsightsAnalysisExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccNetworkInsightsAnalysis_filterInARNs(t *testing.T) {
	resourceName := "aws_ec2_network_insights_analysis.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNetworkInsightsAnalysisDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEC2NetworkInsightsAnalysisFilterInARNsConfig(rName, "vpc-peering-connection/pcx-fakearn1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInsightsAnalysisExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "filter_in_arns.0", "ec2", regexp.MustCompile(`vpc-peering-connection/pcx-fakearn1$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_completion"},
			},
			{
				Config: testAccEC2NetworkInsightsAnalysisFilterInARNsConfig(rName, "vpc-peering-connection/pcx-fakearn2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInsightsAnalysisExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "filter_in_arns.0", "ec2", regexp.MustCompile(`vpc-peering-connection/pcx-fakearn2$`)),
				),
			},
		},
	})
}

func TestAccNetworkInsightsAnalysis_waitForCompletion(t *testing.T) {
	resourceName := "aws_ec2_network_insights_analysis.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNetworkInsightsAnalysisDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEC2NetworkInsightsAnalysisWaitForCompletionConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInsightsAnalysisExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "wait_for_completion", "false"),
					resource.TestCheckResourceAttr(resourceName, "status", "running"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_completion"},
			},
			{
				Config: testAccEC2NetworkInsightsAnalysisWaitForCompletionConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInsightsAnalysisExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "wait_for_completion", "true"),
				),
			},
		},
	})
}

func testAccCheckNetworkInsightsAnalysisExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Network Insights Analysis ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		_, err := tfec2.FindNetworkInsightsAnalysisByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckNetworkInsightsAnalysisDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_network_insights_analysis" {
			continue
		}

		_, err := tfec2.FindNetworkInsightsAnalysisByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 Network Insights Analysis %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccEC2NetworkInsightsAnalysisConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test_source" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test_destination" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_network_insights_path" "test" {
  source      = aws_network_interface.test_source.id
  destination = aws_network_interface.test_destination.id
  protocol    = "tcp"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_network_insights_analysis" "test" {
  network_insights_path_id = aws_ec2_network_insights_path.test.id
}
`, rName)
}

func testAccEC2NetworkInsightsAnalysisFilterInARNsConfig(rName, arnSuffix string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "aws_partition" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test_source" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test_destination" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_network_insights_path" "test" {
  source      = aws_network_interface.test_source.id
  destination = aws_network_interface.test_destination.id
  protocol    = "tcp"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_network_insights_analysis" "test" {
  network_insights_path_id = aws_ec2_network_insights_path.test.id
  filter_in_arns           = ["arn:${data.aws_partition.current.partition}:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.id}:%v"]
}

`, rName, arnSuffix)
}

func testAccEC2NetworkInsightsAnalysisConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test_source" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test_destination" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_network_insights_path" "test" {
  source      = aws_network_interface.test_source.id
  destination = aws_network_interface.test_destination.id
  protocol    = "tcp"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_network_insights_analysis" "test" {
  network_insights_path_id = aws_ec2_network_insights_path.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccEC2NetworkInsightsAnalysisConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test_source" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test_destination" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_network_insights_path" "test" {
  source      = aws_network_interface.test_source.id
  destination = aws_network_interface.test_destination.id
  protocol    = "tcp"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_network_insights_analysis" "test" {
  network_insights_path_id = aws_ec2_network_insights_path.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccEC2NetworkInsightsAnalysisWaitForCompletionConfig(rName string, waitForCompletion bool) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test_source" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test_destination" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_network_insights_path" "test" {
  source      = aws_network_interface.test_source.id
  destination = aws_network_interface.test_destination.id
  protocol    = "tcp"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_network_insights_analysis" "test" {
  network_insights_path_id = aws_ec2_network_insights_path.test.id
  wait_for_completion      = %t
}
`, rName, waitForCompletion)
}
