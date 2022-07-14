package gamelift_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfgamelift "github.com/hashicorp/terraform-provider-aws/internal/service/gamelift"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccGameLiftScript_basic(t *testing.T) {
	var conf gamelift.Script
	resourceName := "aws_gamelift_script.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	region := acctest.Region()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckScriptDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScriptConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScriptExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`script/script-.+`)),
					resource.TestCheckResourceAttr(resourceName, "storage_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.bucket", fmt.Sprintf("prod-gamescale-scripts-%s", region)),
					resource.TestCheckResourceAttrSet(resourceName, "storage_location.0.key"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file"},
			},
			{
				Config: testAccScriptConfig_basic(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScriptExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`script/script-.+`)),
					resource.TestCheckResourceAttr(resourceName, "storage_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccGameLiftScript_tags(t *testing.T) {
	var conf gamelift.Script

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_gamelift_script.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckScriptDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScriptConfig_basicTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScriptExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file"},
			},
			{
				Config: testAccScriptConfig_basicTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScriptExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccScriptConfig_basicTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScriptExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccGameLiftScript_disappears(t *testing.T) {
	var conf gamelift.Script

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_gamelift_script.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckScriptDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScriptConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScriptExists(resourceName, &conf),
					acctest.CheckResourceDisappears(acctest.Provider, tfgamelift.ResourceScript(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfgamelift.ResourceScript(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckScriptExists(n string, res *gamelift.Script) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No GameLift Script ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftConn

		script, err := tfgamelift.FindScriptByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if aws.StringValue(script.ScriptId) != rs.Primary.ID {
			return fmt.Errorf("GameLift Script not found")
		}

		*res = *script

		return nil
	}
}

func testAccCheckScriptDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_gamelift_script" {
			continue
		}

		script, err := tfgamelift.FindScriptByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if script != nil {
			return fmt.Errorf("GameLift Script (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccScriptConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_script" "test" {
  name     = %[1]q
  zip_file = "test-fixtures/script.zip"
}
`, rName)
}

func testAccScriptConfig_basicTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_script" "test" {
  name     = %[1]q
  zip_file = "test-fixtures/script.zip"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccScriptConfig_basicTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_script" "test" {
  name     = %[1]q
  zip_file = "test-fixtures/script.zip"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
