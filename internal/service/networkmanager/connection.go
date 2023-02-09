package networkmanager

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceConnection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConnectionCreate,
		ReadWithoutTimeout:   resourceConnectionRead,
		UpdateWithoutTimeout: resourceConnectionUpdate,
		DeleteWithoutTimeout: resourceConnectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				parsedARN, err := arn.Parse(d.Id())

				if err != nil {
					return nil, fmt.Errorf("error parsing ARN (%s): %w", d.Id(), err)
				}

				// See https://docs.aws.amazon.com/service-authorization/latest/reference/list_networkmanager.html#networkmanager-resources-for-iam-policies.
				resourceParts := strings.Split(parsedARN.Resource, "/")

				if actual, expected := len(resourceParts), 3; actual < expected {
					return nil, fmt.Errorf("expected at least %d resource parts in ARN (%s), got: %d", expected, d.Id(), actual)
				}

				d.SetId(resourceParts[2])
				d.Set("global_network_id", resourceParts[1])

				return []*schema.ResourceData{d}, nil
			},
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connected_device_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"connected_link_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"device_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"global_network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"link_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	globalNetworkID := d.Get("global_network_id").(string)

	input := &networkmanager.CreateConnectionInput{
		ConnectedDeviceId: aws.String(d.Get("connected_device_id").(string)),
		DeviceId:          aws.String(d.Get("device_id").(string)),
		GlobalNetworkId:   aws.String(globalNetworkID),
	}

	if v, ok := d.GetOk("connected_link_id"); ok {
		input.ConnectedLinkId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("link_id"); ok {
		input.LinkId = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating Network Manager Connection: %s", input)
	output, err := conn.CreateConnectionWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error creating Network Manager Connection: %s", err)
	}

	d.SetId(aws.StringValue(output.Connection.ConnectionId))

	if _, err := waitConnectionCreated(ctx, conn, globalNetworkID, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("error waiting for Network Manager Connection (%s) create: %s", d.Id(), err)
	}

	return resourceConnectionRead(ctx, d, meta)
}

func resourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	globalNetworkID := d.Get("global_network_id").(string)
	connection, err := FindConnectionByTwoPartKey(ctx, conn, globalNetworkID, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Network Manager Connection %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading Network Manager Connection (%s): %s", d.Id(), err)
	}

	d.Set("arn", connection.ConnectionArn)
	d.Set("connected_device_id", connection.ConnectedDeviceId)
	d.Set("connected_link_id", connection.ConnectedLinkId)
	d.Set("description", connection.Description)
	d.Set("device_id", connection.DeviceId)
	d.Set("global_network_id", connection.GlobalNetworkId)
	d.Set("link_id", connection.LinkId)

	tags := KeyValueTags(connection.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("error setting tags_all: %s", err)
	}

	return nil
}

func resourceConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn()

	if d.HasChangesExcept("tags", "tags_all") {
		globalNetworkID := d.Get("global_network_id").(string)
		input := &networkmanager.UpdateConnectionInput{
			ConnectedLinkId: aws.String(d.Get("connected_link_id").(string)),
			ConnectionId:    aws.String(d.Id()),
			Description:     aws.String(d.Get("description").(string)),
			GlobalNetworkId: aws.String(globalNetworkID),
			LinkId:          aws.String(d.Get("link_id").(string)),
		}

		log.Printf("[DEBUG] Updating Network Manager Connection: %s", input)
		_, err := conn.UpdateConnectionWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("error updating Network Manager Connection (%s): %s", d.Id(), err)
		}

		if _, err := waitConnectionUpdated(ctx, conn, globalNetworkID, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("error waiting for Network Manager Connection (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("error updating Network Manager Connection (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceConnectionRead(ctx, d, meta)
}

func resourceConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn()

	globalNetworkID := d.Get("global_network_id").(string)

	log.Printf("[DEBUG] Deleting Network Manager Connection: %s", d.Id())
	_, err := conn.DeleteConnectionWithContext(ctx, &networkmanager.DeleteConnectionInput{
		ConnectionId:    aws.String(d.Id()),
		GlobalNetworkId: aws.String(globalNetworkID),
	})

	if globalNetworkIDNotFoundError(err) || tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting Network Manager Connection (%s): %s", d.Id(), err)
	}

	if _, err := waitConnectionDeleted(ctx, conn, globalNetworkID, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("error waiting for Network Manager Connection (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func FindConnection(ctx context.Context, conn *networkmanager.NetworkManager, input *networkmanager.GetConnectionsInput) (*networkmanager.Connection, error) {
	output, err := FindConnections(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindConnections(ctx context.Context, conn *networkmanager.NetworkManager, input *networkmanager.GetConnectionsInput) ([]*networkmanager.Connection, error) {
	var output []*networkmanager.Connection

	err := conn.GetConnectionsPagesWithContext(ctx, input, func(page *networkmanager.GetConnectionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Connections {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if globalNetworkIDNotFoundError(err) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindConnectionByTwoPartKey(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, connectionID string) (*networkmanager.Connection, error) {
	input := &networkmanager.GetConnectionsInput{
		ConnectionIds:   aws.StringSlice([]string{connectionID}),
		GlobalNetworkId: aws.String(globalNetworkID),
	}

	output, err := FindConnection(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.GlobalNetworkId) != globalNetworkID || aws.StringValue(output.ConnectionId) != connectionID {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func statusConnectionState(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, connectionID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindConnectionByTwoPartKey(ctx, conn, globalNetworkID, connectionID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func waitConnectionCreated(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, connectionID string, timeout time.Duration) (*networkmanager.Connection, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkmanager.ConnectionStatePending},
		Target:  []string{networkmanager.ConnectionStateAvailable},
		Timeout: timeout,
		Refresh: statusConnectionState(ctx, conn, globalNetworkID, connectionID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.Connection); ok {
		return output, err
	}

	return nil, err
}

func waitConnectionDeleted(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, connectionID string, timeout time.Duration) (*networkmanager.Connection, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkmanager.ConnectionStateDeleting},
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusConnectionState(ctx, conn, globalNetworkID, connectionID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.Connection); ok {
		return output, err
	}

	return nil, err
}

func waitConnectionUpdated(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, connectionID string, timeout time.Duration) (*networkmanager.Connection, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkmanager.ConnectionStateUpdating},
		Target:  []string{networkmanager.ConnectionStateAvailable},
		Timeout: timeout,
		Refresh: statusConnectionState(ctx, conn, globalNetworkID, connectionID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.Connection); ok {
		return output, err
	}

	return nil, err
}
