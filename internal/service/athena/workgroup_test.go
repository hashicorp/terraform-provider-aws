// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package athena_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/athena/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfathena "github.com/hashicorp/terraform-provider-aws/internal/service/athena"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAthenaWorkGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1 types.WorkGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "athena", fmt.Sprintf("workgroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.bytes_scanned_cutoff_per_query", "0"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.enforce_workgroup_configuration", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.engine_version.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration.0.engine_version.0.effective_engine_version"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.engine_version.0.selected_engine_version", "AUTO"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.execution_role", ""),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.identity_center_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.publish_cloudwatch_metrics_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.managed_query_results_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.requester_pays_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(types.WorkGroupStateEnabled)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_configurationResultConfigurationACL(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "athena", fmt.Sprintf("workgroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.acl_configuration.#", "1"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					acctest.CheckSDKResourceDisappears(ctx, t, tfathena.ResourceWorkGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAthenaWorkGroup_bytesScannedCutoffPerQuery(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1, workgroup2 types.WorkGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_configurationBytesScannedCutoffPerQuery(rName, 12582912),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
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
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.bytes_scanned_cutoff_per_query", "10485760"),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_enforceWorkGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1, workgroup2 types.WorkGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_enforce(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
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
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.enforce_workgroup_configuration", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_configurationEngineVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1, workgroup2 types.WorkGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_configurationEngineVersion(rName, "Athena engine version 3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.engine_version.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.engine_version.0.effective_engine_version", resourceName, "configuration.0.engine_version.0.selected_engine_version"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.engine_version.0.selected_engine_version", "Athena engine version 3"),
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
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.engine_version.#", "1"),
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
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"
	iamRoleResourceName := "aws_iam_role.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_configurationExecutionRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_configurationPublishCloudWatchMetricsEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
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
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.publish_cloudwatch_metrics_enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_ResultEncryption_sseS3(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1 types.WorkGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_configurationResultConfigurationEncryptionConfigurationEncryptionOptionSseS3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.encryption_configuration.#", "1"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"
	rEncryption := string(types.EncryptionOptionSseKms)
	rEncryption2 := string(types.EncryptionOptionCseKms)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_resultEncryptionEncryptionOptionKMS(rName, rEncryption),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.encryption_configuration.#", "1"),
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
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.encryption_configuration.0.encryption_option", rEncryption2),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_SwitchEncryptionManagement(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1, workgroup2 types.WorkGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"
	rEncryption := string(types.EncryptionOptionSseKms)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_resultEncryptionEncryptionOptionKMS(rName, rEncryption),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.encryption_configuration.#", "1"),
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
				Config: testAccWorkGroupConfig_ManagedQueryResultsConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.managed_query_results_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.managed_query_results_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.managed_query_results_configuration.0.encryption_configuration.#", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccAthenaWorkGroup_Result_outputLocation(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1, workgroup2 types.WorkGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"
	rOutputLocation1 := fmt.Sprintf("%s-1", rName)
	rOutputLocation2 := fmt.Sprintf("%s-2", rName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_configurationResultConfigurationOutputLocation(rName, rOutputLocation1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", "1"),
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
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.output_location", "s3://"+rOutputLocation2+"/test/output"),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_requesterPaysEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1 types.WorkGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_configurationRequesterPaysEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "athena", fmt.Sprintf("workgroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
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
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "athena", fmt.Sprintf("workgroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"
	rOutputLocation1 := fmt.Sprintf("%s-1", rName)
	rOutputLocation2 := fmt.Sprintf("%s-2", rName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_configurationResultConfigurationOutputLocationForceDestroy(rName, rOutputLocation1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", "1"),
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
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.output_location", "s3://"+rOutputLocation2+"/test/output"),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_description(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1, workgroup2 types.WorkGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"
	rDescription := sdkacctest.RandString(20)
	rDescriptionUpdate := sdkacctest.RandString(20)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_description(rName, rDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
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
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rDescriptionUpdate),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_state(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1, workgroup2, workgroup3 types.WorkGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_state(rName, string(types.WorkGroupStateDisabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
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
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(types.WorkGroupStateEnabled)),
				),
			},
			{
				Config: testAccWorkGroupConfig_state(rName, string(types.WorkGroupStateDisabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup3),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(types.WorkGroupStateDisabled)),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_forceDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup types.WorkGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dbName := sdkacctest.RandString(5)
	queryName1 := acctest.RandomWithPrefix(t, "tf-athena-named-query-")
	queryName2 := acctest.RandomWithPrefix(t, "tf-athena-named-query-")
	resourceName := "aws_athena_workgroup.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_forceDestroy(rName, dbName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup),
					testAccCheckCreateNamedQuery(ctx, t, &workgroup, dbName, queryName1, fmt.Sprintf("SELECT * FROM %s limit 10;", rName)),
					testAccCheckCreateNamedQuery(ctx, t, &workgroup, dbName, queryName2, fmt.Sprintf("SELECT * FROM %s limit 100;", rName)),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
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
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccWorkGroupConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_ManagedQueryResultsConfiguration_enabled(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1, workGroup2 types.WorkGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_ManagedQueryResultsConfiguration_Enabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.managed_query_results_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.managed_query_results_configuration.0.enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccWorkGroupConfig_ManagedQueryResultsConfiguration_Enabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workGroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.managed_query_results_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.managed_query_results_configuration.0.enabled", acctest.CtFalse),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccAthenaWorkGroup_ManagedQueryResultsConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1, workgroup2, workgroup3 types.WorkGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_ManagedQueryResultsConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.managed_query_results_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.managed_query_results_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.managed_query_results_configuration.0.encryption_configuration.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccWorkGroupConfig_ManagedQueryResultsConfiguration_OutputLocation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.managed_query_results_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.managed_query_results_configuration.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.managed_query_results_configuration.0.encryption_configuration.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccWorkGroupConfig_ManagedQueryResultsConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup3),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.managed_query_results_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.managed_query_results_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.managed_query_results_configuration.0.encryption_configuration.#", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
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

func TestAccAthenaWorkGroup_ManagedQueryResultsConfiguration_EncryptionConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1, workgroup2, workgroup3 types.WorkGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_ManagedQueryResultsConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.managed_query_results_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.managed_query_results_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.managed_query_results_configuration.0.encryption_configuration.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccWorkGroupConfig_ManagedQueryResultsConfiguration_EncryptionConfiguration_removed(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.managed_query_results_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.managed_query_results_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.managed_query_results_configuration.0.encryption_configuration.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccWorkGroupConfig_ManagedQueryResultsConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup3),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.managed_query_results_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.managed_query_results_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.managed_query_results_configuration.0.encryption_configuration.#", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
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

func TestAccAthenaWorkGroup_ManagedQueryResultsConfiguration_conflictValidation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccWorkGroupConfig_ManagedQueryResultsConfiguration_conflictValidation(rName),
				ExpectError: regexache.MustCompile(`output_location cannot be specified when.*enabled is true`),
			},
		},
	})
}

func TestAccAthenaWorkGroup_customerContentEncryptionConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1 types.WorkGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_customerContentEncryptionConfiguration(rName, "test1"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "athena", fmt.Sprintf("workgroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.customer_content_encryption_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.customer_content_encryption_configuration.0.kms_key", "aws_kms_key.test1", names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccWorkGroupConfig_customerContentEncryptionConfiguration(rName, "test2"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "athena", fmt.Sprintf("workgroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.customer_content_encryption_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.customer_content_encryption_configuration.0.kms_key", "aws_kms_key.test2", names.AttrARN),
				),
			},
			{
				Config: testAccWorkGroupConfig_customerContentEncryptionConfigurationRemoved(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "athena", fmt.Sprintf("workgroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.customer_content_encryption_configuration.#", "0"),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_enableMinimumEncryptionConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1 types.WorkGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_enableMinimumEncryptionConfiguration(rName, true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "athena", fmt.Sprintf("workgroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.enable_minimum_encryption_configuration", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccWorkGroupConfig_enableMinimumEncryptionConfiguration(rName, false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "athena", fmt.Sprintf("workgroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.enable_minimum_encryption_configuration", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_monitoringConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var workgroup1 types.WorkGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_monitoringConfiguration(rName, true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "athena", fmt.Sprintf("workgroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.cloud_watch_logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.cloud_watch_logging_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"configuration.0.monitoring_configuration.0.cloud_watch_logging_configuration.0.log_group",
						"aws_cloudwatch_log_group.test1",
						names.AttrName,
					),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.cloud_watch_logging_configuration.0.log_stream_name_prefix", "test1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.cloud_watch_logging_configuration.0.log_type.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(
						resourceName,
						"configuration.0.monitoring_configuration.0.cloud_watch_logging_configuration.0.log_type.*",
						map[string]string{
							names.AttrKey: "SPARK_DRIVER",
							"values.#":    "2",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						resourceName,
						"configuration.0.monitoring_configuration.0.cloud_watch_logging_configuration.0.log_type.*",
						map[string]string{
							names.AttrKey: "SPARK_EXECUTOR",
							"values.#":    "0",
						},
					),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.managed_logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.managed_logging_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"configuration.0.monitoring_configuration.0.managed_logging_configuration.0.kms_key",
						"aws_kms_key.test1",
						names.AttrARN,
					),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.s3_logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.s3_logging_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"configuration.0.monitoring_configuration.0.s3_logging_configuration.0.kms_key",
						"aws_kms_key.test2",
						names.AttrARN,
					),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.s3_logging_configuration.0.log_location", fmt.Sprintf("s3://%s/athena_spark1/", rName)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccWorkGroupConfig_monitoringConfigurationUpdated(rName, true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "athena", fmt.Sprintf("workgroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.cloud_watch_logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.cloud_watch_logging_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"configuration.0.monitoring_configuration.0.cloud_watch_logging_configuration.0.log_group",
						"aws_cloudwatch_log_group.test2",
						names.AttrName,
					),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.cloud_watch_logging_configuration.0.log_stream_name_prefix", "test2"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.cloud_watch_logging_configuration.0.log_type.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(
						resourceName,
						"configuration.0.monitoring_configuration.0.cloud_watch_logging_configuration.0.log_type.*",
						map[string]string{
							names.AttrKey: "SPARK_DRIVER",
							"values.#":    "1",
							"values.0":    "STDOUT",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						resourceName,
						"configuration.0.monitoring_configuration.0.cloud_watch_logging_configuration.0.log_type.*",
						map[string]string{
							names.AttrKey: "SPARK_EXECUTOR",
							"values.#":    "0",
						},
					),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.managed_logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.managed_logging_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"configuration.0.monitoring_configuration.0.managed_logging_configuration.0.kms_key",
						"aws_kms_key.test2",
						names.AttrARN,
					),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.s3_logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.s3_logging_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"configuration.0.monitoring_configuration.0.s3_logging_configuration.0.kms_key",
						"aws_kms_key.test1",
						names.AttrARN,
					),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.s3_logging_configuration.0.log_location", fmt.Sprintf("s3://%s/athena_spark2/", rName)),
				),
			},
			{
				Config: testAccWorkGroupConfig_monitoringConfigurationUpdated(rName, false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "athena", fmt.Sprintf("workgroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.cloud_watch_logging_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.managed_logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.managed_logging_configuration.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.s3_logging_configuration.#", "0"),
				),
			},
			{
				Config: testAccWorkGroupConfig_monitoringConfigurationUpdated(rName, true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "athena", fmt.Sprintf("workgroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.cloud_watch_logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.cloud_watch_logging_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"configuration.0.monitoring_configuration.0.cloud_watch_logging_configuration.0.log_group",
						"aws_cloudwatch_log_group.test2",
						names.AttrName,
					),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.cloud_watch_logging_configuration.0.log_stream_name_prefix", "test2"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.cloud_watch_logging_configuration.0.log_type.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(
						resourceName,
						"configuration.0.monitoring_configuration.0.cloud_watch_logging_configuration.0.log_type.*",
						map[string]string{
							names.AttrKey: "SPARK_DRIVER",
							"values.#":    "1",
							"values.0":    "STDOUT",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						resourceName,
						"configuration.0.monitoring_configuration.0.cloud_watch_logging_configuration.0.log_type.*",
						map[string]string{
							names.AttrKey: "SPARK_EXECUTOR",
							"values.#":    "0",
						},
					),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.managed_logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.managed_logging_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"configuration.0.monitoring_configuration.0.managed_logging_configuration.0.kms_key",
						"aws_kms_key.test2",
						names.AttrARN,
					),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.s3_logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.s3_logging_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"configuration.0.monitoring_configuration.0.s3_logging_configuration.0.kms_key",
						"aws_kms_key.test1",
						names.AttrARN,
					),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.s3_logging_configuration.0.log_location", fmt.Sprintf("s3://%s/athena_spark2/", rName)),
				),
			},
			{
				Config: testAccWorkGroupConfig_monitoringConfigurationRemoved(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkGroupExists(ctx, t, resourceName, &workgroup1),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "athena", fmt.Sprintf("workgroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.cloud_watch_logging_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.managed_logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.managed_logging_configuration.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.monitoring_configuration.0.s3_logging_configuration.#", "0"),
				),
			},
		},
	})
}

func testAccCheckCreateNamedQuery(ctx context.Context, t *testing.T, workGroup *types.WorkGroup, databaseName, queryName, query string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AthenaClient(ctx)

		input := &athena.CreateNamedQueryInput{
			Name:        aws.String(queryName),
			WorkGroup:   workGroup.Name,
			Database:    aws.String(databaseName),
			QueryString: aws.String(query),
			Description: aws.String("tf test"),
		}

		if _, err := conn.CreateNamedQuery(ctx, input); err != nil {
			return fmt.Errorf("error creating Named Query (%s) on Workgroup (%s): %w", queryName, aws.ToString(workGroup.Name), err)
		}

		return nil
	}
}

func testAccCheckWorkGroupDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AthenaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_athena_workgroup" {
				continue
			}

			_, err := tfathena.FindWorkGroupByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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

func testAccCheckWorkGroupExists(ctx context.Context, t *testing.T, n string, v *types.WorkGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AthenaClient(ctx)

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
  enable_key_rotation     = true
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

func testAccWorkGroupConfig_ManagedQueryResultsConfiguration_Enabled(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = %[1]q

  configuration {
    managed_query_results_configuration {
      enabled = %[2]t
    }
  }
}
`, rName, enabled)
}

func testAccWorkGroupConfig_ManagedQueryResultsConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  description             = "Terraform Acceptance Testing"
  enable_key_rotation     = true
}

resource "aws_athena_workgroup" "test" {
  name = %[1]q

  configuration {
    managed_query_results_configuration {
      enabled = true
      encryption_configuration {
        kms_key = aws_kms_key.test.arn
      }
    }
  }
}
`, rName)
}

func testAccWorkGroupConfig_ManagedQueryResultsConfiguration_EncryptionConfiguration_removed(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  description             = "Terraform Acceptance Testing"
  enable_key_rotation     = true
}

resource "aws_athena_workgroup" "test" {
  name = %[1]q

  configuration {
    managed_query_results_configuration {
      enabled = %[2]t
    }
  }
}
`, rName, enabled)
}

func testAccWorkGroupConfig_ManagedQueryResultsConfiguration_OutputLocation(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  description             = "Terraform Acceptance Testing"
  enable_key_rotation     = true
}

resource "aws_athena_workgroup" "test" {
  name = %[1]q

  configuration {
    managed_query_results_configuration {
      enabled = false
    }

    result_configuration {
      output_location = "s3://${aws_s3_bucket.test.bucket}/output/"
    }
  }
}
`, rName)
}

func testAccWorkGroupConfig_ManagedQueryResultsConfiguration_conflictValidation(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_athena_workgroup" "test" {
  name = %[1]q

  configuration {
    managed_query_results_configuration {
      enabled = true
    }

    result_configuration {
      output_location = "s3://${aws_s3_bucket.test.bucket}/"
    }
  }
}
`, rName)
}

func testAccWorkGroupConfig_customerContentEncryptionConfigurationBase(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["athena.amazonaws.com"]
    }
    effect = "Allow"
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_kms_key" "test1" {
  deletion_window_in_days = 7
  description             = "Terraform Acceptance Testing"
  enable_key_rotation     = true
}

resource "aws_kms_key" "test2" {
  deletion_window_in_days = 7
  description             = "Terraform Acceptance Testing"
  enable_key_rotation     = true
}
`, rName)
}

func testAccWorkGroupConfig_customerContentEncryptionConfiguration(rName, kmsKeyIdentifier string) string {
	return acctest.ConfigCompose(
		testAccWorkGroupConfig_customerContentEncryptionConfigurationBase(rName),
		fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = %[1]q
  configuration {
    customer_content_encryption_configuration {
      kms_key = aws_kms_key.%[2]s.arn
    }
    engine_version {
      selected_engine_version = "PySpark engine version 3"
    }
    result_configuration {
      output_location = "s3://${aws_s3_bucket.test.bucket}/athena_spark/"
    }
    execution_role = aws_iam_role.test.arn
  }
}
`, rName, kmsKeyIdentifier))
}

func testAccWorkGroupConfig_customerContentEncryptionConfigurationRemoved(rName string) string {
	return acctest.ConfigCompose(
		testAccWorkGroupConfig_customerContentEncryptionConfigurationBase(rName),
		fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = %[1]q
  configuration {
    engine_version {
      selected_engine_version = "PySpark engine version 3"
    }
    result_configuration {
      output_location = "s3://${aws_s3_bucket.test.bucket}/athena_spark/"
    }
    execution_role = aws_iam_role.test.arn
  }
}
`, rName))
}

func testAccWorkGroupConfig_enableMinimumEncryptionConfiguration(rName string, enableMinimumEncryptionConfiguration bool) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = %[1]q
  configuration {
    enable_minimum_encryption_configuration = %[2]t
  }
}
`, rName, enableMinimumEncryptionConfiguration)
}

func testAccWorkGroupConfig_monitoringConfigurationBase(rName string) string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["athena.amazonaws.com"]
    }
    effect = "Allow"
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_cloudwatch_log_group" "test1" {
  name = "%[1]s-1"
}

resource "aws_cloudwatch_log_group" "test2" {
  name = "%[1]s-2"
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_kms_key" "test1" {
  deletion_window_in_days = 7
  description             = "Terraform Acceptance Testing"
  enable_key_rotation     = true
}

resource "aws_kms_key" "test2" {
  deletion_window_in_days = 7
  description             = "Terraform Acceptance Testing"
  enable_key_rotation     = true
}
`, rName))
}

func testAccWorkGroupConfig_monitoringConfiguration(rName string, enabled bool) string {
	return acctest.ConfigCompose(
		testAccWorkGroupConfig_monitoringConfigurationBase(rName),
		fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = %[1]q
  configuration {
    engine_version {
      selected_engine_version = "Apache Spark version 3.5"
    }
    monitoring_configuration {
      cloud_watch_logging_configuration {
        enabled                = %[2]t
        log_group              = aws_cloudwatch_log_group.test1.name
        log_stream_name_prefix = "test1"
        log_type {
          key    = "SPARK_DRIVER"
          values = ["STDOUT", "STDERR"]
        }
        log_type {
          key    = "SPARK_EXECUTOR"
          values = []
        }
      }
      managed_logging_configuration {
        enabled = %[2]t
        kms_key = aws_kms_key.test1.arn
      }
      s3_logging_configuration {
        enabled      = %[2]t
        kms_key      = aws_kms_key.test2.arn
        log_location = "s3://${aws_s3_bucket.test.bucket}/athena_spark1/"
      }
    }
    execution_role = aws_iam_role.test.arn
  }
}
`, rName, enabled))
}

func testAccWorkGroupConfig_monitoringConfigurationUpdated(rName string, enabled bool) string {
	return acctest.ConfigCompose(
		testAccWorkGroupConfig_monitoringConfigurationBase(rName),
		fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = %[1]q
  configuration {
    engine_version {
      selected_engine_version = "Apache Spark version 3.5"
    }
    monitoring_configuration {
      cloud_watch_logging_configuration {
        enabled                = %[2]t
        log_group              = aws_cloudwatch_log_group.test2.name
        log_stream_name_prefix = "test2"
        log_type {
          key    = "SPARK_DRIVER"
          values = ["STDOUT"]
        }
        log_type {
          key    = "SPARK_EXECUTOR"
          values = []
        }
      }
      managed_logging_configuration {
        enabled = %[2]t
        kms_key = aws_kms_key.test2.arn
      }
      s3_logging_configuration {
        enabled      = %[2]t
        kms_key      = aws_kms_key.test1.arn
        log_location = "s3://${aws_s3_bucket.test.bucket}/athena_spark2/"
      }
    }
    execution_role = aws_iam_role.test.arn
  }
}
`, rName, enabled))
}

func testAccWorkGroupConfig_monitoringConfigurationRemoved(rName string) string {
	return acctest.ConfigCompose(
		testAccWorkGroupConfig_monitoringConfigurationBase(rName),
		fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = %[1]q
  configuration {
    engine_version {
      selected_engine_version = "Apache Spark version 3.5"
    }
    execution_role = aws_iam_role.test.arn
  }
}
`, rName))
}
