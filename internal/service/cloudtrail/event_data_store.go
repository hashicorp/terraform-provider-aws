// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudtrail

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudtrail_event_data_store", name="Event Data Store")
// @Tags(identifierAttribute="id")
func ResourceEventDataStore() *schema.Resource {
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

		CustomizeDiff: verify.SetTagsDiff,

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
									"field": {
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
						"name": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringLenBetween(0, 1000),
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"multi_region_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"name": {
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
			"retention_period": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  2555,
				ValidateFunc: validation.All(
					validation.IntBetween(7, 2555),
				),
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

func resourceEventDataStoreCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CloudTrailConn(ctx)

	name := d.Get("name").(string)
	input := &cloudtrail.CreateEventDataStoreInput{
		Name:                         aws.String(name),
		OrganizationEnabled:          aws.Bool(d.Get("organization_enabled").(bool)),
		MultiRegionEnabled:           aws.Bool(d.Get("multi_region_enabled").(bool)),
		TerminationProtectionEnabled: aws.Bool(d.Get("termination_protection_enabled").(bool)),
		RetentionPeriod:              aws.Int64(int64(d.Get("retention_period").(int))),
		TagsList:                     getTagsIn(ctx),
	}

	if _, ok := d.GetOk("advanced_event_selector"); ok {
		input.AdvancedEventSelectors = expandAdvancedEventSelector(d.Get("advanced_event_selector").([]interface{}))
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	output, err := conn.CreateEventDataStoreWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudTrail Event Data Store (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.EventDataStoreArn))

	if err := waitEventDataStoreAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudTrail Event Data Store (%s) create: %s", name, err)
	}

	return append(diags, resourceEventDataStoreRead(ctx, d, meta)...)
}

func resourceEventDataStoreRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CloudTrailConn(ctx)

	eventDataStore, err := FindEventDataStoreByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudTrail Event Data Store (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudTrail Event Data Store (%s): %s", d.Id(), err)
	}

	if err := d.Set("advanced_event_selector", flattenAdvancedEventSelector(eventDataStore.AdvancedEventSelectors)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting advanced_event_selector: %s", err)
	}
	d.Set("arn", eventDataStore.EventDataStoreArn)
	d.Set("kms_key_id", eventDataStore.KmsKeyId)
	d.Set("multi_region_enabled", eventDataStore.MultiRegionEnabled)
	d.Set("name", eventDataStore.Name)
	d.Set("organization_enabled", eventDataStore.OrganizationEnabled)
	d.Set("retention_period", eventDataStore.RetentionPeriod)
	d.Set("termination_protection_enabled", eventDataStore.TerminationProtectionEnabled)

	return diags
}

func resourceEventDataStoreUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CloudTrailConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &cloudtrail.UpdateEventDataStoreInput{
			EventDataStore: aws.String(d.Id()),
		}

		if d.HasChange("advanced_event_selector") {
			input.AdvancedEventSelectors = expandAdvancedEventSelector(d.Get("advanced_event_selector").([]interface{}))
		}

		if d.HasChange("multi_region_enabled") {
			input.MultiRegionEnabled = aws.Bool(d.Get("multi_region_enabled").(bool))
		}

		if d.HasChange("name") {
			input.Name = aws.String(d.Get("name").(string))
		}

		if d.HasChange("organization_enabled") {
			input.OrganizationEnabled = aws.Bool(d.Get("organization_enabled").(bool))
		}

		if d.HasChange("retention_period") {
			input.RetentionPeriod = aws.Int64(int64(d.Get("retention_period").(int)))
		}

		if d.HasChange("termination_protection_enabled") {
			input.TerminationProtectionEnabled = aws.Bool(d.Get("termination_protection_enabled").(bool))
		}

		_, err := conn.UpdateEventDataStoreWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CloudTrail Event Data Store (%s): %s", d.Id(), err)
		}

		if err := waitEventDataStoreAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for CloudTrail Event Data Store (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceEventDataStoreRead(ctx, d, meta)...)
}

func resourceEventDataStoreDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CloudTrailConn(ctx)

	log.Printf("[DEBUG] Deleting CloudTrail Event Data Store: %s", d.Id())
	_, err := conn.DeleteEventDataStoreWithContext(ctx, &cloudtrail.DeleteEventDataStoreInput{
		EventDataStore: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, cloudtrail.ErrCodeEventDataStoreNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudTrail Event Data Store (%s): %s", d.Id(), err)
	}

	if err := waitEventDataStoreDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudTrail Event Data Store (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindEventDataStoreByARN(ctx context.Context, conn *cloudtrail.CloudTrail, arn string) (*cloudtrail.GetEventDataStoreOutput, error) {
	input := cloudtrail.GetEventDataStoreInput{
		EventDataStore: aws.String(arn),
	}

	output, err := conn.GetEventDataStoreWithContext(ctx, &input)

	if tfawserr.ErrCodeEquals(err, cloudtrail.ErrCodeEventDataStoreNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if status := aws.StringValue(output.Status); status == cloudtrail.EventDataStoreStatusPendingDeletion {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output, nil
}

func statusEventDataStore(ctx context.Context, conn *cloudtrail.CloudTrail, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		eventDataStore, err := FindEventDataStoreByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return eventDataStore, aws.StringValue(eventDataStore.Status), nil
	}
}

func waitEventDataStoreAvailable(ctx context.Context, conn *cloudtrail.CloudTrail, arn string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{cloudtrail.EventDataStoreStatusCreated},
		Target:  []string{cloudtrail.EventDataStoreStatusEnabled},
		Refresh: statusEventDataStore(ctx, conn, arn),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitEventDataStoreDeleted(ctx context.Context, conn *cloudtrail.CloudTrail, arn string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{cloudtrail.EventDataStoreStatusCreated, cloudtrail.EventDataStoreStatusEnabled},
		Target:  []string{},
		Refresh: statusEventDataStore(ctx, conn, arn),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
