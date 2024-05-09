// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_macie2_custom_data_identifier", name="Custom Data Identifier")
// @Tags
func ResourceCustomDataIdentifier() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCustomDataIdentifierCreate,
		ReadWithoutTimeout:   resourceCustomDataIdentifierRead,
		DeleteWithoutTimeout: resourceCustomDataIdentifierDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"keywords": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 50,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(3, 90),
				},
			},
			"ignore_words": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 10,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(4, 90),
				},
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc:  validation.StringLenBetween(0, 128),
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validation.StringLenBetween(0, 128-id.UniqueIDSuffixLength),
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"maximum_match_distance": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 300),
			},
			names.AttrTags:    tftags.TagsSchemaForceNew(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCustomDataIdentifierCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Conn(ctx)

	input := &macie2.CreateCustomDataIdentifierInput{
		ClientToken: aws.String(id.UniqueId()),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk("regex"); ok {
		input.Regex = aws.String(v.(string))
	}
	if v, ok := d.GetOk("keywords"); ok {
		input.Keywords = flex.ExpandStringSet(v.(*schema.Set))
	}
	if v, ok := d.GetOk("ignore_words"); ok {
		input.IgnoreWords = flex.ExpandStringSet(v.(*schema.Set))
	}
	input.Name = aws.String(create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string)))
	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("maximum_match_distance"); ok {
		input.MaximumMatchDistance = aws.Int64(int64(v.(int)))
	}

	var err error
	var output *macie2.CreateCustomDataIdentifierOutput
	err = retry.RetryContext(ctx, 4*time.Minute, func() *retry.RetryError {
		output, err = conn.CreateCustomDataIdentifierWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, macie2.ErrorCodeClientError) {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateCustomDataIdentifierWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Macie CustomDataIdentifier: %s", err)
	}

	d.SetId(aws.StringValue(output.CustomDataIdentifierId))

	return append(diags, resourceCustomDataIdentifierRead(ctx, d, meta)...)
}

func resourceCustomDataIdentifierRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Conn(ctx)

	input := &macie2.GetCustomDataIdentifierInput{
		Id: aws.String(d.Id()),
	}

	resp, err := conn.GetCustomDataIdentifierWithContext(ctx, input)

	if !d.IsNewResource() && (tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
		tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled")) {
		log.Printf("[WARN] Macie CustomDataIdentifier (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Macie CustomDataIdentifier (%s): %s", d.Id(), err)
	}

	d.Set("regex", resp.Regex)
	if err = d.Set("keywords", flex.FlattenStringList(resp.Keywords)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting `%s` for Macie CustomDataIdentifier (%s): %s", "keywords", d.Id(), err)
	}
	if err = d.Set("ignore_words", flex.FlattenStringList(resp.IgnoreWords)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting `%s` for Macie CustomDataIdentifier (%s): %s", "ignore_words", d.Id(), err)
	}
	d.Set(names.AttrName, resp.Name)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.StringValue(resp.Name)))
	d.Set(names.AttrDescription, resp.Description)
	d.Set("maximum_match_distance", resp.MaximumMatchDistance)

	setTagsOut(ctx, resp.Tags)

	if aws.BoolValue(resp.Deleted) {
		log.Printf("[WARN] Macie CustomDataIdentifier (%s) is soft deleted, removing from state", d.Id())
		d.SetId("")
	}

	d.Set(names.AttrCreatedAt, aws.TimeValue(resp.CreatedAt).Format(time.RFC3339))
	d.Set(names.AttrARN, resp.Arn)

	return diags
}

func resourceCustomDataIdentifierDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Conn(ctx)

	input := &macie2.DeleteCustomDataIdentifierInput{
		Id: aws.String(d.Id()),
	}

	_, err := conn.DeleteCustomDataIdentifierWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Macie CustomDataIdentifier (%s): %s", d.Id(), err)
	}
	return diags
}
