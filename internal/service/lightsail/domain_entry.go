package lightsail

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	DomainEntryIdPartsCount = 4
	ResNameDomainEntry      = "DomainEntry"
)

// @SDKResource("aws_lightsail_domain_entry")
func ResourceDomainEntry() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainEntryCreate,
		ReadWithoutTimeout:   resourceDomainEntryRead,
		DeleteWithoutTimeout: resourceDomainEntryDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"is_alias": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"target": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"A",
					"CNAME",
					"MX",
					"NS",
					"SOA",
					"SRV",
					"TXT",
				}, false),
				ForceNew: true,
			},
		},
	}
}

func resourceDomainEntryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()
	name := d.Get("name").(string)
	req := &lightsail.CreateDomainEntryInput{
		DomainName: aws.String(d.Get("domain_name").(string)),

		DomainEntry: &lightsail.DomainEntry{
			IsAlias: aws.Bool(d.Get("is_alias").(bool)),
			Name:    aws.String(expandDomainEntryName(name, d.Get("domain_name").(string))),
			Target:  aws.String(d.Get("target").(string)),
			Type:    aws.String(d.Get("type").(string)),
		},
	}

	resp, err := conn.CreateDomainEntryWithContext(ctx, req)

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateDomain, ResNameDomainEntry, name, err)
	}

	diag := expandOperations(ctx, conn, []*lightsail.Operation{resp.Operation}, lightsail.OperationTypeCreateDomain, ResNameDomainEntry, name)

	if diag != nil {
		return diag
	}

	// Generate an ID
	idParts := []string{
		name,
		d.Get("domain_name").(string),
		d.Get("type").(string),
		d.Get("target").(string),
	}

	id, err := flex.FlattenResourceId(idParts, DomainEntryIdPartsCount, true)

	if err != nil {
		return create.DiagError(names.DynamoDB, create.ErrActionFlatteningResourceId, ResNameDomainEntry, d.Get("domain_name").(string), err)
	}

	d.SetId(id)

	return resourceDomainEntryRead(ctx, d, meta)
}

func resourceDomainEntryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	entry, err := FindDomainEntryById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.Lightsail, create.ErrActionReading, ResNameDomainEntry, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResNameDomainEntry, d.Id(), err)
	}

	domainName, err := expandDomainNameFromId(d.Id())

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionExpandingResourceId, ResNameDomainEntry, d.Id(), err)
	}

	name := flattenDomainEntryName(aws.StringValue(entry.Name), domainName)

	partCount := flex.ResourceIdPartCount(d.Id())

	// This code is intended to update the Id to use the common separator for resources still using the old separator
	if partCount == 1 {

		idParts := []string{
			name,
			domainName,
			aws.StringValue(entry.Type),
			aws.StringValue(entry.Target),
		}

		id, err := flex.FlattenResourceId(idParts, DomainEntryIdPartsCount, true)

		if err != nil {
			return create.DiagError(names.DynamoDB, create.ErrActionFlatteningResourceId, ResNameDomainEntry, d.Get("domain_name").(string), err)
		}

		d.SetId(id)
	}
	d.Set("name", name)
	d.Set("domain_name", domainName)
	d.Set("type", entry.Type)
	d.Set("is_alias", entry.IsAlias)
	d.Set("target", entry.Target)

	return nil
}

func resourceDomainEntryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	domainName, err := expandDomainNameFromId(d.Id())

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionExpandingResourceId, ResNameDomainEntry, d.Id(), err)
	}

	domainEntry, err := expandDomainEntry(d.Id())

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionExpandingResourceId, ResNameDomainEntry, d.Id(), err)
	}

	resp, err := conn.DeleteDomainEntryWithContext(ctx, &lightsail.DeleteDomainEntryInput{
		DomainName:  aws.String(domainName),
		DomainEntry: domainEntry,
	})

	if err != nil && tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionDeleting, ResNameDomainEntry, d.Id(), err)
	}

	diag := expandOperations(ctx, conn, []*lightsail.Operation{resp.Operation}, lightsail.OperationTypeDeleteDomain, ResNameDomainEntry, d.Id())

	if diag != nil {
		return diag
	}

	return nil
}

func expandDomainEntry(id string) (*lightsail.DomainEntry, error) {
	partCount := flex.ResourceIdPartCount(id)

	var name string
	var domainName string
	var recordType string
	var recordTarget string

	if partCount == 1 {
		idParts := strings.Split(id, "_")
		idLength := len(idParts)
		var index int

		if idLength == 5 {
			index = 1
			name = "_" + idParts[index+0]
		} else {
			index = 0
			name = idParts[index+0]
		}

		domainName = idParts[index+1]
		recordType = idParts[index+2]
		recordTarget = idParts[index+3]
	} else {
		idParts, err := flex.ExpandResourceId(id, DomainEntryIdPartsCount, true)

		if err != nil {
			return nil, err
		}
		name = idParts[0]
		domainName = idParts[1]
		recordType = idParts[2]
		recordTarget = idParts[3]
	}
	entry := &lightsail.DomainEntry{
		Name:   aws.String(expandDomainEntryName(name, domainName)),
		Type:   aws.String(recordType),
		Target: aws.String(recordTarget),
	}

	return entry, nil
}

func expandDomainNameFromId(id string) (string, error) {
	partCount := flex.ResourceIdPartCount(id)
	var domainName string

	if partCount == 1 {
		idParts := strings.Split(id, "_")
		idLength := len(idParts)
		var index int

		if idLength == 5 {
			index = 1
		} else {
			index = 0
		}

		domainName = idParts[index+1]
	} else {
		idParts, err := flex.ExpandResourceId(id, DomainEntryIdPartsCount, true)

		if err != nil {
			return "", err
		}

		domainName = idParts[1]
	}
	return domainName, nil
}

func expandDomainEntryName(name, domainName string) string {
	rn := strings.ToLower(strings.TrimSuffix(name, "."))
	domainName = strings.TrimSuffix(domainName, ".")
	if !strings.HasSuffix(rn, domainName) {
		if len(name) == 0 {
			rn = domainName
		} else {
			rn = strings.Join([]string{rn, domainName}, ".")
		}
	}
	return rn
}

func flattenDomainEntryName(name, domainName string) string {
	rn := strings.ToLower(strings.TrimSuffix(name, "."))
	domainName = strings.TrimSuffix(domainName, ".")
	if strings.HasSuffix(rn, domainName) {
		rn = strings.TrimSuffix(rn, fmt.Sprintf(".%s", domainName))
	}
	if rn == domainName {
		rn = ""
	}
	return rn
}

func FindDomainEntryById(ctx context.Context, conn *lightsail.Lightsail, id string) (*lightsail.DomainEntry, error) {
	partCount := flex.ResourceIdPartCount(id)

	in := &lightsail.GetDomainInput{}
	var name string
	var domainName string
	var entryName string
	var recordType string
	var recordTarget string

	// if there is not more than one partCount, the legacy separator will be used.
	if partCount == 1 {

		idParts := strings.Split(id, "_")
		idLength := len(idParts)
		var index int

		if idLength <= 3 {
			return nil, tfresource.NewEmptyResultError(in)
		}

		if idLength == 5 {
			index = 1
			name = "_" + idParts[index]
		} else {
			index = 0
			name = idParts[index]
		}

		domainName = idParts[index+1]
		entryName = expandDomainEntryName(name, domainName)
		recordType = idParts[index+2]
		recordTarget = idParts[index+3]
	} else {
		idParts, err := flex.ExpandResourceId(id, DomainEntryIdPartsCount, true)

		if err != nil {
			return nil, err
		}

		name = idParts[0]
		domainName = idParts[1]
		entryName = expandDomainEntryName(name, domainName)
		recordType = idParts[2]
		recordTarget = idParts[3]
	}

	in.DomainName = aws.String(domainName)

	out, err := conn.GetDomainWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	var entry *lightsail.DomainEntry
	entryExists := false

	for _, n := range out.Domain.DomainEntries {
		if entryName == aws.StringValue(n.Name) && recordType == aws.StringValue(n.Type) && recordTarget == aws.StringValue(n.Target) {
			entry = n
			entryExists = true
			break
		}
	}

	if !entryExists {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return entry, nil
}
