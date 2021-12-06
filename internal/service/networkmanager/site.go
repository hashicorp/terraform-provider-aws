package networkmanager

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/internal/keyvaluetags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceSite() *schema.Resource {
	return &schema.Resource{
		Create: ResourceSiteCreate,
		Read:   ResourceSiteRead,
		Update: ResourceSiteUpdate,
		Delete: ResourceSiteDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("arn", d.Id())

				idErr := fmt.Errorf("Expected ID in format of arn:aws:networkmanager::ACCOUNTID:site/GLOBALNETWORKID/SITEID and provided: %s", d.Id())

				resARN, err := arn.Parse(d.Id())
				if err != nil {
					return nil, idErr
				}

				identifiers := strings.TrimPrefix(resARN.Resource, "site/")
				identifierParts := strings.Split(identifiers, "/")
				if len(identifierParts) != 2 {
					return nil, idErr
				}
				d.SetId(identifierParts[1])
				d.Set("global_network_id", identifierParts[0])

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"global_network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"location": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"latitude": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"longitude": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func ResourceSiteCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).networkmanagerconn

	input := &networkmanager.CreateSiteInput{
		Description:     aws.String(d.Get("description").(string)),
		GlobalNetworkId: aws.String(d.Get("global_network_id").(string)),
		Location:        expandLocation(d.Get("location").([]interface{})),
		Tags:            keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().NetworkmanagerTags(),
	}

	log.Printf("[DEBUG] Creating Network Manager Site: %s", input)
	output, err := conn.CreateSite(input)
	if err != nil {
		return fmt.Errorf("error creating Network Manager Site: %s", err)
	}

	d.SetId(aws.StringValue(output.Site.SiteId))
	d.Set("global_network_id", aws.StringValue(output.Site.GlobalNetworkId))

	stateConf := &resource.StateChangeConf{
		Pending: []string{networkmanager.SiteStatePending},
		Target:  []string{networkmanager.SiteStateAvailable},
		Refresh: networkmanagerSiteRefreshFunc(conn, aws.StringValue(output.Site.GlobalNetworkId), aws.StringValue(output.Site.SiteId)),
		Timeout: 10 * time.Minute,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for Network Manager Site (%s) availability: %s", d.Id(), err)
	}

	return ResourceSiteRead(d, meta)
}

func ResourceSiteRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).networkmanagerconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	site, err := networkmanagerDescribeSite(conn, d.Get("global_network_id").(string), d.Id())

	if tfawserr.ErrCodeEquals(err, "InvalidSiteID.NotFound", "") {
		log.Printf("[WARN] Network Manager Site (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Network Manager Site: %s", err)
	}

	if site == nil {
		log.Printf("[WARN] Network Manager Site (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(site.State) == networkmanager.SiteStateDeleting {
		log.Printf("[WARN] Network Manager Site (%s) in deleted state (%s), removing from state", d.Id(), aws.StringValue(site.State))
		d.SetId("")
		return nil
	}

	d.Set("arn", site.SiteArn)
	d.Set("description", site.Description)

	if err := d.Set("location", flattenLocation(site.Location)); err != nil {
		return fmt.Errorf("error setting location: %s", err)
	}

	if err := d.Set("tags", keyvaluetags.NetworkmanagerKeyValueTags(site.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func ResourceSiteUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).networkmanagerconn

	if d.HasChanges("description", "location") {
		request := &networkmanager.UpdateSiteInput{
			Description:     aws.String(d.Get("description").(string)),
			GlobalNetworkId: aws.String(d.Get("global_network_id").(string)),
			Location:        expandLocation(d.Get("location").([]interface{})),
			SiteId:          aws.String(d.Id()),
		}

		_, err := conn.UpdateSite(request)
		if err != nil {
			return fmt.Errorf("Failure updating Network Manager Site (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.NetworkmanagerUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Network Manager Site (%s) tags: %s", d.Id(), err)
		}
	}

	return nil
}

func ResourceSiteDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).networkmanagerconn

	input := &networkmanager.DeleteSiteInput{
		GlobalNetworkId: aws.String(d.Get("global_network_id").(string)),
		SiteId:          aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Network Manager Site (%s): %s", d.Id(), input)
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteSite(input)

		if tfawserr.ErrMessageContains(err, "IncorrectState", "has non-deleted Link Associations") {
			return resource.RetryableError(err)
		}

		if tfawserr.ErrMessageContains(err, "IncorrectState", "has non-deleted Device") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteSite(input)
	}

	if tfawserr.ErrCodeEquals(err, "InvalidSiteID.NotFound", "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Network Manager Site: %s", err)
	}

	if err := waitForSiteDeletion(conn, d.Get("global_network_id").(string), d.Id()); err != nil {
		return fmt.Errorf("error waiting for Network Manager Site (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func networkmanagerSiteRefreshFunc(conn *networkmanager.NetworkManager, globalNetworkID, siteID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		site, err := networkmanagerDescribeSite(conn, globalNetworkID, siteID)

		if tfawserr.ErrCodeEquals(err, "InvalidSiteID.NotFound", "") {
			return nil, "DELETED", nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("error reading Network Manager Site (%s): %s", siteID, err)
		}

		if site == nil {
			return nil, "DELETED", nil
		}

		return site, aws.StringValue(site.State), nil
	}
}

func networkmanagerDescribeSite(conn *networkmanager.NetworkManager, globalNetworkID, siteID string) (*networkmanager.Site, error) {
	input := &networkmanager.GetSitesInput{
		GlobalNetworkId: aws.String(globalNetworkID),
		SiteIds:         []*string{aws.String(siteID)},
	}

	log.Printf("[DEBUG] Reading Network Manager Site (%s): %s", siteID, input)
	for {
		output, err := conn.GetSites(input)

		if err != nil {
			return nil, err
		}

		if output == nil || len(output.Sites) == 0 {
			return nil, nil
		}

		for _, site := range output.Sites {
			if site == nil {
				continue
			}

			if aws.StringValue(site.SiteId) == siteID {
				return site, nil
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil, nil
}

func waitForSiteDeletion(conn *networkmanager.NetworkManager, globalNetworkID, siteID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			networkmanager.SiteStateAvailable,
			networkmanager.SiteStateDeleting,
		},
		Target:         []string{""},
		Refresh:        networkmanagerSiteRefreshFunc(conn, globalNetworkID, siteID),
		Timeout:        10 * time.Minute,
		NotFoundChecks: 1,
	}

	log.Printf("[DEBUG] Waiting for Network Manager Site (%s) deletion", siteID)
	_, err := stateConf.WaitForState()

	if tfresource.NotFound(err) {
		return nil
	}

	return err
}
