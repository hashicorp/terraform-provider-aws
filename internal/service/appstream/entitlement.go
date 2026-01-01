// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package appstream

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appstream"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	entitlementOperationTimeout = 4 * time.Minute
)

// @SDKResource("aws_appstream_entitlement", name="Entitlement")
func resourceEntitlement() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEntitlementCreate,
		ReadWithoutTimeout:   resourceEntitlementRead,
		UpdateWithoutTimeout: resourceEntitlementUpdate,
		DeleteWithoutTimeout: resourceEntitlementDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				stackName, name, err := entitlementParseResourceID(d.Id())
				if err != nil {
					return nil, err
				}

				d.Set("stack_name", stackName)
				d.SetId(name)

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"app_visibility": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AppVisibility](),
			},
			"attributes": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			names.AttrCreatedTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"stack_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceEntitlementCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	name := d.Get(names.AttrName).(string)
	stackName := d.Get("stack_name").(string)
	input := appstream.CreateEntitlementInput{
		Name:          aws.String(name),
		StackName:     aws.String(stackName),
		AppVisibility: awstypes.AppVisibility(d.Get("app_visibility").(string)),
		Attributes:    expandEntitlementAttributes(d.Get("attributes").(*schema.Set).List()),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	outputRaw, err := tfresource.RetryWhenIsA[any, *awstypes.ResourceNotFoundException](ctx, entitlementOperationTimeout, func(ctx context.Context) (any, error) {
		return conn.CreateEntitlement(ctx, &input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppStream Entitlement (%s/%s): %s", stackName, name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*appstream.CreateEntitlementOutput).Entitlement.Name))

	return append(diags, resourceEntitlementRead(ctx, d, meta)...)
}

func resourceEntitlementRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	stackName := d.Get("stack_name").(string)
	entitlement, err := findEntitlementByTwoPartKey(ctx, conn, stackName, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] AppStream Entitlement (%s/%s) not found, removing from state", stackName, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppStream Entitlement (%s/%s): %s", stackName, d.Id(), err)
	}

	d.Set("app_visibility", entitlement.AppVisibility)
	if err = d.Set("attributes", flattenEntitlementAttributes(entitlement.Attributes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting attributes: %s", err)
	}
	d.Set(names.AttrCreatedTime, aws.ToTime(entitlement.CreatedTime).Format(time.RFC3339))
	d.Set(names.AttrDescription, entitlement.Description)
	d.Set("last_modified_time", aws.ToTime(entitlement.LastModifiedTime).Format(time.RFC3339))
	d.Set(names.AttrName, entitlement.Name)
	d.Set("stack_name", entitlement.StackName)

	return diags
}

func resourceEntitlementUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	stackName := d.Get("stack_name").(string)
	input := appstream.UpdateEntitlementInput{
		Name:      aws.String(d.Id()),
		StackName: aws.String(stackName),
	}

	if d.HasChange("app_visibility") {
		input.AppVisibility = awstypes.AppVisibility(d.Get("app_visibility").(string))
	}

	if d.HasChange("attributes") {
		input.Attributes = expandEntitlementAttributes(d.Get("attributes").(*schema.Set).List())
	}

	if d.HasChange(names.AttrDescription) {
		input.Description = aws.String(d.Get(names.AttrDescription).(string))
	}

	_, err := conn.UpdateEntitlement(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating AppStream Entitlement (%s/%s): %s", stackName, d.Id(), err)
	}

	return append(diags, resourceEntitlementRead(ctx, d, meta)...)
}

func resourceEntitlementDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	stackName := d.Get("stack_name").(string)
	log.Printf("[DEBUG] Deleting AppStream Entitlement: %s/%s", stackName, d.Id())
	input := appstream.DeleteEntitlementInput{
		Name:      aws.String(d.Id()),
		StackName: aws.String(stackName),
	}
	_, err := conn.DeleteEntitlement(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppStream Entitlement (%s/%s): %s", stackName, d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, entitlementOperationTimeout, func(ctx context.Context) (any, error) {
		return findEntitlementByTwoPartKey(ctx, conn, stackName, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for AppStream Entitlement (%s/%s) delete: %s", stackName, d.Id(), err)
	}

	return diags
}

const entitlementResourceIDSeparator = "/"

func entitlementParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, entitlementResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected StackName%[2]sName", id, entitlementResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findEntitlementByTwoPartKey(ctx context.Context, conn *appstream.Client, stackName, name string) (*awstypes.Entitlement, error) {
	input := appstream.DescribeEntitlementsInput{
		StackName: aws.String(stackName),
		Name:      aws.String(name),
	}

	return findEntitlement(ctx, conn, &input)
}

func findEntitlement(ctx context.Context, conn *appstream.Client, input *appstream.DescribeEntitlementsInput) (*awstypes.Entitlement, error) {
	output, err := findEntitlements(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findEntitlements(ctx context.Context, conn *appstream.Client, input *appstream.DescribeEntitlementsInput) ([]awstypes.Entitlement, error) {
	var output []awstypes.Entitlement

	for {
		page, err := conn.DescribeEntitlements(ctx, input)
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &sdkretry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Entitlements...)

		if aws.ToString(page.NextToken) == "" {
			break
		}

		input.NextToken = page.NextToken
	}

	return output, nil
}

func expandEntitlementAttribute(tfMap map[string]any) awstypes.EntitlementAttribute {
	if tfMap == nil {
		return awstypes.EntitlementAttribute{}
	}

	apiObject := awstypes.EntitlementAttribute{
		Name:  aws.String(tfMap[names.AttrName].(string)),
		Value: aws.String(tfMap[names.AttrValue].(string)),
	}

	return apiObject
}

func expandEntitlementAttributes(tfList []any) []awstypes.EntitlementAttribute {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.EntitlementAttribute

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandEntitlementAttribute(tfMap))
	}

	return apiObjects
}

func flattenEntitlementAttribute(apiObject awstypes.EntitlementAttribute) map[string]any {
	tfMap := map[string]any{}

	tfMap[names.AttrName] = aws.ToString(apiObject.Name)
	tfMap[names.AttrValue] = aws.ToString(apiObject.Value)

	return tfMap
}

func flattenEntitlementAttributes(apiObjects []awstypes.EntitlementAttribute) []map[string]any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenEntitlementAttribute(apiObject))
	}

	return tfList
}
