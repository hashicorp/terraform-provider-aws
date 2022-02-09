package aws

import (
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53domains"
	"github.com/hashicorp/terraform/helper/schema"
)

// setTags is a helper to set the tags for a resource. It expects the
// tags field to be named "tags"
func setTagsRoute53Domains(conn *route53domains.Route53Domains, d *schema.ResourceData) error {
	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		o := oraw.(map[string]interface{})
		n := nraw.(map[string]interface{})
		create, remove := diffTagsRoute53Domains(tagsFromMapRoute53Domains(o), tagsFromMapRoute53Domains(n))

		// Set tags
		if len(remove) > 0 {
			log.Printf("[DEBUG] Removing tags: %#v", remove)
			keys := make([]*string, 0, len(remove))
			// for k := range remove {
			// 	keys = append(keys, aws.String(k))
			// }
			_, err := conn.DeleteTagsForDomain(&route53domains.DeleteTagsForDomainInput{
				DomainName:   aws.String(d.Id()),
				TagsToDelete: keys,
			})
			if err != nil {
				return err
			}
		}
		if len(create) > 0 {
			log.Printf("[DEBUG] Creating tags: %#v", create)
			_, err := conn.UpdateTagsForDomain(&route53domains.UpdateTagsForDomainInput{
				DomainName:   aws.String(d.Id()),
				TagsToUpdate: create,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// diffTags takes our tags locally and the ones remotely and returns
// the set of tags that must be created, and the set of tags that must
// be destroyed.
func diffTagsRoute53Domains(oldTags, newTags []*route53domains.Tag) ([]*route53domains.Tag, []*route53domains.Tag) {
	// First, we're creating everything we have
	create := make(map[string]interface{})
	for _, t := range newTags {
		create[*t.Key] = *t.Value
	}

	// Build the list of what to remove
	var remove []*route53domains.Tag
	for _, t := range oldTags {
		old, ok := create[*t.Key]
		if !ok || old != *t.Value {
			// Delete it!
			remove = append(remove, t)
		}
	}

	return tagsFromMapRoute53Domains(create), remove
}

// tagsFromMap returns the tags for the given map of data.
func tagsFromMapRoute53Domains(m map[string]interface{}) []*route53domains.Tag {
	result := make([]*route53domains.Tag, 0, len(m))
	for k, v := range m {
		t := &route53domains.Tag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		}
		if !tagIgnoredRoute53Domains(t) {
			result = append(result, t)
		}
	}

	return result
}

// tagsToMap turns the list of tags into a map.
func tagsToMapRoute53Domains(ts []*route53domains.Tag) map[string]string {
	result := make(map[string]string)
	for _, t := range ts {
		if !tagIgnoredRoute53Domains(t) {
			result[*t.Key] = *t.Value
		}
	}

	return result
}

// compare a tag against a list of strings and checks if it should
// be ignored or not
func tagIgnoredRoute53Domains(t *route53domains.Tag) bool {
	filter := []string{"^aws:"}
	for _, v := range filter {
		log.Printf("[DEBUG] Matching %v with %v\n", v, *t.Key)
		r, _ := regexp.MatchString(v, *t.Key)
		if r {
			log.Printf("[DEBUG] Found AWS specific tag %s (val: %s), ignoring.\n", *t.Key, *t.Value)
			return true
		}
	}
	return false
}
