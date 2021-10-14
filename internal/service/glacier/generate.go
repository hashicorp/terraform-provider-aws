//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsOp=ListTagsForVault -ListTagsInIDElem=VaultName -ServiceTagsMap=yes -TagOp=AddTagsToVault -TagInIDElem=VaultName -UntagOp=RemoveTagsFromVault -UpdateTags=yes
// ONLY generate directives and package declaration! Do not add anything else to this file.

package glacier
