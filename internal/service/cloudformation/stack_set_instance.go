package cloudformation

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceStackSetInstance() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStackSetInstanceCreate,
		ReadWithoutTimeout:   resourceStackSetInstanceRead,
		UpdateWithoutTimeout: resourceStackSetInstanceUpdate,
		DeleteWithoutTimeout: resourceStackSetInstanceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(StackSetInstanceCreatedDefaultTimeout),
			Update: schema.DefaultTimeout(StackSetInstanceUpdatedDefaultTimeout),
			Delete: schema.DefaultTimeout(StackSetInstanceDeletedDefaultTimeout),
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
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"organizational_unit_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringMatch(regexp.MustCompile(`^(ou-[a-z0-9]{4,32}-[a-z0-9]{8,32}|r-[a-z0-9]{4,32})$`), ""),
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
								ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-]{1,128}$`), ""),
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
			"retain_stack": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"stack_id": {
				Type:     schema.TypeString,
				Computed: true,
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
	conn := meta.(*conns.AWSClient).CloudFormationConn()

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

	callAs := d.Get("call_as").(string)
	if v, ok := d.GetOk("call_as"); ok {
		input.CallAs = aws.String(v.(string))
	}

	if v, ok := d.GetOk("deployment_targets"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		dt := expandDeploymentTargets(v.([]interface{}))
		// temporarily set the accountId to the DeploymentTarget IDs
		// to later inform the Read CRUD operation if the true accountID needs to be determined
		accountID = strings.Join(aws.StringValueSlice(dt.OrganizationalUnitIds), "/")
		input.DeploymentTargets = dt
	} else {
		input.Accounts = aws.StringSlice([]string{accountID})
	}

	if v, ok := d.GetOk("parameter_overrides"); ok {
		input.ParameterOverrides = expandParameters(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("operation_preferences"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.OperationPreferences = expandOperationPreferences(v.([]interface{})[0].(map[string]interface{}))
	}

	log.Printf("[DEBUG] Creating CloudFormation StackSet Instance: %s", input)
	_, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (interface{}, error) {
			input.OperationId = aws.String(resource.UniqueId())

			output, err := conn.CreateStackInstancesWithContext(ctx, input)

			if err != nil {
				return nil, err
			}

			d.SetId(StackSetInstanceCreateResourceID(stackSetName, accountID, region))

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

			return false, fmt.Errorf("error waiting for CloudFormation StackSet Instance (%s) creation: %w", d.Id(), err)
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFormation StackSet (%s) Instance: %s", stackSetName, err)
	}

	return append(diags, resourceStackSetInstanceRead(ctx, d, meta)...)
}

func resourceStackSetInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationConn()

	stackSetName, accountID, region, err := StackSetInstanceParseResourceID(d.Id())

	callAs := d.Get("call_as").(string)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFormation StackSet Instance (%s): %s", d.Id(), err)
	}

	// Determine correct account ID for the Instance if created with deployment targets;
	// we only expect the accountID to be the organization root ID or organizational unit (OU) IDs
	// separated by a slash after creation.
	if regexp.MustCompile(`(ou-[a-z0-9]{4,32}-[a-z0-9]{8,32}|r-[a-z0-9]{4,32})`).MatchString(accountID) {
		orgIDs := strings.Split(accountID, "/")
		accountID, err = FindStackInstanceAccountIdByOrgIDs(ctx, conn, stackSetName, region, callAs, orgIDs)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "finding CloudFormation StackSet Instance (%s) Account: %s", d.Id(), err)
		}

		d.SetId(StackSetInstanceCreateResourceID(stackSetName, accountID, region))
	}

	stackInstance, err := FindStackInstanceByName(ctx, conn, stackSetName, accountID, region, callAs)

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

	d.Set("region", stackInstance.Region)
	d.Set("stack_id", stackInstance.StackId)
	d.Set("stack_set_name", stackSetName)

	return diags
}

func resourceStackSetInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationConn()

	if d.HasChanges("deployment_targets", "parameter_overrides", "operation_preferences") {
		stackSetName, accountID, region, err := StackSetInstanceParseResourceID(d.Id())

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CloudFormation StackSet Instance (%s): %s", d.Id(), err)
		}

		input := &cloudformation.UpdateStackInstancesInput{
			Accounts:           aws.StringSlice([]string{accountID}),
			OperationId:        aws.String(resource.UniqueId()),
			ParameterOverrides: []*cloudformation.Parameter{},
			Regions:            aws.StringSlice([]string{region}),
			StackSetName:       aws.String(stackSetName),
		}

		callAs := d.Get("call_as").(string)
		if v, ok := d.GetOk("call_as"); ok {
			input.CallAs = aws.String(v.(string))
		}

		if v, ok := d.GetOk("deployment_targets"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			// reset input Accounts as the API accepts only 1 of Accounts and DeploymentTargets
			input.Accounts = nil
			input.DeploymentTargets = expandDeploymentTargets(v.([]interface{}))
		}

		if v, ok := d.GetOk("parameter_overrides"); ok {
			input.ParameterOverrides = expandParameters(v.(map[string]interface{}))
		}

		if v, ok := d.GetOk("operation_preferences"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.OperationPreferences = expandOperationPreferences(v.([]interface{})[0].(map[string]interface{}))
		}

		log.Printf("[DEBUG] Updating CloudFormation StackSet Instance: %s", input)
		output, err := conn.UpdateStackInstancesWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CloudFormation StackSet Instance (%s): %s", d.Id(), err)
		}

		if _, err := WaitStackSetOperationSucceeded(ctx, conn, stackSetName, aws.StringValue(output.OperationId), callAs, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CloudFormation StackSet Instance (%s): waiting for completion: %s", d.Id(), err)
		}
	}

	return append(diags, resourceStackSetInstanceRead(ctx, d, meta)...)
}

func resourceStackSetInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationConn()

	stackSetName, accountID, region, err := StackSetInstanceParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudFormation StackSet Instance (%s): %s", d.Id(), err)
	}

	input := &cloudformation.DeleteStackInstancesInput{
		Accounts:     aws.StringSlice([]string{accountID}),
		OperationId:  aws.String(resource.UniqueId()),
		Regions:      aws.StringSlice([]string{region}),
		RetainStacks: aws.Bool(d.Get("retain_stack").(bool)),
		StackSetName: aws.String(stackSetName),
	}

	callAs := d.Get("call_as").(string)
	if v, ok := d.GetOk("call_as"); ok {
		input.CallAs = aws.String(v.(string))
	}

	if v, ok := d.GetOk("organizational_unit_id"); ok {
		// For instances associated with stack sets that use a self-managed permission model,
		// the organizational unit must be provided;
		input.Accounts = nil
		input.DeploymentTargets = &cloudformation.DeploymentTargets{
			OrganizationalUnitIds: aws.StringSlice([]string{v.(string)}),
		}
	}

	log.Printf("[DEBUG] Deleting CloudFormation StackSet Instance: %s", d.Id())
	output, err := conn.DeleteStackInstancesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudformation.ErrCodeStackInstanceNotFoundException) || tfawserr.ErrCodeEquals(err, cloudformation.ErrCodeStackSetNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudFormation StackSet Instance (%s): %s", d.Id(), err)
	}

	if _, err := WaitStackSetOperationSucceeded(ctx, conn, stackSetName, aws.StringValue(output.OperationId), callAs, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudFormation StackSet Instance (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}

func expandDeploymentTargets(l []interface{}) *cloudformation.DeploymentTargets {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	dt := &cloudformation.DeploymentTargets{}

	if v, ok := tfMap["organizational_unit_ids"].(*schema.Set); ok && v.Len() > 0 {
		dt.OrganizationalUnitIds = flex.ExpandStringSet(v)
	}

	return dt
}
