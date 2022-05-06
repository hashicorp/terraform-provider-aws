package eks

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceFargateProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceFargateProfileCreate,
		Read:   resourceFargateProfileRead,
		Update: resourceFargateProfileUpdate,
		Delete: resourceFargateProfileDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validClusterName,
			},
			"fargate_profile_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"pod_execution_role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"selector": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"labels": {
							Type:     schema.TypeMap,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"namespace": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.NoZeroValues,
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
				Optional: true,
				ForceNew: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceFargateProfileCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EKSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	clusterName := d.Get("cluster_name").(string)
	fargateProfileName := d.Get("fargate_profile_name").(string)
	id := FargateProfileCreateResourceID(clusterName, fargateProfileName)

	input := &eks.CreateFargateProfileInput{
		ClientRequestToken:  aws.String(resource.UniqueId()),
		ClusterName:         aws.String(clusterName),
		FargateProfileName:  aws.String(fargateProfileName),
		PodExecutionRoleArn: aws.String(d.Get("pod_execution_role_arn").(string)),
		Selectors:           expandFargateProfileSelectors(d.Get("selector").(*schema.Set).List()),
		Subnets:             flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	// mutex lock for creation/deletion serialization
	mutexKey := fmt.Sprintf("%s-fargate-profiles", clusterName)
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		_, err := conn.CreateFargateProfile(input)

		// Retry for IAM eventual consistency on error:
		// InvalidParameterException: Misconfigured PodExecutionRole Trust Policy; Please add the eks-fargate-pods.amazonaws.com Service Principal
		if tfawserr.ErrMessageContains(err, eks.ErrCodeInvalidParameterException, "Misconfigured PodExecutionRole Trust Policy") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.CreateFargateProfile(input)
	}

	if err != nil {
		return fmt.Errorf("error creating EKS Fargate Profile (%s): %w", id, err)
	}

	d.SetId(id)

	_, err = waitFargateProfileCreated(conn, clusterName, fargateProfileName, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return fmt.Errorf("error waiting for EKS Fargate Profile (%s) to create: %w", d.Id(), err)
	}

	return resourceFargateProfileRead(d, meta)
}

func resourceFargateProfileRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EKSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	clusterName, fargateProfileName, err := FargateProfileParseResourceID(d.Id())

	if err != nil {
		return err
	}

	fargateProfile, err := FindFargateProfileByClusterNameAndFargateProfileName(conn, clusterName, fargateProfileName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EKS Fargate Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EKS Fargate Profile (%s): %w", d.Id(), err)
	}

	d.Set("arn", fargateProfile.FargateProfileArn)
	d.Set("cluster_name", fargateProfile.ClusterName)
	d.Set("fargate_profile_name", fargateProfile.FargateProfileName)
	d.Set("pod_execution_role_arn", fargateProfile.PodExecutionRoleArn)

	if err := d.Set("selector", flattenFargateProfileSelectors(fargateProfile.Selectors)); err != nil {
		return fmt.Errorf("error setting selector: %w", err)
	}

	d.Set("status", fargateProfile.Status)

	if err := d.Set("subnet_ids", aws.StringValueSlice(fargateProfile.Subnets)); err != nil {
		return fmt.Errorf("error setting subnet_ids: %w", err)
	}

	tags := KeyValueTags(fargateProfile.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceFargateProfileUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EKSConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceFargateProfileRead(d, meta)
}

func resourceFargateProfileDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EKSConn

	clusterName, fargateProfileName, err := FargateProfileParseResourceID(d.Id())

	if err != nil {
		return err
	}

	// mutex lock for creation/deletion serialization
	mutexKey := fmt.Sprintf("%s-fargate-profiles", d.Get("cluster_name").(string))
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	log.Printf("[DEBUG] Deleting EKS Fargate Profile: %s", d.Id())
	_, err = conn.DeleteFargateProfile(&eks.DeleteFargateProfileInput{
		ClusterName:        aws.String(clusterName),
		FargateProfileName: aws.String(fargateProfileName),
	})

	if tfawserr.ErrCodeEquals(err, eks.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EKS Fargate Profile (%s): %w", d.Id(), err)
	}

	_, err = waitFargateProfileDeleted(conn, clusterName, fargateProfileName, d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return fmt.Errorf("error waiting for EKS Fargate Profile (%s) to delete: %w", d.Id(), err)
	}

	return nil
}

func expandFargateProfileSelectors(l []interface{}) []*eks.FargateProfileSelector {
	if len(l) == 0 {
		return nil
	}

	fargateProfileSelectors := make([]*eks.FargateProfileSelector, 0, len(l))

	for _, mRaw := range l {
		m, ok := mRaw.(map[string]interface{})

		if !ok {
			continue
		}

		fargateProfileSelector := &eks.FargateProfileSelector{}

		if v, ok := m["labels"].(map[string]interface{}); ok && len(v) > 0 {
			fargateProfileSelector.Labels = flex.ExpandStringMap(v)
		}

		if v, ok := m["namespace"].(string); ok && v != "" {
			fargateProfileSelector.Namespace = aws.String(v)
		}

		fargateProfileSelectors = append(fargateProfileSelectors, fargateProfileSelector)
	}

	return fargateProfileSelectors
}

func flattenFargateProfileSelectors(fargateProfileSelectors []*eks.FargateProfileSelector) []map[string]interface{} {
	if len(fargateProfileSelectors) == 0 {
		return []map[string]interface{}{}
	}

	l := make([]map[string]interface{}, 0, len(fargateProfileSelectors))

	for _, fargateProfileSelector := range fargateProfileSelectors {
		m := map[string]interface{}{
			"labels":    aws.StringValueMap(fargateProfileSelector.Labels),
			"namespace": aws.StringValue(fargateProfileSelector.Namespace),
		}

		l = append(l, m)
	}

	return l
}
