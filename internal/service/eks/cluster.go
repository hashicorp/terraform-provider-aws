package eks

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceClusterCreate,
		Read:   resourceClusterRead,
		Update: resourceClusterUpdate,
		Delete: resourceClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled_cluster_log_types": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(eks.LogType_Values(), true),
				},
				Set: schema.HashString,
			},
			"encryption_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
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
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_family": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(eks.IpFamily_Values(), false),
						},
						"service_ipv4_cidr": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
							ValidateFunc: validation.All(
								validation.IsCIDRNetwork(12, 24),
								validation.StringMatch(regexp.MustCompile(`^(10|172\.(1[6-9]|2[0-9]|3[0-1])|192\.168)\..*`), "must be within 10.0.0.0/8, 172.16.0.0/12, or 192.168.0.0/16"),
							),
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
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
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
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

func resourceClusterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EKSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	name := d.Get("name").(string)

	input := &eks.CreateClusterInput{
		EncryptionConfig:   testAccClusterConfig_expandEncryption(d.Get("encryption_config").([]interface{})),
		Logging:            expandLoggingTypes(d.Get("enabled_cluster_log_types").(*schema.Set)),
		Name:               aws.String(name),
		ResourcesVpcConfig: testAccClusterConfig_expandVPCRequest(d.Get("vpc_config").([]interface{})),
		RoleArn:            aws.String(d.Get("role_arn").(string)),
	}

	if _, ok := d.GetOk("kubernetes_network_config"); ok {
		input.KubernetesNetworkConfig = expandNetworkConfigRequest(d.Get("kubernetes_network_config").([]interface{}))
	}

	if v, ok := d.GetOk("version"); ok {
		input.Version = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating EKS Cluster: %s", input)
	var output *eks.CreateClusterOutput
	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error

		output, err = conn.CreateCluster(input)

		// InvalidParameterException: roleArn, arn:aws:iam::123456789012:role/XXX, does not exist
		if tfawserr.ErrMessageContains(err, eks.ErrCodeInvalidParameterException, "does not exist") {
			return resource.RetryableError(err)
		}

		// InvalidParameterException: Error in role params
		if tfawserr.ErrMessageContains(err, eks.ErrCodeInvalidParameterException, "Error in role params") {
			return resource.RetryableError(err)
		}

		if tfawserr.ErrMessageContains(err, eks.ErrCodeInvalidParameterException, "Role could not be assumed because the trusted entity is not correct") {
			return resource.RetryableError(err)
		}

		// InvalidParameterException: The provided role doesn't have the Amazon EKS Managed Policies associated with it. Please ensure the following policy is attached: arn:aws:iam::aws:policy/AmazonEKSClusterPolicy
		if tfawserr.ErrMessageContains(err, eks.ErrCodeInvalidParameterException, "The provided role doesn't have the Amazon EKS Managed Policies associated with it") {
			return resource.RetryableError(err)
		}

		// InvalidParameterException: IAM role's policy must include the `ec2:DescribeSubnets` action
		if tfawserr.ErrMessageContains(err, eks.ErrCodeInvalidParameterException, "IAM role's policy must include") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateCluster(input)
	}

	if err != nil {
		return fmt.Errorf("error creating EKS Cluster (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.Cluster.Name))

	_, err = waitClusterCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return fmt.Errorf("error waiting for EKS Cluster (%s) to create: %w", d.Id(), err)
	}

	return resourceClusterRead(d, meta)
}

func resourceClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EKSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	cluster, err := FindClusterByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EKS Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EKS Cluster (%s): %w", d.Id(), err)
	}

	d.Set("arn", cluster.Arn)

	if err := d.Set("certificate_authority", flattenCertificate(cluster.CertificateAuthority)); err != nil {
		return fmt.Errorf("error setting certificate_authority: %w", err)
	}

	d.Set("created_at", aws.TimeValue(cluster.CreatedAt).String())

	if err := d.Set("enabled_cluster_log_types", flattenEnabledLogTypes(cluster.Logging)); err != nil {
		return fmt.Errorf("error setting enabled_cluster_log_types: %w", err)
	}

	if err := d.Set("encryption_config", flattenEncryptionConfig(cluster.EncryptionConfig)); err != nil {
		return fmt.Errorf("error setting encryption_config: %w", err)
	}

	d.Set("endpoint", cluster.Endpoint)

	if err := d.Set("identity", flattenIdentity(cluster.Identity)); err != nil {
		return fmt.Errorf("error setting identity: %w", err)
	}

	if err := d.Set("kubernetes_network_config", flattenNetworkConfig(cluster.KubernetesNetworkConfig)); err != nil {
		return fmt.Errorf("error setting kubernetes_network_config: %w", err)
	}

	d.Set("name", cluster.Name)
	d.Set("platform_version", cluster.PlatformVersion)
	d.Set("role_arn", cluster.RoleArn)
	d.Set("status", cluster.Status)
	d.Set("version", cluster.Version)

	if err := d.Set("vpc_config", flattenVPCConfigResponse(cluster.ResourcesVpcConfig)); err != nil {
		return fmt.Errorf("error setting vpc_config: %w", err)
	}

	tags := KeyValueTags(cluster.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EKSConn

	// Do any version update first.
	if d.HasChange("version") {
		input := &eks.UpdateClusterVersionInput{
			Name:    aws.String(d.Id()),
			Version: aws.String(d.Get("version").(string)),
		}

		log.Printf("[DEBUG] Updating EKS Cluster (%s) version: %s", d.Id(), input)
		output, err := conn.UpdateClusterVersion(input)

		if err != nil {
			return fmt.Errorf("error updating EKS Cluster (%s) version: %w", d.Id(), err)
		}

		updateID := aws.StringValue(output.Update.Id)

		_, err = waitClusterUpdateSuccessful(conn, d.Id(), updateID, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return fmt.Errorf("error waiting for EKS Cluster (%s) version update (%s): %w", d.Id(), updateID, err)
		}
	}

	if d.HasChange("encryption_config") {
		o, n := d.GetChange("encryption_config")

		if len(o.([]interface{})) == 0 && len(n.([]interface{})) == 1 {
			input := &eks.AssociateEncryptionConfigInput{
				ClusterName:      aws.String(d.Id()),
				EncryptionConfig: testAccClusterConfig_expandEncryption(d.Get("encryption_config").([]interface{})),
			}

			log.Printf("[DEBUG] Associating EKS Cluster (%s) encryption config: %s", d.Id(), input)
			output, err := conn.AssociateEncryptionConfig(input)

			if err != nil {
				return fmt.Errorf("error associating EKS Cluster (%s) encryption config: %w", d.Id(), err)
			}

			updateID := aws.StringValue(output.Update.Id)

			_, err = waitClusterUpdateSuccessful(conn, d.Id(), updateID, d.Timeout(schema.TimeoutUpdate))

			if err != nil {
				return fmt.Errorf("error waiting for EKS Cluster (%s) encryption config association (%s): %w", d.Id(), updateID, err)
			}
		}
	}

	if d.HasChange("enabled_cluster_log_types") {
		input := &eks.UpdateClusterConfigInput{
			Logging: expandLoggingTypes(d.Get("enabled_cluster_log_types").(*schema.Set)),
			Name:    aws.String(d.Id()),
		}

		log.Printf("[DEBUG] Updating EKS Cluster (%s) logging: %s", d.Id(), input)
		output, err := conn.UpdateClusterConfig(input)

		if err != nil {
			return fmt.Errorf("error updating EKS Cluster (%s) logging: %w", d.Id(), err)
		}

		updateID := aws.StringValue(output.Update.Id)

		_, err = waitClusterUpdateSuccessful(conn, d.Id(), updateID, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return fmt.Errorf("error waiting for EKS Cluster (%s) logging update (%s): %w", d.Id(), updateID, err)
		}
	}

	if d.HasChanges("vpc_config.0.endpoint_private_access", "vpc_config.0.endpoint_public_access", "vpc_config.0.public_access_cidrs") {
		input := &eks.UpdateClusterConfigInput{
			Name:               aws.String(d.Id()),
			ResourcesVpcConfig: testAccClusterConfig_expandVPCUpdateRequest(d.Get("vpc_config").([]interface{})),
		}

		log.Printf("[DEBUG] Updating EKS Cluster (%s) VPC config: %s", d.Id(), input)
		output, err := conn.UpdateClusterConfig(input)

		if err != nil {
			return fmt.Errorf("error updating EKS Cluster (%s) VPC config: %w", d.Id(), err)
		}

		updateID := aws.StringValue(output.Update.Id)

		_, err = waitClusterUpdateSuccessful(conn, d.Id(), updateID, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return fmt.Errorf("error waiting for EKS Cluster (%s) VPC config update (%s): %w", d.Id(), updateID, err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceClusterRead(d, meta)
}

func resourceClusterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EKSConn

	log.Printf("[DEBUG] Deleting EKS Cluster: %s", d.Id())

	input := &eks.DeleteClusterInput{
		Name: aws.String(d.Id()),
	}

	// If a cluster is scaling up due to load a delete request will fail
	// This is a temporary workaround until EKS supports multiple parallel mutating operations
	err := tfresource.RetryConfigContext(context.Background(), 0*time.Second, 1*time.Minute, 0*time.Second, 30*time.Second, clusterDeleteRetryTimeout, func() *resource.RetryError {
		var err error

		_, err = conn.DeleteCluster(input)

		if tfawserr.ErrMessageContains(err, eks.ErrCodeResourceInUseException, "in progress") {
			log.Printf("[DEBUG] eks cluster update in progress: %v", err)
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteCluster(input)
	}

	if tfawserr.ErrCodeEquals(err, eks.ErrCodeResourceNotFoundException) {
		return nil
	}

	// Sometimes the EKS API returns the ResourceNotFound error in this form:
	// ClientException: No cluster found for name: tf-acc-test-0o1f8
	if tfawserr.ErrMessageContains(err, eks.ErrCodeClientException, "No cluster found for name:") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EKS Cluster (%s): %w", d.Id(), err)
	}

	if _, err = waitClusterDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for EKS Cluster (%s) to delete: %w", d.Id(), err)
	}

	return nil
}

func testAccClusterConfig_expandEncryption(tfList []interface{}) []*eks.EncryptionConfig {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*eks.EncryptionConfig

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := &eks.EncryptionConfig{
			Provider: expandProvider(tfMap["provider"].([]interface{})),
		}

		if v, ok := tfMap["resources"].(*schema.Set); ok && v.Len() > 0 {
			apiObject.Resources = flex.ExpandStringSet(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandProvider(tfList []interface{}) *eks.Provider {
	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	apiObject := &eks.Provider{}

	if v, ok := tfMap["key_arn"].(string); ok && v != "" {
		apiObject.KeyArn = aws.String(v)
	}

	return apiObject
}

func testAccClusterConfig_expandVPCRequest(l []interface{}) *eks.VpcConfigRequest {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	vpcConfigRequest := &eks.VpcConfigRequest{
		EndpointPrivateAccess: aws.Bool(m["endpoint_private_access"].(bool)),
		EndpointPublicAccess:  aws.Bool(m["endpoint_public_access"].(bool)),
		SecurityGroupIds:      flex.ExpandStringSet(m["security_group_ids"].(*schema.Set)),
		SubnetIds:             flex.ExpandStringSet(m["subnet_ids"].(*schema.Set)),
	}

	if v, ok := m["public_access_cidrs"].(*schema.Set); ok && v.Len() > 0 {
		vpcConfigRequest.PublicAccessCidrs = flex.ExpandStringSet(v)
	}

	return vpcConfigRequest
}

func testAccClusterConfig_expandVPCUpdateRequest(l []interface{}) *eks.VpcConfigRequest {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	vpcConfigRequest := &eks.VpcConfigRequest{
		EndpointPrivateAccess: aws.Bool(m["endpoint_private_access"].(bool)),
		EndpointPublicAccess:  aws.Bool(m["endpoint_public_access"].(bool)),
	}

	if v, ok := m["public_access_cidrs"].(*schema.Set); ok && v.Len() > 0 {
		vpcConfigRequest.PublicAccessCidrs = flex.ExpandStringSet(v)
	}

	return vpcConfigRequest
}

func expandNetworkConfigRequest(tfList []interface{}) *eks.KubernetesNetworkConfigRequest {
	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	apiObject := &eks.KubernetesNetworkConfigRequest{}

	if v, ok := tfMap["service_ipv4_cidr"].(string); ok && v != "" {
		apiObject.ServiceIpv4Cidr = aws.String(v)
	}

	if v, ok := tfMap["ip_family"].(string); ok && v != "" {
		apiObject.IpFamily = aws.String(v)
	}

	return apiObject
}

func expandLoggingTypes(vEnabledLogTypes *schema.Set) *eks.Logging {
	vEksLogTypes := []interface{}{}
	for _, eksLogType := range eks.LogType_Values() {
		vEksLogTypes = append(vEksLogTypes, eksLogType)
	}
	vAllLogTypes := schema.NewSet(schema.HashString, vEksLogTypes)

	return &eks.Logging{
		ClusterLogging: []*eks.LogSetup{
			{
				Enabled: aws.Bool(true),
				Types:   flex.ExpandStringSet(vEnabledLogTypes),
			},
			{
				Enabled: aws.Bool(false),
				Types:   flex.ExpandStringSet(vAllLogTypes.Difference(vEnabledLogTypes)),
			},
		},
	}
}

func flattenCertificate(certificate *eks.Certificate) []map[string]interface{} {
	if certificate == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"data": aws.StringValue(certificate.Data),
	}

	return []map[string]interface{}{m}
}

func flattenIdentity(identity *eks.Identity) []map[string]interface{} {
	if identity == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"oidc": flattenOIDC(identity.Oidc),
	}

	return []map[string]interface{}{m}
}

func flattenOIDC(oidc *eks.OIDC) []map[string]interface{} {
	if oidc == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"issuer": aws.StringValue(oidc.Issuer),
	}

	return []map[string]interface{}{m}
}

func flattenEncryptionConfig(apiObjects []*eks.EncryptionConfig) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"provider":  flattenProvider(apiObject.Provider),
			"resources": aws.StringValueSlice(apiObject.Resources),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenProvider(apiObject *eks.Provider) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"key_arn": aws.StringValue(apiObject.KeyArn),
	}

	return []interface{}{tfMap}
}

func flattenVPCConfigResponse(vpcConfig *eks.VpcConfigResponse) []map[string]interface{} {
	if vpcConfig == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"cluster_security_group_id": aws.StringValue(vpcConfig.ClusterSecurityGroupId),
		"endpoint_private_access":   aws.BoolValue(vpcConfig.EndpointPrivateAccess),
		"endpoint_public_access":    aws.BoolValue(vpcConfig.EndpointPublicAccess),
		"security_group_ids":        flex.FlattenStringSet(vpcConfig.SecurityGroupIds),
		"subnet_ids":                flex.FlattenStringSet(vpcConfig.SubnetIds),
		"public_access_cidrs":       flex.FlattenStringSet(vpcConfig.PublicAccessCidrs),
		"vpc_id":                    aws.StringValue(vpcConfig.VpcId),
	}

	return []map[string]interface{}{m}
}

func flattenEnabledLogTypes(logging *eks.Logging) *schema.Set {
	enabledLogTypes := []*string{}

	if logging != nil {
		logSetups := logging.ClusterLogging
		for _, logSetup := range logSetups {
			if logSetup == nil || !aws.BoolValue(logSetup.Enabled) {
				continue
			}

			enabledLogTypes = append(enabledLogTypes, logSetup.Types...)
		}
	}

	return flex.FlattenStringSet(enabledLogTypes)
}

func flattenNetworkConfig(apiObject *eks.KubernetesNetworkConfigResponse) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"service_ipv4_cidr": aws.StringValue(apiObject.ServiceIpv4Cidr),
		"ip_family":         aws.StringValue(apiObject.IpFamily),
	}

	return []interface{}{tfMap}
}
