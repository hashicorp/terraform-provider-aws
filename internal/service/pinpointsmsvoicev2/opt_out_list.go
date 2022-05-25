package pinpointsmsvoicev2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/pinpointsmsvoicev2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceOptOutList() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOptOutListCreate,
		ReadWithoutTimeout:   resourceOptOutListRead,
		UpdateWithoutTimeout: resourceOptOutListUpdate,
		DeleteWithoutTimeout: resourceOptOutListDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceOptOutListCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).PinpointSMSVoiceV2Conn

	in := &pinpointsmsvoicev2.CreateOptOutListInput{
		OptOutListName: aws.String(d.Get("name").(string)),
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateOptOutListWithContext(ctx, in)
	if err != nil {
		return diag.Errorf("creating Amazon Pinpoint SMS and Voice V2 OptOutList (%s): %s", d.Get("name").(string), err)
	}

	if out == nil || out.OptOutListArn == nil {
		return diag.Errorf("creating Amazon Pinpoint SMS and Voice V2 OptOutList (%s): empty output", d.Get("name").(string))
	}

	d.SetId(aws.ToString(out.OptOutListName))

	return resourceOptOutListRead(ctx, d, meta)
}

func resourceOptOutListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).PinpointSMSVoiceV2Conn

	out, err := findOptOutListByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] PinpointSMSVoiceV2 OptOutList (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading PinpointSMSVoiceV2 OptOutList (%s): %s", d.Id(), err)
	}

	d.Set("arn", out.OptOutListArn)
	d.Set("name", out.OptOutListName)

	tags, err := ListTags(ctx, conn, d.Get("arn").(string))
	if err != nil {
		return diag.Errorf("listing tags for PinpointSMSVoiceV2 OptOutList (%s): %s", d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceOptOutListUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).PinpointSMSVoiceV2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("error updating tags: %s", err)
		}
	}

	return resourceOptOutListRead(ctx, d, meta)
}

func resourceOptOutListDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).PinpointSMSVoiceV2Conn

	log.Printf("[INFO] Deleting PinpointSMSVoiceV2 OptOutList %s", d.Id())

	_, err := conn.DeleteOptOutListWithContext(ctx, &pinpointsmsvoicev2.DeleteOptOutListInput{
		OptOutListName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, pinpointsmsvoicev2.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting PinpointSMSVoiceV2 OptOutList (%s): %s", d.Id(), err)
	}

	return nil
}
