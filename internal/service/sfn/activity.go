// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sfn

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sfn_activity", name="Activity")
// @Tags(identifierAttribute="id")
func resourceActivity() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceActivityCreate,
		ReadWithoutTimeout:   resourceActivityRead,
		UpdateWithoutTimeout: resourceActivityUpdate,
		DeleteWithoutTimeout: resourceActivityDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrCreationDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEncryptionConfiguration: {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kms_data_key_reuse_period_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(60, 900),
						},
						names.AttrKMSKeyID: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.EncryptionType](),
						},
					},
				},
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 80),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceActivityCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SFNClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &sfn.CreateActivityInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrEncryptionConfiguration); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.EncryptionConfiguration = expandEncryptionConfiguration(v.([]any)[0].(map[string]any))
	}

	output, err := conn.CreateActivity(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Step Functions Activity (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ActivityArn))

	return append(diags, resourceActivityRead(ctx, d, meta)...)
}

func resourceActivityRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SFNClient(ctx)

	output, err := findActivityByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Step Functions Activity (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Step Functions Activity (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrCreationDate, output.CreationDate.Format(time.RFC3339))
	if output.EncryptionConfiguration != nil {
		if err := d.Set(names.AttrEncryptionConfiguration, []any{flattenEncryptionConfiguration(output.EncryptionConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting encryption_configuration: %s", err)
		}
	} else {
		d.Set(names.AttrEncryptionConfiguration, nil)
	}
	d.Set(names.AttrName, output.Name)

	return diags
}

func resourceActivityUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Tags only.
	return resourceActivityRead(ctx, d, meta)
}

func resourceActivityDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SFNClient(ctx)

	log.Printf("[DEBUG] Deleting Step Functions Activity: %s", d.Id())
	_, err := conn.DeleteActivity(ctx, &sfn.DeleteActivityInput{
		ActivityArn: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Step Functions Activity (%s): %s", d.Id(), err)
	}

	return diags
}

func findActivityByARN(ctx context.Context, conn *sfn.Client, arn string) (*sfn.DescribeActivityOutput, error) {
	input := &sfn.DescribeActivityInput{
		ActivityArn: aws.String(arn),
	}

	output, err := conn.DescribeActivity(ctx, input)

	if errs.IsA[*awstypes.ActivityDoesNotExist](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.CreationDate == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
