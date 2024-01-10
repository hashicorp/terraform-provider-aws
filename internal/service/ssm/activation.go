// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ssm_activation", name="Activation")
// @Tags
func ResourceActivation() *schema.Resource {
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
			"description": {
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
			"name": {
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
	conn := meta.(*conns.AWSClient).SSMConn(ctx)

	name := d.Get("name").(string)
	input := &ssm.CreateActivationInput{
		DefaultInstanceName: aws.String(name),
		IamRole:             aws.String(d.Get("iam_role").(string)),
		Tags:                getTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("expiration_date"); ok {
		t, _ := time.Parse(time.RFC3339, v.(string))
		input.ExpirationDate = aws.Time(t)
	}

	if v, ok := d.GetOk("registration_limit"); ok {
		input.RegistrationLimit = aws.Int64(int64(v.(int)))
	}

	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreateActivationWithContext(ctx, input)
	}, "ValidationException", "Nonexistent role")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SSM Activation (%s): %s", name, err)
	}

	output := outputRaw.(*ssm.CreateActivationOutput)

	d.SetId(aws.StringValue(output.ActivationId))
	d.Set("activation_code", output.ActivationCode)

	return append(diags, resourceActivationRead(ctx, d, meta)...)
}

func resourceActivationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn(ctx)

	activation, err := FindActivationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSM Activation %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSM Activation (%s): %s", d.Id(), err)
	}

	d.Set("description", activation.Description)
	d.Set("expiration_date", aws.TimeValue(activation.ExpirationDate).Format(time.RFC3339))
	d.Set("expired", activation.Expired)
	d.Set("iam_role", activation.IamRole)
	d.Set("name", activation.DefaultInstanceName)
	d.Set("registration_count", activation.RegistrationsCount)
	d.Set("registration_limit", activation.RegistrationLimit)

	setTagsOut(ctx, activation.Tags)

	return diags
}

func resourceActivationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn(ctx)

	log.Printf("[DEBUG] Deleting SSM Activation: %s", d.Id())
	_, err := conn.DeleteActivationWithContext(ctx, &ssm.DeleteActivationInput{
		ActivationId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, ssm.ErrCodeInvalidActivation) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SSM Activation (%s): %s", d.Id(), err)
	}

	return diags
}

func FindActivationByID(ctx context.Context, conn *ssm.SSM, id string) (*ssm.Activation, error) {
	input := &ssm.DescribeActivationsInput{
		Filters: []*ssm.DescribeActivationsFilter{
			{
				FilterKey:    aws.String("ActivationIds"),
				FilterValues: aws.StringSlice([]string{id}),
			},
		},
	}

	return findActivation(ctx, conn, input)
}

func findActivation(ctx context.Context, conn *ssm.SSM, input *ssm.DescribeActivationsInput) (*ssm.Activation, error) {
	var output []*ssm.Activation

	err := conn.DescribeActivationsPagesWithContext(ctx, input, func(page *ssm.DescribeActivationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ActivationList {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}
