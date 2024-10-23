// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/aws/aws-sdk-go-v2/service/configservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_config_organization_managed_rule", name="Organization Managed Rule")
func resourceOrganizationManagedRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOrganizationManagedRuleCreate,
		ReadWithoutTimeout:   resourceOrganizationManagedRuleRead,
		UpdateWithoutTimeout: resourceOrganizationManagedRuleUpdate,
		DeleteWithoutTimeout: resourceOrganizationManagedRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"excluded_accounts": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1000,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidAccountID,
				},
			},
			"input_parameters": {
				Type:                  schema.TypeString,
				Optional:              true,
				DiffSuppressFunc:      verify.SuppressEquivalentJSONDiffs,
				DiffSuppressOnRefresh: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 2048),
					validation.StringIsJSON,
				),
			},
			"maximum_execution_frequency": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.MaximumExecutionFrequency](),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"resource_id_scope": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 768),
			},
			"resource_types_scope": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 100,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(0, 256),
				},
			},
			"rule_identifier": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"tag_key_scope": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 128),
			},
			"tag_value_scope": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
		},
	}
}

func resourceOrganizationManagedRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &configservice.PutOrganizationConfigRuleInput{
		OrganizationConfigRuleName: aws.String(name),
		OrganizationManagedRuleMetadata: &types.OrganizationManagedRuleMetadata{
			RuleIdentifier: aws.String(d.Get("rule_identifier").(string)),
		},
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.OrganizationManagedRuleMetadata.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("excluded_accounts"); ok && v.(*schema.Set).Len() > 0 {
		input.ExcludedAccounts = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("input_parameters"); ok {
		input.OrganizationManagedRuleMetadata.InputParameters = aws.String(v.(string))
	}

	if v, ok := d.GetOk("maximum_execution_frequency"); ok {
		input.OrganizationManagedRuleMetadata.MaximumExecutionFrequency = types.MaximumExecutionFrequency(v.(string))
	}

	if v, ok := d.GetOk("resource_id_scope"); ok {
		input.OrganizationManagedRuleMetadata.ResourceIdScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resource_types_scope"); ok && v.(*schema.Set).Len() > 0 {
		input.OrganizationManagedRuleMetadata.ResourceTypesScope = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("tag_key_scope"); ok {
		input.OrganizationManagedRuleMetadata.TagKeyScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tag_value_scope"); ok {
		input.OrganizationManagedRuleMetadata.TagValueScope = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhenIsA[*types.OrganizationAccessDeniedException](ctx, organizationsPropagationTimeout, func() (interface{}, error) {
		return conn.PutOrganizationConfigRule(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ConfigService Organization Managed Rule (%s): %s", name, err)
	}

	d.SetId(name)

	if _, err := waitOrganizationConfigRuleCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ConfigService Organization Managed Rule (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceOrganizationManagedRuleRead(ctx, d, meta)...)
}

func resourceOrganizationManagedRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	configRule, err := findOrganizationManagedRuleByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ConfigService Organization Managed Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ConfigService Organization Managed Rule (%s): %s", d.Id(), err)
	}

	managedRule := configRule.OrganizationManagedRuleMetadata
	d.Set(names.AttrARN, configRule.OrganizationConfigRuleArn)
	d.Set(names.AttrDescription, managedRule.Description)
	d.Set("excluded_accounts", configRule.ExcludedAccounts)
	d.Set("input_parameters", managedRule.InputParameters)
	d.Set("maximum_execution_frequency", managedRule.MaximumExecutionFrequency)
	d.Set(names.AttrName, configRule.OrganizationConfigRuleName)
	d.Set("resource_id_scope", managedRule.ResourceIdScope)
	d.Set("resource_types_scope", managedRule.ResourceTypesScope)
	d.Set("rule_identifier", managedRule.RuleIdentifier)
	d.Set("tag_key_scope", managedRule.TagKeyScope)
	d.Set("tag_value_scope", managedRule.TagValueScope)

	return diags
}

func resourceOrganizationManagedRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	input := &configservice.PutOrganizationConfigRuleInput{
		OrganizationConfigRuleName: aws.String(d.Id()),
		OrganizationManagedRuleMetadata: &types.OrganizationManagedRuleMetadata{
			RuleIdentifier: aws.String(d.Get("rule_identifier").(string)),
		},
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.OrganizationManagedRuleMetadata.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("excluded_accounts"); ok && v.(*schema.Set).Len() > 0 {
		input.ExcludedAccounts = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("input_parameters"); ok {
		input.OrganizationManagedRuleMetadata.InputParameters = aws.String(v.(string))
	}

	if v, ok := d.GetOk("maximum_execution_frequency"); ok {
		input.OrganizationManagedRuleMetadata.MaximumExecutionFrequency = types.MaximumExecutionFrequency(v.(string))
	}

	if v, ok := d.GetOk("resource_id_scope"); ok {
		input.OrganizationManagedRuleMetadata.ResourceIdScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resource_types_scope"); ok && v.(*schema.Set).Len() > 0 {
		input.OrganizationManagedRuleMetadata.ResourceTypesScope = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("tag_key_scope"); ok {
		input.OrganizationManagedRuleMetadata.TagKeyScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tag_value_scope"); ok {
		input.OrganizationManagedRuleMetadata.TagValueScope = aws.String(v.(string))
	}

	_, err := conn.PutOrganizationConfigRule(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating ConfigService Organization Managed Rule (%s): %s", d.Id(), err)
	}

	if _, err := waitOrganizationConfigRuleUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ConfigService Organization Managed Rule (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceOrganizationManagedRuleRead(ctx, d, meta)...)
}

func resourceOrganizationManagedRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	const (
		timeout = 2 * time.Minute
	)
	log.Printf("[DEBUG] Deleting ConfigService Organization Managed Rule: %s", d.Id())
	_, err := tfresource.RetryWhenIsA[*types.ResourceInUseException](ctx, timeout, func() (interface{}, error) {
		return conn.DeleteOrganizationConfigRule(ctx, &configservice.DeleteOrganizationConfigRuleInput{
			OrganizationConfigRuleName: aws.String(d.Id()),
		})
	})

	if errs.IsA[*types.NoSuchOrganizationConfigRuleException](err) || errs.IsA[*types.OrganizationAccessDeniedException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ConfigService Organization Managed Rule (%s): %s", d.Id(), err)
	}

	if _, err := waitOrganizationConfigRuleDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ConfigService Organization Managed Rule (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findOrganizationManagedRuleByName(ctx context.Context, conn *configservice.Client, name string) (*types.OrganizationConfigRule, error) {
	output, err := findOrganizationConfigRuleByName(ctx, conn, name)

	if err != nil {
		return nil, err
	}

	if output.OrganizationManagedRuleMetadata == nil {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	return output, nil
}

func findOrganizationConfigRuleByName(ctx context.Context, conn *configservice.Client, name string) (*types.OrganizationConfigRule, error) {
	input := &configservice.DescribeOrganizationConfigRulesInput{
		OrganizationConfigRuleNames: []string{name},
	}

	return findOrganizationConfigRule(ctx, conn, input)
}

func findOrganizationConfigRule(ctx context.Context, conn *configservice.Client, input *configservice.DescribeOrganizationConfigRulesInput) (*types.OrganizationConfigRule, error) {
	output, err := findOrganizationConfigRules(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findOrganizationConfigRules(ctx context.Context, conn *configservice.Client, input *configservice.DescribeOrganizationConfigRulesInput) ([]types.OrganizationConfigRule, error) {
	var output []types.OrganizationConfigRule

	pages := configservice.NewDescribeOrganizationConfigRulesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.NoSuchOrganizationConfigRuleException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if errs.IsAErrorMessageContains[*types.OrganizationAccessDeniedException](err, "This action can only be made by accounts in an AWS Organization") {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.OrganizationConfigRules...)
	}

	return output, nil
}

func findOrganizationConfigRuleStatusByName(ctx context.Context, conn *configservice.Client, name string) (*types.OrganizationConfigRuleStatus, error) {
	input := &configservice.DescribeOrganizationConfigRuleStatusesInput{
		OrganizationConfigRuleNames: []string{name},
	}

	output, err := findOrganizationConfigRuleStatus(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := output.OrganizationRuleStatus; status == types.OrganizationRuleStatusDeleteSuccessful {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output, nil
}

func findOrganizationConfigRuleStatus(ctx context.Context, conn *configservice.Client, input *configservice.DescribeOrganizationConfigRuleStatusesInput) (*types.OrganizationConfigRuleStatus, error) {
	output, err := findOrganizationConfigRuleStatuses(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findOrganizationConfigRuleStatuses(ctx context.Context, conn *configservice.Client, input *configservice.DescribeOrganizationConfigRuleStatusesInput) ([]types.OrganizationConfigRuleStatus, error) {
	var output []types.OrganizationConfigRuleStatus

	pages := configservice.NewDescribeOrganizationConfigRuleStatusesPaginator(conn, input)
	for pages.HasMorePages() {
		const (
			timeout = 15 * time.Second
		)
		outputRaw, err := tfresource.RetryWhenIsA[*types.OrganizationAccessDeniedException](ctx, timeout, func() (interface{}, error) {
			return pages.NextPage(ctx)
		})

		if errs.IsA[*types.NoSuchOrganizationConfigRuleException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, outputRaw.(*configservice.DescribeOrganizationConfigRuleStatusesOutput).OrganizationConfigRuleStatuses...)
	}

	return output, nil
}

func findOrganizationConfigRuleDetailedStatusesByTwoPartKey(ctx context.Context, conn *configservice.Client, name string, status types.MemberAccountRuleStatus) ([]types.MemberAccountStatus, error) {
	input := &configservice.GetOrganizationConfigRuleDetailedStatusInput{
		Filters: &types.StatusDetailFilters{
			MemberAccountRuleStatus: status,
		},
		OrganizationConfigRuleName: aws.String(name),
	}

	return findOrganizationConfigRuleDetailedStatuses(ctx, conn, input)
}

func findOrganizationConfigRuleDetailedStatuses(ctx context.Context, conn *configservice.Client, input *configservice.GetOrganizationConfigRuleDetailedStatusInput) ([]types.MemberAccountStatus, error) {
	var output []types.MemberAccountStatus

	pages := configservice.NewGetOrganizationConfigRuleDetailedStatusPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.NoSuchOrganizationConfigRuleException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.OrganizationConfigRuleDetailedStatus...)
	}

	return output, nil
}

func statusOrganizationConfigRule(ctx context.Context, conn *configservice.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findOrganizationConfigRuleStatusByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.OrganizationRuleStatus), err
	}
}

func waitOrganizationConfigRuleCreated(ctx context.Context, conn *configservice.Client, name string, timeout time.Duration) (*types.OrganizationConfigRuleStatus, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:        enum.Slice(types.OrganizationRuleStatusCreateInProgress),
		Target:         enum.Slice(types.OrganizationRuleStatusCreateSuccessful),
		Refresh:        statusOrganizationConfigRule(ctx, conn, name),
		Timeout:        timeout,
		Delay:          30 * time.Second,
		NotFoundChecks: 10,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.OrganizationConfigRuleStatus); ok {
		tfresource.SetLastError(err, organizationConfigRuleStatusError(ctx, conn, output))

		return output, err
	}

	return nil, err
}

func waitOrganizationConfigRuleUpdated(ctx context.Context, conn *configservice.Client, name string, timeout time.Duration) (*types.OrganizationConfigRuleStatus, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.OrganizationRuleStatusUpdateInProgress),
		Target:  enum.Slice(types.OrganizationRuleStatusUpdateSuccessful),
		Refresh: statusOrganizationConfigRule(ctx, conn, name),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.OrganizationConfigRuleStatus); ok {
		tfresource.SetLastError(err, organizationConfigRuleStatusError(ctx, conn, output))

		return output, err
	}

	return nil, err
}

func waitOrganizationConfigRuleDeleted(ctx context.Context, conn *configservice.Client, name string, timeout time.Duration) (*types.OrganizationConfigRuleStatus, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.OrganizationRuleStatusDeleteInProgress),
		Target:                    []string{},
		Refresh:                   statusOrganizationConfigRule(ctx, conn, name),
		Timeout:                   timeout,
		Delay:                     10 * time.Second,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.OrganizationConfigRuleStatus); ok {
		tfresource.SetLastError(err, organizationConfigRuleStatusError(ctx, conn, output))

		return output, err
	}

	return nil, err
}

func organizationConfigRuleStatusError(ctx context.Context, conn *configservice.Client, apiObject *types.OrganizationConfigRuleStatus) error {
	errs := []error{fmt.Errorf("%s: %s", aws.ToString(apiObject.ErrorCode), aws.ToString(apiObject.ErrorMessage))}

	var detailedStatus types.MemberAccountRuleStatus
	switch apiObject.OrganizationRuleStatus {
	case types.OrganizationRuleStatusCreateFailed:
		detailedStatus = types.MemberAccountRuleStatusCreateFailed
	case types.OrganizationRuleStatusUpdateFailed:
		detailedStatus = types.MemberAccountRuleStatusUpdateFailed
	case types.OrganizationRuleStatusDeleteFailed:
		detailedStatus = types.MemberAccountRuleStatusDeleteFailed
	}

	if detailedStatus != "" {
		if v, err := findOrganizationConfigRuleDetailedStatusesByTwoPartKey(ctx, conn, aws.ToString(apiObject.OrganizationConfigRuleName), detailedStatus); err == nil {
			for _, v := range v {
				err := fmt.Errorf("%s: %s", aws.ToString(v.ErrorCode), aws.ToString(v.ErrorMessage))
				errs = append(errs, fmt.Errorf("Account ID (%s): %w", aws.ToString(v.AccountId), err))
			}
		}
	}

	return errors.Join(errs...)
}
