// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudtrail

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types/nullable"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudtrail_event_data_store", name="Event Data Store")
// @Tags(identifierAttribute="id")
func resourceEventDataStore() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEventDataStoreCreate,
		ReadWithoutTimeout:   resourceEventDataStoreRead,
		UpdateWithoutTimeout: resourceEventDataStoreUpdate,
		DeleteWithoutTimeout: resourceEventDataStoreDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"advanced_event_selector": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field_selector": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ends_with": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 2048),
										},
									},
									"equals": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 2048),
										},
									},
									names.AttrField: {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.StringInSlice(field_Values(), false),
									},
									"not_ends_with": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 2048),
										},
									},
									"not_equals": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 2048),
										},
									},
									"not_starts_with": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 2048),
										},
									},
									"starts_with": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 2048),
										},
									},
								},
							},
						},
						names.AttrName: {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringLenBetween(0, 1000),
						},
					},
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"billing_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          types.BillingModeExtendableRetentionPricing,
				ValidateDiagFunc: enum.Validate[types.BillingMode](),
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"multi_region_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(3, 128),
			},
			"organization_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrRetentionPeriod: {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  2555,
				ValidateFunc: validation.All(
					validation.IntBetween(7, 2555),
				),
			},
			"suspend": {
				Type:         nullable.TypeNullableBool,
				Optional:     true,
				ValidateFunc: nullable.ValidateTypeStringNullableBool,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"termination_protection_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceEventDataStoreCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudTrailClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &cloudtrail.CreateEventDataStoreInput{
		BillingMode:                  types.BillingMode(d.Get("billing_mode").(string)),
		MultiRegionEnabled:           aws.Bool(d.Get("multi_region_enabled").(bool)),
		Name:                         aws.String(name),
		OrganizationEnabled:          aws.Bool(d.Get("organization_enabled").(bool)),
		RetentionPeriod:              aws.Int32(int32(d.Get(names.AttrRetentionPeriod).(int))),
		TagsList:                     getTagsIn(ctx),
		TerminationProtectionEnabled: aws.Bool(d.Get("termination_protection_enabled").(bool)),
	}

	if _, ok := d.GetOk("advanced_event_selector"); ok {
		input.AdvancedEventSelectors = expandAdvancedEventSelector(d.Get("advanced_event_selector").([]any))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("suspend"); ok {
		if v, null, _ := nullable.Bool(v.(string)).ValueBool(); !null {
			input.StartIngestion = aws.Bool(!v)
		}
	}

	output, err := conn.CreateEventDataStore(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudTrail Event Data Store (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.EventDataStoreArn))

	if _, err := waitEventDataStoreCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudTrail Event Data Store (%s) create: %s", name, err)
	}

	return append(diags, resourceEventDataStoreRead(ctx, d, meta)...)
}

func resourceEventDataStoreRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudTrailClient(ctx)

	output, err := findEventDataStoreByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudTrail Event Data Store (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudTrail Event Data Store (%s): %s", d.Id(), err)
	}

	if err := d.Set("advanced_event_selector", flattenAdvancedEventSelector(output.AdvancedEventSelectors)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting advanced_event_selector: %s", err)
	}
	d.Set(names.AttrARN, output.EventDataStoreArn)
	d.Set(names.AttrKMSKeyID, output.KmsKeyId)
	d.Set("billing_mode", output.BillingMode)
	d.Set("multi_region_enabled", output.MultiRegionEnabled)
	d.Set(names.AttrName, output.Name)
	d.Set("organization_enabled", output.OrganizationEnabled)
	d.Set(names.AttrRetentionPeriod, output.RetentionPeriod)
	d.Set("termination_protection_enabled", output.TerminationProtectionEnabled)

	return diags
}

func resourceEventDataStoreUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudTrailClient(ctx)

	if d.HasChangesExcept("suspend", names.AttrTags, names.AttrTagsAll) {
		input := &cloudtrail.UpdateEventDataStoreInput{
			EventDataStore: aws.String(d.Id()),
		}

		if d.HasChange("advanced_event_selector") {
			input.AdvancedEventSelectors = expandAdvancedEventSelector(d.Get("advanced_event_selector").([]any))
		}

		if d.HasChange("billing_mode") {
			input.BillingMode = types.BillingMode(d.Get("billing_mode").(string))
		}

		if d.HasChange("multi_region_enabled") {
			input.MultiRegionEnabled = aws.Bool(d.Get("multi_region_enabled").(bool))
		}

		if d.HasChange(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		if d.HasChange("organization_enabled") {
			input.OrganizationEnabled = aws.Bool(d.Get("organization_enabled").(bool))
		}

		if d.HasChange(names.AttrRetentionPeriod) {
			input.RetentionPeriod = aws.Int32(int32(d.Get(names.AttrRetentionPeriod).(int)))
		}

		if d.HasChange("termination_protection_enabled") {
			input.TerminationProtectionEnabled = aws.Bool(d.Get("termination_protection_enabled").(bool))
		}

		_, err := conn.UpdateEventDataStore(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CloudTrail Event Data Store (%s): %s", d.Id(), err)
		}

		if _, err := waitEventDataStoreUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for CloudTrail Event Data Store (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("suspend") {
		v := d.Get("suspend")
		if v, null, _ := nullable.Bool(v.(string)).ValueBool(); !null {
			if v {
				if err := stopEventDataStoreIngestion(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}
			} else {
				if err := startEventDataStoreIngestion(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}
			}
		}
	}

	return append(diags, resourceEventDataStoreRead(ctx, d, meta)...)
}

func resourceEventDataStoreDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudTrailClient(ctx)

	log.Printf("[DEBUG] Deleting CloudTrail Event Data Store: %s", d.Id())
	input := cloudtrail.DeleteEventDataStoreInput{
		EventDataStore: aws.String(d.Id()),
	}
	_, err := conn.DeleteEventDataStore(ctx, &input)

	if errs.IsA[*types.EventDataStoreNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudTrail Event Data Store (%s): %s", d.Id(), err)
	}

	if _, err := waitEventDataStoreDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudTrail Event Data Store (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findEventDataStoreByARN(ctx context.Context, conn *cloudtrail.Client, arn string) (*cloudtrail.GetEventDataStoreOutput, error) {
	input := cloudtrail.GetEventDataStoreInput{
		EventDataStore: aws.String(arn),
	}

	output, err := findEventDataStore(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if status := output.Status; status == types.EventDataStoreStatusPendingDeletion {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output, nil
}

func findEventDataStore(ctx context.Context, conn *cloudtrail.Client, input *cloudtrail.GetEventDataStoreInput) (*cloudtrail.GetEventDataStoreOutput, error) {
	output, err := conn.GetEventDataStore(ctx, input)

	if errs.IsA[*types.EventDataStoreNotFoundException](err) {
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

func statusEventDataStore(ctx context.Context, conn *cloudtrail.Client, arn string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findEventDataStoreByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func startEventDataStoreIngestion(ctx context.Context, conn *cloudtrail.Client, arn string, timeout time.Duration) error {
	input := cloudtrail.StartEventDataStoreIngestionInput{
		EventDataStore: aws.String(arn),
	}

	_, err := conn.StartEventDataStoreIngestion(ctx, &input)

	if err != nil {
		return fmt.Errorf("starting CloudTrail Event Data Store (%s) ingestion: %w", arn, err)
	}

	if _, err := waitEventDataStoreIngestionStarted(ctx, conn, arn, timeout); err != nil {
		return fmt.Errorf("waiting for CloudTrail Event Data Store (%s) ingestion start: %w", arn, err)
	}

	return nil
}

func stopEventDataStoreIngestion(ctx context.Context, conn *cloudtrail.Client, arn string, timeout time.Duration) error {
	input := cloudtrail.StopEventDataStoreIngestionInput{
		EventDataStore: aws.String(arn),
	}

	_, err := conn.StopEventDataStoreIngestion(ctx, &input)

	if err != nil {
		return fmt.Errorf("stopping CloudTrail Event Data Store (%s) ingestion: %w", arn, err)
	}

	if _, err := waitEventDataStoreIngestionStopped(ctx, conn, arn, timeout); err != nil {
		return fmt.Errorf("waiting for CloudTrail Event Data Store (%s) ingestion stop: %w", arn, err)
	}

	return nil
}

func waitEventDataStoreCreated(ctx context.Context, conn *cloudtrail.Client, arn string, timeout time.Duration) (*cloudtrail.GetEventDataStoreOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.EventDataStoreStatusCreated),
		Target:  enum.Slice(types.EventDataStoreStatusEnabled, types.EventDataStoreStatusStoppedIngestion),
		Refresh: statusEventDataStore(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudtrail.GetEventDataStoreOutput); ok {
		return output, err
	}

	return nil, err
}

func waitEventDataStoreUpdated(ctx context.Context, conn *cloudtrail.Client, arn string, timeout time.Duration) (*cloudtrail.GetEventDataStoreOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.EventDataStoreStatusCreated),
		Target:  enum.Slice(types.EventDataStoreStatusEnabled),
		Refresh: statusEventDataStore(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudtrail.GetEventDataStoreOutput); ok {
		return output, err
	}

	return nil, err
}

func waitEventDataStoreDeleted(ctx context.Context, conn *cloudtrail.Client, arn string, timeout time.Duration) (*cloudtrail.GetEventDataStoreOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.EventDataStoreStatusCreated, types.EventDataStoreStatusEnabled),
		Target:  []string{},
		Refresh: statusEventDataStore(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudtrail.GetEventDataStoreOutput); ok {
		return output, err
	}

	return nil, err
}

func waitEventDataStoreIngestionStarted(ctx context.Context, conn *cloudtrail.Client, arn string, timeout time.Duration) (*cloudtrail.GetEventDataStoreOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.EventDataStoreStatusStartingIngestion),
		Target:  enum.Slice(types.EventDataStoreStatusEnabled),
		Refresh: statusEventDataStore(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudtrail.GetEventDataStoreOutput); ok {
		return output, err
	}

	return nil, err
}

func waitEventDataStoreIngestionStopped(ctx context.Context, conn *cloudtrail.Client, arn string, timeout time.Duration) (*cloudtrail.GetEventDataStoreOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.EventDataStoreStatusStoppingIngestion),
		Target:  enum.Slice(types.EventDataStoreStatusStoppedIngestion),
		Refresh: statusEventDataStore(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudtrail.GetEventDataStoreOutput); ok {
		return output, err
	}

	return nil, err
}
