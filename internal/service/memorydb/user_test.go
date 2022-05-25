package memorydb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/memorydb"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfmemorydb "github.com/hashicorp/terraform-provider-aws/internal/service/memorydb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccMemoryDBUser_basic(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "access_string", "on ~* &* +@all"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "memorydb", "user/"+rName),
					resource.TestCheckResourceAttr(resourceName, "authentication_mode.0.password_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_mode.0.passwords.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "authentication_mode.0.passwords.*", "aaaaaaaaaaaaaaaa"),
					resource.TestCheckResourceAttr(resourceName, "authentication_mode.0.type", "password"),
					resource.TestCheckResourceAttrSet(resourceName, "minimum_engine_version"),
					resource.TestCheckResourceAttr(resourceName, "user_name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Test", "test"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"authentication_mode.0.passwords"},
			},
		},
	})
}

func TestAccMemoryDBUser_disappears(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfmemorydb.ResourceUser(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccMemoryDBUser_update_accessString(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_accessString(rName, "on ~* &* +@all"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "access_string", "on ~* &* +@all"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"authentication_mode.0.passwords"},
			},
			{
				Config: testAccUserConfig_accessString(rName, "off ~* &* +@all"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "access_string", "off ~* &* +@all"),
				),
			},
		},
	})
}

func TestAccMemoryDBUser_update_passwords(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_passwords2(rName, "aaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbb"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "authentication_mode.0.password_count", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"authentication_mode.0.passwords"},
			},
			{
				Config: testAccUserConfig_passwords1(rName, "aaaaaaaaaaaaaaaa"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "authentication_mode.0.password_count", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"authentication_mode.0.passwords"},
			},
			{
				Config: testAccUserConfig_passwords2(rName, "cccccccccccccccc", "dddddddddddddddd"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "authentication_mode.0.password_count", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"authentication_mode.0.passwords"},
			},
		},
	})
}

func TestAccMemoryDBUser_update_tags(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_tags0(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"authentication_mode.0.passwords"},
			},
			{
				Config: testAccUserConfig_tags2(rName, "Key1", "value1", "Key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key2", "value2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"authentication_mode.0.passwords"},
			},
			{
				Config: testAccUserConfig_tags1(rName, "Key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"authentication_mode.0.passwords"},
			},
			{
				Config: testAccUserConfig_tags0(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"authentication_mode.0.passwords"},
			},
		},
	})
}

func testAccCheckUserDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MemoryDBConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_memorydb_user" {
			continue
		}

		_, err := tfmemorydb.FindUserByName(context.Background(), conn, rs.Primary.Attributes["user_name"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("MemoryDB User %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckUserExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No MemoryDB User ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MemoryDBConn

		_, err := tfmemorydb.FindUserByName(context.Background(), conn, rs.Primary.Attributes["user_name"])

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccUserConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_user" "test" {
  access_string = "on ~* &* +@all"
  user_name     = %[1]q

  authentication_mode {
    type      = "password"
    passwords = ["aaaaaaaaaaaaaaaa"]
  }

  tags = {
    Test = "test"
  }
}
`, rName)
}

func testAccUserConfig_accessString(rName, accessString string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_user" "test" {
  access_string = %[2]q
  user_name     = %[1]q

  authentication_mode {
    type      = "password"
    passwords = ["aaaaaaaaaaaaaaaa"]
  }
}
`, rName, accessString)
}

func testAccUserConfig_passwords1(rName, password1 string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_user" "test" {
  access_string = "on ~* &* +@all"
  user_name     = %[1]q

  authentication_mode {
    type      = "password"
    passwords = [%[2]q]
  }
}
`, rName, password1)
}

func testAccUserConfig_passwords2(rName, password1, password2 string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_user" "test" {
  access_string = "on ~* &* +@all"
  user_name     = %[1]q

  authentication_mode {
    type      = "password"
    passwords = [%[2]q, %[3]q]
  }
}
`, rName, password1, password2)
}

func testAccUserConfig_tags0(rName string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_user" "test" {
  access_string = "on ~* &* +@all"
  user_name     = %[1]q

  authentication_mode {
    type      = "password"
    passwords = ["aaaaaaaaaaaaaaaa"]
  }
}
`, rName)
}

func testAccUserConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_user" "test" {
  access_string = "on ~* &* +@all"
  user_name     = %[1]q

  authentication_mode {
    type      = "password"
    passwords = ["aaaaaaaaaaaaaaaa"]
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccUserConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_user" "test" {
  access_string = "on ~* &* +@all"
  user_name     = %[1]q

  authentication_mode {
    type      = "password"
    passwords = ["aaaaaaaaaaaaaaaa"]
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
