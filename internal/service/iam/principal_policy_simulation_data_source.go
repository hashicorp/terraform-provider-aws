// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_iam_principal_policy_simulation", name="Principal Policy Simulation")
func dataSourcePrincipalPolicySimulation() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePrincipalPolicySimulationRead,

		Schema: map[string]*schema.Schema{
			// Arguments
			"action_names": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: `One or more names of actions, like "iam:CreateUser", that should be included in the simulation.`,
			},
			"caller_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
				Description:  `ARN of a user to use as the caller of the simulated requests. If not specified, defaults to the principal specified in policy_source_arn, if it is a user ARN.`,
			},
			"context": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKey: {
							Type:        schema.TypeString,
							Required:    true,
							Description: `The key name of the context entry, such as "aws:CurrentTime".`,
						},
						names.AttrType: {
							Type:        schema.TypeString,
							Required:    true,
							Description: `The type that the simulator should use to interpret the strings given in argument "values".`,
						},
						names.AttrValues: {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: `One or more values to assign to the context key, given as a string in a syntax appropriate for the selected value type.`,
						},
					},
				},
				Description: `Each block specifies one item of additional context entry to include in the simulated requests. These are the additional properties used in the 'Condition' element of an IAM policy, and in dynamic value interpolations.`,
			},
			"permissions_boundary_policies_json": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsJSON,
				},
				Description: `Additional permission boundary policies to use in the simulation.`,
			},
			"additional_policies_json": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsJSON,
				},
				Description: `Additional principal-based policies to use in the simulation.`,
			},
			"policy_source_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
				Description:  `ARN of the principal (e.g. user, role) whose existing configured access policies will be used as the basis for the simulation. If you specify a role ARN here, you can also set caller_arn to simulate a particular user acting with the given role.`,
			},
			"resource_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
				Description: `ARNs of specific resources to use as the targets of the specified actions during simulation. If not specified, the simulator assumes "*" which represents general access across all resources.`,
			},
			"resource_handling_option": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: `Specifies the type of simulation to run. Some API operations need a particular resource handling option in order to produce a correct reesult.`,
			},
			"resource_owner_account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidAccountID,
				Description:  `An AWS account ID to use as the simulated owner for any resource whose ARN does not include a specific owner account ID. Defaults to the account given as part of caller_arn.`,
			},
			"resource_policy_json": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsJSON,
				Description:  `A resource policy to associate with all of the target resources for simulation purposes. The policy simulator does not automatically retrieve resource-level policies, so if a resource policy is crucial to your test then you must specify here the same policy document associated with your target resource(s).`,
			},

			// Result Attributes
			"all_allowed": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: `A summary of the results attribute which is true if all of the results have decision "allowed", and false otherwise.`,
			},
			"results": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The name of the action whose simulation this result is describing.`,
						},
						"decision": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The exact decision keyword returned by the policy simulator: "allowed", "explicitDeny", or "implicitDeny".`,
						},
						"allowed": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: `A summary of attribute "decision" which is true only if the decision is "allowed".`,
						},
						"decision_details": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: `A mapping of various additional details that are relevant to the decision, exactly as returned by the policy simulator.`,
						},
						names.AttrResourceARN: {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `ARN of the resource that the action was tested against.`,
						},
						"matched_statements": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"source_policy_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: `Identifier of one of the policies used as input to the simulation.`,
									},
									"source_policy_type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: `The type of the policy identified in source_policy_id.`,
									},
									// NOTE: start position and end position
									// omitted right now because they would
									// ideally be singleton objects with
									// column/line attributes, but this SDK
									// can't support that. Maybe we later adopt
									// the new framework and add support for
									// those.
								},
							},
							Description: `Detail about which specific policies contributed to this result.`,
						},
						"missing_context_keys": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: `Set of context entry keys that were needed for one or more of the relevant policies but not included in the request. You must specify suitable values for all context keys used in all of the relevant policies in order to obtain a correct simulation result.`,
						},
						// NOTE: organizations decision detail, permissions
						// boundary decision detail, and resource-specific
						// results omitted for now because it isn't clear
						// that they will be useful and they would make the
						// results of this data source considerably larger
						// and more complicated.
					},
				},
			},
			names.AttrID: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: `Do not use`,
			},
		},
	}
}

func dataSourcePrincipalPolicySimulationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	setAsAWSStringSlice := func(raw interface{}) []string {
		if raw.(*schema.Set).Len() == 0 {
			return nil
		}
		return flex.ExpandStringValueSet(raw.(*schema.Set))
	}

	input := &iam.SimulatePrincipalPolicyInput{
		ActionNames:                        setAsAWSStringSlice(d.Get("action_names")),
		PermissionsBoundaryPolicyInputList: setAsAWSStringSlice(d.Get("permissions_boundary_policies_json")),
		PolicyInputList:                    setAsAWSStringSlice(d.Get("additional_policies_json")),
		PolicySourceArn:                    aws.String(d.Get("policy_source_arn").(string)),
		ResourceArns:                       setAsAWSStringSlice(d.Get("resource_arns")),
	}

	for _, entryRaw := range d.Get("context").(*schema.Set).List() {
		entryRaw := entryRaw.(map[string]interface{})
		entry := awstypes.ContextEntry{
			ContextKeyName:   aws.String(entryRaw[names.AttrKey].(string)),
			ContextKeyType:   awstypes.ContextKeyTypeEnum(entryRaw[names.AttrType].(string)),
			ContextKeyValues: setAsAWSStringSlice(entryRaw[names.AttrValues]),
		}
		input.ContextEntries = append(input.ContextEntries, entry)
	}

	if v := d.Get("caller_arn").(string); v != "" {
		input.CallerArn = aws.String(v)
	}
	if v := d.Get("resource_handling_option").(string); v != "" {
		input.ResourceHandlingOption = aws.String(v)
	}
	if v := d.Get("resource_owner_account_id").(string); v != "" {
		input.ResourceOwner = aws.String(v)
	}
	if v := d.Get("resource_policy_json").(string); v != "" {
		input.ResourcePolicy = aws.String(v)
	}

	// We are going to keep fetching through potentially multiple pages of
	// results in order to return a complete result, so we'll ask the API
	// to return as much as possible in each request to minimize the
	// round-trips.
	input.MaxItems = aws.Int32(1000)

	var results []awstypes.EvaluationResult

	for { // Terminates below, once we see a result that does not set IsTruncated.
		output, err := conn.SimulatePrincipalPolicy(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "simulating IAM Principal Policy: %s", err)
		}

		results = append(results, output.EvaluationResults...)

		if !output.IsTruncated {
			break // All done!
		}

		// If we're making another request then we need to specify the marker
		// to get the next page of results.
		input.Marker = output.Marker
	}

	// While we build the result we'll also tally up the number of allowed
	// vs. denied decisions to use for our top-level "all_allowed" summary
	// result.
	allowedCount := 0
	deniedCount := 0

	rawResults := make([]interface{}, len(results))
	for i, result := range results {
		rawResult := map[string]interface{}{}
		rawResult["action_name"] = aws.ToString(result.EvalActionName)
		rawResult["decision"] = string(result.EvalDecision)
		allowed := string(result.EvalDecision) == "allowed"
		rawResult["allowed"] = allowed
		if allowed {
			allowedCount++
		} else {
			deniedCount++
		}
		if result.EvalResourceName != nil {
			rawResult[names.AttrResourceARN] = aws.ToString(result.EvalResourceName)
		}

		var missingContextKeys []string
		for _, mkk := range result.MissingContextValues {
			if mkk != "" {
				missingContextKeys = append(missingContextKeys, mkk)
			}
		}
		rawResult["missing_context_keys"] = missingContextKeys

		decisionDetails := make(map[string]string, len(result.EvalDecisionDetails))
		for k, pv := range result.EvalDecisionDetails {
			if pv != "" {
				decisionDetails[k] = string(pv)
			}
		}
		rawResult["decision_details"] = decisionDetails

		rawMatchedStmts := make([]interface{}, len(result.MatchedStatements))
		for i, stmt := range result.MatchedStatements {
			rawStmt := map[string]interface{}{
				"source_policy_id":   stmt.SourcePolicyId,
				"source_policy_type": stmt.SourcePolicyType,
			}
			rawMatchedStmts[i] = rawStmt
		}
		rawResult["matched_statements"] = rawMatchedStmts

		rawResults[i] = rawResult
	}
	d.Set("results", rawResults)

	// "all" are allowed only if there is at least one result and no other
	// results were denied. We require at least one allowed here just as
	// a safety-net against a confusing result from a degenerate request.
	d.Set("all_allowed", allowedCount > 0 && deniedCount == 0)

	d.SetId("-")

	return diags
}
