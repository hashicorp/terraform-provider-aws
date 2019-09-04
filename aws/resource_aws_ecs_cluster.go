package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsEcsCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEcsClusterCreate,
		Read:   resourceAwsEcsClusterRead,
		Update: resourceAwsEcsClusterUpdate,
		Delete: resourceAwsEcsClusterDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsEcsClusterImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags": tagsSchema(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"setting": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								ecs.ClusterSettingNameContainerInsights,
							}, false),
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsEcsClusterImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set("name", d.Id())
	d.SetId(arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Service:   "ecs",
		Resource:  fmt.Sprintf("cluster/%s", d.Id()),
	}.String())
	return []*schema.ResourceData{d}, nil
}

func resourceAwsEcsClusterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	clusterName := d.Get("name").(string)
	log.Printf("[DEBUG] Creating ECS cluster %s", clusterName)

	input := ecs.CreateClusterInput{
		ClusterName: aws.String(clusterName),
		Tags:        tagsFromMapECS(d.Get("tags").(map[string]interface{})),
	}

	if v, ok := d.GetOk("setting"); ok {
		input.Settings = expandEcsSettings(v.(*schema.Set).List())
	}

	out, err := conn.CreateCluster(&input)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] ECS cluster %s created", *out.Cluster.ClusterArn)

	d.SetId(*out.Cluster.ClusterArn)

	return resourceAwsEcsClusterRead(d, meta)
}

func resourceAwsEcsClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	input := &ecs.DescribeClustersInput{
		Clusters: []*string{aws.String(d.Id())},
		Include:  []*string{aws.String(ecs.ClusterFieldTags)},
	}

	log.Printf("[DEBUG] Reading ECS Cluster: %s", input)
	var out *ecs.DescribeClustersOutput
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		var err error
		out, err = conn.DescribeClusters(input)

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if out == nil || len(out.Failures) > 0 {
			if d.IsNewResource() {
				return resource.RetryableError(&resource.NotFoundError{})
			}
			return resource.NonRetryableError(&resource.NotFoundError{})
		}

		return nil
	})
	if isResourceTimeoutError(err) {
		out, err = conn.DescribeClusters(input)
	}

	if isResourceNotFoundError(err) {
		log.Printf("[WARN] ECS Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading ECS Cluster (%s): %s", d.Id(), err)
	}

	var cluster *ecs.Cluster
	for _, c := range out.Clusters {
		if aws.StringValue(c.ClusterArn) == d.Id() {
			cluster = c
			break
		}
	}

	if cluster == nil {
		log.Printf("[WARN] ECS Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	// Status==INACTIVE means deleted cluster
	if aws.StringValue(cluster.Status) == "INACTIVE" {
		log.Printf("[WARN] ECS Cluster (%s) deleted, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", cluster.ClusterArn)
	d.Set("name", cluster.ClusterName)

	if err := d.Set("setting", flattenEcsSettings(cluster.Settings)); err != nil {
		return fmt.Errorf("error setting setting: %s", err)
	}

	if err := d.Set("tags", tagsToMapECS(cluster.Tags)); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsEcsClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	if d.HasChange("setting") {
		input := ecs.UpdateClusterSettingsInput{
			Cluster:  aws.String(d.Id()),
			Settings: expandEcsSettings(d.Get("setting").(*schema.Set).List()),
		}

		_, err := conn.UpdateClusterSettings(&input)
		if err != nil {
			return fmt.Errorf("error changing ECS cluster settings (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		oldTagsRaw, newTagsRaw := d.GetChange("tags")
		oldTagsMap := oldTagsRaw.(map[string]interface{})
		newTagsMap := newTagsRaw.(map[string]interface{})
		createTags, removeTags := diffTagsECS(tagsFromMapECS(oldTagsMap), tagsFromMapECS(newTagsMap))

		if len(removeTags) > 0 {
			removeTagKeys := make([]*string, len(removeTags))
			for i, removeTag := range removeTags {
				removeTagKeys[i] = removeTag.Key
			}

			input := &ecs.UntagResourceInput{
				ResourceArn: aws.String(d.Id()),
				TagKeys:     removeTagKeys,
			}

			log.Printf("[DEBUG] Untagging ECS Cluster: %s", input)
			if _, err := conn.UntagResource(input); err != nil {
				return fmt.Errorf("error untagging ECS Cluster (%s): %s", d.Id(), err)
			}
		}

		if len(createTags) > 0 {
			input := &ecs.TagResourceInput{
				ResourceArn: aws.String(d.Id()),
				Tags:        createTags,
			}

			log.Printf("[DEBUG] Tagging ECS Cluster: %s", input)
			if _, err := conn.TagResource(input); err != nil {
				return fmt.Errorf("error tagging ECS Cluster (%s): %s", d.Id(), err)
			}
		}
	}

	return nil
}

func resourceAwsEcsClusterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	log.Printf("[DEBUG] Deleting ECS cluster %s", d.Id())
	input := &ecs.DeleteClusterInput{
		Cluster: aws.String(d.Id()),
	}
	err := resource.Retry(10*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteCluster(input)

		if err == nil {
			log.Printf("[DEBUG] ECS cluster %s deleted", d.Id())
			return nil
		}

		if isAWSErr(err, "ClusterContainsContainerInstancesException", "") {
			log.Printf("[TRACE] Retrying ECS cluster %q deletion after %s", d.Id(), err)
			return resource.RetryableError(err)
		}
		if isAWSErr(err, "ClusterContainsServicesException", "") {
			log.Printf("[TRACE] Retrying ECS cluster %q deletion after %s", d.Id(), err)
			return resource.RetryableError(err)
		}
		return resource.NonRetryableError(err)
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DeleteCluster(input)
	}
	if err != nil {
		return fmt.Errorf("Error deleting ECS cluster: %s", err)
	}

	clusterName := d.Get("name").(string)
	dcInput := &ecs.DescribeClustersInput{
		Clusters: []*string{aws.String(clusterName)},
	}
	var out *ecs.DescribeClustersOutput
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		log.Printf("[DEBUG] Checking if ECS Cluster %q is INACTIVE", d.Id())
		out, err = conn.DescribeClusters(dcInput)

		if err != nil {
			return resource.NonRetryableError(err)
		}
		if !ecsClusterInactive(out, clusterName) {
			return resource.RetryableError(fmt.Errorf("ECS Cluster %q is not inactive", clusterName))
		}

		return nil
	})
	if isResourceTimeoutError(err) {
		out, err = conn.DescribeClusters(dcInput)
		if err != nil {
			return fmt.Errorf("Error waiting for ECS cluster to become inactive: %s", err)
		}
		if !ecsClusterInactive(out, clusterName) {
			return fmt.Errorf("ECS Cluster %q is still not inactive", clusterName)
		}
	}
	if err != nil {
		return fmt.Errorf("Error waiting for ECS cluster to become inactive: %s", err)
	}

	log.Printf("[DEBUG] ECS cluster %q deleted", d.Id())
	return nil
}

func ecsClusterInactive(out *ecs.DescribeClustersOutput, clusterName string) bool {
	for _, c := range out.Clusters {
		if aws.StringValue(c.ClusterName) == clusterName {
			if *c.Status == "INACTIVE" {
				return true
			}
		}
	}
	return false
}

func expandEcsSettings(configured []interface{}) []*ecs.ClusterSetting {
	if len(configured) == 0 {
		return nil
	}

	settings := make([]*ecs.ClusterSetting, 0, len(configured))

	for _, raw := range configured {
		data := raw.(map[string]interface{})

		setting := &ecs.ClusterSetting{
			Name:  aws.String(data["name"].(string)),
			Value: aws.String(data["value"].(string)),
		}

		settings = append(settings, setting)
	}

	return settings
}

func flattenEcsSettings(list []*ecs.ClusterSetting) []map[string]interface{} {
	if len(list) == 0 {
		return nil
	}

	result := make([]map[string]interface{}, 0, len(list))
	for _, setting := range list {
		l := map[string]interface{}{
			"name":  aws.StringValue(setting.Name),
			"value": aws.StringValue(setting.Value),
		}

		result = append(result, l)
	}
	return result
}
