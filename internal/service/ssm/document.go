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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

// @SDKResource("aws_ssm_document")
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
							Optional: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
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

				return nil
			},
		),
	}
}

func resourceDocumentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(ctx, d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &ssm.CreateDocumentInput{
		Content:        aws.String(d.Get("content").(string)),
		DocumentFormat: aws.String(d.Get("document_format").(string)),
		DocumentType:   aws.String(d.Get("document_type").(string)),
		Name:           aws.String(name),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	if v, ok := d.GetOk("attachments_source"); ok {
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
	conn := meta.(*conns.AWSClient).SSMConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

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

	content, err := findDocumentContentByThreePartKey(ctx, conn, d.Id(), "$LATEST", aws.StringValue(doc.DocumentFormat))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSM Document (%s) content: %s", d.Id(), err)
	}

	d.Set("content", content)

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

	tags := KeyValueTags(ctx, doc.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceDocumentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), ssm.ResourceTypeForTaggingDocument, o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SSM Document (%s) tags: %s", d.Id(), err)
		}
	}

	if d.HasChange("permissions") {
		diags = append(diags, sdkdiag.WrapDiagsf(setDocumentPermissions(ctx, d, meta), "updating SSM Document (%s): setting permissions", d.Id())...)
		if diags.HasError() {
			return diags
		}
	} else {
		log.Printf("[DEBUG] Not setting document permissions on %q", d.Id())
	}

	// update for schema version 1.x is not allowed
	isSchemaVersion1, _ := regexp.MatchString("^1[.][0-9]$", d.Get("schema_version").(string))

	if !d.HasChange("content") && isSchemaVersion1 {
		return diags
	}

	if d.HasChangesExcept("tags", "tags_all", "permissions") {
		diags = append(diags, sdkdiag.WrapDiagsf(updateDocument(ctx, d, meta), "updating SSM Document (%s)", d.Id())...)
		if diags.HasError() {
			return diags
		}

		_, err := waitDocumentActive(ctx, conn, d.Id())
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for SSM Document (%s) to be Active: %s", d.Id(), err)
		}
	}

	return append(diags, resourceDocumentRead(ctx, d, meta)...)
}

func resourceDocumentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn()

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

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SSM Document (%s): %s", d.Id(), err)
	}

	if tfawserr.ErrMessageContains(err, ssm.ErrCodeInvalidDocument, "does not exist") {
		return diags
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
		return nil, &resource.NotFoundError{
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

func findDocumentContentByThreePartKey(ctx context.Context, conn *ssm.SSM, name, version, format string) (string, error) {
	input := &ssm.GetDocumentInput{
		DocumentFormat:  aws.String(format),
		DocumentVersion: aws.String(version),
		Name:            aws.String(name),
	}

	output, err := conn.GetDocumentWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, ssm.ErrCodeInvalidDocument, "does not exist") {
		return "", &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return "", err
	}

	if output == nil || output.Content == nil {
		return "", tfresource.NewEmptyResultError(input)
	}

	return aws.StringValue(output.Content), nil
}

func statusDocument(ctx context.Context, conn *ssm.SSM, name string) resource.StateRefreshFunc {
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
	stateConf := &resource.StateChangeConf{
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
	stateConf := &resource.StateChangeConf{
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

func setDocumentPermissions(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn()

	log.Printf("[INFO] Setting permissions for document: %s", d.Id())

	if d.HasChange("permissions") {
		o, n := d.GetChange("permissions")
		oldPermissions := o.(map[string]interface{})
		newPermissions := n.(map[string]interface{})
		oldPermissionsAccountIds := make([]interface{}, 0)
		if v, ok := oldPermissions["account_ids"]; ok && v.(string) != "" {
			parts := strings.Split(v.(string), ",")
			oldPermissionsAccountIds = make([]interface{}, len(parts))
			for i, v := range parts {
				oldPermissionsAccountIds[i] = v
			}
		}
		newPermissionsAccountIds := make([]interface{}, 0)
		if v, ok := newPermissions["account_ids"]; ok && v.(string) != "" {
			parts := strings.Split(v.(string), ",")
			newPermissionsAccountIds = make([]interface{}, len(parts))
			for i, v := range parts {
				newPermissionsAccountIds[i] = v
			}
		}

		// Since AccountIdsToRemove has higher priority than AccountIdsToAdd,
		// we filter out accounts from both lists
		accountIdsToRemove := make([]interface{}, 0)
		for _, oldPermissionsAccountId := range oldPermissionsAccountIds {
			if _, contains := verify.SliceContainsString(newPermissionsAccountIds, oldPermissionsAccountId.(string)); !contains {
				accountIdsToRemove = append(accountIdsToRemove, oldPermissionsAccountId.(string))
			}
		}
		accountIdsToAdd := make([]interface{}, 0)
		for _, newPermissionsAccountId := range newPermissionsAccountIds {
			if _, contains := verify.SliceContainsString(oldPermissionsAccountIds, newPermissionsAccountId.(string)); !contains {
				accountIdsToAdd = append(accountIdsToAdd, newPermissionsAccountId.(string))
			}
		}

		if err := modifyDocumentPermissions(ctx, conn, d.Get("name").(string), accountIdsToAdd, accountIdsToRemove); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return diags
}

func getDocumentPermissions(ctx context.Context, d *schema.ResourceData, meta interface{}) (map[string]interface{}, error) {
	conn := meta.(*conns.AWSClient).SSMConn()

	log.Printf("[INFO] Getting permissions for document: %s", d.Id())

	//How to get from nested scheme resource?
	permissionType := "Share"

	permInput := &ssm.DescribeDocumentPermissionInput{
		Name:           aws.String(d.Get("name").(string)),
		PermissionType: aws.String(permissionType),
	}

	resp, err := conn.DescribeDocumentPermissionWithContext(ctx, permInput)

	if err != nil {
		return nil, fmt.Errorf("Error setting permissions for SSM document: %s", err)
	}

	ids := ""
	accountIds := aws.StringValueSlice(resp.AccountIds)

	if len(accountIds) == 1 {
		ids = accountIds[0]
	} else if len(accountIds) > 1 {
		ids = strings.Join(accountIds, ",")
	}

	if ids == "" {
		return nil, nil
	}

	perms := make(map[string]interface{})
	perms["type"] = permissionType
	perms["account_ids"] = ids

	return perms, nil
}

func deleteDocumentPermissions(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn()

	log.Printf("[INFO] Removing permissions from document: %s", d.Id())

	permission := d.Get("permissions").(map[string]interface{})

	accountIdsToRemove := make([]interface{}, 0)

	if permission["account_ids"] != nil {
		if v, ok := permission["account_ids"]; ok && v.(string) != "" {
			parts := strings.Split(v.(string), ",")
			accountIdsToRemove = make([]interface{}, len(parts))
			for i, v := range parts {
				accountIdsToRemove[i] = v
			}
		}

		if err := modifyDocumentPermissions(ctx, conn, d.Get("name").(string), nil, accountIdsToRemove); err != nil {
			return sdkdiag.AppendErrorf(diags, "removing SSM document permissions: %s", err)
		}
	}

	return diags
}

func modifyDocumentPermissions(ctx context.Context, conn *ssm.SSM, name string, accountIdsToAdd []interface{}, accountIdstoRemove []interface{}) error {
	if accountIdsToAdd != nil {
		accountIdsToAddBatch := make([]string, 0, documentPermissionsBatchLimit)
		accountIdsToAddBatches := make([][]string, 0, len(accountIdsToAdd)/documentPermissionsBatchLimit+1)
		for _, accountId := range accountIdsToAdd {
			if len(accountIdsToAddBatch) == documentPermissionsBatchLimit {
				accountIdsToAddBatches = append(accountIdsToAddBatches, accountIdsToAddBatch)
				accountIdsToAddBatch = make([]string, 0, documentPermissionsBatchLimit)
			}
			accountIdsToAddBatch = append(accountIdsToAddBatch, accountId.(string))
		}
		accountIdsToAddBatches = append(accountIdsToAddBatches, accountIdsToAddBatch)

		for _, accountIdsToAdd := range accountIdsToAddBatches {
			_, err := conn.ModifyDocumentPermissionWithContext(ctx, &ssm.ModifyDocumentPermissionInput{
				Name:            aws.String(name),
				PermissionType:  aws.String("Share"),
				AccountIdsToAdd: aws.StringSlice(accountIdsToAdd),
			})
			if err != nil {
				return err
			}
		}
	}

	if accountIdstoRemove != nil {
		accountIdsToRemoveBatch := make([]string, 0, documentPermissionsBatchLimit)
		accountIdsToRemoveBatches := make([][]string, 0, len(accountIdstoRemove)/documentPermissionsBatchLimit+1)
		for _, accountId := range accountIdstoRemove {
			if len(accountIdsToRemoveBatch) == documentPermissionsBatchLimit {
				accountIdsToRemoveBatches = append(accountIdsToRemoveBatches, accountIdsToRemoveBatch)
				accountIdsToRemoveBatch = make([]string, 0, documentPermissionsBatchLimit)
			}
			accountIdsToRemoveBatch = append(accountIdsToRemoveBatch, accountId.(string))
		}
		accountIdsToRemoveBatches = append(accountIdsToRemoveBatches, accountIdsToRemoveBatch)

		for _, accountIdsToRemove := range accountIdsToRemoveBatches {
			_, err := conn.ModifyDocumentPermissionWithContext(ctx, &ssm.ModifyDocumentPermissionInput{
				Name:               aws.String(name),
				PermissionType:     aws.String("Share"),
				AccountIdsToRemove: aws.StringSlice(accountIdsToRemove),
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func updateDocument(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("[INFO] Updating SSM Document: %s", d.Id())

	name := d.Get("name").(string)

	updateDocInput := &ssm.UpdateDocumentInput{
		Name:            aws.String(name),
		Content:         aws.String(d.Get("content").(string)),
		DocumentFormat:  aws.String(d.Get("document_format").(string)),
		DocumentVersion: aws.String(d.Get("default_version").(string)),
	}

	if v, ok := d.GetOk("target_type"); ok {
		updateDocInput.TargetType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("version_name"); ok {
		updateDocInput.VersionName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("attachments_source"); ok {
		updateDocInput.Attachments = expandAttachmentsSources(v.([]interface{}))
	}

	newDefaultVersion := d.Get("default_version").(string)

	conn := meta.(*conns.AWSClient).SSMConn()
	updated, err := conn.UpdateDocumentWithContext(ctx, updateDocInput)

	if tfawserr.ErrCodeEquals(err, ssm.ErrCodeDuplicateDocumentContent) {
		log.Printf("[DEBUG] Content is a duplicate of the latest version so update is not necessary: %s", d.Id())
		log.Printf("[INFO] Updating the default version to the latest version %s: %s", newDefaultVersion, d.Id())

		newDefaultVersion = d.Get("latest_version").(string)
	} else if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating SSM document: %s", err)
	} else {
		log.Printf("[INFO] Updating the default version to the new version %s: %s", newDefaultVersion, d.Id())
		newDefaultVersion = aws.StringValue(updated.DocumentDescription.DocumentVersion)
	}

	updateDefaultInput := &ssm.UpdateDocumentDefaultVersionInput{
		Name:            aws.String(name),
		DocumentVersion: aws.String(newDefaultVersion),
	}

	_, err = conn.UpdateDocumentDefaultVersionWithContext(ctx, updateDefaultInput)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating the default document version to that of the updated document: %s", err)
	}
	return diags
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
