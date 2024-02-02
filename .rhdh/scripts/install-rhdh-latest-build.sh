# code yoinked from https://gitlab.cee.redhat.com/rhidp/rhdh/-/blob/rhdh-1-rhel-9/build/scripts/installCatalogSourceFromIIB.sh

TMPDIR=/tmp
NAMESPACE="openshift-operators"
INSTALL_PLAN_APPROVAL="Automatic"
OLM_CHANNEL="fast"

# log into your OCP cluster before running this or you'll get null values for OCP vars!
OCP_VER="v$(oc version -o json | jq -r '.openshiftVersion' | sed -r -e "s#([0-9]+\.[0-9]+)\..+#\1#")"
OCP_ARCH="$(oc version -o json | jq -r '.serverVersion.platform' | sed -r -e "s#linux/##")"
if [[ $OCP_ARCH == "amd64" ]]; then OCP_ARCH="x86_64"; fi
# if logged in, this should return something like latest-v4.12-x86_64
UPSTREAM_IIB="quay.io/rhdh/iib:latest-${OCP_VER}-${OCP_ARCH}";

TO_INSTALL="rhdh"

# Add ImageContentSourcePolicy to let resolve references to images not on quay as if from quay.io
ICSP_URL="quay.io/rhdh/"
ICSP_URL_PRE=${ICSP_URL%%/*}
# echo "[DEBUG] ${ICSP_URL_PRE}, ${ICSP_URL_PRE//./-}, ${ICSP_URL}"
echo "apiVersion: operator.openshift.io/v1alpha1
kind: ImageContentSourcePolicy
metadata:
  name: ${ICSP_URL_PRE//./-}
spec:
  repositoryDigestMirrors:
  ## 1. add mappings for Developer Hub bundle, operator, hub
  - mirrors:
    - ${ICSP_URL}rhdh-operator-bundle
    source: registry.redhat.io/rhdh/rhdh-operator-bundle
  - mirrors:
    - ${ICSP_URL}rhdh-operator-bundle
    source: registry.stage.redhat.io/rhdh/rhdh-operator-bundle
  - mirrors:
    - ${ICSP_URL}rhdh-operator-bundle
    source: registry-proxy.engineering.redhat.com/rh-osbs/rhdh-rhdh-operator-bundle

  - mirrors:
    - ${ICSP_URL}rhdh-rhel9-operator
    source: registry.redhat.io/rhdh/rhdh-rhel9-operator
  - mirrors:
    - ${ICSP_URL}rhdh-rhel9-operator
    source: registry.stage.redhat.io/rhdh/rhdh-rhel9-operator
  - mirrors:
    - ${ICSP_URL}rhdh-rhel9-operator
    source: registry-proxy.engineering.redhat.com/rh-osbs/rhdh-rhdh-rhel9-operator

  - mirrors:
    - ${ICSP_URL}rhdh-hub-rhel9
    source: registry.redhat.io/rhdh/rhdh-hub-rhel9
  - mirrors:
    - ${ICSP_URL}rhdh-hub-rhel9
    source: registry.stage.redhat.io/rhdh/rhdh-hub-rhel9
  - mirrors:
    - ${ICSP_URL}rhdh-hub-rhel9
    source: registry-proxy.engineering.redhat.com/rh-osbs/rhdh-rhdh-hub-rhel9

  ## 2. general repo mappings
  - mirrors:
    - ${ICSP_URL_PRE}
    source: registry.redhat.io
  - mirrors:
    - ${ICSP_URL_PRE}
    source: registry.stage.redhat.io
  - mirrors:
    - ${ICSP_URL_PRE}
    source: registry-proxy.engineering.redhat.com

  ### now add mappings to resolve internal references
  - mirrors:
    - registry.redhat.io
    source: registry.stage.redhat.io
  - mirrors:
    - registry.stage.redhat.io
    source: registry-proxy.engineering.redhat.com
  - mirrors:
    - registry.redhat.io
    source: registry-proxy.engineering.redhat.com
" > $TMPDIR/ImageContentSourcePolicy_${ICSP_URL_PRE}.yml && oc apply -f $TMPDIR/ImageContentSourcePolicy_${ICSP_URL_PRE}.yml

echo "[INFO] Using iib $TO_INSTALL image $UPSTREAM_IIB"
IIB_IMAGE="${UPSTREAM_IIB}"
CATALOGSOURCE_NAME="${TO_INSTALL}-${OLM_CHANNEL}"

echo "apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: ${CATALOGSOURCE_NAME}
  namespace: $NAMESPACE
spec:
  sourceType: grpc
  image: ${IIB_IMAGE}
  publisher: IIB testing ${TO_INSTALL}
  displayName: IIB testing catalog ${TO_INSTALL}
" > $TMPDIR/CatalogSource.yml && oc apply -f $TMPDIR/CatalogSource.yml

# Create subscription for operator
echo "apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: $TO_INSTALL
  namespace: $NAMESPACE
spec:
  channel: $OLM_CHANNEL
  installPlanApproval: $INSTALL_PLAN_APPROVAL
  name: $TO_INSTALL
  source: ${CATALOGSOURCE_NAME}
  sourceNamespace: $NAMESPACE
" > $TMPDIR/Subscription.yml && oc apply -f $TMPDIR/Subscription.yml
