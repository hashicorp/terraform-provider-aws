// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package globalaccelerator_test

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	globalaccelerator_test "github.com/hashicorp/terraform-provider-aws/internal/service/globalaccelerator"
)

func generateAccountID() string {
	source := rand.NewSource(42)
	rand := rand.New(source)

	accountID := ""
	for i := 0; i < 12; i++ {
		digit := rand.Intn(10)
		accountID += strconv.Itoa(digit)
	}
	return accountID
}

func TestExpandResources(t *testing.T) {
	t.Parallel()
	cases := []struct {
		Input          []globalaccelerator_test.ResourceData
		ExpectedOutput []*globalaccelerator.Resource
	}{
		{
			Input:          []globalaccelerator_test.ResourceData{},
			ExpectedOutput: nil,
		},
		{
			Input: []globalaccelerator_test.ResourceData{
				{
					EndpointID: types.StringValue("endpoint-1"),
					Region:     types.StringValue(acctest.Region()),
				},
				{
					EndpointID: types.StringValue("endpoint-2"),
					Region:     types.StringValue(""),
				},
			},
			ExpectedOutput: []*globalaccelerator.Resource{
				{
					EndpointId: aws.String("endpoint-1"),
					Region:     aws.String(acctest.Region()),
				},
				{
					EndpointId: aws.String("endpoint-2"),
				},
			},
		},
	}

	for _, tc := range cases {
		output := globalaccelerator_test.ExpandResources(tc.Input)
		if !reflect.DeepEqual(output, tc.ExpectedOutput) {
			t.Fatalf("bad: expected %v, got %v", tc.ExpectedOutput, output)
		}
	}
}

func TestFlattenResources(t *testing.T) {
	t.Parallel()
	elem := globalaccelerator_test.ResourceDataElementType
	partition := acctest.Partition()
	region := acctest.Region()
	endpointID := fmt.Sprintf("arn:%s:ec2:%s:171405876253:elastic-ip/eipalloc-1234567890abcdef0", partition, region)

	endpoint1, _ := types.ObjectValue(elem.AttrTypes, map[string]attr.Value{
		"endpoint_id": types.StringValue(endpointID),
		"region":      types.StringValue(region),
	})

	expectedList, _ := types.ListValue(elem, []attr.Value{endpoint1})

	testCases := []struct {
		Name     string
		Input    []*globalaccelerator.Resource
		Expected types.List
	}{
		{
			Name:     "empty input",
			Input:    []*globalaccelerator.Resource{},
			Expected: types.ListNull(elem),
		},
		{
			Name: "non-empty input",
			Input: []*globalaccelerator.Resource{
				{
					EndpointId: aws.String(endpointID),
					Region:     aws.String(region),
				},
			},
			Expected: expectedList,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			output, err := globalaccelerator_test.FlattenResources(ctx, tc.Input)

			if err != nil {
				t.Fatalf("flattenResources() error = %v, wantErr %v", err, nil)
			}

			if !reflect.DeepEqual(output, tc.Expected) {
				t.Errorf("flattenResources() got = %v, want %v", output, tc.Expected)
			}
		})
	}
}

func TestDiffResources(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	elem := globalaccelerator_test.ResourceDataElementType

	partition := acctest.Partition()
	region := acctest.Region()
	alternateRegion := acctest.AlternateRegion()
	endpointID := fmt.Sprintf("arn:%s:ec2:%s:171405876253:elastic-ip/eipalloc-1234567890abcdef0", partition, region)
	endpointID2 := fmt.Sprintf("arn:%s:ec2:%s:171405876253:elastic-ip/eipalloc-1234567890abcdef1", partition, alternateRegion)

	endpoint1Object, _ := types.ObjectValue(elem.AttrTypes, map[string]attr.Value{
		"endpoint_id": types.StringValue(endpointID),
		"region":      types.StringValue(region),
	})
	endpoint2Object, _ := types.ObjectValue(elem.AttrTypes, map[string]attr.Value{
		"endpoint_id": types.StringValue(endpointID2),
		"region":      types.StringValue(alternateRegion),
	})

	expectedResource1 := &globalaccelerator.Resource{
		EndpointId: aws.String(endpointID),
		Region:     aws.String(region),
	}
	expectedResource2 := &globalaccelerator.Resource{
		EndpointId: aws.String(endpointID2),
		Region:     aws.String(alternateRegion),
	}

	cases := []struct {
		Name             string
		OldList          types.List
		NewList          types.List
		ExpectedToAdd    []*globalaccelerator.Resource
		ExpectedToRemove []*globalaccelerator.Resource
	}{
		{
			Name:             "EmptyLists",
			OldList:          types.ListNull(elem),
			NewList:          types.ListNull(elem),
			ExpectedToAdd:    []*globalaccelerator.Resource{},
			ExpectedToRemove: []*globalaccelerator.Resource{},
		},
		{
			Name:    "Resource to add",
			OldList: types.ListValueMust(elem, []attr.Value{endpoint1Object}),
			NewList: types.ListValueMust(elem, []attr.Value{endpoint1Object, endpoint2Object}),
			ExpectedToAdd: []*globalaccelerator.Resource{
				expectedResource2,
			},
			ExpectedToRemove: []*globalaccelerator.Resource{},
		},
		{
			Name:          "Resource to remove",
			OldList:       types.ListValueMust(elem, []attr.Value{endpoint1Object, endpoint2Object}),
			NewList:       types.ListValueMust(elem, []attr.Value{endpoint1Object}),
			ExpectedToAdd: []*globalaccelerator.Resource{},
			ExpectedToRemove: []*globalaccelerator.Resource{
				expectedResource2,
			},
		},
		{
			Name:    "Resource to add and remove",
			OldList: types.ListValueMust(elem, []attr.Value{endpoint1Object}),
			NewList: types.ListValueMust(elem, []attr.Value{endpoint2Object}),
			ExpectedToAdd: []*globalaccelerator.Resource{
				expectedResource2,
			},
			ExpectedToRemove: []*globalaccelerator.Resource{
				expectedResource1,
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			toAdd, toRemove, _ := globalaccelerator_test.DiffResources(ctx, tc.OldList, tc.NewList)

			if !reflect.DeepEqual(toAdd, tc.ExpectedToAdd) {
				t.Errorf("expected to add: %#v, got: %#v", tc.ExpectedToAdd, toAdd)
			}

			if !reflect.DeepEqual(toRemove, tc.ExpectedToRemove) {
				t.Errorf("expected to remove: %#v, got: %#v", tc.ExpectedToRemove, toRemove)
			}
		})
	}
}

func TestDiffPrincipals(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	elemType := types.StringType

	principal1 := types.StringValue("principal-1")
	principal2 := types.StringValue("principal-2")

	oldList, _ := types.ListValue(elemType, []attr.Value{principal1})
	newList, _ := types.ListValue(elemType, []attr.Value{principal2})

	cases := []struct {
		Name             string
		OldList          types.List
		NewList          types.List
		ExpectedToAdd    []*string
		ExpectedToRemove []*string
	}{
		{
			Name:             "EmptyLists",
			OldList:          types.ListNull(elemType),
			NewList:          types.ListNull(elemType),
			ExpectedToAdd:    []*string{},
			ExpectedToRemove: []*string{},
		},
		{
			Name:             "NonEmptyLists",
			OldList:          oldList,
			NewList:          newList,
			ExpectedToAdd:    []*string{aws.String("principal-2")},
			ExpectedToRemove: []*string{aws.String("principal-1")},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			toAdd, toRemove, _ := globalaccelerator_test.DiffPrincipals(ctx, tc.OldList, tc.NewList)

			if !reflect.DeepEqual(toAdd, tc.ExpectedToAdd) {
				t.Errorf("expected to add: %#v, got: %#v", tc.ExpectedToAdd, toAdd)
			}

			if !reflect.DeepEqual(toRemove, tc.ExpectedToRemove) {
				t.Errorf("expected to remove: %#v, got: %#v", tc.ExpectedToRemove, toRemove)
			}
		})
	}
}

func TestAccGlobalAcceleratorCrossAccountAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_cross_account_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var v *globalaccelerator.DescribeCrossAccountAttachmentOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, "aws_globalaccelerator_cross_account_attachment"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrossAccountAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrossAccountAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrossAccountAttachmentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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
func TestAccGlobalAcceleratorCrossAccountAttachment_principals(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_cross_account_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var v *globalaccelerator.DescribeCrossAccountAttachmentOutput
	accountId := generateAccountID()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, "aws_globalaccelerator_cross_account_attachment"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrossAccountAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrossAccountAttachmentConfig_principals(rName, accountId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrossAccountAttachmentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckTypeSetElemAttr(resourceName, "principals.*", accountId),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_time"),
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

func TestAccGlobalAcceleratorCrossAccountAttachment_resources(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_cross_account_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	partition := acctest.Partition()
	region := acctest.Region()
	alternateRegion := acctest.AlternateRegion()
	endpointID := fmt.Sprintf("arn:%s:ec2:%s:171405876253:elastic-ip/eipalloc-1234567890abcdef0", partition, region)
	endpointID2 := fmt.Sprintf("arn:%s:ec2:%s:171405876253:elastic-ip/eipalloc-1234567890abcdef1", partition, alternateRegion)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, "aws_globalaccelerator_cross_account_attachment"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrossAccountAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrossAccountAttachmentConfig_resources(rName, []globalaccelerator_test.ResourceData{
					{EndpointID: types.StringValue(endpointID), Region: types.StringValue(region)},
					{EndpointID: types.StringValue(endpointID2), Region: types.StringValue(alternateRegion)},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "resources.*", map[string]string{
						"endpoint_id": endpointID,
						"region":      region,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "resources.*", map[string]string{
						"endpoint_id": endpointID2,
						"region":      alternateRegion,
					}),
				),
			},
		},
	})
}

func TestAccGlobalAcceleratorCrossAccountAttachment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_cross_account_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var v *globalaccelerator.DescribeCrossAccountAttachmentOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, "aws_globalaccelerator_cross_account_attachment"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrossAccountAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrossAccountAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrossAccountAttachmentExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, globalaccelerator_test.ResourceCrossAccountAttachment, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCrossAccountAttachmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GlobalAcceleratorConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_globalaccelerator_cross_account_attachment" {
				continue
			}

			_, err := conn.DescribeCrossAccountAttachmentWithContext(ctx, &globalaccelerator.DescribeCrossAccountAttachmentInput{
				AttachmentArn: aws.String(rs.Primary.ID),
			})
			if err != nil && strings.Contains(err.Error(), "AttachmentNotFoundException") {
				return nil
			} else if err != nil {
				return fmt.Errorf("error checking if Global Accelerator Cross Account Attachment %s still exists: %s", rs.Primary.ID, err)
			}

			return fmt.Errorf("Global Accelerator Cross Account Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCrossAccountAttachmentExists(ctx context.Context, resourceName string, v **globalaccelerator.DescribeCrossAccountAttachmentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlobalAcceleratorConn(ctx)

		output, err := conn.DescribeCrossAccountAttachmentWithContext(ctx, &globalaccelerator.DescribeCrossAccountAttachmentInput{
			AttachmentArn: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if output == nil || output.CrossAccountAttachment == nil {
			return fmt.Errorf("Global Accelerator Cross Account Attachment %s does not exist", rs.Primary.ID)
		}

		*v = output

		return nil
	}
}

func testAccCrossAccountAttachmentConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_cross_account_attachment" "test" {
  name = %[1]q
}
`, rName)
}

func testAccCrossAccountAttachmentConfig_principals(rName string, accountId string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_cross_account_attachment" "test" {
  name       = %[1]q
  principals = [%[2]q]
}
`, rName, accountId)
}

func testAccCrossAccountAttachmentConfig_resources(rName string, resources []globalaccelerator_test.ResourceData) string {
	var resourcesStr []string
	for _, r := range resources {
		resourcesStr = append(resourcesStr, fmt.Sprintf(`{ endpoint_id = "%s", region = "%s" }`, r.EndpointID.ValueString(), r.Region.ValueString()))
	}
	return fmt.Sprintf(`
resource "aws_globalaccelerator_cross_account_attachment" "test" {
  name      = "%s"
  resources = [%s]
}
`, rName, strings.Join(resourcesStr, ", "))
}
