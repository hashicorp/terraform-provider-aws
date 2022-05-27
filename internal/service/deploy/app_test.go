package deploy_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codedeploy"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodedeploy "github.com/hashicorp/terraform-provider-aws/internal/service/deploy"
)

func TestAccDeployApp_basic(t *testing.T) {
	var application1 codedeploy.ApplicationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codedeploy.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application1),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "codedeploy", fmt.Sprintf(`application:%s`, rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Server"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "linked_to_github", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "application_id"),
				),
			},
			// Import by ID
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Import by name
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployApp_computePlatform(t *testing.T) {
	var application1, application2 codedeploy.ApplicationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codedeploy.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_computePlatform(rName, "Lambda"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application1),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Lambda"),
				),
			},
			{
				Config: testAccAppConfig_computePlatform(rName, "Server"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application2),
					testAccCheckAppRecreated(&application1, &application2),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Server"),
				),
			},
		},
	})
}

func TestAccDeployApp_ComputePlatform_ecs(t *testing.T) {
	var application1 codedeploy.ApplicationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codedeploy.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_computePlatform(rName, "ECS"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application1),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "ECS"),
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

func TestAccDeployApp_ComputePlatform_lambda(t *testing.T) {
	var application1 codedeploy.ApplicationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codedeploy.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_computePlatform(rName, "Lambda"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application1),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Lambda"),
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

func TestAccDeployApp_name(t *testing.T) {
	var application1, application2 codedeploy.ApplicationInfo
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codedeploy.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_name(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application1),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
				),
			},
			{
				Config: testAccAppConfig_name(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application2),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
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

func TestAccDeployApp_tags(t *testing.T) {
	var application codedeploy.ApplicationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codedeploy.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application),
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
				Config: testAccAppConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAppConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccDeployApp_disappears(t *testing.T) {
	var application1 codedeploy.ApplicationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codedeploy.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application1),
					acctest.CheckResourceDisappears(acctest.Provider, tfcodedeploy.ResourceApp(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAppDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DeployConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codedeploy_app" {
			continue
		}

		_, err := conn.GetApplication(&codedeploy.GetApplicationInput{
			ApplicationName: aws.String(rs.Primary.Attributes["name"]),
		})

		if tfawserr.ErrCodeEquals(err, codedeploy.ErrCodeApplicationDoesNotExistException) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("still exists")
	}

	return nil
}

func testAccCheckAppExists(name string, application *codedeploy.ApplicationInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DeployConn

		input := &codedeploy.GetApplicationInput{
			ApplicationName: aws.String(rs.Primary.Attributes["name"]),
		}

		output, err := conn.GetApplication(input)

		if err != nil {
			return err
		}

		if output == nil || output.Application == nil {
			return fmt.Errorf("error reading CodeDeploy Application (%s): empty response", rs.Primary.ID)
		}

		*application = *output.Application

		return nil
	}
}

func testAccCheckAppRecreated(i, j *codedeploy.ApplicationInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreateTime).Equal(aws.TimeValue(j.CreateTime)) {
			return errors.New("CodeDeploy Application was not recreated")
		}

		return nil
	}
}

func testAccAppConfig_computePlatform(rName string, computePlatform string) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_app" "test" {
  compute_platform = %q
  name             = %q
}
`, computePlatform, rName)
}

func testAccAppConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_app" "test" {
  name = %q
}
`, rName)
}

func testAccAppConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_app" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAppConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_app" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
