package aws

import (
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform/helper/schema"
)

// getTags is a helper to get the tags for a resource. It expects the
// tags field to be named "tags"
func getTagsRoute53Resolver(conn *route53resolver.Route53Resolver, d *schema.ResourceData, arn string) error {
	tags := make([]*route53resolver.Tag, 0)
	req := &route53resolver.ListTagsForResourceInput{
		ResourceArn: aws.String(arn),
	}
	for {
		resp, err := conn.ListTagsForResource(req)
		if err != nil {
			return err
		}

		for _, tag := range resp.Tags {
			tags = append(tags, tag)
		}

		if resp.NextToken == nil {
			break
		}
		req.NextToken = resp.NextToken
	}

	if err := d.Set("tags", tagsToMapRoute53Resolver(tags)); err != nil {
		return err
	}

	return nil
}

// setTags is a helper to set the tags for a resource. It expects the
// tags field to be named "tags"
func setTagsRoute53Resolver(conn *route53resolver.Route53Resolver, d *schema.ResourceData, arn string) error {
	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		o := oraw.(map[string]interface{})
		n := nraw.(map[string]interface{})
		create, remove := diffTagsRoute53Resolver(tagsFromMapRoute53Resolver(o), tagsFromMapRoute53Resolver(n))

		// Set tags
		if len(remove) > 0 {
			log.Printf("[DEBUG] Removing tags: %#v", remove)
			k := make([]*string, len(remove), len(remove))
			for i, t := range remove {
				k[i] = t.Key
			}

			_, err := conn.UntagResource(&route53resolver.UntagResourceInput{
				ResourceArn: aws.String(arn),
				TagKeys:     k,
			})
			if err != nil {
				return err
			}
		}
		if len(create) > 0 {
			log.Printf("[DEBUG] Creating tags: %#v", create)
			_, err := conn.TagResource(&route53resolver.TagResourceInput{
				ResourceArn: aws.String(arn),
				Tags:        create,
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
func diffTagsRoute53Resolver(oldTags, newTags []*route53resolver.Tag) ([]*route53resolver.Tag, []*route53resolver.Tag) {
	// First, we're creating everything we have
	create := make(map[string]interface{})
	for _, t := range newTags {
		create[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}

	// Build the list of what to remove
	var remove []*route53resolver.Tag
	for _, t := range oldTags {
		old, ok := create[aws.StringValue(t.Key)]
		if !ok || old != aws.StringValue(t.Value) {
			// Delete it!
			remove = append(remove, t)
		}
	}

	return tagsFromMapRoute53Resolver(create), remove
}

// tagsFromMap returns the tags for the given map of data.
func tagsFromMapRoute53Resolver(m map[string]interface{}) []*route53resolver.Tag {
	result := make([]*route53resolver.Tag, 0, len(m))
	for k, v := range m {
		t := &route53resolver.Tag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		}
		if !tagIgnoredRoute53Resolver(t) {
			result = append(result, t)
		}
	}

	return result
}

// tagsToMap turns the list of tags into a map.
func tagsToMapRoute53Resolver(ts []*route53resolver.Tag) map[string]string {
	result := make(map[string]string)
	for _, t := range ts {
		if !tagIgnoredRoute53Resolver(t) {
			result[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
		}
	}

	return result
}

// compare a tag against a list of strings and checks if it should
// be ignored or not
func tagIgnoredRoute53Resolver(t *route53resolver.Tag) bool {
	filter := []string{"^aws:"}
	for _, v := range filter {
		log.Printf("[DEBUG] Matching %v with %v\n", v, *t.Key)
		if r, _ := regexp.MatchString(v, *t.Key); r == true {
			log.Printf("[DEBUG] Found AWS specific tag %s (val: %s), ignoring.\n", *t.Key, *t.Value)
			return true
		}
	}
	return false
}
