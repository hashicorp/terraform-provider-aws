// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_eks_capability", name="Capability")
// @Tags(identifierAttribute="arn")
func resourceCapability() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCapabilityCreate,
		ReadWithoutTimeout:   resourceCapabilityRead,
		UpdateWithoutTimeout: resourceCapabilityUpdate,
		DeleteWithoutTimeout: resourceCapabilityDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"capability_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrClusterName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validClusterName,
			},
			"configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"argo_cd": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"aws_idc": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"idc_instance_arn": {
													Type:     schema.TypeString,
													Required: true,
												},
												"idc_region": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
									"namespace": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"network_access": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"vpce_ids": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"rbac_role_mappings": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"identities": {
													Type:     schema.TypeList,
													Required: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"id": {
																Type:     schema.TypeString,
																Required: true,
															},
															"type": {
																Type:     schema.TypeString,
																Required: true,
															},
														},
													},
												},
												"role": {
													Type:     schema.TypeString,
													Required: true,
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
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"delete_propagation_policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.CapabilityDeletePropagationPolicy](),
			},
			"modified_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.CapabilityType](),
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCapabilityCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	capabilityName := d.Get("capability_name").(string)
	clusterName := d.Get(names.AttrClusterName).(string)
	id := capabilityCreateResourceID(clusterName, capabilityName)

	cluster, err := findClusterByName(ctx, conn, clusterName)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EKS Cluster (%s): %s", clusterName, err)
	}

	if cluster.Status == types.ClusterStatusCreating || cluster.Status == types.ClusterStatusDeleting || cluster.Status == types.ClusterStatusFailed {
		return sdkdiag.AppendErrorf(diags, "cannot create EKS Capability when cluster is in %s state", cluster.Status)
	}

	input := &eks.CreateCapabilityInput{
		CapabilityName:          aws.String(capabilityName),
		ClusterName:             aws.String(clusterName),
		ClientRequestToken:      aws.String(sdkid.UniqueId()),
		DeletePropagationPolicy: types.CapabilityDeletePropagationPolicy(d.Get("delete_propagation_policy").(string)),
		RoleArn:                 aws.String(d.Get(names.AttrRoleARN).(string)),
		Type:                    types.CapabilityType(d.Get("type").(string)),
		Tags:                    getTagsIn(ctx),
	}

	if v, ok := d.GetOk("configuration"); ok {
		input.Configuration = expandCapabilityConfiguration(v.([]any))
	}

	_, err = conn.CreateCapability(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EKS Capability (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitCapabilityCreated(ctx, conn, clusterName, capabilityName, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EKS Capability (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceCapabilityRead(ctx, d, meta)...)
}

func resourceCapabilityRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	clusterName, capabilityName, err := capabilityParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	capability, err := findCapabilityByTwoPartKey(ctx, conn, clusterName, capabilityName)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] EKS Capability (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EKS Capability (%s): %s", d.Id(), err)
	}

	d.Set("capability_name", capability.CapabilityName)
	d.Set(names.AttrARN, capability.Arn)
	d.Set(names.AttrClusterName, capability.ClusterName)
	d.Set(names.AttrCreatedAt, aws.ToTime(capability.CreatedAt).Format(time.RFC3339))
	d.Set("delete_propagation_policy", capability.DeletePropagationPolicy)
	d.Set("modified_at", aws.ToTime(capability.ModifiedAt).Format(time.RFC3339))
	d.Set(names.AttrRoleARN, capability.RoleArn)
	d.Set("status", capability.Status)
	d.Set("type", capability.Type)
	d.Set("version", capability.Version)

	if capability.Configuration != nil {
		if err := d.Set("configuration", flattenCapabilityConfiguration(capability.Configuration)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting configuration: %s", err)
		}
	}

	setTagsOut(ctx, capability.Tags)

	return diags
}

func resourceCapabilityUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	clusterName, capabilityName, err := capabilityParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChanges("configuration", "delete_propagation_policy", names.AttrRoleARN) {
		cluster, err := findClusterByName(ctx, conn, clusterName)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading EKS Cluster (%s): %s", clusterName, err)
		}

		if cluster.Status == types.ClusterStatusCreating || cluster.Status == types.ClusterStatusDeleting || cluster.Status == types.ClusterStatusFailed {
			return sdkdiag.AppendErrorf(diags, "cannot update EKS Capability when cluster is in %s state", cluster.Status)
		}
		input := &eks.UpdateCapabilityInput{
			CapabilityName:     aws.String(capabilityName),
			ClusterName:        aws.String(clusterName),
			ClientRequestToken: aws.String(sdkid.UniqueId()),
		}

		hasUpdate := false

		if d.HasChange("configuration") {
			if v, ok := d.GetOk("configuration"); ok {
				input.Configuration = expandUpdateCapabilityConfiguration(v.([]any))
				hasUpdate = true
			}
		}

		if d.HasChange("delete_propagation_policy") {
			input.DeletePropagationPolicy = types.CapabilityDeletePropagationPolicy(d.Get("delete_propagation_policy").(string))
			hasUpdate = true
		}

		if d.HasChange(names.AttrRoleARN) {
			input.RoleArn = aws.String(d.Get(names.AttrRoleARN).(string))
			hasUpdate = true
		}

		if !hasUpdate {
			return append(diags, resourceCapabilityRead(ctx, d, meta)...)
		}

		output, err := conn.UpdateCapability(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EKS Capability (%s): %s", d.Id(), err)
		}

		updateID := aws.ToString(output.Update.Id)
		if _, err := waitCapabilityUpdateSuccessful(ctx, conn, clusterName, capabilityName, updateID, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EKS Capability (%s) update (%s): %s", d.Id(), updateID, err)
		}
	}

	return append(diags, resourceCapabilityRead(ctx, d, meta)...)
}

func resourceCapabilityDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	clusterName, capabilityName, err := capabilityParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting EKS Capability: %s", d.Id())
	_, err = conn.DeleteCapability(ctx, &eks.DeleteCapabilityInput{
		CapabilityName: aws.String(capabilityName),
		ClusterName:    aws.String(clusterName),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EKS Capability (%s): %s", d.Id(), err)
	}

	if _, err := waitCapabilityDeleted(ctx, conn, clusterName, capabilityName, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EKS Capability (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandCapabilityConfiguration(tfList []any) *types.CapabilityConfigurationRequest {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	config := &types.CapabilityConfigurationRequest{}

	if v, ok := tfMap["argo_cd"].([]any); ok && len(v) > 0 {
		config.ArgoCd = expandArgoCdConfigRequest(v)
	}

	return config
}

func expandUpdateCapabilityConfiguration(tfList []any) *types.UpdateCapabilityConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	config := &types.UpdateCapabilityConfiguration{}

	if v, ok := tfMap["argo_cd"].([]any); ok && len(v) > 0 {
		config.ArgoCd = expandUpdateArgoCdConfig(v)
	}

	return config
}

func expandArgoCdConfigRequest(tfList []any) *types.ArgoCdConfigRequest {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	config := &types.ArgoCdConfigRequest{}

	if v, ok := tfMap["aws_idc"].([]any); ok && len(v) > 0 {
		config.AwsIdc = expandArgoCdAwsIdcConfigRequest(v)
	}

	if v, ok := tfMap["namespace"].(string); ok && v != "" {
		config.Namespace = aws.String(v)
	}

	if v, ok := tfMap["network_access"].([]any); ok && len(v) > 0 {
		config.NetworkAccess = expandArgoCdNetworkAccessConfigRequest(v)
	}

	if v, ok := tfMap["rbac_role_mappings"].([]any); ok && len(v) > 0 {
		config.RbacRoleMappings = expandArgoCdRoleMappings(v)
	}

	return config
}

func expandUpdateArgoCdConfig(tfList []any) *types.UpdateArgoCdConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	config := &types.UpdateArgoCdConfig{}

	if v, ok := tfMap["network_access"].([]any); ok && len(v) > 0 {
		config.NetworkAccess = expandArgoCdNetworkAccessConfigRequest(v)
	}

	if v, ok := tfMap["rbac_role_mappings"].([]any); ok && len(v) > 0 {
		config.RbacRoleMappings = expandUpdateRoleMappings(v)
	}

	return config
}

func expandArgoCdAwsIdcConfigRequest(tfList []any) *types.ArgoCdAwsIdcConfigRequest {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	config := &types.ArgoCdAwsIdcConfigRequest{}

	if v, ok := tfMap["idc_instance_arn"].(string); ok && v != "" {
		config.IdcInstanceArn = aws.String(v)
	}

	if v, ok := tfMap["idc_region"].(string); ok && v != "" {
		config.IdcRegion = aws.String(v)
	}

	return config
}

func expandArgoCdNetworkAccessConfigRequest(tfList []any) *types.ArgoCdNetworkAccessConfigRequest {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	config := &types.ArgoCdNetworkAccessConfigRequest{}

	if v, ok := tfMap["vpce_ids"].(*schema.Set); ok && v.Len() > 0 {
		for _, item := range v.List() {
			if str, ok := item.(string); ok {
				config.VpceIds = append(config.VpceIds, str)
			}
		}
	}

	return config
}

func expandArgoCdRoleMappings(tfList []any) []types.ArgoCdRoleMapping {
	if len(tfList) == 0 {
		return nil
	}

	var mappings []types.ArgoCdRoleMapping
	for _, raw := range tfList {
		tfMap := raw.(map[string]any)
		mapping := types.ArgoCdRoleMapping{}

		if v, ok := tfMap["identities"].([]any); ok && len(v) > 0 {
			mapping.Identities = expandSsoIdentities(v)
		}

		if v, ok := tfMap["role"].(string); ok && v != "" {
			mapping.Role = types.ArgoCdRole(v)
		}

		mappings = append(mappings, mapping)
	}

	return mappings
}

func expandUpdateRoleMappings(tfList []any) *types.UpdateRoleMappings {
	if len(tfList) == 0 {
		return nil
	}

	mappings := &types.UpdateRoleMappings{}
	for _, raw := range tfList {
		tfMap := raw.(map[string]any)
		mapping := types.ArgoCdRoleMapping{}

		if v, ok := tfMap["identities"].([]any); ok && len(v) > 0 {
			mapping.Identities = expandSsoIdentities(v)
		}

		if v, ok := tfMap["role"].(string); ok && v != "" {
			mapping.Role = types.ArgoCdRole(v)
		}

		mappings.AddOrUpdateRoleMappings = append(mappings.AddOrUpdateRoleMappings, mapping)
	}

	return mappings
}

func expandSsoIdentities(tfList []any) []types.SsoIdentity {
	if len(tfList) == 0 {
		return nil
	}

	var identities []types.SsoIdentity
	for _, raw := range tfList {
		tfMap := raw.(map[string]any)
		identity := types.SsoIdentity{}

		if v, ok := tfMap["id"].(string); ok && v != "" {
			identity.Id = aws.String(v)
		}

		if v, ok := tfMap["type"].(string); ok && v != "" {
			identity.Type = types.SsoIdentityType(v)
		}

		identities = append(identities, identity)
	}

	return identities
}

func flattenCapabilityConfiguration(config *types.CapabilityConfigurationResponse) []any {
	if config == nil {
		return nil
	}

	tfMap := map[string]any{}

	if config.ArgoCd != nil {
		tfMap["argo_cd"] = flattenArgoCdConfigResponse(config.ArgoCd)
	}

	if len(tfMap) == 0 {
		return nil
	}

	return []any{tfMap}
}

func flattenArgoCdConfigResponse(config *types.ArgoCdConfigResponse) []any {
	if config == nil {
		return nil
	}

	tfMap := map[string]any{}

	if config.AwsIdc != nil {
		tfMap["aws_idc"] = flattenArgoCdAwsIdcConfigResponse(config.AwsIdc)
	}

	if config.Namespace != nil {
		tfMap["namespace"] = aws.ToString(config.Namespace)
	}

	if config.NetworkAccess != nil {
		tfMap["network_access"] = flattenArgoCdNetworkAccessConfigResponse(config.NetworkAccess)
	}

	if len(config.RbacRoleMappings) > 0 {
		tfMap["rbac_role_mappings"] = flattenArgoCdRoleMappings(config.RbacRoleMappings)
	}

	return []any{tfMap}
}

func flattenArgoCdAwsIdcConfigResponse(config *types.ArgoCdAwsIdcConfigResponse) []any {
	if config == nil {
		return nil
	}

	tfMap := map[string]any{}

	if config.IdcInstanceArn != nil {
		tfMap["idc_instance_arn"] = aws.ToString(config.IdcInstanceArn)
	}

	if config.IdcRegion != nil {
		tfMap["idc_region"] = aws.ToString(config.IdcRegion)
	}

	return []any{tfMap}
}

func flattenArgoCdNetworkAccessConfigResponse(config *types.ArgoCdNetworkAccessConfigResponse) []any {
	if config == nil {
		return nil
	}

	tfMap := map[string]any{}

	if len(config.VpceIds) > 0 {
		tfMap["vpce_ids"] = config.VpceIds
	}

	return []any{tfMap}
}

func flattenArgoCdRoleMappings(mappings []types.ArgoCdRoleMapping) []any {
	if len(mappings) == 0 {
		return nil
	}

	var tfList []any
	for _, mapping := range mappings {
		tfMap := map[string]any{}

		if len(mapping.Identities) > 0 {
			tfMap["identities"] = flattenSsoIdentities(mapping.Identities)
		}

		tfMap["role"] = string(mapping.Role)

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenSsoIdentities(identities []types.SsoIdentity) []any {
	if len(identities) == 0 {
		return nil
	}

	var tfList []any
	for _, identity := range identities {
		tfMap := map[string]any{
			"id":   aws.ToString(identity.Id),
			"type": string(identity.Type),
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func findCapabilityByTwoPartKey(ctx context.Context, conn *eks.Client, clusterName, capabilityName string) (*types.Capability, error) {
	input := &eks.DescribeCapabilityInput{
		CapabilityName: aws.String(capabilityName),
		ClusterName:    aws.String(clusterName),
	}

	output, err := conn.DescribeCapability(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &sdkretry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Capability == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Capability, nil
}

func findCapabilityUpdateByThreePartKey(ctx context.Context, conn *eks.Client, clusterName, capabilityName, id string) (*types.Update, error) {
	input := &eks.DescribeUpdateInput{
		Name:           aws.String(clusterName),
		UpdateId:       aws.String(id),
		CapabilityName: aws.String(capabilityName),
	}

	output, err := conn.DescribeUpdate(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &sdkretry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Update == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Update, nil
}

func statusCapability(ctx context.Context, conn *eks.Client, clusterName, capabilityName string) sdkretry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findCapabilityByTwoPartKey(ctx, conn, clusterName, capabilityName)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusCapabilityUpdate(ctx context.Context, conn *eks.Client, clusterName, capabilityName, id string) sdkretry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findCapabilityUpdateByThreePartKey(ctx, conn, clusterName, capabilityName, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitCapabilityCreated(ctx context.Context, conn *eks.Client, clusterName, capabilityName string, timeout time.Duration) (*types.Capability, error) {
	stateConf := sdkretry.StateChangeConf{
		Pending: enum.Slice(types.CapabilityStatusCreating),
		Target:  enum.Slice(types.CapabilityStatusActive),
		Refresh: statusCapability(ctx, conn, clusterName, capabilityName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Capability); ok {
		return output, err
	}

	return nil, err
}

func waitCapabilityDeleted(ctx context.Context, conn *eks.Client, clusterName, capabilityName string, timeout time.Duration) (*types.Capability, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending: enum.Slice(types.CapabilityStatusActive, types.CapabilityStatusDeleting),
		Target:  []string{},
		Refresh: statusCapability(ctx, conn, clusterName, capabilityName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Capability); ok {
		return output, err
	}

	return nil, err
}

func waitCapabilityUpdateSuccessful(ctx context.Context, conn *eks.Client, clusterName, capabilityName, id string, timeout time.Duration) (*types.Update, error) {
	stateConf := sdkretry.StateChangeConf{
		Pending: enum.Slice(types.UpdateStatusInProgress),
		Target:  enum.Slice(types.UpdateStatusSuccessful),
		Refresh: statusCapabilityUpdate(ctx, conn, clusterName, capabilityName, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Update); ok {
		return output, err
	}

	return nil, err
}
