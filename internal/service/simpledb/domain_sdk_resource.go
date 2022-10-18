package simpledb

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// Stump of the SDK v2 implementation retained for acceptance tests and sweepers.
// TODO Remove.
func ResourceDomain() *schema.Resource {
	return &schema.Resource{
		DeleteWithoutTimeout: resourceDomainDelete,
	}
}

func resourceDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SimpleDBConn

	log.Printf("[DEBUG] Deleting SimpleDB Domain: %s", d.Id())
	_, err := conn.DeleteDomainWithContext(ctx, &simpledb.DeleteDomainInput{
		DomainName: aws.String(d.Id()),
	})

	if err != nil {
		return diag.Errorf("deleting SimpleDB Domain (%s): %s", d.Id(), err)
	}

	return nil
}
