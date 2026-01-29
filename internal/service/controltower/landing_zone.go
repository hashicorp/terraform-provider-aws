// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package controltower

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/controltower"
	"github.com/aws/aws-sdk-go-v2/service/controltower/document"
	"github.com/aws/aws-sdk-go-v2/service/controltower/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsmithy "github.com/hashicorp/terraform-provider-aws/internal/smithy"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_controltower_landing_zone", name="Landing Zone")
// @Tags(identifierAttribute="arn")
func resourceLandingZone() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLandingZoneCreate,
		ReadWithoutTimeout:   resourceLandingZoneRead,
		UpdateWithoutTimeout: resourceLandingZoneUpdate,
		DeleteWithoutTimeout: resourceLandingZoneDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

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
			"drift_status": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrStatus: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"latest_available_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"manifest_json": {
				Type:                  schema.TypeString,
				Required:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      suppressEquivalentLandingZoneManifestDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v any) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			names.AttrVersion: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceLandingZoneCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ControlTowerClient(ctx)

	manifest, err := tfsmithy.DocumentFromJSONString(d.Get("manifest_json").(string), document.NewLazyDocument)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &controltower.CreateLandingZoneInput{
		Manifest: manifest,
		Tags:     getTagsIn(ctx),
		Version:  aws.String(d.Get(names.AttrVersion).(string)),
	}

	output, err := conn.CreateLandingZone(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ControlTower Landing Zone: %s", err)
	}

	id, err := landingZoneIDFromARN(aws.ToString(output.Arn))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.SetId(id)

	if _, err := waitLandingZoneOperationSucceeded(ctx, conn, aws.ToString(output.OperationIdentifier), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ControlTower Landing Zone (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceLandingZoneRead(ctx, d, meta)...)
}

func resourceLandingZoneRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ControlTowerClient(ctx)

	landingZone, err := findLandingZoneByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] ControlTower Landing Zone (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ControlTower Landing Zone (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, landingZone.Arn)
	if landingZone.DriftStatus != nil {
		if err := d.Set("drift_status", []any{flattenLandingZoneDriftStatusSummary(landingZone.DriftStatus)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting drift_status: %s", err)
		}
	} else {
		d.Set("drift_status", nil)
	}
	d.Set("latest_available_version", landingZone.LatestAvailableVersion)
	if landingZone.Manifest != nil {
		v, err := tfsmithy.DocumentToJSONString(landingZone.Manifest)

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		// Normalize the manifest before setting it in state
		// This ensures that governedRegions order and retentionDays type are consistent
		normalizedManifest, err := normalizeManifestJSON(v)
		if err != nil {
			log.Printf("[WARN] Failed to normalize manifest JSON: %s", err)
			d.Set("manifest_json", v) // Fall back to original if normalization fails
		} else {
			d.Set("manifest_json", normalizedManifest)
		}
	} else {
		d.Set("manifest_json", nil)
	}
	d.Set(names.AttrVersion, landingZone.Version)

	return diags
}

func resourceLandingZoneUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ControlTowerClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		manifest, err := tfsmithy.DocumentFromJSONString(d.Get("manifest_json").(string), document.NewLazyDocument)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &controltower.UpdateLandingZoneInput{
			LandingZoneIdentifier: aws.String(d.Id()),
			Manifest:              manifest,
			Version:               aws.String(d.Get(names.AttrVersion).(string)),
		}

		output, err := conn.UpdateLandingZone(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ControlTower Landing Zone (%s): %s", d.Id(), err)
		}

		if _, err := waitLandingZoneOperationSucceeded(ctx, conn, aws.ToString(output.OperationIdentifier), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for ControlTower Landing Zone (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceLandingZoneRead(ctx, d, meta)...)
}

func resourceLandingZoneDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ControlTowerClient(ctx)

	log.Printf("[DEBUG] Deleting ControlTower Landing Zone: %s", d.Id())
	input := controltower.DeleteLandingZoneInput{
		LandingZoneIdentifier: aws.String(d.Id()),
	}
	output, err := conn.DeleteLandingZone(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ControlTower Landing Zone: %s", err)
	}

	if _, err := waitLandingZoneOperationSucceeded(ctx, conn, aws.ToString(output.OperationIdentifier), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ControlTower Landing Zone (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func landingZoneIDFromARN(arnString string) (string, error) {
	arn, err := arn.Parse(arnString)
	if err != nil {
		return "", err
	}

	// arn:${Partition}:controltower:${Region}:${Account}:landingzone/${LandingZoneId}
	return strings.TrimPrefix(arn.Resource, "landingzone/"), nil
}

func findLandingZoneByID(ctx context.Context, conn *controltower.Client, id string) (*types.LandingZoneDetail, error) {
	input := &controltower.GetLandingZoneInput{
		LandingZoneIdentifier: aws.String(id),
	}

	output, err := conn.GetLandingZone(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.LandingZone == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.LandingZone, nil
}

func findLandingZoneOperationByID(ctx context.Context, conn *controltower.Client, id string) (*types.LandingZoneOperationDetail, error) {
	input := &controltower.GetLandingZoneOperationInput{
		OperationIdentifier: aws.String(id),
	}

	output, err := conn.GetLandingZoneOperation(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.OperationDetails == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.OperationDetails, nil
}

func statusLandingZoneOperation(conn *controltower.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findLandingZoneOperationByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitLandingZoneOperationSucceeded(ctx context.Context, conn *controltower.Client, id string, timeout time.Duration) (*types.LandingZoneOperationDetail, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.LandingZoneOperationStatusInProgress),
		Target:  enum.Slice(types.LandingZoneOperationStatusSucceeded),
		Refresh: statusLandingZoneOperation(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.LandingZoneOperationDetail); ok {
		if status := output.Status; status == types.LandingZoneOperationStatusFailed {
			retry.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func flattenLandingZoneDriftStatusSummary(apiObject *types.LandingZoneDriftStatusSummary) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrStatus: apiObject.Status,
	}

	return tfMap
}

// suppressEquivalentLandingZoneManifestDiffs provides custom diff suppression for Landing Zone manifests.
// It normalizes the JSON to handle two specific issues:
// 1. AWS API returns numeric fields (like retentionDays) as strings, but they should be accepted as numbers
// 2. AWS API may return arrays (like governedRegions) in different order
func suppressEquivalentLandingZoneManifestDiffs(k, old, new string, d *schema.ResourceData) bool {
	if strings.TrimSpace(old) == "" && strings.TrimSpace(new) == "" {
		return true
	}

	if strings.TrimSpace(old) == "" || strings.TrimSpace(new) == "" {
		return false
	}

	// Normalize both JSON strings and compare
	normalizedOld, err := normalizeManifestJSON(old)
	if err != nil {
		log.Printf("[WARN] Failed to normalize old manifest JSON: %s", err)
		return false
	}

	normalizedNew, err := normalizeManifestJSON(new)
	if err != nil {
		log.Printf("[WARN] Failed to normalize new manifest JSON: %s", err)
		return false
	}

	return normalizedOld == normalizedNew
}

// normalizeManifest recursively normalizes a manifest for comparison:
// - Converts numeric strings to numbers where appropriate (retentionDays)
// - Sorts arrays that should be order-independent (governedRegions)
func normalizeManifest(manifest map[string]interface{}) {
	for key, value := range manifest {
		switch key {
		case "governedRegions":
			// Sort the regions array to ignore order differences
			if regions, ok := value.([]interface{}); ok {
				strRegions := make([]string, 0, len(regions))
				for _, r := range regions {
					if str, ok := r.(string); ok {
						strRegions = append(strRegions, str)
					}
				}
				sort.Strings(strRegions)
				newRegions := make([]interface{}, len(strRegions))
				for i, r := range strRegions {
					newRegions[i] = r
				}
				manifest[key] = newRegions
			}
		case "centralizedLogging":
			// Handle nested configurations
			if logging, ok := value.(map[string]interface{}); ok {
				if configs, ok := logging["configurations"].(map[string]interface{}); ok {
					// Normalize retentionDays in loggingBucket and accessLoggingBucket
					for _, bucketKey := range []string{"loggingBucket", "accessLoggingBucket"} {
						if bucket, ok := configs[bucketKey].(map[string]interface{}); ok {
							normalizeRetentionDays(bucket)
						}
					}
				}
			}
		default:
			// Recursively normalize nested objects
			if nested, ok := value.(map[string]interface{}); ok {
				normalizeManifest(nested)
			}
		}
	}
}

// normalizeRetentionDays converts string retentionDays to numbers for comparison
func normalizeRetentionDays(bucket map[string]interface{}) {
	if retention, ok := bucket["retentionDays"]; ok {
		// Convert string to number if it's a numeric string
		if strVal, ok := retention.(string); ok {
			// Parse as float64 (JSON numbers are float64)
			if numVal, err := strconv.ParseFloat(strVal, 64); err == nil {
				bucket["retentionDays"] = numVal
			}
		}
		// If it's already a number, leave it as is
	}
}

// normalizeManifestJSON takes a JSON string, normalizes it (sorting governedRegions and
// converting retentionDays strings to numbers), and returns the normalized JSON string.
func normalizeManifestJSON(manifestJSON string) (string, error) {
	if strings.TrimSpace(manifestJSON) == "" {
		return manifestJSON, nil
	}

	var manifest map[string]interface{}
	if err := json.Unmarshal([]byte(manifestJSON), &manifest); err != nil {
		return "", err
	}

	// Normalize the manifest
	normalizeManifest(manifest)

	// Convert back to JSON string
	normalized, err := json.Marshal(manifest)
	if err != nil {
		return "", err
	}

	// Use structure.NormalizeJsonString to ensure consistent formatting
	return structure.NormalizeJsonString(string(normalized))
}
