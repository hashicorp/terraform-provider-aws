package aws

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

// This is used for testing aws_codebuild_webhook as well as aws_codebuild_project.
// In order for that resource to work the Terraform AWS user must have done a GitHub
// OAuth dance. Additionally, the GitHub user that the Terraform AWS user logs in as
// must have access to the GitHub repository.
func testAccAWSCodeBuildGitHubSourceLocationFromEnv() string {
	sourceLocation := os.Getenv("AWS_CODEBUILD_GITHUB_SOURCE_LOCATION")
	if sourceLocation == "" {
		return "https://github.com/hashibot-test/aws-test.git"
	}
	return sourceLocation
}

func TestAccAWSCodeBuildProject_basic(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.1178773975.encryption_disabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.1178773975.location", ""),
					resource.TestCheckResourceAttr(resourceName, "artifacts.1178773975.name", ""),
					resource.TestCheckResourceAttr(resourceName, "artifacts.1178773975.namespace_type", ""),
					resource.TestCheckResourceAttr(resourceName, "artifacts.1178773975.packaging", ""),
					resource.TestCheckResourceAttr(resourceName, "artifacts.1178773975.path", ""),
					resource.TestCheckResourceAttr(resourceName, "artifacts.1178773975.type", "NO_ARTIFACTS"),
					resource.TestCheckResourceAttr(resourceName, "badge_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "build_timeout", "60"),
					resource.TestCheckResourceAttr(resourceName, "cache.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", "NO_CACHE"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestMatchResourceAttr(resourceName, "encryption_key", regexp.MustCompile(`^arn:[^:]+:kms:[^:]+:[^:]+:alias/aws/s3$`)),
					resource.TestCheckResourceAttr(resourceName, "environment.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "environment.1974383098.compute_type", "BUILD_GENERAL1_SMALL"),
					resource.TestCheckResourceAttr(resourceName, "environment.1974383098.environment_variable.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "environment.1974383098.image", "2"),
					resource.TestCheckResourceAttr(resourceName, "environment.1974383098.privileged_mode", "false"),
					resource.TestCheckResourceAttr(resourceName, "environment.1974383098.type", "LINUX_CONTAINER"),
					resource.TestMatchResourceAttr(resourceName, "service_role", regexp.MustCompile(`^arn:[^:]+:iam::[^:]+:role/tf-acc-test-[0-9]+$`)),
					resource.TestCheckResourceAttr(resourceName, "source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.1441597390.auth.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source.1441597390.git_clone_depth", "0"),
					resource.TestCheckResourceAttr(resourceName, "source.1441597390.insecure_ssl", "false"),
					resource.TestCheckResourceAttr(resourceName, "source.1441597390.location", "https://github.com/hashibot-test/aws-test.git"),
					resource.TestCheckResourceAttr(resourceName, "source.1441597390.report_build_status", "false"),
					resource.TestCheckResourceAttr(resourceName, "source.1441597390.type", "GITHUB"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_BadgeEnabled(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodebuildProjectConfig_BadgeEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "badge_enabled", "true"),
					resource.TestMatchResourceAttr(resourceName, "badge_url", regexp.MustCompile(`\b(https?).*\b`)),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_BuildTimeout(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_BuildTimeout(rName, 120),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "build_timeout", "120"),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_BuildTimeout(rName, 240),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "build_timeout", "240"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Cache(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSCodeBuildProjectConfig_Cache(rName, "", "S3"),
				ExpectError: regexp.MustCompile(`cache location is required when cache type is "S3"`),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_Cache(rName, "", "NO_CACHE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "cache.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", "NO_CACHE"),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "cache.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", "NO_CACHE"),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_Cache(rName, "some-bucket", "S3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "cache.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.location", "some-bucket"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", "S3"),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_Cache(rName, "some-new-bucket", "S3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "cache.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.location", "some-new-bucket"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", "S3"),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "cache.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", "NO_CACHE"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Description(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_Description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_EncryptionKey(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_EncryptionKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestMatchResourceAttr(resourceName, "encryption_key", regexp.MustCompile(`.+`)),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Environment_EnvironmentVariable_Type(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Environment_EnvironmentVariable_Type(rName, "PLAINTEXT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "environment.3925601246.environment_variable.0.type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "environment.3925601246.environment_variable.1.type", "PLAINTEXT"),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_Environment_EnvironmentVariable_Type(rName, "PARAMETER_STORE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "environment.2414912039.environment_variable.0.type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "environment.2414912039.environment_variable.1.type", "PARAMETER_STORE"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Source_Auth(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSCodeBuildProjectConfig_Source_Auth(rName, "FAKERESOURCE1", "INVALID"),
				ExpectError: regexp.MustCompile(`expected source.0.auth.0.type to be one of`),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_Auth(rName, "FAKERESOURCE1", "OAUTH"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.3680505372.auth.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.3680505372.auth.2706882902.resource", "FAKERESOURCE1"),
					resource.TestCheckResourceAttr(resourceName, "source.3680505372.auth.2706882902.type", "OAUTH"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Source_GitCloneDepth(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_GitCloneDepth(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.1181740906.git_clone_depth", "1"),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_GitCloneDepth(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.974047921.git_clone_depth", "2"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Source_InsecureSSL(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_InsecureSSL(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.1976396802.insecure_ssl", "true"),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_InsecureSSL(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.3680505372.insecure_ssl", "false"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Source_ReportBuildStatus(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_ReportBuildStatus(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.4215890488.report_build_status", "true"),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_ReportBuildStatus(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.3680505372.report_build_status", "false"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Source_Type_Bitbucket(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_Type_Bitbucket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.2806293607.type", "BITBUCKET"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Source_Type_CodeCommit(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_Type_CodeCommit(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.3715340088.type", "CODECOMMIT"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Source_Type_CodePipeline(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_Type_CodePipeline(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.2280907000.type", "CODEPIPELINE"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Source_Type_GitHubEnterprise(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_Type_GitHubEnterprise(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.553628638.type", "GITHUB_ENTERPRISE"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Source_Type_S3(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_Type_S3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.2751363124.type", "S3"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Tags(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Tags(rName, "tag2", "tag2value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1value"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "tag2value"),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_Tags(rName, "tag2", "tag2value-updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1value"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "tag2value-updated"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_VpcConfig(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_VpcConfig(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnets.#", "2"),
					resource.TestMatchResourceAttr(resourceName, "vpc_config.0.vpc_id", regexp.MustCompile(`^vpc-`)),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_VpcConfig(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnets.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "vpc_config.0.vpc_id", regexp.MustCompile(`^vpc-`)),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_WindowsContainer(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_WindowsContainer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "environment.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "environment.3935046469.compute_type", "BUILD_GENERAL1_MEDIUM"),
					resource.TestCheckResourceAttr(resourceName, "environment.3935046469.environment_variable.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "environment.3935046469.image", "2"),
					resource.TestCheckResourceAttr(resourceName, "environment.3935046469.privileged_mode", "false"),
					resource.TestCheckResourceAttr(resourceName, "environment.3935046469.type", "WINDOWS_CONTAINER"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Artifacts_EncryptionDisabled(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	bName := acctest.RandomWithPrefix("tf-acc-test-bucket")
	resourceName := "aws_codebuild_project.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodebuildProjectConfig_Artifacts_EncryptionDisabled(rName, bName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.537882814.encryption_disabled", "true"),
				),
			},
		},
	})
}

func TestAWSCodeBuildProject_nameValidation(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{Value: "_test", ErrCount: 1},
		{Value: "test", ErrCount: 0},
		{Value: "1_test", ErrCount: 0},
		{Value: "test**1", ErrCount: 1},
		{Value: acctest.RandString(256), ErrCount: 1},
	}

	for _, tc := range cases {
		_, errors := validateAwsCodeBuildProjectName(tc.Value, "aws_codebuild_project")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the AWS CodeBuild project name to trigger a validation error - %s", errors)
		}
	}
}

func testAccCheckAWSCodeBuildProjectExists(n string, project *codebuild.Project) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CodeBuild Project ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).codebuildconn

		out, err := conn.BatchGetProjects(&codebuild.BatchGetProjectsInput{
			Names: []*string{
				aws.String(rs.Primary.ID),
			},
		})

		if err != nil {
			return err
		}

		if len(out.Projects) < 1 {
			return fmt.Errorf("No project found")
		}

		*project = *out.Projects[0]

		return nil
	}
}

func testAccCheckAWSCodeBuildProjectDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).codebuildconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codebuild_project" {
			continue
		}

		out, err := conn.BatchGetProjects(&codebuild.BatchGetProjectsInput{
			Names: []*string{
				aws.String(rs.Primary.ID),
			},
		})

		if err != nil {
			return err
		}

		if out != nil && len(out.Projects) > 0 {
			return fmt.Errorf("Expected AWS CodeBuild Project to be gone, but was still found")
		}

		return nil
	}

	return nil
}

func testAccAWSCodeBuildProjectConfig_Base_Bucket(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = "%s"
  force_destroy = true
}`, rName)
}

func testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = "%s"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "codebuild.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  role   = "${aws_iam_role.test.name}"
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Resource": [
        "*"
      ],
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "ec2:CreateNetworkInterface",
        "ec2:DescribeDhcpOptions",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DeleteNetworkInterface",
        "ec2:DescribeSubnets",
        "ec2:DescribeSecurityGroups",
        "ec2:DescribeVpcs"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:*"
      ],
      "Resource": "*"
    }
  ]
}
POLICY
}`, rName)
}

func testAccAWSCodeBuildProjectConfig_basic(rName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = "%s"
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = "%s"
    type     = "GITHUB"
  }
}
`, rName, testAccAWSCodeBuildGitHubSourceLocationFromEnv())
}

func testAccAWSCodebuildProjectConfig_BadgeEnabled(rName string, badgeEnabled bool) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  badge_enabled = %t
  name          = "%s"
  service_role  = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, badgeEnabled, rName)
}

func testAccAWSCodeBuildProjectConfig_BuildTimeout(rName string, buildTimeout int) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  build_timeout = %d
  name          = "%s"
  service_role  = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, buildTimeout, rName)
}

func testAccAWSCodeBuildProjectConfig_Cache(rName, cacheLocation, cacheType string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = "%s"
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  cache {
    location = "%s"
    type     = "%s"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, cacheLocation, cacheType)
}

func testAccAWSCodeBuildProjectConfig_Description(rName, description string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  description  = "%s"
  name         = "%s"
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, description, rName)
}

func testAccAWSCodeBuildProjectConfig_EncryptionKey(rName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = "Terraform acc test"
  deletion_window_in_days = 7
}

resource "aws_codebuild_project" "test" {
  encryption_key = "${aws_kms_key.test.arn}"
  name           = "%s"
  service_role   = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName)
}

func testAccAWSCodeBuildProjectConfig_Environment_EnvironmentVariable_Type(rName, environmentVariableType string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = "%s"
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"

    environment_variable {
      name  = "SOME_KEY"
      value = "SOME_VALUE"
    }

    environment_variable {
      name  = "SOME_KEY2"
      value = "SOME_VALUE2"
      type  = "%s"
    }
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, environmentVariableType)
}

func testAccAWSCodeBuildProjectConfig_Source_Auth(rName, authResource, authType string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = "%s"
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type            = "GITHUB"
    location        = "https://github.com/hashicorp/packer.git"

    auth {
      resource = "%s"
      type     = "%s"
    }
  }
}
`, rName, authResource, authType)
}

func testAccAWSCodeBuildProjectConfig_Source_GitCloneDepth(rName string, gitCloneDepth int) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = "%s"
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    git_clone_depth = %d
    location        = "https://github.com/hashicorp/packer.git"
    type            = "GITHUB"
  }
}
`, rName, gitCloneDepth)
}

func testAccAWSCodeBuildProjectConfig_Source_InsecureSSL(rName string, insecureSSL bool) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = "%s"
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    insecure_ssl = %t
    location     = "https://github.com/hashicorp/packer.git"
    type         = "GITHUB"
  }
}
`, rName, insecureSSL)
}

func testAccAWSCodeBuildProjectConfig_Source_ReportBuildStatus(rName string, reportBuildStatus bool) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %q
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location            = "https://github.com/hashicorp/packer.git"
    report_build_status = %t
    type                = "GITHUB"
  }
}
`, rName, reportBuildStatus)
}

func testAccAWSCodeBuildProjectConfig_Source_Type_Bitbucket(rName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %q
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = "https://bitbucket.org/organization/repository.git"
    type     = "BITBUCKET"
  }
}
`, rName)
}

func testAccAWSCodeBuildProjectConfig_Source_Type_CodeCommit(rName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %q
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = "https://git-codecommit.region-id.amazonaws.com/v1/repos/repo-name"
    type     = "CODECOMMIT"
  }
}
`, rName)
}

func testAccAWSCodeBuildProjectConfig_Source_Type_CodePipeline(rName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %q
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "CODEPIPELINE"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type = "CODEPIPELINE"
  }
}
`, rName)
}

func testAccAWSCodeBuildProjectConfig_Source_Type_GitHubEnterprise(rName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %q
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = "https://example.com/organization/repository.git"
    type     = "GITHUB_ENTERPRISE"
  }
}
`, rName)
}

func testAccAWSCodeBuildProjectConfig_Source_Type_S3(rName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %q
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = "bucket-name/object-name.zip"
    type     = "S3"
  }
}
`, rName)
}

func testAccAWSCodeBuildProjectConfig_Tags(rName, tagKey, tagValue string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = "%s"
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = "https://github.com/hashicorp/packer.git"
    type     = "GITHUB"
  }

  tags {
    tag1 = "tag1value"
    %s = "%s"
  }
}
`, rName, tagKey, tagValue)
}

func testAccAWSCodeBuildProjectConfig_VpcConfig(rName string, subnetCount int) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count = %d

  cidr_block = "10.0.${count.index}.0/24"
  vpc_id     = "${aws_vpc.test.id}"

  tags {
    Name = "tf-acc-codebuild-project"
  }
}

resource "aws_security_group" "test" {
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_codebuild_project" "test" {
  name         = "%s"
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = "https://github.com/hashicorp/packer.git"
    type     = "GITHUB"
  }

  vpc_config {
    security_group_ids = ["${aws_security_group.test.id}"]
    subnets            = ["${aws_subnet.test.*.id}"]
    vpc_id             = "${aws_vpc.test.id}"
  }
}
`, subnetCount, rName)
}

func testAccAWSCodeBuildProjectConfig_WindowsContainer(rName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = "%s"
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_MEDIUM"
    image        = "2"
    type         = "WINDOWS_CONTAINER"
  }

  source {
    location = "%s"
    type     = "GITHUB"
  }
}
`, rName, testAccAWSCodeBuildGitHubSourceLocationFromEnv())
}

func testAccAWSCodebuildProjectConfig_Artifacts_EncryptionDisabled(rName string, bName string, encryptionDisabled bool) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + testAccAWSCodeBuildProjectConfig_Base_Bucket(bName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name          = "%s"
  service_role  = "${aws_iam_role.test.arn}"

  artifacts {
    encryption_disabled = %t
    location            = "${aws_s3_bucket.test.bucket}"
    type                = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, encryptionDisabled)
}
