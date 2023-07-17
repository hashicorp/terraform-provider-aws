// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	documentPermissionsBatchLimit = 20
)

// @SDKResource("aws_ssm_document", name="Document")
// @Tags(identifierAttribute="id", resourceType="Document")
func ResourceDocument() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDocumentCreate,
		ReadWithoutTimeout:   resourceDocumentRead,
		UpdateWithoutTimeout: resourceDocumentUpdate,
		DeleteWithoutTimeout: resourceDocumentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attachments_source": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 20,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(ssm.AttachmentsSourceKey_Values(), false),
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.All(
								validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_\-.]+$`), "must contain only alphanumeric, underscore, hyphen, or period characters"),
								validation.StringLenBetween(3, 128),
							),
						},
						"values": {
							Type:     schema.TypeList,
							MinItems: 1,
							Required: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(1, 1024),
							},
						},
					},
				},
			},
			"content": {
				Type:     schema.TypeString,
				Required: true,
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"document_format": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ssm.DocumentFormatJson,
				ValidateFunc: validation.StringInSlice(ssm.DocumentFormat_Values(), false),
			},
			"document_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ssm.DocumentType_Values(), false),
			},
			"document_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hash": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hash_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"latest_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_\-.]+$`), "must contain only alphanumeric, underscore, hyphen, or period characters"),
					validation.StringLenBetween(3, 128),
				),
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parameter": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default_value": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"permissions": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"platform_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"schema_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"target_type": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexp.MustCompile(`^\/[\w\.\-\:\/]*$`), "must contain a forward slash optionally followed by a resource type such as AWS::EC2::Instance"),
					validation.StringLenBetween(1, 200),
				),
			},
			"version_name": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_\-.]{3,128}$`), "must contain only alphanumeric, underscore, hyphen, or period characters"),
					validation.StringLenBetween(3, 128),
				),
			},
		},

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
			func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
				if v, ok := d.GetOk("permissions"); ok && len(v.(map[string]interface{})) > 0 {
					// Validates permissions keys, if set, to be type and account_ids
					// since ValidateFunc validates only the value not the key.
					tfMap := flex.ExpandStringValueMap(v.(map[string]interface{}))

					if v, ok := tfMap["type"]; ok {
						if v != ssm.DocumentPermissionTypeShare {
							return fmt.Errorf("%q: only %s \"type\" supported", "permissions", ssm.DocumentPermissionTypeShare)
						}
					} else {
						return fmt.Errorf("%q: \"type\" must be defined", "permissions")
					}

					if _, ok := tfMap["account_ids"]; !ok {
						return fmt.Errorf("%q: \"account_ids\" must be defined", "permissions")
					}
				}

				if d.HasChange("content") {
					if err := d.SetNewComputed("default_version"); err != nil {
						return err
					}
					if err := d.SetNewComputed("document_version"); err != nil {
						return err
					}
					if err := d.SetNewComputed("hash"); err != nil {
						return err
					}
					if err := d.SetNewComputed("latest_version"); err != nil {
						return err
					}
					if err := d.SetNewComputed("parameter"); err != nil {
						return err
					}
				}

				return nil
			},
		),
	}
}

func resourceDocumentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn(ctx)

	name := d.Get("name").(string)
	input := &ssm.CreateDocumentInput{
		Content:        aws.String(d.Get("content").(string)),
		DocumentFormat: aws.String(d.Get("document_format").(string)),
		DocumentType:   aws.String(d.Get("document_type").(string)),
		Name:           aws.String(name),
		Tags:           getTagsIn(ctx),
	}

	if v, ok := d.GetOk("attachments_source"); ok && len(v.([]interface{})) > 0 {
		input.Attachments = expandAttachmentsSources(v.([]interface{}))
	}

	if v, ok := d.GetOk("target_type"); ok {
		input.TargetType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("version_name"); ok {
		input.VersionName = aws.String(v.(string))
	}

	output, err := conn.CreateDocumentWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SSM Document (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.DocumentDescription.Name))

	if v, ok := d.GetOk("permissions"); ok && len(v.(map[string]interface{})) > 0 {
		tfMap := flex.ExpandStringValueMap(v.(map[string]interface{}))

		if v, ok := tfMap["account_ids"]; ok && v != "" {
			chunks := slices.Chunks(strings.Split(v, ","), documentPermissionsBatchLimit)

			for _, chunk := range chunks {
				input := &ssm.ModifyDocumentPermissionInput{
					AccountIdsToAdd: aws.StringSlice(chunk),
					Name:            aws.String(d.Id()),
					PermissionType:  aws.String(ssm.DocumentPermissionTypeShare),
				}

				_, err := conn.ModifyDocumentPermissionWithContext(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "modifying SSM Document (%s) permissions: %s", d.Id(), err)
				}
			}
		}
	}

	if _, err := waitDocumentActive(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SSM Document (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceDocumentRead(ctx, d, meta)...)
}

func resourceDocumentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn(ctx)

	doc, err := FindDocumentByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSM Document %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSM Document (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ssm",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("document/%s", aws.StringValue(doc.Name)),
	}.String()
	d.Set("arn", arn)
	d.Set("created_date", aws.TimeValue(doc.CreatedDate).Format(time.RFC3339))
	d.Set("default_version", doc.DefaultVersion)
	d.Set("description", doc.Description)
	d.Set("document_format", doc.DocumentFormat)
	d.Set("document_type", doc.DocumentType)
	d.Set("document_version", doc.DocumentVersion)
	d.Set("hash", doc.Hash)
	d.Set("hash_type", doc.HashType)
	d.Set("latest_version", doc.LatestVersion)
	d.Set("name", doc.Name)
	d.Set("owner", doc.Owner)
	if err := d.Set("parameter", flattenDocumentParameters(doc.Parameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameter: %s", err)
	}
	d.Set("platform_types", aws.StringValueSlice(doc.PlatformTypes))
	d.Set("schema_version", doc.SchemaVersion)
	d.Set("status", doc.Status)
	d.Set("target_type", doc.TargetType)
	d.Set("version_name", doc.VersionName)

	{
		input := &ssm.GetDocumentInput{
			DocumentFormat:  aws.String(d.Get("document_format").(string)),
			DocumentVersion: aws.String("$LATEST"),
			Name:            aws.String(d.Id()),
		}

		output, err := conn.GetDocumentWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading SSM Document (%s) content: %s", d.Id(), err)
		}

		d.Set("content", output.Content)
	}

	{
		input := &ssm.DescribeDocumentPermissionInput{
			Name:           aws.String(d.Id()),
			PermissionType: aws.String(ssm.DocumentPermissionTypeShare),
		}

		output, err := conn.DescribeDocumentPermissionWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading SSM Document (%s) permissions: %s", d.Id(), err)
		}

		if accountsIDs := aws.StringValueSlice(output.AccountIds); len(accountsIDs) > 0 {
			d.Set("permissions", map[string]string{
				"account_ids": strings.Join(accountsIDs, ","),
				"type":        ssm.DocumentPermissionTypeShare,
			})
		} else {
			d.Set("permissions", nil)
		}
	}

	setTagsOut(ctx, doc.Tags)

	return diags
}

func resourceDocumentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn(ctx)

	if d.HasChange("permissions") {
		var oldAccountIDs, newAccountIDs flex.Set[string]
		o, n := d.GetChange("permissions")

		if v := o.(map[string]interface{}); len(v) > 0 {
			tfMap := flex.ExpandStringValueMap(v)

			if v, ok := tfMap["account_ids"]; ok && v != "" {
				oldAccountIDs = strings.Split(v, ",")
			}
		}

		if v := n.(map[string]interface{}); len(v) > 0 {
			tfMap := flex.ExpandStringValueMap(v)

			if v, ok := tfMap["account_ids"]; ok && v != "" {
				newAccountIDs = strings.Split(v, ",")
			}
		}

		for _, chunk := range slices.Chunks(newAccountIDs.Difference(oldAccountIDs), documentPermissionsBatchLimit) {
			input := &ssm.ModifyDocumentPermissionInput{
				AccountIdsToAdd: aws.StringSlice(chunk),
				Name:            aws.String(d.Id()),
				PermissionType:  aws.String(ssm.DocumentPermissionTypeShare),
			}

			_, err := conn.ModifyDocumentPermissionWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "modifying SSM Document (%s) permissions: %s", d.Id(), err)
			}
		}

		for _, chunk := range slices.Chunks(oldAccountIDs.Difference(newAccountIDs), documentPermissionsBatchLimit) {
			input := &ssm.ModifyDocumentPermissionInput{
				AccountIdsToRemove: aws.StringSlice(chunk),
				Name:               aws.String(d.Id()),
				PermissionType:     aws.String(ssm.DocumentPermissionTypeShare),
			}

			_, err := conn.ModifyDocumentPermissionWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "modifying SSM Document (%s) permissions: %s", d.Id(), err)
			}
		}
	}

	if d.HasChangesExcept("permissions", "tags", "tags_all") {
		// Update for schema version 1.x is not allowed.
		isSchemaVersion1, _ := regexp.MatchString(`^1[.][0-9]$`, d.Get("schema_version").(string))

		if d.HasChange("content") || !isSchemaVersion1 {
			input := &ssm.UpdateDocumentInput{
				Content:         aws.String(d.Get("content").(string)),
				DocumentFormat:  aws.String(d.Get("document_format").(string)),
				DocumentVersion: aws.String(d.Get("default_version").(string)),
				Name:            aws.String(d.Id()),
			}

			if v, ok := d.GetOk("attachments_source"); ok && len(v.([]interface{})) > 0 {
				input.Attachments = expandAttachmentsSources(v.([]interface{}))
			}

			if v, ok := d.GetOk("target_type"); ok {
				input.TargetType = aws.String(v.(string))
			}

			if v, ok := d.GetOk("version_name"); ok {
				input.VersionName = aws.String(v.(string))
			}

			var defaultVersion string

			output, err := conn.UpdateDocumentWithContext(ctx, input)

			if tfawserr.ErrCodeEquals(err, ssm.ErrCodeDuplicateDocumentContent) {
				defaultVersion = d.Get("latest_version").(string)
			} else if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating SSM Document (%s): %s", d.Id(), err)
			} else {
				defaultVersion = aws.StringValue(output.DocumentDescription.DocumentVersion)
			}

			_, err = conn.UpdateDocumentDefaultVersionWithContext(ctx, &ssm.UpdateDocumentDefaultVersionInput{
				DocumentVersion: aws.String(defaultVersion),
				Name:            aws.String(d.Id()),
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating SSM Document (%s) default version: %s", d.Id(), err)
			}

			if _, err := waitDocumentActive(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for SSM Document (%s) update: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceDocumentRead(ctx, d, meta)...)
}

func resourceDocumentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn(ctx)

	if v, ok := d.GetOk("permissions"); ok && len(v.(map[string]interface{})) > 0 {
		tfMap := flex.ExpandStringValueMap(v.(map[string]interface{}))

		if v, ok := tfMap["account_ids"]; ok && v != "" {
			chunks := slices.Chunks(strings.Split(v, ","), documentPermissionsBatchLimit)

			for _, chunk := range chunks {
				input := &ssm.ModifyDocumentPermissionInput{
					AccountIdsToRemove: aws.StringSlice(chunk),
					Name:               aws.String(d.Id()),
					PermissionType:     aws.String(ssm.DocumentPermissionTypeShare),
				}

				_, err := conn.ModifyDocumentPermissionWithContext(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "modifying SSM Document (%s) permissions: %s", d.Id(), err)
				}
			}
		}
	}

	log.Printf("[INFO] Deleting SSM Document: %s", d.Id())
	_, err := conn.DeleteDocumentWithContext(ctx, &ssm.DeleteDocumentInput{
		Name: aws.String(d.Get("name").(string)),
	})

	if tfawserr.ErrMessageContains(err, ssm.ErrCodeInvalidDocument, "does not exist") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SSM Document (%s): %s", d.Id(), err)
	}

	if _, err := waitDocumentDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SSM Document (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindDocumentByName(ctx context.Context, conn *ssm.SSM, name string) (*ssm.DocumentDescription, error) {
	input := &ssm.DescribeDocumentInput{
		Name: aws.String(name),
	}

	output, err := conn.DescribeDocumentWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, ssm.ErrCodeInvalidDocument, "does not exist") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Document == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Document, nil
}

func statusDocument(ctx context.Context, conn *ssm.SSM, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDocumentByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitDocumentActive(ctx context.Context, conn *ssm.SSM, name string) (*ssm.DocumentDescription, error) { //nolint:unparam
	const (
		timeout = 2 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{ssm.DocumentStatusCreating, ssm.DocumentStatusUpdating},
		Target:  []string{ssm.DocumentStatusActive},
		Refresh: statusDocument(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ssm.DocumentDescription); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusInformation)))

		return output, err
	}

	return nil, err
}

func waitDocumentDeleted(ctx context.Context, conn *ssm.SSM, name string) (*ssm.DocumentDescription, error) {
	const (
		timeout = 2 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{ssm.DocumentStatusDeleting},
		Target:  []string{},
		Refresh: statusDocument(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ssm.DocumentDescription); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusInformation)))

		return output, err
	}

	return nil, err
}

func expandAttachmentsSource(tfMap map[string]interface{}) *ssm.AttachmentsSource {
	if tfMap == nil {
		return nil
	}

	apiObject := &ssm.AttachmentsSource{}

	if v, ok := tfMap["key"].(string); ok && v != "" {
		apiObject.Key = aws.String(v)
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["values"].([]interface{}); ok && len(v) > 0 {
		apiObject.Values = flex.ExpandStringList(v)
	}

	return apiObject
}

func expandAttachmentsSources(tfList []interface{}) []*ssm.AttachmentsSource {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ssm.AttachmentsSource

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandAttachmentsSource(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenDocumentParameter(apiObject *ssm.DocumentParameter) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DefaultValue; v != nil {
		tfMap["default_value"] = aws.StringValue(v)
	}

	if v := apiObject.Description; v != nil {
		tfMap["description"] = aws.StringValue(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenDocumentParameters(apiObjects []*ssm.DocumentParameter) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenDocumentParameter(apiObject))
	}

	return tfList
}
