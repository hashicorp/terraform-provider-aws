package transfer

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceWorkflow() *schema.Resource {
	return &schema.Resource{
		Create: resourceWorkflowCreate,
		Read:   resourceWorkflowRead,
		Update: resourceWorkflowUpdate,
		Delete: resourceWorkflowDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"on_exception_steps": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 8,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"copy_step_details": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"destination_file_location": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"efs_file_location": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"file_system_id": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"path": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(1, 65536),
															},
														},
													},
												},
												"s3_file_location": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"bucket": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"key": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 1024),
															},
														},
													},
												},
											},
										},
									},
									"name": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 30),
											validation.StringMatch(regexp.MustCompile(`^[\w-]*$`), "Must be of the pattern ^[\\w-]*$"),
										),
									},
									"overwrite_existing": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      transfer.OverwriteExistingFalse,
										ValidateFunc: validation.StringInSlice(transfer.OverwriteExisting_Values(), false),
									},
									"source_file_location": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 256),
											validation.StringMatch(regexp.MustCompile(`^\$\{(\w+.)+\w+\}$`), "Must be of the pattern ^\\$\\{(\\w+.)+\\w+\\}$"),
										),
									},
								},
							},
						},
						"custom_step_details": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 30),
											validation.StringMatch(regexp.MustCompile(`^[\w-]*$`), "Must be of the pattern ^[\\w-]*$"),
										),
									},
									"source_file_location": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 256),
											validation.StringMatch(regexp.MustCompile(`^\$\{(\w+.)+\w+\}$`), "Must be of the pattern ^\\$\\{(\\w+.)+\\w+\\}$"),
										),
									},
									"target": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
									"timeout_seconds": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 1800),
									},
								},
							},
						},
						"delete_step_details": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 30),
											validation.StringMatch(regexp.MustCompile(`^[\w-]*$`), "Must be of the pattern ^[\\w-]*$"),
										),
									},
									"source_file_location": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 256),
											validation.StringMatch(regexp.MustCompile(`^\$\{(\w+.)+\w+\}$`), "Must be of the pattern ^\\$\\{(\\w+.)+\\w+\\}$"),
										),
									},
								},
							},
						},
						"tag_step_details": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 30),
											validation.StringMatch(regexp.MustCompile(`^[\w-]*$`), "Must be of the pattern ^[\\w-]*$"),
										),
									},
									"source_file_location": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 256),
											validation.StringMatch(regexp.MustCompile(`^\$\{(\w+.)+\w+\}$`), "Must be of the pattern ^\\$\\{(\\w+.)+\\w+\\}$"),
										),
									},
									"tags": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 10,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(0, 128),
												},
												"value": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(0, 256),
												},
											},
										},
									},
								},
							},
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(transfer.WorkflowStepType_Values(), false),
						},
					},
				},
			},
			"steps": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 8,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"copy_step_details": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"destination_file_location": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"efs_file_location": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"file_system_id": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"path": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(1, 65536),
															},
														},
													},
												},
												"s3_file_location": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"bucket": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"key": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 1024),
															},
														},
													},
												},
											},
										},
									},
									"name": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 30),
											validation.StringMatch(regexp.MustCompile(`^[\w-]*$`), "Must be of the pattern ^[\\w-]*$"),
										),
									},
									"overwrite_existing": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      transfer.OverwriteExistingFalse,
										ValidateFunc: validation.StringInSlice(transfer.OverwriteExisting_Values(), false),
									},
									"source_file_location": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 256),
											validation.StringMatch(regexp.MustCompile(`^\$\{(\w+.)+\w+\}$`), "Must be of the pattern ^\\$\\{(\\w+.)+\\w+\\}$"),
										),
									},
								},
							},
						},
						"custom_step_details": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 30),
											validation.StringMatch(regexp.MustCompile(`^[\w-]*$`), "Must be of the pattern ^[\\w-]*$"),
										),
									},
									"source_file_location": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 256),
											validation.StringMatch(regexp.MustCompile(`^\$\{(\w+.)+\w+\}$`), "Must be of the pattern ^\\$\\{(\\w+.)+\\w+\\}$"),
										),
									},
									"target": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
									"timeout_seconds": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 1800),
									},
								},
							},
						},
						"delete_step_details": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 30),
											validation.StringMatch(regexp.MustCompile(`^[\w-]*$`), "Must be of the pattern ^[\\w-]*$"),
										),
									},
									"source_file_location": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 256),
											validation.StringMatch(regexp.MustCompile(`^\$\{(\w+.)+\w+\}$`), "Must be of the pattern ^\\$\\{(\\w+.)+\\w+\\}$"),
										),
									},
								},
							},
						},
						"tag_step_details": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 30),
											validation.StringMatch(regexp.MustCompile(`^[\w-]*$`), "Must be of the pattern ^[\\w-]*$"),
										),
									},
									"source_file_location": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 256),
											validation.StringMatch(regexp.MustCompile(`^\$\{(\w+.)+\w+\}$`), "Must be of the pattern ^\\$\\{(\\w+.)+\\w+\\}$"),
										),
									},
									"tags": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 10,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(0, 128),
												},
												"value": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(0, 256),
												},
											},
										},
									},
								},
							},
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(transfer.WorkflowStepType_Values(), false),
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceWorkflowCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).TransferConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &transfer.CreateWorkflowInput{}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("on_exception_steps"); ok && len(v.([]interface{})) > 0 {
		input.OnExceptionSteps = expandWorkflows(v.([]interface{}))
	}

	if v, ok := d.GetOk("steps"); ok && len(v.([]interface{})) > 0 {
		input.Steps = expandWorkflows(v.([]interface{}))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating Transfer Workflow: %s", input)
	output, err := conn.CreateWorkflow(input)

	if err != nil {
		return fmt.Errorf("error creating Transfer Workflow: %w", err)
	}

	d.SetId(aws.StringValue(output.WorkflowId))

	return resourceWorkflowRead(d, meta)
}

func resourceWorkflowRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).TransferConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindWorkflowByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Transfer Workflow (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Transfer Workflow (%s): %w", d.Id(), err)
	}

	d.Set("arn", output.Arn)
	d.Set("description", output.Description)

	if err := d.Set("on_exception_steps", flattenWorkflows(output.OnExceptionSteps)); err != nil {
		return fmt.Errorf("error setting on_exception_steps: %w", err)
	}

	if err := d.Set("steps", flattenWorkflows(output.Steps)); err != nil {
		return fmt.Errorf("error setting steps: %w", err)
	}

	tags := KeyValueTags(output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceWorkflowUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).TransferConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceWorkflowRead(d, meta)
}

func resourceWorkflowDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).TransferConn

	log.Printf("[DEBUG] Deleting Transfer Workflow: (%s)", d.Id())
	_, err := conn.DeleteWorkflow(&transfer.DeleteWorkflowInput{
		WorkflowId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, transfer.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Transfer Workflow (%s): %w", d.Id(), err)
	}

	return nil
}

func expandWorkflows(tfList []interface{}) []*transfer.WorkflowStep {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*transfer.WorkflowStep

	for _, tfMapRaw := range tfList {
		tfMap, _ := tfMapRaw.(map[string]interface{})

		apiObject := &transfer.WorkflowStep{
			Type: aws.String(tfMap["type"].(string)),
		}

		if v, ok := tfMap["delete_step_details"].([]interface{}); ok && len(v) > 0 {
			apiObject.DeleteStepDetails = expandDeleteStepDetails(v)
		}

		if v, ok := tfMap["copy_step_details"].([]interface{}); ok && len(v) > 0 {
			apiObject.CopyStepDetails = expandCopyStepDetails(v)
		}

		if v, ok := tfMap["custom_step_details"].([]interface{}); ok && len(v) > 0 {
			apiObject.CustomStepDetails = expandCustomStepDetails(v)
		}

		if v, ok := tfMap["tag_step_details"].([]interface{}); ok && len(v) > 0 {
			apiObject.TagStepDetails = expandTagStepDetails(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenWorkflows(apiObjects []*transfer.WorkflowStep) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		flattenedObject := map[string]interface{}{
			"type": aws.StringValue(apiObject.Type),
		}

		if apiObject.DeleteStepDetails != nil {
			flattenedObject["delete_step_details"] = flattenDeleteStepDetails(apiObject.DeleteStepDetails)
		}

		if apiObject.DeleteStepDetails != nil {
			flattenedObject["copy_step_details"] = flattenCopyStepDetails(apiObject.CopyStepDetails)
		}

		if apiObject.CustomStepDetails != nil {
			flattenedObject["custom_step_details"] = flattenCustomStepDetails(apiObject.CustomStepDetails)
		}

		if apiObject.TagStepDetails != nil {
			flattenedObject["tag_step_details"] = flattenTagStepDetails(apiObject.TagStepDetails)
		}

		tfList = append(tfList, flattenedObject)
	}

	return tfList
}

func expandDeleteStepDetails(tfMap []interface{}) *transfer.DeleteStepDetails {
	if tfMap == nil {
		return nil
	}

	tfMapRaw := tfMap[0].(map[string]interface{})

	apiObject := &transfer.DeleteStepDetails{}

	if v, ok := tfMapRaw["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMapRaw["source_file_location"].(string); ok && v != "" {
		apiObject.SourceFileLocation = aws.String(v)
	}

	return apiObject
}

func flattenDeleteStepDetails(apiObject *transfer.DeleteStepDetails) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	if v := apiObject.SourceFileLocation; v != nil {
		tfMap["source_file_location"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

func expandCopyStepDetails(tfMap []interface{}) *transfer.CopyStepDetails {
	if tfMap == nil {
		return nil
	}

	tfMapRaw := tfMap[0].(map[string]interface{})

	apiObject := &transfer.CopyStepDetails{}

	if v, ok := tfMapRaw["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMapRaw["overwrite_existing"].(string); ok && v != "" {
		apiObject.OverwriteExisting = aws.String(v)
	}

	if v, ok := tfMapRaw["source_file_location"].(string); ok && v != "" {
		apiObject.SourceFileLocation = aws.String(v)
	}

	if v, ok := tfMapRaw["destination_file_location"].([]interface{}); ok && len(v) > 0 {
		apiObject.DestinationFileLocation = expandDestinationFileLocation(v)
	}

	return apiObject
}

func flattenCopyStepDetails(apiObject *transfer.CopyStepDetails) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	if v := apiObject.OverwriteExisting; v != nil {
		tfMap["overwrite_existing"] = aws.StringValue(v)
	}

	if v := apiObject.SourceFileLocation; v != nil {
		tfMap["source_file_location"] = aws.StringValue(v)
	}

	if v := apiObject.DestinationFileLocation; v != nil {
		tfMap["destination_file_location"] = flattenDestinationFileLocation(v)
	}

	return []interface{}{tfMap}
}

func expandCustomStepDetails(tfMap []interface{}) *transfer.CustomStepDetails {
	if tfMap == nil {
		return nil
	}

	tfMapRaw := tfMap[0].(map[string]interface{})

	apiObject := &transfer.CustomStepDetails{}

	if v, ok := tfMapRaw["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMapRaw["source_file_location"].(string); ok && v != "" {
		apiObject.SourceFileLocation = aws.String(v)
	}

	if v, ok := tfMapRaw["target"].(string); ok && v != "" {
		apiObject.Target = aws.String(v)
	}

	if v, ok := tfMapRaw["timeout_seconds"].(int); ok && v > 0 {
		apiObject.TimeoutSeconds = aws.Int64(int64(v))
	}

	return apiObject
}

func flattenCustomStepDetails(apiObject *transfer.CustomStepDetails) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	if v := apiObject.SourceFileLocation; v != nil {
		tfMap["source_file_location"] = aws.StringValue(v)
	}

	if v := apiObject.Target; v != nil {
		tfMap["target"] = aws.StringValue(v)
	}

	if v := apiObject.TimeoutSeconds; v != nil {
		tfMap["timeout_seconds"] = aws.Int64Value(v)
	}

	return []interface{}{tfMap}
}

func expandTagStepDetails(tfMap []interface{}) *transfer.TagStepDetails {
	if tfMap == nil {
		return nil
	}

	tfMapRaw := tfMap[0].(map[string]interface{})

	apiObject := &transfer.TagStepDetails{}

	if v, ok := tfMapRaw["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMapRaw["source_file_location"].(string); ok && v != "" {
		apiObject.SourceFileLocation = aws.String(v)
	}

	if v, ok := tfMapRaw["tags"].([]interface{}); ok && len(v) > 0 {
		apiObject.Tags = expandS3Tags(v)
	}

	return apiObject
}

func flattenTagStepDetails(apiObject *transfer.TagStepDetails) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	if v := apiObject.SourceFileLocation; v != nil {
		tfMap["source_file_location"] = aws.StringValue(v)
	}

	if apiObject.Tags != nil {
		tfMap["tags"] = flattenS3Tags(apiObject.Tags)
	}

	return []interface{}{tfMap}
}

func expandDestinationFileLocation(tfMap []interface{}) *transfer.InputFileLocation {
	if tfMap == nil {
		return nil
	}

	tfMapRaw := tfMap[0].(map[string]interface{})

	apiObject := &transfer.InputFileLocation{}

	if v, ok := tfMapRaw["efs_file_location"].([]interface{}); ok && len(v) > 0 {
		apiObject.EfsFileLocation = expandEFSFileLocation(v)
	}

	if v, ok := tfMapRaw["s3_file_location"].([]interface{}); ok && len(v) > 0 {
		apiObject.S3FileLocation = expandS3FileLocation(v)
	}

	return apiObject
}

func flattenDestinationFileLocation(apiObject *transfer.InputFileLocation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EfsFileLocation; v != nil {
		tfMap["efs_file_location"] = flattenEFSFileLocation(v)
	}

	if v := apiObject.S3FileLocation; v != nil {
		tfMap["s3_file_location"] = flattenS3FileLocation(v)
	}

	return []interface{}{tfMap}
}

func expandEFSFileLocation(tfMap []interface{}) *transfer.EfsFileLocation {
	if tfMap == nil {
		return nil
	}

	tfMapRaw := tfMap[0].(map[string]interface{})

	apiObject := &transfer.EfsFileLocation{}

	if v, ok := tfMapRaw["file_system_id"].(string); ok && v != "" {
		apiObject.FileSystemId = aws.String(v)
	}

	if v, ok := tfMapRaw["path"].(string); ok && v != "" {
		apiObject.Path = aws.String(v)
	}

	return apiObject
}

func flattenEFSFileLocation(apiObject *transfer.EfsFileLocation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.FileSystemId; v != nil {
		tfMap["file_system_id"] = aws.StringValue(v)
	}

	if v := apiObject.Path; v != nil {
		tfMap["path"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

func expandS3FileLocation(tfMap []interface{}) *transfer.S3InputFileLocation {
	if tfMap == nil {
		return nil
	}

	tfMapRaw := tfMap[0].(map[string]interface{})

	apiObject := &transfer.S3InputFileLocation{}

	if v, ok := tfMapRaw["bucket"].(string); ok && v != "" {
		apiObject.Bucket = aws.String(v)
	}

	if v, ok := tfMapRaw["key"].(string); ok && v != "" {
		apiObject.Key = aws.String(v)
	}

	return apiObject
}

func flattenS3FileLocation(apiObject *transfer.S3InputFileLocation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Bucket; v != nil {
		tfMap["bucket"] = aws.StringValue(v)
	}

	if v := apiObject.Key; v != nil {
		tfMap["key"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

func expandS3Tags(tfList []interface{}) []*transfer.S3Tag {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*transfer.S3Tag

	for _, tfMapRaw := range tfList {
		tfMap, _ := tfMapRaw.(map[string]interface{})

		apiObject := &transfer.S3Tag{}

		if v, ok := tfMap["key"].(string); ok && v != "" {
			apiObject.Key = aws.String(v)
		}

		if v, ok := tfMap["value"].(string); ok && v != "" {
			apiObject.Value = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenS3Tags(apiObjects []*transfer.S3Tag) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		flattenedObject := map[string]interface{}{}

		if v := apiObject.Key; v != nil {
			flattenedObject["key"] = aws.StringValue(v)
		}

		if v := apiObject.Value; v != nil {
			flattenedObject["value"] = aws.StringValue(v)
		}

		tfList = append(tfList, flattenedObject)
	}

	return tfList
}
