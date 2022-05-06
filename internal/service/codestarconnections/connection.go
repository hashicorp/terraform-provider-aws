package codestarconnections

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codestarconnections"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceConnection() *schema.Resource {
	return &schema.Resource{
		Create: resourceConnectionCreate,
		Read:   resourceConnectionRead,
		Update: resourceConnectionUpdate,
		Delete: resourceConnectionDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"host_arn": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ValidateFunc:  verify.ValidARN,
				ConflictsWith: []string{"provider_type"},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"provider_type": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Computed:      true,
				ValidateFunc:  validation.StringInSlice(codestarconnections.ProviderType_Values(), false),
				ConflictsWith: []string{"host_arn"},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeStarConnectionsConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &codestarconnections.CreateConnectionInput{
		ConnectionName: aws.String(name),
	}

	if v, ok := d.GetOk("host_arn"); ok {
		input.HostArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("provider_type"); ok {
		input.ProviderType = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating CodeStar Connections Connection: %s", input)
	output, err := conn.CreateConnection(input)

	if err != nil {
		return fmt.Errorf("creating CodeStar Connections Connection (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.ConnectionArn))

	return resourceConnectionRead(d, meta)
}

func resourceConnectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeStarConnectionsConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	connection, err := FindConnectionByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		names.LogNotFoundRemoveState(names.CodeStarConnections, names.ErrActionReading, ResConnection, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.Error(names.CodeStarConnections, names.ErrActionReading, ResConnection, d.Id(), err)
	}

	arn := aws.StringValue(connection.ConnectionArn)
	d.SetId(arn)
	d.Set("arn", connection.ConnectionArn)
	d.Set("connection_status", connection.ConnectionStatus)
	d.Set("host_arn", connection.HostArn)
	d.Set("name", connection.ConnectionName)
	d.Set("provider_type", connection.ProviderType)

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("listing tags for CodeStar Connections Connection (%s): %w", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func resourceConnectionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeStarConnectionsConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("updating CodeStar Connections Connection (%s) tags: %w", d.Id(), err)
		}
	}

	return nil
}

func resourceConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeStarConnectionsConn

	log.Printf("[DEBUG] Deleting CodeStar Connections Connection: %s", d.Id())
	_, err := conn.DeleteConnection(&codestarconnections.DeleteConnectionInput{
		ConnectionArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, codestarconnections.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting CodeStar Connections Connection (%s): %w", d.Id(), err)
	}

	return nil
}
