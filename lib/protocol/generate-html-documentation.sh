#!/bin/bash

set -e

xpl_path=$(cd -P -- "$(dirname -- "$0")" && pwd -P)
project_path=${xpl_path}
cd $xpl_path

version=4.3.1
codegen_url=https://repo1.maven.org/maven2/org/openapitools/openapi-generator-cli
codegen=/tmp/openapi-generator-cli-${version}.jar

last_version=$(curl -s ${codegen_url}/ | grep -e "title=\"[0-9]*" | grep -e "title=\"[0-9]" | sed "s#<.*>\([0-9]*.*\)/</.*#\1#g" | tail -n1)
if [ "$version" != "$last_version" ]
then
	echo "New version $last_version available"
fi

if [ ! -f ${codegen} ]
then
# 	openapi
	wget ${codegen_url}/${version}/openapi-generator-cli-${version}.jar  -O /tmp/openapi-generator-cli-${version}.jar
fi

output_path=${xpl_path}/html
rm -rf ${output_path}
mkdir -p ${output_path}

java -jar ${codegen} generate \
  -i safescale.swagger.json \
  -g html \
  -o ${output_path}
