package redshift

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourcePartner() *schema.Resource {
	return &schema.Resource{
		Create: resourcePartnerCreate,
		Read:   resourcePartnerRead,
		Delete: resourcePartnerDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"cluster_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"database_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"partner_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status_message": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourcePartnerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	account := d.Get("account_id").(string)
	clusterId := d.Get("cluster_identifier").(string)
	input := redshift.AddPartnerInput{
		AccountId:         aws.String(account),
		ClusterIdentifier: aws.String(clusterId),
		DatabaseName:      aws.String(d.Get("database_name").(string)),
		PartnerName:       aws.String(d.Get("partner_name").(string)),
	}

	out, err := conn.AddPartner(&input)

	if err != nil {
		return fmt.Errorf("creating Redshift Partner: %w", err)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s:%s", account, clusterId, aws.StringValue(out.DatabaseName), aws.StringValue(out.PartnerName)))

	return resourcePartnerRead(d, meta)
}

func resourcePartnerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	out, err := FindPartnerById(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Partner (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading Redshift Partner (%s): %w", d.Id(), err)
	}

	d.Set("account_id", d.Get("account_id").(string))
	d.Set("cluster_identifier", d.Get("cluster_identifier").(string))
	d.Set("partner_name", out.PartnerName)
	d.Set("database_name", out.DatabaseName)
	d.Set("status", out.Status)
	d.Set("status_message", out.StatusMessage)

	return nil
}

func resourcePartnerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	account, clusterId, dbName, partnerName, err := DecodePartnerID(d.Id())
	if err != nil {
		return err
	}

	deleteInput := redshift.DeletePartnerInput{
		AccountId:         aws.String(account),
		ClusterIdentifier: aws.String(clusterId),
		DatabaseName:      aws.String(dbName),
		PartnerName:       aws.String(partnerName),
	}

	_, err = conn.DeletePartner(&deleteInput)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, redshift.ErrCodePartnerNotFoundFault) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("deleting Redshift Partner (%s): %w", d.Id(), err)
		}
	}

	return nil
}

func DecodePartnerID(id string) (string, string, string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 4 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" || idParts[3] == "" {
		return "", "", "", "", fmt.Errorf("expected ID to be the form account:cluster_identifier:database_name:partner_name, given: %s", id)
	}

	return idParts[0], idParts[1], idParts[2], idParts[3], nil
}
