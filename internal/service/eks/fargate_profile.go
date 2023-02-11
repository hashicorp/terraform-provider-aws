package eks

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceFargateProfile() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFargateProfileCreate,
		ReadWithoutTimeout:   resourceFargateProfileRead,
		UpdateWithoutTimeout: resourceFargateProfileUpdate,
		DeleteWithoutTimeout: resourceFargateProfileDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourceFargateProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSConn()
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

	err := resource.RetryContext(ctx, propagationTimeout, func() *resource.RetryError {
		_, err := conn.CreateFargateProfileWithContext(ctx, input)

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
		_, err = conn.CreateFargateProfileWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EKS Fargate Profile (%s): %s", id, err)
	}

	d.SetId(id)

	_, err = waitFargateProfileCreated(ctx, conn, clusterName, fargateProfileName, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EKS Fargate Profile (%s) to create: %s", d.Id(), err)
	}

	return append(diags, resourceFargateProfileRead(ctx, d, meta)...)
}

func resourceFargateProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	clusterName, fargateProfileName, err := FargateProfileParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EKS Fargate Profile (%s): %s", d.Id(), err)
	}

	fargateProfile, err := FindFargateProfileByClusterNameAndFargateProfileName(ctx, conn, clusterName, fargateProfileName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EKS Fargate Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EKS Fargate Profile (%s): %s", d.Id(), err)
	}

	d.Set("arn", fargateProfile.FargateProfileArn)
	d.Set("cluster_name", fargateProfile.ClusterName)
	d.Set("fargate_profile_name", fargateProfile.FargateProfileName)
	d.Set("pod_execution_role_arn", fargateProfile.PodExecutionRoleArn)

	if err := d.Set("selector", flattenFargateProfileSelectors(fargateProfile.Selectors)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting selector: %s", err)
	}

	d.Set("status", fargateProfile.Status)

	if err := d.Set("subnet_ids", aws.StringValueSlice(fargateProfile.Subnets)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting subnet_ids: %s", err)
	}

	tags := KeyValueTags(fargateProfile.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceFargateProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags: %s", err)
		}
	}

	return append(diags, resourceFargateProfileRead(ctx, d, meta)...)
}

func resourceFargateProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSConn()

	clusterName, fargateProfileName, err := FargateProfileParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EKS Fargate Profile (%s): %s", d.Id(), err)
	}

	// mutex lock for creation/deletion serialization
	mutexKey := fmt.Sprintf("%s-fargate-profiles", d.Get("cluster_name").(string))
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	log.Printf("[DEBUG] Deleting EKS Fargate Profile: %s", d.Id())
	_, err = conn.DeleteFargateProfileWithContext(ctx, &eks.DeleteFargateProfileInput{
		ClusterName:        aws.String(clusterName),
		FargateProfileName: aws.String(fargateProfileName),
	})

	if tfawserr.ErrCodeEquals(err, eks.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EKS Fargate Profile (%s): %s", d.Id(), err)
	}

	_, err = waitFargateProfileDeleted(ctx, conn, clusterName, fargateProfileName, d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EKS Fargate Profile (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
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
