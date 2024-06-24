// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSageMakerNotebookInstance_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotebookInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "accelerator_types.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "additional_code_repositories.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_code_repository", ""),
					resource.TestCheckResourceAttr(resourceName, "direct_internet_access", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "instance_metadata_service_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_metadata_service_configuration.0.minimum_instance_metadata_service_version", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, "ml.t2.medium"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "platform_identifier", "notebook-al1-v1"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "root_access", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrURL),
					resource.TestCheckResourceAttr(resourceName, names.AttrVolumeSize, "5"),
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

func TestAccSageMakerNotebookInstance_imds(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotebookInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceConfig_imds(rName, acctest.Ct2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "instance_metadata_service_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_metadata_service_configuration.0.minimum_instance_metadata_service_version", acctest.Ct2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotebookInstanceConfig_imds(rName, acctest.Ct1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "instance_metadata_service_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_metadata_service_configuration.0.minimum_instance_metadata_service_version", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccSageMakerNotebookInstance_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotebookInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, "ml.t2.medium"),
				),
			},
			{
				Config: testAccNotebookInstanceConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, "ml.m4.xlarge"),
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

func TestAccSageMakerNotebookInstance_volumeSize(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook1, notebook2, notebook3 sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var resourceName = "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotebookInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook1),
					resource.TestCheckResourceAttr(resourceName, names.AttrVolumeSize, "5"),
				),
			},
			{
				Config: testAccNotebookInstanceConfig_volume(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook2),
					resource.TestCheckResourceAttr(resourceName, names.AttrVolumeSize, "8"),
					testAccCheckNotebookInstanceNotRecreated(&notebook1, &notebook2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotebookInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook3),
					resource.TestCheckResourceAttr(resourceName, names.AttrVolumeSize, "5"),
					testAccCheckNotebookInstanceRecreated(&notebook2, &notebook3),
				),
			},
		},
	})
}

func TestAccSageMakerNotebookInstance_lifecycleName(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_notebook_instance.test"
	sagemakerLifecycleConfigResourceName := "aws_sagemaker_notebook_instance_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotebookInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceConfig_lifecycleName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttrPair(resourceName, "lifecycle_config_name", sagemakerLifecycleConfigResourceName, names.AttrName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotebookInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_config_name", ""),
				),
			},
			{
				Config: testAccNotebookInstanceConfig_lifecycleName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttrPair(resourceName, "lifecycle_config_name", sagemakerLifecycleConfigResourceName, names.AttrName),
				),
			},
		},
	})
}

func TestAccSageMakerNotebookInstance_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotebookInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotebookInstanceConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccNotebookInstanceConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccSageMakerNotebookInstance_kms(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotebookInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceConfig_kms(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, "aws_kms_key.test", names.AttrID),
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

func TestAccSageMakerNotebookInstance_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotebookInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceNotebookInstance(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSageMakerNotebookInstance_Root_access(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotebookInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceConfig_rootAccess(rName, "Disabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "root_access", "Disabled"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotebookInstanceConfig_rootAccess(rName, "Enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "root_access", "Enabled"),
				),
			},
		},
	})
}

func TestAccSageMakerNotebookInstance_Platform_identifier(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotebookInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceConfig_platformIdentifier(rName, "notebook-al2-v1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "platform_identifier", "notebook-al2-v1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotebookInstanceConfig_platformIdentifier(rName, "notebook-al1-v1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "platform_identifier", "notebook-al1-v1"),
				),
			},
		},
	})
}

func TestAccSageMakerNotebookInstance_DirectInternet_access(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotebookInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceConfig_directInternetAccess(rName, "Disabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "direct_internet_access", "Disabled"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSubnetID, "aws_subnet.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, names.AttrNetworkInterfaceID, regexache.MustCompile("eni-.*")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotebookInstanceConfig_directInternetAccess(rName, "Enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "direct_internet_access", "Enabled"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSubnetID, "aws_subnet.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, names.AttrNetworkInterfaceID, regexache.MustCompile("eni-.*")),
				),
			},
		},
	})
}

func TestAccSageMakerNotebookInstance_DefaultCode_repository(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var resourceName = "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotebookInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceConfig_defaultCodeRepository(rName, "https://github.com/hashicorp/terraform-provider-aws.git"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "default_code_repository", "https://github.com/hashicorp/terraform-provider-aws.git"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotebookInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "default_code_repository", ""),
				),
			},
			{
				Config: testAccNotebookInstanceConfig_defaultCodeRepository(rName, "https://github.com/hashicorp/terraform-provider-aws.git"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "default_code_repository", "https://github.com/hashicorp/terraform-provider-aws.git"),
				),
			},
		},
	})
}

func TestAccSageMakerNotebookInstance_AdditionalCode_repositories(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var resourceName = "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotebookInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceConfig_additionalCodeRepository1(rName, "https://github.com/hashicorp/terraform-provider-aws.git"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "additional_code_repositories.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "additional_code_repositories.*", "https://github.com/hashicorp/terraform-provider-aws.git"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotebookInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "additional_code_repositories.#", acctest.Ct0),
				),
			},
			{
				Config: testAccNotebookInstanceConfig_additionalCodeRepository2(rName, "https://github.com/hashicorp/terraform-provider-aws.git", "https://github.com/hashicorp/terraform.git"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "additional_code_repositories.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "additional_code_repositories.*", "https://github.com/hashicorp/terraform-provider-aws.git"),
					resource.TestCheckTypeSetElemAttr(resourceName, "additional_code_repositories.*", "https://github.com/hashicorp/terraform.git"),
				),
			},
			{
				Config: testAccNotebookInstanceConfig_additionalCodeRepository1(rName, "https://github.com/hashicorp/terraform-provider-aws.git"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "additional_code_repositories.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "additional_code_repositories.*", "https://github.com/hashicorp/terraform-provider-aws.git"),
				),
			},
		},
	})
}

func TestAccSageMakerNotebookInstance_DefaultCodeRepository_sageMakerRepo(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var resourceName = "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotebookInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceConfig_defaultCodeRepositoryRepo(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttrPair(resourceName, "default_code_repository", "aws_sagemaker_code_repository.test", "code_repository_name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotebookInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "default_code_repository", ""),
				),
			},
			{
				Config: testAccNotebookInstanceConfig_defaultCodeRepositoryRepo(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttrPair(resourceName, "default_code_repository", "aws_sagemaker_code_repository.test", "code_repository_name")),
			},
		},
	})
}

func TestAccSageMakerNotebookInstance_acceleratorTypes(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var resourceName = "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotebookInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceConfig_acceleratorType(rName, "ml.eia2.medium"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "accelerator_types.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "accelerator_types.*", "ml.eia2.medium"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotebookInstanceConfig_acceleratorType(rName, "ml.eia2.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "accelerator_types.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "accelerator_types.*", "ml.eia2.large"),
				),
			},
			{
				Config: testAccNotebookInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceExists(ctx, resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "accelerator_types.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccCheckNotebookInstanceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_notebook_instance" {
				continue
			}

			_, err := tfsagemaker.FindNotebookInstanceByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SageMaker Notebook Instance %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckNotebookInstanceExists(ctx context.Context, n string, v *sagemaker.DescribeNotebookInstanceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SageMaker Notebook Instance ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)

		output, err := tfsagemaker.FindNotebookInstanceByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckNotebookInstanceNotRecreated(i, j *sagemaker.DescribeNotebookInstanceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.CreationTime).Equal(aws.TimeValue(j.CreationTime)) {
			return errors.New("SageMaker Notebook Instance was recreated")
		}

		return nil
	}
}

func testAccCheckNotebookInstanceRecreated(i, j *sagemaker.DescribeNotebookInstanceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreationTime).Equal(aws.TimeValue(j.CreationTime)) {
			return errors.New("SageMaker Notebook Instance was not recreated")
		}

		return nil
	}
}

func testAccNotebookInstanceBaseConfig(rName string) string {
	return fmt.Sprintf(`
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
`, rName)
}

func testAccNotebookInstanceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccNotebookInstanceBaseConfig(rName), fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name          = %[1]q
  role_arn      = aws_iam_role.test.arn
  instance_type = "ml.t2.medium"
}
`, rName))
}

func testAccNotebookInstanceConfig_update(rName string) string {
	return acctest.ConfigCompose(testAccNotebookInstanceBaseConfig(rName), fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name          = %[1]q
  role_arn      = aws_iam_role.test.arn
  instance_type = "ml.m4.xlarge"
}
`, rName))
}

func testAccNotebookInstanceConfig_lifecycleName(rName string) string {
	return acctest.ConfigCompose(testAccNotebookInstanceBaseConfig(rName), fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance_lifecycle_configuration" "test" {
  name = %[1]q
}

resource "aws_sagemaker_notebook_instance" "test" {
  instance_type         = "ml.t2.medium"
  lifecycle_config_name = aws_sagemaker_notebook_instance_lifecycle_configuration.test.name
  name                  = %[1]q
  role_arn              = aws_iam_role.test.arn
}
`, rName))
}

func testAccNotebookInstanceConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccNotebookInstanceBaseConfig(rName), fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name          = %[1]q
  role_arn      = aws_iam_role.test.arn
  instance_type = "ml.t2.medium"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccNotebookInstanceConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccNotebookInstanceBaseConfig(rName), fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name          = %[1]q
  role_arn      = aws_iam_role.test.arn
  instance_type = "ml.t2.medium"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccNotebookInstanceConfig_rootAccess(rName string, rootAccess string) string {
	return acctest.ConfigCompose(testAccNotebookInstanceBaseConfig(rName), fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name          = %[1]q
  role_arn      = aws_iam_role.test.arn
  instance_type = "ml.t2.medium"
  root_access   = %[2]q
}
`, rName, rootAccess))
}

func testAccNotebookInstanceConfig_platformIdentifier(rName string, platformIdentifier string) string {
	return acctest.ConfigCompose(testAccNotebookInstanceBaseConfig(rName), fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name                = %[1]q
  role_arn            = aws_iam_role.test.arn
  instance_type       = "ml.t2.medium"
  platform_identifier = %[2]q
}
`, rName, platformIdentifier))
}
func testAccNotebookInstanceConfig_directInternetAccess(rName string, directInternetAccess string) string {
	return acctest.ConfigCompose(testAccNotebookInstanceBaseConfig(rName), fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name                   = %[1]q
  role_arn               = aws_iam_role.test.arn
  instance_type          = "ml.t2.medium"
  security_groups        = [aws_security_group.test.id]
  subnet_id              = aws_subnet.test.id
  direct_internet_access = %[2]q
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.0.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, directInternetAccess))
}

func testAccNotebookInstanceConfig_volume(rName string) string {
	return acctest.ConfigCompose(testAccNotebookInstanceBaseConfig(rName), fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name          = %[1]q
  role_arn      = aws_iam_role.test.arn
  instance_type = "ml.t2.medium"
  volume_size   = 8
}
  `, rName))
}

func testAccNotebookInstanceConfig_defaultCodeRepository(rName string, defaultCodeRepository string) string {
	return acctest.ConfigCompose(testAccNotebookInstanceBaseConfig(rName), fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name                    = %[1]q
  role_arn                = aws_iam_role.test.arn
  instance_type           = "ml.t2.medium"
  default_code_repository = %[2]q
}
`, rName, defaultCodeRepository))
}

func testAccNotebookInstanceConfig_additionalCodeRepository1(rName, repo1 string) string {
	return acctest.ConfigCompose(testAccNotebookInstanceBaseConfig(rName), fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name                         = %[1]q
  role_arn                     = aws_iam_role.test.arn
  instance_type                = "ml.t2.medium"
  additional_code_repositories = ["%[2]s"]
}
`, rName, repo1))
}

func testAccNotebookInstanceConfig_additionalCodeRepository2(rName, repo1, repo2 string) string {
	return acctest.ConfigCompose(testAccNotebookInstanceBaseConfig(rName), fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name                         = %[1]q
  role_arn                     = aws_iam_role.test.arn
  instance_type                = "ml.t2.medium"
  additional_code_repositories = ["%[2]s", "%[3]s"]
}
`, rName, repo1, repo2))
}

func testAccNotebookInstanceConfig_defaultCodeRepositoryRepo(rName string) string {
	return acctest.ConfigCompose(testAccNotebookInstanceBaseConfig(rName), fmt.Sprintf(`
resource "aws_sagemaker_code_repository" "test" {
  code_repository_name = %[1]q

  git_config {
    repository_url = "https://github.com/hashicorp/terraform-provider-aws.git"
  }
}

resource "aws_sagemaker_notebook_instance" "test" {
  name                    = %[1]q
  role_arn                = aws_iam_role.test.arn
  instance_type           = "ml.t2.medium"
  default_code_repository = aws_sagemaker_code_repository.test.code_repository_name
}
`, rName))
}

func testAccNotebookInstanceConfig_kms(rName string) string {
	return acctest.ConfigCompose(testAccNotebookInstanceBaseConfig(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_sagemaker_notebook_instance" "test" {
  name          = %[1]q
  role_arn      = aws_iam_role.test.arn
  instance_type = "ml.t2.medium"
  kms_key_id    = aws_kms_key.test.id
}
`, rName))
}

func testAccNotebookInstanceConfig_imds(rName, version string) string {
	return acctest.ConfigCompose(testAccNotebookInstanceBaseConfig(rName), fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name          = %[1]q
  role_arn      = aws_iam_role.test.arn
  instance_type = "ml.t2.medium"
  instance_metadata_service_configuration {
    minimum_instance_metadata_service_version = %[2]q
  }
}
`, rName, version))
}

func testAccNotebookInstanceConfig_acceleratorType(rName, acceleratorType string) string {
	return acctest.ConfigCompose(testAccNotebookInstanceBaseConfig(rName), fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name              = %[1]q
  role_arn          = aws_iam_role.test.arn
  instance_type     = "ml.t2.xlarge"
  accelerator_types = [%[2]q]
}
  `, rName, acceleratorType))
}
