package ec2

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceNetworkInsightsPath() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkInsightsPathCreate,
		ReadContext:   resourceNetworkInsightsPathRead,
		UpdateContext: resourceNetworkInsightsPathUpdate,
		DeleteContext: resourceNetworkInsightsPathDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"source": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source_ip": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"destination_ip": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"destination_port": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"tcp",
					"udp",
				}, false),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
		),
	}
}

func resourceNetworkInsightsPathCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateNetworkInsightsPathInput{
		Source:            aws.String(d.Get("source").(string)),
		Destination:       aws.String(d.Get("destination").(string)),
		Protocol:          aws.String(d.Get("protocol").(string)),
		TagSpecifications: ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeNetworkInsightsPath),
	}

	if v, ok := d.GetOk("source_ip"); ok {
		input.SourceIp = aws.String(v.(string))
	}

	if v, ok := d.GetOk("destination_ip"); ok {
		input.DestinationIp = aws.String(v.(string))
	}

	if v, ok := d.GetOk("destination_port"); ok {
		input.DestinationPort = aws.Int64(int64(v.(int)))
	}

	response, err := conn.CreateNetworkInsightsPath(input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Network Insights Path: %w", err))
	}

	d.SetId(aws.StringValue(response.NetworkInsightsPath.NetworkInsightsPathId))
	return resourceNetworkInsightsPathRead(ctx, d, meta)
}

func resourceNetworkInsightsPathRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	nip, err := FindNetworkInsightsPathByID(conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ErrCodeInvalidNetworkInsightsPathIDNotFound) {
		log.Printf("[WARN] Network Insights Path (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting Network Insights Path (%s): %w", d.Id(), err))
	}

	if nip == nil {
		return diag.FromErr(fmt.Errorf("error getting Network Insights Path (%s): empty output", d.Id()))
	}

	d.Set("source", nip.Source)
	d.Set("destination", nip.Destination)
	d.Set("protocol", nip.Protocol)
	d.Set("arn", nip.NetworkInsightsPathArn)
	d.Set("source_ip", nip.SourceIp)
	d.Set("destination_ip", nip.DestinationIp)
	d.Set("destination_port", nip.DestinationPort)

	tags := KeyValueTags(nip.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}
	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}

	return nil
}

func resourceNetworkInsightsPathUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn
	if d.HasChange("tags_all") && !d.IsNewResource() {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating Network Insights Path (%s) tags: %w", d.Id(), err))
		}
	}
	return resourceNetworkInsightsPathRead(ctx, d, meta)
}

func resourceNetworkInsightsPathDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn
	_, err := conn.DeleteNetworkInsightsPath(&ec2.DeleteNetworkInsightsPathInput{
		NetworkInsightsPathId: aws.String(d.Id()),
	})

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting Network Insights Path (%s): %w", d.Id(), err))
	}

	return nil
}
