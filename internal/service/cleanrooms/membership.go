// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cleanrooms_membership")
// @Tags(identifierAttribute="arn")
func ResourceMembership() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMembershipCreate,
		ReadWithoutTimeout:   resourceMembershipRead,
		UpdateWithoutTimeout: resourceMembershipUpdate,
		DeleteWithoutTimeout: resourceMembershipDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"collaboration_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"collaboration_creator_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"collaboration_creator_display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"collaboration_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"collaboration_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_result_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"role_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"output_configuration": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"s3": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"bucket": {
													Type:     schema.TypeString,
													Required: true,
												},
												"result_format": {
													Type:     schema.TypeString,
													Required: true,
												},
												"key_prefix": {
													Type:     schema.TypeString,
													Optional: true,
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
			names.AttrID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"member_abilities": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"payment_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"query_compute": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"is_responsible": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"query_log_status": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"update_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameMembership = "Membership"
)

func resourceMembershipCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CleanRoomsClient(ctx)

	input := &cleanrooms.CreateMembershipInput{
		CollaborationIdentifier: aws.String(d.Get("collaboration_id").(string)),
		Tags:                    getTagsIn(ctx),
	}

	queryLogStatus, err := expandMembershipQueryLogStatus(d.Get("query_log_status").(string))
	if err != nil {
		return create.AppendDiagError(diags, names.CleanRooms, create.ErrActionCreating, ResNameCollaboration, d.Get("name").(string), err)
	}
	input.QueryLogStatus = queryLogStatus

	if v, ok := d.GetOk("default_result_configuration"); ok {
		defaultResultConfiguration, err := expandDefaultResultConfiguration(v.([]interface{})[0].(map[string]interface{}))
		if err != nil {
			return create.AppendDiagError(diags, names.CleanRooms, create.ErrActionCreating, ResNameMembership, d.Get("collaboration_id").(string), err)
		}
		input.DefaultResultConfiguration = defaultResultConfiguration
	}

	out, err := conn.CreateMembership(ctx, input)
	if err != nil {
		return create.AppendDiagError(diags, names.CleanRooms, create.ErrActionCreating, ResNameMembership, d.Get("collaboration_id").(string), err)
	}

	if out == nil || out.Membership == nil {
		return create.AppendDiagError(diags, names.CleanRooms, create.ErrActionCreating, ResNameMembership, d.Get("collaboration_id").(string), errors.New("empty output"))
	}
	d.SetId(aws.ToString(out.Membership.Id))

	return append(diags, resourceMembershipRead(ctx, d, meta)...)
}

func resourceMembershipRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CleanRoomsClient(ctx)
	out, err := findMembershipByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Clean Rooms Membership (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.CleanRooms, create.ErrActionReading, ResNameMembership, d.Id(), err)
	}

	membership := out.Membership
	d.Set(names.AttrARN, membership.Arn)
	d.Set("collaboration_arn", membership.CollaborationArn)
	d.Set("collaboration_creator_account_id", membership.CollaborationCreatorAccountId)
	d.Set("collaboration_creator_display_name", membership.CollaborationCreatorDisplayName)
	d.Set("collaboration_id", membership.CollaborationId)
	d.Set("collaboration_name", membership.CollaborationName)
	d.Set("create_time", membership.CreateTime.String())
	d.Set("member_abilities", flattenMemberAbilities(membership.MemberAbilities))
	d.Set("update_time", membership.UpdateTime.String())
	d.Set("status", membership.Status)
	d.Set("query_log_status", membership.QueryLogStatus)

	if err := d.Set("default_result_configuration", flattenDefaultResultConfiguration(membership.DefaultResultConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting default_result_configuration: %s", err)
	}

	if err := d.Set("payment_configuration", flattenPaymentConfiguration(membership.PaymentConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting payment_configuration: %s", err)
	}

	return diags
}

func resourceMembershipUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CleanRoomsClient(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &cleanrooms.UpdateMembershipInput{
			MembershipIdentifier: aws.String(d.Id()),
		}

		if d.HasChanges("query_log_status") {
			queryLogStatus, err := expandMembershipQueryLogStatus(d.Get("query_log_status").(string))
			if err != nil {
				return create.AppendDiagError(diags, names.CleanRooms, create.ErrActionCreating, ResNameMembership, d.Id(), err)
			}
			input.QueryLogStatus = queryLogStatus
		}

		if d.HasChanges("default_result_configuration") {
			defaultResultConfiguration, err := expandDefaultResultConfiguration(d.Get("default_result_configuration").([]interface{})[0].(map[string]interface{}))
			if err != nil {
				return create.AppendDiagError(diags, names.CleanRooms, create.ErrActionCreating, ResNameMembership, d.Id(), err)
			}
			input.DefaultResultConfiguration = defaultResultConfiguration
		}

		_, err := conn.UpdateMembership(ctx, input)
		if err != nil {
			return create.AppendDiagError(diags, names.CleanRooms, create.ErrActionUpdating, ResNameMembership, d.Id(), err)
		}
	}

	return append(diags, resourceMembershipRead(ctx, d, meta)...)
}

func resourceMembershipDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CleanRoomsClient(ctx)

	log.Printf("[INFO] Deleting CleanRooms Membership %s", d.Id())
	_, err := conn.DeleteMembership(ctx, &cleanrooms.DeleteMembershipInput{
		MembershipIdentifier: aws.String(d.Id()),
	})

	if err != nil {
		return create.AppendDiagError(diags, names.CleanRooms, create.ErrActionCreating, ResNameMembership, d.Id(), err)
	}

	return diags
}

func findMembershipByID(ctx context.Context, conn *cleanrooms.Client, id string) (*cleanrooms.GetMembershipOutput, error) {
	in := &cleanrooms.GetMembershipInput{
		MembershipIdentifier: aws.String(id),
	}

	out, err := conn.GetMembership(ctx, in)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Membership == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func expandMembershipQueryLogStatus(queryLogStatus string) (types.MembershipQueryLogStatus, error) {
	switch queryLogStatus {
	case "ENABLED":
		return types.MembershipQueryLogStatusEnabled, nil
	case "DISABLED":
		return types.MembershipQueryLogStatusDisabled, nil
	default:
		return types.MembershipQueryLogStatusDisabled, errors.New("invalid query_log_status - only ENABLED and DISABLED are supported")
	}
}

func expandDefaultResultConfiguration(data map[string]interface{}) (*types.MembershipProtectedQueryResultConfiguration, error) {
	defaultResultConfiguration := &types.MembershipProtectedQueryResultConfiguration{}
	defaultResultConfiguration.RoleArn = aws.String(data["role_arn"].(string))

	if aws.ToString(defaultResultConfiguration.RoleArn) == "" {
		defaultResultConfiguration.RoleArn = nil
	}

	outputConfiguration := data["output_configuration"].([]interface{})[0].(map[string]interface{})
	for destination, configuration := range outputConfiguration {
		configurationMap := configuration.([]interface{})[0].(map[string]interface{})
		switch destination {
		case "s3":
			resultFormat, err := expandResultFormat(configurationMap["result_format"].(string))
			if err != nil {
				return nil, err
			}

			defaultResultConfiguration.OutputConfiguration = &types.MembershipProtectedQueryOutputConfigurationMemberS3{
				Value: types.ProtectedQueryS3OutputConfiguration{
					Bucket:       aws.String(configurationMap["bucket"].(string)),
					ResultFormat: resultFormat,
					KeyPrefix:    aws.String(configurationMap["key_prefix"].(string)),
				},
			}
		default:
			return nil, errors.New("invalid default_result_configuration - unsupported destination")
		}
	}

	return defaultResultConfiguration, nil
}

func expandResultFormat(resultFormat string) (types.ResultFormat, error) {
	switch resultFormat {
	case "CSV":
		return types.ResultFormatCsv, nil
	case "PARQUET":
		return types.ResultFormatParquet, nil
	default:
		return types.ResultFormatParquet, errors.New("invalid result_format - only CSV and PARQUET are supported")
	}
}

func flattenDefaultResultConfiguration(defaultResultConfiguration *types.MembershipProtectedQueryResultConfiguration) []interface{} {
	if defaultResultConfiguration == nil {
		return nil
	}

	m := map[string]interface{}{}

	if defaultResultConfiguration.RoleArn != nil {
		m["role_arn"] = aws.ToString(defaultResultConfiguration.RoleArn)
	}

	switch v := defaultResultConfiguration.OutputConfiguration.(type) {
	case *types.MembershipProtectedQueryOutputConfigurationMemberS3:
		outputConfiguration := map[string]interface{}{
			"s3": []interface{}{map[string]interface{}{
				"bucket":        v.Value.Bucket,
				"result_format": v.Value.ResultFormat,
				"key_prefix":    v.Value.KeyPrefix,
			}},
		}
		m["output_configuration"] = []interface{}{outputConfiguration}
	default:
		return nil
	}

	return []interface{}{m}
}

func flattenPaymentConfiguration(paymentConfiguration *types.MembershipPaymentConfiguration) []interface{} {
	if paymentConfiguration == nil {
		return nil
	}

	m := map[string]interface{}{}

	queryCompute := map[string]interface{}{
		"is_responsible": aws.Bool(*paymentConfiguration.QueryCompute.IsResponsible),
	}

	m["query_compute"] = []interface{}{queryCompute}

	return []interface{}{m}
}
