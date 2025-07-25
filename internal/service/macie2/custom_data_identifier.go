// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/macie2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/macie2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_macie2_custom_data_identifier", name="Custom Data Identifier")
// @Tags(identifierAttribute="arn")
func resourceCustomDataIdentifier() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCustomDataIdentifierCreate,
		ReadWithoutTimeout:   resourceCustomDataIdentifierRead,
		UpdateWithoutTimeout: resourceCustomDataIdentifierUpdate,
		DeleteWithoutTimeout: resourceCustomDataIdentifierDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
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
			"maximum_match_distance": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 300),
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
			"regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(4 * time.Minute),
		},
	}
}

func resourceCustomDataIdentifierCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := macie2.CreateCustomDataIdentifierInput{
		ClientToken: aws.String(id.UniqueId()),
		Name:        aws.String(name),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ignore_words"); ok {
		input.IgnoreWords = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("keywords"); ok {
		input.Keywords = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("maximum_match_distance"); ok {
		input.MaximumMatchDistance = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("regex"); ok {
		input.Regex = aws.String(v.(string))
	}

	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutCreate), func() (any, error) {
		return conn.CreateCustomDataIdentifier(ctx, &input)
	}, errCodeClientError)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Macie Custom Data Identifier (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*macie2.CreateCustomDataIdentifierOutput).CustomDataIdentifierId))

	return append(diags, resourceCustomDataIdentifierRead(ctx, d, meta)...)
}

func resourceCustomDataIdentifierRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	output, err := findCustomDataIdentifierByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Macie Custom Data Identifier (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Macie Custom Data Identifier (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.Arn)
	d.Set(names.AttrCreatedAt, aws.ToTime(output.CreatedAt).Format(time.RFC3339))
	d.Set(names.AttrDescription, output.Description)
	d.Set("ignore_words", output.IgnoreWords)
	d.Set("keywords", output.Keywords)
	d.Set("maximum_match_distance", output.MaximumMatchDistance)
	d.Set(names.AttrName, output.Name)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(output.Name)))
	d.Set("regex", output.Regex)

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceCustomDataIdentifierUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Tags only.
	return resourceCustomDataIdentifierRead(ctx, d, meta)
}

func resourceCustomDataIdentifierDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	_, err := conn.DeleteCustomDataIdentifier(ctx, &macie2.DeleteCustomDataIdentifierInput{
		Id: aws.String(d.Id()),
	})

	if isCustomDataIdentifierNotFoundError(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Macie Custom Data Identifier (%s): %s", d.Id(), err)
	}

	return diags
}

func findCustomDataIdentifierByID(ctx context.Context, conn *macie2.Client, id string) (*macie2.GetCustomDataIdentifierOutput, error) {
	input := macie2.GetCustomDataIdentifierInput{
		Id: aws.String(id),
	}

	output, err := findCustomDataIdentifier(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if aws.ToBool(output.Deleted) {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}

func findCustomDataIdentifier(ctx context.Context, conn *macie2.Client, input *macie2.GetCustomDataIdentifierInput) (*macie2.GetCustomDataIdentifierOutput, error) {
	output, err := conn.GetCustomDataIdentifier(ctx, input)

	if isCustomDataIdentifierNotFoundError(err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func isCustomDataIdentifierNotFoundError(err error) bool {
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return true
	}
	if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "Macie is not enabled") {
		return true
	}

	return false
}
