// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codegurureviewer

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codegurureviewer"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_codegurureviewer_repository_association", name="Repository Association")
// @Tags(identifierAttribute="id")
func ResourceRepositoryAssociation() *schema.Resource {
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"association_id": {
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
					if defaultEncryptionOption.(string) != codegurureviewer.EncryptionOptionAwsOwnedCmk {
						return defaultEncryptionOption.(string) == codegurureviewer.EncryptionOptionAwsOwnedCmk
					}
					return true
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encryption_option": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(codegurureviewer.EncryptionOption_Values(), false),
						},
						"kms_key_id": {
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
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner": {
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
									"name": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 100),
											validation.StringMatch(regexache.MustCompile(`^\S[\w.-]*$`), ""),
										),
									},
									"owner": {
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
									"name": {
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
									"name": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 100),
											validation.StringMatch(regexache.MustCompile(`^\S[\w.-]*$`), ""),
										),
									},
									"owner": {
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
						"s3_bucket": {
							Type:     schema.TypeList,
							ForceNew: true,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bucket_name": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 63),
											validation.StringMatch(regexache.MustCompile(`^\S(.*\S)?$`), ""),
										),
									},
									"name": {
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
						"bucket_name": {
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
			"state": {
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
		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
		),
	}
}

const (
	ResNameRepositoryAssociation = "RepositoryAssociation"
)

func resourceRepositoryAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodeGuruReviewerConn(ctx)

	in := &codegurureviewer.AssociateRepositoryInput{
		Tags: getTagsIn(ctx),
	}

	in.KMSKeyDetails = expandKMSKeyDetails(d.Get("kms_key_details").([]interface{}))

	if v, ok := d.GetOk("repository"); ok {
		in.Repository = expandRepository(v.([]interface{}))
	}

	out, err := conn.AssociateRepositoryWithContext(ctx, in)

	if err != nil {
		return create.AppendDiagError(diags, names.CodeGuruReviewer, create.ErrActionCreating, ResNameRepositoryAssociation, d.Get("name").(string), err)
	}

	if out == nil || out.RepositoryAssociation == nil {
		return create.AppendDiagError(diags, names.CodeGuruReviewer, create.ErrActionCreating, ResNameRepositoryAssociation, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.StringValue(out.RepositoryAssociation.AssociationArn))

	if _, err := waitRepositoryAssociationCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.CodeGuruReviewer, create.ErrActionWaitingForCreation, ResNameRepositoryAssociation, d.Id(), err)
	}

	return append(diags, resourceRepositoryAssociationRead(ctx, d, meta)...)
}

func resourceRepositoryAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodeGuruReviewerConn(ctx)

	out, err := findRepositoryAssociationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodeGuruReviewer RepositoryAssociation (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return create.AppendDiagError(diags, names.CodeGuruReviewer, create.ErrActionReading, ResNameRepositoryAssociation, d.Id(), err)
	}

	d.Set("arn", out.AssociationArn)
	d.Set("association_id", out.AssociationId)
	d.Set("connection_arn", out.ConnectionArn)

	if err := d.Set("kms_key_details", flattenKMSKeyDetails(out.KMSKeyDetails)); err != nil {
		return create.AppendDiagError(diags, names.CodeGuruReviewer, create.ErrActionSetting, ResNameRepositoryAssociation, d.Id(), err)
	}

	d.Set("name", out.Name)
	d.Set("owner", out.Owner)
	d.Set("provider_type", out.ProviderType)

	if err := d.Set("s3_repository_details", flattenS3RepositoryDetails(out.S3RepositoryDetails)); err != nil {
		return create.AppendDiagError(diags, names.CodeGuruReviewer, create.ErrActionSetting, ResNameRepositoryAssociation, d.Id(), err)
	}

	d.Set("state", out.State)
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

	conn := meta.(*conns.AWSClient).CodeGuruReviewerConn(ctx)

	log.Printf("[INFO] Deleting CodeGuruReviewer RepositoryAssociation %s", d.Id())

	_, err := conn.DisassociateRepositoryWithContext(ctx, &codegurureviewer.DisassociateRepositoryInput{
		AssociationArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, codegurureviewer.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.CodeGuruReviewer, create.ErrActionDeleting, ResNameRepositoryAssociation, d.Id(), err)
	}

	if _, err := waitRepositoryAssociationDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.AppendDiagError(diags, names.CodeGuruReviewer, create.ErrActionWaitingForDeletion, ResNameRepositoryAssociation, d.Id(), err)
	}

	return diags
}

func waitRepositoryAssociationCreated(ctx context.Context, conn *codegurureviewer.CodeGuruReviewer, id string, timeout time.Duration) (*codegurureviewer.RepositoryAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{codegurureviewer.RepositoryAssociationStateAssociating},
		Target:                    []string{codegurureviewer.RepositoryAssociationStateAssociated},
		Refresh:                   statusRepositoryAssociation(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*codegurureviewer.RepositoryAssociation); ok {
		return out, err
	}

	return nil, err
}

func waitRepositoryAssociationDeleted(ctx context.Context, conn *codegurureviewer.CodeGuruReviewer, id string, timeout time.Duration) (*codegurureviewer.RepositoryAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{codegurureviewer.RepositoryAssociationStateDisassociating, codegurureviewer.RepositoryAssociationStateAssociated},
		Target:  []string{},
		Refresh: statusRepositoryAssociation(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*codegurureviewer.RepositoryAssociation); ok {
		return out, err
	}

	return nil, err
}

func statusRepositoryAssociation(ctx context.Context, conn *codegurureviewer.CodeGuruReviewer, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findRepositoryAssociationByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.StringValue(out.State), nil
	}
}

func findRepositoryAssociationByID(ctx context.Context, conn *codegurureviewer.CodeGuruReviewer, id string) (*codegurureviewer.RepositoryAssociation, error) {
	in := &codegurureviewer.DescribeRepositoryAssociationInput{
		AssociationArn: aws.String(id),
	}
	out, err := conn.DescribeRepositoryAssociationWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, codegurureviewer.ErrCodeNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.RepositoryAssociation == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.RepositoryAssociation, nil
}

func flattenKMSKeyDetails(kmsKeyDetails *codegurureviewer.KMSKeyDetails) []interface{} {
	if kmsKeyDetails == nil {
		return nil
	}

	values := map[string]interface{}{}

	if v := kmsKeyDetails.EncryptionOption; v != nil {
		values["encryption_option"] = aws.StringValue(v)
	}

	if v := kmsKeyDetails.KMSKeyId; v != nil {
		values["kms_key_id"] = aws.StringValue(v)
	}

	return []interface{}{values}
}

func flattenS3RepositoryDetails(s3RepositoryDetails *codegurureviewer.S3RepositoryDetails) []interface{} {
	if s3RepositoryDetails == nil {
		return nil
	}

	values := map[string]interface{}{}

	if v := s3RepositoryDetails.BucketName; v != nil {
		values["bucket_name"] = aws.StringValue(v)
	}

	if v := s3RepositoryDetails.CodeArtifacts; v != nil {
		values["code_artifacts"] = flattenCodeArtifacts(v)
	}

	return []interface{}{values}
}

func flattenCodeArtifacts(apiObject *codegurureviewer.CodeArtifacts) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.BuildArtifactsObjectKey; v != nil {
		m["build_artifacts_object_key"] = aws.StringValue(v)
	}

	if v := apiObject.SourceCodeArtifactsObjectKey; v != nil {
		m["source_code_artifacts_object_key"] = aws.StringValue(v)
	}

	return m
}

func expandKMSKeyDetails(kmsKeyDetails []interface{}) *codegurureviewer.KMSKeyDetails {
	if len(kmsKeyDetails) == 0 || kmsKeyDetails[0] == nil {
		return nil
	}

	tfMap, ok := kmsKeyDetails[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &codegurureviewer.KMSKeyDetails{}

	if v, ok := tfMap["encryption_option"].(string); ok && v != "" {
		result.EncryptionOption = aws.String(v)
	}

	if v, ok := tfMap["kms_key_id"].(string); ok && v != "" {
		result.KMSKeyId = aws.String(v)
	}

	return result
}

func expandCodeCommitRepository(repository []interface{}) *codegurureviewer.CodeCommitRepository {
	if len(repository) == 0 || repository[0] == nil {
		return nil
	}

	tfMap, ok := repository[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &codegurureviewer.CodeCommitRepository{}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		result.Name = aws.String(v)
	}

	return result
}

func expandRepository(repository []interface{}) *codegurureviewer.Repository {
	if len(repository) == 0 || repository[0] == nil {
		return nil
	}

	tfMap, ok := repository[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &codegurureviewer.Repository{}

	if v, ok := tfMap["bitbucket"]; ok {
		result.Bitbucket = expandThirdPartySourceRepository(v.([]interface{}))
	}
	if v, ok := tfMap["codecommit"]; ok {
		result.CodeCommit = expandCodeCommitRepository(v.([]interface{}))
	}
	if v, ok := tfMap["github_enterprise_server"]; ok {
		result.GitHubEnterpriseServer = expandThirdPartySourceRepository(v.([]interface{}))
	}
	if v, ok := tfMap["s3_bucket"]; ok {
		result.S3Bucket = expandS3Repository(v.([]interface{}))
	}

	return result
}

func expandS3Repository(repository []interface{}) *codegurureviewer.S3Repository {
	if len(repository) == 0 || repository[0] == nil {
		return nil
	}

	tfMap, ok := repository[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &codegurureviewer.S3Repository{}

	if v, ok := tfMap["bucket_name"].(string); ok && v != "" {
		result.BucketName = aws.String(v)
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		result.Name = aws.String(v)
	}

	return result
}

func expandThirdPartySourceRepository(repository []interface{}) *codegurureviewer.ThirdPartySourceRepository {
	if len(repository) == 0 || repository[0] == nil {
		return nil
	}

	tfMap, ok := repository[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &codegurureviewer.ThirdPartySourceRepository{}

	if v, ok := tfMap["connection_arn"].(string); ok && v != "" {
		result.ConnectionArn = aws.String(v)
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		result.Name = aws.String(v)
	}

	if v, ok := tfMap["owner"].(string); ok && v != "" {
		result.Owner = aws.String(v)
	}

	return result
}
