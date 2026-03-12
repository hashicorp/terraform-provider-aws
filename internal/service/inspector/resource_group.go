// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package inspector

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/inspector"
	awstypes "github.com/aws/aws-sdk-go-v2/service/inspector/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_inspector_resource_group", name="Resource Group")
// @ArnIdentity
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/inspector/types;types.ResourceGroup")
// @Testing(preIdentityVersion="v6.4.0")
// @Testing(checkDestroyNoop=true)
// @Testing(preCheck="testAccPreCheck")
func resourceResourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourceGroupCreate,
		ReadWithoutTimeout:   resourceResourceGroupRead,
		DeleteWithoutTimeout: schema.NoopContext,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: {
				Type:     schema.TypeMap,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceResourceGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InspectorClient(ctx)

	input := inspector.CreateResourceGroupInput{
		ResourceGroupTags: expandResourceGroupTags(d.Get(names.AttrTags).(map[string]any)),
	}
	output, err := conn.CreateResourceGroup(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Inspector Classic Resource Group: %s", err)
	}

	d.SetId(aws.ToString(output.ResourceGroupArn))

	return append(diags, resourceResourceGroupRead(ctx, d, meta)...)
}

func resourceResourceGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InspectorClient(ctx)

	resourceGroup, err := findResourceGroupByARN(ctx, conn, d.Id())

	if retry.NotFound(err) {
		log.Printf("[WARN] Inspector Classic Resource Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Inspector Classic Resource Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, resourceGroup.Arn)
	if err := d.Set(names.AttrTags, flattenResourceGroupTags(resourceGroup.Tags)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}

func findResourceGroups(ctx context.Context, conn *inspector.Client, input *inspector.DescribeResourceGroupsInput) ([]awstypes.ResourceGroup, error) {
	output, err := conn.DescribeResourceGroups(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	if err := failedItemsError(output.FailedItems); err != nil {
		return nil, err
	}

	return output.ResourceGroups, nil
}

func findResourceGroup(ctx context.Context, conn *inspector.Client, input *inspector.DescribeResourceGroupsInput) (*awstypes.ResourceGroup, error) {
	output, err := findResourceGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findResourceGroupByARN(ctx context.Context, conn *inspector.Client, arn string) (*awstypes.ResourceGroup, error) {
	input := inspector.DescribeResourceGroupsInput{
		ResourceGroupArns: []string{arn},
	}

	output, err := findResourceGroup(ctx, conn, &input)

	if tfawserr.ErrMessageContains(err, string(awstypes.FailedItemErrorCodeItemDoesNotExist), arn) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func expandResourceGroupTags(tfMap map[string]any) []awstypes.ResourceGroupTag {
	var apiObjects []awstypes.ResourceGroupTag

	for k, v := range tfMap {
		apiObjects = append(apiObjects, awstypes.ResourceGroupTag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		})
	}

	return apiObjects
}

func flattenResourceGroupTags(apiObjects []awstypes.ResourceGroupTag) map[string]any {
	tfMap := map[string]any{}

	for _, apiObject := range apiObjects {
		tfMap[aws.ToString(apiObject.Key)] = aws.ToString(apiObject.Value)
	}

	return tfMap
}
