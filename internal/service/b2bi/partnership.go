// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package b2bi

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/b2bi"
	awstypes "github.com/aws/aws-sdk-go-v2/service/b2bi/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_b2bi_partnership", name="Partnership")
// @Tags(identifierAttribute="partnership_arn")
func resourcePartnership() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePartnershipCreate,
		ReadWithoutTimeout:   resourcePartnershipRead,
		UpdateWithoutTimeout: resourcePartnershipUpdate,
		DeleteWithoutTimeout: resourcePartnershipDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"capabilities": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 64),
				},
			},
			"email": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(5, 254),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 254),
			},
			"partnership_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"partnership_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"phone": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(7, 22),
			},
			"profile_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"trading_partner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourcePartnershipCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).B2BIClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &b2bi.CreatePartnershipInput{
		Capabilities: flex.ExpandStringValueList(d.Get("capabilities").([]any)),
		Email:        aws.String(d.Get("email").(string)),
		Name:         aws.String(name),
		ProfileId:    aws.String(d.Get("profile_id").(string)),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("phone"); ok {
		input.Phone = aws.String(v.(string))
	}

	output, err := conn.CreatePartnership(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating B2BI Partnership (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.PartnershipId))

	return append(diags, resourcePartnershipRead(ctx, d, meta)...)
}

func resourcePartnershipRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).B2BIClient(ctx)

	output, err := findPartnershipByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] B2BI Partnership (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading B2BI Partnership (%s): %s", d.Id(), err)
	}

	d.Set("capabilities", output.Capabilities)
	d.Set("email", output.Email)
	d.Set(names.AttrName, output.Name)
	d.Set("partnership_arn", output.PartnershipArn)
	d.Set("partnership_id", output.PartnershipId)
	d.Set("phone", output.Phone)
	d.Set("profile_id", output.ProfileId)
	d.Set("trading_partner_id", output.TradingPartnerId)

	return diags
}

func resourcePartnershipUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).B2BIClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &b2bi.UpdatePartnershipInput{
			PartnershipId: aws.String(d.Id()),
		}

		if d.HasChange("capabilities") {
			input.Capabilities = flex.ExpandStringValueList(d.Get("capabilities").([]any))
		}

		if d.HasChange(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		_, err := conn.UpdatePartnership(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating B2BI Partnership (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourcePartnershipRead(ctx, d, meta)...)
}

func resourcePartnershipDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).B2BIClient(ctx)

	log.Printf("[DEBUG] Deleting B2BI Partnership: %s", d.Id())
	_, err := conn.DeletePartnership(ctx, &b2bi.DeletePartnershipInput{
		PartnershipId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting B2BI Partnership (%s): %s", d.Id(), err)
	}

	return diags
}

func findPartnershipByID(ctx context.Context, conn *b2bi.Client, id string) (*b2bi.GetPartnershipOutput, error) {
	input := &b2bi.GetPartnershipInput{
		PartnershipId: aws.String(id),
	}

	output, err := conn.GetPartnership(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{LastError: err}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}
