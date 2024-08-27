// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glacier

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glacier"
	"github.com/aws/aws-sdk-go-v2/service/glacier/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_glacier_vault_lock")
func resourceVaultLock() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVaultLockCreate,
		ReadWithoutTimeout:   resourceVaultLockRead,
		UpdateWithoutTimeout: schema.NoopContext, // Allow ignore_deletion_error update.
		DeleteWithoutTimeout: resourceVaultLockDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"complete_lock": {
				Type:     schema.TypeBool,
				Required: true,
				ForceNew: true,
			},
			"ignore_deletion_error": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrPolicy: {
				Type:                  schema.TypeString,
				Required:              true,
				ForceNew:              true,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				ValidateFunc:          verify.ValidIAMPolicyJSON,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"vault_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

const (
	lockStateInProgress = "InProgress"
	lockStateLocked     = "Locked"
)

func resourceVaultLockCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlacierClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	vaultName := d.Get("vault_name").(string)
	input := &glacier.InitiateVaultLockInput{
		AccountId: aws.String("-"),
		Policy: &types.VaultLockPolicy{
			Policy: aws.String(policy),
		},
		VaultName: aws.String(vaultName),
	}

	output, err := conn.InitiateVaultLock(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glacier Vault Lock (%s): %s", vaultName, err)
	}

	d.SetId(vaultName)

	if d.Get("complete_lock").(bool) {
		input := &glacier.CompleteVaultLockInput{
			LockId:    output.LockId,
			VaultName: aws.String(vaultName),
		}

		_, err := conn.CompleteVaultLock(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "completing Glacier Vault Lock (%s): %s", d.Id(), err)
		}

		if err := waitVaultLockComplete(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Glacier Vault Lock (%s) completion: %s", d.Id(), err)
		}
	}

	return append(diags, resourceVaultLockRead(ctx, d, meta)...)
}

func resourceVaultLockRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlacierClient(ctx)

	output, err := findVaultLockByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Glaier Vault Lock (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glacier Vault Lock (%s): %s", d.Id(), err)
	}

	d.Set("complete_lock", aws.ToString(output.State) == lockStateLocked)
	d.Set("vault_name", d.Id())

	policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), aws.ToString(output.Policy))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrPolicy, policyToSet)

	return diags
}

func resourceVaultLockDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlacierClient(ctx)

	log.Printf("[DEBUG] Deleting Glacier Vault Lock: %s", d.Id())
	_, err := conn.AbortVaultLock(ctx, &glacier.AbortVaultLockInput{
		VaultName: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil && !d.Get("ignore_deletion_error").(bool) {
		return sdkdiag.AppendErrorf(diags, "deleting Glacier Vault Lock (%s): %s", d.Id(), err)
	}

	return diags
}

func findVaultLockByName(ctx context.Context, conn *glacier.Client, name string) (*glacier.GetVaultLockOutput, error) {
	input := &glacier.GetVaultLockInput{
		AccountId: aws.String("-"),
		VaultName: aws.String(name),
	}

	output, err := conn.GetVaultLock(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
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

func statusLockState(ctx context.Context, conn *glacier.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findVaultLockByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.State), nil
	}
}

func waitVaultLockComplete(ctx context.Context, conn *glacier.Client, name string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{lockStateInProgress},
		Target:  []string{lockStateLocked},
		Refresh: statusLockState(ctx, conn, name),
		Timeout: 5 * time.Minute,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
