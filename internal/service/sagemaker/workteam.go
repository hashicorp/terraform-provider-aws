// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_workteam", name="Workteam")
// @Tags(identifierAttribute="arn")
func ResourceWorkteam() *schema.Resource {
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
													Type:         schema.TypeString,
													Optional:     true,
													Computed:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.EnabledOrDisabled_Values(), false),
													ExactlyOneOf: []string{"worker_access_configuration.0.s3_presign.0.iam_policy_constraints.0.source_ip", "worker_access_configuration.0.s3_presign.0.iam_policy_constraints.0.vpc_source_ip"},
												},
												"vpc_source_ip": {
													Type:         schema.TypeString,
													Optional:     true,
													Computed:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.EnabledOrDisabled_Values(), false),
													ExactlyOneOf: []string{"worker_access_configuration.0.s3_presign.0.iam_policy_constraints.0.source_ip", "worker_access_configuration.0.s3_presign.0.iam_policy_constraints.0.vpc_source_ip"},
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
				Required: true,
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

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceWorkteamCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	name := d.Get("workteam_name").(string)
	input := &sagemaker.CreateWorkteamInput{
		WorkteamName:      aws.String(name),
		WorkforceName:     aws.String(d.Get("workforce_name").(string)),
		Description:       aws.String(d.Get(names.AttrDescription).(string)),
		MemberDefinitions: expandWorkteamMemberDefinition(d.Get("member_definition").([]interface{})),
		Tags:              getTagsIn(ctx),
	}

	if v, ok := d.GetOk("notification_configuration"); ok {
		input.NotificationConfiguration = expandWorkteamNotificationConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("worker_access_configuration"); ok {
		input.WorkerAccessConfiguration = expandWorkerAccessConfiguration(v.([]interface{}))
	}

	log.Printf("[DEBUG] Updating SageMaker Workteam: %s", input)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 2*time.Minute, func() (interface{}, error) {
		return conn.CreateWorkteamWithContext(ctx, input)
	}, "ValidationException")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Workteam (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceWorkteamRead(ctx, d, meta)...)
}

func resourceWorkteamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	workteam, err := FindWorkteamByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SageMaker Workteam (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Workteam (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(workteam.WorkteamArn)
	d.Set(names.AttrARN, arn)
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

func resourceWorkteamUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &sagemaker.UpdateWorkteamInput{
			WorkteamName:      aws.String(d.Id()),
			MemberDefinitions: expandWorkteamMemberDefinition(d.Get("member_definition").([]interface{})),
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange("notification_configuration") {
			input.NotificationConfiguration = expandWorkteamNotificationConfiguration(d.Get("notification_configuration").([]interface{}))
		}

		if d.HasChange("worker_access_configuration") {
			input.WorkerAccessConfiguration = expandWorkerAccessConfiguration(d.Get("worker_access_configuration").([]interface{}))
		}

		log.Printf("[DEBUG] Updating SageMaker Workteam: %s", input)
		_, err := conn.UpdateWorkteamWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker Workteam (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceWorkteamRead(ctx, d, meta)...)
}

func resourceWorkteamDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	log.Printf("[DEBUG] Deleting SageMaker Workteam: %s", d.Id())
	_, err := conn.DeleteWorkteamWithContext(ctx, &sagemaker.DeleteWorkteamInput{
		WorkteamName: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, "ValidationException", "The work team") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker Workteam (%s): %s", d.Id(), err)
	}

	return diags
}

func expandWorkteamMemberDefinition(l []interface{}) []*sagemaker.MemberDefinition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var members []*sagemaker.MemberDefinition

	for _, mem := range l {
		memRaw := mem.(map[string]interface{})
		member := &sagemaker.MemberDefinition{}

		if v, ok := memRaw["cognito_member_definition"].([]interface{}); ok && len(v) > 0 {
			member.CognitoMemberDefinition = expandWorkteamCognitoMemberDefinition(v)
		}

		if v, ok := memRaw["oidc_member_definition"].([]interface{}); ok && len(v) > 0 {
			member.OidcMemberDefinition = expandWorkteamOIDCMemberDefinition(v)
		}

		members = append(members, member)
	}

	return members
}

func flattenWorkteamMemberDefinition(config []*sagemaker.MemberDefinition) []map[string]interface{} {
	members := make([]map[string]interface{}, 0, len(config))

	for _, raw := range config {
		member := make(map[string]interface{})

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

func expandWorkteamCognitoMemberDefinition(l []interface{}) *sagemaker.CognitoMemberDefinition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.CognitoMemberDefinition{
		ClientId:  aws.String(m[names.AttrClientID].(string)),
		UserPool:  aws.String(m["user_pool"].(string)),
		UserGroup: aws.String(m["user_group"].(string)),
	}

	return config
}

func flattenWorkteamCognitoMemberDefinition(config *sagemaker.CognitoMemberDefinition) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		names.AttrClientID: aws.StringValue(config.ClientId),
		"user_pool":        aws.StringValue(config.UserPool),
		"user_group":       aws.StringValue(config.UserGroup),
	}

	return []map[string]interface{}{m}
}

func expandWorkteamOIDCMemberDefinition(l []interface{}) *sagemaker.OidcMemberDefinition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.OidcMemberDefinition{
		Groups: flex.ExpandStringSet(m["groups"].(*schema.Set)),
	}

	return config
}

func flattenWorkteamOIDCMemberDefinition(config *sagemaker.OidcMemberDefinition) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"groups": flex.FlattenStringSet(config.Groups),
	}

	return []map[string]interface{}{m}
}

func expandWorkteamNotificationConfiguration(l []interface{}) *sagemaker.NotificationConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.NotificationConfiguration{}

	if v, ok := m["notification_topic_arn"].(string); ok && v != "" {
		config.NotificationTopicArn = aws.String(v)
	} else {
		return nil
	}

	return config
}

func flattenWorkteamNotificationConfiguration(config *sagemaker.NotificationConfiguration) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"notification_topic_arn": aws.StringValue(config.NotificationTopicArn),
	}

	return []map[string]interface{}{m}
}

func expandWorkerAccessConfiguration(l []interface{}) *sagemaker.WorkerAccessConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.WorkerAccessConfiguration{}

	if v, ok := m["s3_presign"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		config.S3Presign = expandS3Presign(v)
	} else {
		return nil
	}

	return config
}

func flattenWorkerAccessConfiguration(config *sagemaker.WorkerAccessConfiguration) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"s3_presign": flattenS3Presign(config.S3Presign),
	}

	return []map[string]interface{}{m}
}

func expandS3Presign(l []interface{}) *sagemaker.S3Presign {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.S3Presign{}

	if v, ok := m["iam_policy_constraints"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		config.IamPolicyConstraints = expandIAMPolicyConstraints(v)
	} else {
		return nil
	}

	return config
}

func flattenS3Presign(config *sagemaker.S3Presign) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"iam_policy_constraints": flattenIAMPolicyConstraints(config.IamPolicyConstraints),
	}

	return []map[string]interface{}{m}
}

func expandIAMPolicyConstraints(l []interface{}) *sagemaker.IamPolicyConstraints {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.IamPolicyConstraints{}

	if v, ok := m["source_ip"].(string); ok && v != "" {
		config.SourceIp = aws.String(v)
	}

	if v, ok := m["vpc_source_ip"].(string); ok && v != "" {
		config.VpcSourceIp = aws.String(v)
	}

	return config
}

func flattenIAMPolicyConstraints(config *sagemaker.IamPolicyConstraints) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"source_ip":     aws.StringValue(config.SourceIp),
		"vpc_source_ip": aws.StringValue(config.VpcSourceIp),
	}

	return []map[string]interface{}{m}
}
