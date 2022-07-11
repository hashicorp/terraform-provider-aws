package fms

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fms"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceProtocol() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProtocolListCreate,
		ReadContext:   resourceProtocolListRead,
		UpdateContext: resourceProtocolListUpdate,
		DeleteContext: resourceProtocolListDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"protocol_update_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"protocols": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 20,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceProtocolListCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).FMSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	name := d.Get("name").(string)
	input := &fms.PutProtocolsListInput{
		ProtocolsList: resourceProtocolExpand(d),
	}

	if len(tags) > 0 {
		input.TagList = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating FMS Protocol List %s", name)

	output, err := conn.PutProtocolsListWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating FMS Protocol List (%s): %w", name, err))
	}

	if output == nil || output.ProtocolsList == nil {
		return diag.FromErr(fmt.Errorf("error reading FMS Protocol List (%s): empty output", name))
	}

	d.SetId(aws.StringValue(output.ProtocolsList.ListId))

	return resourceProtocolListRead(ctx, d, meta)
}

func resourceProtocolListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).FMSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindProtocolListByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FMS Protocol List %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading FMS Protocol List (%s): %w", d.Id(), err))
	}

	if output == nil || output.ProtocolsList == nil {
		return diag.FromErr(fmt.Errorf("error reading FMS Protocol List (%s): empty output", d.Id()))
	}

	d.Set("arn", output.ProtocolsListArn)
	d.Set("name", output.ProtocolsList.ListName)
	d.Set("protocols", aws.StringValueSlice(output.ProtocolsList.ProtocolsList))
	d.Set("protocol_update_token", output.ProtocolsList.ListUpdateToken)

	tags, err := ListTags(conn, d.Get("arn").(string))

	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing tags for FMS Protocol Lists (%s): %w", d.Id(), err))
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}

	return nil
}

func resourceProtocolListUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).FMSConn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &fms.PutProtocolsListInput{
			ProtocolsList: resourceProtocolExpand(d),
		}

		_, err := conn.PutProtocolsList(input)

		if err != nil {
			return diag.FromErr(fmt.Errorf("error updating FMS Protocol Lists (%s): %w", d.Id(), err))
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating FMS Protocol Lists (%s) tags: %w", d.Id(), err))
		}
	}

	return resourceProtocolListRead(ctx, d, meta)
}

func resourceProtocolListDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).FMSConn

	log.Printf("[DEBUG] Deleting FMS Protocol List: %s", d.Id())
	_, err := conn.DeleteProtocolsList(&fms.DeleteProtocolsListInput{
		ListId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, fms.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting FMS Policy (%s): %w", d.Id(), err))
	}

	return nil
}

func FindProtocolListByID(ctx context.Context, conn *fms.FMS, id string) (*fms.GetProtocolsListOutput, error) {
	input := &fms.GetProtocolsListInput{
		ListId: aws.String(id),
	}

	output, err := conn.GetProtocolsListWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, fms.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func resourceProtocolExpand(d *schema.ResourceData) *fms.ProtocolsListData {
	fmsProtocol := &fms.ProtocolsListData{
		ListName:      aws.String(d.Get("name").(string)),
		ProtocolsList: flex.ExpandStringList(d.Get("protocols").([]interface{})),
	}

	if d.Id() != "" {
		fmsProtocol.ListId = aws.String(d.Id())
		fmsProtocol.ListUpdateToken = aws.String(d.Get("protocol_update_token").(string))
	}
	return fmsProtocol
}
