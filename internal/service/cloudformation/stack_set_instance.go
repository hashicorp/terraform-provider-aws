package cloudformation

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceStackSetInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceStackSetInstanceCreate,
		Read:   resourceStackSetInstanceRead,
		Update: resourceStackSetInstanceUpdate,
		Delete: resourceStackSetInstanceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(StackSetInstanceCreatedDefaultTimeout),
			Update: schema.DefaultTimeout(StackSetInstanceUpdatedDefaultTimeout),
			Delete: schema.DefaultTimeout(StackSetInstanceDeletedDefaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
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

func resourceStackSetInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFormationConn

	accountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
	}

	region := meta.(*conns.AWSClient).Region
	if v, ok := d.GetOk("region"); ok {
		region = v.(string)
	}

	stackSetName := d.Get("stack_set_name").(string)

	input := &cloudformation.CreateStackInstancesInput{
		Accounts:     aws.StringSlice([]string{accountID}),
		OperationId:  aws.String(resource.UniqueId()),
		Regions:      aws.StringSlice([]string{region}),
		StackSetName: aws.String(stackSetName),
	}

	if v, ok := d.GetOk("parameter_overrides"); ok {
		input.ParameterOverrides = expandParameters(v.(map[string]interface{}))
	}

	log.Printf("[DEBUG] Creating CloudFormation StackSet Instance: %s", input)
	err := resource.Retry(tfiam.PropagationTimeout, func() *resource.RetryError {
		output, err := conn.CreateStackInstances(input)

		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("error creating CloudFormation StackSet Instance: %w", err))
		}

		d.SetId(fmt.Sprintf("%s,%s,%s", stackSetName, accountID, region))

		err = WaitStackSetOperationSucceeded(conn, stackSetName, aws.StringValue(output.OperationId), d.Timeout(schema.TimeoutCreate))

		if err != nil {
			// IAM eventual consistency
			if strings.Contains(err.Error(), "AccountGate check failed") {
				input.OperationId = aws.String(resource.UniqueId())
				return resource.RetryableError(err)
			}

			// IAM eventual consistency
			// User: XXX is not authorized to perform: cloudformation:CreateStack on resource: YYY
			if strings.Contains(err.Error(), "is not authorized") {
				input.OperationId = aws.String(resource.UniqueId())
				return resource.RetryableError(err)
			}

			// IAM eventual consistency
			// XXX role has insufficient YYY permissions
			if strings.Contains(err.Error(), "role has insufficient") {
				input.OperationId = aws.String(resource.UniqueId())
				return resource.RetryableError(err)
			}

			// IAM eventual consistency
			// Account XXX should have YYY role with trust relationship to Role ZZZ
			if strings.Contains(err.Error(), "role with trust relationship") {
				input.OperationId = aws.String(resource.UniqueId())
				return resource.RetryableError(err)
			}

			// IAM eventual consistency
			if strings.Contains(err.Error(), "The security token included in the request is invalid") {
				input.OperationId = aws.String(resource.UniqueId())
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(fmt.Errorf("error waiting for CloudFormation StackSet Instance (%s) creation: %w", d.Id(), err))
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		var output *cloudformation.CreateStackInstancesOutput
		output, err = conn.CreateStackInstances(input)

		if err != nil {
			return fmt.Errorf("error creating CloudFormation StackSet Instance: %w", err)
		}

		d.SetId(fmt.Sprintf("%s,%s,%s", stackSetName, accountID, region))

		err = WaitStackSetOperationSucceeded(conn, stackSetName, aws.StringValue(output.OperationId), d.Timeout(schema.TimeoutCreate))

		if err != nil {
			return fmt.Errorf("error waiting for CloudFormation StackSet Instance (%s) creation: %w", d.Id(), err)
		}
	}

	if err != nil {
		return err
	}

	return resourceStackSetInstanceRead(d, meta)
}

func resourceStackSetInstanceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFormationConn

	stackSetName, accountID, region, err := StackSetInstanceParseID(d.Id())

	if err != nil {
		return err
	}

	input := &cloudformation.DescribeStackInstanceInput{
		StackInstanceAccount: aws.String(accountID),
		StackInstanceRegion:  aws.String(region),
		StackSetName:         aws.String(stackSetName),
	}

	log.Printf("[DEBUG] Reading CloudFormation StackSet Instance: %s", d.Id())
	output, err := conn.DescribeStackInstance(input)

	if tfawserr.ErrMessageContains(err, cloudformation.ErrCodeStackInstanceNotFoundException, "") {
		log.Printf("[WARN] CloudFormation StackSet Instance (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if tfawserr.ErrMessageContains(err, cloudformation.ErrCodeStackSetNotFoundException, "") {
		log.Printf("[WARN] CloudFormation StackSet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading CloudFormation StackSet Instance (%s): %s", d.Id(), err)
	}

	if output == nil || output.StackInstance == nil {
		return fmt.Errorf("error reading CloudFormation StackSet Instance (%s): empty response", d.Id())
	}

	stackInstance := output.StackInstance

	d.Set("account_id", stackInstance.Account)

	if err := d.Set("parameter_overrides", flattenAllCloudFormationParameters(stackInstance.ParameterOverrides)); err != nil {
		return fmt.Errorf("error setting parameters: %s", err)
	}

	d.Set("region", stackInstance.Region)
	d.Set("stack_id", stackInstance.StackId)
	d.Set("stack_set_name", stackSetName)

	return nil
}

func resourceStackSetInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFormationConn

	if d.HasChange("parameter_overrides") {
		stackSetName, accountID, region, err := StackSetInstanceParseID(d.Id())

		if err != nil {
			return err
		}

		input := &cloudformation.UpdateStackInstancesInput{
			Accounts:           aws.StringSlice([]string{accountID}),
			OperationId:        aws.String(resource.UniqueId()),
			ParameterOverrides: []*cloudformation.Parameter{},
			Regions:            aws.StringSlice([]string{region}),
			StackSetName:       aws.String(stackSetName),
		}

		if v, ok := d.GetOk("parameter_overrides"); ok {
			input.ParameterOverrides = expandParameters(v.(map[string]interface{}))
		}

		log.Printf("[DEBUG] Updating CloudFormation StackSet Instance: %s", input)
		output, err := conn.UpdateStackInstances(input)

		if err != nil {
			return fmt.Errorf("error updating CloudFormation StackSet Instance (%s): %s", d.Id(), err)
		}

		if err := WaitStackSetOperationSucceeded(conn, stackSetName, aws.StringValue(output.OperationId), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for CloudFormation StackSet Instance (%s) update: %s", d.Id(), err)
		}
	}

	return resourceStackSetInstanceRead(d, meta)
}

func resourceStackSetInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFormationConn

	stackSetName, accountID, region, err := StackSetInstanceParseID(d.Id())

	if err != nil {
		return err
	}

	input := &cloudformation.DeleteStackInstancesInput{
		Accounts:     aws.StringSlice([]string{accountID}),
		OperationId:  aws.String(resource.UniqueId()),
		Regions:      aws.StringSlice([]string{region}),
		RetainStacks: aws.Bool(d.Get("retain_stack").(bool)),
		StackSetName: aws.String(stackSetName),
	}

	log.Printf("[DEBUG] Deleting CloudFormation StackSet Instance: %s", d.Id())
	output, err := conn.DeleteStackInstances(input)

	if tfawserr.ErrMessageContains(err, cloudformation.ErrCodeStackInstanceNotFoundException, "") {
		return nil
	}

	if tfawserr.ErrMessageContains(err, cloudformation.ErrCodeStackSetNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting CloudFormation StackSet Instance (%s): %s", d.Id(), err)
	}

	if err := WaitStackSetOperationSucceeded(conn, stackSetName, aws.StringValue(output.OperationId), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for CloudFormation StackSet Instance (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func StackSetInstanceParseID(id string) (string, string, string, error) {
	idFormatErr := fmt.Errorf("unexpected format of ID (%s), expected NAME,ACCOUNT,REGION", id)

	parts := strings.SplitN(id, ",", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", idFormatErr
	}

	return parts[0], parts[1], parts[2], nil
}

func ListStackSetInstances(conn *cloudformation.CloudFormation, stackSetName string) ([]*cloudformation.StackInstanceSummary, error) {
	input := &cloudformation.ListStackInstancesInput{
		StackSetName: aws.String(stackSetName),
	}
	result := make([]*cloudformation.StackInstanceSummary, 0)

	for {
		output, err := conn.ListStackInstances(input)

		if err != nil {
			return result, err
		}

		result = append(result, output.Summaries...)

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return result, nil
}
