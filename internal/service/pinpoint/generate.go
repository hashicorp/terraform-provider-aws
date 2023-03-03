//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOutTagsElem=TagsModel.Tags -ServiceTagsMap "-TagInCustomVal=&pinpoint.TagsModel{Tags: Tags(updatedTags.IgnoreAWS())}" -TagInTagsElem=TagsModel -UpdateTags -ContextOnly
// ONLY generate directives and package declaration! Do not add anything else to this file.

package pinpoint
