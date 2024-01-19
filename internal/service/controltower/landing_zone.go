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
	d.Set("manifest", landingZone.Manifest)
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
	AccessManagement      landingZoneManifestAccessManagement      `json:"accessManagement,omitempty"`
	CentralizedLogging    landingZoneManifestCentralizedLogging    `json:"centralizedLogging,omitempty"`
	GovernedRegions       []string                                 `json:"governedRegions,omitempty"`
	OrganizationStructure landingZoneManifestOrganizationStructure `json:"organizationStructure,omitempty"`
	SecurityRoles         landingZoneManifestSecurityRoles         `json:"securityRoles,omitempty"`
}

type landingZoneManifestAccessManagement struct {
	Enabled bool `json:"enabled,omitempty"`
}

type landingZoneManifestCentralizedLogging struct {
	AccountID      string                                              `json:"accountId,omitempty"`
	Configurations landingZoneManifestCentralizedLoggingConfigurations `json:"configurations,omitempty"`
	Enabled        bool                                                `json:"enabled,omitempty"`
}

type landingZoneManifestCentralizedLoggingConfigurations struct {
	AccessLoggingBucket landingZoneManifestCentralizedLoggingConfigurationsAccessLoggingBucket `json:"accessLoggingBucket,omitempty"`
	KmsKeyARN           string                                                                 `json:"kmsKeyArn,omitempty"`
	LoggingBucket       landingZoneManifestCentralizedLoggingConfigurationsLoggingBucket       `json:"loggingBucket,omitempty"`
}

type landingZoneManifestCentralizedLoggingConfigurationsAccessLoggingBucket struct {
	RetentionDays int `json:"retentionDays,omitempty"`
}

type landingZoneManifestCentralizedLoggingConfigurationsLoggingBucket struct {
	RetentionDays int `json:"retentionDays,omitempty"`
}

type landingZoneManifestOrganizationStructure struct {
	Sandbox  landingZoneManifestOrganizationStructureSandbox  `json:"sandbox,omitempty"`
	Security landingZoneManifestOrganizationStructureSecurity `json:"security,omitempty"`
}

type landingZoneManifestOrganizationStructureSandbox struct {
	Name string `json:"name,omitempty"`
}

type landingZoneManifestOrganizationStructureSecurity struct {
	Name string `json:"name,omitempty"`
}

type landingZoneManifestSecurityRoles struct {
	AccountID string `json:"accountId,omitempty"`
}

func expandLandingZoneManifest(tfMap map[string]interface{}) *landingZoneManifest {
	if tfMap == nil {
		return nil
	}

	apiObject := &landingZoneManifest{}

	if v, ok := tfMap["governed_regions"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.GovernedRegions = flex.ExpandStringValueSet(v)
	}

	return apiObject
}
