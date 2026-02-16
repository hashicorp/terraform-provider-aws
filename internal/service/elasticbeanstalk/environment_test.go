// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elasticbeanstalk_test

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"slices"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfelasticbeanstalk "github.com/hashicorp/terraform-provider-aws/internal/service/elasticbeanstalk"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElasticBeanstalkEnvironment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"
	beanstalkAsgNameRegexp := regexache.MustCompile("awseb.+?AutoScalingGroup[^,]+")
	beanstalkElbNameRegexp := regexache.MustCompile("awseb.+?EBLoa[^,]+")
	beanstalkInstancesNameRegexp := regexache.MustCompile("i-([0-9A-Fa-f]{8}|[0-9A-Fa-f]{17})")
	beanstalkLcNameRegexp := regexache.MustCompile("awseb.+?AutoScalingLaunch[^,]+")
	beanstalkEndpointURL := regexache.MustCompile("awseb.+?EBLoa[^,].+?elb.amazonaws.com")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "elasticbeanstalk", "environment/{application}/{name}"),
					resource.TestMatchResourceAttr(resourceName, "autoscaling_groups.0", beanstalkAsgNameRegexp),
					resource.TestMatchResourceAttr(resourceName, "endpoint_url", beanstalkEndpointURL),
					resource.TestMatchResourceAttr(resourceName, "instances.0", beanstalkInstancesNameRegexp),
					resource.TestMatchResourceAttr(resourceName, "launch_configurations.0", beanstalkLcNameRegexp),
					resource.TestMatchResourceAttr(resourceName, "load_balancers.0", beanstalkElbNameRegexp),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("setting"), knownvalue.SetExact(settingsChecks_basic())),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("all_settings"), knownvalue.SetPartial(settingsChecks_basic())),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"setting",
					"wait_for_ready_timeout",
				},
			},
			{
				Config: testAccEnvironmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccElasticBeanstalkEnvironment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
					acctest.CheckSDKResourceDisappears(ctx, t, tfelasticbeanstalk.ResourceEnvironment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccElasticBeanstalkEnvironment_tier(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"
	beanstalkQueuesNameRegexp := regexache.MustCompile("https://sqs.+?awseb[^,]+")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_worker(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
					resource.TestMatchResourceAttr(resourceName, "queues.0", beanstalkQueuesNameRegexp),
					resource.TestCheckResourceAttr(resourceName, "tier", "Worker"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"setting",
					"wait_for_ready_timeout",
				},
			},
		},
	})
}

func TestAccElasticBeanstalkEnvironment_cnamePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"
	beanstalkCnameRegexp := regexache.MustCompile("^" + rName + ".+?elasticbeanstalk.com$")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_cnamePrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
					resource.TestMatchResourceAttr(resourceName, "cname", beanstalkCnameRegexp),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"setting",
					"wait_for_ready_timeout",
				},
			},
		},
	})
}

func TestAccElasticBeanstalkEnvironment_beanstalkEnv(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_template(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
					testAccCheckEnvironmentConfigValue(ctx, t, resourceName, "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"setting",
					"template_name",
					"wait_for_ready_timeout",
				},
			},
			{
				Config: testAccEnvironmentConfig_template(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
					testAccCheckEnvironmentConfigValue(ctx, t, resourceName, "2"),
				),
			},
			{
				Config: testAccEnvironmentConfig_template(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
					testAccCheckEnvironmentConfigValue(ctx, t, resourceName, "3"),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkEnvironment_resource(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_resourceOptionSetting(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"setting",
					"wait_for_ready_timeout",
				},
			},
		},
	})
}

func TestAccElasticBeanstalkEnvironment_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"setting",
					"wait_for_ready_timeout",
				},
			},
			{
				Config: testAccEnvironmentConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccEnvironmentConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkEnvironment_changeStack(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_templateChangeStack(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
				),
			},
			{
				Config: testAccEnvironmentConfig_templateChangeTemp(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
				),
			},
			{
				Config: testAccEnvironmentConfig_templateChangeStack(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkEnvironment_update(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
					testAccVerifyConfig(ctx, t, &app, []string{}),
				),
			},
			{
				Config: testAccEnvironmentConfig_settings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
					testAccVerifyConfig(ctx, t, &app, []string{"ENV_STATIC", "ENV_UPDATE"}),
				),
			},
			{
				Config: testAccEnvironmentConfig_settings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
					testAccVerifyConfig(ctx, t, &app, []string{"ENV_STATIC", "ENV_UPDATE"}),
				),
			},
			{
				Config: testAccEnvironmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
					testAccVerifyConfig(ctx, t, &app, []string{}),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkEnvironment_label(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_applicationVersion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
					resource.TestCheckResourceAttrSet(resourceName, "version_label"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"setting",
					"wait_for_ready_timeout",
				},
			},
			{
				Config: testAccEnvironmentConfig_applicationVersionUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
					resource.TestCheckResourceAttrSet(resourceName, "version_label"),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkEnvironment_settingWithJSONValue(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"
	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)

	if err != nil {
		t.Fatalf("generating random SSH key: %s", err)
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_settingJSONValue(rName, publicKey, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"setting",
					"wait_for_ready_timeout",
				},
			},
		},
	})
}

func TestAccElasticBeanstalkEnvironment_platformARN(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"
	platformNameWithVersion1 := "Python 3.12 running on 64bit Amazon Linux 2023/4.7.2"
	rValue1 := acctest.RandIntRange(t, 1000, 2000)
	rValue1Str := strconv.Itoa(rValue1)
	platformNameWithVersion2 := "Python 3.13 running on 64bit Amazon Linux 2023/4.7.2"
	rValue2 := acctest.RandIntRange(t, 3000, 4000)
	rValue2Str := strconv.Itoa(rValue2)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_platformARN(rName, platformNameWithVersion1, rValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
					acctest.CheckResourceAttrRegionalARNNoAccount(resourceName, "platform_arn", "elasticbeanstalk", "platform/"+platformNameWithVersion1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "5"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", rValue1Str),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", rValue1Str),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", rValue1Str),
					resource.TestCheckResourceAttr(resourceName, "tags.Key4", rValue1Str),
					resource.TestCheckResourceAttr(resourceName, "tags.Key5", rValue1Str),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"setting",
					"wait_for_ready_timeout",
				},
			},
			{
				Config: testAccEnvironmentConfig_platformARN(rName, platformNameWithVersion2, rValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
					acctest.CheckResourceAttrRegionalARNNoAccount(resourceName, "platform_arn", "elasticbeanstalk", "platform/"+platformNameWithVersion2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "5"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", rValue2Str),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", rValue2Str),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", rValue2Str),
					resource.TestCheckResourceAttr(resourceName, "tags.Key4", rValue2Str),
					resource.TestCheckResourceAttr(resourceName, "tags.Key5", rValue2Str),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkEnvironment_migrate_settingsResourceDefault(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		CheckDestroy: testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "6.14.1",
					},
				},
				Config: testAccEnvironmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("setting"), knownvalue.SetExact(settingsChecks_basic())),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccEnvironmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccElasticBeanstalkEnvironment_taint(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	value1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	value2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_setting_ComputedValue(rName, value1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("setting"), knownvalue.SetExact(settingsChecks_ValueChanged(value1))),
				},
			},
			{
				Taint:  []string{"terraform_data.test"},
				Config: testAccEnvironmentConfig_setting_ComputedValue(rName, value2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectResourceAction("terraform_data.test", plancheck.ResourceActionReplace),

						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("setting"), knownvalue.SetExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrNamespace: knownvalue.StringExact("aws:ec2:vpc"),
								names.AttrName:      knownvalue.StringExact("Subnets"),
								"resource":          knownvalue.StringExact(""),
								// "value":          Unknown value,
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrNamespace: knownvalue.StringExact("aws:ec2:vpc"),
								names.AttrName:      knownvalue.StringExact("AssociatePublicIpAddress"),
								"resource":          knownvalue.StringExact(""),
								names.AttrValue:     knownvalue.StringExact(acctest.CtTrue),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrNamespace: knownvalue.StringExact("aws:autoscaling:launchconfiguration"),
								names.AttrName:      knownvalue.StringExact("IamInstanceProfile"),
								"resource":          knownvalue.StringExact(""),
								names.AttrValue:     knownvalue.NotNull(), // Pair: aws_iam_instance_profile.test.name
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrNamespace: knownvalue.StringExact("aws:elasticbeanstalk:application:environment"),
								names.AttrName:      knownvalue.StringExact("ENV_TEST"),
								"resource":          knownvalue.StringExact(""),
								// "value":          Unknown value,
							}),
						})),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("setting"), knownvalue.SetExact(settingsChecks_ValueChanged(value2))),
				},
			},
		},
	})
}

func TestAccElasticBeanstalkEnvironment_setting_ComputedValue(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	value1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	value2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_setting_ComputedValue(rName, value1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("setting"), knownvalue.SetExact(settingsChecks_ValueChanged(value1))),
				},
			},
			{
				Config: testAccEnvironmentConfig_setting_ComputedValue(rName, value2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectResourceAction("terraform_data.test", plancheck.ResourceActionUpdate),

						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("setting"), knownvalue.SetExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrNamespace: knownvalue.StringExact("aws:ec2:vpc"),
								names.AttrName:      knownvalue.StringExact("Subnets"),
								"resource":          knownvalue.StringExact(""),
								// "value":          Unknown value,
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrNamespace: knownvalue.StringExact("aws:ec2:vpc"),
								names.AttrName:      knownvalue.StringExact("AssociatePublicIpAddress"),
								"resource":          knownvalue.StringExact(""),
								names.AttrValue:     knownvalue.StringExact(acctest.CtTrue),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrNamespace: knownvalue.StringExact("aws:autoscaling:launchconfiguration"),
								names.AttrName:      knownvalue.StringExact("IamInstanceProfile"),
								"resource":          knownvalue.StringExact(""),
								names.AttrValue:     knownvalue.NotNull(), // Pair: aws_iam_instance_profile.test.name
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrNamespace: knownvalue.StringExact("aws:elasticbeanstalk:application:environment"),
								names.AttrName:      knownvalue.StringExact("ENV_TEST"),
								"resource":          knownvalue.StringExact(""),
								// "value":          Unknown value,
							}),
						})),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("setting"), knownvalue.SetExact(settingsChecks_ValueChanged(value2))),
				},
			},
		},
	})
}

func TestAccElasticBeanstalkEnvironment_setting_ForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	value1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	value2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_setting_ForceNew(rName, value1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("setting"), knownvalue.SetExact(settingsChecks_ValueChanged(value1))),
				},
			},
			{
				Config: testAccEnvironmentConfig_setting_ForceNew(rName, value2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &app),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectResourceAction("terraform_data.test", plancheck.ResourceActionReplace),

						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("setting"), knownvalue.SetExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrNamespace: knownvalue.StringExact("aws:ec2:vpc"),
								names.AttrName:      knownvalue.StringExact("Subnets"),
								"resource":          knownvalue.StringExact(""),
								// "value":          Unknown value,
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrNamespace: knownvalue.StringExact("aws:ec2:vpc"),
								names.AttrName:      knownvalue.StringExact("AssociatePublicIpAddress"),
								"resource":          knownvalue.StringExact(""),
								names.AttrValue:     knownvalue.StringExact(acctest.CtTrue),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrNamespace: knownvalue.StringExact("aws:autoscaling:launchconfiguration"),
								names.AttrName:      knownvalue.StringExact("IamInstanceProfile"),
								"resource":          knownvalue.StringExact(""),
								names.AttrValue:     knownvalue.NotNull(), // Pair: aws_iam_instance_profile.test.name
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrNamespace: knownvalue.StringExact("aws:elasticbeanstalk:application:environment"),
								names.AttrName:      knownvalue.StringExact("ENV_TEST"),
								"resource":          knownvalue.StringExact(""),
								// "value":          Unknown value,
							}),
						})),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("setting"), knownvalue.SetExact(settingsChecks_ValueChanged(value2))),
				},
			},
		},
	})
}

func testAccCheckEnvironmentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ElasticBeanstalkClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_elastic_beanstalk_environment" {
				continue
			}

			_, err := tfelasticbeanstalk.FindEnvironmentByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Elastic Beanstalk Environment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckEnvironmentExists(ctx context.Context, t *testing.T, n string, v *awstypes.EnvironmentDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Elastic Beanstalk Environment ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).ElasticBeanstalkClient(ctx)

		output, err := tfelasticbeanstalk.FindEnvironmentByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVerifyConfig(ctx context.Context, t *testing.T, env *awstypes.EnvironmentDescription, expected []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if env == nil {
			return fmt.Errorf("Nil environment in testAccVerifyConfig")
		}
		conn := acctest.ProviderMeta(ctx, t).ElasticBeanstalkClient(ctx)

		resp, err := conn.DescribeConfigurationSettings(ctx, &elasticbeanstalk.DescribeConfigurationSettingsInput{
			ApplicationName: env.ApplicationName,
			EnvironmentName: env.EnvironmentName,
		})

		if err != nil {
			return fmt.Errorf("Error describing config settings in testAccVerifyConfig: %w", err)
		}

		// should only be 1 environment
		if len(resp.ConfigurationSettings) != 1 {
			return fmt.Errorf("Expected only 1 set of Configuration Settings in testAccVerifyConfig, got (%d)", len(resp.ConfigurationSettings))
		}

		cs := resp.ConfigurationSettings[0]

		var foundEnvs []string
		testStrings := []string{"ENV_STATIC", "ENV_UPDATE"}
		for _, os := range cs.OptionSettings {
			for _, k := range testStrings {
				if *os.OptionName == k {
					foundEnvs = append(foundEnvs, k)
				}
			}
		}

		// if expected is zero, then we should not have found any of the predefined
		// env vars
		if len(expected) == 0 {
			if len(foundEnvs) > 0 {
				return fmt.Errorf("Found configs we should not have: %#v", foundEnvs)
			}
			return nil
		}

		slices.Sort(testStrings)
		slices.Sort(expected)
		if !reflect.DeepEqual(testStrings, expected) {
			return fmt.Errorf("error matching strings, expected:\n\n%#v\n\ngot:\n\n%#v", testStrings, foundEnvs)
		}

		return nil
	}
}

func testAccCheckEnvironmentConfigValue(ctx context.Context, t *testing.T, n string, expectedValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ElasticBeanstalkClient(ctx)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Elastic Beanstalk ENV is not set")
		}

		resp, err := conn.DescribeConfigurationOptions(ctx, &elasticbeanstalk.DescribeConfigurationOptionsInput{
			ApplicationName: aws.String(rs.Primary.Attributes["application"]),
			EnvironmentName: aws.String(rs.Primary.Attributes[names.AttrName]),
			Options: []awstypes.OptionSpecification{
				{
					Namespace:  aws.String("aws:elasticbeanstalk:application:environment"),
					OptionName: aws.String("TEMPLATE"),
				},
			},
		})
		if err != nil {
			return err
		}

		if len(resp.Options) != 1 {
			return fmt.Errorf("Found %d options, expected 1.", len(resp.Options))
		}

		log.Printf("[DEBUG] %d Elastic Beanstalk Option values returned.", len(resp.Options[0].ValueOptions))

		for _, value := range resp.Options[0].ValueOptions {
			if value != expectedValue {
				return fmt.Errorf("Option setting value: %s. Expected %s", value, expectedValue)
			}
		}

		return nil
	}
}

func testAccEnvironmentConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
data "aws_elastic_beanstalk_solution_stack" "test" {
  most_recent = true
  name_regex  = "64bit Amazon Linux .* running Python .*"
}

data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.test.id
  route_table_id         = aws_vpc.test.main_route_table_id
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_elastic_beanstalk_application" "test" {
  description = "tf-test-desc"
  name        = %[1]q
}

# Create custom service role per test to remove dependency on
# Service-Linked Role existing.
resource "aws_iam_role" "service_role" {
  name = "%[1]s-service"
  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Condition = {
        StringEquals = {
          "sts:ExternalId" = "elasticbeanstalk"
        }
      }
      Effect = "Allow"
      Principal = {
        Service = "elasticbeanstalk.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role_policy_attachment" "service_role-AWSElasticBeanstalkEnhancedHealth" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSElasticBeanstalkEnhancedHealth"
  role       = aws_iam_role.service_role.id
}

resource "aws_iam_role_policy_attachment" "service_role-AWSElasticBeanstalkService" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSElasticBeanstalkService"
  role       = aws_iam_role.service_role.id
}

# Amazon Linux 2 environments require IAM Instance Profile.
resource "aws_iam_role" "instance_profile" {
  name = "%[1]s-instance"
  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role_policy_attachment" "instance_profile-AWSElasticBeanstalkWebTier" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AWSElasticBeanstalkWebTier"
  role       = aws_iam_role.instance_profile.id
}

resource "aws_iam_role_policy_attachment" "instance_profile-AWSElasticBeanstalkWorkerTier" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AWSElasticBeanstalkWorkerTier"
  role       = aws_iam_role.instance_profile.id
}

# Since the IAM Instance Profile is required anyways, also use it to
# ensure IAM Role permissions for both roles are attached.
resource "aws_iam_instance_profile" "test" {
  depends_on = [
    aws_iam_role_policy_attachment.instance_profile-AWSElasticBeanstalkWebTier,
    aws_iam_role_policy_attachment.instance_profile-AWSElasticBeanstalkWorkerTier,
    aws_iam_role_policy_attachment.service_role-AWSElasticBeanstalkEnhancedHealth,
    aws_iam_role_policy_attachment.service_role-AWSElasticBeanstalkService,
  ]

  name = %[1]q
  role = aws_iam_role.instance_profile.name
}
`, rName))
}

func testAccEnvironmentConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_elastic_beanstalk_environment" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  name                = %[1]q
  solution_stack_name = data.aws_elastic_beanstalk_solution_stack.test.name

  setting {
    namespace = "aws:ec2:vpc"
    name      = "VPCId"
    value     = aws_vpc.test.id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "Subnets"
    value     = aws_subnet.test[0].id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }
}
`, rName))
}

func settingsChecks_basic() []knownvalue.Check {
	return []knownvalue.Check{
		knownvalue.ObjectExact(map[string]knownvalue.Check{
			names.AttrNamespace: knownvalue.StringExact("aws:ec2:vpc"),
			names.AttrName:      knownvalue.StringExact("VPCId"),
			"resource":          knownvalue.StringExact(""),
			names.AttrValue:     knownvalue.NotNull(), // Pair: aws_vpc.test.id
		}),

		knownvalue.ObjectExact(map[string]knownvalue.Check{
			names.AttrNamespace: knownvalue.StringExact("aws:ec2:vpc"),
			names.AttrName:      knownvalue.StringExact("Subnets"),
			"resource":          knownvalue.StringExact(""),
			names.AttrValue:     knownvalue.NotNull(), // Pair: aws_subnet.test[0].id
		}),

		knownvalue.ObjectExact(map[string]knownvalue.Check{
			names.AttrNamespace: knownvalue.StringExact("aws:ec2:vpc"),
			names.AttrName:      knownvalue.StringExact("AssociatePublicIpAddress"),
			"resource":          knownvalue.StringExact(""),
			names.AttrValue:     knownvalue.StringExact(acctest.CtTrue),
		}),

		knownvalue.ObjectExact(map[string]knownvalue.Check{
			names.AttrNamespace: knownvalue.StringExact("aws:autoscaling:launchconfiguration"),
			names.AttrName:      knownvalue.StringExact("SecurityGroups"),
			"resource":          knownvalue.StringExact(""),
			names.AttrValue:     knownvalue.NotNull(), // Pair: aws_security_group.test.id
		}),

		knownvalue.ObjectExact(map[string]knownvalue.Check{
			names.AttrNamespace: knownvalue.StringExact("aws:autoscaling:launchconfiguration"),
			names.AttrName:      knownvalue.StringExact("IamInstanceProfile"),
			"resource":          knownvalue.StringExact(""),
			names.AttrValue:     knownvalue.NotNull(), // Pair: aws_iam_instance_profile.test.name
		}),

		knownvalue.ObjectExact(map[string]knownvalue.Check{
			names.AttrNamespace: knownvalue.StringExact("aws:elasticbeanstalk:environment"),
			names.AttrName:      knownvalue.StringExact("ServiceRole"),
			"resource":          knownvalue.StringExact(""),
			names.AttrValue:     knownvalue.NotNull(), // Pair: aws_iam_role.service_role.name
		}),
	}
}

func testAccEnvironmentConfig_platformARN(rName, platformNameWithVersion string, rValue int) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_elastic_beanstalk_environment" "test" {
  application  = aws_elastic_beanstalk_application.test.name
  name         = %[1]q
  platform_arn = "arn:${data.aws_partition.current.partition}:elasticbeanstalk:${data.aws_region.current.region}::platform/%[2]s"

  setting {
    namespace = "aws:ec2:vpc"
    name      = "VPCId"
    value     = aws_vpc.test.id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "Subnets"
    value     = aws_subnet.test[0].id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }

  tags = {
    Key1 = "%[3]d"
    Key2 = "%[3]d"
    Key3 = "%[3]d"
    Key4 = "%[3]d"
    Key5 = "%[3]d"
  }
}
`, rName, platformNameWithVersion, rValue))
}

func testAccEnvironmentConfig_settings(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_elastic_beanstalk_environment" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  name                = %[1]q
  solution_stack_name = data.aws_elastic_beanstalk_solution_stack.test.name

  setting {
    namespace = "aws:ec2:vpc"
    name      = "VPCId"
    value     = aws_vpc.test.id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "Subnets"
    value     = aws_subnet.test[0].id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:application:environment"
    name      = "ENV_STATIC"
    value     = "true"
  }

  setting {
    namespace = "aws:elasticbeanstalk:application:environment"
    name      = "ENV_UPDATE"
    value     = "true"
  }

  setting {
    namespace = "aws:elasticbeanstalk:application:environment"
    name      = "ENV_REMOVE"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:scheduledaction"
    resource  = "ScheduledAction01"
    name      = "MinSize"
    value     = 2
  }

  setting {
    namespace = "aws:autoscaling:scheduledaction"
    resource  = "ScheduledAction01"
    name      = "MaxSize"
    value     = 3
  }

  setting {
    namespace = "aws:autoscaling:scheduledaction"
    resource  = "ScheduledAction01"
    name      = "StartTime"
    value     = "2016-07-28T04:07:02Z"
  }
}
`, rName))
}

func testAccEnvironmentConfig_worker(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_elastic_beanstalk_environment" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  name                = %[1]q
  solution_stack_name = data.aws_elastic_beanstalk_solution_stack.test.name
  tier                = "Worker"

  setting {
    namespace = "aws:ec2:vpc"
    name      = "VPCId"
    value     = aws_vpc.test.id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "Subnets"
    value     = aws_subnet.test[0].id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }
}
`, rName))
}

func testAccEnvironmentConfig_cnamePrefix(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_elastic_beanstalk_environment" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  cname_prefix        = %[1]q
  name                = %[1]q
  solution_stack_name = data.aws_elastic_beanstalk_solution_stack.test.name

  setting {
    namespace = "aws:ec2:vpc"
    name      = "VPCId"
    value     = aws_vpc.test.id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "Subnets"
    value     = aws_subnet.test[0].id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }
}
`, rName))
}

func testAccEnvironmentConfig_template(rName string, cfgTplValue int) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_elastic_beanstalk_environment" "test" {
  application   = aws_elastic_beanstalk_application.test.name
  name          = %[1]q
  template_name = aws_elastic_beanstalk_configuration_template.test.name
}

resource "aws_elastic_beanstalk_configuration_template" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  name                = %[1]q
  solution_stack_name = data.aws_elastic_beanstalk_solution_stack.test.name

  setting {
    namespace = "aws:ec2:vpc"
    name      = "VPCId"
    value     = aws_vpc.test.id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "Subnets"
    value     = aws_subnet.test[0].id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:application:environment"
    name      = "TEMPLATE"
    value     = %[2]d
  }
}
`, rName, cfgTplValue))
}

func testAccEnvironmentConfig_resourceOptionSetting(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_elastic_beanstalk_environment" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  name                = %[1]q
  solution_stack_name = data.aws_elastic_beanstalk_solution_stack.test.name

  setting {
    namespace = "aws:ec2:vpc"
    name      = "VPCId"
    value     = aws_vpc.test.id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "Subnets"
    value     = aws_subnet.test[0].id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }

  setting {
    namespace = "aws:autoscaling:scheduledaction"
    resource  = "ScheduledAction01"
    name      = "MinSize"
    value     = "2"
  }

  setting {
    namespace = "aws:autoscaling:scheduledaction"
    resource  = "ScheduledAction01"
    name      = "MaxSize"
    value     = "6"
  }

  setting {
    namespace = "aws:autoscaling:scheduledaction"
    resource  = "ScheduledAction01"
    name      = "Recurrence"
    value     = "0 8 * * *"
  }
}
`, rName))
}

func testAccEnvironmentConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_elastic_beanstalk_environment" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  name                = %[1]q
  solution_stack_name = data.aws_elastic_beanstalk_solution_stack.test.name

  setting {
    namespace = "aws:ec2:vpc"
    name      = "VPCId"
    value     = aws_vpc.test.id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "Subnets"
    value     = aws_subnet.test[0].id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccEnvironmentConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_elastic_beanstalk_environment" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  name                = %[1]q
  solution_stack_name = data.aws_elastic_beanstalk_solution_stack.test.name

  setting {
    namespace = "aws:ec2:vpc"
    name      = "VPCId"
    value     = aws_vpc.test.id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "Subnets"
    value     = aws_subnet.test[0].id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccEnvironmentConfig_templateChangeStack(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_elastic_beanstalk_environment" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  name                = %[1]q
  solution_stack_name = data.aws_elastic_beanstalk_solution_stack.test.name

  setting {
    namespace = "aws:ec2:vpc"
    name      = "VPCId"
    value     = aws_vpc.test.id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "Subnets"
    value     = aws_subnet.test[0].id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }
}

resource "aws_elastic_beanstalk_configuration_template" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  name                = %[1]q
  solution_stack_name = data.aws_elastic_beanstalk_solution_stack.test.name
}
`, rName))
}

func testAccEnvironmentConfig_templateChangeTemp(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_elastic_beanstalk_environment" "test" {
  application   = aws_elastic_beanstalk_application.test.name
  name          = %[1]q
  template_name = aws_elastic_beanstalk_configuration_template.test.name

  setting {
    namespace = "aws:ec2:vpc"
    name      = "VPCId"
    value     = aws_vpc.test.id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "Subnets"
    value     = aws_subnet.test[0].id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }
}

resource "aws_elastic_beanstalk_configuration_template" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  name                = %[1]q
  solution_stack_name = data.aws_elastic_beanstalk_solution_stack.test.name
}
`, rName))
}

func testAccEnvironmentConfig_applicationVersion(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = "python-v1.zip"
  source = "test-fixtures/python-v1.zip"
}

resource "aws_elastic_beanstalk_application_version" "test" {
  application = aws_elastic_beanstalk_application.test.name
  bucket      = aws_s3_object.test.bucket
  key         = aws_s3_object.test.key
  name        = "%[1]s-1"
}

resource "aws_elastic_beanstalk_environment" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  name                = %[1]q
  solution_stack_name = data.aws_elastic_beanstalk_solution_stack.test.name
  version_label       = aws_elastic_beanstalk_application_version.test.name

  setting {
    namespace = "aws:ec2:vpc"
    name      = "VPCId"
    value     = aws_vpc.test.id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "Subnets"
    value     = aws_subnet.test[0].id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }
}
`, rName))
}

func testAccEnvironmentConfig_applicationVersionUpdate(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = "python-v1.zip"
  source = "test-fixtures/python-v1.zip"
}

resource "aws_elastic_beanstalk_application_version" "test" {
  application = aws_elastic_beanstalk_application.test.name
  bucket      = aws_s3_object.test.bucket
  key         = aws_s3_object.test.key
  name        = "%[1]s-2"
}

resource "aws_elastic_beanstalk_environment" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  name                = %[1]q
  solution_stack_name = data.aws_elastic_beanstalk_solution_stack.test.name
  version_label       = aws_elastic_beanstalk_application_version.test.name

  setting {
    namespace = "aws:ec2:vpc"
    name      = "VPCId"
    value     = aws_vpc.test.id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "Subnets"
    value     = aws_subnet.test[0].id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }
}
`, rName))
}

func testAccEnvironmentConfig_settingJSONValue(rName, publicKey, email string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = %[2]q
}

resource "aws_elastic_beanstalk_environment" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  name                = %[1]q
  solution_stack_name = data.aws_elastic_beanstalk_solution_stack.test.name
  tier                = "Worker"

  setting {
    namespace = "aws:ec2:vpc"
    name      = "VPCId"
    value     = aws_vpc.test.id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "Subnets"
    value     = aws_subnet.test[0].id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:elasticbeanstalk:command"
    name      = "BatchSize"
    value     = "30"
  }

  setting {
    namespace = "aws:elasticbeanstalk:command"
    name      = "BatchSizeType"
    value     = "Percentage"
  }

  setting {
    namespace = "aws:elasticbeanstalk:command"
    name      = "DeploymentPolicy"
    value     = "Rolling"
  }

  setting {
    namespace = "aws:elasticbeanstalk:sns:topics"
    name      = "Notification Endpoint"
    value     = %[3]q
  }

  setting {
    namespace = "aws:elasticbeanstalk:sqsd"
    name      = "ErrorVisibilityTimeout"
    value     = "2"
  }

  setting {
    namespace = "aws:elasticbeanstalk:sqsd"
    name      = "HttpPath"
    value     = "/event-message"
  }

  setting {
    namespace = "aws:elasticbeanstalk:sqsd"
    name      = "WorkerQueueURL"
    value     = aws_sqs_queue.test.id
  }

  setting {
    namespace = "aws:elasticbeanstalk:sqsd"
    name      = "VisibilityTimeout"
    value     = "300"
  }

  setting {
    namespace = "aws:elasticbeanstalk:sqsd"
    name      = "HttpConnections"
    value     = "10"
  }

  setting {
    namespace = "aws:elasticbeanstalk:sqsd"
    name      = "InactivityTimeout"
    value     = "299"
  }

  setting {
    namespace = "aws:elasticbeanstalk:sqsd"
    name      = "MimeType"
    value     = "application/json"
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "EnvironmentType"
    value     = "LoadBalanced"
  }

  setting {
    namespace = "aws:elasticbeanstalk:application"
    name      = "Application Healthcheck URL"
    value     = "/health"
  }

  setting {
    namespace = "aws:elasticbeanstalk:healthreporting:system"
    name      = "SystemType"
    value     = "enhanced"
  }

  setting {
    namespace = "aws:elasticbeanstalk:healthreporting:system"
    name      = "HealthCheckSuccessThreshold"
    value     = "Ok"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "InstanceType"
    value     = "t2.micro"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "EC2KeyName"
    value     = aws_key_pair.test.key_name
  }

  setting {
    namespace = "aws:autoscaling:updatepolicy:rollingupdate"
    name      = "RollingUpdateEnabled"
    value     = "false"
  }

  setting {
    namespace = "aws:elasticbeanstalk:healthreporting:system"
    name      = "ConfigDocument"

    value = <<EOF
{
	"Version": 1,
	"CloudWatchMetrics": {
		"Instance": {
			"ApplicationRequestsTotal": 60
		},
		"Environment": {
			"ApplicationRequests5xx": 60,
			"ApplicationRequests4xx": 60,
			"ApplicationRequests2xx": 60
		}
	}
}
EOF
  }
}
`, rName, publicKey, email))
}

func testAccEnvironmentConfig_setting_ComputedValue(rName, value string) string {
	return acctest.ConfigCompose(
		testAccEnvironmentConfig_setting_ValueChange(rName),
		fmt.Sprintf(`
resource "terraform_data" "test" {
  input = %[1]q
}
`, value))
}

func testAccEnvironmentConfig_setting_ForceNew(rName, value string) string {
	return acctest.ConfigCompose(
		testAccEnvironmentConfig_setting_ValueChange(rName),
		fmt.Sprintf(`
resource "terraform_data" "test" {
  input            = %[2]q
  triggers_replace = [%[2]q]
}
`, rName, value))
}

func testAccEnvironmentConfig_setting_ValueChange(rName string) string {
	return acctest.ConfigCompose(
		testAccEnvironmentConfig_base(rName),
		fmt.Sprintf(`
resource "aws_elastic_beanstalk_environment" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  name                = %[1]q
  solution_stack_name = data.aws_elastic_beanstalk_solution_stack.test.name

  setting {
    namespace = "aws:ec2:vpc"
    name      = "Subnets"
    # This contrived example is a simple way to trigger the error with computed values.
    # It should not be used in production configurations.
    value = replace("${aws_subnet.test[0].id}${terraform_data.test.output}", terraform_data.test.output, "")
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:application:environment"
    name      = "ENV_TEST"
    value     = terraform_data.test.output
  }
}
`, rName))
}

func settingsChecks_ValueChanged(envVal string) []knownvalue.Check {
	return []knownvalue.Check{
		knownvalue.ObjectExact(map[string]knownvalue.Check{
			names.AttrNamespace: knownvalue.StringExact("aws:ec2:vpc"),
			names.AttrName:      knownvalue.StringExact("Subnets"),
			"resource":          knownvalue.StringExact(""),
			names.AttrValue:     knownvalue.NotNull(), // Pair: aws_subnet.test[0].id
		}),

		knownvalue.ObjectExact(map[string]knownvalue.Check{
			names.AttrNamespace: knownvalue.StringExact("aws:ec2:vpc"),
			names.AttrName:      knownvalue.StringExact("AssociatePublicIpAddress"),
			"resource":          knownvalue.StringExact(""),
			names.AttrValue:     knownvalue.StringExact(acctest.CtTrue),
		}),

		knownvalue.ObjectExact(map[string]knownvalue.Check{
			names.AttrNamespace: knownvalue.StringExact("aws:autoscaling:launchconfiguration"),
			names.AttrName:      knownvalue.StringExact("IamInstanceProfile"),
			"resource":          knownvalue.StringExact(""),
			names.AttrValue:     knownvalue.NotNull(), // Pair: aws_iam_instance_profile.test.name
		}),

		knownvalue.ObjectExact(map[string]knownvalue.Check{
			names.AttrNamespace: knownvalue.StringExact("aws:elasticbeanstalk:application:environment"),
			names.AttrName:      knownvalue.StringExact("ENV_TEST"),
			"resource":          knownvalue.StringExact(""),
			names.AttrValue:     knownvalue.StringExact(envVal),
		}),
	}
}
