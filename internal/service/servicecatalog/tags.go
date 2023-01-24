//go:build !generate
// +build !generate

package servicecatalog

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// Custom Service Catalog tag service update functions using the same format as generated code.

func productUpdateTags(ctx context.Context, conn *servicecatalog.ServiceCatalog, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := tftags.New(oldTagsMap)
	newTags := tftags.New(newTagsMap)

	input := &servicecatalog.UpdateProductInput{
		Id: aws.String(identifier),
	}

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input.RemoveTags = aws.StringSlice(removedTags.IgnoreAWS().Keys())
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input.AddTags = Tags(updatedTags.IgnoreAWS())
	}

	_, err := conn.UpdateProductWithContext(ctx, input)

	if err != nil {
		return fmt.Errorf("error updating tags for Service Catalog Product (%s): %w", identifier, err)
	}

	return nil
}

func recordKeyValueTags(tags []*servicecatalog.RecordTag) tftags.KeyValueTags {
	m := make(map[string]*string, len(tags))

	for _, tag := range tags {
		m[aws.StringValue(tag.Key)] = tag.Value
	}

	return tftags.New(m)
}
