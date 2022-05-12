package serverlessrepo

import ( // nosemgrep: aws-sdk-go-multiple-service-imports
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	serverlessrepo "github.com/aws/aws-sdk-go/service/serverlessapplicationrepository"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	CloudFormationStackNamePrefix = "serverlessrepo-"

	serverlessApplicationRepositoryCloudFormationStackTagApplicationID   = "serverlessrepo:applicationId"
	serverlessApplicationRepositoryCloudFormationStackTagSemanticVersion = "serverlessrepo:semanticVersion"
)

func ResourceCloudFormationStack() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudFormationStackCreate,
		Read:   resourceCloudFormationStackRead,
		Update: resourceCloudFormationStackUpdate,
		Delete: resourceCloudFormationStackDelete,

		Importer: &schema.ResourceImporter{
			State: resourceCloudFormationStackImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(cloudFormationStackCreatedDefaultTimeout),
			Update: schema.DefaultTimeout(cloudFormationStackUpdatedDefaultTimeout),
			Delete: schema.DefaultTimeout(cloudFormationStackDeletedDefaultTimeout),
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
				ValidateFunc: verify.ValidARN,
			},
			"capabilities": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(serverlessrepo.Capability_Values(), false),
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
			},
			"outputs": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCloudFormationStackCreate(d *schema.ResourceData, meta interface{}) error {
	cfConn := meta.(*conns.AWSClient).CloudFormationConn

	changeSet, err := createCloudFormationChangeSet(d, meta.(*conns.AWSClient))
	if err != nil {
		return fmt.Errorf("error creating Serverless Application Repository CloudFormation change set: %w", err)
	}

	log.Printf("[INFO] Serverless Application Repository CloudFormation Stack (%s) change set created", d.Id())

	d.SetId(aws.StringValue(changeSet.StackId))

	requestToken := resource.UniqueId()
	executeRequest := cloudformation.ExecuteChangeSetInput{
		ChangeSetName:      changeSet.ChangeSetId,
		ClientRequestToken: aws.String(requestToken),
	}
	log.Printf("[DEBUG] Executing Serverless Application Repository CloudFormation change set: %s", executeRequest)
	_, err = cfConn.ExecuteChangeSet(&executeRequest)
	if err != nil {
		return fmt.Errorf("executing Serverless Application Repository CloudFormation Stack (%s) change set failed: %w", d.Id(), err)
	}

	_, err = tfcloudformation.WaitStackCreated(cfConn, d.Id(), requestToken, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return fmt.Errorf("error waiting for Serverless Application Repository CloudFormation Stack (%s) creation: %w", d.Id(), err)
	}

	log.Printf("[INFO] Serverless Application Repository CloudFormation Stack (%s) created", d.Id())

	return resourceCloudFormationStackRead(d, meta)
}

func resourceCloudFormationStackRead(d *schema.ResourceData, meta interface{}) error {
	serverlessConn := meta.(*conns.AWSClient).ServerlessRepoConn
	cfConn := meta.(*conns.AWSClient).CloudFormationConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	stack, err := tfcloudformation.FindStackByID(cfConn, d.Id())

	if tfresource.NotFound(err) {
		log.Printf("[WARN] Serverless Application Repository CloudFormation Stack (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing Serverless Application Repository CloudFormation Stack (%s): %w", d.Id(), err)
	}

	// Serverless Application Repo prefixes the stack name with "serverlessrepo-", so remove it from the saved string
	stackName := strings.TrimPrefix(aws.StringValue(stack.StackName), CloudFormationStackNamePrefix)
	d.Set("name", &stackName)

	tags := tfcloudformation.KeyValueTags(stack.Tags)
	var applicationID, semanticVersion string
	if v, ok := tags[serverlessApplicationRepositoryCloudFormationStackTagApplicationID]; ok {
		applicationID = aws.StringValue(v.Value)
		d.Set("application_id", applicationID)
	} else {
		return fmt.Errorf("error describing Serverless Application Repository CloudFormation Stack (%s): missing required tag \"%s\"", d.Id(), serverlessApplicationRepositoryCloudFormationStackTagApplicationID)
	}
	if v, ok := tags[serverlessApplicationRepositoryCloudFormationStackTagSemanticVersion]; ok {
		semanticVersion = aws.StringValue(v.Value)
		d.Set("semantic_version", semanticVersion)
	} else {
		return fmt.Errorf("error describing Serverless Application Repository CloudFormation Stack (%s): missing required tag \"%s\"", d.Id(), serverlessApplicationRepositoryCloudFormationStackTagSemanticVersion)
	}

	tags = tags.IgnoreServerlessApplicationRepository().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err = d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("failed to set tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("failed to set tags_all: %w", err)
	}

	if err = d.Set("outputs", flattenCloudFormationOutputs(stack.Outputs)); err != nil {
		return fmt.Errorf("failed to set outputs: %w", err)
	}

	getApplicationOutput, err := findApplication(serverlessConn, applicationID, semanticVersion)
	if err != nil {
		return fmt.Errorf("error getting Serverless Application Repository application (%s, v%s): %w", applicationID, semanticVersion, err)
	}

	if getApplicationOutput == nil || getApplicationOutput.Version == nil {
		return fmt.Errorf("error getting Serverless Application Repository application (%s, v%s): empty response", applicationID, semanticVersion)
	}

	version := getApplicationOutput.Version

	if err = d.Set("parameters", flattenNonDefaultCloudFormationParameters(stack.Parameters, version.ParameterDefinitions)); err != nil {
		return fmt.Errorf("failed to set parameters: %w", err)
	}

	if err = d.Set("capabilities", flattenStackCapabilities(stack.Capabilities, version.RequiredCapabilities)); err != nil {
		return fmt.Errorf("failed to set capabilities: %w", err)
	}

	return nil
}

func flattenNonDefaultCloudFormationParameters(cfParams []*cloudformation.Parameter, rawParameterDefinitions []*serverlessrepo.ParameterDefinition) map[string]interface{} {
	parameterDefinitions := flattenParameterDefinitions(rawParameterDefinitions)
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

func flattenParameterDefinitions(parameterDefinitions []*serverlessrepo.ParameterDefinition) map[string]*serverlessrepo.ParameterDefinition {
	result := make(map[string]*serverlessrepo.ParameterDefinition, len(parameterDefinitions))
	for _, p := range parameterDefinitions {
		result[aws.StringValue(p.Name)] = p
	}
	return result
}

func resourceCloudFormationStackUpdate(d *schema.ResourceData, meta interface{}) error {
	cfConn := meta.(*conns.AWSClient).CloudFormationConn

	changeSet, err := createCloudFormationChangeSet(d, meta.(*conns.AWSClient))
	if err != nil {
		return fmt.Errorf("error creating Serverless Application Repository CloudFormation Stack (%s) change set: %w", d.Id(), err)
	}

	log.Printf("[INFO] Serverless Application Repository CloudFormation Stack (%s) change set created", d.Id())

	requestToken := resource.UniqueId()
	executeRequest := cloudformation.ExecuteChangeSetInput{
		ChangeSetName:      changeSet.ChangeSetId,
		ClientRequestToken: aws.String(requestToken),
	}
	log.Printf("[DEBUG] Executing Serverless Application Repository CloudFormation change set: %s", executeRequest)
	_, err = cfConn.ExecuteChangeSet(&executeRequest)
	if err != nil {
		return fmt.Errorf("executing Serverless Application Repository CloudFormation change set failed: %w", err)
	}

	_, err = tfcloudformation.WaitStackUpdated(cfConn, d.Id(), requestToken, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return fmt.Errorf("error waiting for Serverless Application Repository CloudFormation Stack (%s) update: %w", d.Id(), err)
	}

	log.Printf("[INFO] Serverless Application Repository CloudFormation Stack (%s) updated", d.Id())

	return resourceCloudFormationStackRead(d, meta)
}

func resourceCloudFormationStackDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFormationConn

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

	_, err = tfcloudformation.WaitStackDeleted(conn, d.Id(), requestToken, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return fmt.Errorf("error waiting for Serverless Application Repository CloudFormation Stack deletion: %w", err)
	}

	log.Printf("[INFO] Serverless Application Repository CloudFormation stack (%s) deleted", d.Id())

	return nil
}

func resourceCloudFormationStackImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	stackID := d.Id()

	// If this isn't an ARN, it's the stack name
	if _, err := arn.Parse(stackID); err != nil {
		if !strings.HasPrefix(stackID, CloudFormationStackNamePrefix) {
			stackID = CloudFormationStackNamePrefix + stackID
		}
	}

	cfConn := meta.(*conns.AWSClient).CloudFormationConn
	stack, err := tfcloudformation.FindStackByID(cfConn, stackID)
	if err != nil {
		return nil, fmt.Errorf("error describing Serverless Application Repository CloudFormation Stack (%s): %w", stackID, err)
	}

	d.SetId(aws.StringValue(stack.StackId))

	return []*schema.ResourceData{d}, nil
}

func createCloudFormationChangeSet(d *schema.ResourceData, client *conns.AWSClient) (*cloudformation.DescribeChangeSetOutput, error) {
	serverlessConn := client.ServerlessRepoConn
	cfConn := client.CloudFormationConn
	defaultTagsConfig := client.DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	stackName := d.Get("name").(string)
	changeSetRequest := serverlessrepo.CreateCloudFormationChangeSetRequest{
		StackName:     aws.String(stackName),
		ApplicationId: aws.String(d.Get("application_id").(string)),
		Capabilities:  flex.ExpandStringSet(d.Get("capabilities").(*schema.Set)),
		Tags:          Tags(tags.IgnoreServerlessApplicationRepository()),
	}
	if v, ok := d.GetOk("semantic_version"); ok {
		changeSetRequest.SemanticVersion = aws.String(v.(string))
	}
	if v, ok := d.GetOk("parameters"); ok {
		changeSetRequest.ParameterOverrides = expandCloudFormationChangeSetParameters(v.(map[string]interface{}))
	}

	log.Printf("[DEBUG] Creating Serverless Application Repository CloudFormation change set: %s", changeSetRequest)
	changeSetResponse, err := serverlessConn.CreateCloudFormationChangeSet(&changeSetRequest)
	if err != nil {
		return nil, err
	}

	return tfcloudformation.WaitChangeSetCreated(cfConn, aws.StringValue(changeSetResponse.StackId), aws.StringValue(changeSetResponse.ChangeSetId))
}

func expandCloudFormationChangeSetParameters(params map[string]interface{}) []*serverlessrepo.ParameterValue {
	var appParams []*serverlessrepo.ParameterValue
	for k, v := range params {
		appParams = append(appParams, &serverlessrepo.ParameterValue{
			Name:  aws.String(k),
			Value: aws.String(v.(string)),
		})
	}
	return appParams
}

func flattenStackCapabilities(stackCapabilities []*string, applicationRequiredCapabilities []*string) *schema.Set {
	// We need to preserve "CAPABILITY_RESOURCE_POLICY" if it has been set. It is not
	// returned by the CloudFormation APIs.
	capabilities := flex.FlattenStringSet(stackCapabilities)
	for _, capability := range applicationRequiredCapabilities {
		if aws.StringValue(capability) == serverlessrepo.CapabilityCapabilityResourcePolicy {
			capabilities.Add(serverlessrepo.CapabilityCapabilityResourcePolicy)
			break
		}
	}
	return capabilities
}
func flattenCloudFormationOutputs(cfOutputs []*cloudformation.Output) map[string]string {
	outputs := make(map[string]string, len(cfOutputs))
	for _, o := range cfOutputs {
		outputs[aws.StringValue(o.OutputKey)] = aws.StringValue(o.OutputValue)
	}
	return outputs
}
