// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
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

// @SDKResource("aws_verifiedaccess_instance", name="Verified Access Instance")
// @Tags(identifierAttribute="id")
func ResourceVerifiedAccessInstance() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVerifiedAccessInstanceCreate,
		ReadWithoutTimeout:   resourceVerifiedAccessInstanceRead,
		UpdateWithoutTimeout: resourceVerifiedAccessInstanceUpdate,
		DeleteWithoutTimeout: resourceVerifiedAccessInstanceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"last_updated_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"verified_access_trust_providers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device_trust_provider_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"trust_provider_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"user_trust_provider_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"verified_access_trust_provider_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVerifiedAccessInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateVerifiedAccessInstanceInput{
		ClientToken:       aws.String(id.UniqueId()),
		TagSpecifications: getTagSpecificationsInV2(ctx, types.ResourceTypeVerifiedAccessInstance),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateVerifiedAccessInstance(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Verified Access Instance: %s", err)
	}

	d.SetId(aws.ToString(output.VerifiedAccessInstance.VerifiedAccessInstanceId))

	return append(diags, resourceVerifiedAccessInstanceRead(ctx, d, meta)...)
}

func resourceVerifiedAccessInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	output, err := FindVerifiedAccessInstanceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Verified Access Instance (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Verified Access Instance (%s): %s", d.Id(), err)
	}

	d.Set("description", output.Description)
	d.Set("creation_time", output.CreationTime)
	d.Set("last_updated_time", output.LastUpdatedTime)

	if v := output.VerifiedAccessTrustProviders; v != nil {
		if err := d.Set("verified_access_trust_providers", flattenVerifiedAccessTrustProviders(v)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting verified access trust providers: %s", err)
		}
	} else {
		d.Set("verified_access_trust_providers", nil)
	}

	setTagsOutV2(ctx, output.Tags)

	return diags
}

func resourceVerifiedAccessInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &ec2.ModifyVerifiedAccessInstanceInput{
			ClientToken:              aws.String(id.UniqueId()),
			VerifiedAccessInstanceId: aws.String(d.Id()),
		}

		if d.HasChanges("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		_, err := conn.ModifyVerifiedAccessInstance(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Verified Access Instance (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceVerifiedAccessInstanceRead(ctx, d, meta)...)
}

func resourceVerifiedAccessInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[INFO] Deleting Verified Access Instance: %s", d.Id())
	_, err := conn.DeleteVerifiedAccessInstance(ctx, &ec2.DeleteVerifiedAccessInstanceInput{
		ClientToken:              aws.String(id.UniqueId()),
		VerifiedAccessInstanceId: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Verified Access Instance (%s): %s", d.Id(), err)
	}

	return diags
}

func flattenVerifiedAccessTrustProviders(apiObjects []types.VerifiedAccessTrustProviderCondensed) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		v := flattenVerifiedAccessTrustProvider(apiObject)

		if len(v) > 0 {
			tfList = append(tfList, v)
		}
	}

	return tfList
}

func flattenVerifiedAccessTrustProvider(apiObject types.VerifiedAccessTrustProviderCondensed) map[string]interface{} {
	tfMap := map[string]interface{}{
		"device_trust_provider_type": string(apiObject.DeviceTrustProviderType),
		"trust_provider_type":        string(apiObject.TrustProviderType),
		"user_trust_provider_type":   string(apiObject.UserTrustProviderType),
	}

	if v := apiObject.Description; v != nil {
		tfMap["description"] = aws.ToString(v)
	}

	if v := apiObject.VerifiedAccessTrustProviderId; v != nil {
		tfMap["verified_access_trust_provider_id"] = aws.ToString(v)
	}

	return tfMap
}
