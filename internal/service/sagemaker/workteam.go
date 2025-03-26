// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_workteam", name="Workteam")
// @Tags(identifierAttribute="arn")
func resourceWorkteam() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWorkteamCreate,
		ReadWithoutTimeout:   resourceWorkteamRead,
		UpdateWithoutTimeout: resourceWorkteamUpdate,
		DeleteWithoutTimeout: resourceWorkteamDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 200),
			},
			"member_definition": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cognito_member_definition": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrClientID: {
										Type:     schema.TypeString,
										Required: true,
									},
									"user_group": {
										Type:     schema.TypeString,
										Required: true,
									},
									"user_pool": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"oidc_member_definition": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"groups": {
										Type:     schema.TypeSet,
										MaxItems: 10,
										Required: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 63),
										},
									},
								},
							},
						},
					},
				},
			},
			"notification_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"notification_topic_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
			},
			"worker_access_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3_presign": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"iam_policy_constraints": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"source_ip": {
													Type:             schema.TypeString,
													Optional:         true,
													Computed:         true,
													ValidateDiagFunc: enum.Validate[awstypes.EnabledOrDisabled](),
													ExactlyOneOf:     []string{"worker_access_configuration.0.s3_presign.0.iam_policy_constraints.0.source_ip", "worker_access_configuration.0.s3_presign.0.iam_policy_constraints.0.vpc_source_ip"},
												},
												"vpc_source_ip": {
													Type:             schema.TypeString,
													Optional:         true,
													Computed:         true,
													ValidateDiagFunc: enum.Validate[awstypes.EnabledOrDisabled](),
													ExactlyOneOf:     []string{"worker_access_configuration.0.s3_presign.0.iam_policy_constraints.0.source_ip", "worker_access_configuration.0.s3_presign.0.iam_policy_constraints.0.vpc_source_ip"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"subdomain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"workforce_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"workteam_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z](-*[0-9A-Za-z])*$`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
		},
	}
}

func resourceWorkteamCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	name := d.Get("workteam_name").(string)
	input := &sagemaker.CreateWorkteamInput{
		WorkteamName:      aws.String(name),
		Description:       aws.String(d.Get(names.AttrDescription).(string)),
		MemberDefinitions: expandWorkteamMemberDefinition(d.Get("member_definition").([]any)),
		Tags:              getTagsIn(ctx),
	}

	if v, ok := d.GetOk("notification_configuration"); ok {
		input.NotificationConfiguration = expandWorkteamNotificationConfiguration(v.([]any))
	}

	if v, ok := d.GetOk("worker_access_configuration"); ok {
		input.WorkerAccessConfiguration = expandWorkerAccessConfiguration(v.([]any))
	}

	if v, ok := d.GetOk("workforce_name"); ok {
		input.WorkforceName = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 2*time.Minute, func() (any, error) {
		return conn.CreateWorkteam(ctx, input)
	}, ErrCodeValidationException)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI Workteam (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceWorkteamRead(ctx, d, meta)...)
}

func resourceWorkteamRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	workteam, err := findWorkteamByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SageMaker AI Workteam (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Workteam (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, workteam.WorkteamArn)
	d.Set("subdomain", workteam.SubDomain)
	d.Set(names.AttrDescription, workteam.Description)
	d.Set("workteam_name", workteam.WorkteamName)

	if err := d.Set("member_definition", flattenWorkteamMemberDefinition(workteam.MemberDefinitions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting member_definition: %s", err)
	}

	if err := d.Set("notification_configuration", flattenWorkteamNotificationConfiguration(workteam.NotificationConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting notification_configuration: %s", err)
	}

	if err := d.Set("worker_access_configuration", flattenWorkerAccessConfiguration(workteam.WorkerAccessConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting worker_access_configuration: %s", err)
	}

	return diags
}

func resourceWorkteamUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &sagemaker.UpdateWorkteamInput{
			WorkteamName:      aws.String(d.Id()),
			MemberDefinitions: expandWorkteamMemberDefinition(d.Get("member_definition").([]any)),
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange("notification_configuration") {
			input.NotificationConfiguration = expandWorkteamNotificationConfiguration(d.Get("notification_configuration").([]any))
		}

		if d.HasChange("worker_access_configuration") {
			input.WorkerAccessConfiguration = expandWorkerAccessConfiguration(d.Get("worker_access_configuration").([]any))
		}

		_, err := conn.UpdateWorkteam(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker AI Workteam (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceWorkteamRead(ctx, d, meta)...)
}

func resourceWorkteamDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	log.Printf("[DEBUG] Deleting SageMaker AI Workteam: %s", d.Id())
	_, err := conn.DeleteWorkteam(ctx, &sagemaker.DeleteWorkteamInput{
		WorkteamName: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "The work team") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI Workteam (%s): %s", d.Id(), err)
	}

	return diags
}

func findWorkteamByName(ctx context.Context, conn *sagemaker.Client, name string) (*awstypes.Workteam, error) {
	input := &sagemaker.DescribeWorkteamInput{
		WorkteamName: aws.String(name),
	}

	output, err := conn.DescribeWorkteam(ctx, input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "The work team") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Workteam == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Workteam, nil
}

func expandWorkteamMemberDefinition(l []any) []awstypes.MemberDefinition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var members []awstypes.MemberDefinition

	for _, mem := range l {
		memRaw := mem.(map[string]any)
		member := awstypes.MemberDefinition{}

		if v, ok := memRaw["cognito_member_definition"].([]any); ok && len(v) > 0 {
			member.CognitoMemberDefinition = expandWorkteamCognitoMemberDefinition(v)
		}

		if v, ok := memRaw["oidc_member_definition"].([]any); ok && len(v) > 0 {
			member.OidcMemberDefinition = expandWorkteamOIDCMemberDefinition(v)
		}

		members = append(members, member)
	}

	return members
}

func flattenWorkteamMemberDefinition(config []awstypes.MemberDefinition) []map[string]any {
	members := make([]map[string]any, 0, len(config))

	for _, raw := range config {
		member := make(map[string]any)

		if raw.CognitoMemberDefinition != nil {
			member["cognito_member_definition"] = flattenWorkteamCognitoMemberDefinition(raw.CognitoMemberDefinition)
		}

		if raw.OidcMemberDefinition != nil {
			member["oidc_member_definition"] = flattenWorkteamOIDCMemberDefinition(raw.OidcMemberDefinition)
		}

		members = append(members, member)
	}

	return members
}

func expandWorkteamCognitoMemberDefinition(l []any) *awstypes.CognitoMemberDefinition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.CognitoMemberDefinition{
		ClientId:  aws.String(m[names.AttrClientID].(string)),
		UserPool:  aws.String(m["user_pool"].(string)),
		UserGroup: aws.String(m["user_group"].(string)),
	}

	return config
}

func flattenWorkteamCognitoMemberDefinition(config *awstypes.CognitoMemberDefinition) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		names.AttrClientID: aws.ToString(config.ClientId),
		"user_pool":        aws.ToString(config.UserPool),
		"user_group":       aws.ToString(config.UserGroup),
	}

	return []map[string]any{m}
}

func expandWorkteamOIDCMemberDefinition(l []any) *awstypes.OidcMemberDefinition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.OidcMemberDefinition{
		Groups: flex.ExpandStringValueSet(m["groups"].(*schema.Set)),
	}

	return config
}

func flattenWorkteamOIDCMemberDefinition(config *awstypes.OidcMemberDefinition) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"groups": flex.FlattenStringValueSet(config.Groups),
	}

	return []map[string]any{m}
}

func expandWorkteamNotificationConfiguration(l []any) *awstypes.NotificationConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.NotificationConfiguration{}

	if v, ok := m["notification_topic_arn"].(string); ok && v != "" {
		config.NotificationTopicArn = aws.String(v)
	} else {
		return nil
	}

	return config
}

func flattenWorkteamNotificationConfiguration(config *awstypes.NotificationConfiguration) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"notification_topic_arn": aws.ToString(config.NotificationTopicArn),
	}

	return []map[string]any{m}
}

func expandWorkerAccessConfiguration(l []any) *awstypes.WorkerAccessConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.WorkerAccessConfiguration{}

	if v, ok := m["s3_presign"].([]any); ok && len(v) > 0 && v[0] != nil {
		config.S3Presign = expandS3Presign(v)
	} else {
		return nil
	}

	return config
}

func flattenWorkerAccessConfiguration(config *awstypes.WorkerAccessConfiguration) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"s3_presign": flattenS3Presign(config.S3Presign),
	}

	return []map[string]any{m}
}

func expandS3Presign(l []any) *awstypes.S3Presign {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.S3Presign{}

	if v, ok := m["iam_policy_constraints"].([]any); ok && len(v) > 0 && v[0] != nil {
		config.IamPolicyConstraints = expandIAMPolicyConstraints(v)
	} else {
		return nil
	}

	return config
}

func flattenS3Presign(config *awstypes.S3Presign) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"iam_policy_constraints": flattenIAMPolicyConstraints(config.IamPolicyConstraints),
	}

	return []map[string]any{m}
}

func expandIAMPolicyConstraints(l []any) *awstypes.IamPolicyConstraints {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.IamPolicyConstraints{}

	if v, ok := m["source_ip"].(string); ok && v != "" {
		config.SourceIp = awstypes.EnabledOrDisabled(v)
	}

	if v, ok := m["vpc_source_ip"].(string); ok && v != "" {
		config.VpcSourceIp = awstypes.EnabledOrDisabled(v)
	}

	return config
}

func flattenIAMPolicyConstraints(config *awstypes.IamPolicyConstraints) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"source_ip":     config.SourceIp,
		"vpc_source_ip": config.VpcSourceIp,
	}

	return []map[string]any{m}
}
