package ec2_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccEC2KeyPair_basic(t *testing.T) {
	var keyPair ec2.KeyPairInfo
	resourceName := "aws_key_pair.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckKeyPairDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPairConfig(rName, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyPairExists(resourceName, &keyPair),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "ec2", fmt.Sprintf("key-pair/%s", rName)),
					resource.TestMatchResourceAttr(resourceName, "fingerprint", regexp.MustCompile(`[a-f0-9]{2}(:[a-f0-9]{2}){15}`)),
					resource.TestCheckResourceAttr(resourceName, "key_name", rName),
					resource.TestCheckResourceAttr(resourceName, "public_key", publicKey),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"public_key"},
			},
		},
	})
}

func TestAccEC2KeyPair_tags(t *testing.T) {
	var keyPair ec2.KeyPairInfo
	resourceName := "aws_key_pair.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckKeyPairDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPairTags1Config(rName, publicKey, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyPairExists(resourceName, &keyPair),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"public_key"},
			},
			{
				Config: testAccKeyPairTags2Config(rName, publicKey, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyPairExists(resourceName, &keyPair),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccKeyPairTags1Config(rName, publicKey, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyPairExists(resourceName, &keyPair),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEC2KeyPair_generatedName(t *testing.T) {
	var keyPair ec2.KeyPairInfo
	resourceName := "aws_key_pair.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckKeyPairDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPairConfig_generatedName(publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyPairExists(resourceName, &keyPair),
					testAccCheckKeyPairKeyNamePrefix(&keyPair, "terraform-"),
					resource.TestMatchResourceAttr(resourceName, "key_name", regexp.MustCompile(`^terraform-`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"public_key"},
			},
		},
	})
}

func TestAccEC2KeyPair_namePrefix(t *testing.T) {
	var keyPair ec2.KeyPairInfo
	resourceName := "aws_key_pair.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckKeyPairDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckKeyPairPrefixNameConfig(publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyPairExists(resourceName, &keyPair),
					testAccCheckKeyPairKeyNamePrefix(&keyPair, "baz-"),
					resource.TestMatchResourceAttr(resourceName, "key_name", regexp.MustCompile(`^baz-`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"key_name_prefix", "public_key"},
			},
		},
	})
}

func TestAccEC2KeyPair_disappears(t *testing.T) {
	var keyPair ec2.KeyPairInfo
	resourceName := "aws_key_pair.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckKeyPairDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPairConfig(rName, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyPairExists(resourceName, &keyPair),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceKeyPair(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckKeyPairDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_key_pair" {
			continue
		}

		// Try to find key pair
		resp, err := conn.DescribeKeyPairs(&ec2.DescribeKeyPairsInput{
			KeyNames: []*string{aws.String(rs.Primary.ID)},
		})
		if err == nil {
			if len(resp.KeyPairs) > 0 {
				return fmt.Errorf("still exist.")
			}
			return nil
		}

		if !tfawserr.ErrMessageContains(err, "InvalidKeyPair.NotFound", "") {
			return err
		}
	}

	return nil
}

func testAccCheckKeyPairKeyNamePrefix(conf *ec2.KeyPairInfo, namePrefix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !strings.HasPrefix(aws.StringValue(conf.KeyName), namePrefix) {
			return fmt.Errorf("incorrect key name. expected %s prefix, got %s", namePrefix, aws.StringValue(conf.KeyName))
		}
		return nil
	}
}

func testAccCheckKeyPairExists(n string, res *ec2.KeyPairInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No KeyPair name is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		resp, err := conn.DescribeKeyPairs(&ec2.DescribeKeyPairsInput{
			KeyNames: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}
		if len(resp.KeyPairs) != 1 ||
			aws.StringValue(resp.KeyPairs[0].KeyName) != rs.Primary.ID {
			return fmt.Errorf("KeyPair not found")
		}

		*res = *resp.KeyPairs[0]

		return nil
	}
}

func testAccKeyPairConfig(rName, publicKey string) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = %[2]q
}
`, rName, publicKey)
}

func testAccKeyPairTags1Config(rName, publicKey, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = %[2]q

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, publicKey, tagKey1, tagValue1)
}

func testAccKeyPairTags2Config(rName, publicKey, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = %[2]q

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, publicKey, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccKeyPairConfig_generatedName(publicKey string) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "test" {
  public_key = %[1]q
}
`, publicKey)
}

func testAccCheckKeyPairPrefixNameConfig(publicKey string) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "test" {
  key_name_prefix = "baz-"
  public_key      = %[1]q
}
`, publicKey)
}
