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

output_path=${xpl_path}/_output_
rm -rf ${output_path}
mkdir -p ${output_path}

cd ${xpl_path}/../model/api

# generate python rest server
# java -jar ${codegen} generate \
#   -i safescale.swagger.json \
#   -g python-flask \
#   -o ${output_path}
  
# generate python rest server
java -jar ${codegen} generate \
  -i openapi.yaml \
  -g python-flask \
  -o ${output_path}

mkdir -p ${output_path}html_
java -jar ${codegen} generate \
  -i safescale.swagger.json \
  -g html \
  -o ${output_path}html_

# copy server inside backend
if [ ! -d ${project_path}/controllers ]
then
	cp -vr ${output_path}/openapi_server/controllers ${project_path}
fi
cp -f ${output_path}/openapi_server/{encoder.py,util.py,typing_utils.py} ${project_path}
rsync --delete -rv ${output_path}/openapi_server/models/ ${project_path}/models/
rsync --delete -rv ${output_path}/openapi_server/openapi/ ${project_path}/openapi/
# patch files
find ${project_path} -name "*.py" -exec sed -i -e "s# openapi_server# rest#g" {} \;
sed -i -e "s#openapi_server#rest#g" ${project_path}/openapi/openapi.yaml
# add an init for packaging
touch ${project_path}/openapi/__init__.py

#rm -rf ${output_path}
