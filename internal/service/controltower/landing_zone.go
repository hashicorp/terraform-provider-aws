// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package controltower

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/controltower"
	"github.com/aws/aws-sdk-go-v2/service/controltower/document"
	"github.com/aws/aws-sdk-go-v2/service/controltower/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"latest_available_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"manifest": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_management": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
									},
								},
							},
						},
						"centralized_logging": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"account_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"configurations": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"access_logging_bucket": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"retention_days": {
																Type:     schema.TypeInt,
																Optional: true,
																Computed: true,
															},
														},
													},
												},
												"kms_key_arn": {
													Type:     schema.TypeString,
													Optional: true,
													Computed: true,
												},
												"logging_bucket": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"retention_days": {
																Type:     schema.TypeInt,
																Optional: true,
																Computed: true,
															},
														},
													},
												},
											},
										},
									},
									"enabled": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
									},
								},
							},
						},
						"governed_regions": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"organization_structure": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"sandbox": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:     schema.TypeString,
													Optional: true,
													Computed: true,
												},
											},
										},
									},
									"security": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:     schema.TypeString,
													Optional: true,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"security_roles": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"account_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLandingZoneCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ControlTowerClient(ctx)

	input := &controltower.CreateLandingZoneInput{
		Manifest: document.NewLazyDocument(expandLandingZoneManifest(d.Get("manifest").([]interface{})[0].(map[string]interface{}))),
		Tags:     getTagsIn(ctx),
		Version:  aws.String(d.Get("version").(string)),
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

func resourceLandingZoneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ControlTowerClient(ctx)

	landingZone, err := findLandingZoneByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ControlTower Landing Zone (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ControlTower Landing Zone (%s): %s", d.Id(), err)
	}

	d.Set("arn", landingZone.Arn)
	d.Set("latest_available_version", landingZone.LatestAvailableVersion)
	if landingZone.Manifest != nil {
		var v landingZoneManifest

		if err := landingZone.Manifest.UnmarshalSmithyDocument(&v); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		if err := d.Set("manifest", []interface{}{flattenLandingZoneManifest(&v)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting manifest: %s", err)
		}
	} else {
		d.Set("manifest", nil)
	}
	d.Set("version", landingZone.Version)

	return diags
}

func resourceLandingZoneUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceLandingZoneRead(ctx, d, meta)...)
}

func resourceLandingZoneDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ControlTowerClient(ctx)

	log.Printf("[DEBUG] Deleting ControlTower Landing Zone: %s", d.Id())
	output, err := conn.DeleteLandingZone(ctx, &controltower.DeleteLandingZoneInput{
		LandingZoneIdentifier: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ControlTower Landing Zone: %s", err)
	}

	if _, err := waitLandingZoneOperationSucceeded(ctx, conn, aws.ToString(output.OperationIdentifier), d.Timeout(schema.TimeoutDelete)); err != nil {
		sdkdiag.AppendErrorf(diags, "waiting for ControlTower Landing Zone (%s) delete: %s", d.Id(), err)
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
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.LandingZone == nil {
		return nil, tfresource.NewEmptyResultError(input)
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
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.OperationDetails == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.OperationDetails, nil
}

func statusLandingZoneOperation(ctx context.Context, conn *controltower.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findLandingZoneOperationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
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
		Refresh: statusLandingZoneOperation(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.LandingZoneOperationDetail); ok {
		if status := output.Status; status == types.LandingZoneOperationStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

// https://mholt.github.io/json-to-go/
// https://docs.aws.amazon.com/controltower/latest/userguide/lz-api-launch.html
type landingZoneManifest struct {
	GovernedRegions       []string `json:"governedRegions,omitempty"`
	OrganizationStructure struct {
		Security struct {
			Name string `json:"name,omitempty"`
		} `json:"security,omitempty"`
		Sandbox struct {
			Name string `json:"name,omitempty"`
		} `json:"sandbox,omitempty"`
	} `json:"organizationStructure,omitempty"`
	CentralizedLogging struct {
		AccountID      string `json:"accountId,omitempty"`
		Configurations struct {
			LoggingBucket struct {
				RetentionDays int `json:"retentionDays,omitempty"`
			} `json:"loggingBucket,omitempty"`
			AccessLoggingBucket struct {
				RetentionDays int `json:"retentionDays,omitempty"`
			} `json:"accessLoggingBucket,omitempty"`
			KmsKeyARN string `json:"kmsKeyArn,omitempty"`
		} `json:"configurations,omitempty"`
		Enabled bool `json:"enabled,omitempty"`
	} `json:"centralizedLogging,omitempty"`
	SecurityRoles struct {
		AccountID string `json:"accountId,omitempty"`
	} `json:"securityRoles,omitempty"`
	AccessManagement struct {
		Enabled bool `json:"enabled,omitempty"`
	} `json:"accessManagement,omitempty"`
}

func expandLandingZoneManifest(tfMap map[string]interface{}) *landingZoneManifest {
	if tfMap == nil {
		return nil
	}

	apiObject := &landingZoneManifest{}

	if v, ok := tfMap["access_management"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		tfMap := v[0].(map[string]interface{})

		if v, ok := tfMap["enabled"].(bool); ok {
			apiObject.AccessManagement.Enabled = v
		}
	}

	if v, ok := tfMap["centralized_logging"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		tfMap := v[0].(map[string]interface{})

		if v, ok := tfMap["account_id"].(string); ok {
			apiObject.CentralizedLogging.AccountID = v
		}

		if v, ok := tfMap["configurations"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			tfMap := v[0].(map[string]interface{})

			if v, ok := tfMap["access_logging_bucket"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
				tfMap := v[0].(map[string]interface{})

				if v, ok := tfMap["retention_days"].(int); ok {
					apiObject.CentralizedLogging.Configurations.AccessLoggingBucket.RetentionDays = v
				}
			}

			if v, ok := tfMap["kms_key_arn"].(string); ok {
				apiObject.CentralizedLogging.Configurations.KmsKeyARN = v
			}

			if v, ok := tfMap["logging_bucket"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
				tfMap := v[0].(map[string]interface{})

				if v, ok := tfMap["retention_days"].(int); ok {
					apiObject.CentralizedLogging.Configurations.AccessLoggingBucket.RetentionDays = v
				}
			}
		}

		if v, ok := tfMap["enabled"].(bool); ok {
			apiObject.CentralizedLogging.Enabled = v
		}
	}

	if v, ok := tfMap["governed_regions"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.GovernedRegions = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["organization_structure"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		tfMap := v[0].(map[string]interface{})

		if v, ok := tfMap["sandbox"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			tfMap := v[0].(map[string]interface{})

			if v, ok := tfMap["name"].(string); ok {
				apiObject.OrganizationStructure.Sandbox.Name = v
			}
		}

		if v, ok := tfMap["security"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			tfMap := v[0].(map[string]interface{})

			if v, ok := tfMap["name"].(string); ok {
				apiObject.OrganizationStructure.Security.Name = v
			}
		}
	}

	if v, ok := tfMap["security_roles"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		tfMap := v[0].(map[string]interface{})

		if v, ok := tfMap["account_id"].(string); ok {
			apiObject.SecurityRoles.AccountID = v
		}
	}

	return apiObject
}

func flattenLandingZoneManifest(apiObject *landingZoneManifest) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"governed_regions": apiObject.GovernedRegions,
	}

	if !itypes.IsZero(&apiObject.AccessManagement) {
		tfMap["access_management"] = []interface{}{map[string]interface{}{
			"enabled": apiObject.AccessManagement.Enabled,
		}}
	}

	if !itypes.IsZero(&apiObject.CentralizedLogging) {
		tfMap["centralized_logging"] = []interface{}{map[string]interface{}{
			"account_id": apiObject.CentralizedLogging.AccountID,
			"enabled":    apiObject.CentralizedLogging.Enabled,
		}}

		if !itypes.IsZero(&apiObject.CentralizedLogging.Configurations) {
			tfMap["centralized_logging"].([]interface{})[0].(map[string]interface{})["configurations"] = []interface{}{map[string]interface{}{
				"kms_key_arn": apiObject.CentralizedLogging.Configurations.KmsKeyARN,
			}}

			if !itypes.IsZero(&apiObject.CentralizedLogging.Configurations.AccessLoggingBucket) {
				tfMap["centralized_logging"].([]interface{})[0].(map[string]interface{})["configurations"].([]interface{})[0].(map[string]interface{})["access_logging_bucket"] = []interface{}{map[string]interface{}{
					"retention_days": apiObject.CentralizedLogging.Configurations.AccessLoggingBucket.RetentionDays,
				}}
			}

			if !itypes.IsZero(&apiObject.CentralizedLogging.Configurations.LoggingBucket) {
				tfMap["centralized_logging"].([]interface{})[0].(map[string]interface{})["configurations"].([]interface{})[0].(map[string]interface{})["logging_bucket"] = []interface{}{map[string]interface{}{
					"retention_days": apiObject.CentralizedLogging.Configurations.LoggingBucket.RetentionDays,
				}}
			}
		}
	}

	if !itypes.IsZero(&apiObject.OrganizationStructure) {
		tfMap["organization_structure"] = []interface{}{map[string]interface{}{}}

		if !itypes.IsZero(&apiObject.OrganizationStructure.Sandbox) {
			tfMap["organization_structure"].([]interface{})[0].(map[string]interface{})["sandbox"] = []interface{}{map[string]interface{}{
				"name": apiObject.OrganizationStructure.Sandbox.Name,
			}}
		}

		if !itypes.IsZero(&apiObject.OrganizationStructure.Security) {
			tfMap["organization_structure"].([]interface{})[0].(map[string]interface{})["security"] = []interface{}{map[string]interface{}{
				"name": apiObject.OrganizationStructure.Security.Name,
			}}
		}
	}

	if !itypes.IsZero(&apiObject.SecurityRoles) {
		tfMap["security_roles"] = []interface{}{map[string]interface{}{
			"account_id": apiObject.SecurityRoles.AccountID,
		}}
	}

	return tfMap
}
