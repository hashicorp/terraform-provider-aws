package appstream

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceImageBuilder() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceImageBuilderCreate,
		ReadWithoutTimeout:   resourceImageBuilderRead,
		UpdateWithoutTimeout: resourceImageBuilderUpdate,
		DeleteWithoutTimeout: resourceImageBuilderDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"access_endpoint": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 4,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoint_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(appstream.AccessEndpointType_Values(), false),
						},
						"vpce_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"appstream_agent_version": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"display_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 100),
			},
			"domain_join_info": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"directory_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"organizational_unit_distinguished_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"enable_default_internet_access": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"iam_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"image_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"image_arn", "image_name"},
				ValidateFunc: verify.ValidARN,
			},
			"image_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"image_name", "image_arn"},
			},
			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_group_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceImageBuilderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppStreamConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)

	input := &appstream.CreateImageBuilderInput{
		Name:         aws.String(name),
		InstanceType: aws.String(d.Get("instance_type").(string)),
	}

	if v, ok := d.GetOk("access_endpoint"); ok && v.(*schema.Set).Len() > 0 {
		input.AccessEndpoints = expandAccessEndpoints(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("appstream_agent_version"); ok {
		input.AppstreamAgentVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("display_name"); ok {
		input.DisplayName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("domain_join_info"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DomainJoinInfo = expandDomainJoinInfo(v.([]interface{}))
	}

	if v, ok := d.GetOk("enable_default_internet_access"); ok {
		input.EnableDefaultInternetAccess = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("image_arn"); ok {
		input.ImageArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("image_name"); ok {
		input.ImageName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("iam_role_arn"); ok {
		input.IamRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vpc_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.VpcConfig = expandImageBuilderVPCConfig(v.([]interface{}))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	output, err := conn.CreateImageBuilderWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Appstream ImageBuilder (%s): %w", name, err))
	}

	d.SetId(aws.StringValue(output.ImageBuilder.Name))

	if _, err = waitImageBuilderStateRunning(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for Appstream ImageBuilder (%s) to be running: %w", d.Id(), err))
	}

	return resourceImageBuilderRead(ctx, d, meta)
}

func resourceImageBuilderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppStreamConn

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	imageBuilder, err := FindImageBuilderByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Appstream ImageBuilder (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Appstream ImageBuilder (%s): %w", d.Id(), err))
	}

	if imageBuilder == nil {
		return diag.FromErr(fmt.Errorf("error reading Appstream ImageBuilder (%s): not found after creation", d.Id()))
	}

	arn := aws.StringValue(imageBuilder.Arn)

	d.Set("appstream_agent_version", imageBuilder.AppstreamAgentVersion)
	d.Set("arn", arn)
	d.Set("created_time", aws.TimeValue(imageBuilder.CreatedTime).Format(time.RFC3339))
	d.Set("description", imageBuilder.Description)
	d.Set("display_name", imageBuilder.DisplayName)
	d.Set("enable_default_internet_access", imageBuilder.EnableDefaultInternetAccess)
	d.Set("image_arn", imageBuilder.ImageArn)
	d.Set("iam_role_arn", imageBuilder.IamRoleArn)
	d.Set("instance_type", imageBuilder.InstanceType)

	if err = d.Set("access_endpoint", flattenAccessEndpoints(imageBuilder.AccessEndpoints)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for AppStream ImageBuilder (%s): %w", "access_endpoints", d.Id(), err))
	}
	if err = d.Set("domain_join_info", flattenDomainInfo(imageBuilder.DomainJoinInfo)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for AppStream ImageBuilder (%s): %w", "domain_join_info", d.Id(), err))
	}

	if err = d.Set("vpc_config", flattenVPCConfig(imageBuilder.VpcConfig)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for AppStream ImageBuilder (%s): %w", "vpc_config", d.Id(), err))
	}

	d.Set("name", imageBuilder.Name)
	d.Set("state", imageBuilder.State)

	tags, err := ListTags(conn, arn)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing tags for AppStream ImageBuilder (%s): %w", arn, err))
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}

	return nil
}

func resourceImageBuilderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.HasChange("tags_all") {
		conn := meta.(*conns.AWSClient).AppStreamConn

		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating tags for AppStream ImageBuilder (%s): %w", d.Id(), err))
		}
	}

	return resourceImageBuilderRead(ctx, d, meta)
}

func resourceImageBuilderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppStreamConn

	log.Printf("[DEBUG] Deleting AppStream Image Builder: (%s)", d.Id())
	_, err := conn.DeleteImageBuilderWithContext(ctx, &appstream.DeleteImageBuilderInput{
		Name: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting Appstream ImageBuilder (%s): %w", d.Id(), err))
	}

	if _, err = waitImageBuilderStateDeleted(ctx, conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error waiting for Appstream ImageBuilder (%s) to be deleted: %w", d.Id(), err))
	}

	return nil
}

func expandImageBuilderVPCConfig(tfList []interface{}) *appstream.VpcConfig {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	apiObject := &appstream.VpcConfig{}

	if v, ok := tfMap["security_group_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroupIds = flex.ExpandStringSet(v)
	}
	if v, ok := tfMap["subnet_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SubnetIds = flex.ExpandStringSet(v)
	}

	return apiObject
}
