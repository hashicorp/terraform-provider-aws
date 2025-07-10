// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	stackInstancesResourceIDPartCount = 3 // stack_set_name, call_as, OU
	regionAcctOrgIDSeparator          = "/"
	ResNameStackInstances             = "Stack Instances"
)

// @SDKResource("aws_cloudformation_stack_instances", name="Stack Instances")
func resourceStackInstances() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStackInstancesCreate,
		ReadWithoutTimeout:   resourceStackInstancesRead,
		UpdateWithoutTimeout: resourceStackInstancesUpdate,
		DeleteWithoutTimeout: resourceStackInstancesDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceStackInstancesImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			AttrAccounts: { // create input, read input (single account), update input, delete input
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"deployment_targets"},
				MinItems:      1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidAccountID,
				},
			},
			"call_as": { // create input, read input, update input, delete input
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.CallAsSelf,
				ValidateDiagFunc: enum.Validate[awstypes.CallAs](),
			},
			"deployment_targets": { // create input, update input, delete input
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account_filter_type": { // create input
							Type:          schema.TypeString,
							Optional:      true,
							ValidateFunc:  validation.StringInSlice(enum.Slice(awstypes.AccountFilterType.Values("")...), false),
							ForceNew:      true,
							ConflictsWith: []string{AttrAccounts},
						},
						AttrAccounts: { // create input
							Type:          schema.TypeSet,
							Optional:      true,
							ConflictsWith: []string{AttrAccounts},
							MinItems:      1,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: verify.ValidAccountID,
							},
						},
						"accounts_url": { // create input
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{AttrAccounts},
							ValidateFunc:  validation.StringMatch(regexache.MustCompile(`(s3://|http(s?)://).+`), ""),
						},
						"organizational_unit_ids": { // create input
							Type:          schema.TypeSet,
							Optional:      true,
							MinItems:      1,
							ConflictsWith: []string{AttrAccounts},
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringMatch(regexache.MustCompile(`^(ou-[0-9a-z]{4,32}-[0-9a-z]{8,32}|r-[0-9a-z]{4,32})$`), ""),
							},
						},
					},
				},
				ConflictsWith: []string{AttrAccounts},
			},
			"operation_preferences": { // create input, update input, delete input
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"concurrency_mode": { // create input
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ConcurrencyMode](),
						},
						"failure_tolerance_count": { // create input
							Type:          schema.TypeInt,
							Optional:      true,
							ValidateFunc:  validation.IntAtLeast(0),
							ConflictsWith: []string{"operation_preferences.0.failure_tolerance_percentage"},
						},
						"failure_tolerance_percentage": { // create input
							Type:          schema.TypeInt,
							Optional:      true,
							ValidateFunc:  validation.IntBetween(0, 100),
							ConflictsWith: []string{"operation_preferences.0.failure_tolerance_count"},
						},
						"max_concurrent_count": { // create input
							Type:          schema.TypeInt,
							Optional:      true,
							ValidateFunc:  validation.IntAtLeast(1),
							ConflictsWith: []string{"operation_preferences.0.max_concurrent_percentage"},
						},
						"max_concurrent_percentage": { // create input
							Type:          schema.TypeInt,
							Optional:      true,
							ValidateFunc:  validation.IntBetween(1, 100),
							ConflictsWith: []string{"operation_preferences.0.max_concurrent_count"},
						},
						"region_concurrency_type": { // create input
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.RegionConcurrencyType](),
						},
						"region_order": { // create input
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z-]{1,128}$`), ""),
							},
						},
					},
				},
			},
			"parameter_overrides": { // create input, update input
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			AttrRegions: { // create input - required, read input (single region), update input - required, delete input - required
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidRegionName,
				},
			},
			"retain_stacks": { // delete input - required
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"stack_instance_summaries": { // read output
				Type:     schema.TypeList,
				Computed: true,
				Description: "List of stack instances created from an organizational unit deployment target. " +
					"This will only be populated when `deployment_targets` is set.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAccountID: { // read output
							Type:     schema.TypeString,
							Computed: true,
						},
						"detailed_status": { // read output
							Type:     schema.TypeString,
							Computed: true,
						},
						"drift_status": { // read output
							Type:     schema.TypeString,
							Computed: true,
						},
						"organizational_unit_id": { // read output
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrRegion: { // read output
							Type:     schema.TypeString,
							Computed: true,
						},
						"stack_id": { // read output
							Type:     schema.TypeString,
							Computed: true,
						},
						"stack_set_id": { // read output
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatus: { // read output
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatusReason: { // read output
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"stack_set_id": { // read output
				Type:     schema.TypeString,
				Computed: true,
			},
			"stack_set_name": { // create input - required, read input - required, update input - required, delete input - required
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func resourceStackInstancesCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationClient(ctx)

	stackSetName := d.Get("stack_set_name").(string)
	input := &cloudformation.CreateStackInstancesInput{
		StackSetName: aws.String(stackSetName),
	}

	if v, ok := d.GetOk(AttrRegions); ok && v.(*schema.Set).Len() > 0 {
		input.Regions = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk(AttrRegions); !ok || v.(*schema.Set).Len() == 0 {
		input.Regions = []string{meta.(*conns.AWSClient).Region(ctx)}
	}

	if v, ok := d.GetOk(AttrAccounts); ok && v.(*schema.Set).Len() > 0 {
		input.Accounts = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	deployedByOU := ""
	if v, ok := d.GetOk("deployment_targets"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.DeploymentTargets = expandDeploymentTargets(v.([]any))
		input.Accounts = nil

		if v, ok := d.GetOk("deployment_targets.0.organizational_unit_ids"); ok && len(v.(*schema.Set).List()) > 0 {
			deployedByOU = "OU"
		}
	} else {
		input.Accounts = []string{meta.(*conns.AWSClient).AccountID(ctx)}
	}

	callAs := d.Get("call_as").(string)
	if v, ok := d.GetOk("call_as"); ok {
		input.CallAs = awstypes.CallAs(v.(string))
	}

	if v, ok := d.GetOk("parameter_overrides"); ok {
		input.ParameterOverrides = expandParameters(v.(map[string]any))
	}

	if v, ok := d.GetOk("operation_preferences"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.OperationPreferences = expandOperationPreferences(v.([]any)[0].(map[string]any))
	}

	id, err := flex.FlattenResourceId([]string{stackSetName, callAs, deployedByOU}, stackInstancesResourceIDPartCount, true)
	if err != nil {
		return create.AppendDiagError(diags, names.CloudFormation, create.ErrActionFlatteningResourceId, ResNameStackInstances, stackSetName, err)
	}

	_, err = tfresource.RetryWhen(ctx, propagationTimeout,
		func() (any, error) {
			input.OperationId = aws.String(sdkid.UniqueId())

			output, err := conn.CreateStackInstances(ctx, input)
			if err != nil {
				return nil, err
			}

			d.SetId(id)

			operation, err := waitStackSetOperationSucceeded(ctx, conn, stackSetName, aws.ToString(output.OperationId), callAs, d.Timeout(schema.TimeoutCreate))
			if err != nil {
				return nil, fmt.Errorf("waiting for create: %w", err)
			}

			return operation, nil
		},
		isRetryableIAMPropagationErr,
	)

	if err != nil {
		return create.AppendDiagError(diags, names.CloudFormation, create.ErrActionCreating, ResNameStackInstances, id, err)
	}

	return append(diags, resourceStackInstancesRead(ctx, d, meta)...)
}

func resourceStackInstancesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Dramatic simplification of the ID from stack_set_instance removing regions and accounts. The upside
	// is simplicity. The downside is that we hoover up all stack instances for the stack set.
	parts, err := flex.ExpandResourceId(d.Id(), stackInstancesResourceIDPartCount, true)
	if err != nil {
		return create.AppendDiagError(diags, names.CloudFormation, create.ErrActionExpandingResourceId, ResNameStackInstances, d.Id(), err)
	}

	stackSetName, callAs, deployedByOU := parts[0], parts[1], parts[2]
	d.Set("stack_set_name", stackSetName)

	if callAs == "" {
		callAs = d.Get("call_as").(string)
	}

	stackInstances, err := findStackInstancesByNameCallAs(ctx, meta, stackSetName, callAs, deployedByOU == "OU", flex.ExpandStringValueSet(d.Get(AttrAccounts).(*schema.Set)), flex.ExpandStringValueSet(d.Get(AttrRegions).(*schema.Set)))
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudFormation Stack Instances (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.CloudFormation, create.ErrActionReading, ResNameStackInstances, d.Id(), err)
	}

	if len(stackInstances.OUs) > 0 {
		d.Set("deployment_targets", replaceOrganizationalUnitIDs(d.Get("deployment_targets").([]any), stackInstances.OUs))
	}

	d.Set(AttrAccounts, flex.FlattenStringValueList(stackInstances.Accounts))
	d.Set(AttrRegions, flex.FlattenStringValueList(stackInstances.Regions))
	d.Set("stack_instance_summaries", flattenStackInstancesSummaries(stackInstances.Summaries))
	d.Set("stack_set_id", stackInstances.StackSetID)

	if err := d.Set("parameter_overrides", flattenAllParameters(stackInstances.ParameterOverrides)); err != nil {
		return create.AppendDiagError(diags, names.CloudFormation, create.ErrActionReading, ResNameStackInstances, d.Id(), err)
	}

	return diags
}

const (
	AttrAccounts   = "accounts"
	AttrDTAccounts = "deployment_targets.0.accounts"
	AttrDTOUs      = "deployment_targets.0.organizational_unit_ids"
	AttrRegions    = "regions"
)

func resourceStackInstancesUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationClient(ctx)

	accounts := flex.ExpandStringValueSet(d.Get(AttrAccounts).(*schema.Set))
	regions := flex.ExpandStringValueSet(d.Get(AttrRegions).(*schema.Set))
	dtAccounts := flex.ExpandStringValueSet(d.Get(AttrDTAccounts).(*schema.Set))
	dtOUs := flex.ExpandStringValueSet(d.Get(AttrDTOUs).(*schema.Set))

	if d.HasChange(AttrRegions) {
		oRaw, nRaw := d.GetChange(AttrRegions)
		o, n := oRaw.(*schema.Set), nRaw.(*schema.Set)

		if axe := o.Difference(n); axe.Len() > 0 {
			if err := deleteStackInstances(ctx, d, meta, accounts, flex.ExpandStringValueSet(axe), dtAccounts, dtOUs); err != nil {
				return create.AppendDiagError(diags, names.CloudFormation, create.ErrActionDeleting, ResNameStackInstances, d.Id(), err)
			}
		}

		if add := n.Difference(o); add.Len() > 0 {
			diags = append(diags, resourceStackInstancesCreate(ctx, d, meta)...)
			if diags.HasError() {
				return diags
			}
		}
	}

	if d.HasChange(AttrAccounts) {
		oRaw, nRaw := d.GetChange(AttrAccounts)
		o, n := oRaw.(*schema.Set), nRaw.(*schema.Set)

		if axe := o.Difference(n); axe.Len() > 0 {
			if err := deleteStackInstances(ctx, d, meta, flex.ExpandStringValueSet(axe), regions, dtAccounts, dtOUs); err != nil {
				return create.AppendDiagError(diags, names.CloudFormation, create.ErrActionDeleting, ResNameStackInstances, d.Id(), err)
			}
		}

		if add := n.Difference(o); add.Len() > 0 {
			diags = append(diags, resourceStackInstancesCreate(ctx, d, meta)...)
			if diags.HasError() {
				return diags
			}
		}
	}

	if d.HasChange(AttrDTAccounts) {
		oRaw, nRaw := d.GetChange(AttrDTAccounts)
		o, n := oRaw.(*schema.Set), nRaw.(*schema.Set)

		if axe := o.Difference(n); axe.Len() > 0 {
			if err := deleteStackInstances(ctx, d, meta, accounts, regions, flex.ExpandStringValueSet(axe), dtOUs); err != nil {
				return create.AppendDiagError(diags, names.CloudFormation, create.ErrActionDeleting, ResNameStackInstances, d.Id(), err)
			}
		}

		if add := n.Difference(o); add.Len() > 0 {
			diags = append(diags, resourceStackInstancesCreate(ctx, d, meta)...)
			if diags.HasError() {
				return diags
			}
		}
	}

	if d.HasChange(AttrDTOUs) {
		oRaw, nRaw := d.GetChange(AttrDTOUs)
		o, n := oRaw.(*schema.Set), nRaw.(*schema.Set)

		if axe := o.Difference(n); axe.Len() > 0 {
			if err := deleteStackInstances(ctx, d, meta, accounts, regions, dtAccounts, flex.ExpandStringValueSet(axe)); err != nil {
				return create.AppendDiagError(diags, names.CloudFormation, create.ErrActionDeleting, ResNameStackInstances, d.Id(), err)
			}
		}

		if add := n.Difference(o); add.Len() > 0 {
			diags = append(diags, resourceStackInstancesCreate(ctx, d, meta)...)
			if diags.HasError() {
				return diags
			}
		}
	}

	if d.HasChanges(
		"call_as",
		"deployment_targets.0.accounts_url",
		"operation_preferences",
		"parameter_overrides",
	) {
		input := &cloudformation.UpdateStackInstancesInput{
			OperationId:        aws.String(sdkid.UniqueId()),
			ParameterOverrides: []awstypes.Parameter{},
			Regions:            flex.ExpandStringValueSet(d.Get(AttrRegions).(*schema.Set)),
			StackSetName:       aws.String(d.Get("stack_set_name").(string)),
		}

		// can only give either accounts or deployment_targets
		input.Accounts = []string{meta.(*conns.AWSClient).AccountID(ctx)}
		if v, ok := d.GetOk(AttrAccounts); ok && v.(*schema.Set).Len() > 0 {
			input.Accounts = flex.ExpandStringValueSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("deployment_targets"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.DeploymentTargets = expandDeploymentTargets(v.([]any))
			input.Accounts = nil
		}

		if v, ok := d.GetOk("call_as"); ok {
			input.CallAs = awstypes.CallAs(v.(string))
		}

		if v, ok := d.GetOk("operation_preferences"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.OperationPreferences = expandOperationPreferences(v.([]any)[0].(map[string]any))
		}

		if v, ok := d.GetOk("parameter_overrides"); ok {
			input.ParameterOverrides = expandParameters(v.(map[string]any))
		}

		output, err := conn.UpdateStackInstances(ctx, input)
		if err != nil {
			return create.AppendDiagError(diags, names.CloudFormation, create.ErrActionUpdating, ResNameStackInstances, d.Id(), err)
		}

		if _, err := waitStackSetOperationSucceeded(ctx, conn, d.Get("stack_set_name").(string), aws.ToString(output.OperationId), d.Get("call_as").(string), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.AppendDiagError(diags, names.CloudFormation, create.ErrActionWaitingForUpdate, ResNameStackInstances, d.Id(), err)
		}
	}

	return append(diags, resourceStackInstancesRead(ctx, d, meta)...)
}

func resourceStackInstancesDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	accounts := flex.ExpandStringValueSet(d.Get(AttrAccounts).(*schema.Set))
	regions := flex.ExpandStringValueSet(d.Get(AttrRegions).(*schema.Set))
	dtAccounts := flex.ExpandStringValueSet(d.Get(AttrDTAccounts).(*schema.Set))
	dtOUs := flex.ExpandStringValueSet(d.Get(AttrDTOUs).(*schema.Set))

	if err := deleteStackInstances(ctx, d, meta, accounts, regions, dtAccounts, dtOUs); err != nil {
		return create.AppendDiagError(diags, names.CloudFormation, create.ErrActionDeleting, ResNameStackInstances, d.Id(), err)
	}

	return diag.Diagnostics{}
}

func deleteStackInstances(ctx context.Context, d *schema.ResourceData, meta any, accounts, regions, dtAccounts, dtOUs []string) error {
	conn := meta.(*conns.AWSClient).CloudFormationClient(ctx)

	input := &cloudformation.DeleteStackInstancesInput{
		OperationId:  aws.String(sdkid.UniqueId()),
		Accounts:     accounts,
		Regions:      regions,
		RetainStacks: aws.Bool(d.Get("retain_stacks").(bool)),
		StackSetName: aws.String(d.Get("stack_set_name").(string)),
	}

	// can only give either accounts or deployment_targets
	if v, ok := d.GetOk("deployment_targets"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.DeploymentTargets = expandDeploymentTargets(v.([]any))
		input.DeploymentTargets.Accounts = dtAccounts
		input.DeploymentTargets.OrganizationalUnitIds = dtOUs
		input.Accounts = nil
	}

	if v, ok := d.GetOk("call_as"); ok {
		input.CallAs = awstypes.CallAs(v.(string))
	}

	if v, ok := d.GetOk("operation_preferences"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.OperationPreferences = expandOperationPreferences(v.([]any)[0].(map[string]any))
	}

	log.Printf("[DEBUG] Deleting CloudFormation Stack Instances: %s", d.Id())
	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.OperationInProgressException](ctx, d.Timeout(schema.TimeoutDelete), func() (any, error) {
		return conn.DeleteStackInstances(ctx, input)
	})

	if errs.IsA[*awstypes.StackInstanceNotFoundException](err) || errs.IsA[*awstypes.StackSetNotFoundException](err) {
		return nil
	}

	if err != nil {
		return err
	}

	if _, err := waitStackSetOperationSucceeded(ctx, conn, d.Get("stack_set_name").(string), aws.ToString(outputRaw.(*cloudformation.DeleteStackInstancesOutput).OperationId), d.Get("call_as").(string), d.Timeout(schema.TimeoutDelete)); err != nil {
		return err
	}

	return nil
}

func resourceStackInstancesImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	switch parts := strings.Split(d.Id(), flex.ResourceIdSeparator); len(parts) {
	case 1:
	case 2:
	case 3:
		d.SetId(strings.Join([]string{parts[0], parts[1], parts[2]}, flex.ResourceIdSeparator))
		d.Set("call_as", parts[1])
	default:
		return []*schema.ResourceData{}, fmt.Errorf("unexpected format for import ID (%[1]s), use: STACKSET_NAME or STACKSET_NAME%[2]sCALLAS or STACKSET_NAME%[2]sCALLAS%[2]sOU", d.Id(), flex.ResourceIdSeparator)
	}

	return []*schema.ResourceData{d}, nil
}

// StackInstances is a helper struct to hold the describeable/listable info for stack instances.
type StackInstances struct {
	Accounts           []string
	OUs                []string
	Regions            []string
	StackSetID         string
	Summaries          []awstypes.StackInstanceSummary
	ParameterOverrides []awstypes.Parameter
}

// findStackInstancesByNameCallAs is a helper function to find stack instances by stackSetName and callAs.
// accounts and regions are not used unless no summaries are returned. When no summaries are returned,
// drift detection is limited because we're just using the accounts and regions from config--we have no
// other choice. When there are no summaries, we use the first account and region to find a stack instance
// to get the parameter_overrides and stack_set_id.
func findStackInstancesByNameCallAs(ctx context.Context, meta any, stackSetName, callAs string, deployedByOU bool, accounts, regions []string) (StackInstances, error) {
	conn := meta.(*conns.AWSClient).CloudFormationClient(ctx)

	input := &cloudformation.ListStackInstancesInput{
		StackSetName: aws.String(stackSetName),
	}

	if callAs != "" {
		input.CallAs = awstypes.CallAs(callAs)
	}

	var output StackInstances
	none := true
	pages := cloudformation.NewListStackInstancesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.StackInstanceNotFoundException](err) || errs.IsA[*awstypes.StackSetNotFoundException](err) {
			return output, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return output, err
		}

		none = false

		for _, v := range page.Summaries {
			if aws.ToString(v.StackSetId) != "" && output.StackSetID == "" {
				output.StackSetID = aws.ToString(v.StackSetId)
			}

			output.Summaries = append(output.Summaries, v)

			if aws.ToString(v.Account) != "" {
				output.Accounts = append(output.Accounts, aws.ToString(v.Account))
			}

			if aws.ToString(v.Region) != "" {
				output.Regions = append(output.Regions, aws.ToString(v.Region))
			}

			if aws.ToString(v.OrganizationalUnitId) != "" {
				output.OUs = append(output.OUs, aws.ToString(v.OrganizationalUnitId))
			}
		}
	}

	if len(output.Accounts) == 0 && len(accounts) > 0 {
		output.Accounts = accounts
	}

	if len(output.Accounts) == 0 && len(accounts) == 0 {
		output.Accounts = []string{meta.(*conns.AWSClient).AccountID(ctx)}
	}

	if len(output.Regions) == 0 && len(regions) > 0 {
		output.Regions = regions
	}

	if len(output.Regions) == 0 && len(regions) == 0 {
		output.Regions = []string{meta.(*conns.AWSClient).Region(ctx)}
	}

	if deployedByOU {
		return output, nil
	}

	// set based on the first account and region which means they may not be accurate for all stack instances
	stackInstance, err := findStackInstanceByFourPartKey(ctx, conn, stackSetName, output.Accounts[0], output.Regions[0], callAs)
	if none || tfresource.NotFound(err) {
		return output, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil && !tfresource.NotFound(err) {
		return output, err
	}

	if stackInstance != nil && output.StackSetID == "" {
		output.StackSetID = aws.ToString(stackInstance.StackSetId)
	}

	if stackInstance != nil && stackInstance.ParameterOverrides != nil {
		output.ParameterOverrides = stackInstance.ParameterOverrides
	}

	return output, nil
}

func replaceOrganizationalUnitIDs(tfList []any, newOUIDs []string) []any {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	// Update the "organizational_unit_ids" with the new value
	tfMap["organizational_unit_ids"] = flex.FlattenStringValueList(newOUIDs)

	return []any{tfMap}
}

func flattenStackInstancesSummaries(apiObject []awstypes.StackInstanceSummary) []any {
	if len(apiObject) == 0 {
		return nil
	}

	tfList := []any{}
	for _, obj := range apiObject {
		m := map[string]any{
			names.AttrAccountID:      obj.Account,
			"drift_status":           obj.DriftStatus,
			"organizational_unit_id": obj.OrganizationalUnitId,
			names.AttrRegion:         obj.Region,
			"stack_id":               obj.StackId,
			"stack_set_id":           obj.StackSetId,
			names.AttrStatus:         obj.Status,
			names.AttrStatusReason:   obj.StatusReason,
		}

		if obj.StackInstanceStatus != nil {
			m["detailed_status"] = obj.StackInstanceStatus.DetailedStatus
		}

		tfList = append(tfList, m)
	}

	return tfList
}
