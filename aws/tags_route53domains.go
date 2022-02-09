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
		updateTags, deleteTags := diffTagsRoute53Domains(o, n)

		// Set tags
		if len(deleteTags) > 0 {
			log.Printf("[DEBUG] Removing tags: %#v", deleteTags)
			_, err := conn.DeleteTagsForDomain(&route53domains.DeleteTagsForDomainInput{
				DomainName:   aws.String(d.Id()),
				TagsToDelete: deleteTags,
			})
			if err != nil {
				return err
			}
		}
		if len(updateTags) > 0 {
			log.Printf("[DEBUG] Updating tags: %#v", updateTags)
			_, err := conn.UpdateTagsForDomain(&route53domains.UpdateTagsForDomainInput{
				DomainName:   aws.String(d.Id()),
				TagsToUpdate: updateTags,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// diffTags takes our tags locally and the ones remotely and returns
// the set of tags that must be updated, and the set of tags that must
// be deleted.
func diffTagsRoute53Domains(oldTags, newTags map[string]interface{}) ([]*route53domains.Tag, []*string) {
	// Is the key from newTags in oldTags
	// No - it's a creation
	// Yes - if it's differnet, it's an update
	updateTags := make(map[string]interface{})
	for k, v := range newTags {
		old, ok := oldTags[k]
		if !ok || old != v {
			updateTags[k] = v
		}
	}
	log.Printf("[DEBUG] Route 53 Domain tags to update: %#v", updateTags)

	// Is the key from oldTags in newTags
	// No - it's a deletion
	var deleteTags []*string
	for k, _ := range oldTags {
		_, ok := newTags[k]
		if !ok {
			deleteTags = append(deleteTags, aws.String(k))
		}
	}
	log.Printf("[DEBUG] Route 53 Domain tags to delete: %#v", deleteTags)

	// aws/tags_route53domains.go:56:22: invalid indirect of k (type string)
	// aws/tags_route53domains.go:57:20: invalid indirect of v (type interface {})
	// aws/tags_route53domains.go:58:15: invalid indirect of k (type string)
	// aws/tags_route53domains.go:58:21: invalid indirect of v (type interface {})
	// aws/tags_route53domains.go:66:20: invalid indirect of k (type string)
	// aws/tags_route53domains.go:68:23: cannot use k (type string) as type *string in append

	return tagsFromMapRoute53Domains(updateTags), deleteTags
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
