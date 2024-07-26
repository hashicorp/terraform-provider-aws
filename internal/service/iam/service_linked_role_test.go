// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestDecodeServiceLinkedRoleID(t *testing.T) {
	t.Parallel()

	var testCases = []struct {
		Input        string
		ServiceName  string
		RoleName     string
		CustomSuffix string
		ErrCount     int
	}{
		{
			Input:    "not-arn",
			ErrCount: 1,
		},
		{
			Input:    "arn:aws:iam::123456789012:role/not-service-linked-role", //lintignore:AWSAT005
			ErrCount: 1,
		},
		{
			Input:        "arn:aws:iam::123456789012:role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling", //lintignore:AWSAT005
			ServiceName:  "autoscaling.amazonaws.com",
			RoleName:     "AWSServiceRoleForAutoScaling",
			CustomSuffix: "",
			ErrCount:     0,
		},
		{
			Input:        "arn:aws:iam::123456789012:role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling_custom-suffix", //lintignore:AWSAT005
			ServiceName:  "autoscaling.amazonaws.com",
			RoleName:     "AWSServiceRoleForAutoScaling_custom-suffix",
			CustomSuffix: "custom-suffix",
			ErrCount:     0,
		},
		{
			Input:        "arn:aws:iam::123456789012:role/aws-service-role/dynamodb.application-autoscaling.amazonaws.com/AWSServiceRoleForApplicationAutoScaling_DynamoDBTable", //lintignore:AWSAT005
			ServiceName:  "dynamodb.application-autoscaling.amazonaws.com",
			RoleName:     "AWSServiceRoleForApplicationAutoScaling_DynamoDBTable",
			CustomSuffix: "DynamoDBTable",
			ErrCount:     0,
		},
	}

	for _, tc := range testCases {
		serviceName, roleName, customSuffix, err := tfiam.DecodeServiceLinkedRoleID(tc.Input)
		if tc.ErrCount == 0 && err != nil {
			t.Fatalf("expected %q not to trigger an error, received: %s", tc.Input, err)
		}
		if tc.ErrCount > 0 && err == nil {
			t.Fatalf("expected %q to trigger an error", tc.Input)
		}
		if serviceName != tc.ServiceName {
			t.Fatalf("expected service name %q to be %q", serviceName, tc.ServiceName)
		}
		if roleName != tc.RoleName {
			t.Fatalf("expected role name %q to be %q", roleName, tc.RoleName)
		}
		if customSuffix != tc.CustomSuffix {
			t.Fatalf("expected custom suffix %q to be %q", customSuffix, tc.CustomSuffix)
		}
	}
}

func TestAccIAMServiceLinkedRole_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_iam_service_linked_role.test"
	awsServiceName := "inspector.amazonaws.com"
	name := "AWSServiceRoleForAmazonInspector"
	path := fmt.Sprintf("/aws-service-role/%s/", awsServiceName)
	arnResource := fmt.Sprintf("role%s%s", path, name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceLinkedRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					// Remove existing if possible
					client := acctest.Provider.Meta().(*conns.AWSClient)
					arn := arn.ARN{
						Partition: client.Partition,
						Service:   "iam",
						Region:    client.Region,
						AccountID: client.AccountID,
						Resource:  arnResource,
					}.String()
					r := tfiam.ResourceServiceLinkedRole()
					d := r.Data(nil)
					d.SetId(arn)
					err := acctest.DeleteResource(ctx, r, d, client)

					if err != nil {
						t.Fatalf("deleting service-linked role %s: %s", name, err)
					}
				},
				Config: testAccServiceLinkedRoleConfig_basic(awsServiceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceLinkedRoleExists(ctx, resourceName),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrARN, "iam", arnResource),
					resource.TestCheckResourceAttr(resourceName, "aws_service_name", awsServiceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "create_date"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, name),
					resource.TestCheckResourceAttr(resourceName, names.AttrPath, path),
					resource.TestCheckResourceAttrSet(resourceName, "unique_id"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccIAMServiceLinkedRole_customSuffix(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_iam_service_linked_role.test"
	awsServiceName := "autoscaling.amazonaws.com"
	customSuffix := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	name := fmt.Sprintf("AWSServiceRoleForAutoScaling_%s", customSuffix)
	path := fmt.Sprintf("/aws-service-role/%s/", awsServiceName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceLinkedRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceLinkedRoleConfig_customSuffix(awsServiceName, customSuffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceLinkedRoleExists(ctx, resourceName),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrARN, "iam", fmt.Sprintf("role%s%s", path, name)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, name),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/4439
func TestAccIAMServiceLinkedRole_CustomSuffix_diffSuppressFunc(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_iam_service_linked_role.test"
	awsServiceName := "custom-resource.application-autoscaling.amazonaws.com"
	name := "AWSServiceRoleForApplicationAutoScaling_CustomResource"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceLinkedRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceLinkedRoleConfig_basic(awsServiceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceLinkedRoleExists(ctx, resourceName),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrARN, "iam", fmt.Sprintf("role/aws-service-role/%s/%s", awsServiceName, name)),
					resource.TestCheckResourceAttr(resourceName, "custom_suffix", "CustomResource"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, name),
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

func TestAccIAMServiceLinkedRole_description(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_iam_service_linked_role.test"
	awsServiceName := "autoscaling.amazonaws.com"
	customSuffix := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceLinkedRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceLinkedRoleConfig_description(awsServiceName, customSuffix, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceLinkedRoleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				Config: testAccServiceLinkedRoleConfig_description(awsServiceName, customSuffix, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceLinkedRoleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
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

func TestAccIAMServiceLinkedRole_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_iam_service_linked_role.test"
	awsServiceName := "autoscaling.amazonaws.com"
	customSuffix := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceLinkedRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceLinkedRoleConfig_customSuffix(awsServiceName, customSuffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceLinkedRoleExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiam.ResourceServiceLinkedRole(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiam.ResourceServiceLinkedRole(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckServiceLinkedRoleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_service_linked_role" {
				continue
			}

			_, roleName, _, err := tfiam.DecodeServiceLinkedRoleID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfiam.FindRoleByName(ctx, conn, roleName)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IAM Service Linked Role %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckServiceLinkedRoleExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		_, roleName, _, err := tfiam.DecodeServiceLinkedRoleID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfiam.FindRoleByName(ctx, conn, roleName)

		return err
	}
}

func testAccServiceLinkedRoleConfig_basic(awsServiceName string) string {
	return fmt.Sprintf(`
resource "aws_iam_service_linked_role" "test" {
  aws_service_name = %[1]q
}
`, awsServiceName)
}

func testAccServiceLinkedRoleConfig_customSuffix(awsServiceName, customSuffix string) string {
	return fmt.Sprintf(`
resource "aws_iam_service_linked_role" "test" {
  aws_service_name = %[1]q
  custom_suffix    = %[2]q
}
`, awsServiceName, customSuffix)
}

func testAccServiceLinkedRoleConfig_description(awsServiceName, customSuffix, description string) string {
	return fmt.Sprintf(`
resource "aws_iam_service_linked_role" "test" {
  aws_service_name = %[1]q
  custom_suffix    = %[2]q
  description      = %[3]q
}
`, awsServiceName, customSuffix, description)
}
