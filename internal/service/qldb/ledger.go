// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qldb

import (
	"context"
	"log"
	"regexp"
	"time"

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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"kms_key": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.Any(
					validation.StringInSlice([]string{"AWS_OWNED_KMS_KEY"}, false),
					verify.ValidARN,
				),
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
					validation.StringMatch(regexp.MustCompile(`^[A-Za-z0-9_-]+`), "must contain only alphanumeric characters, underscores, and hyphens"),
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
	conn := meta.(*conns.AWSClient).QLDBClient(ctx)

	name := create.Name(d.Get("name").(string), "tf")
	input := &qldb.CreateLedgerInput{
		DeletionProtection: aws.Bool(d.Get("deletion_protection").(bool)),
		Name:               aws.String(name),
		PermissionsMode:    types.PermissionsMode(d.Get("permissions_mode").(string)),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("kms_key"); ok {
		input.KmsKey = aws.String(v.(string))
	}

	output, err := conn.CreateLedger(ctx, input)

	if err != nil {
		return diag.Errorf("creating QLDB Ledger (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Name))

	if _, err := waitLedgerCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for QLDB Ledger (%s) create: %s", d.Id(), err)
	}

	return resourceLedgerRead(ctx, d, meta)
}

func resourceLedgerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).QLDBClient(ctx)

	ledger, err := findLedgerByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] QLDB Ledger %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading QLDB Ledger (%s): %s", d.Id(), err)
	}

	d.Set("arn", ledger.Arn)
	d.Set("deletion_protection", ledger.DeletionProtection)
	if ledger.EncryptionDescription != nil {
		d.Set("kms_key", ledger.EncryptionDescription.KmsKeyArn)
	} else {
		d.Set("kms_key", nil)
	}
	d.Set("name", ledger.Name)
	d.Set("permissions_mode", ledger.PermissionsMode)

	return nil
}

func resourceLedgerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).QLDBClient(ctx)

	if d.HasChange("permissions_mode") {
		input := &qldb.UpdateLedgerPermissionsModeInput{
			Name:            aws.String(d.Id()),
			PermissionsMode: types.PermissionsMode(d.Get("permissions_mode").(string)),
		}

		if _, err := conn.UpdateLedgerPermissionsMode(ctx, input); err != nil {
			return diag.Errorf("updating QLDB Ledger (%s) permissions mode: %s", d.Id(), err)
		}
	}

	if d.HasChanges("deletion_protection", "kms_key") {
		input := &qldb.UpdateLedgerInput{
			DeletionProtection: aws.Bool(d.Get("deletion_protection").(bool)),
			Name:               aws.String(d.Id()),
		}

		if d.HasChange("kms_key") {
			input.KmsKey = aws.String(d.Get("kms_key").(string))
		}

		if _, err := conn.UpdateLedger(ctx, input); err != nil {
			return diag.Errorf("updating QLDB Ledger (%s): %s", d.Id(), err)
		}
	}

	return resourceLedgerRead(ctx, d, meta)
}

func resourceLedgerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).QLDBClient(ctx)

	input := &qldb.DeleteLedgerInput{
		Name: aws.String(d.Id()),
	}

	log.Printf("[INFO] Deleting QLDB Ledger: %s", d.Id())
	_, err := tfresource.RetryWhenIsA[*types.ResourceInUseException](ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return conn.DeleteLedger(ctx, input)
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting QLDB Ledger (%s): %s", d.Id(), err)
	}

	if _, err := waitLedgerDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for QLDB Ledger (%s) delete: %s", d.Id(), err)
	}

	return nil
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
