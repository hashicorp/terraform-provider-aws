package glacier

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glacier"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVaultLock() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVaultLockCreate,
		ReadWithoutTimeout:   resourceVaultLockRead,
		// Allow ignore_deletion_error update
		UpdateWithoutTimeout: schema.NoopContext,
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
			"policy": {
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

func resourceVaultLockCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlacierConn()
	vaultName := d.Get("vault_name").(string)

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy is invalid JSON: %s", err)
	}

	input := &glacier.InitiateVaultLockInput{
		AccountId: aws.String("-"),
		Policy: &glacier.VaultLockPolicy{
			Policy: aws.String(policy),
		},
		VaultName: aws.String(vaultName),
	}

	log.Printf("[DEBUG] Initiating Glacier Vault Lock: %s", input)
	output, err := conn.InitiateVaultLockWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "initiating Glacier Vault Lock: %s", err)
	}

	d.SetId(vaultName)

	if !d.Get("complete_lock").(bool) {
		return append(diags, resourceVaultLockRead(ctx, d, meta)...)
	}

	completeLockInput := &glacier.CompleteVaultLockInput{
		LockId:    output.LockId,
		VaultName: aws.String(vaultName),
	}

	log.Printf("[DEBUG] Completing Glacier Vault (%s) Lock: %s", vaultName, completeLockInput)
	if _, err := conn.CompleteVaultLockWithContext(ctx, completeLockInput); err != nil {
		return sdkdiag.AppendErrorf(diags, "completing Glacier Vault (%s) Lock: %s", vaultName, err)
	}

	if err := waitVaultLockCompletion(ctx, conn, vaultName); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Glacier Vault Lock (%s) completion: %s", d.Id(), err)
	}

	return append(diags, resourceVaultLockRead(ctx, d, meta)...)
}

func resourceVaultLockRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlacierConn()

	input := &glacier.GetVaultLockInput{
		AccountId: aws.String("-"),
		VaultName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Glacier Vault Lock (%s): %s", d.Id(), input)
	output, err := conn.GetVaultLockWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, glacier.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Glacier Vault Lock (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glacier Vault Lock (%s): %s", d.Id(), err)
	}

	if output == nil {
		log.Printf("[WARN] Glacier Vault Lock (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("complete_lock", aws.StringValue(output.State) == "Locked")
	d.Set("vault_name", d.Id())

	policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), aws.StringValue(output.Policy))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glacier Vault Lock (%s): setting policy: %s", d.Id(), err)
	}

	d.Set("policy", policyToSet)

	return diags
}

func resourceVaultLockDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlacierConn()

	input := &glacier.AbortVaultLockInput{
		VaultName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Aborting Glacier Vault Lock (%s): %s", d.Id(), input)
	_, err := conn.AbortVaultLockWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, glacier.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil && !d.Get("ignore_deletion_error").(bool) {
		return sdkdiag.AppendErrorf(diags, "aborting Glacier Vault Lock (%s): %s", d.Id(), err)
	}

	return diags
}

func vaultLockRefreshFunc(ctx context.Context, conn *glacier.Glacier, vaultName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &glacier.GetVaultLockInput{
			AccountId: aws.String("-"),
			VaultName: aws.String(vaultName),
		}

		log.Printf("[DEBUG] Reading Glacier Vault Lock (%s): %s", vaultName, input)
		output, err := conn.GetVaultLockWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, glacier.ErrCodeResourceNotFoundException) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("error reading Glacier Vault Lock (%s): %s", vaultName, err)
		}

		if output == nil {
			return nil, "", nil
		}

		return output, aws.StringValue(output.State), nil
	}
}

func waitVaultLockCompletion(ctx context.Context, conn *glacier.Glacier, vaultName string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"InProgress"},
		Target:  []string{"Locked"},
		Refresh: vaultLockRefreshFunc(ctx, conn, vaultName),
		Timeout: 5 * time.Minute,
	}

	log.Printf("[DEBUG] Waiting for Glacier Vault Lock (%s) completion", vaultName)
	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
