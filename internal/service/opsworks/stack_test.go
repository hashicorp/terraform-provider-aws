package opsworks_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/opsworks"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfopsworks "github.com/hashicorp/terraform-provider-aws/internal/service/opsworks"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccOpsWorksStack_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_stack.test"
	var v opsworks.Stack

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t)
			testAccPreCheckStacks(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "agent_version"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "opsworks", regexp.MustCompile(`stack/.+/`)),
					resource.TestCheckResourceAttr(resourceName, "berkshelf_version", "3.2.0"),
					resource.TestCheckResourceAttr(resourceName, "color", ""),
					resource.TestCheckResourceAttr(resourceName, "configuration_manager_name", "Chef"),
					resource.TestCheckResourceAttr(resourceName, "configuration_manager_version", "11.10"),
					resource.TestCheckResourceAttr(resourceName, "custom_cookbooks_source.#", "1"),
					resource.TestCheckNoResourceAttr(resourceName, "custom_json"),
					resource.TestCheckResourceAttrPair(resourceName, "default_availability_zone", "data.aws_availability_zones.available", "names.0"),
					resource.TestCheckResourceAttrSet(resourceName, "default_instance_profile_arn"),
					resource.TestCheckResourceAttr(resourceName, "default_os", "Ubuntu 12.04 LTS"),
					resource.TestCheckResourceAttr(resourceName, "default_root_device_type", "instance-store"),
					resource.TestCheckResourceAttr(resourceName, "default_ssh_key_name", ""),
					resource.TestCheckResourceAttrPair(resourceName, "default_subnet_id", "aws_subnet.test.0", "id"),
					resource.TestCheckResourceAttr(resourceName, "hostname_theme", "Layer_Dependent"),
					resource.TestCheckResourceAttr(resourceName, "manage_berkshelf", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "region", acctest.Region()),
					resource.TestCheckResourceAttrSet(resourceName, "service_role_arn"),
					resource.TestCheckResourceAttr(resourceName, "stack_endpoint", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "use_custom_cookbooks", "false"),
					resource.TestCheckResourceAttr(resourceName, "use_opsworks_security_groups", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", "aws_vpc.test", "id"),
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

func TestAccOpsWorksStack_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_stack.test"
	var v opsworks.Stack

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t)
			testAccPreCheckStacks(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfopsworks.ResourceStack(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccOpsWorksStack_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_stack.test"
	var v opsworks.Stack

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t)
			testAccPreCheckStacks(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &v),
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
				Config: testAccStackConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccStackConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccOpsWorksStack_tagsAlternateRegion(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_stack.test"
	var v opsworks.Stack

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t)
			testAccPreCheckStacks(t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
			acctest.PreCheckAlternateRegionIs(t, endpoints.UsWest1RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(t, 2),
		CheckDestroy:             testAccCheckStackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_tags1AlternateRegion(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &v),
					resource.TestCheckResourceAttrWith(resourceName, "arn", func(value string) error {
						if !regexp.MustCompile(arn.ARN{
							Partition: acctest.Partition(),
							Service:   opsworks.ServiceName,
							Region:    acctest.AlternateRegion(),
							AccountID: acctest.AccountID(),
							Resource:  `stack/.+/`,
						}.String()).MatchString(value) {
							return fmt.Errorf("%s doesn't match ARN pattern", value)
						}

						return nil
					}),
					resource.TestCheckResourceAttr(resourceName, "region", acctest.AlternateRegion()),
					// "In this case, the actual API endpoint of the stack is in us-east-1."
					resource.TestCheckResourceAttr(resourceName, "stack_endpoint", endpoints.UsEast1RegionID),
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
				Config: testAccStackConfig_tags2AlternateRegion(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccStackConfig_tags1AlternateRegion(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

///////////////////////////////
//// Tests for the No-VPC case
///////////////////////////////

func TestAccOpsWorksStack_noVPCBasic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_stack.test"
	var v opsworks.Stack

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t)
			testAccPreCheckStacks(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_noVPCCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &v),
					testAccCheckCreateStackAttributes(rName),
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

func TestAccOpsWorksStack_noVPCChangeServiceRoleForceNew(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_stack.test"
	var v1, v2 opsworks.Stack

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t)
			testAccPreCheckStacks(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_noVPCCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &v1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStackConfig_noVPCCreateUpdateServiceRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &v2),
					testAccCheckStackRecreated(t, &v1, &v2),
				),
			},
		},
	})
}

func TestAccOpsWorksStack_vpc(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_stack.test"
	var v opsworks.Stack

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t)
			testAccPreCheckStacks(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_vpcCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &v),
					testAccCheckCreateStackAttributes(rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStackConfig_vpcUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "default_availability_zone", "data.aws_availability_zones.available", "names.0"),
					resource.TestCheckResourceAttr(resourceName, "default_os", "Amazon Linux 2015.09"),
					resource.TestCheckResourceAttr(resourceName, "default_root_device_type", "ebs"),
					resource.TestCheckResourceAttr(resourceName, "custom_json", customJSON),
					resource.TestCheckResourceAttr(resourceName, "configuration_manager_version", "11.10"),
					resource.TestCheckResourceAttr(resourceName, "use_opsworks_security_groups", "false"),
					resource.TestCheckResourceAttr(resourceName, "use_custom_cookbooks", "true"),
					resource.TestCheckResourceAttr(resourceName, "manage_berkshelf", "true"),
					resource.TestCheckResourceAttr(resourceName, "custom_cookbooks_source.0.type", "git"),
					resource.TestCheckResourceAttr(resourceName, "custom_cookbooks_source.0.revision", "master"),
					resource.TestCheckResourceAttr(resourceName, "custom_cookbooks_source.0.url", "https://github.com/aws/opsworks-example-cookbooks.git"),
				),
			},
		},
	})
}

/////////////////////////////
// Tests for Custom Cookbooks
/////////////////////////////

func TestAccOpsWorksStack_CustomCookbooks_setPrivateProperties(t *testing.T) {
	resourceName := "aws_opsworks_stack.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var v opsworks.Stack

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t)
			testAccPreCheckStacks(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_customCookbooksSet(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "default_availability_zone", "data.aws_availability_zones.available", "names.0"),
					resource.TestCheckResourceAttr(resourceName, "default_os", "Amazon Linux 2016.09"),
					resource.TestCheckResourceAttr(resourceName, "default_root_device_type", "ebs"),
					resource.TestCheckResourceAttr(resourceName, "custom_json", customJSON),
					resource.TestCheckResourceAttr(resourceName, "configuration_manager_version", "11.10"),
					resource.TestCheckResourceAttr(resourceName, "use_opsworks_security_groups", "false"),
					resource.TestCheckResourceAttr(resourceName, "use_custom_cookbooks", "true"),
					resource.TestCheckResourceAttr(resourceName, "custom_cookbooks_source.0.type", "git"),
					resource.TestCheckResourceAttr(resourceName, "custom_cookbooks_source.0.revision", "master"),
					resource.TestCheckResourceAttr(resourceName, "custom_cookbooks_source.0.url", "https://github.com/aws/opsworks-example-cookbooks.git"),
					resource.TestCheckResourceAttr(resourceName, "custom_cookbooks_source.0.username", "username"),
					resource.TestCheckResourceAttr(resourceName, "custom_cookbooks_source.0.password", "password"),
					resource.TestCheckResourceAttr(resourceName, "custom_cookbooks_source.0.ssh_key", sshKey),
				),
			},
		},
	})
}

// Tests the addition of regional endpoints and supporting the classic link used
// to create Stack's prior to v0.9.0.
// See https://github.com/hashicorp/terraform/issues/12842
func TestAccOpsWorksStack_classicEndpoints(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_stack.test"
	var v opsworks.Stack

	// This test cannot be parallel with other tests, because it changes the provider region in a non-standard way
	// https://github.com/hashicorp/terraform-provider-aws/issues/21887
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckRegion(t, endpoints.UsWest2RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_classicEndpoint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &v),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Ensure that changing region results in no plan
			{
				Config:   testAccStackConfig_regionalEndpoint(rName),
				PlanOnly: true,
			},
		},
	})
}

func testAccCheckStackRecreated(t *testing.T, before, after *opsworks.Stack) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(before.StackId) == aws.StringValue(after.StackId) {
			t.Fatalf("Expected change of Opsworks StackIds, but both were %v", before.StackId)
		}
		return nil
	}
}

func testAccPreCheckStacks(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).OpsWorksConn

	input := &opsworks.DescribeStacksInput{}

	_, err := conn.DescribeStacks(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

////////////////////////////
//// Checkers and Utilities
////////////////////////////

func testAccCheckCreateStackAttributes(rName string) resource.TestCheckFunc {
	resourceName := "aws_opsworks_stack.test"
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "name", rName),
		resource.TestCheckResourceAttrPair(resourceName, "default_availability_zone", "data.aws_availability_zones.available", "names.0"),
		resource.TestCheckResourceAttr(resourceName, "default_os", "Amazon Linux 2016.09"),
		resource.TestCheckResourceAttr(resourceName, "default_root_device_type", "ebs"),
		resource.TestCheckResourceAttr(resourceName, "custom_json", customJSON),
		resource.TestCheckResourceAttr(resourceName, "configuration_manager_version", "11.10"),
		resource.TestCheckResourceAttr(resourceName, "use_opsworks_security_groups", "false"),
	)
}

func testAccCheckStackExists(n string, v *opsworks.Stack) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OpsWorksConn

		output, err := tfopsworks.FindStackByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckStackDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).OpsWorksConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_opsworks_stack" {
			continue
		}

		_, err := tfopsworks.FindStackByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("OpsWorks Stack %s still exists", rs.Primary.ID)
	}

	return nil
}

//////////////////////////////////////////////////
//// Helper configs for the necessary IAM objects
//////////////////////////////////////////////////

func testAccStackConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "opsworks_service" {
  name = "%[1]s-service"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "opsworks.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
}
EOT
}

resource "aws_iam_role_policy" "opsworks_service" {
  name = "%[1]s-service"
  role = aws_iam_role.opsworks_service.id

  policy = <<EOT
{
  "Statement": [{
    "Action": [
      "ec2:*",
      "iam:PassRole",
      "cloudwatch:GetMetricStatistics",
      "elasticloadbalancing:*",
      "rds:*"
    ],
    "Effect": "Allow",
    "Resource": ["*"]
  }]
}
EOT
}

resource "aws_iam_role" "opsworks_instance" {
  name = "%[1]s-instance"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "ec2.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
}
EOT
}

resource "aws_iam_instance_profile" "opsworks_instance" {
  name = "%[1]s-instance"
  role = aws_iam_role.opsworks_instance.name
}
`, rName)
}

func testAccStackConfig_baseVPC(rName string) string {
	return acctest.ConfigCompose(testAccStackConfig_base(rName), acctest.ConfigVPCWithSubnets(rName, 2))
}

func testAccStackConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccStackConfig_baseVPC(rName), fmt.Sprintf(`
resource "aws_opsworks_stack" "test" {
  name                          = %[1]q
  region                        = %[2]q
  service_role_arn              = aws_iam_role.opsworks_service.arn
  default_instance_profile_arn  = aws_iam_instance_profile.opsworks_instance.arn
  default_subnet_id             = aws_subnet.test[0].id
  vpc_id                        = aws_vpc.test.id
  use_opsworks_security_groups  = false
}
`, rName, acctest.Region()))
}

func testAccStackConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccStackConfig_baseVPC(rName), fmt.Sprintf(`
resource "aws_opsworks_stack" "test" {
  name                          = %[1]q
  region                        = %[2]q
  service_role_arn              = aws_iam_role.opsworks_service.arn
  default_instance_profile_arn  = aws_iam_instance_profile.opsworks_instance.arn
  default_subnet_id             = aws_subnet.test[0].id
  vpc_id                        = aws_vpc.test.id
  use_opsworks_security_groups  = false

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, acctest.Region(), tagKey1, tagValue1))
}

func testAccStackConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccStackConfig_baseVPC(rName), fmt.Sprintf(`
resource "aws_opsworks_stack" "test" {
  name                          = %[1]q
  region                        = %[2]q
  service_role_arn              = aws_iam_role.opsworks_service.arn
  default_instance_profile_arn  = aws_iam_instance_profile.opsworks_instance.arn
  default_subnet_id             = aws_subnet.test[0].id
  vpc_id                        = aws_vpc.test.id
  use_opsworks_security_groups  = false

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, acctest.Region(), tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccStackConfig_baseTagsAlternateRegion(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		testAccStackConfig_base(rName),
		fmt.Sprintf(`
data "aws_availability_zones" "available" {
  provider = "awsalternate"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  provider = "awsalternate"

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  provider = "awsalternate"

  count = 2

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccStackConfig_tags1AlternateRegion(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccStackConfig_baseTagsAlternateRegion(rName), fmt.Sprintf(`
resource "aws_opsworks_stack" "test" {
  name                          = %[1]q
  region                        = %[2]q
  service_role_arn              = aws_iam_role.opsworks_service.arn
  default_instance_profile_arn  = aws_iam_instance_profile.opsworks_instance.arn
  default_subnet_id             = aws_subnet.test[0].id
  vpc_id                        = aws_vpc.test.id
  use_opsworks_security_groups  = false

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, acctest.AlternateRegion(), tagKey1, tagValue1))
}

func testAccStackConfig_tags2AlternateRegion(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccStackConfig_baseTagsAlternateRegion(rName), fmt.Sprintf(`
resource "aws_opsworks_stack" "test" {
  name                          = %[1]q
  region                        = %[2]q
  service_role_arn              = aws_iam_role.opsworks_service.arn
  default_instance_profile_arn  = aws_iam_instance_profile.opsworks_instance.arn
  default_subnet_id             = aws_subnet.test[0].id
  vpc_id                        = aws_vpc.test.id
  use_opsworks_security_groups  = false

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, acctest.AlternateRegion(), tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccStackConfig_classicEndpoint(rName string) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}

resource "aws_opsworks_stack" "test" {
  name                          = %[1]q
  region                        = "us-west-2"
  service_role_arn              = aws_iam_role.opsworks_service.arn
  default_instance_profile_arn  = aws_iam_instance_profile.opsworks_instance.arn
  configuration_manager_version = "12"
  default_availability_zone     = "us-west-2b"
}

resource "aws_iam_role" "opsworks_service" {
  name = %[1]q

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "opsworks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy" "opsworks_service" {
  name = %[1]q
  role = aws_iam_role.opsworks_service.id

  policy = <<EOT
{
  "Statement": [
    {
      "Action": [
        "ec2:*",
        "iam:PassRole",
        "cloudwatch:GetMetricStatistics",
        "elasticloadbalancing:*",
        "rds:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "*"
      ]
    }
  ]
}
EOT
}

resource "aws_iam_role" "opsworks_instance" {
  name = "%[1]s-instance"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_instance_profile" "opsworks_instance" {
  name = %[1]q
  role = aws_iam_role.opsworks_instance.name
}
`, rName) //lintignore:AWSAT003,AT004
}

func testAccStackConfig_regionalEndpoint(rName string) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-west-2"
}

resource "aws_opsworks_stack" "test" {
  name                          = %[1]q
  region                        = "us-west-2"
  service_role_arn              = aws_iam_role.opsworks_service.arn
  default_instance_profile_arn  = aws_iam_instance_profile.opsworks_instance.arn
  configuration_manager_version = "12"
  default_availability_zone     = "us-west-2b"
}

resource "aws_iam_role" "opsworks_service" {
  name = %[1]q

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "opsworks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy" "opsworks_service" {
  name = %[1]q
  role = aws_iam_role.opsworks_service.id

  policy = <<EOT
{
  "Statement": [
    {
      "Action": [
        "ec2:*",
        "iam:PassRole",
        "cloudwatch:GetMetricStatistics",
        "elasticloadbalancing:*",
        "rds:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "*"
      ]
    }
  ]
}
EOT
}

resource "aws_iam_role" "opsworks_instance" {
  name = "%[1]s-instance"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_instance_profile" "opsworks_instance" {
  name = %[1]q
  role = aws_iam_role.opsworks_instance.name
}
`, rName) //lintignore:AWSAT003,AT004
}

func testAccStackConfig_noVPCCreate(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_opsworks_stack" "test" {
  name                         = %[1]q
  region                       = data.aws_region.current.name
  service_role_arn             = aws_iam_role.opsworks_service.arn
  default_instance_profile_arn = aws_iam_instance_profile.opsworks_instance.arn
  default_availability_zone    = data.aws_availability_zones.available.names[0]
  default_os                   = "Amazon Linux 2016.09"
  default_root_device_type     = "ebs"

  custom_json = <<EOF
{
  "key": "value"
}
EOF

  configuration_manager_version = "11.10"
  use_opsworks_security_groups  = false
}

resource "aws_iam_role" "opsworks_service" {
  name = %[1]q

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "opsworks.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy" "opsworks_service" {
  name = %[1]q
  role = aws_iam_role.opsworks_service.id

  policy = <<EOT
{
  "Statement": [
    {
      "Action": [
        "ec2:*",
        "iam:PassRole",
        "cloudwatch:GetMetricStatistics",
        "elasticloadbalancing:*",
        "rds:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "*"
      ]
    }
  ]
}
EOT
}

resource "aws_iam_role" "opsworks_instance" {
  name = "%[1]s-instance"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_instance_profile" "opsworks_instance" {
  name = %[1]q
  role = aws_iam_role.opsworks_instance.name
}
`, rName))
}

func testAccStackConfig_noVPCCreateUpdateServiceRole(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_opsworks_stack" "test" {
  name                         = %[1]q
  region                       = data.aws_region.current.name
  service_role_arn             = aws_iam_role.opsworks_service_new.arn
  default_instance_profile_arn = aws_iam_instance_profile.opsworks_instance.arn
  default_availability_zone    = data.aws_availability_zones.available.names[0]
  default_os                   = "Amazon Linux 2016.09"
  default_root_device_type     = "ebs"

  custom_json = <<EOF
{
  "key": "value"
}
EOF

  configuration_manager_version = "11.10"
  use_opsworks_security_groups  = false
}

resource "aws_iam_role" "opsworks_service" {
  name = %[1]q

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "opsworks.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role" "opsworks_service_new" {
  name = "%[1]s-new"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "opsworks.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy" "opsworks_service_new" {
  name = "%[1]s-new"
  role = aws_iam_role.opsworks_service_new.id

  policy = <<EOT
{
  "Statement": [
    {
      "Action": [
        "ec2:*",
        "iam:PassRole",
        "cloudwatch:*",
        "elasticloadbalancing:*",
        "rds:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "*"
      ]
    }
  ]
}
EOT
}

resource "aws_iam_role_policy" "opsworks_service" {
  name = %[1]q
  role = aws_iam_role.opsworks_service.id

  policy = <<EOT
{
  "Statement": [
    {
      "Action": [
        "ec2:*",
        "iam:PassRole",
        "cloudwatch:GetMetricStatistics",
        "elasticloadbalancing:*",
        "rds:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "*"
      ]
    }
  ]
}
EOT
}

resource "aws_iam_role" "opsworks_instance" {
  name = "%[1]s-instance"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_instance_profile" "opsworks_instance" {
  name = %[1]q
  role = aws_iam_role.opsworks_instance.name
}
`, rName))
}

////////////////////////////
//// Tests for the VPC case
////////////////////////////

func testAccStackConfig_vpcCreate(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.3.5.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = aws_vpc.test.cidr_block
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_opsworks_stack" "test" {
  name                         = %[1]q
  region                       = data.aws_region.current.name
  vpc_id                       = aws_vpc.test.id
  default_subnet_id            = aws_subnet.test.id
  service_role_arn             = aws_iam_role.opsworks_service.arn
  default_instance_profile_arn = aws_iam_instance_profile.opsworks_instance.arn
  default_os                   = "Amazon Linux 2016.09"
  default_root_device_type     = "ebs"

  custom_json = <<EOF
{
  "key": "value"
}
EOF

  configuration_manager_version = "11.10"
  use_opsworks_security_groups  = false
}

resource "aws_iam_role" "opsworks_service" {
  name = %[1]q

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "opsworks.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy" "opsworks_service" {
  name = %[1]q
  role = aws_iam_role.opsworks_service.id

  policy = <<EOT
{
  "Statement": [
    {
      "Action": [
        "ec2:*",
        "iam:PassRole",
        "cloudwatch:GetMetricStatistics",
        "elasticloadbalancing:*",
        "rds:*",
        "ecs:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "*"
      ]
    }
  ]
}
EOT
}

resource "aws_iam_role" "opsworks_instance" {
  name = "%[1]s-instance"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_instance_profile" "opsworks_instance" {
  name = %[1]q
  role = aws_iam_role.opsworks_instance.name
}
`, rName))
}

func testAccStackConfig_vpcUpdate(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.3.5.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = aws_vpc.test.cidr_block
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_opsworks_stack" "test" {
  name                         = %[1]q
  region                       = data.aws_region.current.name
  vpc_id                       = aws_vpc.test.id
  default_subnet_id            = aws_subnet.test.id
  service_role_arn             = aws_iam_role.opsworks_service.arn
  default_instance_profile_arn = aws_iam_instance_profile.opsworks_instance.arn
  default_os                   = "Amazon Linux 2015.09"
  default_root_device_type     = "ebs"

  custom_json = <<EOF
{
  "key": "value"
}
EOF

  configuration_manager_version = "11.10"
  use_opsworks_security_groups  = false
  use_custom_cookbooks          = true
  manage_berkshelf              = true

  custom_cookbooks_source {
    type     = "git"
    revision = "master"
    url      = "https://github.com/aws/opsworks-example-cookbooks.git"
  }
}

resource "aws_iam_role" "opsworks_service" {
  name = %[1]q

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "opsworks.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy" "opsworks_service" {
  name = %[1]q
  role = aws_iam_role.opsworks_service.id

  policy = <<EOT
{
  "Statement": [
    {
      "Action": [
        "ec2:*",
        "iam:PassRole",
        "cloudwatch:GetMetricStatistics",
        "elasticloadbalancing:*",
        "rds:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "*"
      ]
    }
  ]
}
EOT
}

resource "aws_iam_role" "opsworks_instance" {
  name = "%[1]s-instance"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_instance_profile" "opsworks_instance" {
  name = %[1]q
  role = aws_iam_role.opsworks_instance.name
}
`, rName))
}

/////////////////////////////////////////
// Helpers for Custom Cookbook properties
/////////////////////////////////////////

func testAccStackConfig_customCookbooksSet(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.3.5.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = aws_vpc.test.cidr_block
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_opsworks_stack" "test" {
  name                         = %[1]q
  region                       = data.aws_region.current.name
  vpc_id                       = aws_vpc.test.id
  default_subnet_id            = aws_subnet.test.id
  service_role_arn             = aws_iam_role.opsworks_service.arn
  default_instance_profile_arn = aws_iam_instance_profile.opsworks_instance.arn
  default_os                   = "Amazon Linux 2016.09"
  default_root_device_type     = "ebs"

  custom_json = <<EOF
{
  "key": "value"
}
EOF

  configuration_manager_version = "11.10"
  use_opsworks_security_groups  = false
  use_custom_cookbooks          = true
  manage_berkshelf              = true

  custom_cookbooks_source {
    type     = "git"
    revision = "master"
    url      = "https://github.com/aws/opsworks-example-cookbooks.git"
    username = "username"
    password = "password"
    ssh_key  = %[2]q
  }
}

resource "aws_iam_role" "opsworks_service" {
  name = %[1]q

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "opsworks.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy" "opsworks_service" {
  name = %[1]q
  role = aws_iam_role.opsworks_service.id

  policy = <<EOT
{
  "Statement": [
    {
      "Action": [
        "ec2:*",
        "iam:PassRole",
        "cloudwatch:GetMetricStatistics",
        "elasticloadbalancing:*",
        "rds:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "*"
      ]
    }
  ]
}
EOT
}

resource "aws_iam_role" "opsworks_instance" {
  name = "%[1]s-instance"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_instance_profile" "opsworks_instance" {
  name = %[1]q
  role = aws_iam_role.opsworks_instance.name
}
`, rName, sshKey))
}

// One-off, bogus private key generated for use in testing
const sshKey = "-----BEGIN RSA PRIVATE KEY-----" +
	"MIIEpAIBAAKCAQEAv/1hnOZadSDMbJUVJsqweDwc/4TvhTGf0vl9vtNyjzqUUxgU" +
	"RrSvYrgkvWgAFtQ9J5QDNOPSRvS8F1cu7tR036cecdHPmA+Cxto1qENy8UeYrKzV" +
	"I55i+vJiSn3i22HW+SbW1raBM+PL3sp9i0BQmCr8eh3i/VdUm92OQHtnjhfLB3GX" +
	"xnrvytBfI8p2bx9j7mAAjS/X+QncMawPqI9WGuizmuC2cTQHZpZY7j/w+bItoYIV" +
	"g5qJV3908LNlNZGU6etdEUTWM1VSNxG2Yk6eULeStSA4oSkJSHlwP1/fjab0j1b4" +
	"HeB/TUFpy3ODrRAhuHxlyFFWMSzePkXLx9d0GwIDAQABAoIBAFlwrj/M5Ik6XWGc" +
	"Vj07IdjxkETNZlQzmRRNHHKAyRbGoIDRb+i8lhQ0WxFN2PTJrS+5+YBzPevGabWp" +
	"7PhgS45BqaI2rzJUz4TZ9TNNMMgMpaiT37t3Nv9XWckAOmYff2mU2XMvlKNa1QgW" +
	"Z0QvExzAsdwl/jAttgHixjluBAEib+G3p0Xt2CZMQYNzE9H2gH/nqkysiZ5fC+ng" +
	"RnM843jAHtrfz9Q0ATBADMJZgZepnMZyldaOV+s5L8UB893UGhrfGrBwlHd5U5ug" +
	"Z/p74IvOgDd3/pp/2yuyqE+RWz9sakss196aJ0jUXVXjH3F+QDdqqPx0YIJ7S0eM" +
	"13T7hGkCgYEA4TqpoPFYIVEug4gQ6SDttSMwrbA5uBM13s1vjNpDBFuHWFMqgSRe" +
	"xlIAGCGNhoyTr3xr/34filwGMkMdLw8JFISOIbZ18+qgDOsSW0tXwE03vQgKFNB1" +
	"ClGEfcd/4B/oLwOe/bqnKVBQSnfp05yqHjdc9XNQeFxLL8LfSv7LIIUCgYEA2jgt" +
	"108LF+RtdkmSoqLnexJ0jmbPlcYTw1wzuIg49gLdlRxoC+UwPFc+uzMGNxEzr6hG" +
	"Eg3dJVr3+TMLIcTD6usPWzzuL4ReV/IAhCjzgS/WopqURg4cQ+R4MjvTMg8GCZfE" +
	"QvjcbpKh5ndP/QQEOy7cAP8BLVSG3/ichMcttB8CgYAdzmaebvILzrOKIpqiT4JF" +
	"w3dwtO6ehqRNbQCDMmtGC1rY/ICWgJquQjHS/7W8BaSRx7R/JlDEPbNwOWOGU8YO" +
	"2g/5NC1d70HpE77lKA5f25gxwvuaj4+9otYW0y0AGxjeB+ulhmsS05cck8v0/jmh" +
	"MBB0RyNyGjy1AGQOh7OYBQKBgQCwFq1HFM2K1hVOYkglXPcV5OqRDn1sCo5gEsLZ" +
	"oXL1cZKEhIuhLawixPQl8yKMxSDEGjGQ2Acf4axANuRAt5qwskWOBjjdtx66MNoh" +
	"yznTgVrdk4cakMBWOMKVJplhx6XDj+gbct3NjB2A775oGRmg+Esnsp6siYzcpq0G" +
	"qANFWQKBgQCyv8KoQXsD8f8XMvicRC42uZXfhlDjOUzpo1O7WQKWBYqPBqz4AHzE" +
	"Cdy6djI120bqDOifre1qnBjoHezrG+ejaQOTpocOVwT5Zl7BhjoXQZRGiQXj+2aD" +
	"tmm0+hpmkjX7jiPcljjs8S8gh+uCWieJoO4JNPk2SXRiePpYgKzdlg==" +
	"-----END RSA PRIVATE KEY-----"

const customJSON = `{
  "key": "value"
}`
