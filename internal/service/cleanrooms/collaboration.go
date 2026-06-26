// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package cleanrooms

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cleanrooms"
	"github.com/aws/aws-sdk-go-v2/service/cleanrooms/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cleanrooms_collaboration", name="Collaboration")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
// @IdentityAttribute("id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cleanrooms;cleanrooms.GetCollaborationOutput")
// @Testing(preIdentityVersion="v6.26.0")
func resourceCollaboration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCollaborationCreate,
		ReadWithoutTimeout:   resourceCollaborationRead,
		UpdateWithoutTimeout: resourceCollaborationUpdate,
		DeleteWithoutTimeout: resourceCollaborationDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"allowed_result_regions": {
					Type:     schema.TypeList,
					Optional: true,
					ForceNew: true,
					Elem: &schema.Schema{
						Type:             schema.TypeString,
						ValidateDiagFunc: enum.Validate[types.SupportedS3Region](),
					},
				},
				"analytics_engine": {
					Type:             schema.TypeString,
					Optional:         true,
					Computed:         true,
					ForceNew:         true,
					ValidateDiagFunc: enum.Validate[types.AnalyticsEngine](),
					Deprecated:       "AWS Clean Rooms now uses Spark exclusively for new collaborations. CLEAN_ROOMS_SQL is no longer accepted at create time and AWS will return a ValidationException. See https://docs.aws.amazon.com/clean-rooms/latest/userguide/doc-history.html.",
				},
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"auto_approved_change_request_types": {
					Type:     schema.TypeList,
					Optional: true,
					ForceNew: true,
					Elem: &schema.Schema{
						Type:             schema.TypeString,
						ValidateDiagFunc: enum.Validate[types.AutoApprovedChangeType](),
					},
				},
				names.AttrCreateTime: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"creator_display_name": {
					Type:     schema.TypeString,
					ForceNew: true,
					Required: true,
				},
				"creator_member_abilities": {
					Type:     schema.TypeList,
					Required: true,
					ForceNew: true,
					Elem: &schema.Schema{
						Type:             schema.TypeString,
						ValidateDiagFunc: enum.Validate[types.MemberAbility](),
					},
				},
				"creator_ml_member_abilities":   mlMemberAbilitiesSchema(),
				"creator_payment_configuration": paymentConfigurationSchema(),
				names.AttrDescription: {
					Type:     schema.TypeString,
					Required: true,
				},
				"data_encryption_metadata": {
					Type:     schema.TypeList,
					Optional: true,
					ForceNew: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"allow_clear_text": {
								Type:     schema.TypeBool,
								Required: true,
								ForceNew: true,
							},
							"allow_duplicates": {
								Type:     schema.TypeBool,
								Required: true,
								ForceNew: true,
							},
							"allow_joins_on_columns_with_different_names": {
								Type:     schema.TypeBool,
								Required: true,
								ForceNew: true,
							},
							"preserve_nulls": {
								Type:     schema.TypeBool,
								Required: true,
								ForceNew: true,
							},
						},
					},
				},
				names.AttrID: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"is_metrics_enabled": {
					Type:     schema.TypeBool,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},
				"job_log_status": {
					Type:             schema.TypeString,
					Optional:         true,
					Computed:         true,
					ForceNew:         true,
					ValidateDiagFunc: enum.Validate[types.CollaborationJobLogStatus](),
				},
				"member": {
					Type:     schema.TypeSet,
					Optional: true,
					ForceNew: true,
					Set: func(v any) int {
						m := v.(map[string]any)
						return create.StringHashcode(m[names.AttrAccountID].(string))
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrAccountID: {
								Type:     schema.TypeString,
								Required: true,
								ForceNew: true,
							},
							names.AttrDisplayName: {
								Type:     schema.TypeString,
								Required: true,
								ForceNew: true,
							},
							"member_abilities": {
								Type:     schema.TypeList,
								Required: true,
								ForceNew: true,
								Elem: &schema.Schema{
									Type:             schema.TypeString,
									ValidateDiagFunc: enum.Validate[types.MemberAbility](),
								},
							},
							"ml_member_abilities":   mlMemberAbilitiesSchema(),
							"payment_configuration": paymentConfigurationSchema(),
							names.AttrStatus: {
								Type:     schema.TypeString,
								Computed: true,
							},
						},
					},
				},
				"membership_arn": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"membership_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
				},
				"query_log_status": {
					Type:             schema.TypeString,
					Required:         true,
					ForceNew:         true,
					ValidateDiagFunc: enum.Validate[types.CollaborationQueryLogStatus](),
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
				"update_time": {
					Type:     schema.TypeString,
					Computed: true,
				},
			}
		},
	}
}

func mlMemberAbilitiesSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		ForceNew: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"custom_ml_member_abilities": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					ForceNew: true,
					Elem: &schema.Schema{
						Type:             schema.TypeString,
						ValidateDiagFunc: enum.Validate[types.CustomMLMemberAbility](),
					},
				},
			},
		},
	}
}

func paymentConfigurationSchema() *schema.Schema {
	isResponsibleBlock := func() *schema.Schema {
		return &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			ForceNew: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"is_responsible": {
						Type:     schema.TypeBool,
						Required: true,
						ForceNew: true,
					},
				},
			},
		}
	}
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		ForceNew: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"job_compute": isResponsibleBlock(),
				"machine_learning": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					ForceNew: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"model_inference":           isResponsibleBlock(),
							"model_training":            isResponsibleBlock(),
							"synthetic_data_generation": isResponsibleBlock(),
						},
					},
				},
				// query_compute is required by the AWS API on every PaymentConfiguration.
				// See https://docs.aws.amazon.com/clean-rooms/latest/apireference/API_PaymentConfiguration.html.
				"query_compute": {
					Type:     schema.TypeList,
					Required: true,
					ForceNew: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"is_responsible": {
								Type:     schema.TypeBool,
								Required: true,
								ForceNew: true,
							},
						},
					},
				},
			},
		},
	}
}

const (
	ResNameCollaboration = "Collaboration"
)

func resourceCollaborationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CleanRoomsClient(ctx)

	input := cleanrooms.CreateCollaborationInput{
		Name:                   aws.String(d.Get(names.AttrName).(string)),
		CreatorDisplayName:     aws.String(d.Get("creator_display_name").(string)),
		CreatorMemberAbilities: flex.ExpandStringyValueList[types.MemberAbility](d.Get("creator_member_abilities").([]any)),
		Members:                *expandMembers(d.Get("member").(*schema.Set).List()),
		Tags:                   getTagsIn(ctx),
	}

	if v, ok := d.GetOk("allowed_result_regions"); ok {
		input.AllowedResultRegions = flex.ExpandStringyValueList[types.SupportedS3Region](v.([]any))
	}

	if v, ok := d.GetOk("analytics_engine"); ok {
		input.AnalyticsEngine = types.AnalyticsEngine(v.(string))
	}

	if v, ok := d.GetOk("auto_approved_change_request_types"); ok {
		input.AutoApprovedChangeRequestTypes = flex.ExpandStringyValueList[types.AutoApprovedChangeType](v.([]any))
	}

	if v, ok := d.GetOk("creator_ml_member_abilities"); ok {
		input.CreatorMLMemberAbilities = expandMLMemberAbilities(v.([]any))
	}

	if v, ok := d.GetOk("creator_payment_configuration"); ok {
		input.CreatorPaymentConfiguration = expandPaymentConfiguration(v.([]any))
	}

	input.QueryLogStatus = types.CollaborationQueryLogStatus(d.Get("query_log_status").(string))

	if v, ok := d.GetOk("data_encryption_metadata"); ok {
		input.DataEncryptionMetadata = expandDataEncryptionMetadata(v.([]any))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	// is_metrics_enabled is Optional+Computed+TypeBool. Use GetRawConfig so an
	// explicit `false` is reliably distinguished from absence; d.GetOk and
	// d.GetOkExists are unreliable here per terraform-plugin-sdk guidance.
	if v := d.GetRawConfig().GetAttr("is_metrics_enabled"); v.IsKnown() && !v.IsNull() {
		input.IsMetricsEnabled = aws.Bool(v.True())
	}

	if v, ok := d.GetOk("job_log_status"); ok {
		input.JobLogStatus = types.CollaborationJobLogStatus(v.(string))
	}

	out, err := conn.CreateCollaboration(ctx, &input)
	if err != nil {
		return create.AppendDiagError(diags, names.CleanRooms, create.ErrActionCreating, ResNameCollaboration, d.Get(names.AttrName).(string), err)
	}

	if out == nil || out.Collaboration == nil {
		return create.AppendDiagError(diags, names.CleanRooms, create.ErrActionCreating, ResNameCollaboration, d.Get(names.AttrName).(string), errors.New("empty output"))
	}
	d.SetId(aws.ToString(out.Collaboration.Id))

	return append(diags, resourceCollaborationRead(ctx, d, meta)...)
}

func resourceCollaborationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CleanRoomsClient(ctx)

	out, err := findCollaborationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] CleanRooms Collaboration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.CleanRooms, create.ErrActionReading, ResNameCollaboration, d.Id(), err)
	}

	return append(diags, resourceCollaborationFlatten(ctx, conn, d, out)...)
}

func resourceCollaborationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CleanRoomsClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := cleanrooms.UpdateCollaborationInput{
			CollaborationIdentifier: aws.String(d.Id()),
		}

		if d.HasChanges(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChanges(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		_, err := conn.UpdateCollaboration(ctx, &input)
		if err != nil {
			return create.AppendDiagError(diags, names.CleanRooms, create.ErrActionUpdating, ResNameCollaboration, d.Id(), err)
		}
	}

	return append(diags, resourceCollaborationRead(ctx, d, meta)...)
}

func resourceCollaborationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CleanRoomsClient(ctx)

	log.Printf("[INFO] Deleting CleanRooms Collaboration %s", d.Id())
	input := cleanrooms.DeleteCollaborationInput{
		CollaborationIdentifier: aws.String(d.Id()),
	}
	_, err := conn.DeleteCollaboration(ctx, &input)

	if errs.IsA[*types.AccessDeniedException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.CleanRooms, create.ErrActionDeleting, ResNameCollaboration, d.Id(), err)
	}

	return diags
}

func findCollaborationByID(ctx context.Context, conn *cleanrooms.Client, id string) (*cleanrooms.GetCollaborationOutput, error) {
	input := cleanrooms.GetCollaborationInput{
		CollaborationIdentifier: aws.String(id),
	}
	out, err := conn.GetCollaboration(ctx, &input)

	if errs.IsA[*types.AccessDeniedException](err) {
		//We throw Access Denied for NFE in Cleanrooms for collaborations since they are cross account
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Collaboration == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

func findMembersByCollaborationId(ctx context.Context, conn *cleanrooms.Client, id string) (*cleanrooms.ListMembersOutput, error) {
	input := cleanrooms.ListMembersInput{
		CollaborationIdentifier: aws.String(id),
	}
	out, err := conn.ListMembers(ctx, &input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.MemberSummaries == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

func expandDataEncryptionMetadata(data []any) *types.DataEncryptionMetadata {
	dataEncryptionMetadata := types.DataEncryptionMetadata{}
	if len(data) > 0 {
		metadata := data[0].(map[string]any)
		dataEncryptionMetadata.PreserveNulls = aws.Bool(metadata["preserve_nulls"].(bool))
		dataEncryptionMetadata.AllowCleartext = aws.Bool(metadata["allow_clear_text"].(bool))
		dataEncryptionMetadata.AllowJoinsOnColumnsWithDifferentNames = aws.Bool(metadata["allow_joins_on_columns_with_different_names"].(bool))
		dataEncryptionMetadata.AllowDuplicates = aws.Bool(metadata["allow_duplicates"].(bool))
	}
	return &dataEncryptionMetadata
}

func expandMembers(data []any) *[]types.MemberSpecification {
	members := make([]types.MemberSpecification, 0)
	for _, member := range data {
		memberMap := member.(map[string]any)
		m := types.MemberSpecification{
			AccountId:       aws.String(memberMap[names.AttrAccountID].(string)),
			MemberAbilities: flex.ExpandStringyValueList[types.MemberAbility](memberMap["member_abilities"].([]any)),
			DisplayName:     aws.String(memberMap[names.AttrDisplayName].(string)),
		}
		if v, ok := memberMap["ml_member_abilities"].([]any); ok && len(v) > 0 {
			m.MlMemberAbilities = expandMLMemberAbilities(v)
		}
		if v, ok := memberMap["payment_configuration"].([]any); ok && len(v) > 0 {
			m.PaymentConfiguration = expandPaymentConfiguration(v)
		}
		members = append(members, m)
	}
	return &members
}

func expandMLMemberAbilities(data []any) *types.MLMemberAbilities {
	if len(data) == 0 {
		return nil
	}
	m := data[0].(map[string]any)
	return &types.MLMemberAbilities{
		CustomMLMemberAbilities: flex.ExpandStringyValueList[types.CustomMLMemberAbility](m["custom_ml_member_abilities"].([]any)),
	}
}

// expandIsResponsible reads the `is_responsible` bool from a single-element
// list block (the shape used by every leaf payment-config block) and returns
// nil when the block is absent.
func expandIsResponsible(v any) *bool {
	data, ok := v.([]any)
	if !ok || len(data) == 0 {
		return nil
	}
	return aws.Bool(data[0].(map[string]any)["is_responsible"].(bool))
}

func expandPaymentConfiguration(data []any) *types.PaymentConfiguration {
	if len(data) == 0 {
		return nil
	}
	m := data[0].(map[string]any)
	out := &types.PaymentConfiguration{}
	if r := expandIsResponsible(m["query_compute"]); r != nil {
		out.QueryCompute = &types.QueryComputePaymentConfig{IsResponsible: r}
	}
	if r := expandIsResponsible(m["job_compute"]); r != nil {
		out.JobCompute = &types.JobComputePaymentConfig{IsResponsible: r}
	}
	if v, ok := m["machine_learning"].([]any); ok && len(v) > 0 {
		out.MachineLearning = expandMLPaymentConfig(v)
	}
	return out
}

func expandMLPaymentConfig(data []any) *types.MLPaymentConfig {
	if len(data) == 0 {
		return nil
	}
	m := data[0].(map[string]any)
	out := &types.MLPaymentConfig{}
	if r := expandIsResponsible(m["model_inference"]); r != nil {
		out.ModelInference = &types.ModelInferencePaymentConfig{IsResponsible: r}
	}
	if r := expandIsResponsible(m["model_training"]); r != nil {
		out.ModelTraining = &types.ModelTrainingPaymentConfig{IsResponsible: r}
	}
	if r := expandIsResponsible(m["synthetic_data_generation"]); r != nil {
		out.SyntheticDataGeneration = &types.SyntheticDataGenerationPaymentConfig{IsResponsible: r}
	}
	return out
}

func flattenDataEncryptionMetadata(dataEncryptionMetadata *types.DataEncryptionMetadata) []any {
	if dataEncryptionMetadata == nil {
		return nil
	}
	m := map[string]any{}
	m["preserve_nulls"] = aws.Bool(*dataEncryptionMetadata.PreserveNulls)
	m["allow_clear_text"] = aws.Bool(*dataEncryptionMetadata.AllowCleartext)
	m["allow_joins_on_columns_with_different_names"] = aws.Bool(*dataEncryptionMetadata.AllowJoinsOnColumnsWithDifferentNames)
	m["allow_duplicates"] = aws.Bool(*dataEncryptionMetadata.AllowDuplicates)
	return []any{m}
}

func flattenMembers(members []types.MemberSummary, ownerAccount *string) []any {
	flattenedMembers := make([]any, 0)
	for _, member := range members {
		if aws.ToString(member.AccountId) == aws.ToString(ownerAccount) {
			continue
		}
		flattenedMembers = append(flattenedMembers, map[string]any{
			names.AttrStatus:        member.Status,
			names.AttrAccountID:     member.AccountId,
			names.AttrDisplayName:   member.DisplayName,
			"member_abilities":      flex.FlattenStringyValueList(member.Abilities),
			"ml_member_abilities":   flattenMLMemberAbilities(member.MlAbilities),
			"payment_configuration": flattenPaymentConfiguration(member.PaymentConfiguration),
		})
	}
	return flattenedMembers
}

// findCreatorMember returns the MemberSummary entry matching the creator
// account, or nil if it is not present in the list.
func findCreatorMember(members []types.MemberSummary, ownerAccount *string) *types.MemberSummary {
	for i, member := range members {
		if aws.ToString(member.AccountId) == aws.ToString(ownerAccount) {
			return &members[i]
		}
	}
	return nil
}

func flattenMLMemberAbilities(in *types.MLMemberAbilities) []any {
	if in == nil {
		return nil
	}
	return []any{map[string]any{
		"custom_ml_member_abilities": flex.FlattenStringyValueList(in.CustomMLMemberAbilities),
	}}
}

// flattenIsResponsible builds the single-element list block that wraps
// `is_responsible` for every leaf payment-config block. Returns nil when the
// input pointer is nil so the wrapping caller can skip the key.
func flattenIsResponsible(b *bool) []any {
	if b == nil {
		return nil
	}
	return []any{map[string]any{
		"is_responsible": aws.ToBool(b),
	}}
}

func flattenPaymentConfiguration(in *types.PaymentConfiguration) []any {
	if in == nil {
		return nil
	}
	m := map[string]any{}
	if in.QueryCompute != nil {
		m["query_compute"] = flattenIsResponsible(in.QueryCompute.IsResponsible)
	}
	if in.JobCompute != nil {
		m["job_compute"] = flattenIsResponsible(in.JobCompute.IsResponsible)
	}
	if in.MachineLearning != nil {
		m["machine_learning"] = flattenMLPaymentConfig(in.MachineLearning)
	}
	return []any{m}
}

func flattenMLPaymentConfig(in *types.MLPaymentConfig) []any {
	if in == nil {
		return nil
	}
	m := map[string]any{}
	if in.ModelInference != nil {
		m["model_inference"] = flattenIsResponsible(in.ModelInference.IsResponsible)
	}
	if in.ModelTraining != nil {
		m["model_training"] = flattenIsResponsible(in.ModelTraining.IsResponsible)
	}
	if in.SyntheticDataGeneration != nil {
		m["synthetic_data_generation"] = flattenIsResponsible(in.SyntheticDataGeneration.IsResponsible)
	}
	return []any{m}
}

func resourceCollaborationFlatten(ctx context.Context, conn *cleanrooms.Client, d *schema.ResourceData, out *cleanrooms.GetCollaborationOutput) diag.Diagnostics {
	var diags diag.Diagnostics

	collaboration := out.Collaboration
	d.Set(names.AttrARN, collaboration.Arn)
	d.Set(names.AttrName, collaboration.Name)
	d.Set(names.AttrDescription, collaboration.Description)
	d.Set("allowed_result_regions", flex.FlattenStringyValueList(collaboration.AllowedResultRegions))
	d.Set("analytics_engine", collaboration.AnalyticsEngine)
	d.Set("auto_approved_change_request_types", flex.FlattenStringyValueList(collaboration.AutoApprovedChangeTypes))
	d.Set("creator_display_name", collaboration.CreatorDisplayName)
	d.Set(names.AttrCreateTime, collaboration.CreateTime.String())
	d.Set("is_metrics_enabled", aws.ToBool(collaboration.IsMetricsEnabled))
	d.Set("job_log_status", collaboration.JobLogStatus)
	d.Set("membership_arn", collaboration.MembershipArn)
	d.Set("membership_id", collaboration.MembershipId)
	d.Set("update_time", collaboration.UpdateTime.String())
	d.Set("query_log_status", collaboration.QueryLogStatus)
	d.Set("data_encryption_metadata", flattenDataEncryptionMetadata(collaboration.DataEncryptionMetadata))

	membersOut, err := findMembersByCollaborationId(ctx, conn, d.Id())
	if err != nil {
		return create.AppendDiagError(diags, names.CleanRooms, create.ErrActionSetting, ResNameCollaboration, d.Id(), err)
	}

	d.Set("member", flattenMembers(membersOut.MemberSummaries, collaboration.CreatorAccountId))

	if creator := findCreatorMember(membersOut.MemberSummaries, collaboration.CreatorAccountId); creator != nil {
		d.Set("creator_member_abilities", flex.FlattenStringyValueList(creator.Abilities))
		d.Set("creator_ml_member_abilities", flattenMLMemberAbilities(creator.MlAbilities))
		d.Set("creator_payment_configuration", flattenPaymentConfiguration(creator.PaymentConfiguration))
	}

	return diags
}
