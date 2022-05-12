package appflow_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appflow"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappflow "github.com/hashicorp/terraform-provider-aws/internal/service/appflow"
)

func TestAccAppFlowConnectorProfile_basic(t *testing.T) {
	var connectorProfiles appflow.DescribeConnectorProfilesOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appflow_connector_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appflow.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAppFlowConnectorProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppFlowConnectorProfile_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppFlowConnectorProfileExists(resourceName, &connectorProfiles),
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

func TestAccAWSAppFlowConnectorProfile_Amplitude(t *testing.T) {
	var connectorProfiles appflow.DescribeConnectorProfilesOutput

	connectorProfileName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appflow_connector_profile.amplitude"

	apiKey := os.Getenv("AMPLITUDE_API_KEY")
	secretKey := os.Getenv("AMPLITUDE_SECRET_KEY")

	if apiKey == "" || secretKey == "" {
		t.Skip("All environment variables: AMPLITUDE_API_KEY, AMPLITUDE_SECRET_KEY must be set.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appflow.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAppFlowConnectorProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppFlowConnectorProfile_Amplitude(connectorProfileName, apiKey, secretKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppFlowConnectorProfileExists(resourceName, &connectorProfiles),
					resource.TestCheckResourceAttr(resourceName, "connection_mode", "Public"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "connector_profile_arn", "appflow", fmt.Sprintf("connectorprofile/%s", connectorProfileName)),
					resource.TestCheckResourceAttr(resourceName, "connector_type", "Amplitude"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.amplitude.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.amplitude.0.api_key", apiKey),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.amplitude.0.secret_key", secretKey),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.amplitude.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", connectorProfileName),
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

func TestAccAWSAppFlowConnectorProfile_Datadog(t *testing.T) {
	var connectorProfiles appflow.DescribeConnectorProfilesOutput

	connectorProfileName := sdkacctest.RandomWithPrefix("tf-acc-test-appflow-connector-profile")
	resourceName := "aws_appflow_connector_profile.datadog"

	apiKey := os.Getenv("DATADOG_API_KEY")
	applicationKey := os.Getenv("DATADOG_APPLICATION_KEY")
	instanceUrl := os.Getenv("DATADOG_INSTANCE_URL")

	if apiKey == "" || applicationKey == "" || instanceUrl == "" {
		t.Skip("All environment variables: DATADOG_API_KEY, DATADOG_APPLICATION_KEY, DATADOG_INSTANCE_URL must be set.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appflow.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAppFlowConnectorProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppFlowConnectorProfile_Datadog(connectorProfileName, apiKey, applicationKey, instanceUrl),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppFlowConnectorProfileExists(resourceName, &connectorProfiles),
					resource.TestCheckResourceAttr(resourceName, "connection_mode", "Public"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "connector_profile_arn", "appflow", fmt.Sprintf("connectorprofile/%s", connectorProfileName)),
					resource.TestCheckResourceAttr(resourceName, "connector_type", "Datadog"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.datadog.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.datadog.0.api_key", apiKey),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.datadog.0.application_key", applicationKey),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.datadog.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.datadog.0.instance_url", instanceUrl),
					resource.TestCheckResourceAttr(resourceName, "name", connectorProfileName),
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

func TestAccAWSAppFlowConnectorProfile_Dynatrace(t *testing.T) {
	var connectorProfiles appflow.DescribeConnectorProfilesOutput

	connectorProfileName := sdkacctest.RandomWithPrefix("tf-acc-test-appflow-connector-profile")
	resourceName := "aws_appflow_connector_profile.dynatrace"

	apiToken := os.Getenv("DYNATRACE_API_TOKEN")
	instanceUrl := os.Getenv("DYNATRACE_INSTANCE_URL")

	if apiToken == "" || instanceUrl == "" {
		t.Skip("All environment variables: DYNATRACE_API_TOKEN, DYNATRACE_INSTANCE_URL must be set.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appflow.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAppFlowConnectorProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppFlowConnectorProfile_Dynatrace(connectorProfileName, apiToken, instanceUrl),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppFlowConnectorProfileExists(resourceName, &connectorProfiles),
					resource.TestCheckResourceAttr(resourceName, "connection_mode", "Public"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "connector_profile_arn", "appflow", fmt.Sprintf("connectorprofile/%s", connectorProfileName)),
					resource.TestCheckResourceAttr(resourceName, "connector_type", "Dynatrace"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.dynatrace.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.dynatrace.0.api_token", apiToken),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.dynatrace.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.dynatrace.0.instance_url", instanceUrl),
					resource.TestCheckResourceAttr(resourceName, "name", connectorProfileName),
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

func TestAccAWSAppFlowConnectorProfile_Slack(t *testing.T) {
	var connectorProfiles appflow.DescribeConnectorProfilesOutput

	connectorProfileName := sdkacctest.RandomWithPrefix("tf-acc-test-appflow-connector-profile")
	resourceName := "aws_appflow_connector_profile.slack"

	clientId := os.Getenv("SLACK_CLIENT_ID")
	clientSecret := os.Getenv("SLACK_CLIENT_SECRET")
	accessToken := os.Getenv("SLACK_ACCESS_TOKEN")
	instanceUrl := os.Getenv("SLACK_INSTANCE_URL")

	if clientId == "" || clientSecret == "" || accessToken == "" || instanceUrl == "" {
		t.Skip("All environment variables: SLACK_CLIENT_ID, SLACK_CLIENT_SECRET, SLACK_ACCESS_TOKEN, SLACK_INSTANCE_URL must be set.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appflow.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAppFlowConnectorProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppFlowConnectorProfile_Slack(connectorProfileName, clientId, clientSecret, accessToken, instanceUrl),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppFlowConnectorProfileExists(resourceName, &connectorProfiles),
					resource.TestCheckResourceAttr(resourceName, "connection_mode", "Public"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "connector_profile_arn", "appflow", fmt.Sprintf("connectorprofile/%s", connectorProfileName)),
					resource.TestCheckResourceAttr(resourceName, "connector_type", "Slack"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.slack.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.slack.0.access_token", accessToken),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.slack.0.client_id", clientId),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.slack.0.client_secret", clientSecret),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.slack.0.oauth_request.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "connector_profile_config.0.connector_profile_credentials.0.slack.0.oauth_request.0.redirect_uri"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.slack.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.slack.0.instance_url", instanceUrl),
					resource.TestCheckResourceAttr(resourceName, "name", connectorProfileName),
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

func TestAccAWSAppFlowConnectorProfile_Snowflake(t *testing.T) {
	var connectorProfiles appflow.DescribeConnectorProfilesOutput

	connectorProfileName := sdkacctest.RandomWithPrefix("tf-acc-test-appflow-connector-profile")
	resourceName := "aws_appflow_connector_profile.snowflake"

	password := os.Getenv("SNOWFLAKE_PASSWORD")
	username := os.Getenv("SNOWFLAKE_USERNAME")
	accountName := os.Getenv("SNOWFLAKE_ACCOUNT_NAME")
	bucketName := connectorProfileName
	region := os.Getenv("SNOWFLAKE_REGION")
	stage := os.Getenv("SNOWFLAKE_STAGE")
	warehouse := os.Getenv("SNOWFLAKE_WAREHOUSE")

	if password == "" || username == "" || accountName == "" || region == "" || stage == "" || warehouse == "" {
		t.Skip("All environment variables: SNOWFLAKE_PASSWORD, SNOWFLAKE_USERNAME, SNOWFLAKE_ACCOUNT_NAME, SNOWFLAKE_REGION, SNOWFLAKE_STAGE, SNOWFLAKE_WAREHOUSE must be set.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appflow.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAppFlowConnectorProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppFlowConnectorProfile_Snowflake(connectorProfileName, password, username, accountName, bucketName, region, stage, warehouse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppFlowConnectorProfileExists(resourceName, &connectorProfiles),
					resource.TestCheckResourceAttr(resourceName, "connection_mode", "Public"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "connector_profile_arn", "appflow", fmt.Sprintf("connectorprofile/%s", connectorProfileName)),
					resource.TestCheckResourceAttr(resourceName, "connector_type", "Snowflake"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.snowflake.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.snowflake.0.password", password),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.snowflake.0.username", username),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.snowflake.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.snowflake.0.account_name", accountName),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.snowflake.0.bucket_name", bucketName),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.snowflake.0.region", region),
					resource.TestCheckResourceAttrSet(resourceName, "connector_profile_config.0.connector_profile_properties.0.snowflake.0.stage"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.snowflake.0.warehouse", warehouse),
					resource.TestCheckResourceAttr(resourceName, "name", connectorProfileName),
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

func TestAccAWSAppFlowConnectorProfile_Trendmicro(t *testing.T) {
	var connectorProfiles appflow.DescribeConnectorProfilesOutput

	connectorProfileName := sdkacctest.RandomWithPrefix("tf-acc-test-appflow-connector-profile")
	resourceName := "aws_appflow_connector_profile.trendmicro"

	apiSecretKey := os.Getenv("TRENDMICRO_API_SECRET_KEY")

	if apiSecretKey == "" {
		t.Skip("All environment variables: TRENDMICRO_API_SECRET_KEY must be set.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appflow.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAppFlowConnectorProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppFlowConnectorProfile_Trendmicro(connectorProfileName, apiSecretKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppFlowConnectorProfileExists(resourceName, &connectorProfiles),
					resource.TestCheckResourceAttr(resourceName, "connection_mode", "Public"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "connector_profile_arn", "appflow", fmt.Sprintf("connectorprofile/%s", connectorProfileName)),
					resource.TestCheckResourceAttr(resourceName, "connector_type", "Trendmicro"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.trendmicro.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.trendmicro.0.api_secret_key", apiSecretKey),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", connectorProfileName),
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

func TestAccAWSAppFlowConnectorProfile_Zendesk(t *testing.T) {
	var connectorProfiles appflow.DescribeConnectorProfilesOutput

	connectorProfileName := sdkacctest.RandomWithPrefix("tf-acc-test-appflow-connector-profile")
	resourceName := "aws_appflow_connector_profile.zendesk"

	clientId := os.Getenv("ZENDESK_CLIENT_ID")
	clientSecret := os.Getenv("ZENDESK_CLIENT_SECRET")
	accessToken := os.Getenv("ZENDESK_ACCESS_TOKEN")
	instanceUrl := os.Getenv("ZENDESK_INSTANCE_URL")

	if clientId == "" || clientSecret == "" || accessToken == "" || instanceUrl == "" {
		t.Skip("All environment variables: ZENDESK_CLIENT_ID, ZENDESK_CLIENT_SECRET, ZENDESK_ACCESS_TOKEN, ZENDESK_INSTANCE_URL must be set.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appflow.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAppFlowConnectorProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppFlowConnectorProfile_Zendesk(connectorProfileName, clientId, clientSecret, accessToken, instanceUrl),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppFlowConnectorProfileExists(resourceName, &connectorProfiles),
					resource.TestCheckResourceAttr(resourceName, "connection_mode", "Public"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "connector_profile_arn", "appflow", fmt.Sprintf("connectorprofile/%s", connectorProfileName)),
					resource.TestCheckResourceAttr(resourceName, "connector_type", "Zendesk"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.zendesk.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.zendesk.0.access_token", accessToken),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.zendesk.0.client_id", clientId),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.zendesk.0.client_secret", clientSecret),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.zendesk.0.oauth_request.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "connector_profile_config.0.connector_profile_credentials.0.zendesk.0.oauth_request.0.redirect_uri"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.zendesk.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.zendesk.0.instance_url", instanceUrl),
					resource.TestCheckResourceAttr(resourceName, "name", connectorProfileName),
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

func testAccCheckAppFlowConnectorProfileDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppFlowConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appflow_connector_profile" {
			continue
		}

		_, err := tfappflow.FindConnectorProfileByName(context.Background(), conn, rs.Primary.ID)

		if _, ok := err.(*resource.NotFoundError); ok {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Expected AppFlow Connector Profile to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSAppFlowConnectorProfileExists(n string, res *appflow.DescribeConnectorProfilesOutput) resource.TestCheckFunc {
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

func testAccAppFlowConnectorProfileConfigBase(connectorProfileName string, redshiftPassword string, redshiftUsername string) string {
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
  subnet_ids = [ aws_subnet.test.id ]
}

resource "aws_iam_role" "test" {
  name = %[1]q

  managed_policy_arns = [ "arn:aws:iam::aws:policy/AmazonRedshiftAllCommandsFullAccess" ]

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
  name   = %[1]q

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
  cluster_identifier  = %[1]q

  availability_zone         = data.aws_availability_zones.available.names[0]
  cluster_subnet_group_name = aws_redshift_subnet_group.test.name
  vpc_security_group_ids    = [ aws_security_group.test.id ]

  master_password = %[2]q
  master_username = %[3]q

  publicly_accessible = true

  node_type           = "dc2.large"
  skip_final_snapshot = true
}
`, connectorProfileName, redshiftPassword, redshiftUsername))
}

func testAccAppFlowConnectorProfile_basic(connectorProfileName string) string {
	const redshiftPassword = "testPassword123!"
	const redshiftUsername = "testusername"

	return acctest.ConfigCompose(
		testAccAppFlowConnectorProfileConfigBase(connectorProfileName, redshiftPassword, redshiftUsername),
		fmt.Sprintf(`
resource "aws_appflow_connector_profile" "test" {
  name = %[1]q
  connector_type         = "Redshift"
  connection_mode        = "Public"

  connector_profile_config {

    connector_profile_credentials {
      redshift {
		  password = %[2]q
		  username = %[3]q
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

func testAccAWSAppFlowConnectorProfile_Amplitude(connectorProfileName string, apiKey string, secretKey string) string {
	return fmt.Sprintf(`
resource "aws_appflow_connector_profile" "amplitude" {
  name = %[1]q
  connector_type         = "Amplitude"
  connection_mode        = "Public"
  connector_profile_config {
    connector_profile_credentials {
      amplitude {
        api_key    = %[2]q
        secret_key = %[3]q
      }
    }
    connector_profile_properties {
      amplitude {
      }
    }
  }

}
`, connectorProfileName, apiKey, secretKey)
}

func testAccAWSAppFlowConnectorProfile_Datadog(connectorProfileName string, apiKey string, applicationKey string, instanceUrl string) string {
	return fmt.Sprintf(`
resource "aws_appflow_connector_profile" "datadog" {
  name = %[1]q
  connector_type         = "Datadog"
  connection_mode        = "Public"
  connector_profile_config {
    connector_profile_credentials {
      datadog {
        api_key         = %[2]q
        application_key = %[3]q
      }
    }
    connector_profile_properties {
      datadog {
        instance_url = %[4]q
      }
    }
  }

}
`, connectorProfileName, apiKey, applicationKey, instanceUrl)
}

func testAccAWSAppFlowConnectorProfile_Dynatrace(connectorProfileName string, apiToken string, instanceUrl string) string {
	return fmt.Sprintf(`
resource "aws_appflow_connector_profile" "dynatrace" {
  name = %[1]q
  connector_type         = "Dynatrace"
  connection_mode        = "Public"
  connector_profile_config {
    connector_profile_credentials {
      dynatrace {
        api_token = %[2]q
      }
    }
    connector_profile_properties {
      dynatrace {
        instance_url = %[3]q
      }
    }
  }

}
`, connectorProfileName, apiToken, instanceUrl)
}

func testAccAWSAppFlowConnectorProfile_Slack(connectorProfileName string, clientId string, clientSecret string, accessToken string, instanceUrl string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_appflow_connector_profile" "slack" {
  connection_mode        = "Public"
  name = %[1]q
  connector_type         = "Slack"
  connector_profile_config {
    connector_profile_credentials {
      slack {
        client_id     = %[2]q
        client_secret = %[3]q
        access_token  = %[4]q
        oauth_request {
          redirect_uri = "https://${data.aws_region.current.name}.console.aws.amazon.com/appflow/oauth"
        }
      }
    }
    connector_profile_properties {
      slack {
        instance_url = %[5]q
      }
    }
  }
}
`, connectorProfileName, clientId, clientSecret, accessToken, instanceUrl)
}

func testAccAWSAppFlowConnectorProfile_Snowflake(connectorProfileName string, password string, username string, accountName string, bucketName string, region string, stage string, warehouse string) string {

	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_s3_bucket" "snowflake" {
  bucket = %[1]q
  acl    = "private"
}

resource "aws_appflow_connector_profile" "snowflake" {
  name = %[2]q
  connector_type         = "Snowflake"
  connection_mode        = "Public"
  connector_profile_config {
    connector_profile_credentials {
      snowflake {
        password = %[3]q
        username = %[4]q
      }
    }
    connector_profile_properties {
      snowflake {
        account_name = %[5]q
        bucket_name  = aws_s3_bucket.snowflake.id
        region       = %[6]q
        stage        = %[7]q
        warehouse    = %[8]q
      }
    }
  }

}
`, bucketName, connectorProfileName, password, username, accountName, region, stage, warehouse)
}

func testAccAWSAppFlowConnectorProfile_Trendmicro(connectorProfileName string, apiSecretKey string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_appflow_connector_profile" "trendmicro" {
  name = %[1]q
  connector_type         = "Trendmicro"
  connection_mode        = "Public"
  connector_profile_config {
    connector_profile_credentials {
      trendmicro {
        api_secret_key = %[2]q
      }
    }
    connector_profile_properties {
      trendmicro {
      }
    }
  }

}
`, connectorProfileName, apiSecretKey)
}

func testAccAWSAppFlowConnectorProfile_Zendesk(connectorProfileName string, clientId string, clientSecret string, accessToken string, instanceUrl string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_appflow_connector_profile" "zendesk" {
  connection_mode        = "Public"
  name = %[1]q
  connector_type         = "Zendesk"
  connector_profile_config {
    connector_profile_credentials {
      zendesk {
        client_id     = %[2]q
        client_secret = %[3]q
        access_token  = %[4]q
        oauth_request {
          redirect_uri = "https://${data.aws_region.current.name}.console.aws.amazon.com/appflow/oauth"
        }
      }
    }
    connector_profile_properties {
      zendesk {
        instance_url = %[5]q
      }
    }
  }
}
`, connectorProfileName, clientId, clientSecret, accessToken, instanceUrl)
}
