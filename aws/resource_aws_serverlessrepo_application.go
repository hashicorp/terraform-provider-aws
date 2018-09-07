package aws

import (
	"fmt"
	"log"
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
		//		Update: resourceAwsServerlessRepositoryApplicationUpdate,
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
			//			"semantic_version": {
			//				Type:     schema.TypeString,
			//				Optional: true,
			//				Computed: true,
			//			},
			//			"outputs": {
			//				Type:     schema.TypeMap,
			//				Computed: true,
			//			},
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
	if v, ok := d.GetOk("parameters"); ok {
		changeSetRequest.ParameterOverrides = expandServerlessRepositoryApplicationParameters(v.(map[string]interface{}))
	}

	log.Printf("[DEBUG] Creating Serverless Repo Application Change Set: %s", changeSetRequest)
	changeSetResponse, err := serverlessConn.CreateCloudFormationChangeSet(&changeSetRequest)
	if err != nil {
		return fmt.Errorf("Creating Serverless Repo Application Change Set failed: %s", err.Error())
	}

	d.SetId(*changeSetResponse.StackId)

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
			resp, err := cfConn.DescribeChangeSet(&cloudformation.DescribeChangeSetInput{
				ChangeSetName: changeSetResponse.ChangeSetId,
			})
			if err != nil {
				log.Printf("[ERROR] Failed to describe Change Set: %s", err)
				return nil, "", err
			}
			status := *resp.Status
			lastChangeSetStatus = status
			log.Printf("[DEBUG] Current CloudFormation stack status: %q", status)

			return resp, status, err
		},
	}
	_, err = changeSetWait.WaitForState()
	if err != nil {
		return err
	}

	if lastChangeSetStatus == "FAILED" {
		reasons, err := getCloudFormationFailures(d.Id(), cfConn)
		if err != nil {
			return fmt.Errorf("Failed getting failure reasons: %q", err.Error())
		}
		return fmt.Errorf("%s: %q", lastChangeSetStatus, reasons)
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
	conn := meta.(*AWSClient).cfconn

	input := &cloudformation.DescribeStacksInput{
		StackName: aws.String(d.Id()),
	}
	resp, err := conn.DescribeStacks(input)
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

	// Serverless Application Repo prefixes the stack with "serverlessrepo-", so remove it
	stackName := strings.TrimPrefix(*stack.StackName, "serverlessrepo-")
	d.Set("name", &stackName)

	originalParams := d.Get("parameters").(map[string]interface{})
	err = d.Set("parameters", flattenCloudFormationParameters(stack.Parameters, originalParams))
	if err != nil {
		return err
	}

	//	err = d.Set("tags", flattenCloudFormationTags(stack.Tags))
	//	if err != nil {
	//		return err
	//	}

	//	err = d.Set("outputs", flattenCloudFormationOutputs(stack.Outputs))
	//	if err != nil {
	//		return err
	//	}

	return nil
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
