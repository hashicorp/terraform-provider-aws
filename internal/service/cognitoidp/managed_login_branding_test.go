// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfcognitoidp "github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidp"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCognitoIDPManagedLoginBranding_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var managedLoginBranding awstypes.ManagedLoginBrandingType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_managed_login_branding.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			//acctest.PreCheckPartitionHasService(t, names.CognitoIDPEndpointID)
			//testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedLoginBrandingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccManagedLoginBrandingConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedLoginBrandingExists(ctx, resourceName, &managedLoginBranding),
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
func TestAccCognitoIDPManagedLoginBranding_asset(t *testing.T) {
	ctx := acctest.Context(t)

	var managedLoginBranding awstypes.ManagedLoginBrandingType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_managed_login_branding.test"
	assetBytes, _ := os.ReadFile("test-fixtures/login_branding_asset.svg")
	assetBase64Encoded := itypes.Base64Encode(assetBytes)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			//acctest.PreCheckPartitionHasService(t, names.CognitoIDPEndpointID)
			//testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedLoginBrandingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccManagedLoginBrandingConfig_asset(rName, string(awstypes.AssetCategoryTypePageFooterBackground), string(awstypes.ColorSchemeModeTypeDark)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedLoginBrandingExists(ctx, resourceName, &managedLoginBranding),
					resource.TestCheckResourceAttr(resourceName, "asset.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "asset.0.bytes", assetBase64Encoded),
					resource.TestCheckResourceAttr(resourceName, "asset.0.category", string(awstypes.AssetCategoryTypePageFooterBackground)),
					resource.TestCheckResourceAttr(resourceName, "asset.0.color_mode", string(awstypes.ColorSchemeModeTypeDark)),
					resource.TestCheckResourceAttr(resourceName, "asset.0.extension", "SVG"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccManagedLoginBrandingConfig_asset(rName, string(awstypes.AssetCategoryTypePageHeaderBackground), string(awstypes.ColorSchemeModeTypeLight)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedLoginBrandingExists(ctx, resourceName, &managedLoginBranding),
					resource.TestCheckResourceAttr(resourceName, "asset.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "asset.0.bytes", assetBase64Encoded),
					resource.TestCheckResourceAttr(resourceName, "asset.0.category", string(awstypes.AssetCategoryTypePageHeaderBackground)),
					resource.TestCheckResourceAttr(resourceName, "asset.0.color_mode", string(awstypes.ColorSchemeModeTypeLight)),
					resource.TestCheckResourceAttr(resourceName, "asset.0.extension", "SVG"),
				),
			},
			{
				Config: testAccManagedLoginBrandingConfig_assetMultiple(rName,
					string(awstypes.AssetCategoryTypePageHeaderBackground),
					string(awstypes.ColorSchemeModeTypeLight),
					string(awstypes.AssetCategoryTypePageFooterBackground),
					string(awstypes.ColorSchemeModeTypeLight),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedLoginBrandingExists(ctx, resourceName, &managedLoginBranding),
					resource.TestCheckResourceAttr(resourceName, "asset.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "asset.*", map[string]string{
						"bytes":      assetBase64Encoded,
						"category":   string(awstypes.AssetCategoryTypePageHeaderBackground),
						"color_mode": string(awstypes.ColorSchemeModeTypeLight),
						"extension":  "SVG",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "asset.*", map[string]string{
						"bytes":      assetBase64Encoded,
						"category":   string(awstypes.AssetCategoryTypePageFooterBackground),
						"color_mode": string(awstypes.ColorSchemeModeTypeLight),
						"extension":  "SVG",
					}),
				),
			},
			{
				Config: testAccManagedLoginBrandingConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedLoginBrandingExists(ctx, resourceName, &managedLoginBranding),
					resource.TestCheckResourceAttr(resourceName, "asset.#", "0"),
				),
			},
		},
	})
}

func TestAccCognitoIDPManagedLoginBranding_settings(t *testing.T) {
	ctx := acctest.Context(t)

	var managedLoginBrandingBefore, managedLoginBrandingAfter awstypes.ManagedLoginBrandingType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_managed_login_branding.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			//acctest.PreCheckPartitionHasService(t, names.CognitoIDPEndpointID)
			//testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedLoginBrandingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccManagedLoginBrandingConfig_settings(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedLoginBrandingExists(ctx, resourceName, &managedLoginBrandingBefore),
					acctest.CheckResourceAttrIsJSONString(resourceName, "settings"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"settings"},
			},
			{
				Config: testAccManagedLoginBrandingConfig_settings(rName, false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedLoginBrandingExists(ctx, resourceName, &managedLoginBrandingAfter),
					acctest.CheckResourceAttrIsJSONString(resourceName, "settings"),
					testAccCheckManagedLoginBrandingNotRecreated(&managedLoginBrandingBefore, &managedLoginBrandingAfter),
				),
			},
			{
				Config: testAccManagedLoginBrandingConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedLoginBrandingExists(ctx, resourceName, &managedLoginBrandingBefore),
					resource.TestCheckNoResourceAttr(resourceName, "settings"),
				),
			},
		},
	})
}

func TestAccCognitoIDPManagedLoginBranding_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var managedLoginBranding awstypes.ManagedLoginBrandingType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_managed_login_branding.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			//testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedLoginBrandingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccManagedLoginBrandingConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedLoginBrandingExists(ctx, resourceName, &managedLoginBranding),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfcognitoidp.ResourceManagedLoginBranding, resourceName),
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

func testAccCheckManagedLoginBrandingDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cognito_managed_login_branding" {
				continue
			}

			_, err := tfcognitoidp.FindManagedLoginBrandingByID(ctx, conn, rs.Primary.Attributes["managed_login_branding_id"], rs.Primary.Attributes[names.AttrUserPoolID])
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.CognitoIDP, create.ErrActionCheckingDestroyed, tfcognitoidp.ResNameManagedLoginBranding, rs.Primary.ID, err)
			}

			return create.Error(names.CognitoIDP, create.ErrActionCheckingDestroyed, tfcognitoidp.ResNameManagedLoginBranding, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckManagedLoginBrandingExists(ctx context.Context, name string, managedLoginBranding *awstypes.ManagedLoginBrandingType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.CognitoIDP, create.ErrActionCheckingExistence, tfcognitoidp.ResNameManagedLoginBranding, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.CognitoIDP, create.ErrActionCheckingExistence, tfcognitoidp.ResNameManagedLoginBranding, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPClient(ctx)

		resp, err := tfcognitoidp.FindManagedLoginBrandingByID(ctx, conn, rs.Primary.Attributes["managed_login_branding_id"], rs.Primary.Attributes[names.AttrUserPoolID])
		if err != nil {
			return create.Error(names.CognitoIDP, create.ErrActionCheckingExistence, tfcognitoidp.ResNameManagedLoginBranding, rs.Primary.ID, err)
		}

		*managedLoginBranding = *resp

		return nil
	}
}

//func testAccPreCheck(ctx context.Context, t *testing.T) {
//	conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPClient(ctx)
//
//	input := &cognitoidentityprovider.ListManagedLoginBrandingsInput{}
//
//	_, err := conn.ListManagedLoginBrandings(ctx, input)
//
//	if acctest.PreCheckSkipError(err) {
//		t.Skipf("skipping acceptance testing: %s", err)
//	}
//	if err != nil {
//		t.Fatalf("unexpected PreCheck error: %s", err)
//	}
//}

func testAccCheckManagedLoginBrandingNotRecreated(before, after *awstypes.ManagedLoginBrandingType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.ManagedLoginBrandingId), aws.ToString(after.ManagedLoginBrandingId); before != after {
			return create.Error(names.CognitoIDP, create.ErrActionCheckingNotRecreated, tfcognitoidp.ResNameManagedLoginBranding, before, errors.New("recreated"))
		}

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
	return acctest.ConfigCompose(
		testAccManagedLoginBrandingConfig_base(rName),
		`
resource "aws_cognito_managed_login_branding" "test" {
  client_id                   = aws_cognito_user_pool_client.test.id
  user_pool_id                = aws_cognito_user_pool.test.id
  use_cognito_provided_values = true
}`)
}

// the asset file used in this test comes from an example in the AWS documentation
// https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_CreateManagedLoginBranding.html#API_CreateManagedLoginBranding_Examples
func testAccManagedLoginBrandingConfig_asset(rName, category, colorMode string) string {
	return acctest.ConfigCompose(
		testAccManagedLoginBrandingConfig_base(rName),
		fmt.Sprintf(`
resource "aws_cognito_managed_login_branding" "test" {
  client_id                   = aws_cognito_user_pool_client.test.id
  user_pool_id                = aws_cognito_user_pool.test.id
  use_cognito_provided_values = true

  asset {
    bytes      = filebase64("test-fixtures/login_branding_asset.svg")
    category   = %[1]q
    color_mode = %[2]q
    extension  = "SVG"
  }
}
`, category, colorMode))
}
func testAccManagedLoginBrandingConfig_assetMultiple(rName, category1, colorMode1, category2, colorMode2 string) string {
	return acctest.ConfigCompose(
		testAccManagedLoginBrandingConfig_base(rName),
		fmt.Sprintf(`
resource "aws_cognito_managed_login_branding" "test" {
  client_id                   = aws_cognito_user_pool_client.test.id
  user_pool_id                = aws_cognito_user_pool.test.id
  use_cognito_provided_values = true

  asset {
    bytes      = filebase64("test-fixtures/login_branding_asset.svg")
    category   = %[1]q
    color_mode = %[2]q
    extension  = "SVG"
  }
  asset {
    bytes      = filebase64("test-fixtures/login_branding_asset.svg")
    category   = %[3]q
    color_mode = %[4]q
    extension  = "SVG"
  }
}
`, category1, colorMode1, category2, colorMode2))
}

// using `settings` described in the AWS documentation
// https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_CreateManagedLoginBranding.html#API_CreateManagedLoginBranding_Examples
func testAccManagedLoginBrandingConfig_settings(rName string, formBackgroundImageEnabled bool) string {
	return acctest.ConfigCompose(
		testAccManagedLoginBrandingConfig_base(rName),
		fmt.Sprintf(`
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
          "enabled" : %[1]t
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
`, formBackgroundImageEnabled))
}
