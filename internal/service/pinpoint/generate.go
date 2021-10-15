//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsOutTagsElem=TagsModel.Tags -ServiceTagsMap=yes "-TagInCustomVal=&pinpoint.TagsModel{Tags: Tags(updatedTags.IgnoreAWS())}" -TagInTagsElem=TagsModel -UpdateTags=yes
// ONLY generate directives and package declaration! Do not add anything else to this file.

package pinpoint
