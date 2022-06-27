package mq_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccMQConfiguration_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(mq.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, mq.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "mq", regexp.MustCompile(`configuration:+.`)),
					resource.TestCheckResourceAttr(resourceName, "authentication_strategy", "simple"),
					resource.TestCheckResourceAttr(resourceName, "description", "TfAccTest MQ Configuration"),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "ActiveMQ"),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "5.15.0"),
					resource.TestCheckResourceAttr(resourceName, "latest_revision", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationConfig_descriptionUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "mq", regexp.MustCompile(`configuration:+.`)),
					resource.TestCheckResourceAttr(resourceName, "description", "TfAccTest MQ Configuration Updated"),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "ActiveMQ"),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "5.15.0"),
					resource.TestCheckResourceAttr(resourceName, "latest_revision", "3"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
		},
	})
}

func TestAccMQConfiguration_withData(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(mq.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, mq.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationConfig_data(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "mq", regexp.MustCompile(`configuration:+.`)),
					resource.TestCheckResourceAttr(resourceName, "description", "TfAccTest MQ Configuration"),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "ActiveMQ"),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "5.15.0"),
					resource.TestCheckResourceAttr(resourceName, "latest_revision", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccMQConfiguration_withLdapData(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(mq.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, mq.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationConfig_ldapData(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "mq", regexp.MustCompile(`configuration:+.`)),
					resource.TestCheckResourceAttr(resourceName, "authentication_strategy", "ldap"),
					resource.TestCheckResourceAttr(resourceName, "description", "TfAccTest MQ Configuration"),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "ActiveMQ"),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "5.15.0"),
					resource.TestCheckResourceAttr(resourceName, "latest_revision", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccMQConfiguration_updateTags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(mq.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, mq.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationConfig_updateTags1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.env", "test"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationConfig_updateTags2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.env", "test2"),
					resource.TestCheckResourceAttr(resourceName, "tags.role", "test-role"),
				),
			},
			{
				Config: testAccConfigurationConfig_updateTags3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.role", "test-role"),
				),
			},
		},
	})
}

func testAccCheckConfigurationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MQConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_mq_configuration" {
			continue
		}

		input := &mq.DescribeConfigurationInput{
			ConfigurationId: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeConfiguration(input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, mq.ErrCodeNotFoundException) {
				return nil
			}
			return err
		}

		// TODO: Delete is not available in the API
		return nil
		//return fmt.Errorf("Expected MQ configuration to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckConfigurationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_mq_configuration" "test" {
  description             = "TfAccTest MQ Configuration"
  name                    = %[1]q
  engine_type             = "ActiveMQ"
  engine_version          = "5.15.0"
  authentication_strategy = "simple"

  data = <<DATA
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
</broker>
DATA
}
`, rName)
}

func testAccConfigurationConfig_descriptionUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_mq_configuration" "test" {
  description    = "TfAccTest MQ Configuration Updated"
  name           = %[1]q
  engine_type    = "ActiveMQ"
  engine_version = "5.15.0"

  data = <<DATA
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
</broker>
DATA
}
`, rName)
}

func testAccConfigurationConfig_data(rName string) string {
	return fmt.Sprintf(`
resource "aws_mq_configuration" "test" {
  description    = "TfAccTest MQ Configuration"
  name           = %[1]q
  engine_type    = "ActiveMQ"
  engine_version = "5.15.0"

  data = <<DATA
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
  <plugins>
    <authorizationPlugin>
      <map>
        <authorizationMap>
          <authorizationEntries>
            <authorizationEntry admin="guests,users" queue="GUEST.&gt;" read="guests" write="guests,users"/>
            <authorizationEntry admin="guests,users" read="guests,users" topic="ActiveMQ.Advisory.&gt;" write="guests,users"/>
          </authorizationEntries>
          <tempDestinationAuthorizationEntry>
            <tempDestinationAuthorizationEntry admin="tempDestinationAdmins" read="tempDestinationAdmins" write="tempDestinationAdmins"/>
          </tempDestinationAuthorizationEntry>
        </authorizationMap>
      </map>
    </authorizationPlugin>
    <forcePersistencyModeBrokerPlugin persistenceFlag="true"/>
    <statisticsBrokerPlugin/>
    <timeStampingBrokerPlugin ttlCeiling="86400000" zeroExpirationOverride="86400000"/>
  </plugins>
</broker>
DATA
}
`, rName)
}

func testAccConfigurationConfig_ldapData(rName string) string {
	return fmt.Sprintf(`
resource "aws_mq_configuration" "test" {
  description             = "TfAccTest MQ Configuration"
  name                    = %[1]q
  engine_type             = "ActiveMQ"
  engine_version          = "5.15.0"
  authentication_strategy = "ldap"

  data = <<DATA
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
  <plugins>
    <authorizationPlugin>
      <map>
        <cachedLDAPAuthorizationMap legacyGroupMapping="false" queueSearchBase="ou=Queue,ou=Destination,ou=ActiveMQ,dc=example,dc=org" refreshInterval="0" tempSearchBase="ou=Temp,ou=Destination,ou=ActiveMQ,dc=example,dc=org" topicSearchBase="ou=Topic,ou=Destination,ou=ActiveMQ,dc=example,dc=org"/>
      </map>
    </authorizationPlugin>
    <forcePersistencyModeBrokerPlugin persistenceFlag="true"/>
    <statisticsBrokerPlugin/>
    <timeStampingBrokerPlugin ttlCeiling="86400000" zeroExpirationOverride="86400000"/>
  </plugins>
</broker>
DATA
}
`, rName)
}

func testAccConfigurationConfig_updateTags1(rName string) string {
	return fmt.Sprintf(`
resource "aws_mq_configuration" "test" {
  description    = "TfAccTest MQ Configuration"
  name           = %[1]q
  engine_type    = "ActiveMQ"
  engine_version = "5.15.0"

  data = <<DATA
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
</broker>
DATA

  tags = {
    env = "test"
  }
}
`, rName)
}

func testAccConfigurationConfig_updateTags2(rName string) string {
	return fmt.Sprintf(`
resource "aws_mq_configuration" "test" {
  description    = "TfAccTest MQ Configuration"
  name           = %[1]q
  engine_type    = "ActiveMQ"
  engine_version = "5.15.0"

  data = <<DATA
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
</broker>
DATA

  tags = {
    env  = "test2"
    role = "test-role"
  }
}
`, rName)
}

func testAccConfigurationConfig_updateTags3(rName string) string {
	return fmt.Sprintf(`
resource "aws_mq_configuration" "test" {
  description    = "TfAccTest MQ Configuration"
  name           = %[1]q
  engine_type    = "ActiveMQ"
  engine_version = "5.15.0"

  data = <<DATA
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
</broker>
DATA

  tags = {
    role = "test-role"
  }
}
`, rName)
}
