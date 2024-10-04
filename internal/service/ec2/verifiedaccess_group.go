// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_verifiedaccess_group", name="Verified Access Group")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceVerifiedAccessGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVerifiedAccessGroupCreate,
		ReadWithoutTimeout:   resourceVerifiedAccessGroupRead,
		UpdateWithoutTimeout: resourceVerifiedAccessGroupUpdate,
		DeleteWithoutTimeout: resourceVerifiedAccessGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrCreationTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deletion_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			names.AttrLastUpdatedTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrOwner: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy_document": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"sse_configuration": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"customer_managed_key_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						names.AttrKMSKeyARN: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"verifiedaccess_group_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"verifiedaccess_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"verifiedaccess_instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVerifiedAccessGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateVerifiedAccessGroupInput{
		ClientToken:              aws.String(id.UniqueId()),
		TagSpecifications:        getTagSpecificationsIn(ctx, types.ResourceTypeVerifiedAccessGroup),
		VerifiedAccessInstanceId: aws.String(d.Get("verifiedaccess_instance_id").(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("policy_document"); ok {
		input.PolicyDocument = aws.String(v.(string))
	}

	if v, ok := d.GetOk("sse_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.SseSpecification = expandVerifiedAccessSseSpecificationRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	output, err := conn.CreateVerifiedAccessGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Verified Access Group: %s", err)
	}

	d.SetId(aws.ToString(output.VerifiedAccessGroup.VerifiedAccessGroupId))

	return append(diags, resourceVerifiedAccessGroupRead(ctx, d, meta)...)
}

func resourceVerifiedAccessGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	group, err := findVerifiedAccessGroupByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Verified Access Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Verified Access Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrCreationTime, group.CreationTime)
	d.Set("deletion_time", group.DeletionTime)
	d.Set(names.AttrDescription, group.Description)
	d.Set(names.AttrLastUpdatedTime, group.LastUpdatedTime)
	d.Set(names.AttrOwner, group.Owner)
	if v := group.SseSpecification; v != nil {
		if err := d.Set("sse_configuration", flattenVerifiedAccessSseSpecificationResponse(v)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting sse_configuration: %s", err)
		}
	} else {
		d.Set("sse_configuration", nil)
	}
	d.Set("verifiedaccess_group_arn", group.VerifiedAccessGroupArn)
	d.Set("verifiedaccess_group_id", group.VerifiedAccessGroupId)
	d.Set("verifiedaccess_instance_id", group.VerifiedAccessInstanceId)

	setTagsOut(ctx, group.Tags)

	output, err := findVerifiedAccessGroupPolicyByID(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Verified Access Group (%s) policy: %s", d.Id(), err)
	}

	d.Set("policy_document", output.PolicyDocument)

	return diags
}

func resourceVerifiedAccessGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChangesExcept("policy_document", names.AttrTags, names.AttrTagsAll, "sse_configuration") {
		input := &ec2.ModifyVerifiedAccessGroupInput{
			ClientToken:           aws.String(id.UniqueId()),
			VerifiedAccessGroupId: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange("verified_access_instance_id") {
			input.VerifiedAccessInstanceId = aws.String(d.Get(names.AttrDescription).(string))
		}

		_, err := conn.ModifyVerifiedAccessGroup(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Verified Access Group (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("policy_document") {
		in := &ec2.ModifyVerifiedAccessGroupPolicyInput{
			PolicyDocument:        aws.String(d.Get("policy_document").(string)),
			VerifiedAccessGroupId: aws.String(d.Id()),
			PolicyEnabled:         aws.Bool(true),
		}

		_, err := conn.ModifyVerifiedAccessGroupPolicy(ctx, in)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Verified Access Group (%s) policy: %s", d.Id(), err)
		}
	}

	if d.HasChange("sse_configuration") {
		in := &ec2.ModifyVerifiedAccessGroupPolicyInput{
			VerifiedAccessGroupId: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("sse_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			in.SseSpecification = expandVerifiedAccessSseSpecificationRequest(v.([]interface{})[0].(map[string]interface{}))
		}

		_, err := conn.ModifyVerifiedAccessGroupPolicy(ctx, in)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SSE on Verified Access Group (%s) policy: %s", d.Id(), err)
		}
	}

	return append(diags, resourceVerifiedAccessGroupRead(ctx, d, meta)...)
}

func resourceVerifiedAccessGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[INFO] Deleting Verified Access Group: %s", d.Id())
	_, err := conn.DeleteVerifiedAccessGroup(ctx, &ec2.DeleteVerifiedAccessGroupInput{
		ClientToken:           aws.String(id.UniqueId()),
		VerifiedAccessGroupId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVerifiedAccessGroupIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Verified Access Group (%s): %s", d.Id(), err)
	}

	return diags
}

func expandVerifiedAccessSseSpecificationRequest(tfMap map[string]interface{}) *types.VerifiedAccessSseSpecificationRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.VerifiedAccessSseSpecificationRequest{}

	if v, ok := tfMap[names.AttrKMSKeyARN].(string); ok && v != "" {
		apiObject.KmsKeyArn = aws.String(v)
	}

	if v, ok := tfMap["customer_managed_key_enabled"].(bool); ok {
		apiObject.CustomerManagedKeyEnabled = aws.Bool(v)
	}

	return apiObject
}

func flattenVerifiedAccessSseSpecificationResponse(apiObject *types.VerifiedAccessSseSpecificationResponse) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CustomerManagedKeyEnabled; v != nil {
		tfMap["customer_managed_key_enabled"] = aws.ToBool(v)
	}

	if v := apiObject.KmsKeyArn; v != nil {
		tfMap[names.AttrKMSKeyARN] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}
