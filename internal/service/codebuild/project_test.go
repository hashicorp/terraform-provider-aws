package codebuild_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodebuild "github.com/hashicorp/terraform-provider-aws/internal/service/codebuild"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(codebuild.EndpointsID, testAccErrorCheckSkipCodebuild)
}

func testAccErrorCheckSkipCodebuild(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"InvalidInputException: Region",
	)
}

// This is used for testing aws_codebuild_webhook as well as aws_codebuild_project.
// The Terraform AWS user must have done the manual Bitbucket OAuth dance for this
// functionality to work. Additionally, the Bitbucket user that the Terraform AWS
// user logs in as must have access to the Bitbucket repository.
func testAccBitbucketSourceLocationFromEnv() string {
	sourceLocation := os.Getenv("AWS_CODEBUILD_BITBUCKET_SOURCE_LOCATION")
	if sourceLocation == "" {
		return "https://terraform@bitbucket.org/terraform/aws-test.git" // nosemgrep: email-address
	}
	return sourceLocation
}

// This is used for testing aws_codebuild_webhook as well as aws_codebuild_project.
// The Terraform AWS user must have done the manual GitHub OAuth dance for this
// functionality to work. Additionally, the GitHub user that the Terraform AWS
// user logs in as must have access to the GitHub repository.
func testAccGitHubSourceLocationFromEnv() string {
	sourceLocation := os.Getenv("AWS_CODEBUILD_GITHUB_SOURCE_LOCATION")
	if sourceLocation == "" {
		return "https://github.com/hashibot-test/aws-test.git"
	}
	return sourceLocation
}

func TestAccCodeBuildProject_basic(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_codebuild_project.test"
	roleResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "codebuild", fmt.Sprintf("project/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "badge_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "build_timeout", "60"),
					resource.TestCheckResourceAttr(resourceName, "queued_timeout", "480"),
					resource.TestCheckResourceAttr(resourceName, "cache.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", codebuild.CacheTypeNoCache),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					acctest.CheckResourceAttrRegionalARN(resourceName, "encryption_key", "kms", "alias/aws/s3"),
					resource.TestCheckResourceAttr(resourceName, "environment.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "environment.0.compute_type", codebuild.ComputeTypeBuildGeneral1Small),
					resource.TestCheckResourceAttr(resourceName, "environment.0.environment_variable.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "environment.0.image", "2"),
					resource.TestCheckResourceAttr(resourceName, "environment.0.privileged_mode", "false"),
					resource.TestCheckResourceAttr(resourceName, "environment.0.type", codebuild.EnvironmentTypeLinuxContainer),
					resource.TestCheckResourceAttr(resourceName, "environment.0.image_pull_credentials_type", codebuild.ImagePullCredentialsTypeCodebuild),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.cloudwatch_logs.0.status", codebuild.LogsConfigStatusTypeEnabled),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.s3_logs.0.status", codebuild.LogsConfigStatusTypeDisabled),
					resource.TestCheckResourceAttr(resourceName, "project_visibility", "PRIVATE"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", roleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.auth.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source.0.git_clone_depth", "0"),
					resource.TestCheckResourceAttr(resourceName, "source.0.insecure_ssl", "false"),
					resource.TestCheckResourceAttr(resourceName, "source.0.location", "https://github.com/hashibot-test/aws-test.git"),
					resource.TestCheckResourceAttr(resourceName, "source.0.report_build_status", "false"),
					resource.TestCheckResourceAttr(resourceName, "source.0.type", "GITHUB"),
					resource.TestCheckResourceAttr(resourceName, "secondary_source_version.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "secondary_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "0"),
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

func TestAccCodeBuildProject_publicVisibility(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_codebuild_project.test"
	roleResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_projectVisibility(rName, "PUBLIC_READ"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "project_visibility", "PUBLIC_READ"),
					resource.TestCheckResourceAttrSet(resourceName, "public_project_alias"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_projectVisibility(rName, "PRIVATE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "project_visibility", "PRIVATE"),
				),
			},
			{
				Config: testAccProjectConfig_projectVisibility_resourceRole(rName, "PRIVATE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "project_visibility", "PRIVATE"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_access_role", roleResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_badgeEnabled(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_BadgeEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "badge_enabled", "true"),
					resource.TestMatchResourceAttr(resourceName, "badge_url", regexp.MustCompile(`\b(https?).*\b`)),
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

func TestAccCodeBuildProject_buildTimeout(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_BuildTimeout(rName, 120),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "build_timeout", "120"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_BuildTimeout(rName, 240),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "build_timeout", "240"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_queuedTimeout(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_QueuedTimeout(rName, 120),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "queued_timeout", "120"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_QueuedTimeout(rName, 240),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "queued_timeout", "240"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_cache(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"
	s3Location1 := rName + "-1"
	s3Location2 := rName + "-2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccProjectConfig_Cache(rName, "", "S3"),
				ExpectError: regexp.MustCompile(`cache location is required when cache type is "S3"`),
			},
			{
				Config: testAccProjectConfig_Cache(rName, "", codebuild.CacheTypeNoCache),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "cache.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", codebuild.CacheTypeNoCache),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "cache.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", codebuild.CacheTypeNoCache),
				),
			},
			{
				Config: testAccProjectConfig_Cache(rName, s3Location1, "S3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "cache.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.location", s3Location1),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", "S3"),
				),
			},
			{
				Config: testAccProjectConfig_Cache(rName, s3Location2, "S3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "cache.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.location", s3Location2),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", "S3"),
				),
			},
			{
				Config: testAccProjectConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "cache.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", codebuild.CacheTypeNoCache),
				),
			},
			{
				Config: testAccProjectConfig_LocalCache(rName, "LOCAL_DOCKER_LAYER_CACHE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "cache.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.modes.0", "LOCAL_DOCKER_LAYER_CACHE"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", "LOCAL"),
				),
			},
			{
				Config: testAccProjectConfig_S3_ComputedLocation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "cache.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", codebuild.CacheTypeS3),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_description(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_Description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_fileSystemLocations(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID, "efs"), //using efs.EndpointsID will import efs and make linters sad
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_FileSystemLocations(rName, "/mount1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "environment.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "environment.0.compute_type", codebuild.ComputeTypeBuildGeneral1Small),
					resource.TestCheckResourceAttr(resourceName, "environment.0.environment_variable.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "environment.0.image", "2"),
					resource.TestCheckResourceAttr(resourceName, "environment.0.privileged_mode", "true"),
					resource.TestCheckResourceAttr(resourceName, "environment.0.type", codebuild.EnvironmentTypeLinuxContainer),
					resource.TestCheckResourceAttr(resourceName, "file_system_locations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "file_system_locations.0.identifier", "test"),
					resource.TestMatchResourceAttr(resourceName, "file_system_locations.0.location", regexp.MustCompile(`/directory-path$`)),
					resource.TestCheckResourceAttr(resourceName, "file_system_locations.0.mount_options", "nfsvers=4.1,rsize=1048576,wsize=1048576,hard,timeo=450,retrans=3"),
					resource.TestCheckResourceAttr(resourceName, "file_system_locations.0.mount_point", "/mount1"),
					resource.TestCheckResourceAttr(resourceName, "file_system_locations.0.type", codebuild.FileSystemTypeEfs),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_FileSystemLocations(rName, "/mount2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "file_system_locations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "file_system_locations.0.identifier", "test"),
					resource.TestMatchResourceAttr(resourceName, "file_system_locations.0.location", regexp.MustCompile(`/directory-path$`)),
					resource.TestCheckResourceAttr(resourceName, "file_system_locations.0.mount_options", "nfsvers=4.1,rsize=1048576,wsize=1048576,hard,timeo=450,retrans=3"),
					resource.TestCheckResourceAttr(resourceName, "file_system_locations.0.mount_point", "/mount2"),
					resource.TestCheckResourceAttr(resourceName, "file_system_locations.0.type", codebuild.FileSystemTypeEfs),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_sourceVersion(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_SourceVersion(rName, "master"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source_version", "master"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_encryptionKey(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_EncryptionKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_key", "aws_kms_key.test", "arn"),
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

func TestAccCodeBuildProject_Environment_environmentVariable(t *testing.T) {
	var project1, project2, project3 codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Environment_EnvironmentVariable_One(rName, "KEY1", "VALUE1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_Environment_EnvironmentVariable_Two(rName, "KEY1", "VALUE1UPDATED", "KEY2", "VALUE2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_Environment_EnvironmentVariable_Zero(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project3),
					resource.TestCheckResourceAttr(resourceName, "environment.0.environment_variable.#", "0"),
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

func TestAccCodeBuildProject_EnvironmentEnvironmentVariable_type(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Environment_EnvironmentVariable_Type(rName, codebuild.EnvironmentVariableTypePlaintext),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "environment.0.environment_variable.0.type", codebuild.EnvironmentVariableTypePlaintext),
					resource.TestCheckResourceAttr(resourceName, "environment.0.environment_variable.1.type", codebuild.EnvironmentVariableTypePlaintext),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_Environment_EnvironmentVariable_Type(rName, codebuild.EnvironmentVariableTypeParameterStore),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "environment.0.environment_variable.0.type", codebuild.EnvironmentVariableTypePlaintext),
					resource.TestCheckResourceAttr(resourceName, "environment.0.environment_variable.1.type", codebuild.EnvironmentVariableTypeParameterStore),
				),
			},
			{
				Config: testAccProjectConfig_Environment_EnvironmentVariable_Type(rName, codebuild.EnvironmentVariableTypeSecretsManager),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "environment.0.environment_variable.0.type", codebuild.EnvironmentVariableTypePlaintext),
					resource.TestCheckResourceAttr(resourceName, "environment.0.environment_variable.1.type", codebuild.EnvironmentVariableTypeSecretsManager),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_EnvironmentEnvironmentVariable_value(t *testing.T) {
	var project1, project2, project3 codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Environment_EnvironmentVariable_One(rName, "KEY1", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_Environment_EnvironmentVariable_One(rName, "KEY1", "VALUE1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_Environment_EnvironmentVariable_One(rName, "KEY1", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project3),
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

func TestAccCodeBuildProject_Environment_certificate(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	oName := "certificate.pem"
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Environment_Certificate(rName, oName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					testAccCheckProjectCertificate(&project, fmt.Sprintf("%s/%s", rName, oName)),
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

func TestAccCodeBuildProject_Logs_cloudWatchLogs(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_LogsConfig_CloudWatchLogs(rName, codebuild.LogsConfigStatusTypeEnabled, "group-name", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.cloudwatch_logs.0.status", codebuild.LogsConfigStatusTypeEnabled),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.cloudwatch_logs.0.group_name", "group-name"),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.cloudwatch_logs.0.stream_name", ""),
				),
			},
			{
				Config: testAccProjectConfig_LogsConfig_CloudWatchLogs(rName, codebuild.LogsConfigStatusTypeEnabled, "group-name", "stream-name"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.cloudwatch_logs.0.status", codebuild.LogsConfigStatusTypeEnabled),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.cloudwatch_logs.0.group_name", "group-name"),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.cloudwatch_logs.0.stream_name", "stream-name"),
				),
			},
			{
				Config: testAccProjectConfig_LogsConfig_CloudWatchLogs(rName, codebuild.LogsConfigStatusTypeDisabled, "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.cloudwatch_logs.0.status", codebuild.LogsConfigStatusTypeDisabled),
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

func TestAccCodeBuildProject_Logs_s3Logs(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_LogsConfig_S3Logs(rName, codebuild.LogsConfigStatusTypeEnabled, rName+"/build-log", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.s3_logs.0.status", codebuild.LogsConfigStatusTypeEnabled),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.s3_logs.0.location", rName+"/build-log"),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.s3_logs.0.encryption_disabled", "false"),
				),
			},
			{
				Config: testAccProjectConfig_LogsConfig_S3Logs(rName, codebuild.LogsConfigStatusTypeEnabled, rName+"/build-log", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.s3_logs.0.status", codebuild.LogsConfigStatusTypeEnabled),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.s3_logs.0.location", rName+"/build-log"),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.s3_logs.0.encryption_disabled", "true"),
				),
			},
			{
				Config: testAccProjectConfig_LogsConfig_S3Logs(rName, codebuild.LogsConfigStatusTypeDisabled, "", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.s3_logs.0.status", codebuild.LogsConfigStatusTypeDisabled),
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

func TestAccCodeBuildProject_buildBatch(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	if acctest.Partition() == "aws-us-gov" {
		t.Skip("CodeBuild Project build batch config is not supported in GovCloud partition")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_BuildBatchConfig(rName, true, "BUILD_GENERAL1_SMALL", 10, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.combine_artifacts", "true"),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.restrictions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.restrictions.0.compute_types_allowed.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.restrictions.0.compute_types_allowed.0", "BUILD_GENERAL1_SMALL"),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.restrictions.0.maximum_builds_allowed", "10"),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.timeout_in_mins", "5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_BuildBatchConfig(rName, false, "BUILD_GENERAL1_MEDIUM", 20, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.combine_artifacts", "false"),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.restrictions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.restrictions.0.compute_types_allowed.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.restrictions.0.compute_types_allowed.0", "BUILD_GENERAL1_MEDIUM"),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.restrictions.0.maximum_builds_allowed", "20"),
					resource.TestCheckResourceAttr(resourceName, "build_batch_config.0.timeout_in_mins", "10"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_Source_gitCloneDepth(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Source_GitCloneDepth(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.git_clone_depth", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_Source_GitCloneDepth(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.git_clone_depth", "2"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SourceGitSubmodules_codeCommit(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Source_GitSubmodulesConfig_CodeCommit(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.git_submodules_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.git_submodules_config.0.fetch_submodules", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_Source_GitSubmodulesConfig_CodeCommit(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.git_submodules_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.git_submodules_config.0.fetch_submodules", "false"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SourceGitSubmodules_gitHub(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Source_GitSubmodulesConfig_GitHub(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_Source_GitSubmodulesConfig_GitHub(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SourceGitSubmodules_gitHubEnterprise(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Source_GitSubmodulesConfig_GitHubEnterprise(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_Source_GitSubmodulesConfig_GitHubEnterprise(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SecondarySourcesGitSubmodules_codeCommit(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_SecondarySources_GitSubmodulesConfig_CodeCommit(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_sources.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier":                        "secondarySource1",
						"git_submodules_config.#":                  "1",
						"git_submodules_config.0.fetch_submodules": "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier":                        "secondarySource2",
						"git_submodules_config.#":                  "1",
						"git_submodules_config.0.fetch_submodules": "true",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_SecondarySources_GitSubmodulesConfig_CodeCommit(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_sources.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier":                        "secondarySource1",
						"git_submodules_config.#":                  "1",
						"git_submodules_config.0.fetch_submodules": "false",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier":                        "secondarySource2",
						"git_submodules_config.#":                  "1",
						"git_submodules_config.0.fetch_submodules": "false",
					}),
				),
			},
			{
				Config: testAccProjectConfig_SecondarySources_none(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_sources.#", "0"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SecondarySourcesGitSubmodules_gitHub(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_SecondarySources_GitSubmodulesConfig_GitHub(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_SecondarySources_GitSubmodulesConfig_GitHub(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SecondarySourcesGitSubmodules_gitHubEnterprise(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_SecondarySources_GitSubmodulesConfig_GitHubEnterprise(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_SecondarySources_GitSubmodulesConfig_GitHubEnterprise(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SecondarySourcesVersions(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_SecondarySourceVersions_CodeCommit(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_sources.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "secondary_source_version.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier": "secondarySource1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier": "secondarySource2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_source_version.*", map[string]string{
						"source_identifier": "secondarySource1",
						"source_version":    "master",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_SecondarySourceVersions_CodeCommitUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_sources.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "secondary_source_version.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier": "secondarySource1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier": "secondarySource2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_source_version.*", map[string]string{
						"source_identifier": "secondarySource1",
						"source_version":    "master",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_source_version.*", map[string]string{
						"source_identifier": "secondarySource2",
						"source_version":    "master",
					}),
				),
			},
			{
				Config: testAccProjectConfig_SecondarySourceVersions_CodeCommit(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_sources.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "secondary_source_version.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier": "secondarySource1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier": "secondarySource2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_source_version.*", map[string]string{
						"source_identifier": "secondarySource1",
						"source_version":    "master",
					}),
				),
			},
			{
				Config: testAccProjectConfig_SecondarySources_GitSubmodulesConfig_CodeCommit(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_sources.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "secondary_source_version.#", "0"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier": "secondarySource1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier": "secondarySource2",
					}),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SourceBuildStatus_gitHubEnterprise(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	if acctest.Partition() == "aws-us-gov" {
		t.Skip("CodeBuild Project build status config is not supported in GovCloud partition")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Source_BuildStatusConfig_GitHubEnterprise(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
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

func TestAccCodeBuildProject_Source_insecureSSL(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Source_InsecureSSL(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.insecure_ssl", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_Source_InsecureSSL(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.insecure_ssl", "false"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SourceReportBuildStatus_bitbucket(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	sourceLocation := testAccBitbucketSourceLocationFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Source_ReportBuildStatus_Bitbucket(rName, sourceLocation, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.report_build_status", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_Source_ReportBuildStatus_Bitbucket(rName, sourceLocation, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.report_build_status", "false"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SourceReportBuildStatus_gitHub(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Source_ReportBuildStatus_GitHub(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.report_build_status", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_Source_ReportBuildStatus_GitHub(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.report_build_status", "false"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SourceReportBuildStatus_gitHubEnterprise(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Source_ReportBuildStatus_GitHubEnterprise(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.report_build_status", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_Source_ReportBuildStatus_GitHubEnterprise(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.report_build_status", "false"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SourceType_bitbucket(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	sourceLocation := testAccBitbucketSourceLocationFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Source_Type_Bitbucket(rName, sourceLocation),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.type", "BITBUCKET"),
					resource.TestCheckResourceAttr(resourceName, "source.0.location", sourceLocation),
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

func TestAccCodeBuildProject_SourceType_codeCommit(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Source_Type_CodeCommit(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.type", "CODECOMMIT"),
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

func TestAccCodeBuildProject_SourceType_codePipeline(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Source_Type_CodePipeline(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.type", "CODEPIPELINE"),
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

func TestAccCodeBuildProject_SourceType_gitHubEnterprise(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Source_Type_GitHubEnterprise(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.type", "GITHUB_ENTERPRISE"),
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

func TestAccCodeBuildProject_SourceType_s3(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Source_Type_S3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
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

func TestAccCodeBuildProject_SourceType_noSource(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"
	rBuildspec := `
version: 0.2
phases:
  build:
    commands:
      - rspec hello_world_spec.rb
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Source_Type_NoSource(rName, "", rBuildspec),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.type", "NO_SOURCE"),
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

func TestAccCodeBuildProject_SourceType_noSourceInvalid(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBuildspec := `
version: 0.2
phases:
  build:
    commands:
      - rspec hello_world_spec.rb
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccProjectConfig_Source_Type_NoSource(rName, "", ""),
				ExpectError: regexp.MustCompile("`buildspec` must be set when source's `type` is `NO_SOURCE`"),
			},
			{
				Config:      testAccProjectConfig_Source_Type_NoSource(rName, "location", rBuildspec),
				ExpectError: regexp.MustCompile("`location` must be empty when source's `type` is `NO_SOURCE`"),
			},
		},
	})
}

func TestAccCodeBuildProject_tags(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Tags(rName, "tag2", "tag2value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1value"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "tag2value"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_Tags(rName, "tag2", "tag2value-updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1value"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "tag2value-updated"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_vpc(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_VPCConfig2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnets.#", "2"),
					resource.TestMatchResourceAttr(resourceName, "vpc_config.0.vpc_id", regexp.MustCompile(`^vpc-`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_VPCConfig1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnets.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "vpc_config.0.vpc_id", regexp.MustCompile(`^vpc-`)),
				),
			},
			{
				Config: testAccProjectConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "0"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_windowsServer2019Container(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_WindowsServer2019Container(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "environment.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "environment.0.compute_type", codebuild.ComputeTypeBuildGeneral1Medium),
					resource.TestCheckResourceAttr(resourceName, "environment.0.environment_variable.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "environment.0.image", "2"),
					resource.TestCheckResourceAttr(resourceName, "environment.0.privileged_mode", "false"),
					resource.TestCheckResourceAttr(resourceName, "environment.0.image_pull_credentials_type", codebuild.ImagePullCredentialsTypeCodebuild),
					resource.TestCheckResourceAttr(resourceName, "environment.0.type", codebuild.EnvironmentTypeWindowsServer2019Container),
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

func TestAccCodeBuildProject_armContainer(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_ARMContainer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
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

func TestAccCodeBuildProject_Artifacts_artifactIdentifier(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	artifactIdentifier1 := "artifactIdentifier1"
	artifactIdentifier2 := "artifactIdentifier2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Artifacts_ArtifactIdentifier(rName, artifactIdentifier1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.artifact_identifier", artifactIdentifier1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_Artifacts_ArtifactIdentifier(rName, artifactIdentifier2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.artifact_identifier", artifactIdentifier2),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_Artifacts_encryptionDisabled(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Artifacts_EncryptionDisabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.encryption_disabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_Artifacts_EncryptionDisabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.encryption_disabled", "false"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_Artifacts_location(t *testing.T) {
	var project codebuild.Project
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Artifacts_Location(rName1, rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.location", rName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_Artifacts_Location(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.location", rName2),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_Artifacts_name(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	name1 := "name1"
	name2 := "name2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Artifacts_Name(rName, name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.name", name1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_Artifacts_Name(rName, name2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.name", name2),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_Artifacts_namespaceType(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Artifacts_NamespaceType(rName, codebuild.ArtifactNamespaceBuildId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.namespace_type", codebuild.ArtifactNamespaceBuildId),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_Artifacts_NamespaceType(rName, codebuild.ArtifactNamespaceNone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.namespace_type", codebuild.ArtifactNamespaceNone),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_Artifacts_overrideArtifactName(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Artifacts_OverrideArtifactName(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.override_artifact_name", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_Artifacts_OverrideArtifactName(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.override_artifact_name", "false"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_Artifacts_packaging(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Artifacts_Packaging(rName, codebuild.ArtifactPackagingZip),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.packaging", codebuild.ArtifactPackagingZip),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_Artifacts_Packaging(rName, codebuild.ArtifactPackagingNone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.packaging", codebuild.ArtifactPackagingNone),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_Artifacts_path(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Artifacts_Path(rName, "path1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.path", "path1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_Artifacts_Path(rName, "path2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.path", "path2"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_Artifacts_type(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	type1 := codebuild.ArtifactsTypeS3
	type2 := codebuild.ArtifactsTypeCodepipeline

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Artifacts_Type(rName, type1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.type", type1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_Artifacts_Type(rName, type2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.type", type2),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_Artifacts_bucketOwnerAccess(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Artifacts_BucketOwnerAccess(rName, "FULL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.bucket_owner_access", "FULL"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_Artifacts_BucketOwnerAccess(rName, "READ_ONLY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.bucket_owner_access", "READ_ONLY"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_secondaryArtifacts(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_SecondaryArtifacts(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_SecondaryArtifacts_none(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "0"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SecondaryArtifacts_artifactIdentifier(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	artifactIdentifier1 := "artifactIdentifier1"
	artifactIdentifier2 := "artifactIdentifier2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_SecondaryArtifacts_ArtifactIdentifier(rName, artifactIdentifier1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						"artifact_identifier": artifactIdentifier1,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_SecondaryArtifacts_ArtifactIdentifier(rName, artifactIdentifier2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						"artifact_identifier": artifactIdentifier2,
					}),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SecondaryArtifacts_overrideArtifactName(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_SecondaryArtifacts_OverrideArtifactName(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						"override_artifact_name": "true",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_SecondaryArtifacts_OverrideArtifactName(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						"override_artifact_name": "false",
					}),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SecondaryArtifacts_encryptionDisabled(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_SecondaryArtifacts_EncryptionDisabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						"encryption_disabled": "true",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_SecondaryArtifacts_EncryptionDisabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						"encryption_disabled": "false",
					}),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SecondaryArtifacts_location(t *testing.T) {
	var project codebuild.Project
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_SecondaryArtifacts_Location(rName1, rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						"location": rName1,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_SecondaryArtifacts_Location(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						"location": rName2,
					}),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SecondaryArtifacts_name(t *testing.T) {
	acctest.Skip(t, "Currently no solution to allow updates on name attribute")

	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	name1 := "name1"
	name2 := "name2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_SecondaryArtifacts_Name(rName, name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						"name": name1,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_SecondaryArtifacts_Name(rName, name2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						"name": name2,
					}),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SecondaryArtifacts_namespaceType(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_SecondaryArtifacts_NamespaceType(rName, codebuild.ArtifactNamespaceBuildId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						"namespace_type": codebuild.ArtifactNamespaceBuildId,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_SecondaryArtifacts_NamespaceType(rName, codebuild.ArtifactNamespaceNone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						"namespace_type": codebuild.ArtifactNamespaceNone,
					}),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SecondaryArtifacts_packaging(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_SecondaryArtifacts_Packaging(rName, codebuild.ArtifactPackagingZip),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						"packaging": codebuild.ArtifactPackagingZip,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_SecondaryArtifacts_Packaging(rName, codebuild.ArtifactPackagingNone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						"packaging": codebuild.ArtifactPackagingNone,
					}),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SecondaryArtifacts_path(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	path1 := "path1"
	path2 := "path2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_SecondaryArtifacts_Path(rName, path1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						"path": path1,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_SecondaryArtifacts_Path(rName, path2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						"path": path2,
					}),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_SecondaryArtifacts_type(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_SecondaryArtifacts_Type(rName, codebuild.ArtifactsTypeS3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_artifacts.*", map[string]string{
						"type": codebuild.ArtifactsTypeS3,
					}),
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

func TestAccCodeBuildProject_SecondarySources_codeCommit(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_SecondarySources_CodeCommit(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.0.type", "CODECOMMIT"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier": "secondarySource1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "secondary_sources.*", map[string]string{
						"source_identifier": "secondarySource2",
					}),
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

func TestProject_nameValidation(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{Value: "_test", ErrCount: 1},
		{Value: "test", ErrCount: 0},
		{Value: "1_test", ErrCount: 0},
		{Value: "test**1", ErrCount: 1},
		{Value: sdkacctest.RandString(256), ErrCount: 1},
	}

	for _, tc := range cases {
		_, errors := tfcodebuild.ValidProjectName(tc.Value, "aws_codebuild_project")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the AWS CodeBuild project name to trigger a validation error - %s", errors)
		}
	}
}

func TestAccCodeBuildProject_concurrentBuildLimit(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_ConcurrentBuildLimit(rName, 4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "concurrent_build_limit", "4"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_ConcurrentBuildLimit(rName, 12),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "concurrent_build_limit", "12"),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_Environment_registryCredential(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_Environment_RegistryCredential1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_Environment_RegistryCredential2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
				),
			},
		},
	})
}

func TestAccCodeBuildProject_disappears(t *testing.T) {
	var project codebuild.Project
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					acctest.CheckResourceDisappears(acctest.Provider, tfcodebuild.ResourceProject(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcodebuild.ResourceProject(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckProjectExists(n string, project *codebuild.Project) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CodeBuild Project ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeBuildConn

		output, err := tfcodebuild.FindProjectByARN(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("CodeBuild Project (%s) not found", rs.Primary.ID)
		}

		*project = *output

		return nil
	}
}

func testAccCheckProjectDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CodeBuildConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codebuild_project" {
			continue
		}

		_, err := tfcodebuild.FindProjectByARN(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("CodeBuild Project %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckProjectCertificate(project *codebuild.Project, expectedCertificate string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(project.Environment.Certificate) != expectedCertificate {
			return fmt.Errorf("CodeBuild Project certificate (%s) did not match: %s", aws.StringValue(project.Environment.Certificate), expectedCertificate)
		}
		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CodeBuildConn

	_, err := tfcodebuild.FindProjectByARN(conn, "tf-acc-test-precheck")

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil && !tfresource.NotFound(err) {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccProjectConfig_Base_ServiceRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "codebuild.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Resource": "*",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ]
    },
    {
      "Effect": "Allow",
      "Resource": "*",
      "Action": [
        "s3:PutObject",
        "s3:GetObject",
        "s3:GetObjectVersion",
        "s3:GetBucketAcl",
        "s3:PutBucketAcl",
        "s3:GetBucketLocation"
      ]
    },
    {
      "Effect": "Allow",
      "Resource": "*",
      "Action": [
        "codebuild:CreateReportGroup",
        "codebuild:CreateReport",
        "codebuild:UpdateReport",
        "codebuild:BatchPutTestCases",
        "codebuild:BatchPutCodeCoverages"
      ]
    },
    {
      "Effect": "Allow",
      "Resource": "*",
      "Action": [
        "ec2:CreateNetworkInterface",
        "ec2:CreateNetworkInterfacePermission",
        "ec2:DescribeDhcpOptions",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DeleteNetworkInterface",
        "ec2:DescribeSubnets",
        "ec2:DescribeSecurityGroups",
        "ec2:DescribeVpcs"
      ]
    }
  ]
}
POLICY
}
`, rName)
}

func testAccProjectConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = %[2]q
    type     = "GITHUB"
  }
}
`, rName, testAccGitHubSourceLocationFromEnv()))
}

func testAccProjectConfig_projectVisibility(rName, visibility string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name               = %[1]q
  service_role       = aws_iam_role.test.arn
  project_visibility = %[3]q

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = %[2]q
    type     = "GITHUB"
  }
}
`, rName, testAccGitHubSourceLocationFromEnv(), visibility))
}

func testAccProjectConfig_projectVisibility_resourceRole(rName, visibility string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name                 = %[1]q
  service_role         = aws_iam_role.test.arn
  project_visibility   = %[3]q
  resource_access_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = %[2]q
    type     = "GITHUB"
  }
}
`, rName, testAccGitHubSourceLocationFromEnv(), visibility))
}

func testAccProjectConfig_BadgeEnabled(rName string, badgeEnabled bool) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  badge_enabled = %[1]t
  name          = %[2]q
  service_role  = aws_iam_role.test.arn

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
`, badgeEnabled, rName))
}

func testAccProjectConfig_BuildTimeout(rName string, buildTimeout int) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  build_timeout = %[1]d
  name          = %[2]q
  service_role  = aws_iam_role.test.arn

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
`, buildTimeout, rName))
}

func testAccProjectConfig_QueuedTimeout(rName string, queuedTimeout int) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  queued_timeout = %[1]d
  name           = %[2]q
  service_role   = aws_iam_role.test.arn

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
`, queuedTimeout, rName))
}

func testAccProjectConfig_S3_ComputedLocation(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket_prefix = "cache"
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

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

  cache {
    type     = "S3"
    location = aws_s3_bucket.test.bucket
  }
}
`, rName))
}

func testAccProjectConfig_Cache(rName, cacheLocation, cacheType string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test1" {
  bucket        = "%[1]s-1"
  force_destroy = true
}

resource "aws_s3_bucket" "test2" {
  bucket        = "%[1]s-2"
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  cache {
    location = %[2]q
    type     = %[3]q
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

  depends_on = [aws_s3_bucket.test1, aws_s3_bucket.test2]
}
`, rName, cacheLocation, cacheType))
}

func testAccProjectConfig_LocalCache(rName, modeType string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  cache {
    type  = "LOCAL"
    modes = [%[2]q]
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
`, rName, modeType))
}

func testAccProjectConfig_Description(rName, description string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  description  = %[1]q
  name         = %[2]q
  service_role = aws_iam_role.test.arn

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
`, description, rName))
}

func testAccProjectConfig_SourceVersion(rName, sourceVersion string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source_version = %[2]q

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, sourceVersion))
}

func testAccProjectConfig_EncryptionKey(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codebuild_project" "test" {
  encryption_key = aws_kms_key.test.arn
  name           = %[1]q
  service_role   = aws_iam_role.test.arn

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
`, rName))
}

func testAccProjectConfig_Environment_EnvironmentVariable_One(rName, key1, value1 string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"

    environment_variable {
      name  = %[2]q
      value = %[3]q
    }
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, key1, value1))
}

func testAccProjectConfig_Environment_EnvironmentVariable_Two(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"

    environment_variable {
      name  = %[2]q
      value = %[3]q
    }

    environment_variable {
      name  = %[4]q
      value = %[5]q
    }
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, key1, value1, key2, value2))
}

func testAccProjectConfig_Environment_EnvironmentVariable_Zero(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

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
`, rName))
}

func testAccProjectConfig_Environment_EnvironmentVariable_Type(rName, environmentVariableType string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

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
      type  = %[2]q
    }
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, environmentVariableType))
}

func testAccProjectConfig_Environment_Certificate(rName string, oName string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_Base_ServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.bucket
  key     = %[1]q
  content = "test"
}

resource "aws_codebuild_project" "test" {
  name         = %[2]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
    certificate  = "${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, oName, rName))
}

func testAccProjectConfig_Environment_RegistryCredential1(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type                = "BUILD_GENERAL1_SMALL"
    image                       = "2"
    type                        = "LINUX_CONTAINER"
    image_pull_credentials_type = "SERVICE_ROLE"

    registry_credential {
      credential          = aws_secretsmanager_secret_version.test.arn
      credential_provider = "SECRETS_MANAGER"
    }
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}

resource "aws_secretsmanager_secret" "test" {
  name                    = "%[1]s-1"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = jsonencode({ username : "user", password : "pass" })
}
`, rName))
}

func testAccProjectConfig_Environment_RegistryCredential2(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type                = "BUILD_GENERAL1_SMALL"
    image                       = "2"
    type                        = "LINUX_CONTAINER"
    image_pull_credentials_type = "SERVICE_ROLE"

    registry_credential {
      credential          = aws_secretsmanager_secret_version.test.arn
      credential_provider = "SECRETS_MANAGER"
    }
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}

resource "aws_secretsmanager_secret" "test" {
  name                    = "%[1]s-2"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = jsonencode({ username : "user", password : "pass" })
}
`, rName))
}

func testAccProjectConfig_LogsConfig_CloudWatchLogs(rName, status, gName, sName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

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

  logs_config {
    cloudwatch_logs {
      status      = %[2]q
      group_name  = %[3]q
      stream_name = %[4]q
    }
  }
}
`, rName, status, gName, sName))
}

func testAccProjectConfig_BuildBatchConfig(rName string, combineArtifacts bool, computeTypesAllowed string, maximumBuildsAllowed, timeoutInMins int) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

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

  build_batch_config {
    combine_artifacts = %[2]t

    restrictions {
      compute_types_allowed  = [%[3]q]
      maximum_builds_allowed = %[4]d
    }

    service_role    = aws_iam_role.test.arn
    timeout_in_mins = %[5]d
  }
}
`, rName, combineArtifacts, computeTypesAllowed, maximumBuildsAllowed, timeoutInMins))
}

func testAccProjectConfig_LogsConfig_S3Logs(rName, status, location string, encryptionDisabled bool) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_Base_ServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

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

  logs_config {
    s3_logs {
      status              = %[2]q
      location            = %[3]q
      encryption_disabled = %[4]t
    }
  }
}
`, rName, status, location, encryptionDisabled))
}

func testAccProjectConfig_Source_GitCloneDepth(rName string, gitCloneDepth int) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    git_clone_depth = %[2]d
    location        = "https://github.com/hashicorp/packer.git"
    type            = "GITHUB"
  }
}
`, rName, gitCloneDepth))
}

func testAccProjectConfig_Source_GitSubmodulesConfig_CodeCommit(rName string, fetchSubmodules bool) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

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

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }
}
`, rName, fetchSubmodules))
}

func testAccProjectConfig_Source_GitSubmodulesConfig_GitHub(rName string, fetchSubmodules bool) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

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

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }
}
`, rName, fetchSubmodules))
}

func testAccProjectConfig_Source_GitSubmodulesConfig_GitHubEnterprise(rName string, fetchSubmodules bool) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

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

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }
}
`, rName, fetchSubmodules))
}

func testAccProjectConfig_SecondarySources_GitSubmodulesConfig_CodeCommit(rName string, fetchSubmodules bool) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

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

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }

  secondary_sources {
    location          = "https://git-codecommit.region-id.amazonaws.com/v1/repos/second-repo-name"
    type              = "CODECOMMIT"
    source_identifier = "secondarySource1"

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }

  secondary_sources {
    location          = "https://git-codecommit.region-id.amazonaws.com/v1/repos/third-repo-name"
    type              = "CODECOMMIT"
    source_identifier = "secondarySource2"

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }
}
`, rName, fetchSubmodules))
}

func testAccProjectConfig_SecondarySources_none(rName string, fetchSubmodules bool) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

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

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }
}
`, rName, fetchSubmodules))
}

func testAccProjectConfig_SecondarySources_GitSubmodulesConfig_GitHub(rName string, fetchSubmodules bool) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

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

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }

  secondary_sources {
    location          = "https://github.com/hashicorp/terraform.git"
    type              = "GITHUB"
    source_identifier = "secondarySource1"

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }

  secondary_sources {
    location          = "https://github.com/hashicorp/vault.git"
    type              = "GITHUB"
    source_identifier = "secondarySource2"

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }
}
`, rName, fetchSubmodules))
}

func testAccProjectConfig_SecondarySources_GitSubmodulesConfig_GitHubEnterprise(rName string, fetchSubmodules bool) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = "https://example.com/organization/repository-1.git"
    type     = "GITHUB_ENTERPRISE"

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }

  secondary_sources {
    location          = "https://example.com/organization/repository-2.git"
    type              = "GITHUB_ENTERPRISE"
    source_identifier = "secondarySource1"

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }

  secondary_sources {
    location          = "https://example.com/organization/repository-3.git"
    type              = "GITHUB_ENTERPRISE"
    source_identifier = "secondarySource2"

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }
}
`, rName, fetchSubmodules))
}

func testAccProjectConfig_SecondarySourceVersions_CodeCommit(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

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

  secondary_sources {
    location          = "https://git-codecommit.region-id.amazonaws.com/v1/repos/second-repo-name"
    type              = "CODECOMMIT"
    source_identifier = "secondarySource1"
  }

  secondary_sources {
    location          = "https://git-codecommit.region-id.amazonaws.com/v1/repos/third-repo-name"
    type              = "CODECOMMIT"
    source_identifier = "secondarySource2"
  }

  secondary_source_version {
    source_version    = "master"
    source_identifier = "secondarySource1"
  }
}
`, rName))
}

func testAccProjectConfig_SecondarySourceVersions_CodeCommitUpdated(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

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

  secondary_sources {
    location          = "https://git-codecommit.region-id.amazonaws.com/v1/repos/second-repo-name"
    type              = "CODECOMMIT"
    source_identifier = "secondarySource1"
  }

  secondary_sources {
    location          = "https://git-codecommit.region-id.amazonaws.com/v1/repos/third-repo-name"
    type              = "CODECOMMIT"
    source_identifier = "secondarySource2"
  }

  secondary_source_version {
    source_version    = "master"
    source_identifier = "secondarySource1"
  }

  secondary_source_version {
    source_version    = "master"
    source_identifier = "secondarySource2"
  }
}
`, rName))
}

func testAccProjectConfig_Source_InsecureSSL(rName string, insecureSSL bool) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    insecure_ssl = %[2]t
    location     = "https://github.com/hashicorp/packer.git"
    type         = "GITHUB"
  }
}
`, rName, insecureSSL))
}

func testAccProjectConfig_Source_ReportBuildStatus_Bitbucket(rName, sourceLocation string, reportBuildStatus bool) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location            = %[2]q
    report_build_status = %[3]t
    type                = "BITBUCKET"
  }
}
`, rName, sourceLocation, reportBuildStatus))
}

func testAccProjectConfig_Source_ReportBuildStatus_GitHub(rName string, reportBuildStatus bool) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

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
    report_build_status = %[2]t
    type                = "GITHUB"
  }
}
`, rName, reportBuildStatus))
}

func testAccProjectConfig_Source_ReportBuildStatus_GitHubEnterprise(rName string, reportBuildStatus bool) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location            = "https://example.com/organization/repository.git"
    report_build_status = %[2]t
    type                = "GITHUB_ENTERPRISE"
  }
}
`, rName, reportBuildStatus))
}

func testAccProjectConfig_Source_Type_Bitbucket(rName, sourceLocation string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = %[2]q
    type     = "BITBUCKET"
  }
}
`, rName, sourceLocation))
}

func testAccProjectConfig_Source_Type_CodeCommit(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

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
`, rName))
}

func testAccProjectConfig_Source_Type_CodePipeline(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

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
`, rName))
}

func testAccProjectConfig_Source_Type_GitHubEnterprise(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

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
`, rName))
}

func testAccProjectConfig_Source_Type_S3(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.bucket
  content = "test"
  key     = "test.txt"
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = "${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
    type     = "S3"
  }
}
`, rName))
}

func testAccProjectConfig_Source_Type_NoSource(rName string, rLocation string, rBuildspec string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type      = "NO_SOURCE"
    location  = %[2]q
    buildspec = %[3]q
  }
}
`, rName, rLocation, rBuildspec))
}

func testAccProjectConfig_Tags(rName, tagKey, tagValue string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

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

  tags = {
    tag1 = "tag1value"

    %[2]s = %[3]q
  }
}
`, rName, tagKey, tagValue))
}

func testAccProjectConfig_VPCConfig1(rName string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_Base_ServiceRole(rName),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 1

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

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
    security_group_ids = [aws_security_group.test.id]
    subnets            = aws_subnet.test[*].id
    vpc_id             = aws_vpc.test.id
  }
}
`, rName))
}

func testAccProjectConfig_VPCConfig2(rName string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_Base_ServiceRole(rName),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

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
    security_group_ids = [aws_security_group.test.id]
    subnets            = aws_subnet.test[*].id
    vpc_id             = aws_vpc.test.id
  }
}
`, rName))
}

func testAccProjectConfig_WindowsServer2019Container(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_MEDIUM"
    image        = "2"
    type         = "WINDOWS_SERVER_2019_CONTAINER"
  }

  source {
    location = %[2]q
    type     = "GITHUB"
  }
}
`, rName, testAccGitHubSourceLocationFromEnv()))
}

func testAccProjectConfig_ARMContainer(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_LARGE"
    image        = "2"
    type         = "ARM_CONTAINER"
  }

  source {
    location = %[2]q
    type     = "GITHUB"
  }
}
`, rName, testAccGitHubSourceLocationFromEnv()))
}

func testAccProjectConfig_Artifacts_ArtifactIdentifier(rName string, artifactIdentifier string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_Base_ServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    artifact_identifier = %[2]q
    location            = aws_s3_bucket.test.bucket
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
`, rName, artifactIdentifier))
}

func testAccProjectConfig_Artifacts_EncryptionDisabled(rName string, encryptionDisabled bool) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_Base_ServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    encryption_disabled = %[2]t
    location            = aws_s3_bucket.test.bucket
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
`, rName, encryptionDisabled))
}

func testAccProjectConfig_Artifacts_Location(rName, bucketName string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_Base_ServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location = aws_s3_bucket.test.bucket
    type     = "S3"
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
`, rName, bucketName))
}

func testAccProjectConfig_Artifacts_Name(rName string, name string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_Base_ServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    name     = %[2]q
    location = aws_s3_bucket.test.bucket
    type     = "S3"
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
`, rName, name))
}

func testAccProjectConfig_Artifacts_NamespaceType(rName, namespaceType string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_Base_ServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location       = aws_s3_bucket.test.bucket
    namespace_type = %[2]q
    type           = "S3"
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
`, rName, namespaceType))
}

func testAccProjectConfig_Artifacts_OverrideArtifactName(rName string, overrideArtifactName bool) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_Base_ServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    override_artifact_name = %[2]t
    location               = aws_s3_bucket.test.bucket
    type                   = "S3"
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
`, rName, overrideArtifactName))
}

func testAccProjectConfig_Artifacts_Packaging(rName, packaging string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_Base_ServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location  = aws_s3_bucket.test.bucket
    packaging = %[2]q
    type      = "S3"
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
`, rName, packaging))
}

func testAccProjectConfig_Artifacts_Path(rName, path string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_Base_ServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location = aws_s3_bucket.test.bucket
    path     = %[2]q
    type     = "S3"
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
`, rName, path))
}

func testAccProjectConfig_Artifacts_Type(rName string, artifactType string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_Base_ServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type     = %[2]q
    location = aws_s3_bucket.test.bucket
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = %[2]q
    location = "${aws_s3_bucket.test.bucket}/"
  }
}
`, rName, artifactType))
}

func testAccProjectConfig_Artifacts_BucketOwnerAccess(rName string, typ string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_Base_ServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type                = "S3"
    location            = aws_s3_bucket.test.bucket
    bucket_owner_access = %[2]q
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "S3"
    location = "${aws_s3_bucket.test.bucket}/"
  }
}
`, rName, typ))
}

func testAccProjectConfig_SecondaryArtifacts(rName string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_Base_ServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location = aws_s3_bucket.test.bucket
    type     = "S3"
  }

  secondary_artifacts {
    artifact_identifier = "secondaryArtifact1"
    location            = aws_s3_bucket.test.bucket
    type                = "S3"
  }

  secondary_artifacts {
    artifact_identifier = "secondaryArtifact2"
    location            = aws_s3_bucket.test.bucket
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
`, rName))
}

func testAccProjectConfig_SecondaryArtifacts_none(rName string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_Base_ServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location = aws_s3_bucket.test.bucket
    type     = "S3"
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
`, rName))
}

func testAccProjectConfig_SecondaryArtifacts_ArtifactIdentifier(rName string, artifactIdentifier string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_Base_ServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location = aws_s3_bucket.test.bucket
    type     = "S3"
  }

  secondary_artifacts {
    artifact_identifier = %[2]q
    location            = aws_s3_bucket.test.bucket
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
`, rName, artifactIdentifier))
}

func testAccProjectConfig_SecondaryArtifacts_EncryptionDisabled(rName string, encryptionDisabled bool) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_Base_ServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location = aws_s3_bucket.test.bucket
    type     = "S3"
  }

  secondary_artifacts {
    artifact_identifier = "secondaryArtifact1"
    encryption_disabled = %[2]t
    location            = aws_s3_bucket.test.bucket
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
`, rName, encryptionDisabled))
}

func testAccProjectConfig_SecondaryArtifacts_Location(rName, bucketName string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_Base_ServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location = aws_s3_bucket.test.bucket
    type     = "S3"
  }

  secondary_artifacts {
    artifact_identifier = "secondaryArtifact1"
    location            = aws_s3_bucket.test.bucket
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
`, rName, bucketName))
}

func testAccProjectConfig_SecondaryArtifacts_Name(rName string, name string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_Base_ServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location = aws_s3_bucket.test.bucket
    type     = "S3"
  }

  secondary_artifacts {
    artifact_identifier = "secondaryArtifact1"
    name                = %[2]q
    location            = aws_s3_bucket.test.bucket
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
`, rName, name))
}

func testAccProjectConfig_SecondaryArtifacts_NamespaceType(rName, namespaceType string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_Base_ServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location = aws_s3_bucket.test.bucket
    type     = "S3"
  }

  secondary_artifacts {
    artifact_identifier = "secondaryArtifact1"
    namespace_type      = %[2]q
    location            = aws_s3_bucket.test.bucket
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
`, rName, namespaceType))
}

func testAccProjectConfig_SecondaryArtifacts_OverrideArtifactName(rName string, overrideArtifactName bool) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_Base_ServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location = aws_s3_bucket.test.bucket
    type     = "S3"
  }

  secondary_artifacts {
    artifact_identifier    = "secondaryArtifact1"
    override_artifact_name = %[2]t
    location               = aws_s3_bucket.test.bucket
    type                   = "S3"
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
`, rName, overrideArtifactName))
}

func testAccProjectConfig_SecondaryArtifacts_Packaging(rName, packaging string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_Base_ServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location = aws_s3_bucket.test.bucket
    type     = "S3"
  }

  secondary_artifacts {
    artifact_identifier = "secondaryArtifact1"
    packaging           = %[2]q
    location            = aws_s3_bucket.test.bucket
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
`, rName, packaging))
}

func testAccProjectConfig_SecondaryArtifacts_Path(rName, path string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_Base_ServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location = aws_s3_bucket.test.bucket
    type     = "S3"
  }

  secondary_artifacts {
    artifact_identifier = "secondaryArtifact1"
    path                = %[2]q
    location            = aws_s3_bucket.test.bucket
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
`, rName, path))
}

func testAccProjectConfig_SecondaryArtifacts_Type(rName string, artifactType string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_Base_ServiceRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    location = aws_s3_bucket.test.bucket
    type     = "S3"
  }

  secondary_artifacts {
    artifact_identifier = "secondaryArtifact1"
    type                = %[2]q
    location            = aws_s3_bucket.test.bucket
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
`, rName, artifactType))
}

func testAccProjectConfig_SecondarySources_CodeCommit(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

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

  secondary_sources {
    location          = "https://git-codecommit.region-id.amazonaws.com/v1/repos/second-repo-name"
    type              = "CODECOMMIT"
    source_identifier = "secondarySource1"
  }

  secondary_sources {
    location          = "https://git-codecommit.region-id.amazonaws.com/v1/repos/third-repo-name"
    type              = "CODECOMMIT"
    source_identifier = "secondarySource2"
  }
}
`, rName))
}

func testAccProjectConfig_Source_BuildStatusConfig_GitHubEnterprise(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

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

    build_status_config {
      context    = "codebuild"
      target_url = "https://example.com/$${CODEBUILD_BUILD_ID}"
    }
  }
}
`, rName))
}

func testAccProjectConfig_ConcurrentBuildLimit(rName string, concurrentBuildLimit int) string {
	return acctest.ConfigCompose(testAccProjectConfig_Base_ServiceRole(rName), fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  concurrent_build_limit = %[1]d
  name                   = %[2]q
  service_role           = aws_iam_role.test.arn

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
`, concurrentBuildLimit, rName))
}

func testAccProjectConfig_FileSystemLocations(rName, mountPoint string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_Base_ServiceRole(rName),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "public" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "%[1]s-public"
  }
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "%[1]s-public"
  }
}

resource "aws_route_table_association" "public" {
  route_table_id = aws_route_table.public.id
  subnet_id      = aws_subnet.public.id
}

resource "aws_route" "public" {
  route_table_id         = aws_route_table.public.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.test.id
}

resource "aws_subnet" "private" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.1.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "%[1]s-private"
  }
}

resource "aws_eip" "test" {
  vpc = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_nat_gateway" "test" {
  allocation_id = aws_eip.test.id
  subnet_id     = aws_subnet.public.id

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_route.public]
}

resource "aws_route_table" "private" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "%[1]s-private"
  }
}

resource "aws_route_table_association" "private" {
  route_table_id = aws_route.private.route_table_id
  subnet_id      = aws_subnet.private.id
}

resource "aws_route" "private" {
  route_table_id         = aws_route_table.private.id
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = aws_nat_gateway.test.id
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"

    privileged_mode = true
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }

  vpc_config {
    security_group_ids = [aws_security_group.test.id]
    subnets            = [aws_subnet.private.id]
    vpc_id             = aws_vpc.test.id
  }

  file_system_locations {
    identifier    = "test"
    location      = "${aws_efs_file_system.test.dns_name}:/directory-path"
    type          = "EFS"
    mount_point   = %[2]q
    mount_options = "nfsvers=4.1,rsize=1048576,wsize=1048576,hard,timeo=450,retrans=3"
  }
}
`, rName, mountPoint))
}
