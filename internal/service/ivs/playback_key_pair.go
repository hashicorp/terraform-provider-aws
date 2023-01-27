package ivs

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ivs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourcePlaybackKeyPair() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePlaybackKeyPairCreate,
		ReadWithoutTimeout:   resourcePlaybackKeyPairRead,
		DeleteWithoutTimeout: resourcePlaybackKeyPairDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"public_key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchemaForceNew(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNamePlaybackKeyPair = "Playback Key Pair"
)

func resourcePlaybackKeyPairCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IVSConn()

	in := &ivs.ImportPlaybackKeyPairInput{
		PublicKeyMaterial: aws.String(d.Get("public_key").(string)),
	}

	if v, ok := d.GetOk("name"); ok {
		in.Name = aws.String(v.(string))
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.ImportPlaybackKeyPairWithContext(ctx, in)
	if err != nil {
		return create.DiagError(names.IVS, create.ErrActionCreating, ResNamePlaybackKeyPair, d.Get("name").(string), err)
	}

	if out == nil || out.KeyPair == nil {
		return create.DiagError(names.IVS, create.ErrActionCreating, ResNamePlaybackKeyPair, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.StringValue(out.KeyPair.Arn))

	if _, err := waitPlaybackKeyPairCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.IVS, create.ErrActionWaitingForCreation, ResNamePlaybackKeyPair, d.Id(), err)
	}

	return resourcePlaybackKeyPairRead(ctx, d, meta)
}

func resourcePlaybackKeyPairRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IVSConn()

	out, err := FindPlaybackKeyPairByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IVS PlaybackKeyPair (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.IVS, create.ErrActionReading, ResNamePlaybackKeyPair, d.Id(), err)
	}

	d.Set("arn", out.Arn)
	d.Set("name", out.Name)
	d.Set("fingerprint", out.Fingerprint)

	tags, err := ListTags(ctx, conn, d.Id())
	if err != nil {
		return create.DiagError(names.IVS, create.ErrActionReading, ResNamePlaybackKeyPair, d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.IVS, create.ErrActionSetting, ResNamePlaybackKeyPair, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.IVS, create.ErrActionSetting, ResNamePlaybackKeyPair, d.Id(), err)
	}

	return nil
}

func resourcePlaybackKeyPairDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IVSConn()

	log.Printf("[INFO] Deleting IVS PlaybackKeyPair %s", d.Id())

	_, err := conn.DeletePlaybackKeyPairWithContext(ctx, &ivs.DeletePlaybackKeyPairInput{
		Arn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, ivs.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return create.DiagError(names.IVS, create.ErrActionDeleting, ResNamePlaybackKeyPair, d.Id(), err)
	}

	if _, err := waitPlaybackKeyPairDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.IVS, create.ErrActionWaitingForDeletion, ResNamePlaybackKeyPair, d.Id(), err)
	}

	return nil
}
