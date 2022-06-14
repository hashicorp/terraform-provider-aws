package athena_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfathena "github.com/hashicorp/terraform-provider-aws/internal/service/athena"
)

func TestAccAthenaWorkGroup_basic(t *testing.T) {
	var workgroup1 athena.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup1),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "athena", fmt.Sprintf("workgroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.enforce_workgroup_configuration", "true"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.engine_version.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration.0.engine_version.0.effective_engine_version"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.engine_version.0.selected_engine_version", "AUTO"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.publish_cloudwatch_metrics_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.requester_pays_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "state", athena.WorkGroupStateEnabled),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func TestAccAthenaWorkGroup_aclConfig(t *testing.T) {
	var workgroup1 athena.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_configurationResultConfigurationACL(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup1),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "athena", fmt.Sprintf("workgroup/%s", rName)),
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
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func TestAccAthenaWorkGroup_disappears(t *testing.T) {
	var workgroup1 athena.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup1),
					acctest.CheckResourceDisappears(acctest.Provider, tfathena.ResourceWorkGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAthenaWorkGroup_bytesScannedCutoffPerQuery(t *testing.T) {
	var workgroup1, workgroup2 athena.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_configurationBytesScannedCutoffPerQuery(rName, 12582912),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.bytes_scanned_cutoff_per_query", "12582912"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccWorkGroupConfig_configurationBytesScannedCutoffPerQuery(rName, 10485760),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.bytes_scanned_cutoff_per_query", "10485760"),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_enforceWorkGroup(t *testing.T) {
	var workgroup1, workgroup2 athena.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_enforce(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.enforce_workgroup_configuration", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccWorkGroupConfig_enforce(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.enforce_workgroup_configuration", "true"),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_configurationEngineVersion(t *testing.T) {
	var workgroup1, workgroup2 athena.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_configurationEngineVersion(rName, "Athena engine version 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.engine_version.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.engine_version.0.effective_engine_version", resourceName, "configuration.0.engine_version.0.selected_engine_version"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.engine_version.0.selected_engine_version", "Athena engine version 2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccWorkGroupConfig_configurationEngineVersion(rName, "AUTO"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup2),
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
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccWorkGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttrSet(resourceName, "configuration.0.engine_version.0.effective_engine_version"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.engine_version.0.selected_engine_version", "AUTO"),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_publishCloudWatchMetricsEnabled(t *testing.T) {
	var workgroup1, workgroup2 athena.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_configurationPublishCloudWatchMetricsEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.publish_cloudwatch_metrics_enabled", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccWorkGroupConfig_configurationPublishCloudWatchMetricsEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.publish_cloudwatch_metrics_enabled", "true"),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_ResultEncryption_sseS3(t *testing.T) {
	var workgroup1 athena.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_configurationResultConfigurationEncryptionConfigurationEncryptionOptionSseS3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.encryption_configuration.0.encryption_option", athena.EncryptionOptionSseS3),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func TestAccAthenaWorkGroup_ResultEncryption_kms(t *testing.T) {
	var workgroup1, workgroup2 athena.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"
	rEncryption := athena.EncryptionOptionSseKms
	rEncryption2 := athena.EncryptionOptionCseKms

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_resultEncryptionEncryptionOptionKMS(rName, rEncryption),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup1),
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
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccWorkGroupConfig_resultEncryptionEncryptionOptionKMS(rName, rEncryption2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.encryption_configuration.0.encryption_option", rEncryption2),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_Result_outputLocation(t *testing.T) {
	var workgroup1, workgroup2 athena.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"
	rOutputLocation1 := fmt.Sprintf("%s-1", rName)
	rOutputLocation2 := fmt.Sprintf("%s-2", rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_configurationResultConfigurationOutputLocation(rName, rOutputLocation1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.output_location", "s3://"+rOutputLocation1+"/test/output"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccWorkGroupConfig_configurationResultConfigurationOutputLocation(rName, rOutputLocation2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.output_location", "s3://"+rOutputLocation2+"/test/output"),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_requesterPaysEnabled(t *testing.T) {
	var workgroup1 athena.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_configurationRequesterPaysEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup1),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "athena", fmt.Sprintf("workgroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.requester_pays_enabled", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccWorkGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup1),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "athena", fmt.Sprintf("workgroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.requester_pays_enabled", "false"),
				),
			},
		},
	})
} //

func TestAccAthenaWorkGroup_ResultOutputLocation_forceDestroy(t *testing.T) {
	var workgroup1, workgroup2 athena.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"
	rOutputLocation1 := fmt.Sprintf("%s-1", rName)
	rOutputLocation2 := fmt.Sprintf("%s-2", rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_configurationResultConfigurationOutputLocationForceDestroy(rName, rOutputLocation1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.output_location", "s3://"+rOutputLocation1+"/test/output"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccWorkGroupConfig_configurationResultConfigurationOutputLocationForceDestroy(rName, rOutputLocation2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.output_location", "s3://"+rOutputLocation2+"/test/output"),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_description(t *testing.T) {
	var workgroup1, workgroup2 athena.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"
	rDescription := sdkacctest.RandString(20)
	rDescriptionUpdate := sdkacctest.RandString(20)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_description(rName, rDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "description", rDescription),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccWorkGroupConfig_description(rName, rDescriptionUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "description", rDescriptionUpdate),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_state(t *testing.T) {
	var workgroup1, workgroup2, workgroup3 athena.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_state(rName, athena.WorkGroupStateDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "state", athena.WorkGroupStateDisabled),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccWorkGroupConfig_state(rName, athena.WorkGroupStateEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "state", athena.WorkGroupStateEnabled),
				),
			},
			{
				Config: testAccWorkGroupConfig_state(rName, athena.WorkGroupStateDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup3),
					resource.TestCheckResourceAttr(resourceName, "state", athena.WorkGroupStateDisabled),
				),
			},
		},
	})
}

func TestAccAthenaWorkGroup_forceDestroy(t *testing.T) {
	var workgroup athena.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbName := sdkacctest.RandString(5)
	queryName1 := sdkacctest.RandomWithPrefix("tf-athena-named-query-")
	queryName2 := sdkacctest.RandomWithPrefix("tf-athena-named-query-")
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_forceDestroy(rName, dbName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup),
					testAccCheckCreateNamedQuery(&workgroup, dbName, queryName1, fmt.Sprintf("SELECT * FROM %s limit 10;", rName)),
					testAccCheckCreateNamedQuery(&workgroup, dbName, queryName2, fmt.Sprintf("SELECT * FROM %s limit 100;", rName)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func TestAccAthenaWorkGroup_tags(t *testing.T) {
	var workgroup1, workgroup2, workgroup3 athena.WorkGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkGroupConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccWorkGroupConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccWorkGroupConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkGroupExists(resourceName, &workgroup3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckCreateNamedQuery(workGroup *athena.WorkGroup, databaseName, queryName, query string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AthenaConn

		input := &athena.CreateNamedQueryInput{
			Name:        aws.String(queryName),
			WorkGroup:   workGroup.Name,
			Database:    aws.String(databaseName),
			QueryString: aws.String(query),
			Description: aws.String("tf test"),
		}

		if _, err := conn.CreateNamedQuery(input); err != nil {
			return fmt.Errorf("error creating Named Query (%s) on Workgroup (%s): %s", queryName, aws.StringValue(workGroup.Name), err)
		}

		return nil
	}
}

func testAccCheckWorkGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AthenaConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_athena_workgroup" {
			continue
		}

		input := &athena.GetWorkGroupInput{
			WorkGroup: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetWorkGroup(input)

		if tfawserr.ErrMessageContains(err, athena.ErrCodeInvalidRequestException, "is not found") {
			continue
		}

		if err != nil {
			return err
		}

		if resp.WorkGroup != nil {
			return fmt.Errorf("Athena WorkGroup (%s) found", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckWorkGroupExists(name string, workgroup *athena.WorkGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AthenaConn

		input := &athena.GetWorkGroupInput{
			WorkGroup: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetWorkGroup(input)

		if err != nil {
			return err
		}

		*workgroup = *output.WorkGroup

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
