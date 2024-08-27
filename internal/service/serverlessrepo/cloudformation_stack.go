// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package serverlessrepo

import ( // nosemgrep:ci.semgrep.aws.multiple-service-imports
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cloudformationtypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	serverlessrepo "github.com/aws/aws-sdk-go-v2/service/serverlessapplicationrepository"
	awstypes "github.com/aws/aws-sdk-go-v2/service/serverlessapplicationrepository/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	CloudFormationStackNamePrefix = "serverlessrepo-"

	cloudFormationStackTagApplicationID   = "serverlessrepo:applicationId"
	cloudFormationStackTagSemanticVersion = "serverlessrepo:semanticVersion"
)

// @SDKResource("aws_serverlessapplicationrepository_cloudformation_stack", name="CloudFormation Stack")
// @Tags
func ResourceCloudFormationStack() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCloudFormationStackCreate,
		ReadWithoutTimeout:   resourceCloudFormationStackRead,
		UpdateWithoutTimeout: resourceCloudFormationStackUpdate,
		DeleteWithoutTimeout: resourceCloudFormationStackDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceCloudFormationStackImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(cloudFormationStackCreatedDefaultTimeout),
			Update: schema.DefaultTimeout(cloudFormationStackUpdatedDefaultTimeout),
			Delete: schema.DefaultTimeout(cloudFormationStackDeletedDefaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			names.AttrApplicationID: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"capabilities": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.Capability](),
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"outputs": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrParameters: {
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCloudFormationStackCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cfConn := meta.(*conns.AWSClient).CloudFormationClient(ctx)

	changeSet, err := createCloudFormationChangeSet(ctx, d, meta.(*conns.AWSClient))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Serverless Application Repository CloudFormation change set: %s", err)
	}

	log.Printf("[INFO] Serverless Application Repository CloudFormation Stack (%s) change set created", d.Id())

	d.SetId(aws.ToString(changeSet.StackId))

	requestToken := id.UniqueId()
	executeRequest := cloudformation.ExecuteChangeSetInput{
		ChangeSetName:      changeSet.ChangeSetId,
		ClientRequestToken: aws.String(requestToken),
	}
	log.Printf("[DEBUG] Executing Serverless Application Repository CloudFormation change set: %+v", executeRequest)
	_, err = cfConn.ExecuteChangeSet(ctx, &executeRequest)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "executing Serverless Application Repository CloudFormation Stack (%s) change set failed: %s", d.Id(), err)
	}

	_, err = tfcloudformation.WaitStackCreated(ctx, cfConn, d.Id(), requestToken, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Serverless Application Repository CloudFormation Stack (%s) creation: %s", d.Id(), err)
	}

	log.Printf("[INFO] Serverless Application Repository CloudFormation Stack (%s) created", d.Id())

	return append(diags, resourceCloudFormationStackRead(ctx, d, meta)...)
}

func resourceCloudFormationStackRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	serverlessConn := meta.(*conns.AWSClient).ServerlessRepoClient(ctx)
	cfConn := meta.(*conns.AWSClient).CloudFormationClient(ctx)

	stack, err := tfcloudformation.FindStackByName(ctx, cfConn, d.Id())

	if tfresource.NotFound(err) {
		log.Printf("[WARN] Serverless Application Repository CloudFormation Stack (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Serverless Application Repository CloudFormation Stack (%s): %s", d.Id(), err)
	}

	// Serverless Application Repo prefixes the stack name with "serverlessrepo-", so remove it from the saved string
	stackName := strings.TrimPrefix(aws.ToString(stack.StackName), CloudFormationStackNamePrefix)
	d.Set(names.AttrName, &stackName)

	tags := tfcloudformation.KeyValueTags(ctx, stack.Tags)
	var applicationID, semanticVersion string
	if v, ok := tags[cloudFormationStackTagApplicationID]; ok {
		applicationID = aws.ToString(v.Value)
		d.Set(names.AttrApplicationID, applicationID)
	} else {
		return sdkdiag.AppendErrorf(diags, "describing Serverless Application Repository CloudFormation Stack (%s): missing required tag \"%s\"", d.Id(), cloudFormationStackTagApplicationID)
	}
	if v, ok := tags[cloudFormationStackTagSemanticVersion]; ok {
		semanticVersion = aws.ToString(v.Value)
		d.Set("semantic_version", semanticVersion)
	} else {
		return sdkdiag.AppendErrorf(diags, "describing Serverless Application Repository CloudFormation Stack (%s): missing required tag \"%s\"", d.Id(), cloudFormationStackTagSemanticVersion)
	}

	setTagsOut(ctx, Tags(tags))

	if err = d.Set("outputs", flattenCloudFormationOutputs(stack.Outputs)); err != nil {
		return sdkdiag.AppendErrorf(diags, "to set outputs: %s", err)
	}

	getApplicationOutput, err := findApplication(ctx, serverlessConn, applicationID, semanticVersion)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Serverless Application Repository application (%s, v%s): %s", applicationID, semanticVersion, err)
	}

	if getApplicationOutput == nil || getApplicationOutput.Version == nil {
		return sdkdiag.AppendErrorf(diags, "getting Serverless Application Repository application (%s, v%s): empty response", applicationID, semanticVersion)
	}

	version := getApplicationOutput.Version

	if err = d.Set(names.AttrParameters, flattenNonDefaultCloudFormationParameters(stack.Parameters, version.ParameterDefinitions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "to set parameters: %s", err)
	}

	if err = d.Set("capabilities", flattenStackCapabilities(stack.Capabilities, version.RequiredCapabilities)); err != nil {
		return sdkdiag.AppendErrorf(diags, "to set capabilities: %s", err)
	}

	return diags
}

func flattenNonDefaultCloudFormationParameters(cfParams []cloudformationtypes.Parameter, rawParameterDefinitions []awstypes.ParameterDefinition) map[string]interface{} {
	parameterDefinitions := flattenParameterDefinitions(rawParameterDefinitions)
	params := make(map[string]interface{}, len(cfParams))
	for _, p := range cfParams {
		key := aws.ToString(p.ParameterKey)
		value := aws.ToString(p.ParameterValue)
		if value != aws.ToString(parameterDefinitions[key].DefaultValue) {
			params[key] = value
		}
	}
	return params
}

func flattenParameterDefinitions(parameterDefinitions []awstypes.ParameterDefinition) map[string]awstypes.ParameterDefinition {
	result := make(map[string]awstypes.ParameterDefinition, len(parameterDefinitions))
	for _, p := range parameterDefinitions {
		result[aws.ToString(p.Name)] = p
	}
	return result
}

func resourceCloudFormationStackUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cfConn := meta.(*conns.AWSClient).CloudFormationClient(ctx)

	changeSet, err := createCloudFormationChangeSet(ctx, d, meta.(*conns.AWSClient))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Serverless Application Repository CloudFormation Stack (%s) change set: %s", d.Id(), err)
	}

	log.Printf("[INFO] Serverless Application Repository CloudFormation Stack (%s) change set created", d.Id())

	requestToken := id.UniqueId()
	executeRequest := cloudformation.ExecuteChangeSetInput{
		ChangeSetName:      changeSet.ChangeSetId,
		ClientRequestToken: aws.String(requestToken),
	}
	log.Printf("[DEBUG] Executing Serverless Application Repository CloudFormation change set: %+v", executeRequest)
	_, err = cfConn.ExecuteChangeSet(ctx, &executeRequest)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "executing Serverless Application Repository CloudFormation change set failed: %s", err)
	}

	_, err = tfcloudformation.WaitStackUpdated(ctx, cfConn, d.Id(), requestToken, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Serverless Application Repository CloudFormation Stack (%s) update: %s", d.Id(), err)
	}

	log.Printf("[INFO] Serverless Application Repository CloudFormation Stack (%s) updated", d.Id())

	return append(diags, resourceCloudFormationStackRead(ctx, d, meta)...)
}

func resourceCloudFormationStackDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cfConn := meta.(*conns.AWSClient).CloudFormationClient(ctx)

	requestToken := id.UniqueId()
	input := &cloudformation.DeleteStackInput{
		StackName:          aws.String(d.Id()),
		ClientRequestToken: aws.String(requestToken),
	}
	_, err := cfConn.DeleteStack(ctx, input)
	if tfawserr.ErrCodeEquals(err, "ValidationError") {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Serverless Application Repository CloudFormation Stack (%s): %s", d.Id(), err)
	}

	_, err = tfcloudformation.WaitStackDeleted(ctx, cfConn, d.Id(), requestToken, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Serverless Application Repository CloudFormation Stack (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}

func resourceCloudFormationStackImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	stackID := d.Id()

	// If this isn't an ARN, it's the stack name
	if _, err := arn.Parse(stackID); err != nil {
		if !strings.HasPrefix(stackID, CloudFormationStackNamePrefix) {
			stackID = CloudFormationStackNamePrefix + stackID
		}
	}

	cfConn := meta.(*conns.AWSClient).CloudFormationClient(ctx)
	stack, err := tfcloudformation.FindStackByName(ctx, cfConn, stackID)
	if err != nil {
		return nil, fmt.Errorf("describing Serverless Application Repository CloudFormation Stack (%s): %w", stackID, err)
	}

	d.SetId(aws.ToString(stack.StackId))

	return []*schema.ResourceData{d}, nil
}

func createCloudFormationChangeSet(ctx context.Context, d *schema.ResourceData, client *conns.AWSClient) (*cloudformation.DescribeChangeSetOutput, error) {
	serverlessConn := client.ServerlessRepoClient(ctx)
	cfConn := client.CloudFormationClient(ctx)

	stackName := d.Get(names.AttrName).(string)
	changeSetRequest := serverlessrepo.CreateCloudFormationChangeSetInput{
		StackName:     aws.String(stackName),
		ApplicationId: aws.String(d.Get(names.AttrApplicationID).(string)),
		Capabilities:  flex.ExpandStringValueSet(d.Get("capabilities").(*schema.Set)),
		Tags:          getTagsIn(ctx),
	}
	if v, ok := d.GetOk("semantic_version"); ok {
		changeSetRequest.SemanticVersion = aws.String(v.(string))
	}
	if v, ok := d.GetOk(names.AttrParameters); ok {
		changeSetRequest.ParameterOverrides = expandCloudFormationChangeSetParameters(v.(map[string]interface{}))
	}

	changeSetResponse, err := serverlessConn.CreateCloudFormationChangeSet(ctx, &changeSetRequest)
	if err != nil {
		return nil, err
	}

	return tfcloudformation.WaitChangeSetCreated(ctx, cfConn, aws.ToString(changeSetResponse.StackId), aws.ToString(changeSetResponse.ChangeSetId))
}

func expandCloudFormationChangeSetParameters(params map[string]interface{}) []awstypes.ParameterValue {
	var appParams []awstypes.ParameterValue
	for k, v := range params {
		appParams = append(appParams, awstypes.ParameterValue{
			Name:  aws.String(k),
			Value: aws.String(v.(string)),
		})
	}
	return appParams
}

func flattenStackCapabilities(stackCapabilities []cloudformationtypes.Capability, applicationRequiredCapabilities []awstypes.Capability) *schema.Set {
	// We need to preserve "CAPABILITY_RESOURCE_POLICY" if it has been set. It is not
	// returned by the CloudFormation APIs.
	capabilities := flex.FlattenStringyValueSet(stackCapabilities)
	for _, capability := range applicationRequiredCapabilities {
		if capability == awstypes.CapabilityCapabilityResourcePolicy {
			capabilities.Add(string(awstypes.CapabilityCapabilityResourcePolicy))
			break
		}
	}
	return capabilities
}
func flattenCloudFormationOutputs(cfOutputs []cloudformationtypes.Output) map[string]string {
	outputs := make(map[string]string, len(cfOutputs))
	for _, o := range cfOutputs {
		outputs[aws.ToString(o.OutputKey)] = aws.ToString(o.OutputValue)
	}
	return outputs
}
