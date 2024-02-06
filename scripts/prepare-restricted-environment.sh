#!/bin/bash
#
# Copyright (c) 2024 Red Hat, Inc.
# This program and the accompanying materials are made
# available under the terms of the Eclipse Public License 2.0
# which is available at https://www.eclipse.org/legal/epl-2.0/
#
# SPDX-License-Identifier: EPL-2.0
#

# Fail on error
set -e

# INTERNAL_REGISTRY_URL, eg 'ec2-3-12-71-143.us-east-2.compute.amazonaws.com:5000' or 'default-route-openshift-image-registry.apps.ci-ln-x0yk982-72292.origin-ci-int-gce.dev.rhcloud.com'
# INTERNAL_REG_USERNAME, eg., dummy
# INTERNAL_REG_PASSWORD, eg., dummy

# podman login registry.redhat.io -u ${RRIO_USERNAME} -p ${RRIO_PASSWORD}
# podman login ${INTERNAL_REGISTRY_URL} -u ${INTERNAL_REG_USERNAME} -p ${INTERNAL_REG_PASSWORD} --tls-verify=false

# example usage:
# ./prepare-restricted-environment.sh \
#   --prod_operator_index "registry.redhat.io/redhat/redhat-operator-index:v4.14" \
#   --prod_operator_package_name "devspaces" \
#   --prod_operator_bundle_name "rhdh-operator" \
#   --prod_operator_version "v1.1.0" \
#   --my_registry "$INTERNAL_REGISTRY_URL"
while [ $# -gt 0 ]; do
  if [[ $1 == *"--"* ]]; then
    param="${1/--/}"
    declare "$param"="$2"
  fi
  shift
done

# Display commands
# set -x

# Operators
declare prod_operator_index="${prod_operator_index:?Must set --prod_operator_index: for OCP 4.12, use registry.redhat.io/redhat/redhat-operator-index:v4.12 or quay.io/rhdh/iib:latest-v4.14-x86_64}"
declare prod_operator_package_name="rhdh"
declare prod_operator_bundle_name="rhdh-operator"
declare prod_operator_version="${prod_operator_version:?Must set --prod_operator_version: for stable channel, use v1.1.0; for stable-1.1 channel, use v1.1.1}" # eg., v1.1.0 or v1.1.1

# Destination registry
declare my_registry="${my_registry:?Must set --my_registry: something like 'default-route-openshift-image-registry.apps.ci-ln-x0yk982-72292.origin-ci-int-gce.dev.rhcloud.com' for your cluster}"
declare my_operator_index_image_name_and_tag=${prod_operator_package_name}-index:${prod_operator_version}
declare my_operator_index="${my_registry}/${prod_operator_package_name}/${my_operator_index_image_name_and_tag}"

declare my_catalog=${prod_operator_package_name}-disconnected-install
declare k8s_resource_name=${my_catalog}

## from https://docs.openshift.com/container-platform/4.14/registry/securing-exposing-registry.html
#echo "[INFO] Expose the default registry and log in ..."
#oc patch configs.imageregistry.operator.openshift.io/cluster --patch '{"spec":{"defaultRoute":true}}' --type=merge
#HOST=$(oc get route default-route -n openshift-image-registry --template='{{ .spec.host }}')
#echo "[INFO] Default registry is $HOST"
#my_registry="$HOST"

# if able to run sudo and update your ca trust store
# oc get secret -n openshift-ingress  router-certs-default -o go-template='{{index .data "tls.crt"}}' | base64 -d | sudo tee /etc/pki/ca-trust/source/anchors/${HOST}.crt  > /dev/null
# sudo update-ca-trust enable
# sudo podman login -u kubeadmin -p "$(oc whoami -t)" "$HOST"

## else just login with tls disabled
#podman login -u kubeadmin -p "$(oc whoami -t)" --tls-verify=false "$HOST"

# Create local directory
mkdir -p "${my_catalog}/${prod_operator_package_name}"

echo "[INFO] Fetching metadata for the ${prod_operator_package_name} operator catalog channel, packages, and bundles."
opm render "${prod_operator_index}" \
  | jq "select \
    (\
      (.schema == \"olm.bundle\" and .name == \"${prod_operator_bundle_name}.${prod_operator_version}\") or \
      (.schema == \"olm.package\" and .name == \"${prod_operator_package_name}\") or \
      (.schema == \"olm.channel\" and .package == \"${prod_operator_package_name}\") \
    )" \
  | jq "select \
     (.schema == \"olm.channel\" and .package == \"${prod_operator_package_name}\").entries \
      |= [{name: \"${prod_operator_bundle_name}.${prod_operator_version}\"}]" \
  > "${my_catalog}/${prod_operator_package_name}/render.json"

echo "[DEBUG] Got $(cat "${my_catalog}/${prod_operator_package_name}/render.json" | wc -l) lines of JSON from the index!"
# echo "[DEBUG] Got this from the index:
# ======"
# cat "${my_catalog}/${prod_operator_package_name}/render.json"
# echo "======"

echo "[INFO] Creating the catalog dockerfile."
if [ -f "${my_catalog}.Dockerfile" ]; then
  rm -f "${my_catalog}.Dockerfile"
fi
opm generate dockerfile "./${my_catalog}"

echo "[INFO] Building the catalog image locally."
podman build -t "${my_operator_index}" -f "./${my_catalog}.Dockerfile" --no-cache .

echo "[INFO] Disabling the default Red Hat Ecosystem Catalog."
oc patch OperatorHub cluster --type json \
    --patch '[{"op": "add", "path": "/spec/disableAllDefaultSources", "value": true}]'

echo "[INFO] Deploying your catalog image to the $my_operator_index registry."
# See: https://docs.openshift.com/container-platform/latest/installing/disconnected_install/installing-mirroring-installation-images.html#olm-mirroring-catalog_installing-mirroring-installation-images
### TODO fix this step or switch to oc adm catalog mirror?
# https://docs.openshift.com/container-platform/4.14/installing/disconnected_install/installing-mirroring-installation-images.html#installation-images-samples-disconnected-mirroring-assist_installing-mirroring-installation-images ?
# FATA[0010] writing manifest: uploading manifest v1.1.0 to default-route-openshift-image-registry.apps.ci-ln-x0yk982-72292.origin-ci-int-gce.dev.rhcloud.com/rhdh/rhdh-index: denied 
skopeo copy --src-tls-verify=false --dest-tls-verify=false --all "containers-storage:$my_operator_index" "docker://$my_operator_index"

echo "[INFO] Removing index image from mappings.txt to prepare mirroring."
oc adm catalog mirror "$my_operator_index" "$my_registry" --insecure --manifests-only | tee catalog_mirror.log
MANIFESTS_FOLDER=$(sed -n -e 's/^wrote mirroring manifests to \(.*\)$/\1/p' catalog_mirror.log |xargs) # The xargs here is to trim whitespaces
sed -i -e "/${my_operator_index_image_name_and_tag}/d" "${MANIFESTS_FOLDER}/mapping.txt"
cat "${MANIFESTS_FOLDER}/mapping.txt"

echo "[INFO] Mirroring related images to the $my_registry registry."
# oc image mirror --insecure=true -f "${MANIFESTS_FOLDER}/mapping.txt"
while IFS= read -r line
do
  public_image=$(echo "${line}" | cut -d '=' -f1)
  if [[ "$public_image" == registry.redhat.io/rhdh/* ]]; then
    # CI Builds not public yet
    public_image=${public_image/registry.redhat.io/quay.io}
  fi
  private_image=$(echo "${line}" | cut -d '=' -f2)
  echo "[INFO] Mirroring ${public_image}"
  skopeo copy --dest-tls-verify=false --preserve-digests --all "docker://$public_image" "docker://$private_image"
done < "${MANIFESTS_FOLDER}/mapping.txt"

echo "[INFO] Creating CatalogSource and ImageContentSourcePolicy"
# shellcheck disable=SC2002
cat "${MANIFESTS_FOLDER}/catalogSource.yaml" | sed 's|name: .*|name: '${k8s_resource_name}'|' | oc apply -f -
# shellcheck disable=SC2002
cat "${MANIFESTS_FOLDER}/imageContentSourcePolicy.yaml" | sed 's|name: .*|name: '${k8s_resource_name}'|' | oc apply -f -

echo "[INFO] Catalog $my_operator_index deployed to the $my_registry registry."
