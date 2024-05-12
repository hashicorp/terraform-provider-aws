// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppFabricAppAuthorization_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appfabric_app_authorization.test"
	// var appauthorization types.AppAuthorization

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// acctest.PreCheckPartitionHasService(t, names.AppFabric)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		// CheckDestroy:             testAccCheckAppAuthorizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppAuthorizationConfig_basic(),
				Check:  resource.ComposeTestCheckFunc(
				// testAccCheckAppAuthorizationExists(ctx, resourceName, &appauthorization),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"credential"},
			},
		},
	})
}

// func TestAccAppFabricAppAuthorization_disappears(t *testing.T) {
// 	ctx := acctest.Context(t)
// 	var appauthorization types.AppAuthorization
// 	resourceName := "aws_appfabric_app_authorization.test"

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck: func() {
// 			acctest.PreCheck(ctx, t)
// 			acctest.PreCheckPartitionHasService(t, names.AppFabric)
// 			testAccPreCheck(ctx, t)
// 		},
// 		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
// 		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
// 		CheckDestroy:             testAccCheckAppAuthorizationDestroy(ctx),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccAppAuthorizationConfig_basic(),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckAppAuthorizationExists(ctx, resourceName, &appauthorization),
// 					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfappfabric.ResourceAppAuthorization, resourceName),
// 				),
// 				ExpectNonEmptyPlan: true,
// 			},
// 		},
// 	})
// }

// func testAccCheckAppAuthorizationDestroy(ctx context.Context) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)

// 		for _, rs := range s.RootModule().Resources {
// 			if rs.Type != "aws_appfabric_app_authorization" {
// 				continue
// 			}

// 			input := &appfabric.DescribeAppAuthorizationInput{
// 				AppAuthorizationId: aws.String(rs.Primary.ID),
// 			}
// 			_, err := conn.DescribeAppAuthorization(ctx, &appfabric.DescribeAppAuthorizationInput{
// 				AppAuthorizationId: aws.String(rs.Primary.ID),
// 			})
// 			if errs.IsA[*types.ResourceNotFoundException](err) {
// 				return nil
// 			}
// 			if err != nil {
// 				return create.Error(names.AppFabric, create.ErrActionCheckingDestroyed, tfappfabric.ResNameAppAuthorization, rs.Primary.ID, err)
// 			}

// 			return create.Error(names.AppFabric, create.ErrActionCheckingDestroyed, tfappfabric.ResNameAppAuthorization, rs.Primary.ID, errors.New("not destroyed"))
// 		}

// 		return nil
// 	}
// }

// func testAccCheckAppAuthorizationExists(ctx context.Context, name string, appauthorization *appfabric.DescribeAppAuthorizationResponse) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		rs, ok := s.RootModule().Resources[name]
// 		if !ok {
// 			return create.Error(names.AppFabric, create.ErrActionCheckingExistence, tfappfabric.ResNameAppAuthorization, name, errors.New("not found"))
// 		}

// 		if rs.Primary.ID == "" {
// 			return create.Error(names.AppFabric, create.ErrActionCheckingExistence, tfappfabric.ResNameAppAuthorization, name, errors.New("not set"))
// 		}

// 		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)
// 		resp, err := conn.DescribeAppAuthorization(ctx, &appfabric.DescribeAppAuthorizationInput{
// 			AppAuthorizationId: aws.String(rs.Primary.ID),
// 		})

// 		if err != nil {
// 			return create.Error(names.AppFabric, create.ErrActionCheckingExistence, tfappfabric.ResNameAppAuthorization, rs.Primary.ID, err)
// 		}

// 		*appauthorization = *resp

// 		return nil
// 	}
// }

func testAccAppAuthorizationConfig_basic() string {
	return `

resource "aws_appfabric_app_authorization" "test" {
  app_bundle_identifier   = "arn:aws:appfabric:eu-west-1:926562225508:appbundle/7546871d-363f-4374-a90e-464d8b69e3e7"
  app             		  = "TERRAFORMCLOUD"
  auth_type 			  = "apiKey"
  credential {
	api_key_credential {
		api_key = "testkey"
	}
  }
  tenant {
	tenant_display_name = "mkandylis-org-aws"
	tenant_identifier   = "mkandylis-org-aws"
  }      

}
`
}
