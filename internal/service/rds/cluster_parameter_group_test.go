// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSClusterParameterGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, t, resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, rName),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "rds", fmt.Sprintf("cluster-pg:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrFamily, "aurora5.6"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Test cluster parameter group for terraform"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_results",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_server",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_client",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterParameterGroupConfig_addParameters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, t, resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrFamily, "aurora5.6"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Test cluster parameter group for terraform"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "collation_connection",
						names.AttrValue: "utf8_unicode_ci",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_results",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_server",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "collation_server",
						names.AttrValue: "utf8_unicode_ci",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_client",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				Config: testAccClusterParameterGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, t, resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, rName),
					testAccCheckClusterParameterNotUserDefined(ctx, t, resourceName, "collation_connection"),
					testAccCheckClusterParameterNotUserDefined(ctx, t, resourceName, "collation_server"),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_results",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_server",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_client",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func TestAccRDSClusterParameterGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfrds.ResourceClusterParameterGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRDSClusterParameterGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterParameterGroupConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccClusterParameterGroupConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccRDSClusterParameterGroup_withApplyMethod(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DBClusterParameterGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster_parameter_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_applyMethod(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, t, resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, rName),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "rds", fmt.Sprintf("cluster-pg:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrFamily, "aurora5.6"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Test cluster parameter group for terraform"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_server",
						names.AttrValue: "utf8",
						"apply_method":  "immediate",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_client",
						names.AttrValue: "utf8",
						"apply_method":  "pending-reboot",
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

func TestAccRDSClusterParameterGroup_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_namePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-acc-test-prefix-"),
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

func TestAccRDSClusterParameterGroup_NamePrefix_parameter(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_namePrefixParameter("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-acc-test-prefix-"),
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

func TestAccRDSClusterParameterGroup_generatedName(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, sdkid.UniqueIdPrefix),
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

func TestAccRDSClusterParameterGroup_GeneratedName_parameter(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_generatedName_Parameter,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, sdkid.UniqueIdPrefix),
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

func TestAccRDSClusterParameterGroup_only(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_only(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, t, resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrFamily, "aurora5.6"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
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

func TestAccRDSClusterParameterGroup_updateParameters(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_updateParametersInitial(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, t, resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrFamily, "aurora5.6"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_results",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_server",
						names.AttrValue: "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_client",
						names.AttrValue: "utf8",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterParameterGroupConfig_updateParametersUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, t, resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_results",
						names.AttrValue: "ascii",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_server",
						names.AttrValue: "ascii",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "character_set_client",
						names.AttrValue: "utf8",
					}),
				),
			},
		},
	})
}

func TestAccRDSClusterParameterGroup_caseParameters(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_upperCase(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, t, resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrFamily, "aurora5.6"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "max_connections",
						names.AttrValue: "LEAST({DBInstanceClassMemory/6000000},10)",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterParameterGroupConfig_upperCase(rName),
			},
		},
	})
}

func TestAccRDSClusterParameterGroup_dynamicDiffs(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_dynamicDiffs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrFamily, "aurora-postgresql12"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "track_activity_query_size", // system source
						names.AttrValue: "4096",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "shared_preload_libraries", // system source
						names.AttrValue: "pg_stat_statements",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "track_io_timing", // system source
						names.AttrValue: "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "track_activities", // user source
						names.AttrValue: "1",
					}),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/24035.
func TestAccRDSClusterParameterGroup_charsetAndCollation(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_charsetAndCollation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrFamily), knownvalue.StringExact("aurora-mysql8.0")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrParameter), knownvalue.SetSizeExact(28)),
				},
			},
		},
	})
}

func testAccCheckClusterParameterGroupDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).RDSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rds_cluster_parameter_group" {
				continue
			}

			_, err := tfrds.FindDBClusterParameterGroupByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS Cluster Parameter Group %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckClusterParameterNotUserDefined(ctx context.Context, t *testing.T, n, paramName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).RDSClient(ctx)

		input := &rds.DescribeDBClusterParametersInput{
			DBClusterParameterGroupName: aws.String(rs.Primary.ID),
		}

		userDefined := false
		pages := rds.NewDescribeDBClusterParametersPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)

			if err != nil {
				return err
			}

			for _, param := range page.Parameters {
				if aws.ToString(param.ParameterName) == paramName && aws.ToString(param.ParameterValue) != "" {
					// Some of these resets leave the parameter name present but with a nil value.
					userDefined = true
				}
			}
		}

		if userDefined {
			return fmt.Errorf("Cluster Parameter %s is user defined", paramName)
		}

		return nil
	}
}

func testAccCheckClusterParameterGroupAttributes(v *types.DBClusterParameterGroup, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if v := aws.ToString(v.DBClusterParameterGroupName); v != name {
			return fmt.Errorf("bad name: %#v expected: %v", v, name)
		}

		if v := aws.ToString(v.DBParameterGroupFamily); v != "aurora5.6" {
			return fmt.Errorf("bad family: %#v", v)
		}

		return nil
	}
}

func testAccCheckClusterParameterGroupExists(ctx context.Context, t *testing.T, n string, v *types.DBClusterParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).RDSClient(ctx)

		output, err := tfrds.FindDBClusterParameterGroupByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccClusterParameterGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_parameter_group" "test" {
  name        = %[1]q
  family      = "aurora5.6"
  description = "Test cluster parameter group for terraform"

  parameter {
    name  = "character_set_server"
    value = "utf8"
  }

  parameter {
    name  = "character_set_client"
    value = "utf8"
  }

  parameter {
    name  = "character_set_results"
    value = "utf8"
  }
}
`, rName)
}

func testAccClusterParameterGroupConfig_applyMethod(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_parameter_group" "test" {
  name        = %[1]q
  family      = "aurora5.6"
  description = "Test cluster parameter group for terraform"

  parameter {
    name  = "character_set_server"
    value = "utf8"
  }

  parameter {
    name         = "character_set_client"
    value        = "utf8"
    apply_method = "pending-reboot"
  }
}
`, rName)
}

func testAccClusterParameterGroupConfig_addParameters(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_parameter_group" "test" {
  name        = %[1]q
  family      = "aurora5.6"
  description = "Test cluster parameter group for terraform"

  parameter {
    name  = "character_set_server"
    value = "utf8"
  }

  parameter {
    name  = "character_set_client"
    value = "utf8"
  }

  parameter {
    name  = "character_set_results"
    value = "utf8"
  }

  parameter {
    name  = "collation_server"
    value = "utf8_unicode_ci"
  }

  parameter {
    name  = "collation_connection"
    value = "utf8_unicode_ci"
  }
}
`, rName)
}

func testAccClusterParameterGroupConfig_only(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_parameter_group" "test" {
  name   = %[1]q
  family = "aurora5.6"
}
`, rName)
}

func testAccClusterParameterGroupConfig_updateParametersInitial(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_parameter_group" "test" {
  name   = %[1]q
  family = "aurora5.6"

  parameter {
    name  = "character_set_server"
    value = "utf8"
  }

  parameter {
    name  = "character_set_client"
    value = "utf8"
  }

  parameter {
    name  = "character_set_results"
    value = "utf8"
  }
}
`, rName)
}

func testAccClusterParameterGroupConfig_updateParametersUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_parameter_group" "test" {
  name   = %[1]q
  family = "aurora5.6"

  parameter {
    name  = "character_set_server"
    value = "ascii"
  }

  parameter {
    name  = "character_set_client"
    value = "utf8"
  }

  parameter {
    name  = "character_set_results"
    value = "ascii"
  }
}
`, rName)
}

func testAccClusterParameterGroupConfig_upperCase(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_parameter_group" "test" {
  name   = %[1]q
  family = "aurora5.6"

  parameter {
    name  = "max_connections"
    value = "LEAST({DBInstanceClassMemory/6000000},10)"
  }
}
`, rName)
}

func testAccClusterParameterGroupConfig_namePrefix(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_parameter_group" "test" {
  name_prefix = %[1]q
  family      = "aurora5.6"
}
`, namePrefix)
}

func testAccClusterParameterGroupConfig_namePrefixParameter(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_parameter_group" "test" {
  name_prefix = %[1]q
  family      = "aurora5.6"

  parameter {
    name  = "character_set_server"
    value = "utf8"
  }

}
`, namePrefix)
}

const testAccClusterParameterGroupConfig_generatedName = `
resource "aws_rds_cluster_parameter_group" "test" {
  family = "aurora5.6"
}
`

const testAccClusterParameterGroupConfig_generatedName_Parameter = `
resource "aws_rds_cluster_parameter_group" "test" {
  family = "aurora5.6"

  parameter {
    name  = "character_set_server"
    value = "utf8"
  }
}
`

func testAccClusterParameterGroupConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_parameter_group" "test" {
  name   = %[1]q
  family = "aurora5.6"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccClusterParameterGroupConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_parameter_group" "test" {
  name   = %[1]q
  family = "aurora5.6"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccClusterParameterGroupConfig_dynamicDiffs(rName string) string {
	return fmt.Sprintf(`
locals {
  cluster_parameters = {
    "shared_preload_libraries" = { # system source
      value        = "pg_stat_statements"
      apply_method = "pending-reboot"
    },
    "track_activity_query_size" = { # system source
      value        = "4096"
      apply_method = "pending-reboot"
    },
    "pg_stat_statements.track" = {
      value        = "ALL"
      apply_method = "pending-reboot"
    },
    "pg_stat_statements.max" = {
      value        = "10000"
      apply_method = "pending-reboot"
    },
    "track_activities" = {
      value        = "1"
      apply_method = "pending-reboot"
    },
    "track_counts" = {
      value        = "1"
      apply_method = "pending-reboot"
    },
    "track_io_timing" = { # system source
      value        = "1"
      apply_method = "pending-reboot"
    },
  }
}

resource "aws_rds_cluster_parameter_group" "test" {
  name   = %[1]q
  family = "aurora-postgresql12"

  dynamic "parameter" {
    for_each = local.cluster_parameters
    content {
      name         = parameter.key
      value        = parameter.value["value"]
      apply_method = parameter.value["apply_method"]
    }
  }
}
`, rName)
}

func testAccClusterParameterGroupConfig_charsetAndCollation(rName string) string {
	return fmt.Sprintf(`
variable "aurora_oom_response" {
  type        = string
  default     = "print"
  description = "Determines how the server responds to memory allocation requests that cannot be satisfied."
}

variable "aws_default_logs_role" {
  type        = string
  description = "IAM role for uploading logs to CloudWatch. Defaults to using service-linked role"
  default     = null
}

variable "binlog_cache_size" {
  type        = number
  default     = 32768
  description = "The size of the cache to hold the SQL statements for the binary log during a transaction."
}

variable "binlog_format" {
  type        = string
  default     = "ROW"
  description = "Row or Mixed replication"
}

variable "character_set_client" {
  type        = string
  default     = "utf8"
  description = "The character set for statements that arrive from the client."
}

variable "character_set_connection" {
  type        = string
  default     = "utf8"
  description = "The character set used for literals that do not have a character set introducer and for number-to-string."
}

variable "character_set_database" {
  type        = string
  default     = "utf8"
  description = "The character set used by the default database."
}

variable "character_set_filesystem" {
  type        = string
  default     = "binary"
  description = "The file system character set."
}

variable "character_set_results" {
  type        = string
  default     = "utf8"
  description = "The character set used for returning query results to the client."
}

variable "character_set_server" {
  type        = string
  default     = "utf8"
  description = "The server's default character set."
}

variable "collation_connection" {
  type        = string
  default     = "utf8_general_ci"
  description = "The collation of the connection character set."
}

variable "collation_server" {
  type        = string
  default     = "utf8_general_ci"
  description = "The server's default collation."
}

variable "enforce_gtid_consistency" {
  type        = string
  default     = "OFF"
  description = "Specifies whether to check GTID consistency"
}

variable "eq_range_index_dive_limit" {
  type        = number
  default     = 10
  description = "Number of equality ranges when the optimizer should switch from using index dives to index statistics."
}

variable "event_scheduler" {
  type        = bool
  default     = false
  description = "Indicates the status of the Event Scheduler"
}

variable "explicit_defaults_for_timestamp" {
  type    = bool
  default = true
}

variable "extra_immediate_cluster_parameter" {
  type        = map(string)
  default     = {}
  description = "A list of custom non-sticking parameter values for clusters"
}

variable "extra_immediate_instance_parameter" {
  type        = map(string)
  default     = {}
  description = "A list of custom non-sticking parameter values for instances"
}

variable "extra_pending_cluster_parameter" {
  type        = map(string)
  default     = {}
  description = "A list of custom non-sticking parameter values for clusters"
}

variable "extra_pending_instance_parameter" {
  type        = map(string)
  default     = {}
  description = "A list of custom non-sticking parameter values for instances"
}

variable "general_log" {
  type        = bool
  default     = false
  description = "Whether the general query log is enabled"
}

variable "gtid_mode" {
  type        = string
  default     = "OFF"
  description = "Specifies whether global transaction identifiers (GTIDs) are used to identify transactions."
}

variable "innodb_adaptive_hash_index" {
  type        = bool
  default     = false
  description = "Whether innodb adaptive hash indexes are enabled or disabled"
}

variable "innodb_autoinc_lock_mode" {
  type        = number
  default     = 1
  description = "Tune autoinc generation mode (consecutive, interleaved, etc.)"
}

variable "innodb_buffer_pool_size" {
  type        = number
  default     = null
  description = "Specifies whether to check GTID consistency"
}

variable "innodb_file_per_table" {
  type        = string
  default     = "1"
  description = "Use tablespaces or files for Innodb."
}

variable "innodb_lock_wait_timeout" {
  type        = number
  default     = 1
  description = "Timeout in seconds an innodb transaction may wait for a row lock before giving up"
}

variable "innodb_online_alter_log_max_size" {
  type        = number
  default     = 134217728
  description = "Specifies an upper limit on the size of the temporary log files used during online DDL operations for InnoDB tables."
}

variable "innodb_open_files" {
  type        = number
  default     = 2000
  description = "Relevant only if you use multiple tablespaces in innodb. It specifies the maximum number of .ibd files that innodb can keep open at one time"
}

variable "innodb_print_all_deadlocks" {
  type        = bool
  default     = false
  description = "When this option is enabled, information about all deadlocks in InnoDB user transactions is recorded."
}

variable "innodb_stats_on_metadata" {
  type        = bool
  default     = false
  description = "Controls whether table and index stats are updated when getting status information via SHOW STATUS or the INFORMATION_SCHEMA"
}

variable "innodb_strict_mode" {
  type        = bool
  default     = false
  description = "Whether InnoDB returns errors rather than warnings for exceptional conditions."
}

variable "key_buffer_size" {
  type        = number
  default     = null
  description = "Increase the buffer size to get better index handling used for index blocks (for all reads and multiple writes)."
}

variable "key_buffer_size_lookup" {
  type = map(string)
  default = {
    "db.t3.small"     = 8388608
    "db.t3.medium"    = 8388608
    "db.r5.large"     = 8388608
    "db.r5.xlarge"    = 8388608
    "db.r5.2xlarge"   = 16777216
    "db.r5.4xlarge"   = 16777216
    "db.r5.8xlarge"   = 16777216
    "db.r5.12xlarge"  = 16777216
    "db.r5.16xlarge"  = 16777216
    "db.r5.24xlarge"  = 16777216
    "db.r6i.large"    = 8388608
    "db.r6i.xlarge"   = 8388608
    "db.r6i.2xlarge"  = 16777216
    "db.r6i.4xlarge"  = 16777216
    "db.r6i.8xlarge"  = 16777216
    "db.r6i.12xlarge" = 16777216
    "db.r6i.16xlarge" = 16777216
    "db.r6i.24xlarge" = 16777216
    "db.r6i.32xlarge" = 16777216
    "db.r7g.large"    = 8388608
    "db.r7g.xlarge"   = 8388608
    "db.r7g.2xlarge"  = 16777216
    "db.r7g.4xlarge"  = 16777216
    "db.r7g.8xlarge"  = 16777216
    "db.r7g.12xlarge" = 16777216
    "db.r7g.16xlarge" = 16777216
    "db.r7g.24xlarge" = 16777216
    "db.r7g.32xlarge" = 16777216
  }
}

variable "local_infile" {
  type        = bool
  default     = true
  description = "Controls whether LOCAL is supported for LOAD DATA INFILE"
}

variable "lock_wait_timeout" {
  type        = number
  default     = 31536000
  description = "Specifies the timeout in seconds for attempts to acquire metadata locks"
}

variable "log_bin_trust_function_creators" {
  type        = bool
  default     = true
  description = "Enforces restrictions on stored functions / triggers - logging for replication."
}

variable "log_output" {
  type        = string
  default     = "FILE"
  description = "Controls where to store query logs"
}

variable "long_query_time" {
  type        = number
  default     = 15
  description = "Defines what MySQL considers long queries"
}

variable "max_allowed_packet" {
  type        = number
  default     = 4194304
  description = "This value by default is small, to catch large (possibly incorrect) packets. Must be increased if using large BLOB columns or long strings. As big as largest BLOB. "
}

variable "max_connect_errors" {
  type        = number
  default     = 1000000
  description = "A host is blocked from further connections if there are more than this number of interrupted connections"
}

variable "max_connections" {
  type        = number
  default     = 10000
  description = "The number of simultaneous client connections allowed."
}

variable "max_connections_lookup" {
  type = map(string)
  default = {
    "db.t3.small"     = 50
    "db.t3.medium"    = 100
    "db.r5.large"     = 1000
    "db.r5.xlarge"    = 10000
    "db.r5.2xlarge"   = 15000
    "db.r5.4xlarge"   = 15800
    "db.r5.8xlarge"   = 15800
    "db.r5.12xlarge"  = 15800
    "db.r5.16xlarge"  = 15800
    "db.r5.24xlarge"  = 15800
    "db.r6i.large"    = 1000
    "db.r6i.xlarge"   = 10000
    "db.r6i.2xlarge"  = 15000
    "db.r6i.4xlarge"  = 15800
    "db.r6i.8xlarge"  = 15800
    "db.r6i.12xlarge" = 15800
    "db.r6i.16xlarge" = 15800
    "db.r6i.24xlarge" = 15800
    "db.r6i.32xlarge" = 15800
    "db.r7g.large"    = 1000
    "db.r7g.xlarge"   = 10000
    "db.r7g.2xlarge"  = 15000
    "db.r7g.4xlarge"  = 15800
    "db.r7g.8xlarge"  = 15800
    "db.r7g.12xlarge" = 15800
    "db.r7g.16xlarge" = 15800
    "db.r7g.24xlarge" = 15800
    "db.r7g.32xlarge" = 15800
  }
}

variable "parameter_group_prefix" {
  type        = string
  default     = "aurora-pg-"
  description = "Prefix for RDS parameter group names"
}

variable "range_optimizer_max_mem_size" {
  type        = number
  default     = 67108864
  description = "The limit on memory consumption for the range optimizer."
}

variable "read_only" {
  type        = string
  default     = "{TrueIfClusterReplica}"
  description = "When it is enabled, the server permits no updates except from updates performed by slave threads."
}

variable "replica_preserve_commit_order" {
  type        = bool
  default     = true
  description = "Preserve commit order when replicating from binlog source"
}

variable "replica_parallel_workers" {
  type        = number
  default     = 1
  description = "Set number of threads while replicating from binlog source"
}

variable "binlog_transaction_dependency_tracking" {
  type        = string
  default     = "WRITESET_SESSION"
  description = "Set the binlog transaction dependency tracking method"
}

variable "sort_buffer_size" {
  type        = number
  default     = 262144
  description = "Size of buffer allocated by each session that must perform a sort."
}

variable "sql_mode" {
  type        = string
  default     = "NO_ENGINE_SUBSTITUTION"
  description = "Current SQL Server Mode."
}

variable "thread_stack" {
  type        = number
  default     = 262144
  description = "If the thread stack size is too small, it limits the complexity of the SQL statements that the server can handle, the recursion depth of stored procedures, and other memory-consuming actions."
}

variable "transaction_isolation" {
  type        = string
  default     = "REPEATABLE-READ"
  description = "Specifies the default transaction isolation level."
}

variable "max_heap_table_size" {
  type        = number
  default     = null
  description = "Maximum size to which MEMORY tables are allowed to grow."
}

variable "max_heap_table_size_lookup" {
  type = map(string)
  default = {
    "db.t3.small"     = 16777216
    "db.t3.medium"    = 16777216
    "db.r5.large"     = 16777216
    "db.r5.xlarge"    = 16777216
    "db.r5.2xlarge"   = 16777216
    "db.r5.4xlarge"   = 16777216
    "db.r5.8xlarge"   = 16777216
    "db.r5.12xlarge"  = 33554432
    "db.r5.16xlarge"  = 33554432
    "db.r5.24xlarge"  = 33554432
    "db.r6i.large"    = 16777216
    "db.r6i.xlarge"   = 16777216
    "db.r6i.2xlarge"  = 16777216
    "db.r6i.4xlarge"  = 16777216
    "db.r6i.8xlarge"  = 16777216
    "db.r6i.12xlarge" = 33554432
    "db.r6i.16xlarge" = 33554432
    "db.r6i.24xlarge" = 33554432
    "db.r6i.32xlarge" = 33554432
    "db.r7g.large"    = 16777216
    "db.r7g.xlarge"   = 16777216
    "db.r7g.2xlarge"  = 16777216
    "db.r7g.4xlarge"  = 16777216
    "db.r7g.8xlarge"  = 16777216
    "db.r7g.12xlarge" = 33554432
    "db.r7g.16xlarge" = 33554432
    "db.r7g.24xlarge" = 33554432
    "db.r7g.32xlarge" = 33554432
  }
}

variable "performance_schema" {
  type        = bool
  default     = true
  description = "Enables or disables the Performance Schema"
}

variable "read_buffer_size" {
  type        = number
  default     = 262144
  description = "Each thread that does a sequential scan allocates this buffer. Increased value may help perf if performing many sequential scans."
}

variable "read_rnd_buffer_size" {
  type        = number
  default     = 524288
  description = "Avoids disk reads when reading rows in sorted order following a key-sort operation. Large values can improve ORDER BY perf."
}

variable "server_audit_events" {
  type        = string
  default     = "CONNECT,QUERY_DCL,QUERY_DDL"
  description = "If set it specifies the set of types of events to log."
}

variable "server_audit_logging" {
  type        = bool
  default     = false
  description = "Enables audit logging."
}

variable "server_audit_logs_upload" {
  type        = bool
  default     = true
  description = "Enables audit log upload to CloudWatch Logs."
}

variable "slow_query_log" {
  type        = bool
  default     = true
  description = "Enable or disable the slow query log"
}

variable "table_definition_cache" {
  type        = number
  default     = 4096
  description = "The number of table definitions that can be stored in the definition cache"
}

variable "table_open_cache" {
  type        = string
  default     = null
  description = "The number of open tables for all threads. Increasing this value increases the number of file descriptors."
}

variable "table_open_cache_lookup" {
  type = map(string)
  default = {
    "db.t3.small"     = 100
    "db.t3.medium"    = 200
    "db.r5.large"     = 2000
    "db.r5.xlarge"    = 5000
    "db.r5.2xlarge"   = 7500
    "db.r5.4xlarge"   = 10000
    "db.r5.8xlarge"   = 10000
    "db.r5.12xlarge"  = 10000
    "db.r5.16xlarge"  = 10000
    "db.r5.24xlarge"  = 10000
    "db.r6i.large"    = 2000
    "db.r6i.xlarge"   = 5000
    "db.r6i.2xlarge"  = 7500
    "db.r6i.4xlarge"  = 10000
    "db.r6i.8xlarge"  = 10000
    "db.r6i.12xlarge" = 10000
    "db.r6i.16xlarge" = 10000
    "db.r6i.24xlarge" = 10000
    "db.r6i.32xlarge" = 10000
    "db.r7g.large"    = 2000
    "db.r7g.xlarge"   = 5000
    "db.r7g.2xlarge"  = 7500
    "db.r7g.4xlarge"  = 10000
    "db.r7g.8xlarge"  = 10000
    "db.r7g.12xlarge" = 10000
    "db.r7g.16xlarge" = 10000
    "db.r7g.24xlarge" = 10000
    "db.r7g.32xlarge" = 10000
  }
}

variable "table_open_cache_instances" {
  type        = number
  default     = 16
  description = "The number of open tables cache instances."
}

variable "thread_cache_size" {
  type        = string
  default     = null
  description = "Number of threads to be cached. Doesn't improve perf for good thread implementations."
}

variable "thread_cache_size_lookup" {
  type = map(string)
  default = {
    "db.t3.small"     = 50
    "db.t3.medium"    = 100
    "db.r5.large"     = 1000
    "db.r5.xlarge"    = 2000
    "db.r5.2xlarge"   = 5000
    "db.r5.4xlarge"   = 5000
    "db.r5.8xlarge"   = 5000
    "db.r5.12xlarge"  = 5000
    "db.r5.16xlarge"  = 5000
    "db.r5.24xlarge"  = 5000
    "db.r6i.large"    = 1000
    "db.r6i.xlarge"   = 2000
    "db.r6i.2xlarge"  = 5000
    "db.r6i.4xlarge"  = 5000
    "db.r6i.8xlarge"  = 5000
    "db.r6i.12xlarge" = 5000
    "db.r6i.16xlarge" = 5000
    "db.r6i.24xlarge" = 5000
    "db.r6i.32xlarge" = 5000
    "db.r7g.large"    = 1000
    "db.r7g.xlarge"   = 2000
    "db.r7g.2xlarge"  = 5000
    "db.r7g.4xlarge"  = 5000
    "db.r7g.8xlarge"  = 5000
    "db.r7g.12xlarge" = 5000
    "db.r7g.16xlarge" = 5000
    "db.r7g.24xlarge" = 5000
    "db.r7g.32xlarge" = 5000
  }
}

resource "aws_rds_cluster_parameter_group" "test" {
  name   = %[1]q
  family = "aurora-mysql8.0"

  parameter {
    name         = "binlog_format"
    value        = var.binlog_format
    apply_method = "pending-reboot"
  }

  parameter {
    name         = "character_set_client"
    value        = var.character_set_client
    apply_method = "immediate"
  }

  parameter {
    name         = "character_set_connection"
    value        = var.character_set_connection
    apply_method = "immediate"
  }

  parameter {
    name         = "character_set_database"
    value        = var.character_set_database
    apply_method = "immediate"
  }

  parameter {
    name         = "character_set_server"
    value        = var.character_set_server
    apply_method = "immediate"
  }

  parameter {
    name         = "character_set_filesystem"
    value        = var.character_set_filesystem
    apply_method = "immediate"
  }

  parameter {
    name         = "character_set_results"
    value        = var.character_set_results
    apply_method = "immediate"
  }

  parameter {
    name         = "collation_connection"
    value        = var.collation_connection
    apply_method = "immediate"
  }

  parameter {
    name         = "collation_server"
    value        = var.collation_server
    apply_method = "immediate"
  }

  parameter {
    name         = "innodb_adaptive_hash_index"
    value        = var.innodb_adaptive_hash_index ? 1 : 0
    apply_method = "pending-reboot"
  }

  parameter {
    name         = "innodb_autoinc_lock_mode"
    value        = var.innodb_autoinc_lock_mode
    apply_method = "pending-reboot"
  }

  parameter {
    name         = "innodb_online_alter_log_max_size"
    value        = var.innodb_online_alter_log_max_size
    apply_method = "immediate"
  }

  parameter {
    name         = "innodb_file_per_table"
    value        = var.innodb_file_per_table
    apply_method = "pending-reboot"
  }

  parameter {
    name         = "server_audit_logging"
    value        = var.server_audit_logging ? 1 : 0
    apply_method = "immediate"
  }

  parameter {
    name         = "server_audit_logs_upload"
    value        = var.server_audit_logs_upload ? 1 : 0
    apply_method = "immediate"
  }

  parameter {
    name         = "server_audit_events"
    value        = var.server_audit_events
    apply_method = "immediate"
  }

  parameter {
    name         = "innodb_strict_mode"
    value        = var.innodb_strict_mode ? 1 : 0
    apply_method = "immediate"
  }

  parameter {
    name         = "gtid-mode"
    value        = var.gtid_mode
    apply_method = "pending-reboot"
  }

  parameter {
    name         = "enforce_gtid_consistency"
    value        = var.enforce_gtid_consistency
    apply_method = "pending-reboot"
  }

  parameter {
    name         = "sql_mode"
    value        = var.sql_mode
    apply_method = "immediate"
  }

  parameter {
    name         = "transaction_isolation"
    value        = var.transaction_isolation
    apply_method = "immediate"
  }

  parameter {
    name         = "event_scheduler"
    value        = var.event_scheduler ? "on" : "off"
    apply_method = "pending-reboot"
  }

  parameter {
    name         = "replica_preserve_commit_order"
    value        = var.replica_preserve_commit_order ? "on" : "off"
    apply_method = "pending-reboot"
  }

  parameter {
    name         = "replica_parallel_type"
    value        = "LOGICAL_CLOCK"
    apply_method = "pending-reboot"
  }

  parameter {
    name         = "replica_parallel_workers"
    value        = var.replica_parallel_workers
    apply_method = "pending-reboot"
  }

  parameter {
    name         = "binlog_transaction_dependency_tracking"
    value        = var.binlog_transaction_dependency_tracking
    apply_method = "pending-reboot"
  }

  parameter {
    name         = "binlog_cache_size"
    value        = var.binlog_cache_size
    apply_method = "pending-reboot"
  }

  parameter {
    name         = "read_only"
    value        = var.read_only
    apply_method = "pending-reboot"
  }

  dynamic "parameter" {
    for_each = compact([var.aws_default_logs_role])

    content {
      name         = "aws_default_logs_role"
      value        = parameter.value
      apply_method = "immediate"
    }
  }

  dynamic "parameter" {
    for_each = var.extra_immediate_cluster_parameter
    content {
      name         = parameter.key
      value        = parameter.value
      apply_method = "immediate"
    }
  }

  dynamic "parameter" {
    for_each = var.extra_pending_cluster_parameter
    content {
      name         = parameter.key
      value        = parameter.value
      apply_method = "pending-reboot"
    }
  }
}
`, rName)
}
