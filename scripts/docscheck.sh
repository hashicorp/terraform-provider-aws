#!/usr/bin/env bash

docs=$(ls website/docs/**/*.markdown)
error=false

for doc in $docs; do
  dirname=$(dirname "$doc")
  category=$(basename "$dirname")


  case "$category" in
    "guides")
      # Guides have no requirements
      continue
      ;;

    "d" | "r")
      # Resources and datasources require a subcategory
      grep "^subcategory: " "$doc" > /dev/null
      if [[ "$?" == "1" ]]; then
        echo "Doc is missing a subcategory: $doc"
        error=true
      fi
      ;;

    *)
      # Docs 
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
