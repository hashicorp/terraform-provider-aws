package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSElasticacheParameterGroup_basic(t *testing.T) {
	var v elasticache.CacheParameterGroup
	rName := fmt.Sprintf("parameter-group-test-terraform-%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheParameterGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheParameterGroupExists("aws_elasticache_parameter_group.bar", &v),
					testAccCheckAWSElasticacheParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(
						"aws_elasticache_parameter_group.bar", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_elasticache_parameter_group.bar", "family", "redis2.8"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_parameter_group.bar", "description", "Managed by Terraform"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_parameter_group.bar", "parameter.283487565.name", "appendonly"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_parameter_group.bar", "parameter.283487565.value", "yes"),
				),
			},
			{
				Config: testAccAWSElasticacheParameterGroupAddParametersConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheParameterGroupExists("aws_elasticache_parameter_group.bar", &v),
					testAccCheckAWSElasticacheParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(
						"aws_elasticache_parameter_group.bar", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_elasticache_parameter_group.bar", "family", "redis2.8"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_parameter_group.bar", "description", "Test parameter group for terraform"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_parameter_group.bar", "parameter.283487565.name", "appendonly"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_parameter_group.bar", "parameter.283487565.value", "yes"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_parameter_group.bar", "parameter.2196914567.name", "appendfsync"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_parameter_group.bar", "parameter.2196914567.value", "always"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheParameterGroup_only(t *testing.T) {
	var v elasticache.CacheParameterGroup
	rName := fmt.Sprintf("parameter-group-test-terraform-%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheParameterGroupOnlyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheParameterGroupExists("aws_elasticache_parameter_group.bar", &v),
					testAccCheckAWSElasticacheParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(
						"aws_elasticache_parameter_group.bar", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_elasticache_parameter_group.bar", "family", "redis2.8"),
				),
			},
		},
	})
}

// Regression for https://github.com/terraform-providers/terraform-provider-aws/issues/116
func TestAccAWSElasticacheParameterGroup_removeParam(t *testing.T) {
	var v elasticache.CacheParameterGroup
	rName := fmt.Sprintf("parameter-group-test-terraform-%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheParameterGroupAddParametersConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheParameterGroupExists("aws_elasticache_parameter_group.bar", &v),
					testAccCheckAWSElasticacheParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(
						"aws_elasticache_parameter_group.bar", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_elasticache_parameter_group.bar", "family", "redis2.8"),
				),
			},
			{
				Config: testAccAWSElasticacheParameterGroupOnlyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheParameterGroupExists("aws_elasticache_parameter_group.bar", &v),
					testAccCheckAWSElasticacheParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(
						"aws_elasticache_parameter_group.bar", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_elasticache_parameter_group.bar", "family", "redis2.8"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheParameterGroup_UppercaseName(t *testing.T) {
	var v elasticache.CacheParameterGroup
	rInt := acctest.RandInt()
	rName := fmt.Sprintf("TF-ELASTIPG-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheParameterGroupConfig_UppercaseName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheParameterGroupExists("aws_elasticache_parameter_group.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_elasticache_parameter_group.bar", "name", fmt.Sprintf("tf-elastipg-%d", rInt)),
				),
			},
		},
	})
}

func testAccCheckAWSElasticacheParameterGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elasticacheconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticache_parameter_group" {
			continue
		}

		// Try to find the Group
		resp, err := conn.DescribeCacheParameterGroups(
			&elasticache.DescribeCacheParameterGroupsInput{
				CacheParameterGroupName: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.CacheParameterGroups) != 0 &&
				*resp.CacheParameterGroups[0].CacheParameterGroupName == rs.Primary.ID {
				return fmt.Errorf("Cache Parameter Group still exists")
			}
		}

		// Verify the error
		newerr, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if newerr.Code() != "CacheParameterGroupNotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckAWSElasticacheParameterGroupAttributes(v *elasticache.CacheParameterGroup, rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if *v.CacheParameterGroupName != rName {
			return fmt.Errorf("bad name: %#v", v.CacheParameterGroupName)
		}

		if *v.CacheParameterGroupFamily != "redis2.8" {
			return fmt.Errorf("bad family: %#v", v.CacheParameterGroupFamily)
		}

		return nil
	}
}

func testAccCheckAWSElasticacheParameterGroupExists(n string, v *elasticache.CacheParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Cache Parameter Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).elasticacheconn

		opts := elasticache.DescribeCacheParameterGroupsInput{
			CacheParameterGroupName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeCacheParameterGroups(&opts)

		if err != nil {
			return err
		}

		if len(resp.CacheParameterGroups) != 1 ||
			*resp.CacheParameterGroups[0].CacheParameterGroupName != rs.Primary.ID {
			return fmt.Errorf("Cache Parameter Group not found")
		}

		*v = *resp.CacheParameterGroups[0]

		return nil
	}
}

func testAccAWSElasticacheParameterGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "bar" {
	name = "%s"
	family = "redis2.8"
	parameter {
	  name = "appendonly"
	  value = "yes"
	}
}`, rName)
}

func testAccAWSElasticacheParameterGroupAddParametersConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "bar" {
	name = "%s"
	family = "redis2.8"
	description = "Test parameter group for terraform"
	parameter {
	  name = "appendonly"
	  value = "yes"
	}
	parameter {
	  name = "appendfsync"
	  value = "always"
	}
}`, rName)
}

func testAccAWSElasticacheParameterGroupOnlyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "bar" {
	name = "%s"
	family = "redis2.8"
	description = "Test parameter group for terraform"
}`, rName)
}

func testAccAWSElasticacheParameterGroupConfig_UppercaseName(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "bar" {
  name = "%s"
  family = "redis2.8"
  parameter {
    name = "appendonly"
    value = "yes"
  }
}`, rName)
}
