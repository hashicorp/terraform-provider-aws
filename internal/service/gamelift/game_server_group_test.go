package gamelift_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccGameLiftGameServerGroup_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_gamelift_game_server_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGameServerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGameServerGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameServerGroupExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`gameservergroup/.+`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "auto_scaling_group_arn", "autoscaling", regexp.MustCompile(`autoScalingGroup:.+`)),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "balancing_strategy", gamelift.BalancingStrategySpotPreferred),
					resource.TestCheckResourceAttr(resourceName, "game_server_protection_policy", gamelift.GameServerProtectionPolicyNoProtection),
					resource.TestCheckResourceAttr(resourceName, "launch_template.0.version", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"vpc_subnets"},
			},
		},
	})
}

func TestAccGameLiftGameServerGroup_AutoScalingPolicy(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_gamelift_game_server_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGameServerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGameServerGroupConfig_autoScalingPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameServerGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_policy.0.estimated_instance_warmup", "60"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_policy.0.target_tracking_configuration.0.target_value", "77.7"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"vpc_subnets"},
			},
		},
	})
}

func TestAccGameLiftGameServerGroup_AutoScalingPolicy_EstimatedInstanceWarmup(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_gamelift_game_server_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGameServerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGameServerGroupConfig_autoScalingPolicyEstimatedInstanceWarmup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameServerGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_policy.0.estimated_instance_warmup", "66"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_policy.0.target_tracking_configuration.0.target_value", "77.7"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"vpc_subnets"},
			},
		},
	})
}

func TestAccGameLiftGameServerGroup_BalancingStrategy(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_gamelift_game_server_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGameServerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGameServerGroupConfig_balancingStrategy(rName, gamelift.BalancingStrategySpotOnly),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameServerGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "balancing_strategy", gamelift.BalancingStrategySpotOnly),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"vpc_subnets"},
			},
		},
	})
}

func TestAccGameLiftGameServerGroup_GameServerGroupName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_gamelift_game_server_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGameServerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGameServerGroupConfig_name(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameServerGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "game_server_group_name", rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"vpc_subnets"},
			},
			{
				Config: testAccGameServerGroupConfig_name(rName, rName+"-new"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameServerGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "game_server_group_name", rName+"-new"),
				),
			},
		},
	})
}

func TestAccGameLiftGameServerGroup_InstanceDefinition(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_gamelift_game_server_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGameServerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGameServerGroupConfig_instanceDefinition(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameServerGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_definition.#", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"vpc_subnets"},
			},
			{
				Config: testAccGameServerGroupConfig_instanceDefinition(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameServerGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_definition.#", "3"),
				),
			},
		},
	})
}

func TestAccGameLiftGameServerGroup_InstanceDefinition_WeightedCapacity(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_gamelift_game_server_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGameServerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGameServerGroupConfig_instanceDefinitionWeightedCapacity(rName, "1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameServerGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_definition.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "instance_definition.0.weighted_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_definition.1.weighted_capacity", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"vpc_subnets"},
			},
			{
				Config: testAccGameServerGroupConfig_instanceDefinitionWeightedCapacity(rName, "2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameServerGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_definition.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "instance_definition.0.weighted_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "instance_definition.1.weighted_capacity", "2"),
				),
			},
		},
	})
}

func TestAccGameLiftGameServerGroup_LaunchTemplate_Id(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_gamelift_game_server_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGameServerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGameServerGroupConfig_launchTemplateID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameServerGroupExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", "aws_launch_template.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "launch_template.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "launch_template.0.version", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"vpc_subnets"},
			},
		},
	})
}

func TestAccGameLiftGameServerGroup_LaunchTemplate_Name(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_gamelift_game_server_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGameServerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGameServerGroupConfig_launchTemplateName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameServerGroupExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", "aws_launch_template.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "launch_template.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "launch_template.0.version", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"vpc_subnets"},
			},
		},
	})
}

func TestAccGameLiftGameServerGroup_LaunchTemplate_Version(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_gamelift_game_server_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGameServerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGameServerGroupConfig_launchTemplateVersion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameServerGroupExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", "aws_launch_template.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "launch_template.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "launch_template.0.version", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"vpc_subnets"},
			},
		},
	})
}

func TestAccGameLiftGameServerGroup_GameServerProtectionPolicy(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_gamelift_game_server_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGameServerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGameServerGroupConfig_protectionPolicy(rName, gamelift.GameServerProtectionPolicyFullProtection),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameServerGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "game_server_protection_policy", gamelift.GameServerProtectionPolicyFullProtection),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"vpc_subnets"},
			},
		},
	})
}

func TestAccGameLiftGameServerGroup_MaxSize(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_gamelift_game_server_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGameServerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGameServerGroupConfig_maxSize(rName, "1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameServerGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "max_size", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"vpc_subnets"},
			},
			{
				Config: testAccGameServerGroupConfig_maxSize(rName, "2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameServerGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "max_size", "2"),
				),
			},
		},
	})
}

func TestAccGameLiftGameServerGroup_MinSize(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_gamelift_game_server_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGameServerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGameServerGroupConfig_minSize(rName, "1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameServerGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "min_size", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"vpc_subnets"},
			},
			{
				Config: testAccGameServerGroupConfig_minSize(rName, "2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameServerGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "min_size", "2"),
				),
			},
		},
	})
}

func TestAccGameLiftGameServerGroup_roleARN(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_gamelift_game_server_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGameServerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGameServerGroupConfig_roleARN(rName, "test1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameServerGroupExists(resourceName),
					acctest.CheckResourceAttrGlobalARN(resourceName, "role_arn", "iam", fmt.Sprintf(`role/%s-test1`, rName)),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test1", "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"vpc_subnets"},
			},
			{
				Config: testAccGameServerGroupConfig_roleARN(rName, "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameServerGroupExists(resourceName),
					acctest.CheckResourceAttrGlobalARN(resourceName, "role_arn", "iam", fmt.Sprintf(`role/%s-test2`, rName)),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test2", "arn"),
				),
			},
		},
	})
}

func TestAccGameLiftGameServerGroup_vpcSubnets(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_gamelift_game_server_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGameServerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGameServerGroupConfig_vpcSubnets(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameServerGroupExists(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"vpc_subnets"},
			},
			{
				Config: testAccGameServerGroupConfig_vpcSubnets(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameServerGroupExists(resourceName),
				),
			},
		},
	})
}

func testAccCheckGameServerGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_gamelift_game_server_group" {
			continue
		}

		input := gamelift.DescribeGameServerGroupInput{
			GameServerGroupName: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeGameServerGroup(&input)

		if tfawserr.ErrCodeEquals(err, gamelift.ErrCodeNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil {
			return fmt.Errorf("GameLift Game Server Group (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckGameServerGroupExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource %s has not set its id", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftConn

		input := gamelift.DescribeGameServerGroupInput{
			GameServerGroupName: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeGameServerGroup(&input)

		if err != nil {
			return fmt.Errorf("error reading GameLift Game Server Group (%s): %w", rs.Primary.ID, err)
		}

		if output == nil {
			return fmt.Errorf("GameLift Game Server Group (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccGameServerGroupConfig_iam(rName string, name string) string {
	return fmt.Sprintf(`
data "aws_partition" %[2]q {}

resource "aws_iam_role" %[2]q {
  assume_role_policy = <<-EOF
    {
      "Version": "2012-10-17",
      "Statement": [
        {
          "Effect": "Allow",
          "Principal": {
            "Service": [
              "autoscaling.amazonaws.com",
              "gamelift.amazonaws.com"
            ]
          },
          "Action": "sts:AssumeRole"
        }
      ]
    }
  EOF
  name = "%[1]s-%[2]s"
}

resource "aws_iam_role_policy_attachment" %[2]q {
  policy_arn = "arn:${data.aws_partition.%[2]s.partition}:iam::aws:policy/GameLiftGameServerGroupPolicy"
  role       = aws_iam_role.%[2]s.name
}
`, rName, name)
}

func testAccGameServerGroupLaunchTemplateConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_vpc" "test" {
  default = true
}

data "aws_subnets" "test" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.test.id]
  }
}

resource "aws_launch_template" "test" {
  image_id = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  name     = %[1]q

  network_interfaces {
    subnet_id = data.aws_subnets.test.ids[0]
  }
}
`, rName)
}

func testAccGameServerGroupInstanceTypeOfferingsConfig() string {
	return `
data "aws_ec2_instance_type_offerings" "available" {
  filter {
    name   = "instance-type"
    values = ["c5a.large", "c5a.2xlarge", "c5.large", "c5.2xlarge", "m4.large", "m4.2xlarge", "m5a.large", "m5a.2xlarge", "m5.large", "m5.2xlarge"]
  }
}
`
}

func testAccGameServerGroupConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccGameServerGroupConfig_iam(rName, "test"),
		testAccGameServerGroupInstanceTypeOfferingsConfig(),
		testAccGameServerGroupLaunchTemplateConfig(rName),
		fmt.Sprintf(`
resource "aws_gamelift_game_server_group" "test" {
  game_server_group_name = %[1]q
  dynamic "instance_definition" {
    for_each = data.aws_ec2_instance_type_offerings.available.instance_types
    content {
      instance_type = instance_definition.value
    }
  }
  launch_template {
    id = aws_launch_template.test.id
  }

  max_size = 1
  min_size = 1
  role_arn = aws_iam_role.test.arn

  vpc_subnets = [data.aws_subnets.test.ids[0]]

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccGameServerGroupConfig_autoScalingPolicy(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccGameServerGroupConfig_iam(rName, "test"),
		testAccGameServerGroupInstanceTypeOfferingsConfig(),
		testAccGameServerGroupLaunchTemplateConfig(rName),
		fmt.Sprintf(`
resource "aws_gamelift_game_server_group" "test" {
  auto_scaling_policy {
    target_tracking_configuration {
      target_value = 77.7
    }
  }
  game_server_group_name = %[1]q
  dynamic "instance_definition" {
    for_each = data.aws_ec2_instance_type_offerings.available.instance_types
    content {
      instance_type = instance_definition.value
    }
  }
  launch_template {
    id = aws_launch_template.test.id
  }

  max_size = 1
  min_size = 1
  role_arn = aws_iam_role.test.arn

  vpc_subnets = [data.aws_subnets.test.ids[0]]

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccGameServerGroupConfig_autoScalingPolicyEstimatedInstanceWarmup(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccGameServerGroupConfig_iam(rName, "test"),
		testAccGameServerGroupInstanceTypeOfferingsConfig(),
		testAccGameServerGroupLaunchTemplateConfig(rName),
		fmt.Sprintf(`
resource "aws_gamelift_game_server_group" "test" {
  auto_scaling_policy {
    estimated_instance_warmup = 66
    target_tracking_configuration {
      target_value = 77.7
    }
  }
  game_server_group_name = %[1]q
  dynamic "instance_definition" {
    for_each = data.aws_ec2_instance_type_offerings.available.instance_types
    content {
      instance_type = instance_definition.value
    }
  }
  launch_template {
    id = aws_launch_template.test.id
  }

  max_size = 1
  min_size = 1
  role_arn = aws_iam_role.test.arn

  vpc_subnets = [data.aws_subnets.test.ids[0]]

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccGameServerGroupConfig_balancingStrategy(rName string, balancingStrategy string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccGameServerGroupConfig_iam(rName, "test"),
		testAccGameServerGroupInstanceTypeOfferingsConfig(),
		testAccGameServerGroupLaunchTemplateConfig(rName),
		fmt.Sprintf(`
resource "aws_gamelift_game_server_group" "test" {
  balancing_strategy     = %[2]q
  game_server_group_name = %[1]q
  dynamic "instance_definition" {
    for_each = data.aws_ec2_instance_type_offerings.available.instance_types
    content {
      instance_type = instance_definition.value
    }
  }
  launch_template {
    id = aws_launch_template.test.id
  }

  max_size = 1
  min_size = 1
  role_arn = aws_iam_role.test.arn

  vpc_subnets = [data.aws_subnets.test.ids[0]]

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, balancingStrategy))
}

func testAccGameServerGroupConfig_name(rName string, gameServerGroupName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccGameServerGroupConfig_iam(rName, "test"),
		testAccGameServerGroupInstanceTypeOfferingsConfig(),
		testAccGameServerGroupLaunchTemplateConfig(rName),
		fmt.Sprintf(`
resource "aws_gamelift_game_server_group" "test" {
  game_server_group_name = %[1]q
  dynamic "instance_definition" {
    for_each = data.aws_ec2_instance_type_offerings.available.instance_types
    content {
      instance_type = instance_definition.value
    }
  }
  launch_template {
    id = aws_launch_template.test.id
  }

  max_size = 1
  min_size = 1
  role_arn = aws_iam_role.test.arn

  vpc_subnets = [data.aws_subnets.test.ids[0]]

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, gameServerGroupName))
}

func testAccGameServerGroupConfig_instanceDefinition(rName string, count int) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccGameServerGroupConfig_iam(rName, "test"),
		testAccGameServerGroupInstanceTypeOfferingsConfig(),
		testAccGameServerGroupLaunchTemplateConfig(rName),
		fmt.Sprintf(`
resource "aws_gamelift_game_server_group" "test" {
  game_server_group_name = %[1]q
  dynamic "instance_definition" {
    for_each = slice(sort(tolist(data.aws_ec2_instance_type_offerings.available.instance_types)), 0, %[2]d)
    content {
      instance_type = instance_definition.value
    }
  }
  launch_template {
    id = aws_launch_template.test.id
  }

  max_size = 1
  min_size = 1
  role_arn = aws_iam_role.test.arn

  vpc_subnets = [data.aws_subnets.test.ids[0]]

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, count))
}

func testAccGameServerGroupConfig_instanceDefinitionWeightedCapacity(rName string, weightedCapacity string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccGameServerGroupConfig_iam(rName, "test"),
		testAccGameServerGroupInstanceTypeOfferingsConfig(),
		testAccGameServerGroupLaunchTemplateConfig(rName),
		fmt.Sprintf(`
resource "aws_gamelift_game_server_group" "test" {
  game_server_group_name = %[1]q
  dynamic "instance_definition" {
    for_each = slice(sort(tolist(data.aws_ec2_instance_type_offerings.available.instance_types)), 0, 2)
    content {
      instance_type     = instance_definition.value
      weighted_capacity = %[2]q
    }
  }
  launch_template {
    id = aws_launch_template.test.id
  }

  max_size = 1
  min_size = 1
  role_arn = aws_iam_role.test.arn

  vpc_subnets = [data.aws_subnets.test.ids[0]]

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, weightedCapacity))
}

func testAccGameServerGroupConfig_launchTemplateID(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccGameServerGroupConfig_iam(rName, "test"),
		testAccGameServerGroupInstanceTypeOfferingsConfig(),
		testAccGameServerGroupLaunchTemplateConfig(rName),
		fmt.Sprintf(`
resource "aws_gamelift_game_server_group" "test" {
  game_server_group_name = %[1]q
  dynamic "instance_definition" {
    for_each = data.aws_ec2_instance_type_offerings.available.instance_types
    content {
      instance_type = instance_definition.value
    }
  }
  launch_template {
    id = aws_launch_template.test.id
  }

  max_size = 1
  min_size = 1
  role_arn = aws_iam_role.test.arn

  vpc_subnets = [data.aws_subnets.test.ids[0]]

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccGameServerGroupConfig_launchTemplateName(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccGameServerGroupConfig_iam(rName, "test"),
		testAccGameServerGroupInstanceTypeOfferingsConfig(),
		testAccGameServerGroupLaunchTemplateConfig(rName),
		fmt.Sprintf(`
resource "aws_gamelift_game_server_group" "test" {
  game_server_group_name = %[1]q
  dynamic "instance_definition" {
    for_each = data.aws_ec2_instance_type_offerings.available.instance_types
    content {
      instance_type = instance_definition.value
    }
  }
  launch_template {
    name = aws_launch_template.test.name
  }

  max_size = 1
  min_size = 1
  role_arn = aws_iam_role.test.arn

  vpc_subnets = [data.aws_subnets.test.ids[0]]

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccGameServerGroupConfig_launchTemplateVersion(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccGameServerGroupConfig_iam(rName, "test"),
		testAccGameServerGroupInstanceTypeOfferingsConfig(),
		testAccGameServerGroupLaunchTemplateConfig(rName),
		fmt.Sprintf(`
resource "aws_gamelift_game_server_group" "test" {
  game_server_group_name = %[1]q
  dynamic "instance_definition" {
    for_each = data.aws_ec2_instance_type_offerings.available.instance_types
    content {
      instance_type = instance_definition.value
    }
  }
  launch_template {
    id      = aws_launch_template.test.id
    version = 1
  }

  max_size = 1
  min_size = 1
  role_arn = aws_iam_role.test.arn

  vpc_subnets = [data.aws_subnets.test.ids[0]]

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccGameServerGroupConfig_maxSize(rName string, maxSize string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccGameServerGroupConfig_iam(rName, "test"),
		testAccGameServerGroupInstanceTypeOfferingsConfig(),
		testAccGameServerGroupLaunchTemplateConfig(rName),
		fmt.Sprintf(`
resource "aws_gamelift_game_server_group" "test" {
  game_server_group_name = %[1]q
  dynamic "instance_definition" {
    for_each = data.aws_ec2_instance_type_offerings.available.instance_types
    content {
      instance_type = instance_definition.value
    }
  }
  launch_template {
    id = aws_launch_template.test.id
  }

  max_size = %[2]s
  min_size = 1
  role_arn = aws_iam_role.test.arn

  vpc_subnets = [data.aws_subnets.test.ids[0]]

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, maxSize))
}

func testAccGameServerGroupConfig_minSize(rName string, minSize string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccGameServerGroupConfig_iam(rName, "test"),
		testAccGameServerGroupInstanceTypeOfferingsConfig(),
		testAccGameServerGroupLaunchTemplateConfig(rName),
		fmt.Sprintf(`
resource "aws_gamelift_game_server_group" "test" {
  game_server_group_name = %[1]q
  dynamic "instance_definition" {
    for_each = data.aws_ec2_instance_type_offerings.available.instance_types
    content {
      instance_type = instance_definition.value
    }
  }
  launch_template {
    id = aws_launch_template.test.id
  }

  max_size = 2
  min_size = %[2]s
  role_arn = aws_iam_role.test.arn

  vpc_subnets = [data.aws_subnets.test.ids[0]]

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, minSize))
}

func testAccGameServerGroupConfig_protectionPolicy(rName string, gameServerProtectionPolicy string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccGameServerGroupConfig_iam(rName, "test"),
		testAccGameServerGroupInstanceTypeOfferingsConfig(),
		testAccGameServerGroupLaunchTemplateConfig(rName),
		fmt.Sprintf(`
resource "aws_gamelift_game_server_group" "test" {
  game_server_group_name        = %[1]q
  game_server_protection_policy = %[2]q
  dynamic "instance_definition" {
    for_each = data.aws_ec2_instance_type_offerings.available.instance_types
    content {
      instance_type = instance_definition.value
    }
  }
  launch_template {
    id = aws_launch_template.test.id
  }

  max_size = 1
  min_size = 1
  role_arn = aws_iam_role.test.arn

  vpc_subnets = [data.aws_subnets.test.ids[0]]

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, gameServerProtectionPolicy))
}

func testAccGameServerGroupConfig_roleARN(rName string, roleArn string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccGameServerGroupConfig_iam(rName, roleArn),
		testAccGameServerGroupInstanceTypeOfferingsConfig(),
		testAccGameServerGroupLaunchTemplateConfig(rName),
		fmt.Sprintf(`
resource "aws_gamelift_game_server_group" "test" {
  game_server_group_name = %[1]q
  dynamic "instance_definition" {
    for_each = data.aws_ec2_instance_type_offerings.available.instance_types
    content {
      instance_type = instance_definition.value
    }
  }
  launch_template {
    id = aws_launch_template.test.id
  }

  max_size = 1
  min_size = 1
  role_arn = aws_iam_role.%[2]s.arn

  vpc_subnets = [data.aws_subnets.test.ids[0]]

  depends_on = [aws_iam_role_policy_attachment.%[2]s]
}
`, rName, roleArn))
}

func testAccGameServerGroupConfig_vpcSubnets(rName string, count int) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccGameServerGroupConfig_iam(rName, "test"),
		testAccGameServerGroupInstanceTypeOfferingsConfig(),
		testAccGameServerGroupLaunchTemplateConfig(rName),
		fmt.Sprintf(`
resource "aws_gamelift_game_server_group" "test" {
  game_server_group_name = %[1]q
  dynamic "instance_definition" {
    for_each = data.aws_ec2_instance_type_offerings.available.instance_types
    content {
      instance_type = instance_definition.value
    }
  }
  launch_template {
    id = aws_launch_template.test.id
  }

  max_size    = 1
  min_size    = 1
  role_arn    = aws_iam_role.test.arn
  vpc_subnets = slice(data.aws_subnets.test.ids, 0, %[2]d)

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, count))
}
