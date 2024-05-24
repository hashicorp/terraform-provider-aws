// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfssm "github.com/hashicorp/terraform-provider-aws/internal/service/ssm"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSMAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ssm", regexache.MustCompile(`association/.+`)),
					resource.TestCheckResourceAttr(resourceName, "apply_only_at_cron_interval", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_instance.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "output_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "targets.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "targets.0.key", "InstanceIds"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "targets.0.values.0", "aws_instance.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "document_version", "$DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
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

func TestAccSSMAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfssm.ResourceAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSMAssociation_disappears_document(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfssm.ResourceDocument(), "aws_ssm_document.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSMAssociation_applyOnlyAtCronInterval(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_basicApplyOnlyAtCronInterval(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "apply_only_at_cron_interval", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssociationConfig_basicApplyOnlyAtCronInterval(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "apply_only_at_cron_interval", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccSSMAssociation_withTargets(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_association.test"
	oneTarget := `

targets {
  key    = "tag:Name"
  values = ["acceptanceTest"]
}
`

	twoTargets := `

targets {
  key    = "tag:Name"
  values = ["acceptanceTest"]
}

targets {
  key    = "tag:ExtraName"
  values = ["acceptanceTest"]
}
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_basicTargets(rName, oneTarget),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "targets.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.0", "acceptanceTest"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssociationConfig_basicTargets(rName, twoTargets),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "targets.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.0", "acceptanceTest"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.key", "tag:ExtraName"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.0", "acceptanceTest"),
				),
			},
			{
				Config: testAccAssociationConfig_basicTargets(rName, oneTarget),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "targets.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.0", "acceptanceTest"),
				),
			},
		},
	})
}

func TestAccSSMAssociation_withParameters(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_basicParameters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameters.Directory", "myWorkSpace"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrParameters},
			},
			{
				Config: testAccAssociationConfig_basicParametersUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameters.Directory", "myWorkSpaceUpdated"),
				),
			},
		},
	})
}

func TestAccSSMAssociation_withAssociationName(t *testing.T) {
	ctx := acctest.Context(t)
	assocName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	assocName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_basicName(rName, assocName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "association_name", assocName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssociationConfig_basicName(rName, assocName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "association_name", assocName2),
				),
			},
		},
	})
}

func TestAccSSMAssociation_withAssociationNameAndScheduleExpression(t *testing.T) {
	ctx := acctest.Context(t)
	assocName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_association.test"
	scheduleExpression1 := "cron(0 16 ? * TUE *)"
	scheduleExpression2 := "cron(0 16 ? * WED *)"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_nameAndScheduleExpression(rName, assocName, scheduleExpression1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "association_name", assocName),
					resource.TestCheckResourceAttr(resourceName, names.AttrScheduleExpression, scheduleExpression1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssociationConfig_nameAndScheduleExpression(rName, assocName, scheduleExpression2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "association_name", assocName),
					resource.TestCheckResourceAttr(resourceName, names.AttrScheduleExpression, scheduleExpression2),
				),
			},
		},
	})
}

func TestAccSSMAssociation_withDocumentVersion(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_basicDocumentVersion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "document_version", acctest.Ct1),
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

func TestAccSSMAssociation_withOutputLocation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_basicOutPutLocation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "output_location.0.s3_bucket_name", "aws_s3_bucket.output_location", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "output_location.0.s3_key_prefix", "SSMAssociation"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssociationConfig_basicOutPutLocationUpdateBucketName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "output_location.0.s3_bucket_name", "aws_s3_bucket.output_location_updated", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "output_location.0.s3_key_prefix", "SSMAssociation"),
				),
			},
			{
				Config: testAccAssociationConfig_basicOutPutLocationUpdateKeyPrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "output_location.0.s3_bucket_name", "aws_s3_bucket.output_location_updated", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "output_location.0.s3_key_prefix", "UpdatedAssociation"),
				),
			},
		},
	})
}

func TestAccSSMAssociation_withOutputLocation_s3Region(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckMultipleRegion(t, 2) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_outputLocationS3Region(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "output_location.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "output_location.0.s3_bucket_name", rName),
					resource.TestCheckResourceAttr(resourceName, "output_location.0.s3_region", acctest.Region()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssociationConfig_outputLocationUpdateS3Region(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "output_location.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "output_location.0.s3_bucket_name", rName),
					resource.TestCheckResourceAttr(resourceName, "output_location.0.s3_region", acctest.AlternateRegion()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssociationConfig_outputLocationNoS3Region(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "output_location.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "output_location.0.s3_region", ""),
				),
			},
		},
	})
}

func TestAccSSMAssociation_withOutputLocation_waitForSuccessTimeout(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckMultipleRegion(t, 2) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_outputLocationAndWaitForSuccess(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"wait_for_success_timeout_seconds",
				},
			},
		},
	})
}

func TestAccSSMAssociation_withAutomationTargetParamName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_basicAutomationTargetParamName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameters.Directory", "myWorkSpace"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrParameters},
			},
			{
				Config: testAccAssociationConfig_basicParametersUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameters.Directory", "myWorkSpaceUpdated"),
				),
			},
		},
	})
}

func TestAccSSMAssociation_withScheduleExpression(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_basicScheduleExpression(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrScheduleExpression, "cron(0 16 ? * TUE *)"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssociationConfig_basicScheduleExpressionUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrScheduleExpression, "cron(0 16 ? * WED *)"),
				),
			},
		},
	})
}

func TestAccSSMAssociation_withComplianceSeverity(t *testing.T) {
	ctx := acctest.Context(t)
	assocName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	compSeverity1 := "HIGH"
	compSeverity2 := "LOW"
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_basicComplianceSeverity(compSeverity1, rName, assocName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "association_name", assocName),
					resource.TestCheckResourceAttr(resourceName, "compliance_severity", compSeverity1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssociationConfig_basicComplianceSeverity(compSeverity2, rName, assocName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "association_name", assocName),
					resource.TestCheckResourceAttr(resourceName, "compliance_severity", compSeverity2),
				),
			},
		},
	})
}

func TestAccSSMAssociation_rateControl(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_rateControl(rName, "10%"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "max_concurrency", "10%"),
					resource.TestCheckResourceAttr(resourceName, "max_errors", "10%"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssociationConfig_rateControl(rName, "20%"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "max_concurrency", "20%"),
					resource.TestCheckResourceAttr(resourceName, "max_errors", "20%"),
				),
			},
		},
	})
}

func TestAccSSMAssociation_syncCompliance(t *testing.T) {
	ctx := acctest.Context(t)
	rName := "AWS-RunPatchBaselineAssociation"
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationSyncComplianceConfig(rName, "MANUAL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "sync_compliance", "MANUAL"),
				),
			},
			{
				Config: testAccAssociationSyncComplianceConfig(rName, "AUTO"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "sync_compliance", "AUTO"),
				),
			},
		},
	})
}

func testAccCheckAssociationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMClient(ctx)

		_, err := tfssm.FindAssociationByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssm_association" {
				continue
			}

			_, err := tfssm.FindAssociationByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SSM Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAssociationConfig_basicApplyOnlyAtCronInterval(rName string, applyOnlyAtCronInterval bool) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name                        = aws_ssm_document.test.name
  schedule_expression         = "cron(0 16 ? * TUE *)"
  apply_only_at_cron_interval = %[2]t

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName, applyOnlyAtCronInterval)
}

func testAccAssociationConfig_basicAutomationTargetParamName(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_iam_instance_profile" "ssm_profile" {
  name = %[1]q
  role = aws_iam_role.ssm_role.name
}

data "aws_partition" "current" {}

resource "aws_iam_role" "ssm_role" {
  name = %[1]q
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

}

resource "aws_ssm_document" "foo" {
  name          = %[1]q
  document_type = "Automation"

  content = <<DOC
{
  "description": "Systems Manager Automation Demo",
  "schemaVersion": "0.3",
  "assumeRole": "${aws_iam_role.ssm_role.arn}",
  "parameters": {
    "Directory": {
      "description": "(Optional) The path to the working directory on your instance.",
      "default": "",
      "type": "String",
      "maxChars": 4096
    }
  },
  "mainSteps": [
    {
      "name": "startInstances",
      "action": "aws:runInstances",
      "timeoutSeconds": 1200,
      "maxAttempts": 1,
      "onFailure": "Abort",
      "inputs": {
        "ImageId": "${data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id}",
        "InstanceType": "t2.small",
        "MinInstanceCount": 1,
        "MaxInstanceCount": 1,
        "IamInstanceProfileName": "${aws_iam_instance_profile.ssm_profile.name}"
      }
    },
    {
      "name": "stopInstance",
      "action": "aws:changeInstanceState",
      "maxAttempts": 1,
      "onFailure": "Continue",
      "inputs": {
        "InstanceIds": [
          "{{ startInstances.InstanceIds }}"
        ],
        "DesiredState": "stopped"
      }
    },
    {
      "name": "terminateInstance",
      "action": "aws:changeInstanceState",
      "maxAttempts": 1,
      "onFailure": "Continue",
      "inputs": {
        "InstanceIds": [
          "{{ startInstances.InstanceIds }}"
        ],
        "DesiredState": "terminated"
      }
    }
  ]
}
DOC

}

resource "aws_ssm_association" "test" {
  name                             = aws_ssm_document.foo.name
  automation_target_parameter_name = "Directory"

  parameters = {
    AutomationAssumeRole = aws_iam_role.ssm_role.id
    Directory            = "myWorkSpace"
  }

  targets {
    key    = "tag:myTagName"
    values = ["myTagValue"]
  }

  schedule_expression = "rate(60 minutes)"
}
`, rName))
}

func testAccAssociationConfig_basicParametersUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "%[1]s-2"
  document_type = "Command"

  content = <<-DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {
    "Directory": {
      "description": "(Optional) The path to the working directory on your instance.",
      "default": "",
      "type": "String",
      "maxChars": 4096
    }
  },
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
  DOC

}

resource "aws_ssm_association" "test" {
  name = aws_ssm_document.test.name

  parameters = {
    Directory = "myWorkSpaceUpdated"
  }

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName)
}

func testAccAssociationConfig_basicParameters(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Command"

  content = <<-DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {
    "Directory": {
      "description": "(Optional) The path to the working directory on your instance.",
      "default": "",
      "type": "String",
      "maxChars": 4096
    }
  },
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
  DOC

}

resource "aws_ssm_association" "test" {
  name = aws_ssm_document.test.name

  parameters = {
    Directory = "myWorkSpace"
  }

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName)
}

func testAccAssociationConfig_basicTargets(rName, targetsStr string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name = aws_ssm_document.test.name
  %s
}
`, rName, targetsStr)
}

func testAccAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_ami" "amzn" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "first" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "foo"
  vpc_id      = aws_vpc.main.id

  ingress {
    protocol    = "icmp"
    from_port   = -1
    to_port     = -1
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "test" {
  ami                    = data.aws_ami.amzn.image_id
  availability_zone      = data.aws_availability_zones.available.names[0]
  instance_type          = "t2.micro"
  vpc_security_group_ids = [aws_security_group.test.id]
  subnet_id              = aws_subnet.first.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name        = %[1]q
  instance_id = aws_instance.test.id
}
`, rName)
}

func testAccAssociationConfig_basicDocumentVersion(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name             = %[1]q
  document_version = aws_ssm_document.test.latest_version

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName)
}

func testAccAssociationConfig_basicScheduleExpression(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name                = aws_ssm_document.test.name
  schedule_expression = "cron(0 16 ? * TUE *)"

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName)
}

func testAccAssociationConfig_basicScheduleExpressionUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name                = aws_ssm_document.test.name
  schedule_expression = "cron(0 16 ? * WED *)"

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName)
}

func testAccAssociationConfig_basicOutPutLocation(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "output_location" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name = aws_ssm_document.test.name

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }

  output_location {
    s3_bucket_name = aws_s3_bucket.output_location.id
    s3_key_prefix  = "SSMAssociation"
  }
}
`, rName)
}

func testAccAssociationWithOutputLocationS3RegionConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC
}
`, rName)
}

func testAccAssociationConfig_outputLocationS3Region(rName string) string {
	return acctest.ConfigCompose(
		testAccAssociationWithOutputLocationS3RegionConfigBase(rName),
		`
resource "aws_ssm_association" "test" {
  name = aws_ssm_document.test.name

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }

  output_location {
    s3_bucket_name = aws_s3_bucket.test.id
    s3_region      = aws_s3_bucket.test.region
  }
}
`)
}

func testAccAssociationConfig_outputLocationUpdateS3Region(rName string) string {
	return acctest.ConfigCompose(
		testAccAssociationWithOutputLocationS3RegionConfigBase(rName),
		fmt.Sprintf(`
resource "aws_ssm_association" "test" {
  name = aws_ssm_document.test.name

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }

  output_location {
    s3_bucket_name = aws_s3_bucket.test.id
    s3_region      = %[1]q
  }
}
`, acctest.AlternateRegion()))
}

func testAccAssociationConfig_outputLocationNoS3Region(rName string) string {
	return acctest.ConfigCompose(
		testAccAssociationWithOutputLocationS3RegionConfigBase(rName),
		`
resource "aws_ssm_association" "test" {
  name = aws_ssm_document.test.name

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }

  output_location {
    s3_bucket_name = aws_s3_bucket.test.id
  }
}
`)
}

func testAccAssociationConfig_basicOutPutLocationUpdateBucketName(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "output_location" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket" "output_location_updated" {
  bucket        = "%[1]s-2"
  force_destroy = true
}

resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name = aws_ssm_document.test.name

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }

  output_location {
    s3_bucket_name = aws_s3_bucket.output_location_updated.id
    s3_key_prefix  = "SSMAssociation"
  }
}
`, rName)
}

func testAccAssociationConfig_basicOutPutLocationUpdateKeyPrefix(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "output_location" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket" "output_location_updated" {
  bucket        = "%[1]s-2"
  force_destroy = true
}

resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name = aws_ssm_document.test.name

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }

  output_location {
    s3_bucket_name = aws_s3_bucket.output_location_updated.id
    s3_key_prefix  = "UpdatedAssociation"
  }
}
`, rName)
}

func testAccAssociationConfig_basicName(rName, assocName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name             = aws_ssm_document.test.name
  association_name = %[2]q

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName, assocName)
}

func testAccAssociationConfig_nameAndScheduleExpression(rName, associationName, scheduleExpression string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  association_name    = %[2]q
  name                = aws_ssm_document.test.name
  schedule_expression = %[3]q

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName, associationName, scheduleExpression)
}

func testAccAssociationConfig_basicComplianceSeverity(compSeverity, rName, assocName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name                = aws_ssm_document.test.name
  association_name    = %[2]q
  compliance_severity = %[3]q

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName, assocName, compSeverity)
}

func testAccAssociationConfig_rateControl(rName, rate string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name            = aws_ssm_document.test.name
  max_concurrency = %[2]q
  max_errors      = %[2]q

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName, rate)
}

func testAccAssociationSyncComplianceConfig(rName, syncCompliance string) string {
	return fmt.Sprintf(`
resource "aws_ssm_association" "test" {
  name = %[1]q
  targets {
    key    = "InstanceIds"
    values = ["*"]
  }
  apply_only_at_cron_interval = false
  sync_compliance             = %[2]q
  parameters = {
    Operation    = "Scan"
    RebootOption = "NoReboot"
  }
  schedule_expression = "cron(0 6 ? * * *)"
  lifecycle {
    ignore_changes = [
      parameters["AssociationId"]
    ]
  }
}
`, rName, syncCompliance)
}

func testAccAssociationConfig_outputLocationAndWaitForSuccess(rName string) string {
	return acctest.ConfigCompose(
		testAccAssociationWithOutputLocationS3RegionConfigBase(rName),
		`
resource "aws_ssm_association" "test" {
  name = aws_ssm_document.test.name

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }

  output_location {
    s3_bucket_name = aws_s3_bucket.test.id
    s3_region      = aws_s3_bucket.test.region
  }

  wait_for_success_timeout_seconds = 1800
}
`)
}
