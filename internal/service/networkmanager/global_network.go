package networkmanager

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceGlobalNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceGlobalNetworkCreate,
		Read:   resourceGlobalNetworkRead,
		Update: resourceGlobalNetworkUpdate,
		Delete: resourceGlobalNetworkDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Second),
			Delete: schema.DefaultTimeout(30 * time.Second),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceGlobalNetworkCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).NetworkManagerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &networkmanager.CreateGlobalNetworkInput{}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Println("[DEBUG] Creating Global Network:", input)
	output, err := conn.CreateGlobalNetwork(input)
	if err != nil {
		return fmt.Errorf("Error creating Global Network: %s", err)
	}

	d.SetId(aws.StringValue(output.GlobalNetwork.GlobalNetworkId))

	if err := waitForNetworkManagerGlobalNetworkCreation(conn, d.Id()); err != nil {
		return fmt.Errorf("Error waiting for Global Network: %w", err)
	}

	return resourceGlobalNetworkRead(d, meta)
}

func resourceGlobalNetworkRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).NetworkManagerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := DescribeGlobalNetwork(conn, d.Id())

	if tfawserr.ErrMessageContains(err, networkmanager.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] No Global Networks by ID (%s) found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error reading Global Network %s: %s", d.Id(), err)
	}

	if output == nil {
		log.Printf("[WARN] No Global Networks by ID (%s) found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(output.State) != networkmanager.ConnectionStateAvailable {
		log.Printf("[WARN] Global Networks (%s) delet(ing|ed), removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", output.GlobalNetworkArn)
	d.Set("description", output.Description)

	tags := KeyValueTags(output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("Error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("Error setting tags_all: %w", err)
	}

	return nil

}

func resourceGlobalNetworkUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).NetworkManagerConn

	if d.HasChanges("description") {
		request := &networkmanager.UpdateGlobalNetworkInput{
			GlobalNetworkId: aws.String(d.Id()),
			Description:     aws.String(d.Get("description").(string)),
		}

		log.Println("[DEBUG] Update Global Network request:", request)
		_, err := conn.UpdateGlobalNetwork(request)
		if err != nil {
			if tfawserr.ErrMessageContains(err, networkmanager.ErrCodeResourceNotFoundException, "") {
				log.Printf("[WARN] No Global Network by ID (%s) found", d.Id())
				d.SetId("")
				return nil
			}
			return fmt.Errorf("Error updating Global Network %s: %s", d.Id(), err)
		}

	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("Error updating Global Network (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceGlobalNetworkRead(d, meta)

}

func resourceGlobalNetworkDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).NetworkManagerConn

	input := &networkmanager.DeleteGlobalNetworkInput{
		GlobalNetworkId: aws.String(d.Id()),
	}

	log.Println("[DEBUG] Delete Global Network request:", input)
	_, err := conn.DeleteGlobalNetwork(input)
	if err != nil {
		return fmt.Errorf("Error deleting Global Network: %s", err)
	}

	if err := waitForNetworkManagerGlobalNetworkDeletion(conn, d.Id()); err != nil {
		return fmt.Errorf("Error waiting for Global Network (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func networkManagerGlobalNetworkRefresh(conn *networkmanager.NetworkManager, globalNetworkId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		networkManagerGlobalNetwork, err := DescribeGlobalNetwork(conn, globalNetworkId)
		if tfawserr.ErrMessageContains(err, networkmanager.ErrCodeResourceNotFoundException, "") {
			return nil, networkmanager.GlobalNetworkStateDeleting, nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("Error reading Global Network (%s): %s", globalNetworkId, err)
		}

		if networkManagerGlobalNetwork == nil {
			return nil, networkmanager.GlobalNetworkStateDeleting, nil
		}

		return networkManagerGlobalNetwork, aws.StringValue(networkManagerGlobalNetwork.State), nil
	}
}

func DescribeGlobalNetwork(conn *networkmanager.NetworkManager, globalNetworkID string) (*networkmanager.GlobalNetwork, error) {
	request := &networkmanager.DescribeGlobalNetworksInput{
		GlobalNetworkIds: []*string{aws.String(globalNetworkID)},
	}

	log.Printf("[DEBUG] Reading Global Network (%s): %s", globalNetworkID, request)
	for {
		output, err := conn.DescribeGlobalNetworks(request)

		if err != nil {
			return nil, err
		}

		if output == nil || len(output.GlobalNetworks) == 0 {
			return nil, nil
		}

		for _, globalNetwork := range output.GlobalNetworks {
			if globalNetwork == nil {
				continue
			}

			if aws.StringValue(globalNetwork.GlobalNetworkId) == globalNetworkID {
				return globalNetwork, nil
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		request.NextToken = output.NextToken
	}

	return nil, nil
}

func waitForNetworkManagerGlobalNetworkCreation(conn *networkmanager.NetworkManager, globalNetworkId string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkmanager.GlobalNetworkStatePending},
		Target:  []string{networkmanager.GlobalNetworkStateAvailable},
		Refresh: networkManagerGlobalNetworkRefresh(conn, globalNetworkId),
		Timeout: 10 * time.Minute,
	}

	log.Printf("[DEBUG] Waiting for Global Network (%s) creation", globalNetworkId)
	_, err := stateConf.WaitForState()

	return err
}

func waitForNetworkManagerGlobalNetworkDeletion(conn *networkmanager.NetworkManager, globalNetworkId string) error {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{networkmanager.GlobalNetworkStateAvailable},
		Target:         []string{networkmanager.GlobalNetworkStateDeleting},
		Refresh:        networkManagerGlobalNetworkRefresh(conn, globalNetworkId),
		Timeout:        10 * time.Minute,
		NotFoundChecks: 1,
	}

	log.Printf("[DEBUG] Waiting for Global Network (%s) deletion", globalNetworkId)
	_, err := stateConf.WaitForState()

	if tfresource.NotFound(err) {
		return nil
	}

	return err
}
