// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/shield"
	awstypes "github.com/aws/aws-sdk-go-v2/service/shield/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_shield_protection_group", name="Protection Group")
// @Tags(identifierAttribute="protection_group_arn")
func ResourceProtectionGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProtectionGroupCreate,
		ReadWithoutTimeout:   resourceProtectionGroupRead,
		UpdateWithoutTimeout: resourceProtectionGroupUpdate,
		DeleteWithoutTimeout: resourceProtectionGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"aggregation": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ProtectionGroupAggregation](),
			},
			"members": {
				Type:          schema.TypeList,
				Optional:      true,
				MinItems:      0,
				MaxItems:      10000,
				ConflictsWith: []string{names.AttrResourceType},
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(verify.ValidARN,
						validation.StringLenBetween(1, 2048),
					),
				},
			},
			"pattern": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ProtectionGroupPattern](),
			},
			"protection_group_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 36),
				ForceNew:     true,
			},
			"protection_group_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrResourceType: {
				Type:             schema.TypeString,
				Optional:         true,
				ConflictsWith:    []string{"members"},
				ValidateDiagFunc: enum.Validate[awstypes.ProtectedResourceType](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceProtectionGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ShieldClient(ctx)

	protectionGroupID := d.Get("protection_group_id").(string)
	input := &shield.CreateProtectionGroupInput{
		Aggregation:       awstypes.ProtectionGroupAggregation(d.Get("aggregation").(string)),
		Pattern:           awstypes.ProtectionGroupPattern(d.Get("pattern").(string)),
		ProtectionGroupId: aws.String(protectionGroupID),
		Tags:              getTagsIn(ctx),
	}

	if v, ok := d.GetOk("members"); ok {
		input.Members = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrResourceType); ok {
		input.ResourceType = awstypes.ProtectedResourceType(v.(string))
	}

	_, err := conn.CreateProtectionGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Shield Protection Group (%s): %s", protectionGroupID, err)
	}

	d.SetId(protectionGroupID)

	return append(diags, resourceProtectionGroupRead(ctx, d, meta)...)
}

func resourceProtectionGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ShieldClient(ctx)

	resp, err := findProtectionGroupByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Shield Protection Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Shield Protection Group (%s): %s", d.Id(), err)
	}

	arn := aws.ToString(resp.ProtectionGroupArn)
	d.Set("protection_group_arn", arn)
	d.Set("aggregation", resp.Aggregation)
	d.Set("protection_group_id", resp.ProtectionGroupId)
	d.Set("pattern", resp.Pattern)
	d.Set("members", resp.Members)
	d.Set(names.AttrResourceType, resp.ResourceType)

	return diags
}

func resourceProtectionGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ShieldClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &shield.UpdateProtectionGroupInput{
			Aggregation:       awstypes.ProtectionGroupAggregation(d.Get("aggregation").(string)),
			Pattern:           awstypes.ProtectionGroupPattern(d.Get("pattern").(string)),
			ProtectionGroupId: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("members"); ok {
			input.Members = flex.ExpandStringValueList(v.([]interface{}))
		}

		if v, ok := d.GetOk(names.AttrResourceType); ok {
			input.ResourceType = awstypes.ProtectedResourceType(v.(string))
		}

		_, err := conn.UpdateProtectionGroup(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Shield Protection Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceProtectionGroupRead(ctx, d, meta)...)
}

func resourceProtectionGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ShieldClient(ctx)

	log.Printf("[DEBUG] Deletinh Shield Protection Group: %s", d.Id())
	_, err := conn.DeleteProtectionGroup(ctx, &shield.DeleteProtectionGroupInput{
		ProtectionGroupId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Shield Protection Group (%s): %s", d.Id(), err)
	}

	return diags
}

func findProtectionGroupByID(ctx context.Context, conn *shield.Client, id string) (*awstypes.ProtectionGroup, error) {
	input := &shield.DescribeProtectionGroupInput{
		ProtectionGroupId: aws.String(id),
	}

	resp, err := conn.DescribeProtectionGroup(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if resp.ProtectionGroup == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return resp.ProtectionGroup, nil
}
