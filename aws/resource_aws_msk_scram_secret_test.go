package aws

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var newSecrets = aws.StringSlice([]string{"AmazonMSK_test_aws_example_test_tf-test-2-jbuaA1",
	"AmazonMSK_test_aws_example_test_tf-test-5-I50lRm"})

var existingSecrets = aws.StringSlice([]string{"AmazonMSK_test_aws_example_test_tf-test-2-jbuaA1",
	"AmazonMSK_test_aws_example_test_tf-test-4-I50lRm"})

func TestAccAwsMskScramSecret_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_sasl_scram_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMsk(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskSaslScramSecret_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "scram_secrets.#"),
				),
			},
		},
	})
}
func TestAccAwsMskScramSecret_Delete(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_sasl_scram_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMsk(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskSaslScramSecret_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "scram_secrets.#"),
				),
			},
			{
				Config: testAccMskSaslScramSecretDelete(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSaslScramSecretsDestruction(resourceName),
				),
			},
		},
	})
}

func TestAccAwsMskScramSecret_UpdateRemove(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_sasl_scram_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMsk(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskSaslScramSecret_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "scram_secrets.#"),
				),
			},
			{
				Config: testAccMskSaslScramSecretUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSaslScramSecretsDontExist(resourceName, 2),
				),
			},
		},
	})
}

func testAccCheckAwsSaslScramSecretsDontExist(resourceName string, count int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if len(rs.Primary.Attributes["secret_arn_list"]) > count {
			return fmt.Errorf("Too many secrets for %v", rs.Primary.Attributes["secret_arn_list"])
		}

		return nil
	}
}

func testAccCheckAwsSaslScramSecretsDestruction(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[resourceName]
		if ok {
			return fmt.Errorf("Should not have found: %s", resourceName)
		}

		return nil
	}
}

func TestAwsMskFilterNewSecrets(t *testing.T) {
	expected := []string{"AmazonMSK_test_aws_example_test_tf-test-5-I50lRm"}

	result := filterNewSecrets(newSecrets, existingSecrets)
	if !reflect.DeepEqual(expected, aws.StringValueSlice(result)) {
		t.Fatalf("Expected secret list to be %v, got %v", expected, aws.StringValueSlice(result))
	}
}

func TestAwsMskFilterExistingSecrets(t *testing.T) {
	expected := []string{"AmazonMSK_test_aws_example_test_tf-test-2-jbuaA1"}

	result := filterExistingSecrets(newSecrets, existingSecrets)
	if !reflect.DeepEqual(expected, aws.StringValueSlice(result)) {
		t.Fatalf("Expected secret list to be %v, got %v", expected, aws.StringValueSlice(result))
	}
}

func TestAwsMskFilterDeletionSecrets(t *testing.T) {
	expectedUpdate := []string{"AmazonMSK_test_aws_example_test_tf-test-5-I50lRm"}
	expectedDelete := []string{"AmazonMSK_test_aws_example_test_tf-test-4-I50lRm"}

	updated, deleted := filterSecretsForDeletion(newSecrets, existingSecrets)

	if !reflect.DeepEqual(expectedUpdate, aws.StringValueSlice(updated)) {
		t.Fatalf("Expected secret list to be %v, got %v", expectedUpdate, aws.StringValueSlice(updated))
	}

	if !reflect.DeepEqual(expectedDelete, aws.StringValueSlice(deleted)) {
		t.Fatalf("Expected secret list to be %v, got %v", expectedDelete, aws.StringValueSlice(deleted))
	}
}

func testAccMskSaslScramSecret_basic(rName string) string {
	return testAccMskClusterBaseConfig() + fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.5.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.t3.small"
    security_groups = [aws_security_group.example_sg.id]
  }

  client_authentication {
    sasl {
      scram = true
    }
  }
}

resource "aws_kms_key" "test" {
  description = "%s-kms-key-msk"
}

resource "aws_secretsmanager_secret" "test_basic" {
  name       = "AmazonMSK_test_%s_1_basic"
  kms_key_id = aws_kms_key.test.key_id
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test_basic.id
  secret_string = jsonencode({ username = "%s", password = "foobar" })
}

resource "aws_msk_sasl_scram_secret" "test" {
  cluster_arn     = aws_msk_cluster.test.arn
  secret_arn_list = [aws_secretsmanager_secret.test_basic.arn]
}

resource "aws_secretsmanager_secret_policy" "test" {
  secret_arn = aws_secretsmanager_secret.test_basic.arn
  policy     = <<POLICY
{
  "Version" : "2012-10-17",
  "Statement" : [ {
    "Sid": "AWSKafkaResourcePolicy",
    "Effect" : "Allow",
    "Principal" : {
      "Service" : "kafka.amazonaws.com"
    },
    "Action" : "secretsmanager:getSecretValue",
    "Resource" : "${aws_secretsmanager_secret.test_basic.arn}"
  } ]
}
POLICY
}
`, rName, rName, rName, rName)
}

func testAccMskSaslScramSecretUpdate(rName string) string {
	return testAccMskClusterBaseConfig() + fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.5.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.t3.small"
    security_groups = [aws_security_group.example_sg.id]
  }

  client_authentication {
    sasl {
      scram = true
    }
  }
}

resource "aws_kms_key" "test" {
  description = "%s-kms-key-msk"
}

resource "aws_secretsmanager_secret" "test_update" {
  name       = "AmazonMSK_test_%s_1_update"
  kms_key_id = aws_kms_key.test.key_id
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test_update.id
  secret_string = jsonencode({ username = "%s", password = "foobar" })
}

resource "aws_secretsmanager_secret" "test2" {
  name       = "AmazonMSK_test_%s_2_update"
  kms_key_id = aws_kms_key.test.key_id
}

resource "aws_secretsmanager_secret_version" "test2" {
  secret_id     = aws_secretsmanager_secret.test2.id
  secret_string = jsonencode({ username = "%s", password = "foobar" })
}

resource "aws_secretsmanager_secret" "test3" {
  name       = "AmazonMSK_test_%s_3_update"
  kms_key_id = aws_kms_key.test.key_id
}

resource "aws_secretsmanager_secret_version" "test3" {
  secret_id     = aws_secretsmanager_secret.test3.id
  secret_string = jsonencode({ username = "%s", password = "foobar" })
}

resource "aws_msk_sasl_scram_secret" "test" {
  cluster_arn     = aws_msk_cluster.test.arn
  secret_arn_list = [aws_secretsmanager_secret.test2.arn, aws_secretsmanager_secret.test3.arn]
}

resource "aws_secretsmanager_secret_policy" "test" {
  secret_arn = aws_secretsmanager_secret.test_update.arn
  policy     = <<POLICY
{
  "Version" : "2012-10-17",
  "Statement" : [ {
    "Sid": "AWSKafkaResourcePolicy",
    "Effect" : "Allow",
    "Principal" : {
      "Service" : "kafka.amazonaws.com"
    },
    "Action" : "secretsmanager:getSecretValue",
    "Resource" : "${aws_secretsmanager_secret.test_update.arn}"
  } ]
}
POLICY
}

resource "aws_secretsmanager_secret_policy" "test2" {
  secret_arn = aws_secretsmanager_secret.test2.arn
  policy     = <<POLICY
{
  "Version" : "2012-10-17",
  "Statement" : [ {
    "Sid": "AWSKafkaResourcePolicy",
    "Effect" : "Allow",
    "Principal" : {
      "Service" : "kafka.amazonaws.com"
    },
    "Action" : "secretsmanager:getSecretValue",
    "Resource" : "${aws_secretsmanager_secret.test2.arn}"
  } ]
}
POLICY
}

resource "aws_secretsmanager_secret_policy" "test3" {
  secret_arn = aws_secretsmanager_secret.test3.arn
  policy     = <<POLICY
{
  "Version" : "2012-10-17",
  "Statement" : [ {
    "Sid": "AWSKafkaResourcePolicy",
    "Effect" : "Allow",
    "Principal" : {
      "Service" : "kafka.amazonaws.com"
    },
    "Action" : "secretsmanager:getSecretValue",
    "Resource" : "${aws_secretsmanager_secret.test3.arn}"
  } ]
}
POLICY
}
`, rName, rName, rName, rName, rName, rName, rName, rName)
}

func testAccMskSaslScramSecretDelete(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = "%s-kms-key-msk"
}

resource "aws_secretsmanager_secret" "test" {
  name       = "AmazonMSK_test_%s_1_delete"
  kms_key_id = aws_kms_key.test.key_id
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = jsonencode({ username = "%s", password = "foobar" })
}

resource "aws_secretsmanager_secret_policy" "test" {
  secret_arn = aws_secretsmanager_secret.test.arn
  policy     = <<POLICY
{
  "Version" : "2012-10-17",
  "Statement" : [ {
    "Sid": "AWSKafkaResourcePolicy",
    "Effect" : "Allow",
    "Principal" : {
      "Service" : "kafka.amazonaws.com"
    },
    "Action" : "secretsmanager:getSecretValue",
    "Resource" : "${aws_secretsmanager_secret.test.arn}"
  } ]
}
POLICY
}

`, rName, rName, rName)
}
