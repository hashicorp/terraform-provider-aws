// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"
	"log"
	"reflect"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	"github.com/aws/aws-sdk-go-v2/service/s3control/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3_account_public_access_block", name="Account Public Access Block")
func resourceAccountPublicAccessBlock() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccountPublicAccessBlockCreate,
		ReadWithoutTimeout:   resourceAccountPublicAccessBlockRead,
		UpdateWithoutTimeout: resourceAccountPublicAccessBlockUpdate,
		DeleteWithoutTimeout: resourceAccountPublicAccessBlockDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrAccountID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"block_public_acls": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"block_public_policy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"ignore_public_acls": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"restrict_public_buckets": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceAccountPublicAccessBlockCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk(names.AttrAccountID); ok {
		accountID = v.(string)
	}

	input := &s3control.PutPublicAccessBlockInput{
		AccountId: aws.String(accountID),
		PublicAccessBlockConfiguration: &types.PublicAccessBlockConfiguration{
			BlockPublicAcls:       aws.Bool(d.Get("block_public_acls").(bool)),
			BlockPublicPolicy:     aws.Bool(d.Get("block_public_policy").(bool)),
			IgnorePublicAcls:      aws.Bool(d.Get("ignore_public_acls").(bool)),
			RestrictPublicBuckets: aws.Bool(d.Get("restrict_public_buckets").(bool)),
		},
	}

	_, err := conn.PutPublicAccessBlock(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Account Public Access Block (%s): %s", accountID, err)
	}

	d.SetId(accountID)

	_, err = tfresource.RetryWhenNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return findPublicAccessBlockByAccountID(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Account Public Access Block (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceAccountPublicAccessBlockRead(ctx, d, meta)...)
}

func resourceAccountPublicAccessBlockRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	output, err := findPublicAccessBlockByAccountID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Account Public Access Block (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Account Public Access Block (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAccountID, d.Id())
	d.Set("block_public_acls", output.BlockPublicAcls)
	d.Set("block_public_policy", output.BlockPublicPolicy)
	d.Set("ignore_public_acls", output.IgnorePublicAcls)
	d.Set("restrict_public_buckets", output.RestrictPublicBuckets)

	return diags
}

func resourceAccountPublicAccessBlockUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	publicAccessBlockConfiguration := &types.PublicAccessBlockConfiguration{
		BlockPublicAcls:       aws.Bool(d.Get("block_public_acls").(bool)),
		BlockPublicPolicy:     aws.Bool(d.Get("block_public_policy").(bool)),
		IgnorePublicAcls:      aws.Bool(d.Get("ignore_public_acls").(bool)),
		RestrictPublicBuckets: aws.Bool(d.Get("restrict_public_buckets").(bool)),
	}
	input := &s3control.PutPublicAccessBlockInput{
		AccountId:                      aws.String(d.Id()),
		PublicAccessBlockConfiguration: publicAccessBlockConfiguration,
	}

	_, err := conn.PutPublicAccessBlock(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating S3 Account Public Access Block (%s): %s", d.Id(), err)
	}

	if _, err := waitPublicAccessBlockEqual(ctx, conn, d.Id(), publicAccessBlockConfiguration); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Account Public Access Block (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceAccountPublicAccessBlockRead(ctx, d, meta)...)
}

func resourceAccountPublicAccessBlockDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	log.Printf("[DEBUG] Deleting S3 Account Public Access Block: %s", d.Id())
	_, err := conn.DeletePublicAccessBlock(ctx, &s3control.DeletePublicAccessBlockInput{
		AccountId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchPublicAccessBlockConfiguration) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Account Public Access Block (%s): %s", d.Id(), err)
	}

	return diags
}

func findPublicAccessBlockByAccountID(ctx context.Context, conn *s3control.Client, accountID string) (*types.PublicAccessBlockConfiguration, error) {
	input := &s3control.GetPublicAccessBlockInput{
		AccountId: aws.String(accountID),
	}

	output, err := conn.GetPublicAccessBlock(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchPublicAccessBlockConfiguration) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.PublicAccessBlockConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.PublicAccessBlockConfiguration, nil
}

func statusPublicAccessBlockEqual(ctx context.Context, conn *s3control.Client, accountID string, target *types.PublicAccessBlockConfiguration) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findPublicAccessBlockByAccountID(ctx, conn, accountID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(reflect.DeepEqual(output, target)), nil
	}
}

func waitPublicAccessBlockEqual(ctx context.Context, conn *s3control.Client, accountID string, target *types.PublicAccessBlockConfiguration) (*types.PublicAccessBlockConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{strconv.FormatBool(false)},
		Target:                    []string{strconv.FormatBool(true)},
		Refresh:                   statusPublicAccessBlockEqual(ctx, conn, accountID, target),
		Timeout:                   propagationTimeout,
		MinTimeout:                5 * time.Second,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.PublicAccessBlockConfiguration); ok {
		return output, err
	}

	return nil, err
}
