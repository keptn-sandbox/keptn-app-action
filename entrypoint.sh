#!/usr/bin/env bash

set -e
while getopts "i:o:v:b:" o; do
   case "${o}" in
      i)
        export input=${OPTARG}
        ;;
      o)
        export output=${OPTARG}
        ;;
      v)
        export version=${OPTARG}
        ;;
      b)
        export bump=${OPTARG}
        ;;
  esac
done

ARGS=""

if [ -n "$input" ]; then
  ARGS="$ARGS --input $input"
fi

if [ -n "$output" ]; then
  ARGS="$ARGS --output $output"
fi

if [ -n "$version" ]; then
  ARGS="$ARGS --version $version"
fi

if [ -n "$bump" ]; then
  ARGS="$ARGS --bump $bump"
fi

/keptn-config-generator ${ARGS}



