package appflow_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appflow"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappflow "github.com/hashicorp/terraform-provider-aws/internal/service/appflow"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccAppFlowConnectorProfile_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var connectorProfiles appflow.DescribeConnectorProfilesOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appflow_connector_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, appflow.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConnectorProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigConnectorProfile_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectorProfileExists(resourceName, &connectorProfiles),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "appflow", regexp.MustCompile(`connectorprofile/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "connection_mode"),
					resource.TestCheckResourceAttrSet(resourceName, "connector_profile_config.#"),
					resource.TestCheckResourceAttrSet(resourceName, "connector_profile_config.0.connector_profile_credentials.#"),
					resource.TestCheckResourceAttrSet(resourceName, "connector_profile_config.0.connector_profile_properties.#"),
					resource.TestCheckResourceAttrSet(resourceName, "connector_type"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"connector_profile_config.0.connector_profile_credentials"},
			},
		},
	})
}

func TestAccAppFlowConnectorProfile_update(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var connectorProfiles appflow.DescribeConnectorProfilesOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appflow_connector_profile.test"
	testPrefix := "test-prefix"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, appflow.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConnectorProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigConnectorProfile_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectorProfileExists(resourceName, &connectorProfiles),
				),
			},
			{
				Config: testAccConfigConnectorProfile_update(rName, testPrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectorProfileExists(resourceName, &connectorProfiles),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.redshift.0.bucket_prefix", testPrefix),
				),
			},
		},
	})
}

func TestAccAppFlowConnectorProfile_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var connectorProfiles appflow.DescribeConnectorProfilesOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appflow_connector_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, appflow.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConnectorProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigConnectorProfile_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectorProfileExists(resourceName, &connectorProfiles),
					acctest.CheckResourceDisappears(acctest.Provider, tfappflow.ResourceConnectorProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckConnectorProfileDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppFlowConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appflow_connector_profile" {
			continue
		}

		_, err := tfappflow.FindConnectorProfileByArn(context.Background(), conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Expected AppFlow Connector Profile to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckConnectorProfileExists(n string, res *appflow.DescribeConnectorProfilesOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFlowConn

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		req := &appflow.DescribeConnectorProfilesInput{
			ConnectorProfileNames: []*string{aws.String(rs.Primary.Attributes["name"])},
		}
		describe, err := conn.DescribeConnectorProfiles(req)

		if len(describe.ConnectorProfileDetails) == 0 {
			return fmt.Errorf("AppFlow Connector profile %s does not exist.", n)
		}

		if err != nil {
			return err
		}

		*res = *describe

		return nil
	}
}

func testAccConnectorProfileConfigBase(connectorProfileName string, redshiftPassword string, redshiftUsername string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_route" "test" {
  route_table_id = data.aws_route_table.test.id

  destination_cidr_block = "0.0.0.0/0"

  gateway_id = aws_internet_gateway.test.id
}

resource "aws_subnet" "test" {
  cidr_block        = aws_vpc.test.cidr_block
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_redshift_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = [aws_subnet.test.id]
}

data "aws_iam_policy" "test" {
  name = "AmazonRedshiftAllCommandsFullAccess"
}

resource "aws_iam_role" "test" {
  name = %[1]q

  managed_policy_arns = [data.aws_iam_policy.test.arn]

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_security_group" "test" {
  name = %[1]q

  vpc_id = aws_vpc.test.id
}

resource "aws_security_group_rule" "test" {
  type = "ingress"

  security_group_id = aws_security_group.test.id

  from_port   = 0
  to_port     = 65535
  protocol    = "-1"
  cidr_blocks = ["0.0.0.0/0"]
}

resource "aws_redshift_cluster" "test" {
  cluster_identifier = %[1]q

  availability_zone         = data.aws_availability_zones.available.names[0]
  cluster_subnet_group_name = aws_redshift_subnet_group.test.name
  vpc_security_group_ids    = [aws_security_group.test.id]

  master_password = %[2]q
  master_username = %[3]q

  publicly_accessible = true

  node_type           = "dc2.large"
  skip_final_snapshot = true
}
`, connectorProfileName, redshiftPassword, redshiftUsername))
}

func testAccConfigConnectorProfile_basic(connectorProfileName string) string {
	const redshiftPassword = "testPassword123!"
	const redshiftUsername = "testusername"

	return acctest.ConfigCompose(
		testAccConnectorProfileConfigBase(connectorProfileName, redshiftPassword, redshiftUsername),
		fmt.Sprintf(`
resource "aws_appflow_connector_profile" "test" {
  name            = %[1]q
  connector_type  = "Redshift"
  connection_mode = "Public"

  connector_profile_config {

    connector_profile_credentials {
      redshift {
        password = aws_redshift_cluster.test.master_password
        username = aws_redshift_cluster.test.master_username
      }
    }

    connector_profile_properties {
      redshift {
        bucket_name  = %[1]q
        database_url = "jdbc:redshift://${aws_redshift_cluster.test.endpoint}/dev"
        role_arn     = aws_iam_role.test.arn
      }
    }
  }

  depends_on = [
    aws_route.test,
    aws_security_group_rule.test,
  ]
}
`, connectorProfileName, redshiftPassword, redshiftUsername),
	)
}

func testAccConfigConnectorProfile_update(connectorProfileName string, bucketPrefix string) string {
	const redshiftPassword = "testPassword123!"
	const redshiftUsername = "testusername"

	return acctest.ConfigCompose(
		testAccConnectorProfileConfigBase(connectorProfileName, redshiftPassword, redshiftUsername),
		fmt.Sprintf(`
resource "aws_appflow_connector_profile" "test" {
  name            = %[1]q
  connector_type  = "Redshift"
  connection_mode = "Public"

  connector_profile_config {

    connector_profile_credentials {
      redshift {
        password = aws_redshift_cluster.test.master_password
        username = aws_redshift_cluster.test.master_username
      }
    }

    connector_profile_properties {
      redshift {
        bucket_name   = %[1]q
        bucket_prefix = %[4]q
        database_url  = "jdbc:redshift://${aws_redshift_cluster.test.endpoint}/dev"
        role_arn      = aws_iam_role.test.arn
      }
    }
  }

  depends_on = [
    aws_route.test,
    aws_security_group_rule.test,
  ]
}
`, connectorProfileName, redshiftPassword, redshiftUsername, bucketPrefix),
	)
}
