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

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	documentPermissionsBatchLimit = 20
)

// @SDKResource("aws_ssm_document", name="Document")
// @Tags(identifierAttribute="id", resourceType="Document")
func resourceDocument() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDocumentCreate,
		ReadWithoutTimeout:   resourceDocumentRead,
		UpdateWithoutTimeout: resourceDocumentUpdate,
		DeleteWithoutTimeout: resourceDocumentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attachments_source": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 20,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKey: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.AttachmentsSourceKey](),
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.All(
								validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+$`), "must contain only alphanumeric, underscore, hyphen, or period characters"),
								validation.StringLenBetween(3, 128),
							),
						},
						names.AttrValues: {
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
			names.AttrContent: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrCreatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"document_format": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.DocumentFormatJson,
				ValidateDiagFunc: enum.Validate[awstypes.DocumentFormat](),
			},
			"document_type": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.DocumentType](),
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
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+$`), "must contain only alphanumeric, underscore, hyphen, or period characters"),
					validation.StringLenBetween(3, 128),
				),
			},
			names.AttrOwner: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrParameter: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDefaultValue: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrDescription: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrPermissions: {
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
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"target_type": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexache.MustCompile(`^\/[\w\.\-\:\/]*$`), "must contain a forward slash optionally followed by a resource type such as AWS::EC2::Instance"),
					validation.StringLenBetween(1, 200),
				),
			},
			"version_name": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]{3,128}$`), "must contain only alphanumeric, underscore, hyphen, or period characters"),
					validation.StringLenBetween(3, 128),
				),
			},
		},

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
			func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
				if v, ok := d.GetOk(names.AttrPermissions); ok && len(v.(map[string]interface{})) > 0 {
					// Validates permissions keys, if set, to be type and account_ids
					// since ValidateFunc validates only the value not the key.
					tfMap := flex.ExpandStringValueMap(v.(map[string]interface{}))

					if v, ok := tfMap[names.AttrType]; ok {
						if awstypes.DocumentPermissionType(v) != awstypes.DocumentPermissionTypeShare {
							return fmt.Errorf("%q: only %s \"type\" supported", names.AttrPermissions, awstypes.DocumentPermissionTypeShare)
						}
					} else {
						return fmt.Errorf("%q: \"type\" must be defined", names.AttrPermissions)
					}

					if _, ok := tfMap["account_ids"]; !ok {
						return fmt.Errorf("%q: \"account_ids\" must be defined", names.AttrPermissions)
					}
				}

				if d.HasChange(names.AttrContent) {
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
					if err := d.SetNewComputed(names.AttrParameter); err != nil {
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
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &ssm.CreateDocumentInput{
		Content:        aws.String(d.Get(names.AttrContent).(string)),
		DocumentFormat: awstypes.DocumentFormat(d.Get("document_format").(string)),
		DocumentType:   awstypes.DocumentType(d.Get("document_type").(string)),
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

	output, err := conn.CreateDocument(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SSM Document (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.DocumentDescription.Name))

	if v, ok := d.GetOk(names.AttrPermissions); ok && len(v.(map[string]interface{})) > 0 {
		tfMap := flex.ExpandStringValueMap(v.(map[string]interface{}))

		if v, ok := tfMap["account_ids"]; ok && v != "" {
			chunks := tfslices.Chunks(strings.Split(v, ","), documentPermissionsBatchLimit)

			for _, chunk := range chunks {
				input := &ssm.ModifyDocumentPermissionInput{
					AccountIdsToAdd: chunk,
					Name:            aws.String(d.Id()),
					PermissionType:  awstypes.DocumentPermissionTypeShare,
				}

				_, err := conn.ModifyDocumentPermission(ctx, input)

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
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	doc, err := findDocumentByName(ctx, conn, d.Id())

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
		Resource:  "document/" + aws.ToString(doc.Name),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrCreatedDate, aws.ToTime(doc.CreatedDate).Format(time.RFC3339))
	d.Set("default_version", doc.DefaultVersion)
	d.Set(names.AttrDescription, doc.Description)
	d.Set("document_format", doc.DocumentFormat)
	d.Set("document_type", doc.DocumentType)
	d.Set("document_version", doc.DocumentVersion)
	d.Set("hash", doc.Hash)
	d.Set("hash_type", doc.HashType)
	d.Set("latest_version", doc.LatestVersion)
	d.Set(names.AttrName, doc.Name)
	d.Set(names.AttrOwner, doc.Owner)
	if err := d.Set(names.AttrParameter, flattenDocumentParameters(doc.Parameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameter: %s", err)
	}
	d.Set("platform_types", doc.PlatformTypes)
	d.Set("schema_version", doc.SchemaVersion)
	d.Set(names.AttrStatus, doc.Status)
	d.Set("target_type", doc.TargetType)
	d.Set("version_name", doc.VersionName)

	{
		input := &ssm.GetDocumentInput{
			DocumentFormat:  awstypes.DocumentFormat(d.Get("document_format").(string)),
			DocumentVersion: aws.String("$LATEST"),
			Name:            aws.String(d.Id()),
		}

		output, err := conn.GetDocument(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading SSM Document (%s) content: %s", d.Id(), err)
		}

		d.Set(names.AttrContent, output.Content)
	}

	{
		input := &ssm.DescribeDocumentPermissionInput{
			Name:           aws.String(d.Id()),
			PermissionType: awstypes.DocumentPermissionTypeShare,
		}

		output, err := conn.DescribeDocumentPermission(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading SSM Document (%s) permissions: %s", d.Id(), err)
		}

		if accountsIDs := output.AccountIds; len(accountsIDs) > 0 {
			d.Set(names.AttrPermissions, map[string]interface{}{
				"account_ids":  strings.Join(accountsIDs, ","),
				names.AttrType: awstypes.DocumentPermissionTypeShare,
			})
		} else {
			d.Set(names.AttrPermissions, nil)
		}
	}

	setTagsOut(ctx, doc.Tags)

	return diags
}

func resourceDocumentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	if d.HasChange(names.AttrPermissions) {
		var oldAccountIDs, newAccountIDs itypes.Set[string]
		o, n := d.GetChange(names.AttrPermissions)

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

		for _, chunk := range tfslices.Chunks(newAccountIDs.Difference(oldAccountIDs), documentPermissionsBatchLimit) {
			input := &ssm.ModifyDocumentPermissionInput{
				AccountIdsToAdd: chunk,
				Name:            aws.String(d.Id()),
				PermissionType:  awstypes.DocumentPermissionTypeShare,
			}

			_, err := conn.ModifyDocumentPermission(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "modifying SSM Document (%s) permissions: %s", d.Id(), err)
			}
		}

		for _, chunk := range tfslices.Chunks(oldAccountIDs.Difference(newAccountIDs), documentPermissionsBatchLimit) {
			input := &ssm.ModifyDocumentPermissionInput{
				AccountIdsToRemove: chunk,
				Name:               aws.String(d.Id()),
				PermissionType:     awstypes.DocumentPermissionTypeShare,
			}

			_, err := conn.ModifyDocumentPermission(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "modifying SSM Document (%s) permissions: %s", d.Id(), err)
			}
		}
	}

	if d.HasChangesExcept(names.AttrPermissions, names.AttrTags, names.AttrTagsAll) {
		// Update for schema version 1.x is not allowed.
		isSchemaVersion1, _ := regexp.MatchString(`^1[.][0-9]$`, d.Get("schema_version").(string))

		if d.HasChange(names.AttrContent) || !isSchemaVersion1 {
			input := &ssm.UpdateDocumentInput{
				Content:         aws.String(d.Get(names.AttrContent).(string)),
				DocumentFormat:  awstypes.DocumentFormat(d.Get("document_format").(string)),
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

			output, err := conn.UpdateDocument(ctx, input)

			if errs.IsA[*awstypes.DuplicateDocumentContent](err) {
				defaultVersion = d.Get("latest_version").(string)
			} else if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating SSM Document (%s): %s", d.Id(), err)
			} else {
				defaultVersion = aws.ToString(output.DocumentDescription.DocumentVersion)
			}

			_, err = conn.UpdateDocumentDefaultVersion(ctx, &ssm.UpdateDocumentDefaultVersionInput{
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
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	if v, ok := d.GetOk(names.AttrPermissions); ok && len(v.(map[string]interface{})) > 0 {
		tfMap := flex.ExpandStringValueMap(v.(map[string]interface{}))

		if v, ok := tfMap["account_ids"]; ok && v != "" {
			chunks := tfslices.Chunks(strings.Split(v, ","), documentPermissionsBatchLimit)

			for _, chunk := range chunks {
				input := &ssm.ModifyDocumentPermissionInput{
					AccountIdsToRemove: chunk,
					Name:               aws.String(d.Id()),
					PermissionType:     awstypes.DocumentPermissionTypeShare,
				}

				_, err := conn.ModifyDocumentPermission(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "modifying SSM Document (%s) permissions: %s", d.Id(), err)
				}
			}
		}
	}

	log.Printf("[INFO] Deleting SSM Document: %s", d.Id())
	_, err := conn.DeleteDocument(ctx, &ssm.DeleteDocumentInput{
		Name: aws.String(d.Get(names.AttrName).(string)),
	})

	if errs.IsAErrorMessageContains[*awstypes.InvalidDocument](err, "does not exist") {
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

func findDocumentByName(ctx context.Context, conn *ssm.Client, name string) (*awstypes.DocumentDescription, error) {
	input := &ssm.DescribeDocumentInput{
		Name: aws.String(name),
	}

	output, err := conn.DescribeDocument(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.InvalidDocument](err, "does not exist") {
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

func statusDocument(ctx context.Context, conn *ssm.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDocumentByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitDocumentActive(ctx context.Context, conn *ssm.Client, name string) (*awstypes.DocumentDescription, error) { //nolint:unparam
	const (
		timeout = 2 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DocumentStatusCreating, awstypes.DocumentStatusUpdating),
		Target:  enum.Slice(awstypes.DocumentStatusActive),
		Refresh: statusDocument(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DocumentDescription); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusInformation)))

		return output, err
	}

	return nil, err
}

func waitDocumentDeleted(ctx context.Context, conn *ssm.Client, name string) (*awstypes.DocumentDescription, error) {
	const (
		timeout = 2 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DocumentStatusDeleting),
		Target:  []string{},
		Refresh: statusDocument(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DocumentDescription); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusInformation)))

		return output, err
	}

	return nil, err
}

func expandAttachmentsSource(tfMap map[string]interface{}) *awstypes.AttachmentsSource {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.AttachmentsSource{}

	if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
		apiObject.Key = awstypes.AttachmentsSourceKey(v)
	}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap[names.AttrValues].([]interface{}); ok && len(v) > 0 {
		apiObject.Values = flex.ExpandStringValueList(v)
	}

	return apiObject
}

func expandAttachmentsSources(tfList []interface{}) []awstypes.AttachmentsSource {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.AttachmentsSource

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandAttachmentsSource(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func flattenDocumentParameter(apiObject *awstypes.DocumentParameter) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrType: apiObject.Type,
	}

	if v := apiObject.DefaultValue; v != nil {
		tfMap[names.AttrDefaultValue] = aws.ToString(v)
	}

	if v := apiObject.Description; v != nil {
		tfMap[names.AttrDescription] = aws.ToString(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	return tfMap
}

func flattenDocumentParameters(apiObjects []awstypes.DocumentParameter) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenDocumentParameter(&apiObject))
	}

	return tfList
}
