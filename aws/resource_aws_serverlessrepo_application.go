package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/serverlessapplicationrepository"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsServerlessRepositoryApplication() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServerlessRepositoryApplicationCreate,
		Read:   resourceAwsServerlessRepositoryApplicationRead,
		Update: resourceAwsServerlessRepositoryApplicationUpdate,
		Delete: resourceAwsServerlessRepositoryApplicationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
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
			"parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
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
			},
			"capabilities": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsServerlessRepositoryApplicationCreate(d *schema.ResourceData, meta interface{}) error {
	serverlessConn := meta.(*AWSClient).serverlessapplicationrepositoryconn
	cfConn := meta.(*AWSClient).cfconn

	changeSetRequest := serverlessapplicationrepository.CreateCloudFormationChangeSetRequest{
		StackName:     aws.String(d.Get("name").(string)),
		ApplicationId: aws.String(d.Get("application_id").(string)),
	}
	if v, ok := d.GetOk("semantic_version"); ok {
		version := v.(string)
		changeSetRequest.SemanticVersion = aws.String(version)
	}
	if v, ok := d.GetOk("parameters"); ok {
		changeSetRequest.ParameterOverrides = expandServerlessRepositoryApplicationParameters(v.(map[string]interface{}))
	}
	if v, ok := d.GetOk("tags"); ok {
		changeSetRequest.Tags = expandServerlessRepositoryTags(v.(map[string]interface{}))
	}

	log.Printf("[DEBUG] Creating Serverless Repo Application change set: %s", changeSetRequest)
	changeSetResponse, err := serverlessConn.CreateCloudFormationChangeSet(&changeSetRequest)
	if err != nil {
		return fmt.Errorf("Creating Serverless Repo Application change set failed: %s", err.Error())
	}

	d.SetId(*changeSetResponse.StackId)

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
		return fmt.Errorf("Executing Change Set failed: %s", err.Error())
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
			return fmt.Errorf("Failed getting rollback reasons: %q", err.Error())
		}

		return fmt.Errorf("%s: %q", lastStatus, reasons)
	}
	if lastStatus == "DELETE_COMPLETE" || lastStatus == "DELETE_FAILED" {
		reasons, err := getCloudFormationDeletionReasons(d.Id(), cfConn)
		if err != nil {
			return fmt.Errorf("Failed getting deletion reasons: %q", err.Error())
		}

		d.SetId("")
		return fmt.Errorf("%s: %q", lastStatus, reasons)
	}
	if lastStatus == "CREATE_FAILED" {
		reasons, err := getCloudFormationFailures(d.Id(), cfConn)
		if err != nil {
			return fmt.Errorf("Failed getting failure reasons: %q", err.Error())
		}
		return fmt.Errorf("%s: %q", lastStatus, reasons)
	}

	log.Printf("[INFO] CloudFormation Stack %q created", d.Id())

	return resourceAwsServerlessRepositoryApplicationRead(d, meta)
}

func resourceAwsServerlessRepositoryApplicationRead(d *schema.ResourceData, meta interface{}) error {
	serverlessConn := meta.(*AWSClient).serverlessapplicationrepositoryconn
	cfConn := meta.(*AWSClient).cfconn

	getApplicationInput := &serverlessapplicationrepository.GetApplicationInput{
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
			log.Printf("[DEBUG] Removing CloudFormation stack %s"+
				" as it has been already deleted", d.Id())
			d.SetId("")
			return nil
		}
	}

	stack := stacks[0]
	log.Printf("[DEBUG] Received CloudFormation stack: %s", stack)

	// Serverless Application Repo prefixes the stack with "serverlessrepo-",
	// so remove it from the saved string
	stackName := strings.TrimPrefix(*stack.StackName, "serverlessrepo-")
	d.Set("name", &stackName)

	originalParams := d.Get("parameters").(map[string]interface{})
	err = d.Set("parameters", flattenCloudFormationParameters(stack.Parameters, originalParams))
	if err != nil {
		return err
	}

	err = d.Set("tags", flattenServerlessRepositoryCloudFormationTags(stack.Tags))
	if err != nil {
		return err
	}

	err = d.Set("outputs", flattenCloudFormationOutputs(stack.Outputs))
	if err != nil {
		return err
	}

	if len(stack.Capabilities) > 0 {
		err = d.Set("capabilities", schema.NewSet(schema.HashString, flattenStringList(stack.Capabilities)))
		if err != nil {
			return err
		}
	}

	return nil
}

func expandServerlessRepositoryTags(tags map[string]interface{}) []*serverlessapplicationrepository.Tag {
	var cfTags []*serverlessapplicationrepository.Tag
	for k, v := range tags {
		cfTags = append(cfTags, &serverlessapplicationrepository.Tag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		})
	}
	return cfTags
}

func flattenServerlessRepositoryCloudFormationTags(cfTags []*cloudformation.Tag) map[string]string {
	tags := make(map[string]string, len(cfTags))
	for _, t := range cfTags {
		if !tagIgnoredServerlessRepositoryCloudFormation(*t.Key) {
			tags[*t.Key] = *t.Value
		}
	}
	return tags
}

func tagIgnoredServerlessRepositoryCloudFormation(k string) bool {
	filter := []string{"^aws:", "^serverlessrepo:"}
	for _, v := range filter {
		log.Printf("[DEBUG] Matching %v with %v\n", v, k)
		if r, _ := regexp.MatchString(v, k); r {
			log.Printf("[DEBUG] Found AWS specific tag %s, ignoring.\n", k)
			return true
		}
	}
	return false
}

func resourceAwsServerlessRepositoryApplicationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cfconn

	input := &cloudformation.CreateChangeSetInput{
		StackName:           aws.String(d.Id()),
		UsePreviousTemplate: aws.Bool(true),
		ChangeSetType:       aws.String("UPDATE"),
	}

	input.ChangeSetName = aws.String(fmt.Sprintf("%s-%s",
		d.Get("name").(string),
		time.Now().UTC().Format("20060102150405999999999")))

	// Parameters must be present whether they are changed or not
	if v, ok := d.GetOk("parameters"); ok {
		input.Parameters = expandCloudFormationParameters(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("tags"); ok {
		input.Tags = expandCloudFormationTags(v.(map[string]interface{}))
	}

	capabilities := d.Get("capabilities")
	input.Capabilities = expandStringList(capabilities.(*schema.Set).List())

	log.Printf("[DEBUG] Creating CloudFormation change set: %s", input)
	changeSetResponse, err := conn.CreateChangeSet(input)
	if err != nil {
		return fmt.Errorf("Creating CloudFormation change set failed: %s", err.Error())
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
		return fmt.Errorf("Executing Change Set failed: %s", err.Error())
	}

	lastUpdatedTime, err := getLastCfEventTimestamp(d.Id(), conn)
	if err != nil {
		return err
	}

	var lastStatus string
	var stackId string
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

			stackId = aws.StringValue(resp.Stacks[0].StackId)

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
		reasons, err := getCloudFormationRollbackReasons(stackId, lastUpdatedTime, conn)
		if err != nil {
			return fmt.Errorf("Failed getting details about rollback: %q", err.Error())
		}

		return fmt.Errorf("%s: %q", lastStatus, reasons)
	}

	log.Printf("[DEBUG] CloudFormation stack %q has been updated", stackId)

	return resourceAwsServerlessRepositoryApplicationRead(d, meta)
}

func resourceAwsServerlessRepositoryApplicationDelete(d *schema.ResourceData, meta interface{}) error {
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
			return fmt.Errorf("Failed getting reasons of failure: %q", err.Error())
		}

		return fmt.Errorf("%s: %q", lastStatus, reasons)
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
			return fmt.Errorf("Failed getting failure reasons: %q", err.Error())
		}
		return fmt.Errorf("%s: %q", lastChangeSetStatus, reasons)
	}
	return err
}

// Move to `structure.go`?
func expandServerlessRepositoryApplicationParameters(params map[string]interface{}) []*serverlessapplicationrepository.ParameterValue {
	var appParams []*serverlessapplicationrepository.ParameterValue
	for k, v := range params {
		appParams = append(appParams, &serverlessapplicationrepository.ParameterValue{
			Name:  aws.String(k),
			Value: aws.String(v.(string)),
		})
	}
	return appParams
}
