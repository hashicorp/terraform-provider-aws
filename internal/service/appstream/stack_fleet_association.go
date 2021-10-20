package appstream

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceStackFleetAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStackFleetAssociationCreate,
		ReadWithoutTimeout:   resourceStackFleetAssociationRead,
		DeleteWithoutTimeout: resourceStackFleetAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"fleet_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"stack_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceStackFleetAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppStreamConn
	input := &appstream.AssociateFleetInput{
		FleetName: aws.String(d.Get("fleet_name").(string)),
		StackName: aws.String(d.Get("stack_name").(string)),
	}
	var err error
	err = resource.RetryContext(ctx, fleetOperationTimeout, func() *resource.RetryError {
		_, err = conn.AssociateFleetWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.AssociateFleetWithContext(ctx, input)
	}
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Appstream Stack Fleet Association (%s): %w", d.Id(), err))
	}

	d.SetId(fmt.Sprintf("%s/%s", d.Get("stack_name").(string), d.Get("fleet_name").(string)))

	return resourceStackFleetAssociationRead(ctx, d, meta)
}

func resourceStackFleetAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppStreamConn

	stackName, fleetName, err := DecodeStackFleetID(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("error decoding id Appstream Stack Fleet Association (%s): %w", d.Id(), err))
	}

	resp, err := conn.ListAssociatedStacksWithContext(ctx, &appstream.ListAssociatedStacksInput{FleetName: aws.String(fleetName)})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Appstream Stack Fleet Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	var sName string
	for _, name := range resp.Names {
		if aws.StringValue(name) == stackName {
			sName = aws.StringValue(name)
		}
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Appstream Stack Fleet Association (%s): %w", d.Id(), err))
	}
	if len(sName) == 0 {
		log.Printf("[WARN] Appstream Stack Fleet Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("fleet_name", fleetName)
	d.Set("stack_name", sName)

	return nil
}

func resourceStackFleetAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppStreamConn

	fleetName, stackName, err := DecodeStackFleetID(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("error decoding id Appstream Stack Fleet Association (%s): %w", d.Id(), err))
	}

	_, err = conn.DisassociateFleetWithContext(ctx, &appstream.DisassociateFleetInput{
		StackName: aws.String(stackName),
		FleetName: aws.String(fleetName),
	})

	if err != nil {
		if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting Appstream Stack Fleet Association (%s): %w", d.Id(), err))
	}
	return nil
}

func DecodeStackFleetID(id string) (string, string, error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("expected ID in format StackName-FleetName, received: %s", id)
	}
	return idParts[0], idParts[1], nil
}
