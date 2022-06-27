package ecs_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/ecs"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfecs "github.com/hashicorp/terraform-provider-aws/internal/service/ecs"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(ecs.EndpointsID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"Unsupported field 'inferenceAccelerators'",
	)
}

func TestAccECSTaskDefinition_basic(t *testing.T) {
	var def ecs.TaskDefinition

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &def),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ecs", regexp.MustCompile(`task-definition/.+`)),
				),
			},
			{
				Config: testAccTaskDefinitionConfig_modified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &def),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ecs", regexp.MustCompile(`task-definition/.+`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform/issues/2370
func TestAccECSTaskDefinition_scratchVolume(t *testing.T) {
	var def ecs.TaskDefinition

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_scratchVolume(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &def),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

func TestAccECSTaskDefinition_DockerVolume_basic(t *testing.T) {
	var def ecs.TaskDefinition

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_dockerVolumes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &def),
					resource.TestCheckResourceAttr(resourceName, "volume.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "volume.*", map[string]string{
						"name":                                             rName,
						"docker_volume_configuration.#":                    "1",
						"docker_volume_configuration.0.driver":             "local",
						"docker_volume_configuration.0.scope":              "shared",
						"docker_volume_configuration.0.autoprovision":      "true",
						"docker_volume_configuration.0.driver_opts.%":      "2",
						"docker_volume_configuration.0.driver_opts.device": "tmpfs",
						"docker_volume_configuration.0.driver_opts.uid":    "1000",
						"docker_volume_configuration.0.labels.%":           "2",
						"docker_volume_configuration.0.labels.environment": "test",
						"docker_volume_configuration.0.labels.stack":       "april",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

func TestAccECSTaskDefinition_DockerVolume_minimal(t *testing.T) {
	var def ecs.TaskDefinition

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_dockerVolumesMinimal(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &def),
					resource.TestCheckResourceAttr(resourceName, "volume.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "volume.*", map[string]string{
						"name":                          rName,
						"docker_volume_configuration.#": "1",
						"docker_volume_configuration.0.autoprovision": "true",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

func TestAccECSTaskDefinition_runtimePlatform(t *testing.T) {
	var def ecs.TaskDefinition

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartition(t, endpoints.AwsPartitionID) }, // runtime platform not support on GovCloud
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_runtimePlatformMinimal(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &def),
					resource.TestCheckResourceAttr(resourceName, "runtime_platform.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "runtime_platform.*", map[string]string{
						"operating_system_family": "LINUX",
						"cpu_architecture":        "X86_64",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

func TestAccECSTaskDefinition_Fargate_runtimePlatform(t *testing.T) {
	var def ecs.TaskDefinition

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartition(t, endpoints.AwsPartitionID) }, // runtime platform not support on GovCloud
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_fargateRuntimePlatformMinimal(rName, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &def),
					resource.TestCheckResourceAttr(resourceName, "runtime_platform.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "runtime_platform.*", map[string]string{
						"operating_system_family": "WINDOWS_SERVER_2019_CORE",
						"cpu_architecture":        "X86_64",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

func TestAccECSTaskDefinition_Fargate_runtimePlatformWithoutArch(t *testing.T) {
	var def ecs.TaskDefinition

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartition(t, endpoints.AwsPartitionID) }, // runtime platform not support on GovCloud
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_fargateRuntimePlatformMinimal(rName, false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &def),
					resource.TestCheckResourceAttr(resourceName, "runtime_platform.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "runtime_platform.*", map[string]string{
						"operating_system_family": "WINDOWS_SERVER_2019_CORE",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

func TestAccECSTaskDefinition_EFSVolume_minimal(t *testing.T) {
	var def ecs.TaskDefinition

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_efsVolumeMinimal(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &def),
					resource.TestCheckResourceAttr(resourceName, "volume.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "volume.*", map[string]string{
						"name":                       rName,
						"efs_volume_configuration.#": "1",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "volume.*.efs_volume_configuration.0.file_system_id", "aws_efs_file_system.test", "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

func TestAccECSTaskDefinition_EFSVolume_basic(t *testing.T) {
	var def ecs.TaskDefinition

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_efsVolume(rName, "/home/test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &def),
					resource.TestCheckResourceAttr(resourceName, "volume.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "volume.*", map[string]string{
						"name":                       rName,
						"efs_volume_configuration.#": "1",
						"efs_volume_configuration.0.root_directory": "/home/test",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "volume.*.efs_volume_configuration.0.file_system_id", "aws_efs_file_system.test", "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

func TestAccECSTaskDefinition_EFSVolume_transitEncryption(t *testing.T) {
	var def ecs.TaskDefinition

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_definition.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_transitEncryptionEFSVolume(rName, "ENABLED", 2999),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &def),
					resource.TestCheckResourceAttr(resourceName, "volume.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "volume.*", map[string]string{
						"name":                       rName,
						"efs_volume_configuration.#": "1",
						"efs_volume_configuration.0.root_directory":          "/home/test",
						"efs_volume_configuration.0.transit_encryption":      "ENABLED",
						"efs_volume_configuration.0.transit_encryption_port": "2999",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "volume.*.efs_volume_configuration.0.file_system_id", "aws_efs_file_system.test", "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

func TestAccECSTaskDefinition_EFSVolume_accessPoint(t *testing.T) {
	var def ecs.TaskDefinition

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_efsAccessPoint(rName, "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &def),
					resource.TestCheckResourceAttr(resourceName, "volume.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "volume.*", map[string]string{
						"name":                       rName,
						"efs_volume_configuration.#": "1",
						"efs_volume_configuration.0.root_directory":             "/",
						"efs_volume_configuration.0.transit_encryption":         "ENABLED",
						"efs_volume_configuration.0.transit_encryption_port":    "2999",
						"efs_volume_configuration.0.authorization_config.#":     "1",
						"efs_volume_configuration.0.authorization_config.0.iam": "DISABLED",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "volume.*.efs_volume_configuration.0.file_system_id", "aws_efs_file_system.test", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "volume.*.efs_volume_configuration.0.authorization_config.0.access_point_id", "aws_efs_access_point.test", "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

func TestAccECSTaskDefinition_fsxWinFileSystem(t *testing.T) {
	var def ecs.TaskDefinition

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_definition.test"

	domainName := acctest.RandomDomainName()

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	if acctest.Partition() == "aws-us-gov" {
		t.Skip("Amazon FSx for Windows File Server volumes for ECS tasks are not supported in GovCloud partition")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_fsxVolume(domainName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &def),
					resource.TestCheckResourceAttr(resourceName, "volume.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "volume.*", map[string]string{
						"name": rName,
						"fsx_windows_file_server_volume_configuration.#":                        "1",
						"fsx_windows_file_server_volume_configuration.0.root_directory":         "\\data",
						"fsx_windows_file_server_volume_configuration.0.authorization_config.#": "1",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "volume.*.fsx_windows_file_server_volume_configuration.0.file_system_id", "aws_fsx_windows_file_system.test", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "volume.*.fsx_windows_file_server_volume_configuration.0.authorization_config.0.credentials_parameter", "aws_secretsmanager_secret_version.test", "arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "volume.*.fsx_windows_file_server_volume_configuration.0.authorization_config.0.domain", "aws_directory_service_directory.test", "name"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

func TestAccECSTaskDefinition_DockerVolume_taskScoped(t *testing.T) {
	var def ecs.TaskDefinition

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_scopedDockerVolume(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &def),
					testAccCheckTaskDefinitionDockerVolumeConfigurationAutoprovisionNil(&def),
					resource.TestCheckResourceAttr(resourceName, "volume.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform/issues/2694
func TestAccECSTaskDefinition_service(t *testing.T) {
	var def ecs.TaskDefinition
	var service ecs.Service

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_service(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &def),
					testAccCheckServiceExists("aws_ecs_service.test", &service),
				),
			},
			{
				Config: testAccTaskDefinitionConfig_serviceModified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &def),
					testAccCheckServiceExists("aws_ecs_service.test", &service),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

func TestAccECSTaskDefinition_taskRoleARN(t *testing.T) {
	var def ecs.TaskDefinition

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_roleARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &def),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

func TestAccECSTaskDefinition_networkMode(t *testing.T) {
	var def ecs.TaskDefinition

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_networkMode(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &def),
					resource.TestCheckResourceAttr(resourceName, "network_mode", "bridge"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

func TestAccECSTaskDefinition_ipcMode(t *testing.T) {
	var def ecs.TaskDefinition

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_ipcMode(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &def),
					resource.TestCheckResourceAttr(resourceName, "ipc_mode", "host"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

func TestAccECSTaskDefinition_pidMode(t *testing.T) {
	var def ecs.TaskDefinition

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_pidMode(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &def),
					resource.TestCheckResourceAttr(resourceName, "pid_mode", "host"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

func TestAccECSTaskDefinition_constraint(t *testing.T) {
	var def ecs.TaskDefinition

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_constraint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &def),
					resource.TestCheckResourceAttr(resourceName, "placement_constraints.#", "1"),
					testAccCheckTaskDefinitionConstraintsAttrs(&def),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

func TestAccECSTaskDefinition_changeVolumesForcesNewResource(t *testing.T) {
	var before ecs.TaskDefinition
	var after ecs.TaskDefinition

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &before),
				),
			},
			{
				Config: testAccTaskDefinitionConfig_updatedVolume(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &after),
					testAccCheckTaskDefinitionRecreated(t, &before, &after),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform-provider-aws/issues/2336
func TestAccECSTaskDefinition_arrays(t *testing.T) {
	var conf ecs.TaskDefinition
	resourceName := "aws_ecs_task_definition.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_arrays(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &conf),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

func TestAccECSTaskDefinition_Fargate_basic(t *testing.T) {
	var conf ecs.TaskDefinition

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_fargate(rName, `[{"protocol": "tcp", "containerPort": 8000}]`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "requires_compatibilities.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cpu", "256"),
					resource.TestCheckResourceAttr(resourceName, "memory", "512"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
			{
				ExpectNonEmptyPlan: false,
				PlanOnly:           true,
				Config:             testAccTaskDefinitionConfig_fargate(rName, `[{"protocol": "tcp", "containerPort": 8000, "hostPort": 8000}]`),
			},
		},
	})
}

func TestAccECSTaskDefinition_Fargate_ephemeralStorage(t *testing.T) {
	var conf ecs.TaskDefinition

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_fargateEphemeralStorage(rName, `[{"protocol": "tcp", "containerPort": 8000}]`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "requires_compatibilities.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cpu", "256"),
					resource.TestCheckResourceAttr(resourceName, "memory", "512"),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_storage.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_storage.0.size_in_gib", "30"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

func TestAccECSTaskDefinition_executionRole(t *testing.T) {
	var conf ecs.TaskDefinition

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_executionRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &conf),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform/issues/3582#issuecomment-286409786
func TestAccECSTaskDefinition_disappears(t *testing.T) {
	var def ecs.TaskDefinition

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &def),
					acctest.CheckResourceDisappears(acctest.Provider, tfecs.ResourceTaskDefinition(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccTaskDefinitionConfig_basic(rName),
				Check:  resource.TestCheckResourceAttr(resourceName, "revision", "2"), // should get re-created
			},
		},
	})
}

func TestAccECSTaskDefinition_tags(t *testing.T) {
	var taskDefinition ecs.TaskDefinition
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &taskDefinition),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
			{
				Config: testAccTaskDefinitionConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &taskDefinition),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccTaskDefinitionConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &taskDefinition),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccECSTaskDefinition_proxy(t *testing.T) {
	var taskDefinition ecs.TaskDefinition
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_definition.test"

	containerName := "web"
	proxyType := "APPMESH"
	ignoredUid := "1337"
	ignoredGid := "999"
	appPorts := "80"
	proxyIngressPort := "15000"
	proxyEgressPort := "15001"
	egressIgnoredPorts := "5500"
	egressIgnoredIPs := "169.254.170.2,169.254.169.254"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_proxyConfiguration(rName, containerName, proxyType, ignoredUid, ignoredGid, appPorts, proxyIngressPort, proxyEgressPort, egressIgnoredPorts, egressIgnoredIPs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &taskDefinition),
					testAccCheckTaskDefinitionProxyConfiguration(&taskDefinition, containerName, proxyType, ignoredUid, ignoredGid, appPorts, proxyIngressPort, proxyEgressPort, egressIgnoredPorts, egressIgnoredIPs),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

func TestAccECSTaskDefinition_inferenceAccelerator(t *testing.T) {
	var def ecs.TaskDefinition

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionConfig_inferenceAccelerator(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskDefinitionExists(resourceName, &def),
					resource.TestCheckResourceAttr(resourceName, "inference_accelerator.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

func testAccTaskDefinitionConfig_proxyConfiguration(rName string, containerName string, proxyType string,
	ignoredUid string, ignoredGid string, appPorts string, proxyIngressPort string, proxyEgressPort string,
	egressIgnoredPorts string, egressIgnoredIPs string) string {

	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family       = %[1]q
  network_mode = "awsvpc"

  proxy_configuration {
    type           = %[2]q
    container_name = %[3]q
    properties = {
      IgnoredUID         = %[4]q
      IgnoredGID         = %[5]q
      AppPorts           = %[6]q
      ProxyIngressPort   = %[7]q
      ProxyEgressPort    = %[8]q
      EgressIgnoredPorts = %[9]q
      EgressIgnoredIPs   = %[10]q
    }
  }

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "nginx:latest",
    "memory": 128,
    "name": %[11]q
  }
]
DEFINITION
}
`, rName, proxyType, containerName, ignoredUid, ignoredGid, appPorts, proxyIngressPort, proxyEgressPort, egressIgnoredPorts, egressIgnoredIPs, containerName)
}

func testAccCheckTaskDefinitionProxyConfiguration(after *ecs.TaskDefinition, containerName string, proxyType string,
	ignoredUid string, ignoredGid string, appPorts string, proxyIngressPort string, proxyEgressPort string,
	egressIgnoredPorts string, egressIgnoredIPs string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *after.ProxyConfiguration.Type != proxyType {
			return fmt.Errorf("Expected (%s) ProxyConfiguration.Type, got (%s)", proxyType, *after.ProxyConfiguration.Type)
		}

		if *after.ProxyConfiguration.ContainerName != containerName {
			return fmt.Errorf("Expected (%s) ProxyConfiguration.ContainerName, got (%s)", containerName, *after.ProxyConfiguration.ContainerName)
		}

		properties := after.ProxyConfiguration.Properties
		expectedProperties := []string{"IgnoredUID", "IgnoredGID", "AppPorts", "ProxyIngressPort", "ProxyEgressPort", "EgressIgnoredPorts", "EgressIgnoredIPs"}
		if len(properties) != len(expectedProperties) {
			return fmt.Errorf("Expected (%d) ProxyConfiguration.Property count, got (%d)", len(expectedProperties), len(properties))
		}

		propertyLookups := make(map[string]string)
		for _, property := range properties {
			propertyLookups[aws.StringValue(property.Name)] = aws.StringValue(property.Value)
		}

		if propertyLookups["IgnoredUID"] != ignoredUid {
			return fmt.Errorf("Expected (%s) ProxyConfiguration.Properties.IgnoredUID, got (%s)", ignoredUid, propertyLookups["IgnoredUID"])
		}

		if propertyLookups["IgnoredGID"] != ignoredGid {
			return fmt.Errorf("Expected (%s) ProxyConfiguration.Properties.IgnoredGID, got (%s)", ignoredGid, propertyLookups["IgnoredGID"])
		}

		if propertyLookups["AppPorts"] != appPorts {
			return fmt.Errorf("Expected (%s) ProxyConfiguration.Properties.AppPorts, got (%s)", appPorts, propertyLookups["AppPorts"])
		}

		if propertyLookups["ProxyIngressPort"] != proxyIngressPort {
			return fmt.Errorf("Expected (%s) ProxyConfiguration.Properties.ProxyIngressPort, got (%s)", proxyIngressPort, propertyLookups["ProxyIngressPort"])
		}

		if propertyLookups["ProxyEgressPort"] != proxyEgressPort {
			return fmt.Errorf("Expected (%s) ProxyConfiguration.Properties.ProxyEgressPort, got (%s)", proxyEgressPort, propertyLookups["ProxyEgressPort"])
		}

		if propertyLookups["EgressIgnoredPorts"] != egressIgnoredPorts {
			return fmt.Errorf("Expected (%s) ProxyConfiguration.Properties.EgressIgnoredPorts, got (%s)", egressIgnoredPorts, propertyLookups["EgressIgnoredPorts"])
		}

		if propertyLookups["EgressIgnoredIPs"] != egressIgnoredIPs {
			return fmt.Errorf("Expected (%s) ProxyConfiguration.Properties.EgressIgnoredIPs, got (%s)", egressIgnoredIPs, propertyLookups["EgressIgnoredIPs"])
		}

		return nil
	}
}

func testAccCheckTaskDefinitionRecreated(t *testing.T,
	before, after *ecs.TaskDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.Revision == *after.Revision {
			t.Fatalf("Expected change of TaskDefinition Revisions, but both were %v", before.Revision)
		}
		return nil
	}
}

func testAccCheckTaskDefinitionConstraintsAttrs(def *ecs.TaskDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(def.PlacementConstraints) != 1 {
			return fmt.Errorf("Expected (1) placement_constraints, got (%d)", len(def.PlacementConstraints))
		}
		return nil
	}
}

func TestValidTaskDefinitionContainerDefinitions(t *testing.T) {
	validDefinitions := []string{
		testValidTaskDefinitionValidContainerDefinitions,
	}
	for _, v := range validDefinitions {
		_, errors := tfecs.ValidTaskDefinitionContainerDefinitions(v, "container_definitions")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid AWS ECS Task Definition Container Definitions: %q", v, errors)
		}
	}

	invalidDefinitions := []string{
		testValidTaskDefinitionInvalidCommandContainerDefinitions,
	}
	for _, v := range invalidDefinitions {
		_, errors := tfecs.ValidTaskDefinitionContainerDefinitions(v, "container_definitions")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid AWS ECS Task Definition Container Definitions", v)
		}
	}
}

func testAccCheckTaskDefinitionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ECSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ecs_task_definition" {
			continue
		}

		input := ecs.DescribeTaskDefinitionInput{
			TaskDefinition: aws.String(rs.Primary.Attributes["arn"]),
		}

		out, err := conn.DescribeTaskDefinition(&input)

		if err != nil {
			return err
		}

		if out.TaskDefinition != nil && *out.TaskDefinition.Status != ecs.TaskDefinitionStatusInactive {
			return fmt.Errorf("ECS task definition still exists:\n%#v", *out.TaskDefinition)
		}
	}

	return nil
}

func testAccCheckTaskDefinitionExists(name string, def *ecs.TaskDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECSConn

		out, err := conn.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
			TaskDefinition: aws.String(rs.Primary.Attributes["arn"]),
		})
		if err != nil {
			return err
		}
		*def = *out.TaskDefinition

		return nil
	}
}

func testAccCheckTaskDefinitionDockerVolumeConfigurationAutoprovisionNil(def *ecs.TaskDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(def.Volumes) != 1 {
			return fmt.Errorf("Expected (1) volumes, got (%d)", len(def.Volumes))
		}
		config := def.Volumes[0].DockerVolumeConfiguration
		if config == nil {
			return fmt.Errorf("Expected docker_volume_configuration, got nil")
		}
		if config.Autoprovision != nil {
			return fmt.Errorf("Expected autoprovision to be nil, got %t", *config.Autoprovision)
		}
		return nil
	}
}

func testAccTaskDefinitionConfig_constraint(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<TASK_DEFINITION
[
	{
		"cpu": 10,
		"command": ["sleep", "10"],
		"entryPoint": ["/"],
		"environment": [
			{"name": "VARNAME", "value": "VARVAL"}
		],
		"essential": true,
		"image": "jenkins",
		"links": ["mongodb"],
		"memory": 128,
		"name": "jenkins",
		"portMappings": [
			{
				"containerPort": 80,
				"hostPort": 8080
			}
		]
	},
	{
		"cpu": 10,
		"command": ["sleep", "10"],
		"entryPoint": ["/"],
		"essential": true,
		"image": "mongodb",
		"memory": 128,
		"name": "mongodb",
		"portMappings": [
			{
				"containerPort": 28017,
				"hostPort": 28017
			}
		]
	}
]
TASK_DEFINITION

  volume {
    name      = "jenkins-home"
    host_path = "/ecs/jenkins-home"
  }

  placement_constraints {
    type       = "memberOf"
    expression = "attribute:ecs.availability-zone in [${data.aws_availability_zones.available.names[0]}, ${data.aws_availability_zones.available.names[1]}]"
  }
}
`, rName))
}

func testAccTaskDefinitionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<TASK_DEFINITION
[
	{
		"cpu": 10,
		"command": ["sleep", "10"],
		"entryPoint": ["/"],
		"environment": [
			{"name": "VARNAME", "value": "VARVAL"}
		],
		"essential": true,
		"image": "jenkins",
		"links": ["mongodb"],
		"memory": 128,
		"name": "jenkins",
		"portMappings": [
			{
				"containerPort": 80,
				"hostPort": 8080
			}
		]
	},
	{
		"cpu": 10,
		"command": ["sleep", "10"],
		"entryPoint": ["/"],
		"essential": true,
		"image": "mongodb",
		"memory": 128,
		"name": "mongodb",
		"portMappings": [
			{
				"containerPort": 28017,
				"hostPort": 28017
			}
		]
	}
]
TASK_DEFINITION

  volume {
    name      = "jenkins-home"
    host_path = "/ecs/jenkins-home"
  }
}
`, rName)
}

func testAccTaskDefinitionConfig_updatedVolume(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<TASK_DEFINITION
[
	{
		"cpu": 10,
		"command": ["sleep", "10"],
		"entryPoint": ["/"],
		"environment": [
			{"name": "VARNAME", "value": "VARVAL"}
		],
		"essential": true,
		"image": "jenkins",
		"links": ["mongodb"],
		"memory": 128,
		"name": "jenkins",
		"portMappings": [
			{
				"containerPort": 80,
				"hostPort": 8080
			}
		]
	},
	{
		"cpu": 10,
		"command": ["sleep", "10"],
		"entryPoint": ["/"],
		"essential": true,
		"image": "mongodb",
		"memory": 128,
		"name": "mongodb",
		"portMappings": [
			{
				"containerPort": 28017,
				"hostPort": 28017
			}
		]
	}
]
TASK_DEFINITION

  volume {
    name      = "jenkins-home"
    host_path = "/ecs/jenkins"
  }
}
`, rName)
}

func testAccTaskDefinitionConfig_arrays(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<TASK_DEFINITION
[
    {
      "name": "wordpress",
      "image": "wordpress",
      "essential": true,
      "links": ["container1", "container2", "container3"],
      "portMappings": [
        {"containerPort": 80},
        {"containerPort": 81},
        {"containerPort": 82}
      ],
      "environment": [
        {"name": "VARNAME1", "value": "VARVAL1"},
        {"name": "VARNAME2", "value": "VARVAL2"},
        {"name": "VARNAME3", "value": "VARVAL3"}
      ],
      "extraHosts": [
        {"hostname": "host1", "ipAddress": "127.0.0.1"},
        {"hostname": "host2", "ipAddress": "127.0.0.2"},
        {"hostname": "host3", "ipAddress": "127.0.0.3"}
      ],
      "mountPoints": [
        {"sourceVolume": "vol1", "containerPath": "/vol1"},
        {"sourceVolume": "vol2", "containerPath": "/vol2"},
        {"sourceVolume": "vol3", "containerPath": "/vol3"}
      ],
      "volumesFrom": [
        {"sourceContainer": "container1"},
        {"sourceContainer": "container2"},
        {"sourceContainer": "container3"}
      ],
      "ulimits": [
        {
          "name": "core",
          "softLimit": 10, "hardLimit": 20
        },
        {
          "name": "cpu",
          "softLimit": 10, "hardLimit": 20
        },
        {
          "name": "fsize",
          "softLimit": 10, "hardLimit": 20
        }
      ],
      "linuxParameters": {
        "capabilities": {
          "add": ["AUDIT_CONTROL", "AUDIT_WRITE", "BLOCK_SUSPEND"],
          "drop": ["CHOWN", "IPC_LOCK", "KILL"]
        }
      },
      "devices": [
        {
          "hostPath": "/path1",
          "permissions": ["read", "write", "mknod"]
        },
        {
          "hostPath": "/path2",
          "permissions": ["read", "write"]
        },
        {
          "hostPath": "/path3",
          "permissions": ["read", "mknod"]
        }
      ],
      "dockerSecurityOptions": ["label:one", "label:two", "label:three"],
      "memory": 500,
      "cpu": 10
    },
    {
      "name": "container1",
      "image": "busybox",
      "memory": 100
    },
    {
      "name": "container2",
      "image": "busybox",
      "memory": 100
    },
    {
      "name": "container3",
      "image": "busybox",
      "memory": 100
    }
]
TASK_DEFINITION

  volume {
    name      = "vol1"
    host_path = "/host/vol1"
  }

  volume {
    name      = "vol2"
    host_path = "/host/vol2"
  }

  volume {
    name      = "vol3"
    host_path = "/host/vol3"
  }
}
`, rName)
}

func testAccTaskDefinitionConfig_fargate(rName, portMappings string) string {
	return fmt.Sprintf(`
resource "aws_ecs_task_definition" "test" {
  family                   = %[1]q
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"

  container_definitions = <<TASK_DEFINITION
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 10,
    "command": ["sleep","360"],
    "memory": 10,
    "essential": true,
    "portMappings": %[2]s
  }
]
TASK_DEFINITION
}
`, rName, portMappings)
}

func testAccTaskDefinitionConfig_fargateEphemeralStorage(rName, portMappings string) string {
	return fmt.Sprintf(`
resource "aws_ecs_task_definition" "test" {
  family                   = %[1]q
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"

  ephemeral_storage {
    size_in_gib = 30
  }

  container_definitions = <<TASK_DEFINITION
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 10,
    "command": ["sleep","360"],
    "memory": 10,
    "essential": true,
    "portMappings": %[2]s
  }
]
TASK_DEFINITION
}
`, rName, portMappings)
}

func testAccTaskDefinitionConfig_executionRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ecs-tasks.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test" {
  name        = %[1]q
  description = "A test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken",
        "ecr:BatchCheckLayerAvailability",
        "ecr:GetDownloadUrlForLayer",
        "ecr:BatchGetImage",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test.arn
}

resource "aws_ecs_task_definition" "test" {
  family             = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  container_definitions = <<TASK_DEFINITION
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 10,
    "command": ["sleep","360"],
    "memory": 10,
    "essential": true
  }
]
TASK_DEFINITION
}
`, rName)
}

func testAccTaskDefinitionConfig_scratchVolume(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<TASK_DEFINITION
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 10,
    "command": ["sleep","360"],
    "memory": 10,
    "essential": true
  }
]
TASK_DEFINITION

  volume {
    name = %[1]q
  }
}
`, rName)
}

func testAccTaskDefinitionConfig_dockerVolumes(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<TASK_DEFINITION
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 10,
    "command": ["sleep","360"],
    "memory": 10,
    "essential": true
  }
]
TASK_DEFINITION

  volume {
    name = %[1]q

    docker_volume_configuration {
      driver = "local"
      scope  = "shared"

      driver_opts = {
        device = "tmpfs"
        uid    = "1000"
      }

      labels = {
        environment = "test"
        stack       = "april"
      }

      autoprovision = true
    }
  }
}
`, rName)
}

func testAccTaskDefinitionConfig_dockerVolumesMinimal(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<TASK_DEFINITION
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 10,
    "command": ["sleep","360"],
    "memory": 10,
    "essential": true
  }
]
TASK_DEFINITION

  volume {
    name = %[1]q

    docker_volume_configuration {
      autoprovision = true
    }
  }
}
`, rName)
}

func testAccTaskDefinitionConfig_runtimePlatformMinimal(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_task_definition" "test" {
  family                = %[1]q
  container_definitions = <<TASK_DEFINITION
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 10,
    "command": ["sleep","360"],
    "memory": 10,
    "essential": true
  }
]
TASK_DEFINITION
  runtime_platform {
    operating_system_family = "LINUX"
    cpu_architecture        = "X86_64"
  }
}
`, rName)
}

func testAccTaskDefinitionConfig_fargateRuntimePlatformMinimal(rName string, architecture bool, osFamily bool) string {

	var arch string
	if architecture {
		arch = `cpu_architecture         = "X86_64"`
	} else {
		arch = ``
	}

	var os string
	if osFamily {
		os = `operating_system_family  = "WINDOWS_SERVER_2019_CORE"`
	} else {
		os = ``
	}

	return fmt.Sprintf(`
resource "aws_ecs_task_definition" "test" {
  family                   = %[1]q
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = 1024
  memory                   = 2048

  runtime_platform {
    %[3]s
    %[2]s
  }

  container_definitions = <<TASK_DEFINITION
[
  {
    "name": "iis",
    "image": "mcr.microsoft.com/windows/servercore/iis",
    "cpu": 1024,
    "memory": 2048,
    "essential": true
  }
]
TASK_DEFINITION

}
`, rName, arch, os)
}

func testAccTaskDefinitionConfig_scopedDockerVolume(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<TASK_DEFINITION
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 10,
    "command": ["sleep","360"],
    "memory": 10,
    "essential": true
  }
]
TASK_DEFINITION

  volume {
    name = %[1]q

    docker_volume_configuration {
      scope = "task"
    }
  }
}
`, rName)
}

func testAccTaskDefinitionConfig_efsVolumeMinimal(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<TASK_DEFINITION
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 10,
    "command": ["sleep","360"],
    "memory": 10,
    "essential": true
  }
]
TASK_DEFINITION

  volume {
    name = %[1]q

    efs_volume_configuration {
      file_system_id = aws_efs_file_system.test.id
    }
  }
}
`, rName)
}

func testAccTaskDefinitionConfig_efsVolume(rName, rDir string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<TASK_DEFINITION
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 10,
    "command": ["sleep","360"],
    "memory": 10,
    "essential": true
  }
]
TASK_DEFINITION

  volume {
    name = %[1]q

    efs_volume_configuration {
      file_system_id = aws_efs_file_system.test.id
      root_directory = %[2]q
    }
  }
}
`, rName, rDir)
}

func testAccTaskDefinitionConfig_transitEncryptionEFSVolume(rName, tEnc string, tEncPort int) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<TASK_DEFINITION
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 10,
    "command": ["sleep","360"],
    "memory": 10,
    "essential": true
  }
]
TASK_DEFINITION

  volume {
    name = %[1]q

    efs_volume_configuration {
      file_system_id          = aws_efs_file_system.test.id
      root_directory          = "/home/test"
      transit_encryption      = %[2]q
      transit_encryption_port = %[3]d
    }
  }
}
`, rName, tEnc, tEncPort)
}

func testAccTaskDefinitionConfig_efsAccessPoint(rName, useIam string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id
  posix_user {
    gid = 1001
    uid = 1001
  }
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<TASK_DEFINITION
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 10,
    "command": ["sleep","360"],
    "memory": 10,
    "essential": true
  }
]
TASK_DEFINITION

  volume {
    name = %[1]q

    efs_volume_configuration {
      file_system_id          = aws_efs_file_system.test.id
      transit_encryption      = "ENABLED"
      transit_encryption_port = 2999
      authorization_config {
        access_point_id = aws_efs_access_point.test.id
        iam             = %[2]q
      }
    }
  }
}
`, rName, useIam)
}

func testAccTaskDefinitionConfig_roleARN(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/test/"

  assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Action": "sts:AssumeRole",
			"Principal": {
				"Service": "ec2.amazonaws.com"
			},
			"Effect": "Allow",
			"Sid": ""
		}
	]
}
EOF
}

data "aws_partition" "current" {}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Action": [
				"s3:GetBucketLocation",
				"s3:ListAllMyBuckets"
			],
			"Resource": "arn:${data.aws_partition.current.partition}:s3:::*"
		}
	]
}
EOF
}

resource "aws_ecs_task_definition" "test" {
  family        = %[1]q
  task_role_arn = aws_iam_role.test.arn

  container_definitions = <<TASK_DEFINITION
[
	{
		"name": "sleep",
		"image": "busybox",
		"cpu": 10,
		"command": ["sleep","360"],
		"memory": 10,
		"essential": true
	}
]
TASK_DEFINITION

  volume {
    name = %[1]q
  }
}
`, rName)
}

func testAccTaskDefinitionConfig_ipcMode(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/test/"

  assume_role_policy = <<EOF
{
 "Version": "2012-10-17",
 "Statement": [
	 {
		 "Action": "sts:AssumeRole",
		 "Principal": {
			 "Service": "ec2.amazonaws.com"
		 },
		 "Effect": "Allow",
		 "Sid": ""
	 }
 ]
}
EOF
}

data "aws_partition" "current" {}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
 "Version": "2012-10-17",
 "Statement": [
	 {
		 "Effect": "Allow",
		 "Action": [
			 "s3:GetBucketLocation",
			 "s3:ListAllMyBuckets"
		 ],
		 "Resource": "arn:${data.aws_partition.current.partition}:s3:::*"
	 }
 ]
}
 
EOF
}

resource "aws_ecs_task_definition" "test" {
  family        = %[1]q
  task_role_arn = aws_iam_role.test.arn
  network_mode  = "bridge"
  ipc_mode      = "host"

  container_definitions = <<TASK_DEFINITION
[
 {
	 "name": "sleep",
	 "image": "busybox",
	 "cpu": 10,
	 "command": ["sleep","360"],
	 "memory": 10,
	 "essential": true
 }
]
TASK_DEFINITION

  volume {
    name = %[1]q
  }
}
`, rName)
}

func testAccTaskDefinitionConfig_pidMode(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/test/"

  assume_role_policy = <<EOF
{
 "Version": "2012-10-17",
 "Statement": [
	 {
		 "Action": "sts:AssumeRole",
		 "Principal": {
			 "Service": "ec2.amazonaws.com"
		 },
		 "Effect": "Allow",
		 "Sid": ""
	 }
 ]
}
EOF
}

data "aws_partition" "current" {}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
 "Version": "2012-10-17",
 "Statement": [
	 {
		 "Effect": "Allow",
		 "Action": [
			 "s3:GetBucketLocation",
			 "s3:ListAllMyBuckets"
		 ],
		 "Resource": "arn:${data.aws_partition.current.partition}:s3:::*"
	 }
 ]
}
 
EOF
}

resource "aws_ecs_task_definition" "test" {
  family        = %[1]q
  task_role_arn = aws_iam_role.test.arn
  network_mode  = "bridge"
  pid_mode      = "host"

  container_definitions = <<TASK_DEFINITION
[
 {
	 "name": "sleep",
	 "image": "busybox",
	 "cpu": 10,
	 "command": ["sleep","360"],
	 "memory": 10,
	 "essential": true
 }
]
TASK_DEFINITION

  volume {
    name = %[1]q
  }
}
`, rName)
}

func testAccTaskDefinitionConfig_networkMode(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/test/"

  assume_role_policy = <<EOF
{
 "Version": "2012-10-17",
 "Statement": [
	 {
		 "Action": "sts:AssumeRole",
		 "Principal": {
			 "Service": "ec2.amazonaws.com"
		 },
		 "Effect": "Allow",
		 "Sid": ""
	 }
 ]
}
EOF
}

data "aws_partition" "current" {}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
 "Version": "2012-10-17",
 "Statement": [
	 {
		 "Effect": "Allow",
		 "Action": [
			 "s3:GetBucketLocation",
			 "s3:ListAllMyBuckets"
		 ],
		 "Resource": "arn:${data.aws_partition.current.partition}:s3:::*"
	 }
 ]
}
 
EOF
}

resource "aws_ecs_task_definition" "test" {
  family        = %[1]q
  task_role_arn = aws_iam_role.test.arn
  network_mode  = "bridge"

  container_definitions = <<TASK_DEFINITION
[
 {
	 "name": "sleep",
	 "image": "busybox",
	 "cpu": 10,
	 "command": ["sleep","360"],
	 "memory": 10,
	 "essential": true
 }
]
TASK_DEFINITION

  volume {
    name = %[1]q
  }
}
`, rName)
}

func testAccTaskDefinitionConfig_service(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<TASK_DEFINITION
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 10,
    "command": ["sleep","360"],
    "memory": 10,
    "essential": true
  }
]
TASK_DEFINITION

  volume {
    name = %[1]q
  }
}
`, rName)
}

func testAccTaskDefinitionConfig_serviceModified(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<TASK_DEFINITION
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 20,
    "command": ["sleep","360"],
    "memory": 50,
    "essential": true
  }
]
TASK_DEFINITION

  volume {
    name = %[1]q
  }
}
`, rName)
}

func testAccTaskDefinitionConfig_modified(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<TASK_DEFINITION
[
	{
		"cpu": 10,
		"command": ["sleep", "10"],
		"entryPoint": ["/"],
		"environment": [
			{"name": "VARNAME", "value": "VARVAL"}
		],
		"essential": true,
		"image": "jenkins",
		"links": ["mongodb"],
		"memory": 128,
		"name": "jenkins",
		"portMappings": [
			{
				"containerPort": 80,
				"hostPort": 8080
			}
		]
	},
	{
		"cpu": 20,
		"command": ["sleep", "10"],
		"entryPoint": ["/"],
		"essential": true,
		"image": "mongodb",
		"memory": 128,
		"name": "mongodb",
		"portMappings": [
			{
				"containerPort": 28017,
				"hostPort": 28017
			}
		]
	}
]
TASK_DEFINITION

  volume {
    name      = "jenkins-home"
    host_path = "/ecs/jenkins-home"
  }
}
`, rName)
}

var testValidTaskDefinitionValidContainerDefinitions = `
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 10,
    "command": ["sleep","360"],
    "memory": 10,
    "essential": true
  }
]
`

var testValidTaskDefinitionInvalidCommandContainerDefinitions = `
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 10,
    "command": "sleep 360",
    "memory": 10,
    "essential": true
  }
]
`

func testAccTaskDefinitionConfig_tags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccTaskDefinitionConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccTaskDefinitionConfig_inferenceAccelerator(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<TASK_DEFINITION
[
	{
		"cpu": 10,
		"command": ["sleep", "10"],
		"entryPoint": ["/"],
		"environment": [
			{"name": "VARNAME", "value": "VARVAL"}
		],
		"essential": true,
		"image": "jenkins",
		"memory": 128,
		"name": "jenkins",
		"portMappings": [
			{
				"containerPort": 80,
				"hostPort": 8080
			}
		],
        "resourceRequirements":[
            {
                "type":"InferenceAccelerator",
                "value":"device_1"
            }
        ]
	}
]
TASK_DEFINITION

  inference_accelerator {
    device_name = "device_1"
    device_type = "eia1.medium"
  }
}
`, rName)
}

func testAccTaskDefinitionConfig_fsxVolume(domain, rName string) string {
	return acctest.ConfigCompose(
		testAccFSxWindowsFileSystemSubnetIds1Config(rName, domain),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_secretsmanager_secret" "test" {
  name                    = %[1]q
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = jsonencode({ username : "admin", password : aws_directory_service_directory.test.password })
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Action" : "sts:AssumeRole",
        "Principal" : {
          "Service" : "ecs-tasks.${data.aws_partition.current.dns_suffix}"
        },
        "Effect" : "Allow",
        "Sid" : ""
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_role_policy_attachment" "test2" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/SecretsManagerReadWrite"
}

resource "aws_iam_role_policy_attachment" "test3" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonFSxReadOnlyAccess"
}

resource "aws_ecs_task_definition" "test" {
  family             = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  container_definitions = <<TASK_DEFINITION
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 10,
    "command": ["sleep","360"],
    "memory": 10,
    "essential": true
  }
]
TASK_DEFINITION

  volume {
    name = %[1]q

    fsx_windows_file_server_volume_configuration {
      file_system_id = aws_fsx_windows_file_system.test.id
      root_directory = "\\data"

      authorization_config {
        credentials_parameter = aws_secretsmanager_secret_version.test.arn
        domain                = aws_directory_service_directory.test.name
      }
    }
  }

  depends_on = [
    aws_iam_role_policy_attachment.test,
    aws_iam_role_policy_attachment.test2,
    aws_iam_role_policy_attachment.test3
  ]
}
`, rName))
}

func testAccTaskDefinitionImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["arn"], nil
	}
}

func testAccFSxWindowsFileSystemBaseConfig(rName, domain string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  edition  = "Standard"
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}
`, domain),
	)
}

func testAccFSxWindowsFileSystemSubnetIds1Config(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccFSxWindowsFileSystemBaseConfig(rName, domain),
		`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8
}
`)
}
