#!/bin/bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

set -o errexit
set -o nounset

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__parent="$(dirname "$__dir")"
__root="$(dirname "$__parent")"

CHANGELOG_FILE_NAME="CHANGELOG.md"
CHANGELOG_TMP_FILE_NAME="CHANGELOG.tmp"
TARGET_SHA=$(git rev-parse HEAD)
PREVIOUS_RELEASE_TAG=$(git describe --abbrev=0 --match='v*.*.*' --tags)
PREVIOUS_RELEASE_SHA=$(git rev-list -n 1 $PREVIOUS_RELEASE_TAG)

if [ $TARGET_SHA == $PREVIOUS_RELEASE_SHA ]; then
  echo "Nothing to do"
  exit 0
fi

# Guard: the reassembly below assumes the first line of the CHANGELOG is the
# current "## <version> (Unreleased)" heading that separates not-yet-released
# entries from already-released sections. If that heading is missing (for
# example, the post-release "Update Changelog" job never ran), keying off the
# previous release tag makes this script re-append the previous release section
# on every run, duplicating it. Fail loudly instead of corrupting the file.
FIRST_LINE=$(head -n 1 "$__root/$CHANGELOG_FILE_NAME")
if [[ ! "$FIRST_LINE" =~ ^##[[:space:]].*\(Unreleased\)$ ]]; then
  echo "ERROR: expected the first line of $CHANGELOG_FILE_NAME to be a '## <version> (Unreleased)' heading, but found:" >&2
  echo "  ${FIRST_LINE}" >&2
  echo "Refusing to generate the changelog to avoid duplicating released sections." >&2
  echo "Confirm the 'Update Changelog' workflow added the next '## <version> (Unreleased)' heading, then re-run." >&2
  exit 1
fi

# The unreleased version at the top of the file must differ from the previous
# release tag. If they match, the previous-release marker and the file's top
# section coincide and the reassembly would duplicate that section.
UNRELEASED_VERSION=$(echo "$FIRST_LINE" | cut -d ' ' -f 2)
if [ "$UNRELEASED_VERSION" == "${PREVIOUS_RELEASE_TAG#v}" ]; then
  echo "ERROR: top-of-file version ($UNRELEASED_VERSION) matches the previous release tag (${PREVIOUS_RELEASE_TAG})." >&2
  echo "Refusing to generate the changelog to avoid duplicating the released section." >&2
  exit 1
fi

PREVIOUS_CHANGELOG=$(sed -n -e "/# ${PREVIOUS_RELEASE_TAG#v}/,\$p" $__root/$CHANGELOG_FILE_NAME)

if [ -z "$PREVIOUS_CHANGELOG" ]
then
    echo "Unable to locate previous changelog contents."
    exit 1
fi 

CHANGELOG=$("$(go env GOPATH)"/bin/changelog-build -this-release $TARGET_SHA \
                      -last-release $PREVIOUS_RELEASE_SHA \
                      -git-dir $__root \
                      -entries-dir .changelog \
                      -changelog-template $__dir/changelog.tmpl \
                      -note-template $__dir/release-note.tmpl \
                      -storage-mode filesystem)
if [ -z "$CHANGELOG" ]
then
    echo "No changelog generated."
    exit 0
fi

rm -f $CHANGELOG_TMP_FILE_NAME

sed -n -e "1{/# /p;}" $__root/$CHANGELOG_FILE_NAME > $CHANGELOG_TMP_FILE_NAME
echo "$CHANGELOG" >> $CHANGELOG_TMP_FILE_NAME
echo >> $CHANGELOG_TMP_FILE_NAME
echo "$PREVIOUS_CHANGELOG" >> $CHANGELOG_TMP_FILE_NAME

cp $CHANGELOG_TMP_FILE_NAME $CHANGELOG_FILE_NAME

rm $CHANGELOG_TMP_FILE_NAME

echo "Successfully generated changelog."

exit 0
