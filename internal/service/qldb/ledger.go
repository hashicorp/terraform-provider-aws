// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qldb

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/qldb"
	"github.com/aws/aws-sdk-go-v2/service/qldb/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_qldb_ledger", name="Ledger")
// @Tags(identifierAttribute="arn")
func resourceLedger() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLedgerCreate,
		ReadWithoutTimeout:   resourceLedgerRead,
		UpdateWithoutTimeout: resourceLedgerUpdate,
		DeleteWithoutTimeout: resourceLedgerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDeletionProtection: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			names.AttrKMSKey: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.Any(
					validation.StringInSlice([]string{"AWS_OWNED_KMS_KEY"}, false),
					verify.ValidARN,
				),
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_-]+`), "must contain only alphanumeric characters, underscores, and hyphens"),
				),
			},
			"permissions_mode": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.PermissionsMode](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLedgerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QLDBClient(ctx)

	name := create.Name(d.Get(names.AttrName).(string), "tf")
	input := &qldb.CreateLedgerInput{
		DeletionProtection: aws.Bool(d.Get(names.AttrDeletionProtection).(bool)),
		Name:               aws.String(name),
		PermissionsMode:    types.PermissionsMode(d.Get("permissions_mode").(string)),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrKMSKey); ok {
		input.KmsKey = aws.String(v.(string))
	}

	output, err := conn.CreateLedger(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating QLDB Ledger (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Name))

	if _, err := waitLedgerCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for QLDB Ledger (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceLedgerRead(ctx, d, meta)...)
}

func resourceLedgerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QLDBClient(ctx)

	ledger, err := findLedgerByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] QLDB Ledger %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QLDB Ledger (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, ledger.Arn)
	d.Set(names.AttrDeletionProtection, ledger.DeletionProtection)
	if ledger.EncryptionDescription != nil {
		d.Set(names.AttrKMSKey, ledger.EncryptionDescription.KmsKeyArn)
	} else {
		d.Set(names.AttrKMSKey, nil)
	}
	d.Set(names.AttrName, ledger.Name)
	d.Set("permissions_mode", ledger.PermissionsMode)

	return diags
}

func resourceLedgerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QLDBClient(ctx)

	if d.HasChange("permissions_mode") {
		input := &qldb.UpdateLedgerPermissionsModeInput{
			Name:            aws.String(d.Id()),
			PermissionsMode: types.PermissionsMode(d.Get("permissions_mode").(string)),
		}

		if _, err := conn.UpdateLedgerPermissionsMode(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QLDB Ledger (%s) permissions mode: %s", d.Id(), err)
		}
	}

	if d.HasChanges(names.AttrDeletionProtection, names.AttrKMSKey) {
		input := &qldb.UpdateLedgerInput{
			DeletionProtection: aws.Bool(d.Get(names.AttrDeletionProtection).(bool)),
			Name:               aws.String(d.Id()),
		}

		if d.HasChange(names.AttrKMSKey) {
			input.KmsKey = aws.String(d.Get(names.AttrKMSKey).(string))
		}

		if _, err := conn.UpdateLedger(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QLDB Ledger (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceLedgerRead(ctx, d, meta)...)
}

func resourceLedgerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QLDBClient(ctx)

	input := &qldb.DeleteLedgerInput{
		Name: aws.String(d.Id()),
	}

	log.Printf("[INFO] Deleting QLDB Ledger: %s", d.Id())
	_, err := tfresource.RetryWhenIsA[*types.ResourceInUseException](ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return conn.DeleteLedger(ctx, input)
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting QLDB Ledger (%s): %s", d.Id(), err)
	}

	if _, err := waitLedgerDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for QLDB Ledger (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findLedgerByName(ctx context.Context, conn *qldb.Client, name string) (*qldb.DescribeLedgerOutput, error) {
	input := &qldb.DescribeLedgerInput{
		Name: aws.String(name),
	}

	output, err := conn.DescribeLedger(ctx, input)

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

	if state := output.State; state == types.LedgerStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	return output, nil
}

func statusLedgerState(ctx context.Context, conn *qldb.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findLedgerByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func waitLedgerCreated(ctx context.Context, conn *qldb.Client, name string, timeout time.Duration) (*qldb.DescribeLedgerOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.LedgerStateCreating),
		Target:     enum.Slice(types.LedgerStateActive),
		Refresh:    statusLedgerState(ctx, conn, name),
		Timeout:    timeout,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qldb.DescribeLedgerOutput); ok {
		return output, err
	}

	return nil, err
}

func waitLedgerDeleted(ctx context.Context, conn *qldb.Client, name string, timeout time.Duration) (*qldb.DescribeLedgerOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.LedgerStateActive, types.LedgerStateDeleting),
		Target:     []string{},
		Refresh:    statusLedgerState(ctx, conn, name),
		Timeout:    timeout,
		MinTimeout: 1 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qldb.DescribeLedgerOutput); ok {
		return output, err
	}

	return nil, err
}
