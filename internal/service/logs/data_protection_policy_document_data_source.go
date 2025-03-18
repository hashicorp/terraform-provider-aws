// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cloudwatch_log_data_protection_policy_document", name="Data Protection Policy Document")
func dataSourceDataProtectionPolicyDocument() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDataProtectionPolicyDocumentRead,

		Schema: map[string]*schema.Schema{
			names.AttrConfiguration: {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"custom_data_identifier": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 10,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringIsNotEmpty,
											validation.StringLenBetween(1, 128),
										),
									},
									"regex": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringIsNotEmpty,
											validation.StringLenBetween(1, 200),
										),
									},
								},
							},
						},
					},
				},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrJSON: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"statement": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 2,
				MaxItems: 2,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_identifiers": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"operation": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"audit": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"findings_destination": {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrCloudWatchLogs: {
																Type:     schema.TypeList,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"log_group": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringIsNotEmpty,
																		},
																	},
																},
															},
															"firehose": {
																Type:     schema.TypeList,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"delivery_stream": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringIsNotEmpty,
																		},
																	},
																},
															},
															"s3": {
																Type:     schema.TypeList,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		names.AttrBucket: {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringIsNotEmpty,
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
									"deidentify": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"mask_config": {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{},
													},
												},
											},
										},
									},
								},
							},
						},
						"sid": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			names.AttrVersion: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "2021-06-01",
			},
		},
	}
}

func dataSourceDataProtectionPolicyDocumentRead(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	var diags diag.Diagnostics

	document := dataProtectionPolicyDocument{
		Description: d.Get(names.AttrDescription).(string),
		Name:        d.Get(names.AttrName).(string),
		Version:     d.Get(names.AttrVersion).(string),
	}

	// unwrap expects v to be a configuration block -- a TypeList schema
	// element with MaxItems: 1 and with a sub-schema.
	unwrap := func(v any) (map[string]any, bool) {
		if v == nil {
			return nil, false
		}

		if tfList, ok := v.([]any); ok && len(tfList) > 0 {
			if tfList[0] == nil {
				// Configuration block was present, but the sub-schema is empty.
				return map[string]any{}, true
			}

			if tfMap, ok := tfList[0].(map[string]any); ok && tfMap != nil {
				// This should be the most typical path.
				return tfMap, true
			}
		}

		return nil, false
	}

	if tfMap, ok := unwrap(d.Get(names.AttrConfiguration)); ok {
		document.Configuration = &dataProtectionPolicyStatementConfiguration{}

		if tfList, ok := tfMap["custom_data_identifier"].([]any); ok && len(tfList) > 0 {
			for _, tfMapRaw := range tfList {
				tfMap, ok := tfMapRaw.(map[string]any)
				if !ok {
					continue
				}

				document.Configuration.CustomDataIdentifiers = append(document.Configuration.CustomDataIdentifiers, &dataProtectionPolicyCustomDataIdentifier{
					Name:  tfMap[names.AttrName].(string),
					Regex: tfMap["regex"].(string),
				})
			}
		}
	}

	for _, tfMapRaw := range d.Get("statement").([]any) {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok || tfMap == nil {
			continue
		}

		statement := &dataProtectionPolicyStatement{}
		document.Statements = append(document.Statements, statement)

		if v, ok := tfMap["sid"].(string); ok && v != "" {
			statement.Sid = v
		}

		if v, ok := tfMap["data_identifiers"].(*schema.Set); ok && v.Len() > 0 {
			statement.DataIdentifiers = flex.ExpandStringValueSet(v)
		}

		if tfMap, ok := unwrap(tfMap["operation"]); ok {
			operation := &dataProtectionPolicyStatementOperation{}
			statement.Operation = operation

			if tfMap, ok := unwrap(tfMap["audit"]); ok {
				audit := &dataProtectionPolicyStatementOperationAudit{}
				operation.Audit = audit

				if tfMap, ok := unwrap(tfMap["findings_destination"]); ok {
					findingsDestination := &dataProtectionPolicyStatementOperationAuditFindingsDestination{}
					audit.FindingsDestination = findingsDestination

					if tfMap, ok := unwrap(tfMap[names.AttrCloudWatchLogs]); ok {
						findingsDestination.CloudWatchLogs = &dataProtectionPolicyStatementOperationAuditFindingsDestinationCloudWatchLogs{
							LogGroup: tfMap["log_group"].(string),
						}
					}

					if tfMap, ok := unwrap(tfMap["firehose"]); ok {
						findingsDestination.Firehose = &dataProtectionPolicyStatementOperationAuditFindingsDestinationFirehose{
							DeliveryStream: tfMap["delivery_stream"].(string),
						}
					}

					if tfMap, ok := unwrap(tfMap["s3"]); ok {
						findingsDestination.S3 = &dataProtectionPolicyStatementOperationAuditFindingsDestinationS3{
							Bucket: tfMap[names.AttrBucket].(string),
						}
					}
				}
			}

			if tfMap, ok := unwrap(tfMap["deidentify"]); ok {
				deidentify := &dataProtectionPolicyStatementOperationDeidentify{}
				operation.Deidentify = deidentify

				if _, ok := unwrap(tfMap["mask_config"]); ok {
					maskConfig := &dataProtectionPolicyStatementOperationDeidentifyMaskConfig{}
					deidentify.MaskConfig = maskConfig

					// No fields in this object.
				}
			}
		}
	}

	// The schema requires exactly 2 elements, which is assumed here.

	if op := document.Statements[0].Operation; op.Audit == nil || op.Deidentify != nil {
		return sdkdiag.AppendErrorf(diags, "the first policy statement must contain only the audit operation")
	}

	if op := document.Statements[1].Operation; op.Audit != nil || op.Deidentify == nil {
		return sdkdiag.AppendErrorf(diags, "the second policy statement must contain only the deidentify operation")
	}

	jsonString, err := tfjson.EncodeToStringIndent(document, "", "  ")

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrJSON, jsonString)
	d.SetId(strconv.Itoa(create.StringHashcode(jsonString)))

	return diags
}

type dataProtectionPolicyDocument struct {
	Configuration *dataProtectionPolicyStatementConfiguration `json:",omitempty"`
	Description   string                                      `json:",omitempty"`
	Name          string                                      `json:",omitempty"`
	Statements    []*dataProtectionPolicyStatement            `json:"Statement,omitempty"`
	Version       string                                      `json:",omitempty"`
}

type dataProtectionPolicyStatementConfiguration struct {
	CustomDataIdentifiers []*dataProtectionPolicyCustomDataIdentifier `json:"CustomDataIdentifier,omitempty"`
}

type dataProtectionPolicyCustomDataIdentifier struct {
	Name  string `json:",omitempty"`
	Regex string `json:",omitempty"`
}

type dataProtectionPolicyStatement struct {
	Sid             string                                  `json:",omitempty"`
	DataIdentifiers []string                                `json:"DataIdentifier,omitempty"`
	Operation       *dataProtectionPolicyStatementOperation `json:",omitempty"`
}

type dataProtectionPolicyStatementOperation struct {
	Audit      *dataProtectionPolicyStatementOperationAudit      `json:",omitempty"`
	Deidentify *dataProtectionPolicyStatementOperationDeidentify `json:",omitempty"`
}

type dataProtectionPolicyStatementOperationAudit struct {
	FindingsDestination *dataProtectionPolicyStatementOperationAuditFindingsDestination `json:",omitempty"`
}

type dataProtectionPolicyStatementOperationAuditFindingsDestination struct {
	CloudWatchLogs *dataProtectionPolicyStatementOperationAuditFindingsDestinationCloudWatchLogs `json:",omitempty"`
	Firehose       *dataProtectionPolicyStatementOperationAuditFindingsDestinationFirehose       `json:",omitempty"`
	S3             *dataProtectionPolicyStatementOperationAuditFindingsDestinationS3             `json:",omitempty"`
}

type dataProtectionPolicyStatementOperationAuditFindingsDestinationCloudWatchLogs struct {
	LogGroup string `json:",omitempty"`
}

type dataProtectionPolicyStatementOperationAuditFindingsDestinationFirehose struct {
	DeliveryStream string `json:",omitempty"`
}

type dataProtectionPolicyStatementOperationAuditFindingsDestinationS3 struct {
	Bucket string `json:",omitempty"`
}

type dataProtectionPolicyStatementOperationDeidentify struct {
	MaskConfig *dataProtectionPolicyStatementOperationDeidentifyMaskConfig `json:",omitempty"`
}

type dataProtectionPolicyStatementOperationDeidentifyMaskConfig struct{}
