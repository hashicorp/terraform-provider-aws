// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var dataSourcePolicyDocumentVarReplacer = strings.NewReplacer("&{", "${")

// @SDKDataSource("aws_iam_policy_document", name="Policy Document")
func dataSourcePolicyDocument() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePolicyDocumentRead,

		SchemaFunc: func() map[string]*schema.Schema {
			principalsSchema := func() *schema.Schema {
				return &schema.Schema{
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"identifiers": {
								Type:     schema.TypeSet,
								Required: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
							names.AttrType: {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				}
			}
			setOfStringSchema := func() *schema.Schema {
				return &schema.Schema{
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				}
			}

			return map[string]*schema.Schema{
				names.AttrJSON: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"minified_json": {
					Type:     schema.TypeString,
					Computed: true,
				},
				// https://github.com/hashicorp/terraform-provider-aws/issues/31637.
				"override_json": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringIsEmpty,
					Deprecated:   "Not used",
				},
				"override_policy_documents": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: validation.StringIsJSON,
					},
				},
				"policy_id": {
					Type:     schema.TypeString,
					Optional: true,
				},
				// https://github.com/hashicorp/terraform-provider-aws/issues/31637.
				"source_json": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringIsEmpty,
					Deprecated:   "Not used",
				},
				"source_policy_documents": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: validation.StringIsJSON,
					},
				},
				"statement": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrActions: setOfStringSchema(),
							names.AttrCondition: {
								Type:     schema.TypeSet,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"test": {
											Type:     schema.TypeString,
											Required: true,
										},
										names.AttrValues: {
											Type:     schema.TypeList,
											Required: true,
											Elem: &schema.Schema{
												Type: schema.TypeString,
											},
										},
										"variable": {
											Type:     schema.TypeString,
											Required: true,
										},
									},
								},
							},
							"effect": {
								Type:         schema.TypeString,
								Optional:     true,
								Default:      "Allow",
								ValidateFunc: validation.StringInSlice([]string{"Allow", "Deny"}, false),
							},
							"not_actions":       setOfStringSchema(),
							"not_principals":    principalsSchema(),
							"not_resources":     setOfStringSchema(),
							"principals":        principalsSchema(),
							names.AttrResources: setOfStringSchema(),
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
					Default:  "2012-10-17",
					ValidateFunc: validation.StringInSlice([]string{
						"2008-10-17",
						"2012-10-17",
					}, false),
				},
			}
		},
	}
}

func dataSourcePolicyDocumentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	mergedDoc := &IAMPolicyDoc{}

	if v, ok := d.GetOk("source_policy_documents"); ok && len(v.([]interface{})) > 0 {
		// generate sid map to assure there are no duplicates in source jsons
		sidMap := make(map[string]struct{})
		for _, stmt := range mergedDoc.Statements {
			if stmt.Sid != "" {
				sidMap[stmt.Sid] = struct{}{}
			}
		}

		// merge sourceDocs in order specified
		for sourceJSONIndex, sourceJSON := range v.([]interface{}) {
			if sourceJSON == nil {
				continue
			}

			sourceDoc := &IAMPolicyDoc{}
			if err := json.Unmarshal([]byte(sourceJSON.(string)), sourceDoc); err != nil {
				return sdkdiag.AppendErrorf(diags, "writing IAM Policy Document: merging source document %d: %s", sourceJSONIndex, err)
			}

			// assure all statements in sourceDoc are unique before merging
			for stmtIndex, stmt := range sourceDoc.Statements {
				if stmt.Sid != "" {
					if _, sidExists := sidMap[stmt.Sid]; sidExists {
						return sdkdiag.AppendErrorf(diags, "writing IAM Policy Document: merging source document %d: duplicate Sid (%s) in source_policy_documents (statement %d). Remove the Sid or ensure Sids are unique.", sourceJSONIndex, stmt.Sid, stmtIndex)
					}
					sidMap[stmt.Sid] = struct{}{}
				}
			}

			mergedDoc.Merge(sourceDoc)
		}
	}

	// process the current document
	doc := &IAMPolicyDoc{
		Version: d.Get(names.AttrVersion).(string),
	}

	if policyID, hasPolicyID := d.GetOk("policy_id"); hasPolicyID {
		doc.Id = policyID.(string)
	}

	if cfgStmts, hasCfgStmts := d.GetOk("statement"); hasCfgStmts {
		var cfgStmtIntf = cfgStmts.([]interface{})
		stmts := make([]*IAMPolicyStatement, len(cfgStmtIntf))
		sidMap := make(map[string]struct{})

		for i, stmtI := range cfgStmtIntf {
			cfgStmt := stmtI.(map[string]interface{})
			stmt := &IAMPolicyStatement{
				Effect: cfgStmt["effect"].(string),
			}

			if sid, ok := cfgStmt["sid"]; ok {
				if _, ok := sidMap[sid.(string)]; ok {
					return sdkdiag.AppendErrorf(diags, "writing IAM Policy Document: duplicate Sid (%s). Remove the Sid or ensure the Sid is unique.", sid.(string))
				}
				stmt.Sid = sid.(string)
				if len(stmt.Sid) > 0 {
					sidMap[stmt.Sid] = struct{}{}
				}
			}

			if actions := cfgStmt[names.AttrActions].(*schema.Set).List(); len(actions) > 0 {
				stmt.Actions = policyDecodeConfigStringList(actions)
			}
			if actions := cfgStmt["not_actions"].(*schema.Set).List(); len(actions) > 0 {
				stmt.NotActions = policyDecodeConfigStringList(actions)
			}

			if resources := cfgStmt[names.AttrResources].(*schema.Set).List(); len(resources) > 0 {
				var err error
				stmt.Resources, err = dataSourcePolicyDocumentReplaceVarsInList(
					policyDecodeConfigStringList(resources), doc.Version,
				)
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "reading resources: %s", err)
				}
			}
			if notResources := cfgStmt["not_resources"].(*schema.Set).List(); len(notResources) > 0 {
				var err error
				stmt.NotResources, err = dataSourcePolicyDocumentReplaceVarsInList(
					policyDecodeConfigStringList(notResources), doc.Version,
				)
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "reading not_resources: %s", err)
				}
			}

			if principals := cfgStmt["principals"].(*schema.Set).List(); len(principals) > 0 {
				var err error
				stmt.Principals, err = dataSourcePolicyDocumentMakePrincipals(principals, doc.Version)
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "reading principals: %s", err)
				}
			}

			if notPrincipals := cfgStmt["not_principals"].(*schema.Set).List(); len(notPrincipals) > 0 {
				var err error
				stmt.NotPrincipals, err = dataSourcePolicyDocumentMakePrincipals(notPrincipals, doc.Version)
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "reading not_principals: %s", err)
				}
			}

			if conditions := cfgStmt[names.AttrCondition].(*schema.Set).List(); len(conditions) > 0 {
				var err error
				stmt.Conditions, err = dataSourcePolicyDocumentMakeConditions(conditions, doc.Version)
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "reading condition: %s", err)
				}
			}

			stmts[i] = stmt
		}

		doc.Statements = stmts
	}

	// merge our current document into mergedDoc
	mergedDoc.Merge(doc)

	// merge override_policy_documents policies into mergedDoc in order specified
	if v, ok := d.GetOk("override_policy_documents"); ok && len(v.([]interface{})) > 0 {
		for overrideJSONIndex, overrideJSON := range v.([]interface{}) {
			if overrideJSON == nil {
				continue
			}
			overrideDoc := &IAMPolicyDoc{}
			if err := json.Unmarshal([]byte(overrideJSON.(string)), overrideDoc); err != nil {
				return sdkdiag.AppendErrorf(diags, "writing IAM Policy Document: merging override document %d: %s", overrideJSONIndex, err)
			}

			mergedDoc.Merge(overrideDoc)
		}
	}

	jsonDoc, err := json.MarshalIndent(mergedDoc, "", "  ")
	if err != nil {
		// should never happen if the above code is correct
		return sdkdiag.AppendErrorf(diags, "writing IAM Policy Document: formatting JSON: %s", err)
	}
	jsonString := string(jsonDoc)

	d.Set(names.AttrJSON, jsonString)

	jsonMinDoc, err := json.Marshal(mergedDoc)
	if err != nil {
		// should never happen if the above code is correct
		return sdkdiag.AppendErrorf(diags, "writing IAM Policy Document: formatting JSON: %s", err)
	}
	jsonMinString := string(jsonMinDoc)

	d.Set("minified_json", jsonMinString)

	d.SetId(strconv.Itoa(create.StringHashcode(jsonString)))

	return diags
}

func dataSourcePolicyDocumentReplaceVarsInList(in interface{}, version string) (interface{}, error) {
	switch v := in.(type) {
	case string:
		if version == "2008-10-17" && strings.Contains(v, "&{") {
			return nil, fmt.Errorf("found &{ sequence in (%s), which is not supported in document version 2008-10-17", v)
		}
		return dataSourcePolicyDocumentVarReplacer.Replace(v), nil
	case []string:
		out := make([]string, len(v))
		for i, item := range v {
			if version == "2008-10-17" && strings.Contains(item, "&{") {
				return nil, fmt.Errorf("found &{ sequence in (%s), which is not supported in document version 2008-10-17", item)
			}
			out[i] = dataSourcePolicyDocumentVarReplacer.Replace(item)
		}
		return out, nil
	default:
		return nil, errors.New("dataSourcePolicyDocumentReplaceVarsInList: input not string nor []string")
	}
}

func dataSourcePolicyDocumentMakeConditions(in []interface{}, version string) (IAMPolicyStatementConditionSet, error) {
	out := make([]IAMPolicyStatementCondition, len(in))
	for i, itemI := range in {
		var err error
		item := itemI.(map[string]interface{})
		out[i] = IAMPolicyStatementCondition{
			Test:     item["test"].(string),
			Variable: item["variable"].(string),
		}
		out[i].Values, err = dataSourcePolicyDocumentReplaceVarsInList(
			aws.ToStringSlice(expandStringListKeepEmpty(item[names.AttrValues].([]interface{}))),
			version,
		)
		if err != nil {
			return nil, fmt.Errorf("reading values: %w", err)
		}
		itemValues := out[i].Values.([]string)
		if len(itemValues) == 1 {
			out[i].Values = itemValues[0]
		}
	}
	return IAMPolicyStatementConditionSet(out), nil
}

func dataSourcePolicyDocumentMakePrincipals(in []interface{}, version string) (IAMPolicyStatementPrincipalSet, error) {
	out := make([]IAMPolicyStatementPrincipal, len(in))
	for i, itemI := range in {
		var err error
		item := itemI.(map[string]interface{})
		out[i] = IAMPolicyStatementPrincipal{
			Type: item[names.AttrType].(string),
		}
		out[i].Identifiers, err = dataSourcePolicyDocumentReplaceVarsInList(
			policyDecodeConfigStringList(
				item["identifiers"].(*schema.Set).List(),
			), version,
		)
		if err != nil {
			return nil, fmt.Errorf("reading identifiers: %w", err)
		}
	}
	return IAMPolicyStatementPrincipalSet(out), nil
}
