package sagemaker_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
)

func testAccApp_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var app sagemaker.DescribeAppOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "app_name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "domain_id", "aws_sagemaker_domain.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "user_profile_name", "aws_sagemaker_user_profile.test", "user_profile_name"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "sagemaker", regexp.MustCompile(`app/.+`)),
					resource.TestCheckResourceAttr(resourceName, "resource_spec.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_spec.0.sagemaker_image_arn"),
					resource.TestCheckResourceAttr(resourceName, "resource_spec.0.instance_type", "system"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func testAccApp_resourceSpec(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var app sagemaker.DescribeAppOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppResourceSpecConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "app_name", rName),
					resource.TestCheckResourceAttr(resourceName, "resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_spec.0.instance_type", "system"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_spec.0.sagemaker_image_arn"),
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

func testAccApp_resourceSpecLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var app sagemaker.DescribeAppOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	uName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppResourceSpecLifecycleConfig(rName, uName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "app_name", rName),
					resource.TestCheckResourceAttr(resourceName, "resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_spec.0.instance_type", "system"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_spec.0.sagemaker_image_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_spec.0.lifecycle_config_arn", "aws_sagemaker_studio_lifecycle_config.test", "arn"),
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

func testAccApp_tags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var app sagemaker.DescribeAppOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
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
				Config: testAccAppTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAppTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccApp_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var app sagemaker.DescribeAppOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
					acctest.CheckResourceDisappears(acctest.Provider, tfsagemaker.ResourceApp(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAppDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_app" {
			continue
		}

		domainID := rs.Primary.Attributes["domain_id"]
		userProfileName := rs.Primary.Attributes["user_profile_name"]
		appType := rs.Primary.Attributes["app_type"]
		appName := rs.Primary.Attributes["app_name"]

		app, err := tfsagemaker.FindAppByName(conn, domainID, userProfileName, appType, appName)

		if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading SageMaker App (%s): %w", rs.Primary.ID, err)
		}

		appArn := aws.StringValue(app.AppArn)
		if appArn == rs.Primary.ID && aws.StringValue(app.Status) != sagemaker.AppStatusDeleted {
			return fmt.Errorf("SageMaker App %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAppExists(n string, app *sagemaker.DescribeAppOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker domain ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn
		domainID := rs.Primary.Attributes["domain_id"]
		userProfileName := rs.Primary.Attributes["user_profile_name"]
		appType := rs.Primary.Attributes["app_type"]
		appName := rs.Primary.Attributes["app_name"]

		resp, err := tfsagemaker.FindAppByName(conn, domainID, userProfileName, appType, appName)
		if err != nil {
			return err
		}

		*app = *resp

		return nil
	}
}

func testAccAppBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  # SageMaker compute resources are not available at usw2-az4.
  exclude_zone_ids = ["usw2-az4"]
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

  default_user_settings {
    execution_role = aws_iam_role.test.arn
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}

resource "aws_sagemaker_user_profile" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = %[1]q
}
`, rName)
}

func testAccAppBasicConfig(rName string) string {
	return testAccAppBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_app" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = aws_sagemaker_user_profile.test.user_profile_name
  app_name          = %[1]q
  app_type          = "JupyterServer"
}
`, rName)
}

func testAccAppTags1Config(rName, tagKey1, tagValue1 string) string {
	return testAccAppBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_app" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = aws_sagemaker_user_profile.test.user_profile_name
  app_name          = %[1]q
  app_type          = "JupyterServer"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAppTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAppBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_app" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = aws_sagemaker_user_profile.test.user_profile_name
  app_name          = %[1]q
  app_type          = "JupyterServer"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAppResourceSpecConfig(rName string) string {
	return testAccAppBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_app" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = aws_sagemaker_user_profile.test.user_profile_name
  app_name          = %[1]q
  app_type          = "JupyterServer"

  resource_spec {
    instance_type = "system"
  }
}
`, rName)
}

func testAccAppResourceSpecLifecycleConfig(rName, uName string) string {
	return testAccAppBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_studio_lifecycle_config" "test" {
  studio_lifecycle_config_name     = %[1]q
  studio_lifecycle_config_app_type = "JupyterServer"
  studio_lifecycle_config_content  = base64encode("echo Hello")
}

resource "aws_sagemaker_user_profile" "lifecycletest" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = %[2]q

  user_settings {
    execution_role = aws_iam_role.test.arn

    jupyter_server_app_settings {
      default_resource_spec {
        instance_type        = "system"
        lifecycle_config_arn = aws_sagemaker_studio_lifecycle_config.test.arn
      }

      lifecycle_config_arns = [aws_sagemaker_studio_lifecycle_config.test.arn]
    }
  }
}

resource "aws_sagemaker_app" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = aws_sagemaker_user_profile.lifecycletest.user_profile_name
  app_name          = %[1]q
  app_type          = "JupyterServer"

  resource_spec {
    instance_type        = "system"
    lifecycle_config_arn = aws_sagemaker_studio_lifecycle_config.test.arn
  }
}
`, rName, uName)
}
