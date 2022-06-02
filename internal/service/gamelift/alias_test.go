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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccGameLiftAlias_basic(t *testing.T) {
	var conf gamelift.Alias

	rString := sdkacctest.RandString(8)
	resourceName := "aws_gamelift_alias.test"

	aliasName := fmt.Sprintf("tf_acc_alias_%s", rString)
	description := fmt.Sprintf("tf test description %s", rString)
	message := fmt.Sprintf("tf test message %s", rString)

	uAliasName := fmt.Sprintf("tf_acc_alias_upd_%s", rString)
	uDescription := fmt.Sprintf("tf test updated description %s", rString)
	uMessage := fmt.Sprintf("tf test updated message %s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig_basic(aliasName, description, message),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "gamelift", regexp.MustCompile(`alias/alias-.+`)),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.0.message", message),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.0.type", "TERMINAL"),
					resource.TestCheckResourceAttr(resourceName, "name", aliasName),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAliasConfig_basic(uAliasName, uDescription, uMessage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "gamelift", regexp.MustCompile(`alias/.+`)),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.0.message", uMessage),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.0.type", "TERMINAL"),
					resource.TestCheckResourceAttr(resourceName, "name", uAliasName),
					resource.TestCheckResourceAttr(resourceName, "description", uDescription),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccGameLiftAlias_tags(t *testing.T) {
	var conf gamelift.Alias

	resourceName := "aws_gamelift_alias.test"
	aliasName := sdkacctest.RandomWithPrefix("tf-acc-alias")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig_basicTags1(aliasName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(resourceName, &conf),
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
				Config: testAccAliasConfig_basicTags2(aliasName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAliasConfig_basicTags1(aliasName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccGameLiftAlias_fleetRouting(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf gamelift.Alias

	rString := sdkacctest.RandString(8)

	aliasName := fmt.Sprintf("tf_acc_alias_%s", rString)
	description := fmt.Sprintf("tf test description %s", rString)
	fleetName := fmt.Sprintf("tf_acc_fleet_%s", rString)

	region := acctest.Region()
	g, err := testAccSampleGame(region)

	if tfresource.NotFound(err) {
		t.Skip(err)
	}

	if err != nil {
		t.Fatal(err)
	}

	loc := g.Location
	bucketName := *loc.Bucket
	roleArn := *loc.RoleArn
	key := *loc.Key

	launchPath := g.LaunchPath
	params := g.Parameters(33435)
	resourceName := "aws_gamelift_alias.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig_allFields(aliasName, description,
					fleetName, launchPath, params, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "gamelift", regexp.MustCompile(`alias/alias-.+`)),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "routing_strategy.0.fleet_id"),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.0.type", "SIMPLE"),
					resource.TestCheckResourceAttr(resourceName, "name", aliasName),
					resource.TestCheckResourceAttr(resourceName, "description", description),
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

func TestAccGameLiftAlias_disappears(t *testing.T) {
	var conf gamelift.Alias

	rString := sdkacctest.RandString(8)
	resourceName := "aws_gamelift_alias.test"

	aliasName := fmt.Sprintf("tf_acc_alias_%s", rString)
	description := fmt.Sprintf("tf test description %s", rString)
	message := fmt.Sprintf("tf test message %s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig_basic(aliasName, description, message),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(resourceName, &conf),
					testAccCheckAliasDisappears(&conf),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAliasDisappears(res *gamelift.Alias) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftConn

		input := &gamelift.DeleteAliasInput{AliasId: res.AliasId}

		_, err := conn.DeleteAlias(input)

		return err
	}
}

func testAccCheckAliasExists(n string, res *gamelift.Alias) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No GameLift Alias ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftConn

		out, err := conn.DescribeAlias(&gamelift.DescribeAliasInput{
			AliasId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}
		a := out.Alias

		if *a.AliasId != rs.Primary.ID {
			return fmt.Errorf("GameLift Alias not found")
		}

		*res = *a

		return nil
	}
}

func testAccCheckAliasDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_gamelift_alias" {
			continue
		}

		_, err := conn.DescribeAlias(&gamelift.DescribeAliasInput{
			AliasId: aws.String(rs.Primary.ID),
		})
		if err == nil {
			return fmt.Errorf("GameLift Alias still exists")
		}

		if tfawserr.ErrCodeEquals(err, gamelift.ErrCodeNotFoundException) {
			return nil
		}

		return err
	}

	return nil
}

func testAccAliasConfig_basic(aliasName, description, message string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_alias" "test" {
  name        = "%s"
  description = "%s"

  routing_strategy {
    message = "%s"
    type    = "TERMINAL"
  }
}
`, aliasName, description, message)
}

func testAccAliasConfig_basicTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_alias" "test" {
  name        = %[1]q
  description = "foo"

  routing_strategy {
    message = "bar"
    type    = "TERMINAL"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAliasConfig_basicTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_alias" "test" {
  name        = %[1]q
  description = "foo"

  routing_strategy {
    message = "bar"
    type    = "TERMINAL"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAliasConfig_allFields(aliasName, description,
	fleetName, launchPath, params, bucketName, key, roleArn string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_alias" "test" {
  name        = "%s"
  description = "%s"

  routing_strategy {
    fleet_id = aws_gamelift_fleet.test.id
    type     = "SIMPLE"
  }
}
%s
`, aliasName, description,
		testAccFleetConfig_basic(fleetName, launchPath, params, bucketName, key, roleArn))
}
