// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	policyDocumentsDefaultEffect        = "Allow"
	policyDocumentsDefaultMaxPolicySize = 6144 // Managed policy limit, excluding whitespace.
	policyDocumentsDefaultVersion       = "2012-10-17"
)

// @FrameworkDataSource("aws_iam_policy_documents", name="Policy Documents")
func newPolicyDocumentsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &policyDocumentsDataSource{}, nil
}

type policyDocumentsDataSource struct {
	framework.DataSourceWithModel[policyDocumentsDataSourceModel]
}

func (d *policyDocumentsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	policyDocumentsAttribute := func() schema.ListAttribute {
		return schema.ListAttribute{
			CustomType:  fwtypes.ListOfStringType,
			ElementType: types.StringType,
			Optional:    true,
			Validators: []validator.List{
				// Empty strings are skipped when merging, matching aws_iam_policy_document.
				listvalidator.ValueStringsAre(stringvalidator.Any(
					fwvalidators.JSON(),
					stringvalidator.OneOf(""),
				)),
			},
		}
	}
	setOfStringAttribute := func() schema.SetAttribute {
		return schema.SetAttribute{
			CustomType:  fwtypes.SetOfStringType,
			ElementType: types.StringType,
			Optional:    true,
		}
	}
	principalsBlock := func() schema.SetNestedBlock {
		return schema.SetNestedBlock{
			CustomType: fwtypes.NewSetNestedObjectTypeOf[policyDocumentsPrincipalModel](ctx),
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"identifiers": schema.SetAttribute{
						CustomType:  fwtypes.SetOfStringType,
						ElementType: types.StringType,
						Required:    true,
					},
					names.AttrType: schema.StringAttribute{
						Required: true,
					},
				},
			},
		}
	}

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrJSON: schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			"max_policy_size": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"minified_json": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			"override_policy_documents": policyDocumentsAttribute(),
			"policy_id": schema.StringAttribute{
				Optional: true,
			},
			"source_policy_documents": policyDocumentsAttribute(),
			names.AttrVersion: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("2008-10-17", "2012-10-17"),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"statement": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[policyDocumentsStatementModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrActions: setOfStringAttribute(),
						"effect": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Validators: []validator.String{
								stringvalidator.OneOf("Allow", "Deny"),
							},
						},
						"not_actions":       setOfStringAttribute(),
						"not_resources":     setOfStringAttribute(),
						names.AttrResources: setOfStringAttribute(),
						// Because policy documents are widely used outside IAM, we don't enforce
						// IAM validation rules requiring alphanumeric and no spaces.
						"sid": schema.StringAttribute{
							Optional: true,
						},
					},
					Blocks: map[string]schema.Block{
						names.AttrCondition: schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[policyDocumentsConditionModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"test": schema.StringAttribute{
										Required: true,
									},
									names.AttrValues: schema.ListAttribute{
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Required:    true,
									},
									"variable": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"not_principals": principalsBlock(),
						"principals":     principalsBlock(),
					},
				},
			},
		},
	}
}

func (d *policyDocumentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data policyDocumentsDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Version.IsNull() {
		data.Version = types.StringValue(policyDocumentsDefaultVersion)
	}
	if data.MaxPolicySize.IsNull() {
		data.MaxPolicySize = types.Int64Value(policyDocumentsDefaultMaxPolicySize)
	}

	doc := &iamPolicyDoc{
		Id:      data.PolicyID.ValueString(),
		Version: data.Version.ValueString(),
	}

	statements, diags := data.Statements.ToSlice(ctx)
	smerr.AddEnrich(ctx, &resp.Diagnostics, diags)
	if resp.Diagnostics.HasError() {
		return
	}

	sids := make(map[string]struct{})
	for _, s := range statements {
		if s.Effect.IsNull() {
			s.Effect = types.StringValue(policyDocumentsDefaultEffect)
		}

		if sid := s.Sid.ValueString(); sid != "" {
			if _, ok := sids[sid]; ok {
				smerr.AddError(ctx, &resp.Diagnostics, fmt.Errorf("duplicate Sid (%s). Remove the Sid or ensure the Sid is unique.", sid))
				return
			}
			sids[sid] = struct{}{}
		}

		stmt, err := expandPolicyDocumentsStatement(ctx, s, doc.Version)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err)
			return
		}
		doc.Statements = append(doc.Statements, stmt)
	}

	if len(statements) > 0 {
		statementsValue, diags := fwtypes.NewListNestedObjectValueOfSlice(ctx, statements, nil)
		smerr.AddEnrich(ctx, &resp.Diagnostics, diags)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Statements = statementsValue
	}

	mergedDoc, err := mergePolicyDocuments(
		fwflex.ExpandFrameworkStringValueList(ctx, data.SourcePolicyDocuments),
		doc,
		fwflex.ExpandFrameworkStringValueList(ctx, data.OverridePolicyDocuments),
	)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	docs, err := splitPolicyDocument(mergedDoc, int(data.MaxPolicySize.ValueInt64()))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	jsons := make([]string, len(docs))
	minifiedJSONs := make([]string, len(docs))
	for i, chunk := range docs {
		b, err := json.Marshal(chunk)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err)
			return
		}
		minifiedJSONs[i] = string(b)

		// Identical to json.MarshalIndent, without marshalling again.
		var buf bytes.Buffer
		if err := json.Indent(&buf, b, "", "  "); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err)
			return
		}
		jsons[i] = buf.String()
	}

	data.JSON = fwflex.FlattenFrameworkStringValueListOfString(ctx, jsons)
	data.MinifiedJSON = fwflex.FlattenFrameworkStringValueListOfString(ctx, minifiedJSONs)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

// AutoFlex cannot target iamPolicyStatement: its `any` fields collapse single values
// to strings and need `&{` replacement, matching aws_iam_policy_document.
func expandPolicyDocumentsStatement(ctx context.Context, s *policyDocumentsStatementModel, version string) (*iamPolicyStatement, error) { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	stmt := &iamPolicyStatement{
		Effect: s.Effect.ValueString(),
		Sid:    s.Sid.ValueString(),
	}

	if v := fwflex.ExpandFrameworkStringValueSet(ctx, s.Actions); len(v) > 0 {
		stmt.Actions = policyDecodeConfigStrings(v)
	}
	if v := fwflex.ExpandFrameworkStringValueSet(ctx, s.NotActions); len(v) > 0 {
		stmt.NotActions = policyDecodeConfigStrings(v)
	}

	var err error
	if v := fwflex.ExpandFrameworkStringValueSet(ctx, s.Resources); len(v) > 0 {
		if stmt.Resources, err = dataSourcePolicyDocumentReplaceVarsInList(policyDecodeConfigStrings(v), version); err != nil {
			return nil, fmt.Errorf("reading resources: %w", err)
		}
	}
	if v := fwflex.ExpandFrameworkStringValueSet(ctx, s.NotResources); len(v) > 0 {
		if stmt.NotResources, err = dataSourcePolicyDocumentReplaceVarsInList(policyDecodeConfigStrings(v), version); err != nil {
			return nil, fmt.Errorf("reading not_resources: %w", err)
		}
	}

	if stmt.Principals, err = expandPolicyDocumentsPrincipals(ctx, s.Principals, version); err != nil {
		return nil, fmt.Errorf("reading principals: %w", err)
	}
	if stmt.NotPrincipals, err = expandPolicyDocumentsPrincipals(ctx, s.NotPrincipals, version); err != nil {
		return nil, fmt.Errorf("reading not_principals: %w", err)
	}

	if stmt.Conditions, err = expandPolicyDocumentsConditions(ctx, s.Conditions, version); err != nil {
		return nil, fmt.Errorf("reading condition: %w", err)
	}

	return stmt, nil
}

func expandPolicyDocumentsPrincipals(ctx context.Context, v fwtypes.SetNestedObjectValueOf[policyDocumentsPrincipalModel], version string) (iamPolicyStatementPrincipalSet, error) { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	principals, diags := v.ToSlice(ctx)
	if diags.HasError() {
		return nil, fwdiag.DiagnosticsError(diags)
	}
	if len(principals) == 0 {
		return nil, nil
	}

	out := make(iamPolicyStatementPrincipalSet, len(principals))
	for i, p := range principals {
		identifiers, err := dataSourcePolicyDocumentReplaceVarsInList(
			policyDecodeConfigStrings(fwflex.ExpandFrameworkStringValueSet(ctx, p.Identifiers)), version,
		)
		if err != nil {
			return nil, fmt.Errorf("reading identifiers: %w", err)
		}
		out[i] = iamPolicyStatementPrincipal{
			Type:        p.Type.ValueString(),
			Identifiers: identifiers,
		}
	}
	return out, nil
}

func expandPolicyDocumentsConditions(ctx context.Context, v fwtypes.SetNestedObjectValueOf[policyDocumentsConditionModel], version string) (iamPolicyStatementConditionSet, error) { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	conditions, diags := v.ToSlice(ctx)
	if diags.HasError() {
		return nil, fwdiag.DiagnosticsError(diags)
	}
	if len(conditions) == 0 {
		return nil, nil
	}

	out := make(iamPolicyStatementConditionSet, len(conditions))
	for i, c := range conditions {
		values, err := policyConditionValues(fwflex.ExpandFrameworkStringValueList(ctx, c.Values), version)
		if err != nil {
			return nil, fmt.Errorf("reading values: %w", err)
		}
		out[i] = iamPolicyStatementCondition{
			Test:     c.Test.ValueString(),
			Variable: c.Variable.ValueString(),
			Values:   values,
		}
	}
	return out, nil
}

type policyDocumentsDataSourceModel struct {
	JSON                    fwtypes.ListOfString                                           `tfsdk:"json"`
	MaxPolicySize           types.Int64                                                    `tfsdk:"max_policy_size"`
	MinifiedJSON            fwtypes.ListOfString                                           `tfsdk:"minified_json"`
	OverridePolicyDocuments fwtypes.ListOfString                                           `tfsdk:"override_policy_documents"`
	PolicyID                types.String                                                   `tfsdk:"policy_id"`
	SourcePolicyDocuments   fwtypes.ListOfString                                           `tfsdk:"source_policy_documents"`
	Statements              fwtypes.ListNestedObjectValueOf[policyDocumentsStatementModel] `tfsdk:"statement"`
	Version                 types.String                                                   `tfsdk:"version"`
}

type policyDocumentsStatementModel struct {
	Actions       fwtypes.SetOfString                                           `tfsdk:"actions"`
	Conditions    fwtypes.SetNestedObjectValueOf[policyDocumentsConditionModel] `tfsdk:"condition"`
	Effect        types.String                                                  `tfsdk:"effect"`
	NotActions    fwtypes.SetOfString                                           `tfsdk:"not_actions"`
	NotPrincipals fwtypes.SetNestedObjectValueOf[policyDocumentsPrincipalModel] `tfsdk:"not_principals"`
	NotResources  fwtypes.SetOfString                                           `tfsdk:"not_resources"`
	Principals    fwtypes.SetNestedObjectValueOf[policyDocumentsPrincipalModel] `tfsdk:"principals"`
	Resources     fwtypes.SetOfString                                           `tfsdk:"resources"`
	Sid           types.String                                                  `tfsdk:"sid"`
}

type policyDocumentsPrincipalModel struct {
	Identifiers fwtypes.SetOfString `tfsdk:"identifiers"`
	Type        types.String        `tfsdk:"type"`
}

type policyDocumentsConditionModel struct {
	Test     types.String         `tfsdk:"test"`
	Values   fwtypes.ListOfString `tfsdk:"values"`
	Variable types.String         `tfsdk:"variable"`
}
