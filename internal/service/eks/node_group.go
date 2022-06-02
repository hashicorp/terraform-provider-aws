package eks

import (
	"context"
	"log"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceNodeGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNodeGroupCreate,
		ReadContext:   resourceNodeGroupRead,
		UpdateContext: resourceNodeGroupUpdate,
		DeleteContext: resourceNodeGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"ami_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(eks.AMITypes_Values(), false),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"capacity_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(eks.CapacityTypes_Values(), false),
			},
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validClusterName,
			},
			"disk_size": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"force_update_version": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"instance_types": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"launch_template": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ForceNew:      true,
							ConflictsWith: []string{"launch_template.0.name"},
							ValidateFunc:  verify.ValidLaunchTemplateID,
						},
						"name": {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ForceNew:      true,
							ConflictsWith: []string{"launch_template.0.id"},
							ValidateFunc:  verify.ValidLaunchTemplateName,
						},
						"version": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
					},
				},
			},
			"node_group_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"node_group_name_prefix"},
			},
			"node_group_name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"node_group_name"},
			},
			"node_role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"release_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"remote_access": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ec2_ssh_key": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"source_security_group_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"resources": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"autoscaling_groups": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"remote_access_security_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"scaling_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"desired_size": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"max_size": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"min_size": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"taint": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 50,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 63),
						},
						"value": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 63),
						},
						"effect": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(eks.TaintEffect_Values(), false),
						},
					},
				},
			},
			"update_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_unavailable": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 100),
							ExactlyOneOf: []string{
								"update_config.0.max_unavailable",
								"update_config.0.max_unavailable_percentage",
							},
						},
						"max_unavailable_percentage": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 100),
							ExactlyOneOf: []string{
								"update_config.0.max_unavailable",
								"update_config.0.max_unavailable_percentage",
							},
						},
					},
				},
			},
			"version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceNodeGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	clusterName := d.Get("cluster_name").(string)
	nodeGroupName := create.Name(d.Get("node_group_name").(string), d.Get("node_group_name_prefix").(string))
	id := NodeGroupCreateResourceID(clusterName, nodeGroupName)

	input := &eks.CreateNodegroupInput{
		ClientRequestToken: aws.String(resource.UniqueId()),
		ClusterName:        aws.String(clusterName),
		NodegroupName:      aws.String(nodeGroupName),
		NodeRole:           aws.String(d.Get("node_role_arn").(string)),
		Subnets:            flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
	}

	if v, ok := d.GetOk("ami_type"); ok {
		input.AmiType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("capacity_type"); ok {
		input.CapacityType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("disk_size"); ok {
		input.DiskSize = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("instance_types"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.InstanceTypes = flex.ExpandStringList(v.([]interface{}))
	}

	if v := d.Get("labels").(map[string]interface{}); len(v) > 0 {
		input.Labels = flex.ExpandStringMap(v)
	}

	if v := d.Get("launch_template").([]interface{}); len(v) > 0 {
		input.LaunchTemplate = expandLaunchTemplateSpecification(v)
	}

	if v, ok := d.GetOk("release_version"); ok {
		input.ReleaseVersion = aws.String(v.(string))
	}

	if v := d.Get("remote_access").([]interface{}); len(v) > 0 {
		input.RemoteAccess = expandRemoteAccessConfig(v)
	}

	if v, ok := d.GetOk("scaling_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ScalingConfig = expandNodegroupScalingConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("taint"); ok && v.(*schema.Set).Len() > 0 {
		input.Taints = expandTaints(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("update_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.UpdateConfig = expandNodegroupUpdateConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("version"); ok {
		input.Version = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	_, err := conn.CreateNodegroup(input)

	if err != nil {
		return diag.Errorf("error creating EKS Node Group (%s): %s", id, err)
	}

	d.SetId(id)

	_, err = waitNodegroupCreated(ctx, conn, clusterName, nodeGroupName, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return diag.Errorf("error waiting for EKS Node Group (%s) to create: %s", d.Id(), err)
	}

	return resourceNodeGroupRead(ctx, d, meta)
}

func resourceNodeGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	clusterName, nodeGroupName, err := NodeGroupParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	nodeGroup, err := FindNodegroupByClusterNameAndNodegroupName(conn, clusterName, nodeGroupName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EKS Node Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading EKS Node Group (%s): %s", d.Id(), err)
	}

	d.Set("ami_type", nodeGroup.AmiType)
	d.Set("arn", nodeGroup.NodegroupArn)
	d.Set("capacity_type", nodeGroup.CapacityType)
	d.Set("cluster_name", nodeGroup.ClusterName)
	d.Set("disk_size", nodeGroup.DiskSize)

	if err := d.Set("instance_types", aws.StringValueSlice(nodeGroup.InstanceTypes)); err != nil {
		return diag.Errorf("error setting instance_types: %s", err)
	}

	if err := d.Set("labels", aws.StringValueMap(nodeGroup.Labels)); err != nil {
		return diag.Errorf("error setting labels: %s", err)
	}

	if err := d.Set("launch_template", flattenLaunchTemplateSpecification(nodeGroup.LaunchTemplate)); err != nil {
		return diag.Errorf("error setting launch_template: %s", err)
	}

	d.Set("node_group_name", nodeGroup.NodegroupName)
	d.Set("node_group_name_prefix", create.NamePrefixFromName(aws.StringValue(nodeGroup.NodegroupName)))
	d.Set("node_role_arn", nodeGroup.NodeRole)
	d.Set("release_version", nodeGroup.ReleaseVersion)

	if err := d.Set("remote_access", flattenRemoteAccessConfig(nodeGroup.RemoteAccess)); err != nil {
		return diag.Errorf("error setting remote_access: %s", err)
	}

	if err := d.Set("resources", flattenNodeGroupResources(nodeGroup.Resources)); err != nil {
		return diag.Errorf("error setting resources: %s", err)
	}

	if nodeGroup.ScalingConfig != nil {
		if err := d.Set("scaling_config", []interface{}{flattenNodeGroupScalingConfig(nodeGroup.ScalingConfig)}); err != nil {
			return diag.Errorf("error setting scaling_config: %s", err)
		}
	} else {
		d.Set("scaling_config", nil)
	}

	d.Set("status", nodeGroup.Status)

	if err := d.Set("subnet_ids", aws.StringValueSlice(nodeGroup.Subnets)); err != nil {
		return diag.Errorf("error setting subnets: %s", err)
	}

	if err := d.Set("taint", flattenTaints(nodeGroup.Taints)); err != nil {
		return diag.Errorf("error setting taint: %s", err)
	}

	if nodeGroup.UpdateConfig != nil {
		if err := d.Set("update_config", []interface{}{flattenNodeGroupUpdateConfig(nodeGroup.UpdateConfig)}); err != nil {
			return diag.Errorf("error setting update_config: %s", err)
		}
	} else {
		d.Set("update_config", nil)
	}

	d.Set("version", nodeGroup.Version)

	tags := KeyValueTags(nodeGroup.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("error setting tags_all: %s", err)
	}

	return nil
}

func resourceNodeGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSConn

	clusterName, nodeGroupName, err := NodeGroupParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	// Do any version update first.
	if d.HasChanges("launch_template", "release_version", "version") {
		input := &eks.UpdateNodegroupVersionInput{
			ClientRequestToken: aws.String(resource.UniqueId()),
			ClusterName:        aws.String(clusterName),
			Force:              aws.Bool(d.Get("force_update_version").(bool)),
			NodegroupName:      aws.String(nodeGroupName),
		}

		if v := d.Get("launch_template").([]interface{}); len(v) > 0 {
			input.LaunchTemplate = expandLaunchTemplateSpecification(v)

			// When returning Launch Template information, the API returns all
			// fields. Since both the id and name are saved to the Terraform
			// state for drift detection and the API returns the following
			// error if both are present during update:
			// InvalidParameterException: Either provide launch template ID or launch template name in the request.

			// Remove the name if there are no changes, to prefer the ID.
			if input.LaunchTemplate.Id != nil && input.LaunchTemplate.Name != nil && !d.HasChange("launch_template.0.name") {
				input.LaunchTemplate.Name = nil
			}

			// Otherwise, remove the ID, but only if both are present still.
			if input.LaunchTemplate.Id != nil && input.LaunchTemplate.Name != nil && !d.HasChange("launch_template.0.id") {
				input.LaunchTemplate.Id = nil
			}
		}

		if v, ok := d.GetOk("release_version"); ok && d.HasChange("release_version") {
			input.ReleaseVersion = aws.String(v.(string))
		}

		if v, ok := d.GetOk("version"); ok && d.HasChange("version") {
			input.Version = aws.String(v.(string))
		}

		output, err := conn.UpdateNodegroupVersion(input)

		if err != nil {
			return diag.Errorf("error updating EKS Node Group (%s) version: %s", d.Id(), err)
		}

		updateID := aws.StringValue(output.Update.Id)

		_, err = waitNodegroupUpdateSuccessful(ctx, conn, clusterName, nodeGroupName, updateID, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return diag.Errorf("error waiting for EKS Node Group (%s) version update (%s): %s", d.Id(), updateID, err)
		}
	}

	if d.HasChanges("labels", "scaling_config", "taint", "update_config") {
		oldLabelsRaw, newLabelsRaw := d.GetChange("labels")
		oldTaintsRaw, newTaintsRaw := d.GetChange("taint")

		input := &eks.UpdateNodegroupConfigInput{
			ClientRequestToken: aws.String(resource.UniqueId()),
			ClusterName:        aws.String(clusterName),
			Labels:             expandUpdateLabelsPayload(oldLabelsRaw, newLabelsRaw),
			NodegroupName:      aws.String(nodeGroupName),
			Taints:             expandUpdateTaintsPayload(oldTaintsRaw.(*schema.Set).List(), newTaintsRaw.(*schema.Set).List()),
		}

		if d.HasChange("scaling_config") {
			if v, ok := d.GetOk("scaling_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.ScalingConfig = expandNodegroupScalingConfig(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if d.HasChange("update_config") {
			if v, ok := d.GetOk("update_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.UpdateConfig = expandNodegroupUpdateConfig(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		output, err := conn.UpdateNodegroupConfig(input)

		if err != nil {
			return diag.Errorf("error updating EKS Node Group (%s) config: %s", d.Id(), err)
		}

		updateID := aws.StringValue(output.Update.Id)

		_, err = waitNodegroupUpdateSuccessful(ctx, conn, clusterName, nodeGroupName, updateID, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return diag.Errorf("error waiting for EKS Node Group (%s) config update (%s): %s", d.Id(), updateID, err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("error updating tags: %s", err)
		}
	}

	return resourceNodeGroupRead(ctx, d, meta)
}

func resourceNodeGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSConn

	clusterName, nodeGroupName, err := NodeGroupParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting EKS Node Group: %s", d.Id())
	_, err = conn.DeleteNodegroup(&eks.DeleteNodegroupInput{
		ClusterName:   aws.String(clusterName),
		NodegroupName: aws.String(nodeGroupName),
	})

	if tfawserr.ErrCodeEquals(err, eks.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting EKS Node Group (%s): %s", d.Id(), err)
	}

	_, err = waitNodegroupDeleted(ctx, conn, clusterName, nodeGroupName, d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return diag.Errorf("error waiting for EKS Node Group (%s) to delete: %s", d.Id(), err)
	}

	return nil
}

func expandLaunchTemplateSpecification(l []interface{}) *eks.LaunchTemplateSpecification {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &eks.LaunchTemplateSpecification{}

	if v, ok := m["id"].(string); ok && v != "" {
		config.Id = aws.String(v)
	}

	if v, ok := m["name"].(string); ok && v != "" {
		config.Name = aws.String(v)
	}

	if v, ok := m["version"].(string); ok && v != "" {
		config.Version = aws.String(v)
	}

	return config
}

func expandNodegroupScalingConfig(tfMap map[string]interface{}) *eks.NodegroupScalingConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &eks.NodegroupScalingConfig{}

	if v, ok := tfMap["desired_size"].(int); ok {
		apiObject.DesiredSize = aws.Int64(int64(v))
	}

	if v, ok := tfMap["max_size"].(int); ok && v != 0 {
		apiObject.MaxSize = aws.Int64(int64(v))
	}

	if v, ok := tfMap["min_size"].(int); ok {
		apiObject.MinSize = aws.Int64(int64(v))
	}

	return apiObject
}

func expandTaints(l []interface{}) []*eks.Taint {
	if len(l) == 0 {
		return nil
	}

	var taints []*eks.Taint

	for _, raw := range l {
		t, ok := raw.(map[string]interface{})

		if !ok {
			continue
		}

		taint := &eks.Taint{}

		if k, ok := t["key"].(string); ok {
			taint.Key = aws.String(k)
		}

		if v, ok := t["value"].(string); ok {
			taint.Value = aws.String(v)
		}

		if e, ok := t["effect"].(string); ok {
			taint.Effect = aws.String(e)
		}

		taints = append(taints, taint)
	}

	return taints
}

func expandUpdateTaintsPayload(oldTaintsRaw, newTaintsRaw []interface{}) *eks.UpdateTaintsPayload {
	oldTaints := expandTaints(oldTaintsRaw)
	newTaints := expandTaints(newTaintsRaw)

	var removedTaints []*eks.Taint
	for _, ot := range oldTaints {
		if ot == nil {
			continue
		}

		removed := true
		for _, nt := range newTaints {
			if nt == nil {
				continue
			}

			// if both taint.key and taint.effect are the same, we don't need to remove it.
			if aws.StringValue(nt.Key) == aws.StringValue(ot.Key) &&
				aws.StringValue(nt.Effect) == aws.StringValue(ot.Effect) {
				removed = false
				break
			}
		}

		if removed {
			removedTaints = append(removedTaints, ot)
		}
	}

	var updatedTaints []*eks.Taint
	for _, nt := range newTaints {
		if nt == nil {
			continue
		}

		updated := true
		for _, ot := range oldTaints {
			if nt == nil {
				continue
			}

			if reflect.DeepEqual(nt, ot) {
				updated = false
				break
			}
		}
		if updated {
			updatedTaints = append(updatedTaints, nt)
		}
	}

	if len(removedTaints) == 0 && len(updatedTaints) == 0 {
		return nil
	}

	updateTaintsPayload := &eks.UpdateTaintsPayload{}

	if len(removedTaints) > 0 {
		updateTaintsPayload.RemoveTaints = removedTaints
	}

	if len(updatedTaints) > 0 {
		updateTaintsPayload.AddOrUpdateTaints = updatedTaints
	}

	return updateTaintsPayload
}

func expandRemoteAccessConfig(l []interface{}) *eks.RemoteAccessConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &eks.RemoteAccessConfig{}

	if v, ok := m["ec2_ssh_key"].(string); ok && v != "" {
		config.Ec2SshKey = aws.String(v)
	}

	if v, ok := m["source_security_group_ids"].(*schema.Set); ok && v.Len() > 0 {
		config.SourceSecurityGroups = flex.ExpandStringSet(v)
	}

	return config
}

func expandNodegroupUpdateConfig(tfMap map[string]interface{}) *eks.NodegroupUpdateConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &eks.NodegroupUpdateConfig{}

	if v, ok := tfMap["max_unavailable"].(int); ok && v != 0 {
		apiObject.MaxUnavailable = aws.Int64(int64(v))
	}

	if v, ok := tfMap["max_unavailable_percentage"].(int); ok && v != 0 {
		apiObject.MaxUnavailablePercentage = aws.Int64(int64(v))
	}

	return apiObject
}

func expandUpdateLabelsPayload(oldLabelsMap, newLabelsMap interface{}) *eks.UpdateLabelsPayload {
	// EKS Labels operate similarly to keyvaluetags
	oldLabels := tftags.New(oldLabelsMap)
	newLabels := tftags.New(newLabelsMap)

	removedLabels := oldLabels.Removed(newLabels)
	updatedLabels := oldLabels.Updated(newLabels)

	if len(removedLabels) == 0 && len(updatedLabels) == 0 {
		return nil
	}

	updateLabelsPayload := &eks.UpdateLabelsPayload{}

	if len(removedLabels) > 0 {
		updateLabelsPayload.RemoveLabels = aws.StringSlice(removedLabels.Keys())
	}

	if len(updatedLabels) > 0 {
		updateLabelsPayload.AddOrUpdateLabels = aws.StringMap(updatedLabels.Map())
	}

	return updateLabelsPayload
}

func flattenAutoScalingGroups(autoScalingGroups []*eks.AutoScalingGroup) []map[string]interface{} {
	if len(autoScalingGroups) == 0 {
		return []map[string]interface{}{}
	}

	l := make([]map[string]interface{}, 0, len(autoScalingGroups))

	for _, autoScalingGroup := range autoScalingGroups {
		m := map[string]interface{}{
			"name": aws.StringValue(autoScalingGroup.Name),
		}

		l = append(l, m)
	}

	return l
}

func flattenLaunchTemplateSpecification(config *eks.LaunchTemplateSpecification) []map[string]interface{} {
	if config == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := config.Id; v != nil {
		m["id"] = aws.StringValue(v)
	}

	if v := config.Name; v != nil {
		m["name"] = aws.StringValue(v)
	}

	if v := config.Version; v != nil {
		m["version"] = aws.StringValue(v)
	}

	return []map[string]interface{}{m}
}

func flattenNodeGroupResources(resources *eks.NodegroupResources) []map[string]interface{} {
	if resources == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"autoscaling_groups":              flattenAutoScalingGroups(resources.AutoScalingGroups),
		"remote_access_security_group_id": aws.StringValue(resources.RemoteAccessSecurityGroup),
	}

	return []map[string]interface{}{m}
}

func flattenNodeGroupScalingConfig(apiObject *eks.NodegroupScalingConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DesiredSize; v != nil {
		tfMap["desired_size"] = aws.Int64Value(v)
	}

	if v := apiObject.MaxSize; v != nil {
		tfMap["max_size"] = aws.Int64Value(v)
	}

	if v := apiObject.MinSize; v != nil {
		tfMap["min_size"] = aws.Int64Value(v)
	}

	return tfMap
}

func flattenNodeGroupUpdateConfig(apiObject *eks.NodegroupUpdateConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.MaxUnavailable; v != nil {
		tfMap["max_unavailable"] = aws.Int64Value(v)
	}

	if v := apiObject.MaxUnavailablePercentage; v != nil {
		tfMap["max_unavailable_percentage"] = aws.Int64Value(v)
	}

	return tfMap
}

func flattenRemoteAccessConfig(config *eks.RemoteAccessConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"ec2_ssh_key":               aws.StringValue(config.Ec2SshKey),
		"source_security_group_ids": aws.StringValueSlice(config.SourceSecurityGroups),
	}

	return []map[string]interface{}{m}
}

func flattenTaints(taints []*eks.Taint) []interface{} {
	if len(taints) == 0 {
		return nil
	}

	var results []interface{}

	for _, taint := range taints {
		if taint == nil {
			continue
		}

		t := make(map[string]interface{})
		t["key"] = aws.StringValue(taint.Key)
		t["value"] = aws.StringValue(taint.Value)
		t["effect"] = aws.StringValue(taint.Effect)

		results = append(results, t)
	}
	return results
}
