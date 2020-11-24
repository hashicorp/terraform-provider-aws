package aws

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	serverlessrepository "github.com/aws/aws-sdk-go/service/serverlessapplicationrepository"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	cffinder "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudformation/finder"
	cfwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudformation/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/serverlessrepository/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/serverlessrepository/waiter"
)

const serverlessRepositoryStackNamePrefix = "serverlessrepo-"

func resourceAwsServerlessRepositoryStack() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServerlessRepositoryStackCreate,
		Read:   resourceAwsServerlessRepositoryStackRead,
		Update: resourceAwsServerlessRepositoryStackUpdate,
		Delete: resourceAwsServerlessRepositoryStackDelete,

		Importer: &schema.ResourceImporter{
			State: resourceAwsServerlessRepositoryStackImport,
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
	serverlessConn := meta.(*AWSClient).serverlessapplicationrepositoryconn
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
		return fmt.Errorf("error creating Serverless Application Repository change set (%s): %w", stackName, err)
	}

	d.SetId(aws.StringValue(changeSetResponse.StackId))

	_, err = cfwaiter.ChangeSetCreated(cfConn, d.Id(), aws.StringValue(changeSetResponse.ChangeSetId))
	if err != nil {
		return fmt.Errorf("error waiting for Serverless Application Repository change set (%s) creation: %w", stackName, err)
	}

	log.Printf("[INFO] Serverless Application Repository change set (%s) created", d.Id())

	requestToken := resource.UniqueId()
	executeRequest := cloudformation.ExecuteChangeSetInput{
		ChangeSetName:      changeSetResponse.ChangeSetId,
		ClientRequestToken: aws.String(requestToken),
	}
	log.Printf("[DEBUG] Executing Serverless Application Repository change set: %s", executeRequest)
	_, err = cfConn.ExecuteChangeSet(&executeRequest)
	if err != nil {
		return fmt.Errorf("Executing Serverless Application Repository change set failed: %w", err)
	}

	_, err = cfwaiter.StackCreated(cfConn, d.Id(), requestToken, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return fmt.Errorf("error waiting for Serverless Application Repository CloudFormation Stack (%s) creation: %w", d.Id(), err)
	}

	log.Printf("[INFO] Serverless Application Repository CloudFormation Stack (%s) created", d.Id())

	return resourceAwsServerlessRepositoryStackRead(d, meta)
}

func resourceAwsServerlessRepositoryStackRead(d *schema.ResourceData, meta interface{}) error {
	serverlessConn := meta.(*AWSClient).serverlessapplicationrepositoryconn
	cfConn := meta.(*AWSClient).cfconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	applicationID := d.Get("application_id").(string)
	semanticVersion := d.Get("semantic_version").(string)
	descriptor := applicationID
	if semanticVersion != "" {
		descriptor += fmt.Sprintf(", version %s", semanticVersion)
	}

	getApplicationOutput, err := finder.Application(serverlessConn, applicationID, semanticVersion)
	if err != nil {
		return fmt.Errorf("error getting Serverless Application Repository application (%s): %w", descriptor, err)
	}

	if getApplicationOutput.Version == nil {
		return fmt.Errorf("error getting Serverless Application Repository application (%s): returned empty version record", descriptor)
	}

	version := getApplicationOutput.Version
	d.Set("semantic_version", version.SemanticVersion)

	parameterDefinitions := flattenServerlessRepositoryParameterDefinitions(version.ParameterDefinitions)

	stack, err := cffinder.Stack(cfConn, d.Id())
	var e *resource.NotFoundError
	if errors.As(err, &e) {
		log.Printf("[WARN] CloudFormation stack (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error describing CloudFormation Stack (%s): %w", d.Id(), err)
	}

	// Serverless Application Repo prefixes the stack name with "serverlessrepo-",
	// so remove it from the saved string
	stackName := strings.TrimPrefix(aws.StringValue(stack.StackName), serverlessRepositoryStackNamePrefix)
	d.Set("name", &stackName)

	if err = d.Set("parameters", flattenSomeCloudFormationParameters(stack.Parameters, parameterDefinitions)); err != nil {
		return fmt.Errorf("failed to set parameters: %w", err)
	}

	if err = d.Set("tags", keyvaluetags.CloudformationKeyValueTags(stack.Tags).IgnoreServerlessApplicationRepository().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("failed to set tags: %w", err)
	}

	if err = d.Set("outputs", flattenCloudFormationOutputs(stack.Outputs)); err != nil {
		return fmt.Errorf("failed to set outputs: %w", err)
	}

	if err = d.Set("capabilities", flattenServerlessRepositoryStackCapabilities(stack.Capabilities, version.RequiredCapabilities)); err != nil {
		return fmt.Errorf("failed to set capabilities: %w", err)
	}

	return nil
}

func flattenSomeCloudFormationParameters(cfParams []*cloudformation.Parameter, parameterDefinitions map[string]*serverlessrepository.ParameterDefinition) map[string]interface{} {
	params := make(map[string]interface{}, len(cfParams))
	for _, p := range cfParams {
		key := aws.StringValue(p.ParameterKey)
		value := aws.StringValue(p.ParameterValue)
		if value != aws.StringValue(parameterDefinitions[key].DefaultValue) {
			params[key] = value
		}
	}
	return params
}

func flattenServerlessRepositoryParameterDefinitions(parameterDefinitions []*serverlessrepository.ParameterDefinition) map[string]*serverlessrepository.ParameterDefinition {
	result := make(map[string]*serverlessrepository.ParameterDefinition, len(parameterDefinitions))
	for _, p := range parameterDefinitions {
		result[aws.StringValue(p.Name)] = p
	}
	return result
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

	_, err = cfwaiter.ChangeSetCreated(conn, d.Id(), aws.StringValue(changeSetResponse.Id))
	if err != nil {
		return fmt.Errorf("error waiting for Serverless Application Repository change set (%s) creation: %w", d.Id(), err)
	}

	requestToken := resource.UniqueId()
	executeRequest := cloudformation.ExecuteChangeSetInput{
		ChangeSetName:      changeSetResponse.Id,
		ClientRequestToken: aws.String(requestToken),
	}
	log.Printf("[DEBUG] Executing Serverless Application Repository change set: %s", executeRequest)
	_, err = conn.ExecuteChangeSet(&executeRequest)
	if err != nil {
		return fmt.Errorf("Executing Serverless Application Repository change set failed: %w", err)
	}

	_, err = cfwaiter.StackUpdated(conn, d.Id(), requestToken, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return fmt.Errorf("error waiting for Serverless Application Repository CloudFormation Stack (%s) update: %w", d.Id(), err)
	}

	log.Printf("[INFO] Serverless Application Repository CloudFormation Stack (%s) updated", d.Id())

	return resourceAwsServerlessRepositoryStackRead(d, meta)
}

func resourceAwsServerlessRepositoryStackDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cfconn

	requestToken := resource.UniqueId()
	input := &cloudformation.DeleteStackInput{
		StackName:          aws.String(d.Id()),
		ClientRequestToken: aws.String(requestToken),
	}
	log.Printf("[DEBUG] Deleting Serverless Application Repository CloudFormation stack %s", input)
	_, err := conn.DeleteStack(input)
	if tfawserr.ErrCodeEquals(err, "ValidationError") {
		return nil
	}
	if err != nil {
		return err
	}

	_, err = cfwaiter.StackDeleted(conn, d.Id(), requestToken, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return fmt.Errorf("error waiting for Serverless Application Repository CloudFormation Stack deletion: %w", err)
	}

	log.Printf("[INFO] Serverless Application Repository CloudFormation stack (%s) deleted", d.Id())

	return nil
}

func resourceAwsServerlessRepositoryStackImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	stackID := d.Id()

	// If this isn't an ARN, it's the stack name
	if _, err := arn.Parse(stackID); err != nil {
		if !strings.HasPrefix(stackID, serverlessRepositoryStackNamePrefix) {
			stackID = serverlessRepositoryStackNamePrefix + stackID
		}
	}

	cfConn := meta.(*AWSClient).cfconn
	stack, err := cffinder.Stack(cfConn, stackID)
	if err != nil {
		return nil, fmt.Errorf("error describing Serverless Application Repository CloudFormation Stack (%s): %w", stackID, err)
	}

	tags := stack.Tags
	var applicationID, semanticVersion string
	for _, tag := range tags {
		if aws.StringValue(tag.Key) == "serverlessrepo:applicationId" {
			applicationID = aws.StringValue(tag.Value)
		} else if aws.StringValue(tag.Key) == "serverlessrepo:semanticVersion" {
			semanticVersion = aws.StringValue(tag.Value)
		}
		if applicationID != "" && semanticVersion != "" {
			break
		}
	}
	if applicationID == "" || semanticVersion == "" {
		return nil, fmt.Errorf("cannot import Serverless Application Repository CloudFormation Stack (%s): tags \"serverlessrepo:applicationId\" and \"serverlessrepo:semanticVersion\" must be defined", stackID)
	}

	d.SetId(aws.StringValue(stack.StackId))
	d.Set("application_id", applicationID)
	d.Set("semantic_version", semanticVersion)

	return []*schema.ResourceData{d}, nil
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

func flattenServerlessRepositoryStackCapabilities(stackCapabilities []*string, applicationRequiredCapabilities []*string) *schema.Set {
	// We need to preserve "CAPABILITY_RESOURCE_POLICY" if it has been set. It is not
	// returned by the CloudFormation APIs.
	capabilities := flattenStringSet(stackCapabilities)
	for _, capability := range applicationRequiredCapabilities {
		if aws.StringValue(capability) == serverlessrepository.CapabilityCapabilityResourcePolicy {
			capabilities.Add(serverlessrepository.CapabilityCapabilityResourcePolicy)
			break
		}
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
