package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	serverlessrepository "github.com/aws/aws-sdk-go/service/serverlessapplicationrepository"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/serverlessrepository/waiter"
)

func resourceAwsServerlessRepositoryStack() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServerlessRepositoryStackCreate,
		Read:   resourceAwsServerlessRepositoryStackRead,
		Update: resourceAwsServerlessRepositoryStackUpdate,
		Delete: resourceAwsServerlessRepositoryStackDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(waiter.StackCreatedDefaultTimeout),
			Update: schema.DefaultTimeout(waiter.StackUpdatedDefaultTimeout),
			Delete: schema.DefaultTimeout(waiter.StackDeletedDefaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"application_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"capabilities": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(serverlessrepository.Capability_Values(), false),
				},
				Set: schema.HashString,
			},
			"parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"semantic_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"outputs": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsServerlessRepositoryStackCreate(d *schema.ResourceData, meta interface{}) error {
	serverlessConn := meta.(*AWSClient).serverlessapprepositoryconn
	cfConn := meta.(*AWSClient).cfconn

	stackName := d.Get("name").(string)
	changeSetRequest := serverlessrepository.CreateCloudFormationChangeSetRequest{
		StackName:     aws.String(stackName),
		ApplicationId: aws.String(d.Get("application_id").(string)),
		Capabilities:  expandStringSet(d.Get("capabilities").(*schema.Set)),
		Tags:          keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreServerlessApplicationRepository().ServerlessapplicationrepositoryTags(),
	}
	if v, ok := d.GetOk("semantic_version"); ok {
		version := v.(string)
		changeSetRequest.SemanticVersion = aws.String(version)
	}
	if v, ok := d.GetOk("parameters"); ok {
		changeSetRequest.ParameterOverrides = expandServerlessRepositoryChangeSetParameters(v.(map[string]interface{}))
	}

	log.Printf("[DEBUG] Creating Serverless Application Repository change set: %s", changeSetRequest)
	changeSetResponse, err := serverlessConn.CreateCloudFormationChangeSet(&changeSetRequest)
	if err != nil {
		return fmt.Errorf("Creating Serverless Application Repository change set (%s) failed: %w", stackName, err)
	}

	d.SetId(aws.StringValue(changeSetResponse.StackId))

	err = waitForCreateChangeSet(d, cfConn, changeSetResponse.ChangeSetId)
	if err != nil {
		return err
	}

	executeRequest := cloudformation.ExecuteChangeSetInput{
		ChangeSetName: changeSetResponse.ChangeSetId,
	}
	log.Printf("[DEBUG] Executing Change Set: %s", executeRequest)
	_, err = cfConn.ExecuteChangeSet(&executeRequest)
	if err != nil {
		return fmt.Errorf("Executing Change Set failed: %w", err)
	}
	var lastStatus string

	wait := resource.StateChangeConf{
		Pending: []string{
			"CREATE_IN_PROGRESS",
			"DELETE_IN_PROGRESS",
			"ROLLBACK_IN_PROGRESS",
		},
		Target: []string{
			"CREATE_COMPLETE",
			"CREATE_FAILED",
			"DELETE_COMPLETE",
			"DELETE_FAILED",
			"ROLLBACK_COMPLETE",
			"ROLLBACK_FAILED",
		},
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			resp, err := cfConn.DescribeStacks(&cloudformation.DescribeStacksInput{
				StackName: aws.String(d.Id()),
			})
			if err != nil {
				log.Printf("[ERROR] Failed to describe stacks: %s", err)
				return nil, "", err
			}
			if len(resp.Stacks) == 0 {
				// This shouldn't happen unless CloudFormation is inconsistent
				// See https://github.com/hashicorp/terraform/issues/5487
				log.Printf("[WARN] CloudFormation stack %q not found.\nresponse: %q",
					d.Id(), resp)
				return resp, "", fmt.Errorf(
					"CloudFormation stack %q vanished unexpectedly during creation.\n"+
						"Unless you knowingly manually deleted the stack "+
						"please report this as bug at https://github.com/hashicorp/terraform/issues\n"+
						"along with the config & Terraform version & the details below:\n"+
						"Full API response: %s\n",
					d.Id(), resp)
			}

			status := *resp.Stacks[0].StackStatus
			lastStatus = status
			log.Printf("[DEBUG] Current CloudFormation stack status: %q", status)

			return resp, status, err
		},
	}

	_, err = wait.WaitForState()
	if err != nil {
		return err
	}

	if lastStatus == "ROLLBACK_COMPLETE" || lastStatus == "ROLLBACK_FAILED" {
		reasons, err := getCloudFormationRollbackReasons(d.Id(), nil, cfConn)
		if err != nil {
			return fmt.Errorf("Failed getting rollback reasons: %w", err)
		}

		return fmt.Errorf("Error creating %s: %q", lastStatus, reasons)
	}
	if lastStatus == "DELETE_COMPLETE" || lastStatus == "DELETE_FAILED" {
		reasons, err := getCloudFormationDeletionReasons(d.Id(), cfConn)
		if err != nil {
			return fmt.Errorf("Failed getting deletion reasons: %w", err)
		}

		d.SetId("")
		return fmt.Errorf("Error creating %s: %q", lastStatus, reasons)
	}
	if lastStatus == "CREATE_FAILED" {
		reasons, err := getCloudFormationFailures(d.Id(), cfConn)
		if err != nil {
			return fmt.Errorf("Failed getting failure reasons: %w", err)
		}
		return fmt.Errorf("Error creating %s: %q", lastStatus, reasons)
	}

	log.Printf("[INFO] CloudFormation Stack %q created", d.Id())

	return resourceAwsServerlessRepositoryStackRead(d, meta)
}

func resourceAwsServerlessRepositoryStackRead(d *schema.ResourceData, meta interface{}) error {
	serverlessConn := meta.(*AWSClient).serverlessapprepositoryconn
	cfConn := meta.(*AWSClient).cfconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	getApplicationInput := &serverlessrepository.GetApplicationInput{
		ApplicationId: aws.String(d.Get("application_id").(string)),
	}

	_, ok := d.GetOk("semantic_version")
	if !ok {
		getApplicationOutput, err := serverlessConn.GetApplication(getApplicationInput)
		if err != nil {
			return err
		}
		d.Set("semantic_version", getApplicationOutput.Version.SemanticVersion)
	}

	input := &cloudformation.DescribeStacksInput{
		StackName: aws.String(d.Id()),
	}
	resp, err := cfConn.DescribeStacks(input)
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		// ValidationError: Stack with id % does not exist
		if ok && awsErr.Code() == "ValidationError" {
			log.Printf("[WARN] Removing CloudFormation stack %s as it's already gone", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	stacks := resp.Stacks
	if len(stacks) < 1 {
		log.Printf("[WARN] Removing CloudFormation stack %s as it's already gone", d.Id())
		d.SetId("")
		return nil
	}
	for _, s := range stacks {
		if *s.StackId == d.Id() && *s.StackStatus == "DELETE_COMPLETE" {
			log.Printf("[DEBUG] Removing CloudFormation stack %s as it has been already deleted", d.Id())
			d.SetId("")
			return nil
		}
	}

	stack := stacks[0]
	log.Printf("[DEBUG] Received CloudFormation stack: %s", stack)

	// Serverless Application Repo prefixes the stack name with "serverlessrepo-",
	// so remove it from the saved string
	// FIXME: this should be a StateFunc
	stackName := strings.TrimPrefix(aws.StringValue(stack.StackName), "serverlessrepo-")
	d.Set("name", &stackName)

	originalParams := d.Get("parameters").(map[string]interface{})
	if err = d.Set("parameters", flattenCloudFormationParameters(stack.Parameters, originalParams)); err != nil {
		return fmt.Errorf("failed to set parameters: %w", err)
	}

	if err = d.Set("tags", keyvaluetags.CloudformationKeyValueTags(stack.Tags).IgnoreServerlessApplicationRepository().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("failed to set tags: %w", err)
	}

	if err = d.Set("outputs", flattenCloudFormationOutputs(stack.Outputs)); err != nil {
		return fmt.Errorf("failed to set outputs: %w", err)
	}

	if err = d.Set("capabilities", flattenServerlessRepositoryStackCapabilities(d, stack.Capabilities)); err != nil {
		return fmt.Errorf("failed to set capabilities: %w", err)
	}

	return nil
}

func resourceAwsServerlessRepositoryStackUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cfconn

	input := &cloudformation.CreateChangeSetInput{
		StackName:           aws.String(d.Id()),
		UsePreviousTemplate: aws.Bool(true),
		ChangeSetType:       aws.String("UPDATE"),
	}

	input.ChangeSetName = aws.String(resource.PrefixedUniqueId(d.Get("name").(string)))

	// Parameters must be present whether they are changed or not
	if v, ok := d.GetOk("parameters"); ok {
		input.Parameters = expandCloudFormationParameters(v.(map[string]interface{}))
	}

	if d.HasChange("tags") {
		if v, ok := d.GetOk("tags"); ok {
			input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreServerlessApplicationRepository().CloudformationTags()
		}
	}

	input.Capabilities = expandServerlessRepositoryStackChangeSetCapabilities(d.Get("capabilities").(*schema.Set))

	log.Printf("[DEBUG] Creating CloudFormation change set: %s", input)
	changeSetResponse, err := conn.CreateChangeSet(input)
	if err != nil {
		return fmt.Errorf("Creating CloudFormation change set failed: %w", err)
	}

	err = waitForCreateChangeSet(d, conn, changeSetResponse.Id)
	if err != nil {
		return err
	}

	executeRequest := cloudformation.ExecuteChangeSetInput{
		ChangeSetName: changeSetResponse.Id,
	}
	log.Printf("[DEBUG] Executing Change Set: %s", executeRequest)
	_, err = conn.ExecuteChangeSet(&executeRequest)
	if err != nil {
		return fmt.Errorf("Executing Change Set failed: %w", err)
	}

	lastUpdatedTime, err := getLastCfEventTimestamp(d.Id(), conn)
	if err != nil {
		return err
	}

	var lastStatus string
	var stackID string
	wait := resource.StateChangeConf{
		Pending: []string{
			"UPDATE_COMPLETE_CLEANUP_IN_PROGRESS",
			"UPDATE_IN_PROGRESS",
			"UPDATE_ROLLBACK_IN_PROGRESS",
			"UPDATE_ROLLBACK_COMPLETE_CLEANUP_IN_PROGRESS",
		},
		Target: []string{
			"CREATE_COMPLETE", // If no stack update was performed
			"UPDATE_COMPLETE",
			"UPDATE_ROLLBACK_COMPLETE",
			"UPDATE_ROLLBACK_FAILED",
		},
		Timeout:    d.Timeout(schema.TimeoutUpdate),
		MinTimeout: 5 * time.Second,
		Refresh: func() (interface{}, string, error) {
			resp, err := conn.DescribeStacks(&cloudformation.DescribeStacksInput{
				StackName: aws.String(d.Id()),
			})
			if err != nil {
				log.Printf("[ERROR] Failed to describe stacks: %s", err)
				return nil, "", err
			}

			stackID = aws.StringValue(resp.Stacks[0].StackId)

			status := *resp.Stacks[0].StackStatus
			lastStatus = status
			log.Printf("[DEBUG] Current CloudFormation stack status: %q", status)

			return resp, status, err
		},
	}

	_, err = wait.WaitForState()
	if err != nil {
		return err
	}

	if lastStatus == "UPDATE_ROLLBACK_COMPLETE" || lastStatus == "UPDATE_ROLLBACK_FAILED" {
		reasons, err := getCloudFormationRollbackReasons(stackID, lastUpdatedTime, conn)
		if err != nil {
			return fmt.Errorf("Failed getting details about rollback: %w", err)
		}

		return fmt.Errorf("Error updating %s: %q", lastStatus, reasons)
	}

	log.Printf("[DEBUG] CloudFormation stack %q has been updated", stackID)

	return resourceAwsServerlessRepositoryStackRead(d, meta)
}

func resourceAwsServerlessRepositoryStackDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cfconn

	input := &cloudformation.DeleteStackInput{
		StackName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Deleting CloudFormation stack %s", input)
	_, err := conn.DeleteStack(input)
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if !ok {
			return err
		}

		if awsErr.Code() == "ValidationError" {
			// Ignore stack which has been already deleted
			return nil
		}
		return err
	}
	var lastStatus string
	wait := resource.StateChangeConf{
		Pending: []string{
			"DELETE_IN_PROGRESS",
			"ROLLBACK_IN_PROGRESS",
		},
		Target: []string{
			"DELETE_COMPLETE",
			"DELETE_FAILED",
		},
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 5 * time.Second,
		Refresh: func() (interface{}, string, error) {
			resp, err := conn.DescribeStacks(&cloudformation.DescribeStacksInput{
				StackName: aws.String(d.Id()),
			})
			if err != nil {
				awsErr, ok := err.(awserr.Error)
				if !ok {
					return nil, "", err
				}

				log.Printf("[DEBUG] Error when deleting CloudFormation stack: %s: %s",
					awsErr.Code(), awsErr.Message())

				// ValidationError: Stack with id % does not exist
				if awsErr.Code() == "ValidationError" {
					return resp, "DELETE_COMPLETE", nil
				}
				return nil, "", err
			}

			if len(resp.Stacks) == 0 {
				log.Printf("[DEBUG] CloudFormation stack %q is already gone", d.Id())
				return resp, "DELETE_COMPLETE", nil
			}

			status := *resp.Stacks[0].StackStatus
			lastStatus = status
			log.Printf("[DEBUG] Current CloudFormation stack status: %q", status)

			return resp, status, err
		},
	}

	_, err = wait.WaitForState()
	if err != nil {
		return err
	}

	if lastStatus == "DELETE_FAILED" {
		reasons, err := getCloudFormationFailures(d.Id(), conn)
		if err != nil {
			return fmt.Errorf("Failed getting reasons of failure: %w", err)
		}

		return fmt.Errorf("Error deleting %s: %q", lastStatus, reasons)
	}

	log.Printf("[DEBUG] CloudFormation stack %q has been deleted", d.Id())

	return nil
}

func waitForCreateChangeSet(d *schema.ResourceData, conn *cloudformation.CloudFormation, changeSetName *string) error {
	var lastChangeSetStatus string
	changeSetWait := resource.StateChangeConf{
		Pending: []string{
			"CREATE_PENDING",
			"CREATE_IN_PROGRESS",
		},
		Target: []string{
			"CREATE_COMPLETE",
			"DELETE_COMPLETE",
			"FAILED",
		},
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			resp, err := conn.DescribeChangeSet(&cloudformation.DescribeChangeSetInput{
				ChangeSetName: changeSetName,
			})
			if err != nil {
				log.Printf("[ERROR] Failed to describe change set: %s", err)
				return nil, "", err
			}
			status := *resp.Status
			lastChangeSetStatus = status
			log.Printf("[DEBUG] Current CloudFormation stack status: %q", status)

			return resp, status, err
		},
	}
	_, err := changeSetWait.WaitForState()

	if lastChangeSetStatus == "FAILED" {
		reasons, err := getCloudFormationFailures(d.Id(), conn)
		if err != nil {
			return fmt.Errorf("Failed getting failure reasons: %w", err)
		}
		return fmt.Errorf("Error waiting %s: %q", lastChangeSetStatus, reasons)
	}
	return err
}

func expandServerlessRepositoryChangeSetParameters(params map[string]interface{}) []*serverlessrepository.ParameterValue {
	var appParams []*serverlessrepository.ParameterValue
	for k, v := range params {
		appParams = append(appParams, &serverlessrepository.ParameterValue{
			Name:  aws.String(k),
			Value: aws.String(v.(string)),
		})
	}
	return appParams
}

func flattenServerlessRepositoryStackCapabilities(d *schema.ResourceData, c []*string) *schema.Set {
	// We need to preserve "CAPABILITY_RESOURCE_POLICY" if it has been set. It is not
	// returned by the CloudFormation APIs.
	existingCapabilities := d.Get("capabilities").(*schema.Set)
	capabilities := flattenStringSet(c)
	if existingCapabilities.Contains(serverlessrepository.CapabilityCapabilityResourcePolicy) {
		capabilities.Add(serverlessrepository.CapabilityCapabilityResourcePolicy)
	}
	return capabilities
}

func expandServerlessRepositoryStackChangeSetCapabilities(capabilities *schema.Set) []*string {
	// Filter the capabilities for the CloudFormation Change Set. CloudFormation supports a
	// subset of the capabilities supported by Serverless Application Repository.
	result := make([]*string, 0, capabilities.Len())
	for _, c := range cloudformation.Capability_Values() {
		if capabilities.Contains(c) {
			result = append(result, aws.String(c))
		}
	}
	return result
}
