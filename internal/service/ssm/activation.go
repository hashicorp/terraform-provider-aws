// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ssm_activation", name="Activation")
// @Tags
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/ssm/types;awstypes;awstypes.Activation")
// @Testing(importIgnore="activation_code")
// @Testing(tagsUpdateForceNew=true)
func resourceActivation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceActivationCreate,
		ReadWithoutTimeout:   resourceActivationRead,
		DeleteWithoutTimeout: resourceActivationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"activation_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"expiration_date": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"expired": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"iam_role": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"registration_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"registration_limit": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchemaForceNew(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceActivationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &ssm.CreateActivationInput{
		DefaultInstanceName: aws.String(name),
		IamRole:             aws.String(d.Get("iam_role").(string)),
		Tags:                getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("expiration_date"); ok {
		t, _ := time.Parse(time.RFC3339, v.(string))
		input.ExpirationDate = aws.Time(t)
	}

	if v, ok := d.GetOk("registration_limit"); ok {
		input.RegistrationLimit = aws.Int32(int32(v.(int)))
	}

	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreateActivation(ctx, input)
	}, errCodeValidationException, "Nonexistent role")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SSM Activation (%s): %s", name, err)
	}

	output := outputRaw.(*ssm.CreateActivationOutput)

	d.SetId(aws.ToString(output.ActivationId))
	d.Set("activation_code", output.ActivationCode)

	return append(diags, resourceActivationRead(ctx, d, meta)...)
}

func resourceActivationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	activation, err := findActivationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSM Activation %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSM Activation (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrDescription, activation.Description)
	d.Set("expiration_date", aws.ToTime(activation.ExpirationDate).Format(time.RFC3339))
	d.Set("expired", activation.Expired)
	d.Set("iam_role", activation.IamRole)
	d.Set(names.AttrName, activation.DefaultInstanceName)
	d.Set("registration_count", activation.RegistrationsCount)
	d.Set("registration_limit", activation.RegistrationLimit)

	setTagsOut(ctx, activation.Tags)

	return diags
}

func resourceActivationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	log.Printf("[DEBUG] Deleting SSM Activation: %s", d.Id())
	_, err := conn.DeleteActivation(ctx, &ssm.DeleteActivationInput{
		ActivationId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.InvalidActivation](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SSM Activation (%s): %s", d.Id(), err)
	}

	return diags
}

func findActivationByID(ctx context.Context, conn *ssm.Client, id string) (*awstypes.Activation, error) {
	input := &ssm.DescribeActivationsInput{
		Filters: []awstypes.DescribeActivationsFilter{
			{
				FilterKey:    awstypes.DescribeActivationsFilterKeysActivationIds,
				FilterValues: []string{id},
			},
		},
	}

	return findActivation(ctx, conn, input)
}

func findActivation(ctx context.Context, conn *ssm.Client, input *ssm.DescribeActivationsInput) (*awstypes.Activation, error) {
	output, err := findActivations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findActivations(ctx context.Context, conn *ssm.Client, input *ssm.DescribeActivationsInput) ([]awstypes.Activation, error) {
	var output []awstypes.Activation

	pages := ssm.NewDescribeActivationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.ActivationList...)
	}

	return output, nil
}
