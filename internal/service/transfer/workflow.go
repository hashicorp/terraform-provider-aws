// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/transfer"
	awstypes "github.com/aws/aws-sdk-go-v2/service/transfer/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_transfer_workflow", name="Workflow")
// @Tags(identifierAttribute="arn")
func resourceWorkflow() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWorkflowCreate,
		ReadWithoutTimeout:   resourceWorkflowRead,
		UpdateWithoutTimeout: resourceWorkflowUpdate,
		DeleteWithoutTimeout: resourceWorkflowDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
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
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"destination_file_location": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"efs_file_location": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrFileSystemID: {
																Type:     schema.TypeString,
																Optional: true,
																ForceNew: true,
															},
															names.AttrPath: {
																Type:         schema.TypeString,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: validation.StringLenBetween(1, 65536),
															},
														},
													},
												},
												"s3_file_location": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrBucket: {
																Type:     schema.TypeString,
																Optional: true,
																ForceNew: true,
															},
															names.AttrKey: {
																Type:         schema.TypeString,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: validation.StringLenBetween(0, 1024),
															},
														},
													},
												},
											},
										},
									},
									names.AttrName: {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 30),
											validation.StringMatch(regexache.MustCompile(`^[\w-]*$`), "Must be of the pattern ^[\\w-]*$"),
										),
									},
									"overwrite_existing": {
										Type:             schema.TypeString,
										Optional:         true,
										ForceNew:         true,
										Default:          awstypes.OverwriteExistingFalse,
										ValidateDiagFunc: enum.Validate[awstypes.OverwriteExisting](),
									},
									"source_file_location": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 256),
											validation.StringMatch(regexache.MustCompile(`^\$\{(\w+.)+\w+\}$`), "Must be of the pattern ^\\$\\{(\\w+.)+\\w+\\}$"),
										),
									},
								},
							},
						},
						"custom_step_details": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 30),
											validation.StringMatch(regexache.MustCompile(`^[\w-]*$`), "Must be of the pattern ^[\\w-]*$"),
										),
									},
									"source_file_location": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 256),
											validation.StringMatch(regexache.MustCompile(`^\$\{(\w+.)+\w+\}$`), "Must be of the pattern ^\\$\\{(\\w+.)+\\w+\\}$"),
										),
									},
									names.AttrTarget: {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: verify.ValidARN,
									},
									"timeout_seconds": {
										Type:         schema.TypeInt,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntBetween(1, 1800),
									},
								},
							},
						},
						"decrypt_step_details": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"destination_file_location": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"efs_file_location": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrFileSystemID: {
																Type:     schema.TypeString,
																Optional: true,
																ForceNew: true,
															},
															names.AttrPath: {
																Type:         schema.TypeString,
																Optional:     true,
																ForceNew:     true,
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
															names.AttrBucket: {
																Type:     schema.TypeString,
																Optional: true,
																ForceNew: true,
															},
															names.AttrKey: {
																Type:         schema.TypeString,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: validation.StringLenBetween(0, 1024),
															},
														},
													},
												},
											},
										},
									},
									names.AttrName: {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 30),
											validation.StringMatch(regexache.MustCompile(`^[\w-]*$`), "Must be of the pattern ^[\\w-]*$"),
										),
									},
									"overwrite_existing": {
										Type:             schema.TypeString,
										Optional:         true,
										ForceNew:         true,
										Default:          awstypes.OverwriteExistingFalse,
										ValidateDiagFunc: enum.Validate[awstypes.OverwriteExisting](),
									},
									"source_file_location": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 256),
											validation.StringMatch(regexache.MustCompile(`^\$\{(\w+.)+\w+\}$`), "Must be of the pattern ^\\$\\{(\\w+.)+\\w+\\}$"),
										),
									},
									names.AttrType: {
										Type:             schema.TypeString,
										Required:         true,
										ForceNew:         true,
										ValidateDiagFunc: enum.Validate[awstypes.EncryptionType](),
									},
								},
							},
						},
						"delete_step_details": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 30),
											validation.StringMatch(regexache.MustCompile(`^[\w-]*$`), "Must be of the pattern ^[\\w-]*$"),
										),
									},
									"source_file_location": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 256),
											validation.StringMatch(regexache.MustCompile(`^\$\{(\w+.)+\w+\}$`), "Must be of the pattern ^\\$\\{(\\w+.)+\\w+\\}$"),
										),
									},
								},
							},
						},
						"tag_step_details": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 30),
											validation.StringMatch(regexache.MustCompile(`^[\w-]*$`), "Must be of the pattern ^[\\w-]*$"),
										),
									},
									"source_file_location": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 256),
											validation.StringMatch(regexache.MustCompile(`^\$\{(\w+.)+\w+\}$`), "Must be of the pattern ^\\$\\{(\\w+.)+\\w+\\}$"),
										),
									},
									names.AttrTags: {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										MaxItems: 10,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrKey: {
													Type:         schema.TypeString,
													Required:     true,
													ForceNew:     true,
													ValidateFunc: validation.StringLenBetween(0, 128),
												},
												names.AttrValue: {
													Type:         schema.TypeString,
													Required:     true,
													ForceNew:     true,
													ValidateFunc: validation.StringLenBetween(0, 256),
												},
											},
										},
									},
								},
							},
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.WorkflowStepType](),
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
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"destination_file_location": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"efs_file_location": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrFileSystemID: {
																Type:     schema.TypeString,
																Optional: true,
																ForceNew: true,
															},
															names.AttrPath: {
																Type:         schema.TypeString,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: validation.StringLenBetween(1, 65536),
															},
														},
													},
												},
												"s3_file_location": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrBucket: {
																Type:     schema.TypeString,
																Optional: true,
																ForceNew: true,
															},
															names.AttrKey: {
																Type:         schema.TypeString,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: validation.StringLenBetween(0, 1024),
															},
														},
													},
												},
											},
										},
									},
									names.AttrName: {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 30),
											validation.StringMatch(regexache.MustCompile(`^[\w-]*$`), "Must be of the pattern ^[\\w-]*$"),
										),
									},
									"overwrite_existing": {
										Type:             schema.TypeString,
										Optional:         true,
										ForceNew:         true,
										Default:          awstypes.OverwriteExistingFalse,
										ValidateDiagFunc: enum.Validate[awstypes.OverwriteExisting](),
									},
									"source_file_location": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 256),
											validation.StringMatch(regexache.MustCompile(`^\$\{(\w+.)+\w+\}$`), "Must be of the pattern ^\\$\\{(\\w+.)+\\w+\\}$"),
										),
									},
								},
							},
						},
						"custom_step_details": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 30),
											validation.StringMatch(regexache.MustCompile(`^[\w-]*$`), "Must be of the pattern ^[\\w-]*$"),
										),
									},
									"source_file_location": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 256),
											validation.StringMatch(regexache.MustCompile(`^\$\{(\w+.)+\w+\}$`), "Must be of the pattern ^\\$\\{(\\w+.)+\\w+\\}$"),
										),
									},
									names.AttrTarget: {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: verify.ValidARN,
									},
									"timeout_seconds": {
										Type:         schema.TypeInt,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntBetween(1, 1800),
									},
								},
							},
						},
						"decrypt_step_details": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"destination_file_location": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"efs_file_location": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrFileSystemID: {
																Type:     schema.TypeString,
																Optional: true,
																ForceNew: true,
															},
															names.AttrPath: {
																Type:         schema.TypeString,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: validation.StringLenBetween(1, 65536),
															},
														},
													},
												},
												"s3_file_location": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrBucket: {
																Type:     schema.TypeString,
																Optional: true,
																ForceNew: true,
															},
															names.AttrKey: {
																Type:         schema.TypeString,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: validation.StringLenBetween(0, 1024),
															},
														},
													},
												},
											},
										},
									},
									names.AttrName: {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 30),
											validation.StringMatch(regexache.MustCompile(`^[\w-]*$`), "Must be of the pattern ^[\\w-]*$"),
										),
									},
									"overwrite_existing": {
										Type:             schema.TypeString,
										Optional:         true,
										ForceNew:         true,
										Default:          awstypes.OverwriteExistingFalse,
										ValidateDiagFunc: enum.Validate[awstypes.OverwriteExisting](),
									},
									"source_file_location": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 256),
											validation.StringMatch(regexache.MustCompile(`^\$\{(\w+.)+\w+\}$`), "Must be of the pattern ^\\$\\{(\\w+.)+\\w+\\}$"),
										),
									},
									names.AttrType: {
										Type:             schema.TypeString,
										Required:         true,
										ForceNew:         true,
										ValidateDiagFunc: enum.Validate[awstypes.EncryptionType](),
									},
								},
							},
						},
						"delete_step_details": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 30),
											validation.StringMatch(regexache.MustCompile(`^[\w-]*$`), "Must be of the pattern ^[\\w-]*$"),
										),
									},
									"source_file_location": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 256),
											validation.StringMatch(regexache.MustCompile(`^\$\{(\w+.)+\w+\}$`), "Must be of the pattern ^\\$\\{(\\w+.)+\\w+\\}$"),
										),
									},
								},
							},
						},
						"tag_step_details": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 30),
											validation.StringMatch(regexache.MustCompile(`^[\w-]*$`), "Must be of the pattern ^[\\w-]*$"),
										),
									},
									"source_file_location": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 256),
											validation.StringMatch(regexache.MustCompile(`^\$\{(\w+.)+\w+\}$`), "Must be of the pattern ^\\$\\{(\\w+.)+\\w+\\}$"),
										),
									},
									names.AttrTags: {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										MaxItems: 10,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrKey: {
													Type:         schema.TypeString,
													Required:     true,
													ForceNew:     true,
													ValidateFunc: validation.StringLenBetween(0, 128),
												},
												names.AttrValue: {
													Type:         schema.TypeString,
													Required:     true,
													ForceNew:     true,
													ValidateFunc: validation.StringLenBetween(0, 256),
												},
											},
										},
									},
								},
							},
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.WorkflowStepType](),
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceWorkflowCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	input := &transfer.CreateWorkflowInput{
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("on_exception_steps"); ok && len(v.([]interface{})) > 0 {
		input.OnExceptionSteps = expandWorkflowSteps(v.([]interface{}))
	}

	if v, ok := d.GetOk("steps"); ok && len(v.([]interface{})) > 0 {
		input.Steps = expandWorkflowSteps(v.([]interface{}))
	}

	output, err := conn.CreateWorkflow(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Transfer Workflow: %s", err)
	}

	d.SetId(aws.ToString(output.WorkflowId))

	return append(diags, resourceWorkflowRead(ctx, d, meta)...)
}

func resourceWorkflowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	output, err := findWorkflowByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Transfer Workflow (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Transfer Workflow (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.Arn)
	d.Set(names.AttrDescription, output.Description)
	if err := d.Set("on_exception_steps", flattenWorkflowSteps(output.OnExceptionSteps)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting on_exception_steps: %s", err)
	}
	if err := d.Set("steps", flattenWorkflowSteps(output.Steps)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting steps: %s", err)
	}

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceWorkflowUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceWorkflowRead(ctx, d, meta)...)
}

func resourceWorkflowDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	log.Printf("[DEBUG] Deleting Transfer Workflow: %s", d.Id())
	_, err := conn.DeleteWorkflow(ctx, &transfer.DeleteWorkflowInput{
		WorkflowId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Transfer Workflow (%s): %s", d.Id(), err)
	}

	return diags
}

func findWorkflowByID(ctx context.Context, conn *transfer.Client, id string) (*awstypes.DescribedWorkflow, error) {
	input := &transfer.DescribeWorkflowInput{
		WorkflowId: aws.String(id),
	}

	output, err := conn.DescribeWorkflow(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Workflow == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Workflow, nil
}

func expandWorkflowSteps(tfList []interface{}) []awstypes.WorkflowStep {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.WorkflowStep

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := awstypes.WorkflowStep{
			Type: awstypes.WorkflowStepType(tfMap[names.AttrType].(string)),
		}

		if v, ok := tfMap["copy_step_details"].([]interface{}); ok && len(v) > 0 {
			apiObject.CopyStepDetails = expandCopyStepDetails(v)
		}

		if v, ok := tfMap["custom_step_details"].([]interface{}); ok && len(v) > 0 {
			apiObject.CustomStepDetails = expandCustomStepDetails(v)
		}

		if v, ok := tfMap["decrypt_step_details"].([]interface{}); ok && len(v) > 0 {
			apiObject.DecryptStepDetails = expandDecryptStepDetails(v)
		}

		if v, ok := tfMap["delete_step_details"].([]interface{}); ok && len(v) > 0 {
			apiObject.DeleteStepDetails = expandDeleteStepDetails(v)
		}

		if v, ok := tfMap["tag_step_details"].([]interface{}); ok && len(v) > 0 {
			apiObject.TagStepDetails = expandTagStepDetails(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenWorkflowSteps(apiObjects []awstypes.WorkflowStep) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			names.AttrType: apiObject.Type,
		}

		if apiObject.CopyStepDetails != nil {
			tfMap["copy_step_details"] = flattenCopyStepDetails(apiObject.CopyStepDetails)
		}

		if apiObject.CustomStepDetails != nil {
			tfMap["custom_step_details"] = flattenCustomStepDetails(apiObject.CustomStepDetails)
		}

		if apiObject.DecryptStepDetails != nil {
			tfMap["decrypt_step_details"] = flattenDecryptStepDetails(apiObject.DecryptStepDetails)
		}

		if apiObject.DeleteStepDetails != nil {
			tfMap["delete_step_details"] = flattenDeleteStepDetails(apiObject.DeleteStepDetails)
		}

		if apiObject.TagStepDetails != nil {
			tfMap["tag_step_details"] = flattenTagStepDetails(apiObject.TagStepDetails)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandCopyStepDetails(tfMap []interface{}) *awstypes.CopyStepDetails {
	if tfMap == nil {
		return nil
	}

	tfMapRaw := tfMap[0].(map[string]interface{})

	apiObject := &awstypes.CopyStepDetails{}

	if v, ok := tfMapRaw["destination_file_location"].([]interface{}); ok && len(v) > 0 {
		apiObject.DestinationFileLocation = expandInputFileLocation(v)
	}

	if v, ok := tfMapRaw[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMapRaw["overwrite_existing"].(string); ok && v != "" {
		apiObject.OverwriteExisting = awstypes.OverwriteExisting(v)
	}

	if v, ok := tfMapRaw["source_file_location"].(string); ok && v != "" {
		apiObject.SourceFileLocation = aws.String(v)
	}

	return apiObject
}

func flattenCopyStepDetails(apiObject *awstypes.CopyStepDetails) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"overwrite_existing": apiObject.OverwriteExisting,
	}

	if v := apiObject.DestinationFileLocation; v != nil {
		tfMap["destination_file_location"] = flattenInputFileLocation(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.SourceFileLocation; v != nil {
		tfMap["source_file_location"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func expandCustomStepDetails(tfMap []interface{}) *awstypes.CustomStepDetails {
	if tfMap == nil {
		return nil
	}

	tfMapRaw := tfMap[0].(map[string]interface{})

	apiObject := &awstypes.CustomStepDetails{}

	if v, ok := tfMapRaw[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMapRaw["source_file_location"].(string); ok && v != "" {
		apiObject.SourceFileLocation = aws.String(v)
	}

	if v, ok := tfMapRaw[names.AttrTarget].(string); ok && v != "" {
		apiObject.Target = aws.String(v)
	}

	if v, ok := tfMapRaw["timeout_seconds"].(int); ok && v > 0 {
		apiObject.TimeoutSeconds = aws.Int32(int32(v))
	}

	return apiObject
}

func flattenCustomStepDetails(apiObject *awstypes.CustomStepDetails) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.SourceFileLocation; v != nil {
		tfMap["source_file_location"] = aws.ToString(v)
	}

	if v := apiObject.Target; v != nil {
		tfMap[names.AttrTarget] = aws.ToString(v)
	}

	if v := apiObject.TimeoutSeconds; v != nil {
		tfMap["timeout_seconds"] = aws.ToInt32(v)
	}

	return []interface{}{tfMap}
}

func expandDecryptStepDetails(tfMap []interface{}) *awstypes.DecryptStepDetails {
	if tfMap == nil {
		return nil
	}

	tfMapRaw := tfMap[0].(map[string]interface{})

	apiObject := &awstypes.DecryptStepDetails{}

	if v, ok := tfMapRaw["destination_file_location"].([]interface{}); ok && len(v) > 0 {
		apiObject.DestinationFileLocation = expandInputFileLocation(v)
	}

	if v, ok := tfMapRaw[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMapRaw["overwrite_existing"].(string); ok && v != "" {
		apiObject.OverwriteExisting = awstypes.OverwriteExisting(v)
	}

	if v, ok := tfMapRaw["source_file_location"].(string); ok && v != "" {
		apiObject.SourceFileLocation = aws.String(v)
	}

	if v, ok := tfMapRaw[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.EncryptionType(v)
	}

	return apiObject
}

func flattenDecryptStepDetails(apiObject *awstypes.DecryptStepDetails) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"overwrite_existing": apiObject.OverwriteExisting,
		names.AttrType:       apiObject.Type,
	}

	if v := apiObject.DestinationFileLocation; v != nil {
		tfMap["destination_file_location"] = flattenInputFileLocation(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.SourceFileLocation; v != nil {
		tfMap["source_file_location"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func expandDeleteStepDetails(tfMap []interface{}) *awstypes.DeleteStepDetails {
	if tfMap == nil {
		return nil
	}

	tfMapRaw := tfMap[0].(map[string]interface{})

	apiObject := &awstypes.DeleteStepDetails{}

	if v, ok := tfMapRaw[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMapRaw["source_file_location"].(string); ok && v != "" {
		apiObject.SourceFileLocation = aws.String(v)
	}

	return apiObject
}

func flattenDeleteStepDetails(apiObject *awstypes.DeleteStepDetails) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.SourceFileLocation; v != nil {
		tfMap["source_file_location"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func expandTagStepDetails(tfMap []interface{}) *awstypes.TagStepDetails {
	if tfMap == nil {
		return nil
	}

	tfMapRaw := tfMap[0].(map[string]interface{})

	apiObject := &awstypes.TagStepDetails{}

	if v, ok := tfMapRaw[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMapRaw["source_file_location"].(string); ok && v != "" {
		apiObject.SourceFileLocation = aws.String(v)
	}

	if v, ok := tfMapRaw[names.AttrTags].([]interface{}); ok && len(v) > 0 {
		apiObject.Tags = expandS3Tags(v)
	}

	return apiObject
}

func flattenTagStepDetails(apiObject *awstypes.TagStepDetails) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.SourceFileLocation; v != nil {
		tfMap["source_file_location"] = aws.ToString(v)
	}

	if apiObject.Tags != nil {
		tfMap[names.AttrTags] = flattenS3Tags(apiObject.Tags)
	}

	return []interface{}{tfMap}
}

func expandInputFileLocation(tfMap []interface{}) *awstypes.InputFileLocation {
	if tfMap == nil {
		return nil
	}

	tfMapRaw := tfMap[0].(map[string]interface{})

	apiObject := &awstypes.InputFileLocation{}

	if v, ok := tfMapRaw["efs_file_location"].([]interface{}); ok && len(v) > 0 {
		apiObject.EfsFileLocation = expandEFSFileLocation(v)
	}

	if v, ok := tfMapRaw["s3_file_location"].([]interface{}); ok && len(v) > 0 {
		apiObject.S3FileLocation = expandS3InputFileLocation(v)
	}

	return apiObject
}

func flattenInputFileLocation(apiObject *awstypes.InputFileLocation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EfsFileLocation; v != nil {
		tfMap["efs_file_location"] = flattenEFSFileLocation(v)
	}

	if v := apiObject.S3FileLocation; v != nil {
		tfMap["s3_file_location"] = flattenS3InputFileLocation(v)
	}

	return []interface{}{tfMap}
}

func expandEFSFileLocation(tfMap []interface{}) *awstypes.EfsFileLocation {
	if tfMap == nil {
		return nil
	}

	tfMapRaw := tfMap[0].(map[string]interface{})

	apiObject := &awstypes.EfsFileLocation{}

	if v, ok := tfMapRaw[names.AttrFileSystemID].(string); ok && v != "" {
		apiObject.FileSystemId = aws.String(v)
	}

	if v, ok := tfMapRaw[names.AttrPath].(string); ok && v != "" {
		apiObject.Path = aws.String(v)
	}

	return apiObject
}

func flattenEFSFileLocation(apiObject *awstypes.EfsFileLocation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.FileSystemId; v != nil {
		tfMap[names.AttrFileSystemID] = aws.ToString(v)
	}

	if v := apiObject.Path; v != nil {
		tfMap[names.AttrPath] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func expandS3InputFileLocation(tfMap []interface{}) *awstypes.S3InputFileLocation {
	if tfMap == nil {
		return nil
	}

	tfMapRaw := tfMap[0].(map[string]interface{})

	apiObject := &awstypes.S3InputFileLocation{}

	if v, ok := tfMapRaw[names.AttrBucket].(string); ok && v != "" {
		apiObject.Bucket = aws.String(v)
	}

	if v, ok := tfMapRaw[names.AttrKey].(string); ok && v != "" {
		apiObject.Key = aws.String(v)
	}

	return apiObject
}

func flattenS3InputFileLocation(apiObject *awstypes.S3InputFileLocation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Bucket; v != nil {
		tfMap[names.AttrBucket] = aws.ToString(v)
	}

	if v := apiObject.Key; v != nil {
		tfMap[names.AttrKey] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func expandS3Tags(tfList []interface{}) []awstypes.S3Tag {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.S3Tag

	for _, tfMapRaw := range tfList {
		tfMap, _ := tfMapRaw.(map[string]interface{})

		apiObject := awstypes.S3Tag{}

		if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
			apiObject.Key = aws.String(v)
		}

		if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
			apiObject.Value = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenS3Tags(apiObjects []awstypes.S3Tag) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		flattenedObject := map[string]interface{}{}

		if v := apiObject.Key; v != nil {
			flattenedObject[names.AttrKey] = aws.ToString(v)
		}

		if v := apiObject.Value; v != nil {
			flattenedObject[names.AttrValue] = aws.ToString(v)
		}

		tfList = append(tfList, flattenedObject)
	}

	return tfList
}
