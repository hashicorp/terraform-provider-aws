package ec2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceNetworkInsightsAnalysis() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNetworkInsightsAnalysisCreate,
		ReadWithoutTimeout:   resourceNetworkInsightsAnalysisRead,
		UpdateWithoutTimeout: resourceNetworkInsightsAnalysisUpdate,
		DeleteWithoutTimeout: resourceNetworkInsightsAnalysisDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"filter_in_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"network_insights_path_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"path_found": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"start_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status_message": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"wait_for_completion": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"warning_message": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceNetworkInsightsAnalysisCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.StartNetworkInsightsAnalysisInput{
		NetworkInsightsPathId: aws.String(d.Get("network_insights_path_id").(string)),
		TagSpecifications:     tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeNetworkInsightsAnalysis),
	}

	if v, ok := d.GetOk("filter_in_arns"); ok && v.(*schema.Set).Len() > 0 {
		input.FilterInArns = flex.ExpandStringSet(v.(*schema.Set))
	}

	log.Printf("[DEBUG] Creating EC2 Network Insights Analysis: %s", input)
	response, err := conn.StartNetworkInsightsAnalysisWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error creating EC2 Network Insights Analysis: %s", err)
	}

	d.SetId(aws.StringValue(response.NetworkInsightsAnalysis.NetworkInsightsAnalysisId))

	if d.Get("wait_for_completion").(bool) {
		if _, err := WaitNetworkInsightsAnalysisCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return diag.Errorf("error waiting for EC2 Network Insights Analysis (%s) create: %s", d.Id(), err)
		}
	}

	return resourceNetworkInsightsAnalysisRead(ctx, d, meta)
}

func resourceNetworkInsightsAnalysisRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	nia, err := FindNetworkInsightsAnalysisByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Network Insights Analysis (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading EC2 Network Insights Analysis (%s): %s", d.Id(), err)
	}

	d.Set("arn", nia.NetworkInsightsAnalysisArn)
	d.Set("filter_in_arns", aws.StringValueSlice(nia.FilterInArns))
	d.Set("network_insights_path_id", nia.NetworkInsightsPathId)
	d.Set("path_found", nia.NetworkPathFound)
	d.Set("start_date", nia.StartDate.Format(time.RFC3339))
	d.Set("status", nia.Status)
	d.Set("status_message", nia.StatusMessage)
	d.Set("warning_message", nia.WarningMessage)

	tags := KeyValueTags(nia.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceNetworkInsightsAnalysisUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTagsWithContext(ctx, conn, d.Id(), o, n); err != nil {
			return diag.Errorf("updating EC2 Network Insights Analysis (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceNetworkInsightsAnalysisRead(ctx, d, meta)
}

func resourceNetworkInsightsAnalysisDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[DEBUG] Deleting EC2 Network Insights Analysis: %s", d.Id())
	_, err := conn.DeleteNetworkInsightsAnalysisWithContext(ctx, &ec2.DeleteNetworkInsightsAnalysisInput{
		NetworkInsightsAnalysisId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidNetworkInsightsAnalysisIdNotFound) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting EC2 Network Insights Analysis (%s): %s", d.Id(), err)
	}

	return nil
}
