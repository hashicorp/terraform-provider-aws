// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codegurureviewer

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codegurureviewer"
	"github.com/aws/aws-sdk-go-v2/service/codegurureviewer/types"
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

// @SDKResource("aws_codegurureviewer_repository_association", name="Repository Association")
// @Tags(identifierAttribute="id")
func resourceRepositoryAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRepositoryAssociationCreate,
		ReadWithoutTimeout:   resourceRepositoryAssociationRead,
		UpdateWithoutTimeout: resourceRepositoryAssociationUpdate,
		DeleteWithoutTimeout: resourceRepositoryAssociationDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAssociationID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kms_key_details": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Show difference for new resources
					if d.Id() == "" {
						return false
					}
					// Show difference if existing state reflects different default type
					_, defaultEncryptionOption := d.GetChange("kms_key_details.0.encryption_option")
					if defaultEncryptionOption := types.EncryptionOption(defaultEncryptionOption.(string)); defaultEncryptionOption != types.EncryptionOptionAoCmk {
						return defaultEncryptionOption == types.EncryptionOptionAoCmk
					}
					return true
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encryption_option": {
							Type:             schema.TypeString,
							ForceNew:         true,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.EncryptionOption](),
						},
						names.AttrKMSKeyID: {
							Type:     schema.TypeString,
							ForceNew: true,
							Optional: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 2048),
								validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z-]+`), ""),
							),
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrOwner: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provider_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"repository": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bitbucket": {
							Type:     schema.TypeList,
							ForceNew: true,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"connection_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									names.AttrName: {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 100),
											validation.StringMatch(regexache.MustCompile(`^\S[\w.-]*$`), ""),
										),
									},
									names.AttrOwner: {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 100),
											validation.StringMatch(regexache.MustCompile(`^\S(.*\S)?$`), ""),
										),
									},
								},
							},
						},
						"codecommit": {
							Type:     schema.TypeList,
							ForceNew: true,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 100),
											validation.StringMatch(regexache.MustCompile(`^\S[\w.-]*$`), ""),
										),
									},
								},
							},
						},
						"github_enterprise_server": {
							Type:     schema.TypeList,
							ForceNew: true,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"connection_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									names.AttrName: {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 100),
											validation.StringMatch(regexache.MustCompile(`^\S[\w.-]*$`), ""),
										),
									},
									names.AttrOwner: {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 100),
											validation.StringMatch(regexache.MustCompile(`^\S(.*\S)?$`), ""),
										),
									},
								},
							},
						},
						names.AttrS3Bucket: {
							Type:     schema.TypeList,
							ForceNew: true,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrBucketName: {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 63),
											validation.StringMatch(regexache.MustCompile(`^\S(.*\S)?$`), ""),
										),
									},
									names.AttrName: {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 100),
											validation.StringMatch(regexache.MustCompile(`^\S[\w.-]*$`), ""),
										),
									},
								},
							},
						},
					},
				},
			},
			"s3_repository_details": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrBucketName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"code_artifacts": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"build_artifacts_object_key": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"source_code_artifacts_object_key": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRepositoryAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeGuruReviewerClient(ctx)

	input := &codegurureviewer.AssociateRepositoryInput{
		Tags: getTagsIn(ctx),
	}

	input.KMSKeyDetails = expandKMSKeyDetails(d.Get("kms_key_details").([]interface{}))

	if v, ok := d.GetOk("repository"); ok {
		input.Repository = expandRepository(v.([]interface{}))
	}

	output, err := conn.AssociateRepository(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeGuru Repository Association: %s", err)
	}

	d.SetId(aws.ToString(output.RepositoryAssociation.AssociationArn))

	if _, err := waitRepositoryAssociationCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CodeGuru Repository Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceRepositoryAssociationRead(ctx, d, meta)...)
}

func resourceRepositoryAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeGuruReviewerClient(ctx)

	out, err := findRepositoryAssociationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodeGuru Reviewer Repository Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeGuru Repository Association (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, out.AssociationArn)
	d.Set(names.AttrAssociationID, out.AssociationId)
	d.Set("connection_arn", out.ConnectionArn)
	if err := d.Set("kms_key_details", flattenKMSKeyDetails(out.KMSKeyDetails)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting kms_key_details: %s", err)
	}
	d.Set(names.AttrName, out.Name)
	d.Set(names.AttrOwner, out.Owner)
	d.Set("provider_type", out.ProviderType)
	if err := d.Set("s3_repository_details", flattenS3RepositoryDetails(out.S3RepositoryDetails)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting s3_repository_details: %s", err)
	}
	d.Set(names.AttrState, out.State)
	d.Set("state_reason", out.StateReason)

	return diags
}

func resourceRepositoryAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceRepositoryAssociationRead(ctx, d, meta)...)
}

func resourceRepositoryAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeGuruReviewerClient(ctx)

	log.Printf("[INFO] Deleting CodeGuru Repository Association %s", d.Id())
	_, err := conn.DisassociateRepository(ctx, &codegurureviewer.DisassociateRepositoryInput{
		AssociationArn: aws.String(d.Id()),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeGuru Repository Association (%s): %s", d.Id(), err)
	}

	if _, err := waitRepositoryAssociationDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CodeGuru Repository Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findRepositoryAssociationByID(ctx context.Context, conn *codegurureviewer.Client, id string) (*types.RepositoryAssociation, error) {
	input := &codegurureviewer.DescribeRepositoryAssociationInput{
		AssociationArn: aws.String(id),
	}

	output, err := conn.DescribeRepositoryAssociation(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.RepositoryAssociation == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.RepositoryAssociation, nil
}

func statusRepositoryAssociation(ctx context.Context, conn *codegurureviewer.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findRepositoryAssociationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func waitRepositoryAssociationCreated(ctx context.Context, conn *codegurureviewer.Client, id string, timeout time.Duration) (*types.RepositoryAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.RepositoryAssociationStateAssociating),
		Target:                    enum.Slice(types.RepositoryAssociationStateAssociated),
		Refresh:                   statusRepositoryAssociation(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.RepositoryAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StateReason)))

		return output, err
	}

	return nil, err
}

func waitRepositoryAssociationDeleted(ctx context.Context, conn *codegurureviewer.Client, id string, timeout time.Duration) (*types.RepositoryAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.RepositoryAssociationStateDisassociating, types.RepositoryAssociationStateAssociated),
		Target:  []string{},
		Refresh: statusRepositoryAssociation(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.RepositoryAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StateReason)))

		return output, err
	}

	return nil, err
}

func flattenKMSKeyDetails(kmsKeyDetails *types.KMSKeyDetails) []interface{} {
	if kmsKeyDetails == nil {
		return nil
	}

	values := map[string]interface{}{
		"encryption_option": kmsKeyDetails.EncryptionOption,
	}

	if v := kmsKeyDetails.KMSKeyId; v != nil {
		values[names.AttrKMSKeyID] = aws.ToString(v)
	}

	return []interface{}{values}
}

func flattenS3RepositoryDetails(s3RepositoryDetails *types.S3RepositoryDetails) []interface{} {
	if s3RepositoryDetails == nil {
		return nil
	}

	values := map[string]interface{}{}

	if v := s3RepositoryDetails.BucketName; v != nil {
		values[names.AttrBucketName] = aws.ToString(v)
	}

	if v := s3RepositoryDetails.CodeArtifacts; v != nil {
		values["code_artifacts"] = flattenCodeArtifacts(v)
	}

	return []interface{}{values}
}

func flattenCodeArtifacts(apiObject *types.CodeArtifacts) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.BuildArtifactsObjectKey; v != nil {
		m["build_artifacts_object_key"] = aws.ToString(v)
	}

	if v := apiObject.SourceCodeArtifactsObjectKey; v != nil {
		m["source_code_artifacts_object_key"] = aws.ToString(v)
	}

	return m
}

func expandKMSKeyDetails(kmsKeyDetails []interface{}) *types.KMSKeyDetails {
	if len(kmsKeyDetails) == 0 || kmsKeyDetails[0] == nil {
		return nil
	}

	tfMap, ok := kmsKeyDetails[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.KMSKeyDetails{}

	if v, ok := tfMap["encryption_option"].(string); ok && v != "" {
		result.EncryptionOption = types.EncryptionOption(v)
	}

	if v, ok := tfMap[names.AttrKMSKeyID].(string); ok && v != "" {
		result.KMSKeyId = aws.String(v)
	}

	return result
}

func expandCodeCommitRepository(repository []interface{}) *types.CodeCommitRepository {
	if len(repository) == 0 || repository[0] == nil {
		return nil
	}

	tfMap, ok := repository[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.CodeCommitRepository{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		result.Name = aws.String(v)
	}

	return result
}

func expandRepository(repository []interface{}) *types.Repository {
	if len(repository) == 0 || repository[0] == nil {
		return nil
	}

	tfMap, ok := repository[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.Repository{}

	if v, ok := tfMap["bitbucket"]; ok {
		result.Bitbucket = expandThirdPartySourceRepository(v.([]interface{}))
	}
	if v, ok := tfMap["codecommit"]; ok {
		result.CodeCommit = expandCodeCommitRepository(v.([]interface{}))
	}
	if v, ok := tfMap["github_enterprise_server"]; ok {
		result.GitHubEnterpriseServer = expandThirdPartySourceRepository(v.([]interface{}))
	}
	if v, ok := tfMap[names.AttrS3Bucket]; ok {
		result.S3Bucket = expandS3Repository(v.([]interface{}))
	}

	return result
}

func expandS3Repository(repository []interface{}) *types.S3Repository {
	if len(repository) == 0 || repository[0] == nil {
		return nil
	}

	tfMap, ok := repository[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.S3Repository{}

	if v, ok := tfMap[names.AttrBucketName].(string); ok && v != "" {
		result.BucketName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		result.Name = aws.String(v)
	}

	return result
}

func expandThirdPartySourceRepository(repository []interface{}) *types.ThirdPartySourceRepository {
	if len(repository) == 0 || repository[0] == nil {
		return nil
	}

	tfMap, ok := repository[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.ThirdPartySourceRepository{}

	if v, ok := tfMap["connection_arn"].(string); ok && v != "" {
		result.ConnectionArn = aws.String(v)
	}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		result.Name = aws.String(v)
	}

	if v, ok := tfMap[names.AttrOwner].(string); ok && v != "" {
		result.Owner = aws.String(v)
	}

	return result
}
