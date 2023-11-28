// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_eks_cluster", name="Cluster")
// @Tags(identifierAttribute="arn")
func resourceCluster() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterCreate,
		ReadWithoutTimeout:   resourceClusterRead,
		UpdateWithoutTimeout: resourceClusterUpdate,
		DeleteWithoutTimeout: resourceClusterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
			customdiff.ForceNewIfChange("encryption_config", func(_ context.Context, old, new, meta interface{}) bool {
				// You cannot disable envelope encryption after enabling it. This action is irreversible.
				return len(old.([]interface{})) == 1 && len(new.([]interface{})) == 0
			}),
		),

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_authority": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled_cluster_log_types": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[types.LogType](),
				},
			},
			"encryption_config": {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ConflictsWith: []string{"outpost_config"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"provider": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key_arn": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"resources": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(Resources_Values(), false),
							},
						},
					},
				},
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identity": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"oidc": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"issuer": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"kubernetes_network_config": {
				Type:          schema.TypeList,
				Optional:      true,
				Computed:      true,
				MaxItems:      1,
				ConflictsWith: []string{"outpost_config"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_family": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[types.IpFamily](),
						},
						"service_ipv4_cidr": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
							ValidateFunc: validation.All(
								validation.IsCIDRNetwork(12, 24),
								validation.StringMatch(regexache.MustCompile(`^(10|172\.(1[6-9]|2[0-9]|3[0-1])|192\.168)\..*`), "must be within 10.0.0.0/8, 172.16.0.0/12, or 192.168.0.0/16"),
							),
						},
						"service_ipv6_cidr": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validClusterName,
			},
			"outpost_config": {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ConflictsWith: []string{"encryption_config", "kubernetes_network_config"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"control_plane_instance_type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"control_plane_placement": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"group_name": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
								},
							},
						},
						"outpost_arns": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"platform_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"vpc_config": {
				Type:     schema.TypeList,
				MinItems: 1,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cluster_security_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"endpoint_private_access": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"endpoint_public_access": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"public_access_cidrs": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: verify.ValidCIDRNetworkAddress,
							},
						},
						"security_group_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	name := d.Get("name").(string)
	input := &eks.CreateClusterInput{
		EncryptionConfig:   expandEncryptionConfig(d.Get("encryption_config").([]interface{})),
		Logging:            expandLogging(d.Get("enabled_cluster_log_types").(*schema.Set)),
		Name:               aws.String(name),
		ResourcesVpcConfig: expandVPCConfigRequestForCreate(d.Get("vpc_config").([]interface{})),
		RoleArn:            aws.String(d.Get("role_arn").(string)),
		Tags:               getTagsIn(ctx),
	}

	if _, ok := d.GetOk("kubernetes_network_config"); ok {
		input.KubernetesNetworkConfig = expandKubernetesNetworkConfigRequest(d.Get("kubernetes_network_config").([]interface{}))
	}

	if _, ok := d.GetOk("outpost_config"); ok {
		input.OutpostConfig = expandOutpostConfigRequest(d.Get("outpost_config").([]interface{}))
	}

	if v, ok := d.GetOk("version"); ok {
		input.Version = aws.String(v.(string))
	}

	outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateCluster(ctx, input)
		},
		func(err error) (bool, error) {
			// InvalidParameterException: roleArn, arn:aws:iam::123456789012:role/XXX, does not exist
			if errs.IsAErrorMessageContains[*types.InvalidParameterException](err, "does not exist") {
				return true, err
			}

			// InvalidParameterException: Error in role params
			if errs.IsAErrorMessageContains[*types.InvalidParameterException](err, "Error in role params") {
				return true, err
			}

			if errs.IsAErrorMessageContains[*types.InvalidParameterException](err, "Role could not be assumed because the trusted entity is not correct") {
				return true, err
			}

			// InvalidParameterException: The provided role doesn't have the Amazon EKS Managed Policies associated with it. Please ensure the following policy is attached: arn:aws:iam::aws:policy/AmazonEKSClusterPolicy
			if errs.IsAErrorMessageContains[*types.InvalidParameterException](err, "The provided role doesn't have the Amazon EKS Managed Policies associated with it") {
				return true, err
			}

			// InvalidParameterException: IAM role's policy must include the `ec2:DescribeSubnets` action
			if errs.IsAErrorMessageContains[*types.InvalidParameterException](err, "IAM role's policy must include") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return diag.Errorf("creating EKS Cluster (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*eks.CreateClusterOutput).Cluster.Name))

	if _, err := waitClusterCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for EKS Cluster (%s) create: %s", d.Id(), err)
	}

	return resourceClusterRead(ctx, d, meta)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	cluster, err := findClusterByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EKS Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading EKS Cluster (%s): %s", d.Id(), err)
	}

	d.Set("arn", cluster.Arn)
	if err := d.Set("certificate_authority", flattenCertificate(cluster.CertificateAuthority)); err != nil {
		return diag.Errorf("setting certificate_authority: %s", err)
	}
	// cluster_id is only relevant for clusters on Outposts.
	if cluster.OutpostConfig != nil {
		d.Set("cluster_id", cluster.Id)
	}
	d.Set("created_at", aws.ToTime(cluster.CreatedAt).String())
	if err := d.Set("enabled_cluster_log_types", flattenLogging(cluster.Logging)); err != nil {
		return diag.Errorf("setting enabled_cluster_log_types: %s", err)
	}
	if err := d.Set("encryption_config", flattenEncryptionConfig(cluster.EncryptionConfig)); err != nil {
		return diag.Errorf("setting encryption_config: %s", err)
	}
	d.Set("endpoint", cluster.Endpoint)
	if err := d.Set("identity", flattenIdentity(cluster.Identity)); err != nil {
		return diag.Errorf("setting identity: %s", err)
	}
	if err := d.Set("kubernetes_network_config", flattenKubernetesNetworkConfigResponse(cluster.KubernetesNetworkConfig)); err != nil {
		return diag.Errorf("setting kubernetes_network_config: %s", err)
	}
	d.Set("name", cluster.Name)
	if err := d.Set("outpost_config", flattenOutpostConfigResponse(cluster.OutpostConfig)); err != nil {
		return diag.Errorf("setting outpost_config: %s", err)
	}
	d.Set("platform_version", cluster.PlatformVersion)
	d.Set("role_arn", cluster.RoleArn)
	d.Set("status", cluster.Status)
	d.Set("version", cluster.Version)
	if err := d.Set("vpc_config", flattenVPCConfigResponse(cluster.ResourcesVpcConfig)); err != nil {
		return diag.Errorf("setting vpc_config: %s", err)
	}

	setTagsOut(ctx, cluster.Tags)

	return nil
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	// Do any version update first.
	if d.HasChange("version") {
		input := &eks.UpdateClusterVersionInput{
			Name:    aws.String(d.Id()),
			Version: aws.String(d.Get("version").(string)),
		}

		output, err := conn.UpdateClusterVersion(ctx, input)

		if err != nil {
			return diag.Errorf("updating EKS Cluster (%s) version: %s", d.Id(), err)
		}

		updateID := aws.ToString(output.Update.Id)

		if _, err := waitClusterUpdateSuccessful(ctx, conn, d.Id(), updateID, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("waiting for EKS Cluster (%s) version update (%s): %s", d.Id(), updateID, err)
		}
	}

	if d.HasChange("encryption_config") {
		o, n := d.GetChange("encryption_config")

		if len(o.([]interface{})) == 0 && len(n.([]interface{})) == 1 {
			input := &eks.AssociateEncryptionConfigInput{
				ClusterName:      aws.String(d.Id()),
				EncryptionConfig: expandEncryptionConfig(d.Get("encryption_config").([]interface{})),
			}

			output, err := conn.AssociateEncryptionConfig(ctx, input)

			if err != nil {
				return diag.Errorf("associating EKS Cluster (%s) encryption config: %s", d.Id(), err)
			}

			updateID := aws.ToString(output.Update.Id)

			if _, err := waitClusterUpdateSuccessful(ctx, conn, d.Id(), updateID, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return diag.Errorf("waiting for EKS Cluster (%s) encryption config association (%s): %s", d.Id(), updateID, err)
			}
		}
	}

	if d.HasChange("enabled_cluster_log_types") {
		input := &eks.UpdateClusterConfigInput{
			Logging: expandLogging(d.Get("enabled_cluster_log_types").(*schema.Set)),
			Name:    aws.String(d.Id()),
		}

		output, err := conn.UpdateClusterConfig(ctx, input)

		if err != nil {
			return diag.Errorf("updating EKS Cluster (%s) logging: %s", d.Id(), err)
		}

		updateID := aws.ToString(output.Update.Id)

		if _, err := waitClusterUpdateSuccessful(ctx, conn, d.Id(), updateID, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("waiting for EKS Cluster (%s) logging update (%s): %s", d.Id(), updateID, err)
		}
	}
	if d.HasChanges("vpc_config.0.endpoint_private_access", "vpc_config.0.endpoint_public_access", "vpc_config.0.public_access_cidrs") {
		config := &types.VpcConfigRequest{
			EndpointPrivateAccess: aws.Bool(d.Get("vpc_config.0.endpoint_private_access").(bool)),
			EndpointPublicAccess:  aws.Bool(d.Get("vpc_config.0.endpoint_public_access").(bool)),
		}

		if v, ok := d.GetOk("vpc_config.0.public_access_cidrs"); ok && v.(*schema.Set).Len() > 0 {
			config.PublicAccessCidrs = flex.ExpandStringValueSet(v.(*schema.Set))
		}

		if err := updateVPCConfig(ctx, conn, d.Id(), config, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.FromErr(err)
		}
	}

	// API only allows one type of update at at time.
	if d.HasChange("vpc_config.0.subnet_ids") {
		config := &types.VpcConfigRequest{
			SubnetIds: flex.ExpandStringValueSet(d.Get("vpc_config.0.subnet_ids").(*schema.Set)),
		}

		if err := updateVPCConfig(ctx, conn, d.Id(), config, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("vpc_config.0.security_group_ids") {
		config := &types.VpcConfigRequest{
			SecurityGroupIds: flex.ExpandStringValueSet(d.Get("vpc_config.0.security_group_ids").(*schema.Set)),
		}

		if err := updateVPCConfig(ctx, conn, d.Id(), config, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceClusterRead(ctx, d, meta)
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	input := &eks.DeleteClusterInput{
		Name: aws.String(d.Id()),
	}

	// If a cluster is scaling up due to load a delete request will fail
	// This is a temporary workaround until EKS supports multiple parallel mutating operations
	const (
		timeout = 60 * time.Minute
	)
	log.Printf("[DEBUG] Deleting EKS Cluster: %s", d.Id())
	err := tfresource.Retry(ctx, timeout, func() *retry.RetryError {
		var err error

		_, err = conn.DeleteCluster(ctx, input)

		if errs.IsAErrorMessageContains[*types.ResourceInUseException](err, "in progress") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	}, tfresource.WithDelayRand(1*time.Minute), tfresource.WithPollInterval(30*time.Second))

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteCluster(ctx, input)
	}

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil
	}

	// Sometimes the EKS API returns the ResourceNotFound error in this form:
	// ClientException: No cluster found for name: tf-acc-test-0o1f8
	if errs.IsAErrorMessageContains[*types.ClientException](err, "No cluster found for name:") {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting EKS Cluster (%s): %s", d.Id(), err)
	}

	if _, err := waitClusterDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for EKS Cluster (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func findClusterByName(ctx context.Context, conn *eks.Client, name string) (*types.Cluster, error) {
	input := &eks.DescribeClusterInput{
		Name: aws.String(name),
	}

	output, err := conn.DescribeCluster(ctx, input)

	// Sometimes the EKS API returns the ResourceNotFound error in this form:
	// ClientException: No cluster found for name: tf-acc-test-0o1f8
	if errs.IsA[*types.ResourceNotFoundException](err) || errs.IsAErrorMessageContains[*types.ClientException](err, "No cluster found for name:") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Cluster == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Cluster, nil
}

func updateVPCConfig(ctx context.Context, conn *eks.Client, name string, config *types.VpcConfigRequest, timeout time.Duration) error {
	input := &eks.UpdateClusterConfigInput{
		Name:               aws.String(name),
		ResourcesVpcConfig: config,
	}

	output, err := conn.UpdateClusterConfig(ctx, input)

	if err != nil {
		return fmt.Errorf("updating EKS Cluster (%s) VPC config: %s", name, err)
	}

	updateID := aws.ToString(output.Update.Id)

	if _, err := waitClusterUpdateSuccessful(ctx, conn, name, updateID, timeout); err != nil {
		return fmt.Errorf("waiting for EKS Cluster (%s) VPC config update (%s): %s", name, updateID, err)
	}

	return nil
}

func findClusterUpdateByTwoPartKey(ctx context.Context, conn *eks.Client, name, id string) (*types.Update, error) {
	input := &eks.DescribeUpdateInput{
		Name:     aws.String(name),
		UpdateId: aws.String(id),
	}

	output, err := conn.DescribeUpdate(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
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

func statusCluster(ctx context.Context, conn *eks.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findClusterByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusClusterUpdate(ctx context.Context, conn *eks.Client, name, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findClusterUpdateByTwoPartKey(ctx, conn, name, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitClusterCreated(ctx context.Context, conn *eks.Client, name string, timeout time.Duration) (*types.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ClusterStatusPending, types.ClusterStatusCreating),
		Target:  enum.Slice(types.ClusterStatusActive),
		Refresh: statusCluster(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Cluster); ok {
		return output, err
	}

	return nil, err
}

func waitClusterDeleted(ctx context.Context, conn *eks.Client, name string, timeout time.Duration) (*types.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ClusterStatusActive, types.ClusterStatusDeleting),
		Target:  []string{},
		Refresh: statusCluster(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Cluster); ok {
		return output, err
	}

	return nil, err
}

func waitClusterUpdateSuccessful(ctx context.Context, conn *eks.Client, name, id string, timeout time.Duration) (*types.Update, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.UpdateStatusInProgress),
		Target:  enum.Slice(types.UpdateStatusSuccessful),
		Refresh: statusClusterUpdate(ctx, conn, name, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Update); ok {
		if status := output.Status; status == types.UpdateStatusCancelled || status == types.UpdateStatusFailed {
			tfresource.SetLastError(err, errorDetailsError(output.Errors))
		}

		return output, err
	}

	return nil, err
}

func expandEncryptionConfig(tfList []interface{}) []types.EncryptionConfig {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.EncryptionConfig

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := types.EncryptionConfig{
			Provider: expandProvider(tfMap["provider"].([]interface{})),
		}

		if v, ok := tfMap["resources"].(*schema.Set); ok && v.Len() > 0 {
			apiObject.Resources = flex.ExpandStringValueSet(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandProvider(tfList []interface{}) *types.Provider {
	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	apiObject := &types.Provider{}

	if v, ok := tfMap["key_arn"].(string); ok && v != "" {
		apiObject.KeyArn = aws.String(v)
	}

	return apiObject
}

func expandOutpostConfigRequest(l []interface{}) *types.OutpostConfigRequest {
	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	outpostConfigRequest := &types.OutpostConfigRequest{}

	if v, ok := tfMap["control_plane_instance_type"].(string); ok && v != "" {
		outpostConfigRequest.ControlPlaneInstanceType = aws.String(v)
	}

	if v, ok := tfMap["control_plane_placement"].([]interface{}); ok {
		outpostConfigRequest.ControlPlanePlacement = expandControlPlanePlacement(v)
	}

	if v, ok := tfMap["outpost_arns"].(*schema.Set); ok && v.Len() > 0 {
		outpostConfigRequest.OutpostArns = flex.ExpandStringValueSet(v)
	}

	return outpostConfigRequest
}

func expandControlPlanePlacement(tfList []interface{}) *types.ControlPlanePlacementRequest {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	apiObject := &types.ControlPlanePlacementRequest{}

	if v, ok := tfMap["group_name"].(string); ok && v != "" {
		apiObject.GroupName = aws.String(v)
	}

	return apiObject
}

func expandVPCConfigRequestForCreate(l []interface{}) *types.VpcConfigRequest {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	vpcConfigRequest := &types.VpcConfigRequest{
		EndpointPrivateAccess: aws.Bool(m["endpoint_private_access"].(bool)),
		EndpointPublicAccess:  aws.Bool(m["endpoint_public_access"].(bool)),
		SecurityGroupIds:      flex.ExpandStringValueSet(m["security_group_ids"].(*schema.Set)),
		SubnetIds:             flex.ExpandStringValueSet(m["subnet_ids"].(*schema.Set)),
	}

	if v, ok := m["public_access_cidrs"].(*schema.Set); ok && v.Len() > 0 {
		vpcConfigRequest.PublicAccessCidrs = flex.ExpandStringValueSet(v)
	}

	return vpcConfigRequest
}

func expandKubernetesNetworkConfigRequest(tfList []interface{}) *types.KubernetesNetworkConfigRequest {
	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	apiObject := &types.KubernetesNetworkConfigRequest{}

	if v, ok := tfMap["service_ipv4_cidr"].(string); ok && v != "" {
		apiObject.ServiceIpv4Cidr = aws.String(v)
	}

	if v, ok := tfMap["ip_family"].(string); ok && v != "" {
		apiObject.IpFamily = types.IpFamily(v)
	}

	return apiObject
}

func expandLogging(vEnabledLogTypes *schema.Set) *types.Logging {
	allLogTypes := enum.EnumValues[types.LogType]()
	enabledLogTypes := flex.ExpandStringyValueSet[types.LogType](vEnabledLogTypes)

	return &types.Logging{
		ClusterLogging: []types.LogSetup{
			{
				Enabled: aws.Bool(true),
				Types:   enabledLogTypes,
			},
			{
				Enabled: aws.Bool(false),
				Types:   tfslices.RemoveAll(allLogTypes, enabledLogTypes...),
			},
		},
	}
}

func flattenCertificate(certificate *types.Certificate) []map[string]interface{} {
	if certificate == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"data": aws.ToString(certificate.Data),
	}

	return []map[string]interface{}{m}
}

func flattenIdentity(identity *types.Identity) []map[string]interface{} {
	if identity == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"oidc": flattenOIDC(identity.Oidc),
	}

	return []map[string]interface{}{m}
}

func flattenOIDC(oidc *types.OIDC) []map[string]interface{} {
	if oidc == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"issuer": aws.ToString(oidc.Issuer),
	}

	return []map[string]interface{}{m}
}

func flattenEncryptionConfig(apiObjects []types.EncryptionConfig) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"provider":  flattenProvider(apiObject.Provider),
			"resources": apiObject.Resources,
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenProvider(apiObject *types.Provider) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"key_arn": aws.ToString(apiObject.KeyArn),
	}

	return []interface{}{tfMap}
}

func flattenVPCConfigResponse(vpcConfig *types.VpcConfigResponse) []map[string]interface{} {
	if vpcConfig == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"cluster_security_group_id": aws.ToString(vpcConfig.ClusterSecurityGroupId),
		"endpoint_private_access":   vpcConfig.EndpointPrivateAccess,
		"endpoint_public_access":    vpcConfig.EndpointPublicAccess,
		"security_group_ids":        vpcConfig.SecurityGroupIds,
		"subnet_ids":                vpcConfig.SubnetIds,
		"public_access_cidrs":       vpcConfig.PublicAccessCidrs,
		"vpc_id":                    aws.ToString(vpcConfig.VpcId),
	}

	return []map[string]interface{}{m}
}

func flattenLogging(logging *types.Logging) []string {
	enabledLogTypes := []types.LogType{}

	if logging != nil {
		logSetups := logging.ClusterLogging
		for _, logSetup := range logSetups {
			if !aws.ToBool(logSetup.Enabled) {
				continue
			}

			enabledLogTypes = append(enabledLogTypes, logSetup.Types...)
		}
	}

	return enum.Slice(enabledLogTypes...)
}

func flattenKubernetesNetworkConfigResponse(apiObject *types.KubernetesNetworkConfigResponse) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"service_ipv4_cidr": aws.ToString(apiObject.ServiceIpv4Cidr),
		"service_ipv6_cidr": aws.ToString(apiObject.ServiceIpv6Cidr),
		"ip_family":         apiObject.IpFamily,
	}

	return []interface{}{tfMap}
}

func flattenOutpostConfigResponse(apiObject *types.OutpostConfigResponse) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"control_plane_instance_type": aws.ToString(apiObject.ControlPlaneInstanceType),
		"control_plane_placement":     flattenControlPlanePlacementResponse(apiObject.ControlPlanePlacement),
		"outpost_arns":                apiObject.OutpostArns,
	}

	return []interface{}{tfMap}
}

func flattenControlPlanePlacementResponse(apiObject *types.ControlPlanePlacementResponse) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"group_name": aws.ToString(apiObject.GroupName),
	}

	return []interface{}{tfMap}
}
