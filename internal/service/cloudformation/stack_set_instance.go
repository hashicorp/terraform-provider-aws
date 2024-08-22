// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation

import (
	"context"
	"fmt"
	"log"
	"slices"
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
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	stackSetInstanceResourceIDPartCount = 3
)

// @SDKResource("aws_cloudformation_stack_set_instance", name="Stack Set Instance")
func resourceStackSetInstance() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStackSetInstanceCreate,
		ReadWithoutTimeout:   resourceStackSetInstanceRead,
		UpdateWithoutTimeout: resourceStackSetInstanceUpdate,
		DeleteWithoutTimeout: resourceStackSetInstanceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceStackSetInstanceImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrAccountID: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  verify.ValidAccountID,
				ConflictsWith: []string{"deployment_targets"},
			},
			"call_as": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.CallAsSelf,
				ValidateDiagFunc: enum.Validate[awstypes.CallAs](),
			},
			"deployment_targets": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"organizational_unit_ids": {
							Type:          schema.TypeSet,
							Optional:      true,
							ForceNew:      true,
							MinItems:      1,
							ConflictsWith: []string{names.AttrAccountID},
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringMatch(regexache.MustCompile(`^(ou-[0-9a-z]{4,32}-[0-9a-z]{8,32}|r-[0-9a-z]{4,32})$`), ""),
							},
						},
						"account_filter_type": {
							Type:          schema.TypeString,
							Optional:      true,
							ForceNew:      true,
							ValidateFunc:  validation.StringInSlice(enum.Slice(awstypes.AccountFilterType.Values("")...), false),
							ConflictsWith: []string{names.AttrAccountID},
						},
						"accounts": {
							Type:          schema.TypeSet,
							Optional:      true,
							ForceNew:      true,
							ConflictsWith: []string{names.AttrAccountID},
							MinItems:      1,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: verify.ValidAccountID,
							},
						},
						"accounts_url": {
							Type:          schema.TypeString,
							ForceNew:      true,
							Optional:      true,
							ConflictsWith: []string{names.AttrAccountID},
							ValidateFunc:  validation.StringMatch(regexache.MustCompile(`(s3://|http(s?)://).+`), ""),
						},
					},
				},
				ConflictsWith: []string{names.AttrAccountID},
			},
			"operation_preferences": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"failure_tolerance_count": {
							Type:          schema.TypeInt,
							Optional:      true,
							ValidateFunc:  validation.IntAtLeast(0),
							ConflictsWith: []string{"operation_preferences.0.failure_tolerance_percentage"},
						},
						"failure_tolerance_percentage": {
							Type:          schema.TypeInt,
							Optional:      true,
							ValidateFunc:  validation.IntBetween(0, 100),
							ConflictsWith: []string{"operation_preferences.0.failure_tolerance_count"},
						},
						"max_concurrent_count": {
							Type:          schema.TypeInt,
							Optional:      true,
							ValidateFunc:  validation.IntAtLeast(1),
							ConflictsWith: []string{"operation_preferences.0.max_concurrent_percentage"},
						},
						"max_concurrent_percentage": {
							Type:          schema.TypeInt,
							Optional:      true,
							ValidateFunc:  validation.IntBetween(1, 100),
							ConflictsWith: []string{"operation_preferences.0.max_concurrent_count"},
						},
						"concurrency_mode": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ConcurrencyMode](),
						},
						"region_concurrency_type": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.RegionConcurrencyType](),
						},
						"region_order": {
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
			"organizational_unit_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parameter_overrides": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrRegion: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"retain_stack": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"stack_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"stack_instance_summaries": {
				Type:     schema.TypeList,
				Computed: true,
				Description: "List of stack instances created from an organizational unit deployment target. " +
					"This will only be populated when `deployment_targets` is set.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAccountID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"organizational_unit_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"stack_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"stack_set_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func resourceStackSetInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationClient(ctx)

	region := meta.(*conns.AWSClient).Region
	if v, ok := d.GetOk(names.AttrRegion); ok {
		region = v.(string)
	}

	stackSetName := d.Get("stack_set_name").(string)
	input := &cloudformation.CreateStackInstancesInput{
		Regions:      []string{region},
		StackSetName: aws.String(stackSetName),
	}

	accountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk(names.AttrAccountID); ok {
		accountID = v.(string)
	}
	// accountOrOrgID will either be account_id or a slash-delimited list of
	// organizational_unit_id's from the deployment_targets argument. This
	// is composed with stack_set_name and region to form the resources ID.
	accountOrOrgID := accountID

	if v, ok := d.GetOk("deployment_targets"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		dt := expandDeploymentTargets(v.([]interface{}))
		accountOrOrgID = strings.Join(dt.OrganizationalUnitIds, "/")
		input.DeploymentTargets = dt
	} else {
		d.Set(names.AttrAccountID, accountID)
		input.Accounts = []string{accountID}
	}

	callAs := d.Get("call_as").(string)
	if v, ok := d.GetOk("call_as"); ok {
		input.CallAs = awstypes.CallAs(v.(string))
	}

	if v, ok := d.GetOk("parameter_overrides"); ok {
		input.ParameterOverrides = expandParameters(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("operation_preferences"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.OperationPreferences = expandOperationPreferences(v.([]interface{})[0].(map[string]interface{}))
	}

	id := errs.Must(flex.FlattenResourceId([]string{stackSetName, accountOrOrgID, region}, stackSetInstanceResourceIDPartCount, false))
	_, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (interface{}, error) {
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
		func(err error) (bool, error) {
			if err == nil {
				return false, nil
			}

			message := err.Error()

			// IAM eventual consistency
			if strings.Contains(message, "AccountGate check failed") {
				return true, err
			}

			// IAM eventual consistency
			// User: XXX is not authorized to perform: cloudformation:CreateStack on resource: YYY
			if strings.Contains(message, "is not authorized") {
				return true, err
			}

			// IAM eventual consistency
			// XXX role has insufficient YYY permissions
			if strings.Contains(message, "role has insufficient") {
				return true, err
			}

			// IAM eventual consistency
			// Account XXX should have YYY role with trust relationship to Role ZZZ
			if strings.Contains(message, "role with trust relationship") {
				return true, err
			}

			// IAM eventual consistency
			if strings.Contains(message, "The security token included in the request is invalid") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFormation StackSet (%s) Instance: %s", stackSetName, err)
	}

	return append(diags, resourceStackSetInstanceRead(ctx, d, meta)...)
}

func resourceStackSetInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationClient(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), stackSetInstanceResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	stackSetName, accountOrOrgID, region := parts[0], parts[1], parts[2]
	d.Set(names.AttrRegion, region)
	d.Set("stack_set_name", stackSetName)

	callAs := d.Get("call_as").(string)

	if itypes.IsAWSAccountID(accountOrOrgID) {
		// Stack instances deployed by account ID
		stackInstance, err := findStackInstanceByFourPartKey(ctx, conn, stackSetName, accountOrOrgID, region, callAs)

		if !d.IsNewResource() && tfresource.NotFound(err) {
			log.Printf("[WARN] CloudFormation StackSet Instance (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading CloudFormation StackSet Instance (%s): %s", d.Id(), err)
		}

		d.Set(names.AttrAccountID, stackInstance.Account)
		d.Set("organizational_unit_id", stackInstance.OrganizationalUnitId)
		if err := d.Set("parameter_overrides", flattenAllParameters(stackInstance.ParameterOverrides)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting parameters: %s", err)
		}

		d.Set("stack_id", stackInstance.StackId)
		d.Set("stack_instance_summaries", nil)
	} else {
		// Stack instances deployed by organizational unit ID
		orgIDs := strings.Split(accountOrOrgID, "/")

		summaries, err := findStackInstanceSummariesByFourPartKey(ctx, conn, stackSetName, region, callAs, orgIDs)

		if !d.IsNewResource() && tfresource.NotFound(err) {
			log.Printf("[WARN] CloudFormation StackSet Instance (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "finding CloudFormation StackSet Instance (%s): %s", d.Id(), err)
		}

		d.Set("stack_instance_summaries", flattenStackInstanceSummaries(summaries))
	}

	return diags
}

func resourceStackSetInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationClient(ctx)

	if d.HasChanges("parameter_overrides", "operation_preferences") {
		parts, err := flex.ExpandResourceId(d.Id(), stackSetInstanceResourceIDPartCount, false)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		stackSetName, accountOrOrgID, region := parts[0], parts[1], parts[2]
		input := &cloudformation.UpdateStackInstancesInput{
			Accounts:           []string{accountOrOrgID},
			OperationId:        aws.String(sdkid.UniqueId()),
			ParameterOverrides: []awstypes.Parameter{},
			Regions:            []string{region},
			StackSetName:       aws.String(stackSetName),
		}

		callAs := d.Get("call_as").(string)
		if v, ok := d.GetOk("call_as"); ok {
			input.CallAs = awstypes.CallAs(v.(string))
		}

		if v, ok := d.GetOk("parameter_overrides"); ok {
			input.ParameterOverrides = expandParameters(v.(map[string]interface{}))
		}

		if v, ok := d.GetOk("operation_preferences"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.OperationPreferences = expandOperationPreferences(v.([]interface{})[0].(map[string]interface{}))
		}

		output, err := conn.UpdateStackInstances(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CloudFormation StackSet Instance (%s): %s", d.Id(), err)
		}

		if _, err := waitStackSetOperationSucceeded(ctx, conn, stackSetName, aws.ToString(output.OperationId), callAs, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for CloudFormation StackSet Instance (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceStackSetInstanceRead(ctx, d, meta)...)
}

func resourceStackSetInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationClient(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), stackSetInstanceResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	stackSetName, accountOrOrgID, region := parts[0], parts[1], parts[2]
	input := &cloudformation.DeleteStackInstancesInput{
		Accounts:     []string{accountOrOrgID},
		OperationId:  aws.String(sdkid.UniqueId()),
		Regions:      []string{region},
		RetainStacks: aws.Bool(d.Get("retain_stack").(bool)),
		StackSetName: aws.String(stackSetName),
	}

	callAs := d.Get("call_as").(string)
	if v, ok := d.GetOk("call_as"); ok {
		input.CallAs = awstypes.CallAs(v.(string))
	}

	if v, ok := d.GetOk("deployment_targets"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		dt := expandDeploymentTargets(v.([]interface{}))
		// For instances associated with stack sets that use a self-managed permission model,
		// the organizational unit must be provided;
		input.Accounts = nil
		input.DeploymentTargets = dt
	}

	log.Printf("[DEBUG] Deleting CloudFormation StackSet Instance: %s", d.Id())
	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.OperationInProgressException](ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return conn.DeleteStackInstances(ctx, input)
	})

	if errs.IsA[*awstypes.StackInstanceNotFoundException](err) || errs.IsA[*awstypes.StackSetNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudFormation StackSet Instance (%s): %s", d.Id(), err)
	}

	if _, err := waitStackSetOperationSucceeded(ctx, conn, stackSetName, aws.ToString(outputRaw.(*cloudformation.DeleteStackInstancesOutput).OperationId), callAs, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudFormation StackSet Instance (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func resourceStackSetInstanceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	switch parts := strings.Split(d.Id(), flex.ResourceIdSeparator); len(parts) {
	case 3:
	case 4:
		d.SetId(strings.Join([]string{parts[0], parts[1], parts[2]}, flex.ResourceIdSeparator))
		d.Set("call_as", parts[3])
	default:
		return []*schema.ResourceData{}, fmt.Errorf("unexpected format for import ID (%[1]s), use: STACKSETNAME%[2]sACCOUNTID%[2]sREGION or STACKSETNAME%[2]sACCOUNTID%[2]sREGION%[2]sCALLAS", d.Id(), flex.ResourceIdSeparator)
	}

	return []*schema.ResourceData{d}, nil
}

func findStackInstanceSummariesByFourPartKey(ctx context.Context, conn *cloudformation.Client, stackSetName, region, callAs string, orgIDs []string) ([]awstypes.StackInstanceSummary, error) {
	input := &cloudformation.ListStackInstancesInput{
		StackInstanceRegion: aws.String(region),
		StackSetName:        aws.String(stackSetName),
	}
	if callAs != "" {
		input.CallAs = awstypes.CallAs(callAs)
	}
	var output []awstypes.StackInstanceSummary

	pages := cloudformation.NewListStackInstancesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.StackSetNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Summaries {
			if slices.Contains(orgIDs, aws.ToString(v.OrganizationalUnitId)) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func findStackInstanceByFourPartKey(ctx context.Context, conn *cloudformation.Client, stackSetName, accountID, region, callAs string) (*awstypes.StackInstance, error) {
	input := &cloudformation.DescribeStackInstanceInput{
		StackInstanceAccount: aws.String(accountID),
		StackInstanceRegion:  aws.String(region),
		StackSetName:         aws.String(stackSetName),
	}
	if callAs != "" {
		input.CallAs = awstypes.CallAs(callAs)
	}

	output, err := conn.DescribeStackInstance(ctx, input)

	if errs.IsA[*awstypes.StackInstanceNotFoundException](err) || errs.IsA[*awstypes.StackSetNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.StackInstance == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.StackInstance, nil
}

func expandDeploymentTargets(tfList []interface{}) *awstypes.DeploymentTargets {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	dt := &awstypes.DeploymentTargets{}
	if v, ok := tfMap["organizational_unit_ids"].(*schema.Set); ok && v.Len() > 0 {
		dt.OrganizationalUnitIds = flex.ExpandStringValueSet(v)
	}
	if v, ok := tfMap["account_filter_type"].(string); ok && len(v) > 0 {
		dt.AccountFilterType = awstypes.AccountFilterType(v)
	}
	if v, ok := tfMap["accounts"].(*schema.Set); ok && v.Len() > 0 {
		dt.Accounts = flex.ExpandStringValueSet(v)
	}
	if v, ok := tfMap["accounts_url"].(string); ok && len(v) > 0 {
		dt.AccountsUrl = aws.String(v)
	}

	return dt
}

func flattenStackInstanceSummaries(apiObject []awstypes.StackInstanceSummary) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	tfList := []interface{}{}
	for _, obj := range apiObject {
		m := map[string]interface{}{
			names.AttrAccountID:      obj.Account,
			"organizational_unit_id": obj.OrganizationalUnitId,
			"stack_id":               obj.StackId,
		}
		tfList = append(tfList, m)
	}

	return tfList
}
