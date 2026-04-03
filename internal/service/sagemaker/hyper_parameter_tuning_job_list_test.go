// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSageMakerHyperParameterTuningJob_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_sagemaker_hyper_parameter_tuning_job.test[0]"
	resourceName2 := "aws_sagemaker_hyper_parameter_tuning_job.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccHyperParameterTuningJobPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		CheckDestroy:             testAccCheckHyperParameterTuningJobDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/HyperParameterTuningJob/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(acctest.CtName), knownvalue.NotNull()),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(acctest.CtName), knownvalue.NotNull()),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/HyperParameterTuningJob/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_sagemaker_hyper_parameter_tuning_job.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_sagemaker_hyper_parameter_tuning_job.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.NotNull()),
					tfquerycheck.ExpectNoResourceObject("aws_sagemaker_hyper_parameter_tuning_job.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_sagemaker_hyper_parameter_tuning_job.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_sagemaker_hyper_parameter_tuning_job.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.NotNull()),
					tfquerycheck.ExpectNoResourceObject("aws_sagemaker_hyper_parameter_tuning_job.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccSageMakerHyperParameterTuningJob_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_sagemaker_hyper_parameter_tuning_job.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccHyperParameterTuningJobPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		CheckDestroy:             testAccCheckHyperParameterTuningJobDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/HyperParameterTuningJob/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(acctest.CtName), knownvalue.NotNull()),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/HyperParameterTuningJob/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_sagemaker_hyper_parameter_tuning_job.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_sagemaker_hyper_parameter_tuning_job.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.NotNull()),
					querycheck.ExpectResourceKnownValues("aws_sagemaker_hyper_parameter_tuning_job.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(acctest.CtName), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("config"), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectPartial(map[string]knownvalue.Check{
								"strategy": knownvalue.StringExact("Bayesian"),
								"objective": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectPartial(map[string]knownvalue.Check{
										names.AttrMetricName: knownvalue.StringExact("validation:accuracy"),
										names.AttrType:       knownvalue.StringExact("Maximize"),
									}),
								}),
								"resource_limits": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectPartial(map[string]knownvalue.Check{
										"max_number_of_training_jobs": knownvalue.Int64Exact(2),
										"max_parallel_training_jobs":  knownvalue.Int64Exact(1),
									}),
								}),
							}),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("training_job_definition"), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectPartial(map[string]knownvalue.Check{
								names.AttrRoleARN: knownvalue.NotNull(),
								"algorithm_specification": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectPartial(map[string]knownvalue.Check{
										"algorithm_name":      knownvalue.NotNull(),
										"training_input_mode": knownvalue.StringExact("File"),
									}),
								}),
							}),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
					}),
				},
			},
		},
	})
}

func TestAccSageMakerHyperParameterTuningJob_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_sagemaker_hyper_parameter_tuning_job.test[0]"
	resourceName2 := "aws_sagemaker_hyper_parameter_tuning_job.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccHyperParameterTuningJobPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		CheckDestroy:             testAccCheckHyperParameterTuningJobDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/HyperParameterTuningJob/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(acctest.CtName), knownvalue.NotNull()),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(acctest.CtName), knownvalue.NotNull()),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/HyperParameterTuningJob/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_sagemaker_hyper_parameter_tuning_job.test", identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_sagemaker_hyper_parameter_tuning_job.test", identity2.Checks()),
				},
			},
		},
	})
}
