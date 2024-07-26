// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
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

// @SDKResource("aws_config_organization_conformance_pack", name="Organization Conformance Pack")
func resourceOrganizationConformancePack() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOrganizationConformancePackCreate,
		ReadWithoutTimeout:   resourceOrganizationConformancePackRead,
		UpdateWithoutTimeout: resourceOrganizationConformancePackUpdate,
		DeleteWithoutTimeout: resourceOrganizationConformancePackDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"delivery_s3_bucket": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexache.MustCompile(`^awsconfigconforms`), `must begin with "awsconfigconforms"`),
				),
			},
			"delivery_s3_key_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
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
			"input_parameter": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 60,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"parameter_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"parameter_value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexache.MustCompile(`^[A-Za-z]`), "must begin with alphabetic character"),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z-]+$`), "must contain only alphanumeric and hyphen characters"),
				),
			},
			"template_body": {
				Type:                  schema.TypeString,
				Optional:              true,
				DiffSuppressFunc:      verify.SuppressEquivalentJSONOrYAMLDiffs,
				DiffSuppressOnRefresh: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 51200),
					verify.ValidStringIsJSONOrYAML,
				),
				ConflictsWith: []string{"template_s3_uri"},
			},
			"template_s3_uri": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 1024),
					validation.StringMatch(regexache.MustCompile(`^s3://`), "must begin with s3://"),
				),
				ConflictsWith: []string{"template_body"},
			},
		},
	}
}

func resourceOrganizationConformancePackCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &configservice.PutOrganizationConformancePackInput{
		OrganizationConformancePackName: aws.String(name),
	}

	if v, ok := d.GetOk("delivery_s3_bucket"); ok {
		input.DeliveryS3Bucket = aws.String(v.(string))
	}

	if v, ok := d.GetOk("delivery_s3_key_prefix"); ok {
		input.DeliveryS3KeyPrefix = aws.String(v.(string))
	}

	if v, ok := d.GetOk("excluded_accounts"); ok && v.(*schema.Set).Len() > 0 {
		input.ExcludedAccounts = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("input_parameter"); ok {
		input.ConformancePackInputParameters = expandConformancePackInputParameters(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("template_body"); ok {
		input.TemplateBody = aws.String(v.(string))
	}

	if v, ok := d.GetOk("template_s3_uri"); ok {
		input.TemplateS3Uri = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhenIsA[*types.OrganizationAccessDeniedException](ctx, organizationsPropagationTimeout, func() (interface{}, error) {
		return conn.PutOrganizationConformancePack(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ConfigService Organization Conformance Pack (%s): %s", name, err)
	}

	d.SetId(name)

	if _, err := waitOrganizationConformancePackCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ConfigService Organization Conformance Pack (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceOrganizationConformancePackRead(ctx, d, meta)...)
}

func resourceOrganizationConformancePackRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	pack, err := findOrganizationConformancePackByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ConfigService Organization Conformance Pack (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ConfigService Organization Conformance Pack (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, pack.OrganizationConformancePackArn)
	d.Set("delivery_s3_bucket", pack.DeliveryS3Bucket)
	d.Set("delivery_s3_key_prefix", pack.DeliveryS3KeyPrefix)
	d.Set("excluded_accounts", pack.ExcludedAccounts)
	if err = d.Set("input_parameter", flattenConformancePackInputParameters(pack.ConformancePackInputParameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting input_parameter: %s", err)
	}
	d.Set(names.AttrName, pack.OrganizationConformancePackName)

	return diags
}

func resourceOrganizationConformancePackUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	input := &configservice.PutOrganizationConformancePackInput{
		OrganizationConformancePackName: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("delivery_s3_bucket"); ok {
		input.DeliveryS3Bucket = aws.String(v.(string))
	}

	if v, ok := d.GetOk("delivery_s3_key_prefix"); ok {
		input.DeliveryS3KeyPrefix = aws.String(v.(string))
	}

	if v, ok := d.GetOk("excluded_accounts"); ok {
		input.ExcludedAccounts = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("input_parameter"); ok {
		input.ConformancePackInputParameters = expandConformancePackInputParameters(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("template_body"); ok {
		input.TemplateBody = aws.String(v.(string))
	}

	if v, ok := d.GetOk("template_s3_uri"); ok {
		input.TemplateS3Uri = aws.String(v.(string))
	}

	_, err := conn.PutOrganizationConformancePack(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating ConfigService Organization Conformance Pack (%s): %s", d.Id(), err)
	}

	if _, err := waitOrganizationConformancePackUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ConfigService Organization Conformance Pack (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceOrganizationConformancePackRead(ctx, d, meta)...)
}

func resourceOrganizationConformancePackDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	const (
		timeout = 5 * time.Minute
	)
	log.Printf("[DEBUG] Deleting ConfigService Organization Conformance Pack: %s", d.Id())
	_, err := tfresource.RetryWhenIsA[*types.ResourceInUseException](ctx, timeout, func() (interface{}, error) {
		return conn.DeleteOrganizationConformancePack(ctx, &configservice.DeleteOrganizationConformancePackInput{
			OrganizationConformancePackName: aws.String(d.Id()),
		})
	})

	if errs.IsA[*types.NoSuchOrganizationConformancePackException](err) || errs.IsA[*types.OrganizationAccessDeniedException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ConfigService Organization Conformance Pack (%s): %s", d.Id(), err)
	}

	if _, err := waitOrganizationConformancePackDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ConfigService Organization Conformance Pack (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findOrganizationConformancePackByName(ctx context.Context, conn *configservice.Client, name string) (*types.OrganizationConformancePack, error) {
	input := &configservice.DescribeOrganizationConformancePacksInput{
		OrganizationConformancePackNames: []string{name},
	}

	return findOrganizationConformancePack(ctx, conn, input)
}

func findOrganizationConformancePack(ctx context.Context, conn *configservice.Client, input *configservice.DescribeOrganizationConformancePacksInput) (*types.OrganizationConformancePack, error) {
	output, err := findOrganizationConformancePacks(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findOrganizationConformancePacks(ctx context.Context, conn *configservice.Client, input *configservice.DescribeOrganizationConformancePacksInput) ([]types.OrganizationConformancePack, error) {
	var output []types.OrganizationConformancePack

	pages := configservice.NewDescribeOrganizationConformancePacksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.NoSuchOrganizationConformancePackException](err) {
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

		output = append(output, page.OrganizationConformancePacks...)
	}

	return output, nil
}

func findOrganizationConformancePackStatusByName(ctx context.Context, conn *configservice.Client, name string) (*types.OrganizationConformancePackStatus, error) {
	input := &configservice.DescribeOrganizationConformancePackStatusesInput{
		OrganizationConformancePackNames: []string{name},
	}

	output, err := findOrganizationConformancePackStatus(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := output.Status; status == types.OrganizationResourceStatusDeleteSuccessful {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output, nil
}

func findOrganizationConformancePackStatus(ctx context.Context, conn *configservice.Client, input *configservice.DescribeOrganizationConformancePackStatusesInput) (*types.OrganizationConformancePackStatus, error) {
	output, err := findOrganizationConformancePackStatuses(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findOrganizationConformancePackStatuses(ctx context.Context, conn *configservice.Client, input *configservice.DescribeOrganizationConformancePackStatusesInput) ([]types.OrganizationConformancePackStatus, error) {
	var output []types.OrganizationConformancePackStatus

	pages := configservice.NewDescribeOrganizationConformancePackStatusesPaginator(conn, input)
	for pages.HasMorePages() {
		const (
			timeout = 15 * time.Second
		)
		outputRaw, err := tfresource.RetryWhenIsA[*types.OrganizationAccessDeniedException](ctx, timeout, func() (interface{}, error) {
			return pages.NextPage(ctx)
		})

		if errs.IsA[*types.NoSuchOrganizationConformancePackException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, outputRaw.(*configservice.DescribeOrganizationConformancePackStatusesOutput).OrganizationConformancePackStatuses...)
	}

	return output, nil
}

func findOrganizationConformancePackDetailedStatusesByTwoPartKey(ctx context.Context, conn *configservice.Client, name string, status types.OrganizationResourceDetailedStatus) ([]types.OrganizationConformancePackDetailedStatus, error) {
	input := &configservice.GetOrganizationConformancePackDetailedStatusInput{
		Filters: &types.OrganizationResourceDetailedStatusFilters{
			Status: status,
		},
		OrganizationConformancePackName: aws.String(name),
	}

	return findOrganizationConformancePackDetailedStatuses(ctx, conn, input)
}

func findOrganizationConformancePackDetailedStatuses(ctx context.Context, conn *configservice.Client, input *configservice.GetOrganizationConformancePackDetailedStatusInput) ([]types.OrganizationConformancePackDetailedStatus, error) {
	var output []types.OrganizationConformancePackDetailedStatus

	pages := configservice.NewGetOrganizationConformancePackDetailedStatusPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.NoSuchOrganizationConformancePackException](err) {
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

		output = append(output, page.OrganizationConformancePackDetailedStatuses...)
	}

	return output, nil
}

func statusOrganizationConformancePack(ctx context.Context, conn *configservice.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findOrganizationConformancePackStatusByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), err
	}
}

func waitOrganizationConformancePackCreated(ctx context.Context, conn *configservice.Client, name string, timeout time.Duration) (*types.OrganizationConformancePackStatus, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        enum.Slice(types.OrganizationResourceStatusCreateInProgress),
		Target:         enum.Slice(types.OrganizationResourceStatusCreateSuccessful),
		Refresh:        statusOrganizationConformancePack(ctx, conn, name),
		Timeout:        timeout,
		Delay:          30 * time.Second,
		NotFoundChecks: 10,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.OrganizationConformancePackStatus); ok {
		tfresource.SetLastError(err, organizationConformancePackStatusError(ctx, conn, output))

		return output, err
	}

	return nil, err
}

func waitOrganizationConformancePackUpdated(ctx context.Context, conn *configservice.Client, name string, timeout time.Duration) (*types.OrganizationConformancePackStatus, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.OrganizationResourceStatusUpdateInProgress),
		Target:  enum.Slice(types.OrganizationResourceStatusUpdateSuccessful),
		Refresh: statusOrganizationConformancePack(ctx, conn, name),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.OrganizationConformancePackStatus); ok {
		tfresource.SetLastError(err, organizationConformancePackStatusError(ctx, conn, output))

		return output, err
	}

	return nil, err
}

func waitOrganizationConformancePackDeleted(ctx context.Context, conn *configservice.Client, name string, timeout time.Duration) (*types.OrganizationConformancePackStatus, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.OrganizationResourceStatusDeleteInProgress),
		Target:                    []string{},
		Refresh:                   statusOrganizationConformancePack(ctx, conn, name),
		Timeout:                   timeout,
		Delay:                     10 * time.Second,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.OrganizationConformancePackStatus); ok {
		tfresource.SetLastError(err, organizationConformancePackStatusError(ctx, conn, output))

		return output, err
	}

	return nil, err
}

func organizationConformancePackStatusError(ctx context.Context, conn *configservice.Client, apiObject *types.OrganizationConformancePackStatus) error {
	errs := []error{fmt.Errorf("%s: %s", aws.ToString(apiObject.ErrorCode), aws.ToString(apiObject.ErrorMessage))}

	var detailedStatus types.OrganizationResourceDetailedStatus
	switch apiObject.Status {
	case types.OrganizationResourceStatusCreateFailed:
		detailedStatus = types.OrganizationResourceDetailedStatusCreateFailed
	case types.OrganizationResourceStatusUpdateFailed:
		detailedStatus = types.OrganizationResourceDetailedStatusUpdateFailed
	case types.OrganizationResourceStatusDeleteFailed:
		detailedStatus = types.OrganizationResourceDetailedStatusDeleteFailed
	}

	if detailedStatus != "" {
		if v, err := findOrganizationConformancePackDetailedStatusesByTwoPartKey(ctx, conn, aws.ToString(apiObject.OrganizationConformancePackName), detailedStatus); err == nil {
			for _, v := range v {
				err := fmt.Errorf("%s: %s", aws.ToString(v.ErrorCode), aws.ToString(v.ErrorMessage))
				errs = append(errs, fmt.Errorf("Account ID (%s): %w", aws.ToString(v.AccountId), err))
			}
		}
	}

	return errors.Join(errs...)
}
