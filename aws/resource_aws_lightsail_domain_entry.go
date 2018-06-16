package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsLightsailDomainEntry() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLightsailDomainEntryCreate,
		Read:   resourceAwsLightsailDomainEntryRead,
		Update: resourceAwsLightsailDomainEntryUpdate,
		Delete: resourceAwsLightsailDomainEntryDelete,

		Schema: map[string]*schema.Schema{
			"domain_entry": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"is_alias": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"target": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsLightsailDomainEntryCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lightsailconn

	domainName := d.Get("domain_name").(string)
	domainEntry := d.Get("domain_entry").([]interface{})[0].(map[string]interface{})

	input := &lightsail.CreateDomainEntryInput{
		DomainEntry: &lightsail.DomainEntry{
			Name:   aws.String(domainEntry["name"].(string)),
			Target: aws.String(domainEntry["target"].(string)),
			Type:   aws.String(domainEntry["type"].(string)),
		},
		DomainName: aws.String(domainName),
	}

	if domainEntry["is_alias"] != nil {
		input.DomainEntry.IsAlias = aws.Bool(domainEntry["is_alias"].(bool))
	}
	log.Printf("[DEBUG] domain entry to create: %s", input)

	_, err := conn.CreateDomainEntry(input)
	if err != nil {
		return err
	}

	d.SetId(domainEntry["name"].(string))
	return resourceAwsLightsailDomainEntryRead(d, meta)
}

func resourceAwsLightsailDomainEntryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lightsailconn

	domainName := d.Get("domain_name").(string)
	domainEntry := d.Get("domain_entry").([]interface{})[0].(map[string]interface{})

	out, err := conn.GetDomain(&lightsail.GetDomainInput{
		DomainName: aws.String(domainName),
	})
	if err != nil {
		return err
	}

	for _, entry := range out.Domain.DomainEntries {
		if *entry.Name == domainEntry["name"].(string) {
			log.Print("[DEBUG] matched entry:", entry)
			d.Set("domain_entry", []interface{}{
				map[string]interface{}{
					"id":       *entry.Id,
					"is_alias": *entry.IsAlias,
					"name":     *entry.Name,
					"target":   *entry.Target,
					"type":     *entry.Type,
				},
			})
			return nil
		}
	}
	return fmt.Errorf("[WARN] Domain entry (%s) not found", domainEntry["name"].(string))
}

func resourceAwsLightsailDomainEntryUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lightsailconn

	domainName := d.Get("domain_name").(string)
	input := &lightsail.UpdateDomainEntryInput{
		DomainName: aws.String(domainName),
	}

	oldRaw, _ := d.GetChange("domain_entry")
	oldEntry := oldRaw.([]interface{})[0].(map[string]interface{})
	newEntry := d.Get("domain_entry").([]interface{})[0].(map[string]interface{})
	input.DomainEntry = &lightsail.DomainEntry{
		Id:     aws.String(oldEntry["id"].(string)),
		Name:   aws.String(newEntry["name"].(string)),
		Target: aws.String(newEntry["target"].(string)),
		Type:   aws.String(newEntry["type"].(string)),
	}
	if newEntry["is_alias"] != nil {
		input.DomainEntry.IsAlias = aws.Bool(newEntry["is_alias"].(bool))
	}
	log.Printf("[DEBUG] domain entry to update: %s", input)

	_, err := conn.UpdateDomainEntry(input)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NotFoundException" {
				log.Printf("[WARN] Lightsail Domain Entry (%s) not found, removing from state", d.Id())
				d.SetId("")
				return nil
			}
			return err
		}
		return err
	}
	return resourceAwsLightsailDomainEntryRead(d, meta)
}

func resourceAwsLightsailDomainEntryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lightsailconn

	domainEntry := d.Get("domain_entry").([]interface{})[0].(map[string]interface{})
	input := &lightsail.DeleteDomainEntryInput{
		DomainEntry: &lightsail.DomainEntry{
			Id:      aws.String(domainEntry["id"].(string)),
			IsAlias: aws.Bool(domainEntry["is_alias"].(bool)),
			Name:    aws.String(domainEntry["name"].(string)),
			Target:  aws.String(domainEntry["target"].(string)),
			Type:    aws.String(domainEntry["type"].(string)),
		},
		DomainName: aws.String(d.Get("domain_name").(string)),
	}
	log.Print("[DEBUG] domain entry to delete:", input)
	_, err := conn.DeleteDomainEntry(input)
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}
