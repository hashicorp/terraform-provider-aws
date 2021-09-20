package kms_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/kms"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfkms "github.com/hashicorp/terraform-provider-aws/internal/service/kms"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccAWSKmsAlias_basic(t *testing.T) {
	var alias kms.AliasListEntry
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_kms_alias.test"
	keyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSKmsAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsAliasConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsAliasExists(resourceName, &alias),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "kms", regexp.MustCompile(`alias/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", tfkms.AliasNamePrefix+rName),
					resource.TestCheckResourceAttrPair(resourceName, "target_key_arn", keyResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target_key_id", keyResourceName, "id"),
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

func TestAccAWSKmsAlias_disappears(t *testing.T) {
	var alias kms.AliasListEntry
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_kms_alias.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSKmsAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsAliasConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsAliasExists(resourceName, &alias),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceAlias(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSKmsAlias_Name_Generated(t *testing.T) {
	var alias kms.AliasListEntry
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_kms_alias.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSKmsAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsAliasConfigNameGenerated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsAliasExists(resourceName, &alias),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile(fmt.Sprintf("%s[[:xdigit:]]{%d}", tfkms.AliasNamePrefix, resource.UniqueIDSuffixLength))),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", tfkms.AliasNamePrefix),
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

func TestAccAWSKmsAlias_NamePrefix(t *testing.T) {
	var alias kms.AliasListEntry
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_kms_alias.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSKmsAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsAliasConfigNamePrefix(rName, tfkms.AliasNamePrefix+"tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsAliasExists(resourceName, &alias),
					create.TestCheckResourceAttrNameFromPrefix(resourceName, "name", tfkms.AliasNamePrefix+"tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", tfkms.AliasNamePrefix+"tf-acc-test-prefix-"),
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

func TestAccAWSKmsAlias_UpdateKeyID(t *testing.T) {
	var alias kms.AliasListEntry
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_kms_alias.test"
	key1ResourceName := "aws_kms_key.test"
	key2ResourceName := "aws_kms_key.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSKmsAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsAliasConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsAliasExists(resourceName, &alias),
					resource.TestCheckResourceAttrPair(resourceName, "target_key_arn", key1ResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target_key_id", key1ResourceName, "id"),
				),
			},
			{
				Config: testAccAWSKmsAliasConfigUpdatedKeyId(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsAliasExists(resourceName, &alias),
					resource.TestCheckResourceAttrPair(resourceName, "target_key_arn", key2ResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target_key_id", key2ResourceName, "id"),
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

func TestAccAWSKmsAlias_MultipleAliasesForSameKey(t *testing.T) {
	var alias kms.AliasListEntry
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_kms_alias.test"
	alias2ResourceName := "aws_kms_alias.test2"
	keyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSKmsAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsAliasConfigMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsAliasExists(resourceName, &alias),
					resource.TestCheckResourceAttrPair(resourceName, "target_key_arn", keyResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target_key_id", keyResourceName, "id"),
					testAccCheckAWSKmsAliasExists(alias2ResourceName, &alias),
					resource.TestCheckResourceAttrPair(alias2ResourceName, "target_key_arn", keyResourceName, "arn"),
					resource.TestCheckResourceAttrPair(alias2ResourceName, "target_key_id", keyResourceName, "id"),
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

func TestAccAWSKmsAlias_ArnDiffSuppress(t *testing.T) {
	var alias kms.AliasListEntry
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_kms_alias.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSKmsAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsAliasConfigDiffSuppress(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsAliasExists(resourceName, &alias),
					resource.TestCheckResourceAttrSet(resourceName, "target_key_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ExpectNonEmptyPlan: false,
				PlanOnly:           true,
				Config:             testAccAWSKmsAliasConfigDiffSuppress(rName),
			},
		},
	})
}

func testAccCheckAWSKmsAliasDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KMSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kms_alias" {
			continue
		}

		_, err := tfkms.FindAliasByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("KMS Alias %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSKmsAliasExists(name string, v *kms.AliasListEntry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No KMS Alias ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KMSConn

		output, err := tfkms.FindAliasByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAWSKmsAliasConfigName(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_kms_alias" "test" {
  name          = "alias/%[1]s"
  target_key_id = aws_kms_key.test.id
}
`, rName)
}

func testAccAWSKmsAliasConfigNameGenerated(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_kms_alias" "test" {
  target_key_id = aws_kms_key.test.id
}
`, rName)
}

func testAccAWSKmsAliasConfigNamePrefix(rName, namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_kms_alias" "test" {
  name_prefix   = %[2]q
  target_key_id = aws_kms_key.test.id
}
`, rName, namePrefix)
}

func testAccAWSKmsAliasConfigUpdatedKeyId(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_kms_key" "test2" {
  description             = "%[1]s-2"
  deletion_window_in_days = 7
}

resource "aws_kms_alias" "test" {
  name          = "alias/%[1]s"
  target_key_id = aws_kms_key.test2.id
}
`, rName)
}

func testAccAWSKmsAliasConfigMultiple(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_kms_alias" "test" {
  name          = "alias/%[1]s-1"
  target_key_id = aws_kms_key.test.key_id
}

resource "aws_kms_alias" "test2" {
  name          = "alias/%[1]s-2"
  target_key_id = aws_kms_key.test.key_id
}
`, rName)
}

func testAccAWSKmsAliasConfigDiffSuppress(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_kms_alias" "test" {
  name          = "alias/%[1]s"
  target_key_id = aws_kms_key.test.arn
}
`, rName)
}
