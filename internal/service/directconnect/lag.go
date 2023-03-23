package directconnect

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceLag() *schema.Resource {
	return &schema.Resource{
		Create: resourceLagCreate,
		Read:   resourceLagRead,
		Update: resourceLagUpdate,
		Delete: resourceLagDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"connections_bandwidth": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validConnectionBandWidth(),
			},
			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"has_logical_redundancy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"jumbo_frame_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"location": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"owner_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provider_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLagCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &directconnect.CreateLagInput{
		ConnectionsBandwidth: aws.String(d.Get("connections_bandwidth").(string)),
		LagName:              aws.String(name),
		Location:             aws.String(d.Get("location").(string)),
	}

	var connectionIDSpecified bool
	if v, ok := d.GetOk("connection_id"); ok {
		connectionIDSpecified = true
		input.ConnectionId = aws.String(v.(string))
		input.NumberOfConnections = aws.Int64(1)
	} else {
		input.NumberOfConnections = aws.Int64(1)
	}

	if v, ok := d.GetOk("provider_name"); ok {
		input.ProviderName = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating Direct Connect LAG: %s", input)
	output, err := conn.CreateLag(input)

	if err != nil {
		return fmt.Errorf("error creating Direct Connect LAG (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.LagId))

	// Delete unmanaged connection.
	if !connectionIDSpecified {
		err = deleteConnection(conn, aws.StringValue(output.Connections[0].ConnectionId), waitConnectionDeleted)

		if err != nil {
			return err
		}
	}

	return resourceLagRead(d, meta)
}

func resourceLagRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	lag, err := FindLagByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Direct Connect LAG (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Direct Connect LAG (%s): %w", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    aws.StringValue(lag.Region),
		Service:   "directconnect",
		AccountID: aws.StringValue(lag.OwnerAccount),
		Resource:  fmt.Sprintf("dxlag/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("connections_bandwidth", lag.ConnectionsBandwidth)
	d.Set("has_logical_redundancy", lag.HasLogicalRedundancy)
	d.Set("jumbo_frame_capable", lag.JumboFrameCapable)
	d.Set("location", lag.Location)
	d.Set("name", lag.LagName)
	d.Set("owner_account_id", lag.OwnerAccount)
	d.Set("provider_name", lag.ProviderName)

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Direct Connect LAG (%s): %w", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceLagUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn

	if d.HasChange("name") {
		input := &directconnect.UpdateLagInput{
			LagId:   aws.String(d.Id()),
			LagName: aws.String(d.Get("name").(string)),
		}

		log.Printf("[DEBUG] Updating Direct Connect LAG: %s", input)
		_, err := conn.UpdateLag(input)

		if err != nil {
			return fmt.Errorf("error updating Direct Connect LAG (%s): %w", d.Id(), err)
		}
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Direct Connect LAG (%s) tags: %w", arn, err)
		}
	}

	return resourceLagRead(d, meta)
}

func resourceLagDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn

	if d.Get("force_destroy").(bool) {
		lag, err := FindLagByID(conn, d.Id())

		if tfresource.NotFound(err) {
			return nil
		}

		for _, connection := range lag.Connections {
			err = deleteConnection(conn, aws.StringValue(connection.ConnectionId), waitConnectionDeleted)

			if err != nil {
				return err
			}
		}
	} else if v, ok := d.GetOk("connection_id"); ok {
		if err := deleteConnectionLAGAssociation(conn, v.(string), d.Id()); err != nil {
			return err
		}
	}

	log.Printf("[DEBUG] Deleting Direct Connect LAG: %s", d.Id())
	_, err := conn.DeleteLag(&directconnect.DeleteLagInput{
		LagId: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, directconnect.ErrCodeClientException, "Could not find Lag with ID") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Direct Connect LAG (%s): %w", d.Id(), err)
	}

	_, err = waitLagDeleted(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for Direct Connect LAG (%s) delete: %w", d.Id(), err)
	}

	return nil
}
