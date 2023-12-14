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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_cloudformation_stack_set_instance")
func ResourceStackSetInstance() *schema.Resource {
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
			"account_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  verify.ValidAccountID,
				ConflictsWith: []string{"deployment_targets"},
			},
			"call_as": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      cloudformation.CallAsSelf,
				ValidateFunc: validation.StringInSlice(cloudformation.CallAs_Values(), false),
			},
			"deployment_targets": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"organizational_unit_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringMatch(regexache.MustCompile(`^(ou-[0-9a-z]{4,32}-[0-9a-z]{8,32}|r-[0-9a-z]{4,32})$`), ""),
							},
						},
					},
				},
				ConflictsWith: []string{"account_id"},
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
						"region_concurrency_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(cloudformation.RegionConcurrencyType_Values(), false),
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
			"region": {
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
						"account_id": {
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

var (
	accountIDRegexp = regexache.MustCompile(`^\d{12}$`)
)

func resourceStackSetInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationConn(ctx)

	region := meta.(*conns.AWSClient).Region
	if v, ok := d.GetOk("region"); ok {
		region = v.(string)
	}

	stackSetName := d.Get("stack_set_name").(string)
	input := &cloudformation.CreateStackInstancesInput{
		Regions:      aws.StringSlice([]string{region}),
		StackSetName: aws.String(stackSetName),
	}

	accountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
	}
	// accountOrOrgID will either be account_id or a slash-delimited list of
	// organizational_unit_id's from the deployment_targets argument. This
	// is composed with stack_set_name and region to form the resources ID.
	accountOrOrgID := accountID

	if v, ok := d.GetOk("deployment_targets"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		dt := expandDeploymentTargets(v.([]interface{}))
		accountOrOrgID = strings.Join(aws.StringValueSlice(dt.OrganizationalUnitIds), "/")
		input.DeploymentTargets = dt
	} else {
		d.Set("account_id", accountID)
		input.Accounts = aws.StringSlice([]string{accountID})
	}

	callAs := d.Get("call_as").(string)
	if v, ok := d.GetOk("call_as"); ok {
		input.CallAs = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parameter_overrides"); ok {
		input.ParameterOverrides = expandParameters(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("operation_preferences"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.OperationPreferences = expandOperationPreferences(v.([]interface{})[0].(map[string]interface{}))
	}

	_, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (interface{}, error) {
			input.OperationId = aws.String(id.UniqueId())

			output, err := conn.CreateStackInstancesWithContext(ctx, input)

			if err != nil {
				return nil, err
			}

			d.SetId(StackSetInstanceCreateResourceID(stackSetName, accountOrOrgID, region))

			operation, err := WaitStackSetOperationSucceeded(ctx, conn, stackSetName, aws.StringValue(output.OperationId), callAs, d.Timeout(schema.TimeoutCreate))
			if err != nil {
				return nil, fmt.Errorf("waiting for completion: %w", err)
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
	conn := meta.(*conns.AWSClient).CloudFormationConn(ctx)

	stackSetName, accountOrOrgID, region, err := StackSetInstanceParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("region", region)
	d.Set("stack_set_name", stackSetName)

	callAs := d.Get("call_as").(string)

	if accountIDRegexp.MatchString(accountOrOrgID) {
		// Stack instances deployed by account ID
		stackInstance, err := FindStackInstanceByName(ctx, conn, stackSetName, accountOrOrgID, region, callAs)

		if !d.IsNewResource() && tfresource.NotFound(err) {
			log.Printf("[WARN] CloudFormation StackSet Instance (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading CloudFormation StackSet Instance (%s): %s", d.Id(), err)
		}

		d.Set("account_id", stackInstance.Account)
		d.Set("organizational_unit_id", stackInstance.OrganizationalUnitId)
		if err := d.Set("parameter_overrides", flattenAllParameters(stackInstance.ParameterOverrides)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting parameters: %s", err)
		}

		d.Set("stack_id", stackInstance.StackId)
		d.Set("stack_instance_summaries", nil)
	} else {
		// Stack instances deployed by organizational unit ID
		orgIDs := strings.Split(accountOrOrgID, "/")

		summaries, err := FindStackInstanceSummariesByOrgIDs(ctx, conn, stackSetName, region, callAs, orgIDs)

		if !d.IsNewResource() && tfresource.NotFound(err) {
			log.Printf("[WARN] CloudFormation StackSet Instance (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "finding CloudFormation StackSet Instance (%s): %s", d.Id(), err)
		}

		d.Set("deployment_targets", flattenDeploymentTargetsFromSlice(orgIDs))
		d.Set("stack_instance_summaries", flattenStackInstanceSummaries(summaries))
	}

	return diags
}

func resourceStackSetInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationConn(ctx)

	if d.HasChanges("deployment_targets", "parameter_overrides", "operation_preferences") {
		stackSetName, accountOrOrgID, region, err := StackSetInstanceParseResourceID(d.Id())
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &cloudformation.UpdateStackInstancesInput{
			Accounts:           aws.StringSlice([]string{accountOrOrgID}),
			OperationId:        aws.String(id.UniqueId()),
			ParameterOverrides: []*cloudformation.Parameter{},
			Regions:            aws.StringSlice([]string{region}),
			StackSetName:       aws.String(stackSetName),
		}

		callAs := d.Get("call_as").(string)
		if v, ok := d.GetOk("call_as"); ok {
			input.CallAs = aws.String(v.(string))
		}

		if v, ok := d.GetOk("deployment_targets"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			dt := expandDeploymentTargets(v.([]interface{}))
			// reset input Accounts as the API accepts only 1 of Accounts and DeploymentTargets
			input.Accounts = nil
			input.DeploymentTargets = dt
		}

		if v, ok := d.GetOk("parameter_overrides"); ok {
			input.ParameterOverrides = expandParameters(v.(map[string]interface{}))
		}

		if v, ok := d.GetOk("operation_preferences"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.OperationPreferences = expandOperationPreferences(v.([]interface{})[0].(map[string]interface{}))
		}

		output, err := conn.UpdateStackInstancesWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CloudFormation StackSet Instance (%s): %s", d.Id(), err)
		}

		if _, err := WaitStackSetOperationSucceeded(ctx, conn, stackSetName, aws.StringValue(output.OperationId), callAs, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for CloudFormation StackSet Instance (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceStackSetInstanceRead(ctx, d, meta)...)
}

func resourceStackSetInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationConn(ctx)

	stackSetName, accountOrOrgID, region, err := StackSetInstanceParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &cloudformation.DeleteStackInstancesInput{
		Accounts:     aws.StringSlice([]string{accountOrOrgID}),
		OperationId:  aws.String(id.UniqueId()),
		Regions:      aws.StringSlice([]string{region}),
		RetainStacks: aws.Bool(d.Get("retain_stack").(bool)),
		StackSetName: aws.String(stackSetName),
	}

	callAs := d.Get("call_as").(string)
	if v, ok := d.GetOk("call_as"); ok {
		input.CallAs = aws.String(v.(string))
	}

	if v, ok := d.GetOk("deployment_targets"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		dt := expandDeploymentTargets(v.([]interface{}))
		// For instances associated with stack sets that use a self-managed permission model,
		// the organizational unit must be provided;
		input.Accounts = nil
		input.DeploymentTargets = dt
	}

	log.Printf("[DEBUG] Deleting CloudFormation StackSet Instance: %s", d.Id())
	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return conn.DeleteStackInstancesWithContext(ctx, input)
	}, cloudformation.ErrCodeOperationInProgressException)

	if tfawserr.ErrCodeEquals(err, cloudformation.ErrCodeStackInstanceNotFoundException, cloudformation.ErrCodeStackSetNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudFormation StackSet Instance (%s): %s", d.Id(), err)
	}

	if _, err := WaitStackSetOperationSucceeded(ctx, conn, stackSetName, aws.StringValue(outputRaw.(*cloudformation.DeleteStackInstancesOutput).OperationId), callAs, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudFormation StackSet Instance (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func resourceStackSetInstanceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	switch parts := strings.Split(d.Id(), stackSetInstanceResourceIDSeparator); len(parts) {
	case 3:
	case 4:
		d.SetId(strings.Join([]string{parts[0], parts[1], parts[2]}, stackSetInstanceResourceIDSeparator))
		d.Set("call_as", parts[3])
	default:
		return []*schema.ResourceData{}, fmt.Errorf("unexpected format for import ID (%[1]s), use: STACKSETNAME%[2]sACCOUNTID%[2]sREGION or STACKSETNAME%[2]sACCOUNTID%[2]sREGION%[2]sCALLAS", d.Id(), stackSetInstanceResourceIDSeparator)
	}

	return []*schema.ResourceData{d}, nil
}

const stackSetInstanceResourceIDSeparator = ","

func StackSetInstanceCreateResourceID(stackSetName, accountID, region string) string {
	parts := []string{stackSetName, accountID, region}
	id := strings.Join(parts, stackSetInstanceResourceIDSeparator)

	return id
}

func StackSetInstanceParseResourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, stackSetInstanceResourceIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected STACKSETNAME%[2]sACCOUNTID%[2]sREGION", id, stackSetInstanceResourceIDSeparator)
}

func expandDeploymentTargets(tfList []interface{}) *cloudformation.DeploymentTargets {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	dt := &cloudformation.DeploymentTargets{}
	if v, ok := tfMap["organizational_unit_ids"].(*schema.Set); ok && v.Len() > 0 {
		dt.OrganizationalUnitIds = flex.ExpandStringSet(v)
	}

	return dt
}

// flattenDeployment targets converts a list of organizational units (typically
// parsed from the resource ID) into the Terraform representation of the
// deployment_targets attribute.
func flattenDeploymentTargetsFromSlice(orgIDs []string) []interface{} {
	tfList := []interface{}{}
	for _, ou := range orgIDs {
		tfList = append(tfList, ou)
	}

	m := map[string]interface{}{
		"organizational_unit_ids": tfList,
	}

	return []interface{}{m}
}

func flattenStackInstanceSummaries(apiObject []*cloudformation.StackInstanceSummary) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	tfList := []interface{}{}
	for _, obj := range apiObject {
		m := map[string]interface{}{
			"account_id":             obj.Account,
			"organizational_unit_id": obj.OrganizationalUnitId,
			"stack_id":               obj.StackId,
		}
		tfList = append(tfList, m)
	}

	return tfList
}
