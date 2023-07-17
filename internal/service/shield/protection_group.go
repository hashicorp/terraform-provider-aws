// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/shield"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
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
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(shield.ProtectionGroupAggregation_Values(), false),
			},
			"members": {
				Type:          schema.TypeList,
				Optional:      true,
				MinItems:      0,
				MaxItems:      10000,
				ConflictsWith: []string{"resource_type"},
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(verify.ValidARN,
						validation.StringLenBetween(1, 2048),
					),
				},
			},
			"pattern": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(shield.ProtectionGroupPattern_Values(), false),
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
			"resource_type": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"members"},
				ValidateFunc:  validation.StringInSlice(shield.ProtectedResourceType_Values(), false),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceProtectionGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ShieldConn(ctx)

	protectionGroupID := d.Get("protection_group_id").(string)
	input := &shield.CreateProtectionGroupInput{
		Aggregation:       aws.String(d.Get("aggregation").(string)),
		Pattern:           aws.String(d.Get("pattern").(string)),
		ProtectionGroupId: aws.String(protectionGroupID),
		Tags:              getTagsIn(ctx),
	}

	if v, ok := d.GetOk("members"); ok {
		input.Members = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("resource_type"); ok {
		input.ResourceType = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Shield Protection Group: %s", input)
	_, err := conn.CreateProtectionGroupWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Shield Protection Group (%s): %s", protectionGroupID, err)
	}

	d.SetId(protectionGroupID)

	return append(diags, resourceProtectionGroupRead(ctx, d, meta)...)
}

func resourceProtectionGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ShieldConn(ctx)

	input := &shield.DescribeProtectionGroupInput{
		ProtectionGroupId: aws.String(d.Id()),
	}

	resp, err := conn.DescribeProtectionGroupWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, shield.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Shield Protection Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Shield Protection Group (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(resp.ProtectionGroup.ProtectionGroupArn)
	d.Set("protection_group_arn", arn)
	d.Set("aggregation", resp.ProtectionGroup.Aggregation)
	d.Set("protection_group_id", resp.ProtectionGroup.ProtectionGroupId)
	d.Set("pattern", resp.ProtectionGroup.Pattern)
	d.Set("members", resp.ProtectionGroup.Members)
	d.Set("resource_type", resp.ProtectionGroup.ResourceType)

	return diags
}

func resourceProtectionGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ShieldConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &shield.UpdateProtectionGroupInput{
			Aggregation:       aws.String(d.Get("aggregation").(string)),
			Pattern:           aws.String(d.Get("pattern").(string)),
			ProtectionGroupId: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("members"); ok {
			input.Members = flex.ExpandStringList(v.([]interface{}))
		}

		if v, ok := d.GetOk("resource_type"); ok {
			input.ResourceType = aws.String(v.(string))
		}

		log.Printf("[DEBUG] Updating Shield Protection Group: %s", input)
		_, err := conn.UpdateProtectionGroupWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Shield Protection Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceProtectionGroupRead(ctx, d, meta)...)
}

func resourceProtectionGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ShieldConn(ctx)

	log.Printf("[DEBUG] Deletinh Shield Protection Group: %s", d.Id())
	_, err := conn.DeleteProtectionGroupWithContext(ctx, &shield.DeleteProtectionGroupInput{
		ProtectionGroupId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, shield.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Shield Protection Group (%s): %s", d.Id(), err)
	}

	return diags
}
