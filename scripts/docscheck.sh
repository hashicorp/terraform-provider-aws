#!/usr/bin/env bash

docs=$(ls website/docs/**/*.markdown)
error=false

for doc in $docs; do
  dirname=$(dirname "$doc")
  category=$(basename "$dirname")


  case "$category" in
    "guides")
      # Guides require a page_title
      grep "^page_title: " "$doc" > /dev/null
      if [[ "$?" == "1" ]]; then
        echo "Guide is missing a page_title: $doc"
        error=true
      fi
      ;;

    "d" | "r")
      # Resources and datasources require a subcategory
      grep "^subcategory: " "$doc" > /dev/null
      if [[ "$?" == "1" ]]; then
        echo "Doc is missing a subcategory: $doc"
        error=true
      fi

      # sidebar_current is deprecated
      # https://github.com/hashicorp/terraform-website#yaml-frontmatter
      grep "^sidebar_current: " "$doc" > /dev/null
      if [[ "$?" == "0" ]]; then
        echo "Doc contains a deprecated sidebar_current: $doc"
      fi
      ;;

    *)
      error=true
      echo "Unknown category \"$category\". " \
        "Docs can only exist in r/, d/, or guides/ folders."
      ;;
  esac
done

if $error; then
  exit 1
fi

exit 0
