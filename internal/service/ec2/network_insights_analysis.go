package ec2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceNetworkInsightsAnalysis() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkInsightsAnalysisCreate,
		ReadContext:   resourceNetworkInsightsAnalysisRead,
		UpdateContext: resourceNetworkInsightsAnalysisUpdate,
		DeleteContext: resourceNetworkInsightsAnalysisDelete,
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

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
		),
	}
}

func resourceNetworkInsightsAnalysisCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.StartNetworkInsightsAnalysisInput{
		NetworkInsightsPathId: aws.String(d.Get("network_insights_path_id").(string)),
		TagSpecifications:     ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeNetworkInsightsAnalysis),
	}

	if v := d.Get("filter_in_arns").(*schema.Set); v.Len() > 0 {
		for _, v := range v.List() {
			input.FilterInArns = append(input.FilterInArns, aws.String(v.(string)))
		}
	}

	response, err := conn.StartNetworkInsightsAnalysis(input)
	if err != nil {
		return diag.Errorf("error creating Network Insights Analysis: %s", err)
	}
	d.SetId(aws.StringValue(response.NetworkInsightsAnalysis.NetworkInsightsAnalysisId))

	if d.Get("wait_for_completion").(bool) {
		log.Printf("[DEBUG] Waiting until Network Insights Analysis (%s) is complete", d.Id())
		stateConf := &resource.StateChangeConf{
			Pending: []string{"running"},
			Target:  []string{"succeeded", "failed"},
			Refresh: func() (result interface{}, state string, err error) {
				nia, err := FindNetworkInsightsAnalysisByID(conn, d.Id())
				if err != nil {
					return nil, "", err
				}
				return nia, *nia.Status, nil
			},
			Timeout:    d.Timeout(schema.TimeoutCreate),
			Delay:      10 * time.Second,
			MinTimeout: 5 * time.Second,
		}

		_, err := stateConf.WaitForStateContext(ctx)
		if err != nil {
			return diag.Errorf("error waiting until Network Insights Analysis (%s) is complete: %s", d.Id(), err)
		}
	}

	return resourceNetworkInsightsAnalysisRead(ctx, d, meta)
}

func resourceNetworkInsightsAnalysisRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	nia, err := FindNetworkInsightsAnalysisByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Network Insights Analysis (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading EC2 Network Insights Analysis (%s): %s", d.Id(), err)
	}

	d.Set("arn", nia.NetworkInsightsAnalysisArn)
	d.Set("filter_in_arns", nia.FilterInArns)
	d.Set("network_insights_path_id", nia.NetworkInsightsPathId)
	d.Set("path_found", nia.NetworkPathFound)
	d.Set("start_date", nia.StartDate.Format(time.RFC3339))
	d.Set("status", nia.Status)
	d.Set("status_message", nia.StatusMessage)
	d.Set("warning_message", nia.WarningMessage)

	tags := KeyValueTags(nia.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("error setting tags_all: %s", err)
	}

	return nil
}

func resourceNetworkInsightsAnalysisUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return diag.Errorf("error updating Network Insights Analysis (%s) tags: %s", d.Id(), err)
		}
	}
	return resourceNetworkInsightsAnalysisRead(ctx, d, meta)
}

func resourceNetworkInsightsAnalysisDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn
	_, err := conn.DeleteNetworkInsightsAnalysisWithContext(ctx, &ec2.DeleteNetworkInsightsAnalysisInput{
		NetworkInsightsAnalysisId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidNetworkInsightsAnalysisIdNotFound) {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting Network Insights Analysis (%s): %s", d.Id(), err)
	}

	return nil
}
