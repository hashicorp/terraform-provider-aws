// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cognitoidp_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcognitoidp "github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidp"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCognitoIDPManagedLoginBranding_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ManagedLoginBrandingType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_managed_login_branding.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedLoginBrandingDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccManagedLoginBrandingConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedLoginBrandingExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("asset"), knownvalue.SetSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("managed_login_branding_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("settings"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("settings_all"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("use_cognito_provided_values"), knownvalue.Bool(true)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "managed_login_branding_id",
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", names.AttrUserPoolID, "managed_login_branding_id"),
			},
		},
	})
}

func TestAccCognitoIDPManagedLoginBranding_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.ManagedLoginBrandingType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_managed_login_branding.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedLoginBrandingDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccManagedLoginBrandingConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedLoginBrandingExists(ctx, t, resourceName, &client),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcognitoidp.ResourceManagedLoginBranding, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCognitoIDPManagedLoginBranding_asset(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ManagedLoginBrandingType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_managed_login_branding.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedLoginBrandingDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccManagedLoginBrandingConfig_asset(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedLoginBrandingExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("asset"), knownvalue.SetSizeExact(1)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "managed_login_branding_id",
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", names.AttrUserPoolID, "managed_login_branding_id"),
			},
		},
	})
}

func TestAccCognitoIDPManagedLoginBranding_settings(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ManagedLoginBrandingType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_managed_login_branding.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedLoginBrandingDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccManagedLoginBrandingConfig_settings(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedLoginBrandingExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("settings"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("use_cognito_provided_values"), knownvalue.Bool(false)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "managed_login_branding_id",
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", names.AttrUserPoolID, "managed_login_branding_id"),
			},
		},
	})
}

func TestAccCognitoIDPManagedLoginBranding_updateFromBasic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ManagedLoginBrandingType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_managed_login_branding.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedLoginBrandingDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccManagedLoginBrandingConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedLoginBrandingExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("settings"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("use_cognito_provided_values"), knownvalue.Bool(true)),
				},
			},
			{
				Config: testAccManagedLoginBrandingConfig_settings(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedLoginBrandingExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("settings"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("use_cognito_provided_values"), knownvalue.Bool(false)),
				},
			},
		},
	})
}

func TestAccCognitoIDPManagedLoginBranding_updateToBasic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ManagedLoginBrandingType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_managed_login_branding.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedLoginBrandingDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccManagedLoginBrandingConfig_settings(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedLoginBrandingExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("settings"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("use_cognito_provided_values"), knownvalue.Bool(false)),
				},
			},
			{
				Config: testAccManagedLoginBrandingConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedLoginBrandingExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("settings"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("use_cognito_provided_values"), knownvalue.Bool(true)),
				},
			},
		},
	})
}

func TestAccCognitoIDPManagedLoginBranding_updateSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ManagedLoginBrandingType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_managed_login_branding.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedLoginBrandingDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccManagedLoginBrandingConfig_settings(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedLoginBrandingExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("settings"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("use_cognito_provided_values"), knownvalue.Bool(false)),
				},
			},
			{
				Config: testAccManagedLoginBrandingConfig_settingsUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedLoginBrandingExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("settings"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("use_cognito_provided_values"), knownvalue.Bool(false)),
				},
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/44188.
func TestAccCognitoIDPManagedLoginBranding_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.ManagedLoginBrandingType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resource1Name := "aws_cognito_managed_login_branding.test1"
	resource2Name := "aws_cognito_managed_login_branding.test2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedLoginBrandingDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccManagedLoginBrandingConfig_multiple(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedLoginBrandingExists(ctx, t, resource1Name, &v1),
					testAccCheckManagedLoginBrandingExists(ctx, t, resource2Name, &v2),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resource1Name, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resource1Name, tfjsonpath.New("use_cognito_provided_values"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resource2Name, tfjsonpath.New("use_cognito_provided_values"), knownvalue.Bool(true)),
				},
			},
			{
				ResourceName:                         resource1Name,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "managed_login_branding_id",
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resource1Name, ",", names.AttrUserPoolID, "managed_login_branding_id"),
			},
			{
				ResourceName:                         resource2Name,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "managed_login_branding_id",
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resource2Name, ",", names.AttrUserPoolID, "managed_login_branding_id"),
			},
		},
	})
}

func testAccCheckManagedLoginBrandingDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CognitoIDPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cognito_managed_login_branding" {
				continue
			}

			_, err := tfcognitoidp.FindManagedLoginBrandingByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID], rs.Primary.Attributes["managed_login_branding_id"], false)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Cognito Managed Login Branding %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckManagedLoginBrandingExists(ctx context.Context, t *testing.T, n string, v *awstypes.ManagedLoginBrandingType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CognitoIDPClient(ctx)

		output, err := tfcognitoidp.FindManagedLoginBrandingByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID], rs.Primary.Attributes["managed_login_branding_id"], false)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccManagedLoginBrandingConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool_client" "test" {
  name                = %[1]q
  user_pool_id        = aws_cognito_user_pool.test.id
  explicit_auth_flows = ["ADMIN_NO_SRP_AUTH"]
}
`, rName)
}

func testAccManagedLoginBrandingConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccManagedLoginBrandingConfig_base(rName), `
resource "aws_cognito_managed_login_branding" "test" {
  client_id    = aws_cognito_user_pool_client.test.id
  user_pool_id = aws_cognito_user_pool.test.id

  use_cognito_provided_values = true
}
`)
}

func testAccManagedLoginBrandingConfig_asset(rName string) string {
	return acctest.ConfigCompose(testAccManagedLoginBrandingConfig_base(rName), `
resource "aws_cognito_managed_login_branding" "test" {
  client_id    = aws_cognito_user_pool_client.test.id
  user_pool_id = aws_cognito_user_pool.test.id

  use_cognito_provided_values = true

  asset {
    bytes      = filebase64("test-fixtures/login_branding_asset.svg")
    category   = "PAGE_FOOTER_BACKGROUND"
    color_mode = "DARK"
    extension  = "SVG"
  }
}
`)
}

func testAccManagedLoginBrandingConfig_settings(rName string) string {
	return acctest.ConfigCompose(testAccManagedLoginBrandingConfig_base(rName), `
resource "aws_cognito_managed_login_branding" "test" {
  client_id    = aws_cognito_user_pool_client.test.id
  user_pool_id = aws_cognito_user_pool.test.id

  settings = jsonencode({
    "categories" : {
      "auth" : {
        "authMethodOrder" : [
          [
            {
              "display" : "BUTTON",
              "type" : "FEDERATED"
            },
            {
              "display" : "INPUT",
              "type" : "USERNAME_PASSWORD"
            }
          ]
        ],
        "federation" : {
          "interfaceStyle" : "BUTTON_LIST",
          "order" : [
          ]
        }
      },
      "form" : {
        "displayGraphics" : true,
        "instructions" : {
          "enabled" : false
        },
        "languageSelector" : {
          "enabled" : false
        },
        "location" : {
          "horizontal" : "CENTER",
          "vertical" : "CENTER"
        },
        "sessionTimerDisplay" : "NONE"
      },
      "global" : {
        "colorSchemeMode" : "LIGHT",
        "pageFooter" : {
          "enabled" : false
        },
        "pageHeader" : {
          "enabled" : false
        },
        "spacingDensity" : "REGULAR"
      },
      "signUp" : {
        "acceptanceElements" : [
          {
            "enforcement" : "NONE",
            "textKey" : "en"
          }
        ]
      }
    },
    "componentClasses" : {
      "buttons" : {
        "borderRadius" : 8.0
      },
      "divider" : {
        "darkMode" : {
          "borderColor" : "232b37ff"
        },
        "lightMode" : {
          "borderColor" : "ebebf0ff"
        }
      },
      "dropDown" : {
        "borderRadius" : 8.0,
        "darkMode" : {
          "defaults" : {
            "itemBackgroundColor" : "192534ff"
          },
          "hover" : {
            "itemBackgroundColor" : "081120ff",
            "itemBorderColor" : "5f6b7aff",
            "itemTextColor" : "e9ebedff"
          },
          "match" : {
            "itemBackgroundColor" : "d1d5dbff",
            "itemTextColor" : "89bdeeff"
          }
        },
        "lightMode" : {
          "defaults" : {
            "itemBackgroundColor" : "ffffffff"
          },
          "hover" : {
            "itemBackgroundColor" : "f4f4f4ff",
            "itemBorderColor" : "7d8998ff",
            "itemTextColor" : "000716ff"
          },
          "match" : {
            "itemBackgroundColor" : "414d5cff",
            "itemTextColor" : "0972d3ff"
          }
        }
      },
      "focusState" : {
        "darkMode" : {
          "borderColor" : "539fe5ff"
        },
        "lightMode" : {
          "borderColor" : "0972d3ff"
        }
      },
      "idpButtons" : {
        "icons" : {
          "enabled" : true
        }
      },
      "input" : {
        "borderRadius" : 8.0,
        "darkMode" : {
          "defaults" : {
            "backgroundColor" : "0f1b2aff",
            "borderColor" : "5f6b7aff"
          },
          "placeholderColor" : "8d99a8ff"
        },
        "lightMode" : {
          "defaults" : {
            "backgroundColor" : "ffffffff",
            "borderColor" : "7d8998ff"
          },
          "placeholderColor" : "5f6b7aff"
        }
      },
      "inputDescription" : {
        "darkMode" : {
          "textColor" : "8d99a8ff"
        },
        "lightMode" : {
          "textColor" : "5f6b7aff"
        }
      },
      "inputLabel" : {
        "darkMode" : {
          "textColor" : "d1d5dbff"
        },
        "lightMode" : {
          "textColor" : "000716ff"
        }
      },
      "link" : {
        "darkMode" : {
          "defaults" : {
            "textColor" : "539fe5ff"
          },
          "hover" : {
            "textColor" : "89bdeeff"
          }
        },
        "lightMode" : {
          "defaults" : {
            "textColor" : "0972d3ff"
          },
          "hover" : {
            "textColor" : "033160ff"
          }
        }
      },
      "optionControls" : {
        "darkMode" : {
          "defaults" : {
            "backgroundColor" : "0f1b2aff",
            "borderColor" : "7d8998ff"
          },
          "selected" : {
            "backgroundColor" : "539fe5ff",
            "foregroundColor" : "000716ff"
          }
        },
        "lightMode" : {
          "defaults" : {
            "backgroundColor" : "ffffffff",
            "borderColor" : "7d8998ff"
          },
          "selected" : {
            "backgroundColor" : "0972d3ff",
            "foregroundColor" : "ffffffff"
          }
        }
      },
      "statusIndicator" : {
        "darkMode" : {
          "error" : {
            "backgroundColor" : "1a0000ff",
            "borderColor" : "eb6f6fff",
            "indicatorColor" : "eb6f6fff"
          },
          "pending" : {
            "indicatorColor" : "AAAAAAAA"
          },
          "success" : {
            "backgroundColor" : "001a02ff",
            "borderColor" : "29ad32ff",
            "indicatorColor" : "29ad32ff"
          },
          "warning" : {
            "backgroundColor" : "1d1906ff",
            "borderColor" : "e0ca57ff",
            "indicatorColor" : "e0ca57ff"
          }
        },
        "lightMode" : {
          "error" : {
            "backgroundColor" : "fff7f7ff",
            "borderColor" : "d91515ff",
            "indicatorColor" : "d91515ff"
          },
          "pending" : {
            "indicatorColor" : "AAAAAAAA"
          },
          "success" : {
            "backgroundColor" : "f2fcf3ff",
            "borderColor" : "037f0cff",
            "indicatorColor" : "037f0cff"
          },
          "warning" : {
            "backgroundColor" : "fffce9ff",
            "borderColor" : "8d6605ff",
            "indicatorColor" : "8d6605ff"
          }
        }
      }
    },
    "components" : {
      "alert" : {
        "borderRadius" : 12.0,
        "darkMode" : {
          "error" : {
            "backgroundColor" : "1a0000ff",
            "borderColor" : "eb6f6fff"
          }
        },
        "lightMode" : {
          "error" : {
            "backgroundColor" : "fff7f7ff",
            "borderColor" : "d91515ff"
          }
        }
      },
      "favicon" : {
        "enabledTypes" : [
          "ICO",
          "SVG"
        ]
      },
      "form" : {
        "backgroundImage" : {
          "enabled" : false
        },
        "borderRadius" : 8.0,
        "darkMode" : {
          "backgroundColor" : "0f1b2aff",
          "borderColor" : "424650ff"
        },
        "lightMode" : {
          "backgroundColor" : "ffffffff",
          "borderColor" : "c6c6cdff"
        },
        "logo" : {
          "enabled" : false,
          "formInclusion" : "IN",
          "location" : "CENTER",
          "position" : "TOP"
        }
      },
      "idpButton" : {
        "custom" : {
        },
        "standard" : {
          "darkMode" : {
            "active" : {
              "backgroundColor" : "354150ff",
              "borderColor" : "89bdeeff",
              "textColor" : "89bdeeff"
            },
            "defaults" : {
              "backgroundColor" : "0f1b2aff",
              "borderColor" : "c6c6cdff",
              "textColor" : "c6c6cdff"
            },
            "hover" : {
              "backgroundColor" : "192534ff",
              "borderColor" : "89bdeeff",
              "textColor" : "89bdeeff"
            }
          },
          "lightMode" : {
            "active" : {
              "backgroundColor" : "d3e7f9ff",
              "borderColor" : "033160ff",
              "textColor" : "033160ff"
            },
            "defaults" : {
              "backgroundColor" : "ffffffff",
              "borderColor" : "424650ff",
              "textColor" : "424650ff"
            },
            "hover" : {
              "backgroundColor" : "f2f8fdff",
              "borderColor" : "033160ff",
              "textColor" : "033160ff"
            }
          }
        }
      },
      "pageBackground" : {
        "darkMode" : {
          "color" : "0f1b2aff"
        },
        "image" : {
          "enabled" : true
        },
        "lightMode" : {
          "color" : "ffffffff"
        }
      },
      "pageFooter" : {
        "backgroundImage" : {
          "enabled" : false
        },
        "darkMode" : {
          "background" : {
            "color" : "0f141aff"
          },
          "borderColor" : "424650ff"
        },
        "lightMode" : {
          "background" : {
            "color" : "fafafaff"
          },
          "borderColor" : "d5dbdbff"
        },
        "logo" : {
          "enabled" : false,
          "location" : "START"
        }
      },
      "pageHeader" : {
        "backgroundImage" : {
          "enabled" : false
        },
        "darkMode" : {
          "background" : {
            "color" : "0f141aff"
          },
          "borderColor" : "424650ff"
        },
        "lightMode" : {
          "background" : {
            "color" : "fafafaff"
          },
          "borderColor" : "d5dbdbff"
        },
        "logo" : {
          "enabled" : false,
          "location" : "START"
        }
      },
      "pageText" : {
        "darkMode" : {
          "bodyColor" : "b6bec9ff",
          "descriptionColor" : "b6bec9ff",
          "headingColor" : "d1d5dbff"
        },
        "lightMode" : {
          "bodyColor" : "414d5cff",
          "descriptionColor" : "414d5cff",
          "headingColor" : "000716ff"
        }
      },
      "phoneNumberSelector" : {
        "displayType" : "TEXT"
      },
      "primaryButton" : {
        "darkMode" : {
          "active" : {
            "backgroundColor" : "539fe5ff",
            "textColor" : "000716ff"
          },
          "defaults" : {
            "backgroundColor" : "539fe5ff",
            "textColor" : "000716ff"
          },
          "disabled" : {
            "backgroundColor" : "ffffffff",
            "borderColor" : "ffffffff"
          },
          "hover" : {
            "backgroundColor" : "89bdeeff",
            "textColor" : "000716ff"
          }
        },
        "lightMode" : {
          "active" : {
            "backgroundColor" : "033160ff",
            "textColor" : "ffffffff"
          },
          "defaults" : {
            "backgroundColor" : "0972d3ff",
            "textColor" : "ffffffff"
          },
          "disabled" : {
            "backgroundColor" : "ffffffff",
            "borderColor" : "ffffffff"
          },
          "hover" : {
            "backgroundColor" : "033160ff",
            "textColor" : "ffffffff"
          }
        }
      },
      "secondaryButton" : {
        "darkMode" : {
          "active" : {
            "backgroundColor" : "354150ff",
            "borderColor" : "89bdeeff",
            "textColor" : "89bdeeff"
          },
          "defaults" : {
            "backgroundColor" : "0f1b2aff",
            "borderColor" : "539fe5ff",
            "textColor" : "539fe5ff"
          },
          "hover" : {
            "backgroundColor" : "192534ff",
            "borderColor" : "89bdeeff",
            "textColor" : "89bdeeff"
          }
        },
        "lightMode" : {
          "active" : {
            "backgroundColor" : "d3e7f9ff",
            "borderColor" : "033160ff",
            "textColor" : "033160ff"
          },
          "defaults" : {
            "backgroundColor" : "ffffffff",
            "borderColor" : "0972d3ff",
            "textColor" : "0972d3ff"
          },
          "hover" : {
            "backgroundColor" : "f2f8fdff",
            "borderColor" : "033160ff",
            "textColor" : "033160ff"
          }
        }
      }
    }
  })
}
`)
}

func testAccManagedLoginBrandingConfig_settingsUpdated(rName string) string {
	return acctest.ConfigCompose(testAccManagedLoginBrandingConfig_base(rName), `
resource "aws_cognito_managed_login_branding" "test" {
  client_id    = aws_cognito_user_pool_client.test.id
  user_pool_id = aws_cognito_user_pool.test.id

  settings = jsonencode({
    "categories" : {
      "auth" : {
        "authMethodOrder" : [
          [
            {
              "display" : "BUTTON",
              "type" : "FEDERATED"
            },
            {
              "display" : "INPUT",
              "type" : "USERNAME_PASSWORD"
            }
          ]
        ],
        "federation" : {
          "interfaceStyle" : "BUTTON_LIST",
          "order" : [
          ]
        }
      },
      "form" : {
        "displayGraphics" : true,
        "instructions" : {
          "enabled" : false
        },
        "languageSelector" : {
          "enabled" : false
        },
        "location" : {
          "horizontal" : "CENTER",
          "vertical" : "CENTER"
        },
        "sessionTimerDisplay" : "NONE"
      },
      "global" : {
        "colorSchemeMode" : "DARK",
        "pageFooter" : {
          "enabled" : false
        },
        "pageHeader" : {
          "enabled" : false
        },
        "spacingDensity" : "REGULAR"
      },
      "signUp" : {
        "acceptanceElements" : [
          {
            "enforcement" : "NONE",
            "textKey" : "en"
          }
        ]
      }
    },
    "componentClasses" : {
      "buttons" : {
        "borderRadius" : 8.0
      },
      "divider" : {
        "darkMode" : {
          "borderColor" : "232b37ff"
        },
        "lightMode" : {
          "borderColor" : "ebebf0ff"
        }
      },
      "dropDown" : {
        "borderRadius" : 8.0,
        "darkMode" : {
          "defaults" : {
            "itemBackgroundColor" : "192534ff"
          },
          "hover" : {
            "itemBackgroundColor" : "081120ff",
            "itemBorderColor" : "5f6b7aff",
            "itemTextColor" : "e9ebedff"
          },
          "match" : {
            "itemBackgroundColor" : "d1d5dbff",
            "itemTextColor" : "89bdeeff"
          }
        },
        "lightMode" : {
          "defaults" : {
            "itemBackgroundColor" : "ffffffff"
          },
          "hover" : {
            "itemBackgroundColor" : "f4f4f4ff",
            "itemBorderColor" : "7d8998ff",
            "itemTextColor" : "000716ff"
          },
          "match" : {
            "itemBackgroundColor" : "414d5cff",
            "itemTextColor" : "0972d3ff"
          }
        }
      },
      "focusState" : {
        "darkMode" : {
          "borderColor" : "539fe5ff"
        },
        "lightMode" : {
          "borderColor" : "0972d3ff"
        }
      },
      "idpButtons" : {
        "icons" : {
          "enabled" : true
        }
      },
      "input" : {
        "borderRadius" : 8.0,
        "darkMode" : {
          "defaults" : {
            "backgroundColor" : "0f1b2aff",
            "borderColor" : "5f6b7aff"
          },
          "placeholderColor" : "8d99a8ff"
        },
        "lightMode" : {
          "defaults" : {
            "backgroundColor" : "ffffffff",
            "borderColor" : "7d8998ff"
          },
          "placeholderColor" : "5f6b7aff"
        }
      },
      "inputDescription" : {
        "darkMode" : {
          "textColor" : "8d99a8ff"
        },
        "lightMode" : {
          "textColor" : "5f6b7aff"
        }
      },
      "inputLabel" : {
        "darkMode" : {
          "textColor" : "d1d5dbff"
        },
        "lightMode" : {
          "textColor" : "000716ff"
        }
      },
      "link" : {
        "darkMode" : {
          "defaults" : {
            "textColor" : "539fe5ff"
          },
          "hover" : {
            "textColor" : "89bdeeff"
          }
        },
        "lightMode" : {
          "defaults" : {
            "textColor" : "0972d3ff"
          },
          "hover" : {
            "textColor" : "033160ff"
          }
        }
      },
      "optionControls" : {
        "darkMode" : {
          "defaults" : {
            "backgroundColor" : "0f1b2aff",
            "borderColor" : "7d8998ff"
          },
          "selected" : {
            "backgroundColor" : "539fe5ff",
            "foregroundColor" : "000716ff"
          }
        },
        "lightMode" : {
          "defaults" : {
            "backgroundColor" : "ffffffff",
            "borderColor" : "7d8998ff"
          },
          "selected" : {
            "backgroundColor" : "0972d3ff",
            "foregroundColor" : "ffffffff"
          }
        }
      },
      "statusIndicator" : {
        "darkMode" : {
          "error" : {
            "backgroundColor" : "1a0000ff",
            "borderColor" : "eb6f6fff",
            "indicatorColor" : "eb6f6fff"
          },
          "pending" : {
            "indicatorColor" : "AAAAAAAA"
          },
          "success" : {
            "backgroundColor" : "001a02ff",
            "borderColor" : "29ad32ff",
            "indicatorColor" : "29ad32ff"
          },
          "warning" : {
            "backgroundColor" : "1d1906ff",
            "borderColor" : "e0ca57ff",
            "indicatorColor" : "e0ca57ff"
          }
        },
        "lightMode" : {
          "error" : {
            "backgroundColor" : "fff7f7ff",
            "borderColor" : "d91515ff",
            "indicatorColor" : "d91515ff"
          },
          "pending" : {
            "indicatorColor" : "AAAAAAAA"
          },
          "success" : {
            "backgroundColor" : "f2fcf3ff",
            "borderColor" : "037f0cff",
            "indicatorColor" : "037f0cff"
          },
          "warning" : {
            "backgroundColor" : "fffce9ff",
            "borderColor" : "8d6605ff",
            "indicatorColor" : "8d6605ff"
          }
        }
      }
    },
    "components" : {
      "alert" : {
        "borderRadius" : 12.0,
        "darkMode" : {
          "error" : {
            "backgroundColor" : "1a0000ff",
            "borderColor" : "eb6f6fff"
          }
        },
        "lightMode" : {
          "error" : {
            "backgroundColor" : "fff7f7ff",
            "borderColor" : "d91515ff"
          }
        }
      },
      "favicon" : {
        "enabledTypes" : [
          "ICO",
          "SVG"
        ]
      },
      "form" : {
        "backgroundImage" : {
          "enabled" : false
        },
        "borderRadius" : 8.0,
        "darkMode" : {
          "backgroundColor" : "0f1b2aff",
          "borderColor" : "424650ff"
        },
        "lightMode" : {
          "backgroundColor" : "ffffffff",
          "borderColor" : "c6c6cdff"
        },
        "logo" : {
          "enabled" : false,
          "formInclusion" : "IN",
          "location" : "CENTER",
          "position" : "TOP"
        }
      },
      "idpButton" : {
        "custom" : {
        },
        "standard" : {
          "darkMode" : {
            "active" : {
              "backgroundColor" : "354150ff",
              "borderColor" : "89bdeeff",
              "textColor" : "89bdeeff"
            },
            "defaults" : {
              "backgroundColor" : "0f1b2aff",
              "borderColor" : "c6c6cdff",
              "textColor" : "c6c6cdff"
            },
            "hover" : {
              "backgroundColor" : "192534ff",
              "borderColor" : "89bdeeff",
              "textColor" : "89bdeeff"
            }
          },
          "lightMode" : {
            "active" : {
              "backgroundColor" : "d3e7f9ff",
              "borderColor" : "033160ff",
              "textColor" : "033160ff"
            },
            "defaults" : {
              "backgroundColor" : "ffffffff",
              "borderColor" : "424650ff",
              "textColor" : "424650ff"
            },
            "hover" : {
              "backgroundColor" : "f2f8fdff",
              "borderColor" : "033160ff",
              "textColor" : "033160ff"
            }
          }
        }
      },
      "pageBackground" : {
        "darkMode" : {
          "color" : "0f1b2aff"
        },
        "image" : {
          "enabled" : true
        },
        "lightMode" : {
          "color" : "ffffffff"
        }
      },
      "pageFooter" : {
        "backgroundImage" : {
          "enabled" : false
        },
        "darkMode" : {
          "background" : {
            "color" : "0f141aff"
          },
          "borderColor" : "424650ff"
        },
        "lightMode" : {
          "background" : {
            "color" : "fafafaff"
          },
          "borderColor" : "d5dbdbff"
        },
        "logo" : {
          "enabled" : false,
          "location" : "START"
        }
      },
      "pageHeader" : {
        "backgroundImage" : {
          "enabled" : false
        },
        "darkMode" : {
          "background" : {
            "color" : "0f141aff"
          },
          "borderColor" : "424650ff"
        },
        "lightMode" : {
          "background" : {
            "color" : "fafafaff"
          },
          "borderColor" : "d5dbdbff"
        },
        "logo" : {
          "enabled" : false,
          "location" : "START"
        }
      },
      "pageText" : {
        "darkMode" : {
          "bodyColor" : "b6bec9ff",
          "descriptionColor" : "b6bec9ff",
          "headingColor" : "d1d5dbff"
        },
        "lightMode" : {
          "bodyColor" : "414d5cff",
          "descriptionColor" : "414d5cff",
          "headingColor" : "000716ff"
        }
      },
      "phoneNumberSelector" : {
        "displayType" : "TEXT"
      },
      "primaryButton" : {
        "darkMode" : {
          "active" : {
            "backgroundColor" : "539fe5ff",
            "textColor" : "000716ff"
          },
          "defaults" : {
            "backgroundColor" : "539fe5ff",
            "textColor" : "000716ff"
          },
          "disabled" : {
            "backgroundColor" : "ffffffff",
            "borderColor" : "ffffffff"
          },
          "hover" : {
            "backgroundColor" : "89bdeeff",
            "textColor" : "000716ff"
          }
        },
        "lightMode" : {
          "active" : {
            "backgroundColor" : "033160ff",
            "textColor" : "ffffffff"
          },
          "defaults" : {
            "backgroundColor" : "0972d3ff",
            "textColor" : "ffffffff"
          },
          "disabled" : {
            "backgroundColor" : "ffffffff",
            "borderColor" : "ffffffff"
          },
          "hover" : {
            "backgroundColor" : "033160ff",
            "textColor" : "ffffffff"
          }
        }
      },
      "secondaryButton" : {
        "darkMode" : {
          "active" : {
            "backgroundColor" : "354150ff",
            "borderColor" : "89bdeeff",
            "textColor" : "89bdeeff"
          },
          "defaults" : {
            "backgroundColor" : "0f1b2aff",
            "borderColor" : "539fe5ff",
            "textColor" : "539fe5ff"
          },
          "hover" : {
            "backgroundColor" : "192534ff",
            "borderColor" : "89bdeeff",
            "textColor" : "89bdeeff"
          }
        },
        "lightMode" : {
          "active" : {
            "backgroundColor" : "d3e7f9ff",
            "borderColor" : "033160ff",
            "textColor" : "033160ff"
          },
          "defaults" : {
            "backgroundColor" : "ffffffff",
            "borderColor" : "0972d3ff",
            "textColor" : "0972d3ff"
          },
          "hover" : {
            "backgroundColor" : "f2f8fdff",
            "borderColor" : "033160ff",
            "textColor" : "033160ff"
          }
        }
      }
    }
  })
}
`)
}

func testAccManagedLoginBrandingConfig_multiple(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool_client" "test0" {
  name                = "%[1]s-0"
  user_pool_id        = aws_cognito_user_pool.test.id
  explicit_auth_flows = ["ADMIN_NO_SRP_AUTH"]
}

resource "aws_cognito_user_pool_client" "test1" {
  name                = "%[1]s-1"
  user_pool_id        = aws_cognito_user_pool_client.test0.user_pool_id
  explicit_auth_flows = ["ADMIN_NO_SRP_AUTH"]
}

resource "aws_cognito_user_pool_client" "test2" {
  name                = "%[1]s-2"
  user_pool_id        = aws_cognito_user_pool_client.test1.user_pool_id
  explicit_auth_flows = ["ADMIN_NO_SRP_AUTH"]
}

resource "aws_cognito_managed_login_branding" "test1" {
  # Cross over user pool client IDs to test read logic.
  client_id    = aws_cognito_user_pool_client.test2.id
  user_pool_id = aws_cognito_user_pool.test.id

  use_cognito_provided_values = true
}

resource "aws_cognito_managed_login_branding" "test2" {
  client_id    = aws_cognito_user_pool_client.test1.id
  user_pool_id = aws_cognito_managed_login_branding.test1.user_pool_id

  use_cognito_provided_values = true
}
`, rName)
}
