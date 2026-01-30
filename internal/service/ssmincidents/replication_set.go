// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ssmincidents

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ssmincidents"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssmincidents/types"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ssmincidents_replication_set", name="Replication Set")
// @Tags(identifierAttribute="id")
// @Region(overrideEnabled=false)
func resourceReplicationSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceReplicationSetCreate,
		ReadWithoutTimeout:   resourceReplicationSetRead,
		UpdateWithoutTimeout: resourceReplicationSetUpdate,
		DeleteWithoutTimeout: resourceReplicationSetDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Update: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deletion_protected": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"last_modified_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrRegion: {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKMSKeyARN: {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          "DefaultKey",
							ValidateDiagFunc: validateNonAliasARN,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrStatus: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatusMessage: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Deprecated: "region is deprecated. Use regions instead.",
			},
			"regions": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKMSKeyARN: {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          "DefaultKey",
							ValidateDiagFunc: validateNonAliasARN,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrStatus: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatusMessage: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceReplicationSetImport,
		},

		CustomizeDiff: func(ctx context.Context, diff *schema.ResourceDiff, meta any) error {
			config := diff.GetRawConfig()
			v := config.GetAttr("regions")
			regionsConfigured := v.IsKnown() && !v.IsNull() && v.LengthInt() > 0
			v = config.GetAttr(names.AttrRegion)
			regionConfigured := v.IsKnown() && !v.IsNull() && v.LengthInt() > 0
			if regionsConfigured == regionConfigured {
				return errors.New("exactly one of `region` or `regions` must be set")
			}

			return nil
		},
	}
}

func resourceReplicationSetCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMIncidentsClient(ctx)

	input := ssmincidents.CreateReplicationSetInput{
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk("regions"); ok && v.(*schema.Set).Len() > 0 {
		input.Regions = expandRegionMapInputValues(v.(*schema.Set).List())
	} else if v, ok := d.GetOk(names.AttrRegion); ok && v.(*schema.Set).Len() > 0 {
		input.Regions = expandRegionMapInputValues(v.(*schema.Set).List())
	}

	output, err := conn.CreateReplicationSet(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SSMIncidents Replication Set: %s", err)
	}

	d.SetId(aws.ToString(output.Arn))

	if _, err := waitReplicationSetCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SSMIncidents Replication Set (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceReplicationSetRead(ctx, d, meta)...)
}

func resourceReplicationSetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMIncidentsClient(ctx)

	replicationSet, err := findReplicationSetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] SSMIncidents Replication Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSMIncidents Replication Set (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, replicationSet.Arn)
	d.Set("created_by", replicationSet.CreatedBy)
	d.Set("deletion_protected", replicationSet.DeletionProtected)
	d.Set("last_modified_by", replicationSet.LastModifiedBy)
	if err := d.Set(names.AttrRegion, flattenRegionInfos(replicationSet.RegionMap)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting region: %s", err)
	}
	if err := d.Set("regions", flattenRegionInfos(replicationSet.RegionMap)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting regions: %s", err)
	}
	d.Set(names.AttrStatus, replicationSet.Status)

	return diags
}

func resourceReplicationSetUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMIncidentsClient(ctx)

	if d.HasChanges(names.AttrRegion, "regions") {
		input, err := expandUpdateReplicationSetInput(d)

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		_, err = conn.UpdateReplicationSet(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SSMIncidents Replication Set (%s): %s", d.Id(), err)
		}

		if _, err := waitReplicationSetUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for SSMIncidents Replication Set (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceReplicationSetRead(ctx, d, meta)...)
}

func resourceReplicationSetDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMIncidentsClient(ctx)

	log.Printf("[INFO] Deleting SSMIncidents Replication Set: %s", d.Id())
	input := ssmincidents.DeleteReplicationSetInput{
		Arn: aws.String(d.Id()),
	}
	_, err := conn.DeleteReplicationSet(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SSMIncidents Replication Set (%s): %s", d.Id(), err)
	}

	if _, err := waitReplicationSetDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SSMIncidents Replication Set (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func resourceReplicationSetImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).SSMIncidentsClient(ctx)

	var input ssmincidents.ListReplicationSetsInput
	arn, err := findReplicationSetARN(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	d.SetId(aws.ToString(arn))

	return []*schema.ResourceData{d}, nil
}

func findReplicationSetByID(ctx context.Context, conn *ssmincidents.Client, arn string) (*awstypes.ReplicationSet, error) {
	input := ssmincidents.GetReplicationSetInput{
		Arn: aws.String(arn),
	}
	output, err := conn.GetReplicationSet(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ReplicationSet == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.ReplicationSet, nil
}

func statusReplicationSet(conn *ssmincidents.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findReplicationSetByID(ctx, conn, arn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitReplicationSetCreated(ctx context.Context, conn *ssmincidents.Client, arn string, timeout time.Duration) (*awstypes.ReplicationSet, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ReplicationSetStatusCreating),
		Target:     enum.Slice(awstypes.ReplicationSetStatusActive),
		Refresh:    statusReplicationSet(conn, arn),
		Timeout:    timeout,
		Delay:      30 * time.Second,
		MinTimeout: 30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ReplicationSet); ok {
		return output, err
	}

	return nil, err
}

func waitReplicationSetUpdated(ctx context.Context, conn *ssmincidents.Client, arn string, timeout time.Duration) (*awstypes.ReplicationSet, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ReplicationSetStatusUpdating),
		Target:     enum.Slice(awstypes.ReplicationSetStatusActive),
		Refresh:    statusReplicationSet(conn, arn),
		Timeout:    timeout,
		Delay:      30 * time.Second,
		MinTimeout: 30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ReplicationSet); ok {
		return output, err
	}

	return nil, err
}

func waitReplicationSetDeleted(ctx context.Context, conn *ssmincidents.Client, arn string, timeout time.Duration) (*awstypes.ReplicationSet, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ReplicationSetStatusDeleting),
		Target:     []string{},
		Refresh:    statusReplicationSet(conn, arn),
		Timeout:    timeout,
		Delay:      30 * time.Second,
		MinTimeout: 30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ReplicationSet); ok {
		return output, err
	}

	return nil, err
}

// converts a list of regions to a map which maps region name to kms key arn
func regionListToKMSKeyMap(tfList []any) map[string]string {
	m := make(map[string]string)

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)
		m[tfMap[names.AttrName].(string)] = tfMap[names.AttrKMSKeyARN].(string)
	}

	return m
}

func expandUpdateReplicationSetInput(d *schema.ResourceData) (*ssmincidents.UpdateReplicationSetInput, error) {
	input := ssmincidents.UpdateReplicationSetInput{
		Arn: aws.String(d.Id()),
	}

	config := d.GetRawConfig()
	var o, n any
	if v := config.GetAttr("regions"); v.IsKnown() && !v.IsNull() && v.LengthInt() > 0 {
		o, n = d.GetChange("regions")
	} else if v := config.GetAttr(names.AttrRegion); v.IsKnown() && !v.IsNull() && v.LengthInt() > 0 {
		o, n = d.GetChange(names.AttrRegion)
	}
	oldRegions, newRegions := regionListToKMSKeyMap(o.(*schema.Set).List()), regionListToKMSKeyMap(n.(*schema.Set).List())

	for k, oldCMK := range oldRegions {
		if newCMK, ok := newRegions[k]; !ok {
			input.Actions = append(input.Actions, &awstypes.UpdateReplicationSetActionMemberDeleteRegionAction{
				Value: awstypes.DeleteRegionAction{
					RegionName: aws.String(k),
				},
			})
		} else if oldCMK != newCMK {
			return nil, fmt.Errorf("SSMIncidents Replication Set does not support updating encryption on a Region. To do this, remove the Region, and then re-create it with the new key")
		}
	}

	for region, newCMK := range newRegions {
		if _, ok := oldRegions[region]; !ok {
			action := &awstypes.UpdateReplicationSetActionMemberAddRegionAction{
				Value: awstypes.AddRegionAction{
					RegionName: aws.String(region),
				},
			}
			if newCMK != "DefaultKey" {
				action.Value.SseKmsKeyId = aws.String(newCMK)
			}

			input.Actions = append(input.Actions, action)
		}
	}

	return &input, nil
}

func expandRegionMapInputValues(tfList []any) map[string]awstypes.RegionMapInputValue {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make(map[string]awstypes.RegionMapInputValue)

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)

		apiObject := awstypes.RegionMapInputValue{}

		if kmsKey := tfMap[names.AttrKMSKeyARN].(string); kmsKey != "DefaultKey" {
			apiObject.SseKmsKeyId = aws.String(kmsKey)
		}

		apiObjects[tfMap[names.AttrName].(string)] = apiObject
	}

	return apiObjects
}

func flattenRegionInfos(apiObjects map[string]awstypes.RegionInfo) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	tfList := make([]any, 0)
	for k, apiObject := range apiObjects {
		tfMap := make(map[string]any)

		tfMap[names.AttrKMSKeyARN] = aws.ToString(apiObject.SseKmsKeyId)
		tfMap[names.AttrName] = k
		tfMap[names.AttrStatus] = apiObject.Status

		if v := apiObject.StatusMessage; v != nil {
			tfMap[names.AttrStatusMessage] = aws.ToString(v)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func validateNonAliasARN(value any, path cty.Path) diag.Diagnostics {
	parsedARN, err := arn.Parse(value.(string))
	var diags diag.Diagnostics

	if err != nil || strings.HasPrefix(parsedARN.Resource, "alias/") {
		diag := diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Invalid kms_key_arn",
			Detail:   "Failed to parse key ARN. ARN must be a valid key ARN, not a key ID, or alias ARN",
		}
		diags = append(diags, diag)
	}

	return diags
}
