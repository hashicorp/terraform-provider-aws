// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package xray

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/xray"
	"github.com/aws/aws-sdk-go-v2/service/xray/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_xray_encryption_config", name="Encryption Config")
// @SingletonIdentity
// @V60SDKv2Fix
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/xray/types;awstypes;awstypes.EncryptionConfig")
// @Testing(generator=false)
// @Testing(checkDestroyNoop=true)
func resourceEncryptionConfig() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEncryptionPutConfig,
		ReadWithoutTimeout:   resourceEncryptionConfigRead,
		UpdateWithoutTimeout: resourceEncryptionPutConfig,
		DeleteWithoutTimeout: schema.NoopContext,

		Schema: map[string]*schema.Schema{
			names.AttrKeyID: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrType: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.EncryptionType](),
			},
		},
	}
}

func resourceEncryptionPutConfig(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).XRayClient(ctx)

	input := xray.PutEncryptionConfigInput{
		Type: types.EncryptionType(d.Get(names.AttrType).(string)),
	}

	if v, ok := d.GetOk(names.AttrKeyID); ok {
		input.KeyId = aws.String(v.(string))
	}

	_, err := conn.PutEncryptionConfig(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating XRay Encryption Config: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))

	if _, err := waitEncryptionConfigAvailable(ctx, conn); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for XRay Encryption Config (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceEncryptionConfigRead(ctx, d, meta)...)
}

func resourceEncryptionConfigRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).XRayClient(ctx)

	config, err := findEncryptionConfig(ctx, conn)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] XRay Encryption Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading XRay Encryption Config (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrKeyID, config.KeyId)
	d.Set(names.AttrType, config.Type)

	return diags
}

func findEncryptionConfig(ctx context.Context, conn *xray.Client) (*types.EncryptionConfig, error) {
	input := xray.GetEncryptionConfigInput{}

	output, err := conn.GetEncryptionConfig(ctx, &input)

	if err != nil {
		return nil, err
	}

	if output == nil || output.EncryptionConfig == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.EncryptionConfig, nil
}

func statusEncryptionConfig(conn *xray.Client) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findEncryptionConfig(ctx, conn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitEncryptionConfigAvailable(ctx context.Context, conn *xray.Client) (*types.EncryptionConfig, error) {
	const (
		timeout = 15 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.EncryptionStatusUpdating),
		Target:  enum.Slice(types.EncryptionStatusActive),
		Refresh: statusEncryptionConfig(conn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.EncryptionConfig); ok {
		return output, err
	}

	return nil, err
}
