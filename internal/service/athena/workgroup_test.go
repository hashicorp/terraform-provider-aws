// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package athena_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/athena/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfathena "github.com/hashicorp/terraform-provider-aws/internal/service/athena"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAthenaWorkGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1 types.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup1),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "athena", fmt.Sprintf("workgroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.bytes_scanned_cutoff_per_query", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.enforce_workgroup_configuration", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.engine_version.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "configuration.0.engine_version.0.effective_engine_version"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.engine_version.0.selected_engine_version", "AUTO"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.execution_role", ""),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.publish_cloudwatch_metrics_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.requester_pays_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(types.WorkGroupStateEnabled)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccAthenaWorkGroup_aclConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1 types.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_configurationResultConfigurationACL(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup1),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "athena", fmt.Sprintf("workgroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.acl_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.acl_configuration.0.s3_acl_option", "BUCKET_OWNER_FULL_CONTROL"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccAthenaWorkGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1 types.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup1),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfathena.ResourceWorkGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAthenaWorkGroup_bytesScannedCutoffPerQuery(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1, workgroup2 types.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_configurationBytesScannedCutoffPerQuery(rName, 12582912),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.bytes_scanned_cutoff_per_query", "12582912"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccWorkGroupConfig_configurationBytesScannedCutoffPerQuery(rName, 10485760),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.bytes_scanned_cutoff_per_query", "10485760"),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_enforceWorkGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1, workgroup2 types.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_enforce(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.enforce_workgroup_configuration", acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccWorkGroupConfig_enforce(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.enforce_workgroup_configuration", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_configurationEngineVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1, workgroup2 types.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_configurationEngineVersion(rName, "Athena engine version 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.engine_version.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.engine_version.0.effective_engine_version", resourceName, "configuration.0.engine_version.0.selected_engine_version"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.engine_version.0.selected_engine_version", "Athena engine version 2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccWorkGroupConfig_configurationEngineVersion(rName, "AUTO"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.engine_version.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "configuration.0.engine_version.0.effective_engine_version"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.engine_version.0.selected_engine_version", "AUTO"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccWorkGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup1),
					resource.TestCheckResourceAttrSet(resourceName, "configuration.0.engine_version.0.effective_engine_version"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.engine_version.0.selected_engine_version", "AUTO"),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_configurationExecutionRole(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1 types.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"
	iamRoleResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_configurationExecutionRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.execution_role", iamRoleResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccAthenaWorkGroup_publishCloudWatchMetricsEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1, workgroup2 types.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_configurationPublishCloudWatchMetricsEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.publish_cloudwatch_metrics_enabled", acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccWorkGroupConfig_configurationPublishCloudWatchMetricsEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.publish_cloudwatch_metrics_enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_ResultEncryption_sseS3(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1 types.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_configurationResultConfigurationEncryptionConfigurationEncryptionOptionSseS3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.encryption_configuration.0.encryption_option", string(types.EncryptionOptionSseS3)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccAthenaWorkGroup_ResultEncryption_kms(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1, workgroup2 types.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"
	rEncryption := string(types.EncryptionOptionSseKms)
	rEncryption2 := string(types.EncryptionOptionCseKms)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_resultEncryptionEncryptionOptionKMS(rName, rEncryption),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.encryption_configuration.0.encryption_option", rEncryption),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccWorkGroupConfig_resultEncryptionEncryptionOptionKMS(rName, rEncryption2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.encryption_configuration.0.encryption_option", rEncryption2),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_Result_outputLocation(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1, workgroup2 types.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"
	rOutputLocation1 := fmt.Sprintf("%s-1", rName)
	rOutputLocation2 := fmt.Sprintf("%s-2", rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_configurationResultConfigurationOutputLocation(rName, rOutputLocation1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.output_location", "s3://"+rOutputLocation1+"/test/output"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccWorkGroupConfig_configurationResultConfigurationOutputLocation(rName, rOutputLocation2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.output_location", "s3://"+rOutputLocation2+"/test/output"),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_requesterPaysEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1 types.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_configurationRequesterPaysEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup1),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "athena", fmt.Sprintf("workgroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.requester_pays_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccWorkGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup1),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "athena", fmt.Sprintf("workgroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.requester_pays_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

//

func TestAccAthenaWorkGroup_ResultOutputLocation_forceDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1, workgroup2 types.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"
	rOutputLocation1 := fmt.Sprintf("%s-1", rName)
	rOutputLocation2 := fmt.Sprintf("%s-2", rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_configurationResultConfigurationOutputLocationForceDestroy(rName, rOutputLocation1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.output_location", "s3://"+rOutputLocation1+"/test/output"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccWorkGroupConfig_configurationResultConfigurationOutputLocationForceDestroy(rName, rOutputLocation2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.output_location", "s3://"+rOutputLocation2+"/test/output"),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_description(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1, workgroup2 types.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"
	rDescription := sdkacctest.RandString(20)
	rDescriptionUpdate := sdkacctest.RandString(20)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_description(rName, rDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rDescription),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccWorkGroupConfig_description(rName, rDescriptionUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rDescriptionUpdate),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_state(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1, workgroup2, workgroup3 types.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_state(rName, string(types.WorkGroupStateDisabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(types.WorkGroupStateDisabled)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccWorkGroupConfig_state(rName, string(types.WorkGroupStateEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(types.WorkGroupStateEnabled)),
				),
			},
			{
				Config: testAccWorkGroupConfig_state(rName, string(types.WorkGroupStateDisabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup3),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(types.WorkGroupStateDisabled)),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_forceDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup types.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbName := sdkacctest.RandString(5)
	queryName1 := sdkacctest.RandomWithPrefix("tf-athena-named-query-")
	queryName2 := sdkacctest.RandomWithPrefix("tf-athena-named-query-")
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_forceDestroy(rName, dbName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup),
					testAccCheckCreateNamedQuery(ctx, &workgroup, dbName, queryName1, fmt.Sprintf("SELECT * FROM %s limit 10;", rName)),
					testAccCheckCreateNamedQuery(ctx, &workgroup, dbName, queryName2, fmt.Sprintf("SELECT * FROM %s limit 100;", rName)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccAthenaWorkGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1, workgroup2, workgroup3 types.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccWorkGroupConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccWorkGroupConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, resourceName, &workgroup3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckCreateNamedQuery(ctx context.Context, workGroup *types.WorkGroup, databaseName, queryName, query string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AthenaClient(ctx)

		input := &athena.CreateNamedQueryInput{
			Name:        aws.String(queryName),
			WorkGroup:   workGroup.Name,
			Database:    aws.String(databaseName),
			QueryString: aws.String(query),
			Description: aws.String("tf test"),
		}

		if _, err := conn.CreateNamedQuery(ctx, input); err != nil {
			return fmt.Errorf("error creating Named Query (%s) on Workgroup (%s): %s", queryName, aws.ToString(workGroup.Name), err)
		}

		return nil
	}
}

func testAccCheckWorkGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AthenaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_athena_workgroup" {
				continue
			}

			_, err := tfathena.FindWorkGroupByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Athena WorkGroup %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckWorkGroupExists(ctx context.Context, n string, v *types.WorkGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AthenaClient(ctx)

		output, err := tfathena.FindWorkGroupByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccWorkGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = %[1]q
}
`, rName)
}

func testAccWorkGroupConfig_description(rName string, description string) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  description = %[2]q
  name        = %[1]q
}
`, rName, description)
}

func testAccWorkGroupConfig_configurationBytesScannedCutoffPerQuery(rName string, bytesScannedCutoffPerQuery int) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = %[1]q

  configuration {
    bytes_scanned_cutoff_per_query = %[2]d
  }
}
`, rName, bytesScannedCutoffPerQuery)
}

func testAccWorkGroupConfig_enforce(rName string, enforceWorkgroupConfiguration bool) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = %[1]q

  configuration {
    enforce_workgroup_configuration = %[2]t
  }
}
`, rName, enforceWorkgroupConfiguration)
}

func testAccWorkGroupConfig_configurationEngineVersion(rName, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = %[1]q

  configuration {
    engine_version {
      selected_engine_version = %[2]q
    }
  }
}
`, rName, engineVersion)
}

func testAccWorkGroupConfig_configurationExecutionRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = <<EOF
{
 "Version": "2012-10-17",
 "Statement": [
  {
   "Action": "sts:AssumeRole",
   "Principal": {
     "Service": "athena.amazonaws.com"
   },
   "Effect": "Allow",
   "Sid": ""
  }
 ]
}
EOF
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_athena_workgroup" "test" {
  name = %[1]q

  configuration {
    execution_role                     = aws_iam_role.test.arn
    enforce_workgroup_configuration    = false
    publish_cloudwatch_metrics_enabled = false

    engine_version {
      selected_engine_version = "PySpark engine version 3"
    }

    result_configuration {
      output_location = "s3://${aws_s3_bucket.test.id}/logs/athena_spark/"
    }
  }
}
`, rName)
}

func testAccWorkGroupConfig_configurationPublishCloudWatchMetricsEnabled(rName string, publishCloudwatchMetricsEnabled bool) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = %[1]q

  configuration {
    publish_cloudwatch_metrics_enabled = %[2]t
  }
}
`, rName, publishCloudwatchMetricsEnabled)
}

func testAccWorkGroupConfig_configurationResultConfigurationOutputLocation(rName string, bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_athena_workgroup" "test" {
  name = %[1]q

  configuration {
    result_configuration {
      output_location = "s3://${aws_s3_bucket.test.bucket}/test/output"
    }
  }
}
`, rName, bucketName)
}

func testAccWorkGroupConfig_configurationRequesterPaysEnabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = %[1]q

  configuration {
    requester_pays_enabled = true
  }
}
`, rName)
}

func testAccWorkGroupConfig_configurationResultConfigurationOutputLocationForceDestroy(rName string, bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_athena_workgroup" "test" {
  name = %[1]q

  force_destroy = true

  configuration {
    result_configuration {
      output_location = "s3://${aws_s3_bucket.test.bucket}/test/output"
    }
  }
}
`, rName, bucketName)
}

func testAccWorkGroupConfig_configurationResultConfigurationACL(rName string) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = %[1]q

  configuration {
    result_configuration {
      acl_configuration {
        s3_acl_option = "BUCKET_OWNER_FULL_CONTROL"
      }
    }
  }
}
`, rName)
}

func testAccWorkGroupConfig_configurationResultConfigurationEncryptionConfigurationEncryptionOptionSseS3(rName string) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = %[1]q

  configuration {
    result_configuration {
      encryption_configuration {
        encryption_option = "SSE_S3"
      }
    }
  }
}
`, rName)
}

func testAccWorkGroupConfig_resultEncryptionEncryptionOptionKMS(rName, encryptionOption string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  description             = "Terraform Acceptance Testing"
}

resource "aws_athena_workgroup" "test" {
  name = %[1]q

  configuration {
    result_configuration {
      encryption_configuration {
        encryption_option = %[2]q
        kms_key_arn       = aws_kms_key.test.arn
      }
    }
  }
}
`, rName, encryptionOption)
}

func testAccWorkGroupConfig_state(rName, state string) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name  = %[1]q
  state = %[2]q
}
`, rName, state)
}

func testAccWorkGroupConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccWorkGroupConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccWorkGroupConfig_forceDestroy(rName, dbName string) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name          = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_athena_database" "test" {
  name          = %[2]q
  bucket        = aws_s3_bucket.test.bucket
  force_destroy = true
}
`, rName, dbName)
}
