// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cloudwatch_log_data_protection_policy_document")
func dataSourceDataProtectionPolicyDocument() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDataProtectionPolicyDocumentRead,

		Schema: map[string]*schema.Schema{
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
			names.AttrVersion: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "2021-06-01",
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
		},
	}
}

const (
	DSNameDataProtectionPolicyDocument = "Data Protection Policy Document Data Source"
)

func dataSourceDataProtectionPolicyDocumentRead(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	document := DataProtectionPolicyDocument{
		Description: d.Get(names.AttrDescription).(string),
		Name:        d.Get(names.AttrName).(string),
		Version:     d.Get(names.AttrVersion).(string),
	}

	// unwrap expects m to be a configuration block -- a TypeList schema
	// element with MaxItems: 1 and with a sub-schema.
	unwrap := func(m interface{}) (map[string]interface{}, bool) {
		if m == nil {
			return nil, false
		}

		if v, ok := m.([]interface{}); ok && len(v) > 0 {
			if v[0] == nil {
				// Configuration block was present, but the sub-schema is empty.
				return map[string]interface{}{}, true
			}

			if m, ok := v[0].(map[string]interface{}); ok && m != nil {
				// This should be the most typical path.
				return m, true
			}
		}

		return nil, false
	}

	for _, statementIface := range d.Get("statement").([]interface{}) {
		m, ok := statementIface.(map[string]interface{})

		if !ok || m == nil {
			continue
		}

		statement := &DataProtectionPolicyStatement{}
		document.Statements = append(document.Statements, statement)

		if v, ok := m["sid"].(string); ok && v != "" {
			statement.Sid = v
		}

		if v, ok := m["data_identifiers"].(*schema.Set); ok && v.Len() > 0 {
			statement.DataIdentifiers = flex.ExpandStringValueSet(v)
		}

		if m, ok := unwrap(m["operation"]); ok {
			operation := &DataProtectionPolicyStatementOperation{}
			statement.Operation = operation

			if m, ok := unwrap(m["audit"]); ok {
				audit := &DataProtectionPolicyStatementOperationAudit{}
				operation.Audit = audit

				if m, ok := unwrap(m["findings_destination"]); ok {
					findingsDestination := &DataProtectionPolicyStatementOperationAuditFindingsDestination{}
					audit.FindingsDestination = findingsDestination

					if m, ok := unwrap(m[names.AttrCloudWatchLogs]); ok {
						findingsDestination.CloudWatchLogs = &DataProtectionPolicyStatementOperationAuditFindingsDestinationCloudWatchLogs{
							LogGroup: m["log_group"].(string),
						}
					}

					if m, ok := unwrap(m["firehose"]); ok {
						findingsDestination.Firehose = &DataProtectionPolicyStatementOperationAuditFindingsDestinationFirehose{
							DeliveryStream: m["delivery_stream"].(string),
						}
					}

					if m, ok := unwrap(m["s3"]); ok {
						findingsDestination.S3 = &DataProtectionPolicyStatementOperationAuditFindingsDestinationS3{
							Bucket: m[names.AttrBucket].(string),
						}
					}
				}
			}

			if m, ok := unwrap(m["deidentify"]); ok {
				deidentify := &DataProtectionPolicyStatementOperationDeidentify{}
				operation.Deidentify = deidentify

				if _, ok := unwrap(m["mask_config"]); ok {
					maskConfig := &DataProtectionPolicyStatementOperationDeidentifyMaskConfig{}
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

	jsonBytes, err := json.MarshalIndent(document, "", "  ")

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	jsonString := string(jsonBytes)

	d.Set(names.AttrJSON, jsonString)
	d.SetId(strconv.Itoa(create.StringHashcode(jsonString)))

	return diags
}

type DataProtectionPolicyDocument struct {
	Description string                           `json:",omitempty"`
	Version     string                           `json:",omitempty"`
	Name        string                           `json:",omitempty"`
	Statements  []*DataProtectionPolicyStatement `json:"Statement,omitempty"`
}

type DataProtectionPolicyStatement struct {
	Sid             string                                  `json:",omitempty"`
	DataIdentifiers []string                                `json:"DataIdentifier,omitempty"`
	Operation       *DataProtectionPolicyStatementOperation `json:",omitempty"`
}

type DataProtectionPolicyStatementOperation struct {
	Audit      *DataProtectionPolicyStatementOperationAudit      `json:",omitempty"`
	Deidentify *DataProtectionPolicyStatementOperationDeidentify `json:",omitempty"`
}

type DataProtectionPolicyStatementOperationAudit struct {
	FindingsDestination *DataProtectionPolicyStatementOperationAuditFindingsDestination `json:",omitempty"`
}

type DataProtectionPolicyStatementOperationAuditFindingsDestination struct {
	CloudWatchLogs *DataProtectionPolicyStatementOperationAuditFindingsDestinationCloudWatchLogs `json:",omitempty"`
	Firehose       *DataProtectionPolicyStatementOperationAuditFindingsDestinationFirehose       `json:",omitempty"`
	S3             *DataProtectionPolicyStatementOperationAuditFindingsDestinationS3             `json:",omitempty"`
}

type DataProtectionPolicyStatementOperationAuditFindingsDestinationCloudWatchLogs struct {
	LogGroup string `json:",omitempty"`
}

type DataProtectionPolicyStatementOperationAuditFindingsDestinationFirehose struct {
	DeliveryStream string `json:",omitempty"`
}

type DataProtectionPolicyStatementOperationAuditFindingsDestinationS3 struct {
	Bucket string `json:",omitempty"`
}

type DataProtectionPolicyStatementOperationDeidentify struct {
	MaskConfig *DataProtectionPolicyStatementOperationDeidentifyMaskConfig `json:",omitempty"`
}

type DataProtectionPolicyStatementOperationDeidentifyMaskConfig struct{}
