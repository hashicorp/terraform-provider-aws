package rds_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccRDSOptionGroup_basic(t *testing.T) {
	var v rds.OptionGroup
	rName := fmt.Sprintf("option-group-test-terraform-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckOptionGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists("aws_db_option_group.bar", &v),
					acctest.MatchResourceAttrRegionalARN("aws_db_option_group.bar", "arn", "rds", regexp.MustCompile(`og:.+`)),
					resource.TestCheckResourceAttr("aws_db_option_group.bar", "engine_name", "mysql"),
					resource.TestCheckResourceAttr("aws_db_option_group.bar", "major_engine_version", "5.6"),
					resource.TestCheckResourceAttr("aws_db_option_group.bar", "name", rName),
					resource.TestCheckResourceAttr("aws_db_option_group.bar", "option.#", "0"),
					resource.TestCheckResourceAttr("aws_db_option_group.bar", "option_group_description", "Managed by Terraform"),
					resource.TestCheckResourceAttr("aws_db_option_group.bar", "tags.%", "0"),
				),
			},
			{
				ResourceName:            "aws_db_option_group.bar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccRDSOptionGroup_timeoutBlock(t *testing.T) {
	var v rds.OptionGroup
	rName := fmt.Sprintf("option-group-test-terraform-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckOptionGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupBasicTimeoutBlockConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists("aws_db_option_group.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_db_option_group.bar", "name", rName),
				),
			},
			{
				ResourceName:            "aws_db_option_group.bar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccRDSOptionGroup_namePrefix(t *testing.T) {
	var v rds.OptionGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckOptionGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroup_namePrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists("aws_db_option_group.test", &v),
					resource.TestMatchResourceAttr(
						"aws_db_option_group.test", "name", regexp.MustCompile("^tf-test-")),
				),
			},
			{
				ResourceName:            "aws_db_option_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccRDSOptionGroup_generatedName(t *testing.T) {
	var v rds.OptionGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckOptionGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroup_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists("aws_db_option_group.test", &v),
				),
			},
			{
				ResourceName:            "aws_db_option_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccRDSOptionGroup_optionGroupDescription(t *testing.T) {
	var optionGroup1 rds.OptionGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_option_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckOptionGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupOptionGroupDescriptionConfig(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(resourceName, &optionGroup1),
					resource.TestCheckResourceAttr(resourceName, "option_group_description", "description1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccRDSOptionGroup_basicDestroyWithInstance(t *testing.T) {
	rName := fmt.Sprintf("option-group-test-terraform-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckOptionGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupBasicDestroyConfig(rName),
			},
			{
				ResourceName:            "aws_db_option_group.bar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccRDSOptionGroup_Option_optionSettings(t *testing.T) {
	var v rds.OptionGroup
	rName := fmt.Sprintf("option-group-test-terraform-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckOptionGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupOptionSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists("aws_db_option_group.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_db_option_group.bar", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_db_option_group.bar", "option.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("aws_db_option_group.bar", "option.*.option_settings.*", map[string]string{
						"value": "UTC",
					}),
				),
			},
			{
				ResourceName:      "aws_db_option_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
				// Ignore option since our current logic skips "unconfigured" default option settings
				// Even with Config set, ImportState TestStep does not "see" the configuration to check against
				ImportStateVerifyIgnore: []string{"name_prefix", "option"},
			},
			{
				Config: testAccOptionGroupOptionSettings_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists("aws_db_option_group.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_db_option_group.bar", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_db_option_group.bar", "option.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("aws_db_option_group.bar", "option.*.option_settings.*", map[string]string{
						"value": "US/Pacific",
					}),
				),
			},
			// Ensure we can import non-default value option settings
			{
				ResourceName:            "aws_db_option_group.bar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccRDSOptionGroup_OptionOptionSettings_iamRole(t *testing.T) {
	var v rds.OptionGroup
	rName := fmt.Sprintf("option-group-test-terraform-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckOptionGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupOptionSettingsIAMRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists("aws_db_option_group.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_db_option_group.bar", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_db_option_group.bar", "option.#", "1"),
					testAccCheckOptionGroupOptionSettingsIAMRole(&v),
				),
			},
			{
				ResourceName:            "aws_db_option_group.bar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccRDSOptionGroup_sqlServerOptionsUpdate(t *testing.T) {
	var v rds.OptionGroup
	rName := fmt.Sprintf("option-group-test-terraform-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckOptionGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupSQLServerEEOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists("aws_db_option_group.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_db_option_group.bar", "name", rName),
				),
			},
			{
				ResourceName:            "aws_db_option_group.bar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
			{
				Config: testAccOptionGroupSQLServerEEOptions_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists("aws_db_option_group.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_db_option_group.bar", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_db_option_group.bar", "option.#", "1"),
				),
			},
		},
	})
}

func TestAccRDSOptionGroup_oracleOptionsUpdate(t *testing.T) {
	var v rds.OptionGroup
	rName := fmt.Sprintf("option-group-test-terraform-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckOptionGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupOracleEEOptionSettings(rName, "13.2.0.0.v2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists("aws_db_option_group.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_db_option_group.bar", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_db_option_group.bar", "option.#", "1"),
					testAccCheckOptionGroupOptionVersionAttribute(&v, "13.2.0.0.v2"),
				),
			},
			{
				ResourceName:      "aws_db_option_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
				// Ignore option since API responds with **** instead of password
				ImportStateVerifyIgnore: []string{"name_prefix", "option"},
			},
			{
				Config: testAccOptionGroupOracleEEOptionSettings(rName, "13.3.0.0.v2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists("aws_db_option_group.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_db_option_group.bar", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_db_option_group.bar", "option.#", "1"),
					testAccCheckOptionGroupOptionVersionAttribute(&v, "13.3.0.0.v2"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/1876
func TestAccRDSOptionGroup_OptionOptionSettings_multipleNonDefault(t *testing.T) {
	var optionGroup1, optionGroup2 rds.OptionGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_option_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckOptionGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupOptionOptionSettingsMultipleConfig(rName, "example1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(resourceName, &optionGroup1),
					resource.TestCheckResourceAttr(resourceName, "option.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
			{
				Config: testAccOptionGroupOptionOptionSettingsMultipleConfig(rName, "example1,example2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(resourceName, &optionGroup2),
					resource.TestCheckResourceAttr(resourceName, "option.#", "1"),
				),
			},
		},
	})
}

func TestAccRDSOptionGroup_multipleOptions(t *testing.T) {
	var v rds.OptionGroup
	rName := fmt.Sprintf("option-group-test-terraform-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckOptionGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupMultipleOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists("aws_db_option_group.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_db_option_group.bar", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_db_option_group.bar", "option.#", "2"),
				),
			},
			{
				ResourceName:            "aws_db_option_group.bar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccRDSOptionGroup_tags(t *testing.T) {
	var optionGroup1, optionGroup2, optionGroup3 rds.OptionGroup
	resourceName := "aws_db_option_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckOptionGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(resourceName, &optionGroup1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
			{
				Config: testAccOptionGroupTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(resourceName, &optionGroup2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccOptionGroupTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(resourceName, &optionGroup3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/7114
func TestAccRDSOptionGroup_Tags_withOptions(t *testing.T) {
	var optionGroup1, optionGroup2, optionGroup3 rds.OptionGroup
	resourceName := "aws_db_option_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckOptionGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupTagsWithOption1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(resourceName, &optionGroup1),
					resource.TestCheckResourceAttr(resourceName, "option.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
			{
				Config: testAccOptionGroupTagsWithOption2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(resourceName, &optionGroup2),
					resource.TestCheckResourceAttr(resourceName, "option.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccOptionGroupTagsWithOption1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(resourceName, &optionGroup3),
					resource.TestCheckResourceAttr(resourceName, "option.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckOptionGroupOptionSettingsIAMRole(optionGroup *rds.OptionGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if optionGroup == nil {
			return errors.New("Option Group does not exist")
		}
		if len(optionGroup.Options) == 0 {
			return errors.New("Option Group does not have any options")
		}
		if len(optionGroup.Options[0].OptionSettings) == 0 {
			return errors.New("Option Group does not have any option settings")
		}

		settingName := aws.StringValue(optionGroup.Options[0].OptionSettings[0].Name)
		if settingName != "IAM_ROLE_ARN" {
			return fmt.Errorf("Expected option setting IAM_ROLE_ARN and received %s", settingName)
		}

		settingValue := aws.StringValue(optionGroup.Options[0].OptionSettings[0].Value)
		iamArnRegExp := regexp.MustCompile(fmt.Sprintf(`^arn:%s:iam::\d{12}:role/.+`, acctest.Partition()))
		if !iamArnRegExp.MatchString(settingValue) {
			return fmt.Errorf("Expected option setting to be a valid IAM role but received %s", settingValue)
		}
		return nil
	}
}

func testAccCheckOptionGroupOptionVersionAttribute(optionGroup *rds.OptionGroup, optionVersion string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if optionGroup == nil {
			return errors.New("Option Group does not exist")
		}
		if len(optionGroup.Options) == 0 {
			return errors.New("Option Group does not have any options")
		}
		foundOptionVersion := aws.StringValue(optionGroup.Options[0].OptionVersion)
		if foundOptionVersion != optionVersion {
			return fmt.Errorf("Expected option version %q and received %q", optionVersion, foundOptionVersion)
		}
		return nil
	}
}

func testAccCheckOptionGroupExists(n string, v *rds.OptionGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DB Option Group Name is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

		opts := rds.DescribeOptionGroupsInput{
			OptionGroupName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeOptionGroups(&opts)

		if err != nil {
			return err
		}

		if len(resp.OptionGroupsList) != 1 ||
			*resp.OptionGroupsList[0].OptionGroupName != rs.Primary.ID {
			return fmt.Errorf("DB Option Group not found")
		}

		*v = *resp.OptionGroupsList[0]

		return nil
	}
}

func testAccCheckOptionGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_option_group" {
			continue
		}

		resp, err := conn.DescribeOptionGroups(
			&rds.DescribeOptionGroupsInput{
				OptionGroupName: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.OptionGroupsList) != 0 &&
				*resp.OptionGroupsList[0].OptionGroupName == rs.Primary.ID {
				return fmt.Errorf("DB Option Group still exists")
			}
		}

		// Verify the error
		if !tfawserr.ErrMessageContains(err, rds.ErrCodeOptionGroupNotFoundFault, "") {
			return err
		}
	}

	return nil
}

func testAccOptionGroupBasicTimeoutBlockConfig(r string) string {
	return fmt.Sprintf(`
resource "aws_db_option_group" "bar" {
  name                     = "%s"
  option_group_description = "Test option group for terraform"
  engine_name              = "mysql"
  major_engine_version     = "5.6"

  timeouts {
    delete = "10m"
  }
}
`, r)
}

func testAccOptionGroupBasicConfig(r string) string {
	return fmt.Sprintf(`
resource "aws_db_option_group" "bar" {
  name                 = "%s"
  engine_name          = "mysql"
  major_engine_version = "5.6"
}
`, r)
}

func testAccOptionGroupBasicDestroyConfig(r string) string {
	return fmt.Sprintf(`
resource "aws_db_instance" "bar" {
  allocated_storage = 10
  engine            = "mysql"
  engine_version    = "5.6.35"
  instance_class    = "db.t2.micro"
  name              = "baz"
  password          = "barbarbarbar"
  username          = "foo"

  # Maintenance Window is stored in lower case in the API, though not strictly
  # documented. Terraform will downcase this to match (as opposed to throw a
  # validation error).
  maintenance_window = "Fri:09:00-Fri:09:30"

  backup_retention_period = 0
  skip_final_snapshot     = true

  option_group_name = aws_db_option_group.bar.name
}

resource "aws_db_option_group" "bar" {
  name                     = "%s"
  option_group_description = "Test option group for terraform"
  engine_name              = "mysql"
  major_engine_version     = "5.6"
}
`, r)
}

func testAccOptionGroupOptionSettings(r string) string {
	return fmt.Sprintf(`
resource "aws_db_option_group" "bar" {
  name                     = "%s"
  option_group_description = "Test option group for terraform"
  engine_name              = "oracle-ee"
  major_engine_version     = "12.2"

  option {
    option_name = "Timezone"

    option_settings {
      name  = "TIME_ZONE"
      value = "UTC"
    }
  }
}
`, r)
}

func testAccOptionGroupOptionSettingsIAMRole(r string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "rds_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["rds.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role" "sql_server_backup" {
  name               = "rds-backup-%[1]s"
  assume_role_policy = data.aws_iam_policy_document.rds_assume_role.json
}

resource "aws_db_option_group" "bar" {
  name                     = "%[1]s"
  option_group_description = "Test option group for terraform"
  engine_name              = "sqlserver-ex"
  major_engine_version     = "14.00"

  option {
    option_name = "SQLSERVER_BACKUP_RESTORE"

    option_settings {
      name  = "IAM_ROLE_ARN"
      value = aws_iam_role.sql_server_backup.arn
    }
  }
}
`, r)
}

func testAccOptionGroupOptionSettings_update(r string) string {
	return fmt.Sprintf(`
resource "aws_db_option_group" "bar" {
  name                     = "%s"
  option_group_description = "Test option group for terraform"
  engine_name              = "oracle-ee"
  major_engine_version     = "12.2"

  option {
    option_name = "Timezone"

    option_settings {
      name  = "TIME_ZONE"
      value = "US/Pacific"
    }
  }
}
`, r)
}

func testAccOptionGroupSQLServerEEOptions(r string) string {
	return fmt.Sprintf(`
resource "aws_db_option_group" "bar" {
  name                     = "%s"
  option_group_description = "Test option group for terraform"
  engine_name              = "sqlserver-ee"
  major_engine_version     = "11.00"
}
`, r)
}

func testAccOptionGroupSQLServerEEOptions_update(r string) string {
	return fmt.Sprintf(`
resource "aws_db_option_group" "bar" {
  name                     = "%s"
  option_group_description = "Test option group for terraform"
  engine_name              = "sqlserver-ee"
  major_engine_version     = "11.00"

  option {
    option_name = "TDE"
  }
}
`, r)
}

func testAccOptionGroupOracleEEOptionSettings(r, optionVersion string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "foo" {
  name = "%[1]s"
}

resource "aws_db_option_group" "bar" {
  name                     = "%[1]s"
  option_group_description = "Test option group for terraform issue 748"
  engine_name              = "oracle-ee"
  major_engine_version     = "12.2"

  option {
    option_name = "OEM_AGENT"
    port        = "3872"
    version     = "%[2]s"

    vpc_security_group_memberships = [aws_security_group.foo.id]

    option_settings {
      name  = "OMS_PORT"
      value = "4903"
    }

    option_settings {
      name  = "OMS_HOST"
      value = "oem.host.value"
    }

    option_settings {
      name  = "AGENT_REGISTRATION_PASSWORD"
      value = "password"
    }
  }
}
`, r, optionVersion)
}

func testAccOptionGroupMultipleOptions(r string) string {
	return fmt.Sprintf(`
resource "aws_db_option_group" "bar" {
  name                     = "%s"
  option_group_description = "Test option group for terraform"
  engine_name              = "oracle-ee"
  major_engine_version     = "12.2"

  option {
    option_name = "SPATIAL"
  }

  option {
    option_name = "STATSPACK"
  }
}
`, r)
}

const testAccOptionGroup_namePrefix = `
resource "aws_db_option_group" "test" {
  name_prefix              = "tf-test-"
  option_group_description = "Test option group for terraform"
  engine_name              = "mysql"
  major_engine_version     = "5.6"
}
`

const testAccOptionGroup_generatedName = `
resource "aws_db_option_group" "test" {
  option_group_description = "Test option group for terraform"
  engine_name              = "mysql"
  major_engine_version     = "5.6"
}
`

func testAccOptionGroupOptionGroupDescriptionConfig(rName, optionGroupDescription string) string {
	return fmt.Sprintf(`
resource "aws_db_option_group" "test" {
  engine_name              = "mysql"
  major_engine_version     = "5.6"
  name                     = %q
  option_group_description = %q
}
`, rName, optionGroupDescription)
}

func testAccOptionGroupOptionOptionSettingsMultipleConfig(rName, value string) string {
	return fmt.Sprintf(`
resource "aws_db_option_group" "test" {
  engine_name          = "mysql"
  major_engine_version = "5.6"
  name                 = %q

  option {
    option_name = "MARIADB_AUDIT_PLUGIN"

    option_settings {
      name  = "SERVER_AUDIT_EXCL_USERS"
      value = %q
    }

    option_settings {
      name  = "SERVER_AUDIT_FILE_ROTATIONS"
      value = "15"
    }

    option_settings {
      name  = "SERVER_AUDIT_FILE_ROTATE_SIZE"
      value = "52428800"
    }
  }
}
`, rName, value)
}

func testAccOptionGroupTags1Config(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_db_option_group" "test" {
  engine_name          = "mysql"
  major_engine_version = "5.6"
  name                 = %q

  tags = {
    %q = %q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccOptionGroupTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_db_option_group" "test" {
  engine_name          = "mysql"
  major_engine_version = "5.6"
  name                 = %q

  tags = {
    %q = %q
    %q = %q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccOptionGroupTagsWithOption1Config(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_db_option_group" "test" {
  engine_name          = "mysql"
  major_engine_version = "5.6"
  name                 = %q

  option {
    option_name = "MARIADB_AUDIT_PLUGIN"

    option_settings {
      name  = "SERVER_AUDIT_FILE_ROTATIONS"
      value = "0"
    }
  }

  tags = {
    %q = %q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccOptionGroupTagsWithOption2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_db_option_group" "test" {
  engine_name          = "mysql"
  major_engine_version = "5.6"
  name                 = %q

  option {
    option_name = "MARIADB_AUDIT_PLUGIN"

    option_settings {
      name  = "SERVER_AUDIT_FILE_ROTATIONS"
      value = "0"
    }
  }

  tags = {
    %q = %q
    %q = %q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
