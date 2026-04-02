// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appsync_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfappsync "github.com/hashicorp/terraform-provider-aws/internal/service/appsync"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppSyncChannelNamespace_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ChannelNamespace
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_channel_namespace.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelNamespaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelNamespaceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckChannelNamespaceExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("channel_namespace_arn"), tfknownvalue.RegionalARNRegexp("appsync", regexache.MustCompile(`apis/.+/channelNamespace/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("code_handlers"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("handler_configs"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("publish_auth_mode"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("subscribe_auth_mode"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccChannelNamespaceImportStateID(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccAppSyncChannelNamespace_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ChannelNamespace
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appsync_channel_namespace.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelNamespaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelNamespaceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckChannelNamespaceExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfappsync.ResourceChannelNamespace, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccAppSyncChannelNamespace_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ChannelNamespace
	rName := fmt.Sprintf("tfacctest%d", acctest.RandInt(t))
	resourceName := "aws_appsync_channel_namespace.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelNamespaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelNamespaceConfig_comprehensive(rName, awstypes.InvokeTypeEvent),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckChannelNamespaceExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("channel_namespace_arn"), tfknownvalue.RegionalARNRegexp("appsync", regexache.MustCompile(`apis/.+/channelNamespace/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("code_handlers"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("handler_configs"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.MapExact(map[string]knownvalue.Check{
							"on_publish": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.MapExact(map[string]knownvalue.Check{
									"behavior": tfknownvalue.StringExact(awstypes.HandlerBehaviorDirect),
									"integration": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.MapExact(map[string]knownvalue.Check{
											"data_source_name": knownvalue.StringExact(rName),
											"lambda_config":    knownvalue.ListSizeExact(0),
										}),
									}),
								}),
							}),
							"on_subscribe": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.MapExact(map[string]knownvalue.Check{
									"behavior": tfknownvalue.StringExact(awstypes.HandlerBehaviorDirect),
									"integration": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.MapExact(map[string]knownvalue.Check{
											"data_source_name": knownvalue.StringExact(rName),
											"lambda_config": knownvalue.ListExact([]knownvalue.Check{
												knownvalue.MapExact(map[string]knownvalue.Check{
													"invoke_type": tfknownvalue.StringExact(awstypes.InvokeTypeEvent),
												}),
											}),
										}),
									}),
								}),
							}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("publish_auth_mode"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.MapExact(map[string]knownvalue.Check{
							"auth_type": tfknownvalue.StringExact(awstypes.AuthenticationTypeApiKey),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("subscribe_auth_mode"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.MapExact(map[string]knownvalue.Check{
							"auth_type": tfknownvalue.StringExact(awstypes.AuthenticationTypeApiKey),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccChannelNamespaceImportStateID(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
			{
				Config: testAccChannelNamespaceConfig_comprehensive(rName, awstypes.InvokeTypeRequestResponse),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckChannelNamespaceExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("api_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("channel_namespace_arn"), tfknownvalue.RegionalARNRegexp("appsync", regexache.MustCompile(`apis/.+/channelNamespace/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("code_handlers"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("handler_configs"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.MapExact(map[string]knownvalue.Check{
							"on_publish": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.MapExact(map[string]knownvalue.Check{
									"behavior": tfknownvalue.StringExact(awstypes.HandlerBehaviorDirect),
									"integration": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.MapExact(map[string]knownvalue.Check{
											"data_source_name": knownvalue.StringExact(rName),
											"lambda_config":    knownvalue.ListSizeExact(0),
										}),
									}),
								}),
							}),
							"on_subscribe": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.MapExact(map[string]knownvalue.Check{
									"behavior": tfknownvalue.StringExact(awstypes.HandlerBehaviorDirect),
									"integration": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.MapExact(map[string]knownvalue.Check{
											"data_source_name": knownvalue.StringExact(rName),
											"lambda_config": knownvalue.ListExact([]knownvalue.Check{
												knownvalue.MapExact(map[string]knownvalue.Check{
													"invoke_type": tfknownvalue.StringExact(awstypes.InvokeTypeRequestResponse),
												}),
											}),
										}),
									}),
								}),
							}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("publish_auth_mode"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.MapExact(map[string]knownvalue.Check{
							"auth_type": tfknownvalue.StringExact(awstypes.AuthenticationTypeApiKey),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("subscribe_auth_mode"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.MapExact(map[string]knownvalue.Check{
							"auth_type": tfknownvalue.StringExact(awstypes.AuthenticationTypeApiKey),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
		},
	})
}

func testAccCheckChannelNamespaceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AppSyncClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appsync_channel_namespace" {
				continue
			}

			_, err := tfappsync.FindChannelNamespaceByTwoPartKey(ctx, conn, rs.Primary.Attributes["api_id"], rs.Primary.Attributes[names.AttrName])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AppSync Channel Namespace %s still exists", rs.Primary.Attributes[names.AttrName])
		}

		return nil
	}
}

func testAccCheckChannelNamespaceExists(ctx context.Context, t *testing.T, n string, v *awstypes.ChannelNamespace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AppSyncClient(ctx)

		output, err := tfappsync.FindChannelNamespaceByTwoPartKey(ctx, conn, rs.Primary.Attributes["api_id"], rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccChannelNamespaceImportStateID(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		return acctest.AttrsImportStateIdFunc(n, ",", "api_id", names.AttrName)(s)
	}
}

func testAccChannelNamespaceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccAPIConfig_basic(rName), fmt.Sprintf(`
resource "aws_appsync_channel_namespace" "test" {
  name   = %[1]q
  api_id = aws_appsync_api.test.api_id
}
`, rName))
}

func testAccChannelNamespaceConfig_comprehensive(rName string, invokeType awstypes.InvokeType) string {
	return acctest.ConfigCompose(testAccAPIConfig_basic(rName), testAccDatasourceConfig_baseLambda(rName), fmt.Sprintf(`
resource "aws_appsync_datasource" "test" {
  api_id           = aws_appsync_api.test.api_id
  name             = %[1]q
  service_role_arn = aws_iam_role.test.arn
  type             = "AWS_LAMBDA"

  lambda_config {
    function_arn = aws_lambda_function.test.arn
  }
}

resource "aws_appsync_channel_namespace" "test" {
  name   = %[1]q
  api_id = aws_appsync_api.test.api_id

  handler_configs {
    on_publish {
      behavior = "DIRECT"

      integration {
        data_source_name = aws_appsync_datasource.test.name
      }
    }

    on_subscribe {
      behavior = "DIRECT"

      integration {
        data_source_name = aws_appsync_datasource.test.name

        lambda_config {
          invoke_type = %[2]q
        }
      }
    }
  }

  publish_auth_mode {
    auth_type = "API_KEY"
  }

  subscribe_auth_mode {
    auth_type = "API_KEY"
  }
}
`, rName, invokeType))
}
