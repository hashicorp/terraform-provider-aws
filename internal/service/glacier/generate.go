//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=ListTagsForVault -ListTagsInIDElem=VaultName -ServiceTagsMap -TagOp=AddTagsToVault -TagInIDElem=VaultName -UntagOp=RemoveTagsFromVault -UpdateTags -ContextOnly
// ONLY generate directives and package declaration! Do not add anything else to this file.

package glacier
