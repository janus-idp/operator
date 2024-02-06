#!/bin/bash

# Fail on error
set -e

# example usage:
# ./deploy-mirror-registry.sh \
#   --namespace "airgap-registry" \
#   --image "registry:2" \
#   --storage_class "" \
#   --storage_capacity "100Gi" \
#   --expose "true" \
#   --username "$USER" \
#   --password "$PASS"

while [ $# -gt 0 ]; do
  if [[ $1 == *"--"* ]]; then
    param="${1/--/}"
    declare "$param"="$2"
  fi
  shift
done

declare namespace="${namespace:-"airgap-helper-ns"}"
declare image="${image:-"registry:2"}"
declare storage_capacity=${storage_capacity:-"100Gi"}
declare username="${username:?Must set --username}"
declare password="${password:?Must set --username}"

if ! oc get namespace "${namespace}" > /dev/null 2>&1; then
  oc create namespace "${namespace}"
fi

registry_htpasswd=$(htpasswd -Bbn "${username}" "${password}")
cat <<EOF | oc apply -f -
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: airgap-registry-auth
  namespace: "${namespace}"
  labels:
    app: airgap-registry
stringData:
  htpasswd: "${registry_htpasswd}"
EOF

if [ -z "$storage_class" ]; then
  # use default storage class
  storage_class=$(oc get storageclasses -o=jsonpath='{.items[?(@.metadata.annotations.storageclass\.kubernetes\.io/is-default-class=="true")].metadata.name}')
fi
cat <<EOF | oc apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: airgap-registry-storage
  namespace: "${namespace}"
spec:
  resources:
    requests:
      storage: "${storage_capacity}"
  storageClassName: ${storage_class}
  accessModes:
    - ReadWriteOnce
EOF

cat <<EOF | oc apply -f -
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: airgap-registry
  namespace: "${namespace}"
  labels:
    app: airgap-registry
spec:
  replicas: 1
  selector:
    matchLabels:
      app: airgap-registry
  template:
    metadata:
      labels:
        app: airgap-registry
    spec:
      # -----------------------------------------------------------------------
      containers:
        - image: "${image}"
          name: airgap-registry
          imagePullPolicy: IfNotPresent
          env:
            - name: REGISTRY_AUTH
              value: "htpasswd"
            - name: REGISTRY_AUTH_HTPASSWD_REALM
              value: "RHDH Private Registry"
            - name: REGISTRY_AUTH_HTPASSWD_PATH
              value: "/auth/htpasswd"
            - name: REGISTRY_STORAGE_DELETE_ENABLED
              value: "true"
#            - name: REGISTRY_HTTP_TLS_CERTIFICATE
#              value: "/certs/tls.crt"
#            - name: REGISTRY_HTTP_TLS_KEY
#              value: "/certs/tls.key"
          ports:
            - containerPort: 5000
          volumeMounts:
            - name: registry-vol
              mountPath: /var/lib/registry
#            - name: tls-vol
#              mountPath: /certs
#              readOnly: true
            - name: auth-vol
              mountPath: "/auth"
              readOnly: true
      # -----------------------------------------------------------------------
      volumes:
        - name: registry-vol
          persistentVolumeClaim:
            claimName: airgap-registry-storage
#        - name: tls-vol
#          secret:
#            secretName: airgap-registry-certificate
        - name: auth-vol
          secret:
            secretName: airgap-registry-auth
EOF

cat <<EOF | oc apply -f -
apiVersion: v1
kind: Service
metadata:
  name: airgap-registry
  namespace: "${namespace}"
  labels:
    app: airgap-registry
spec:
  type: ClusterIP
  ports:
    - port: 5000
      protocol: TCP
      targetPort: 5000
  selector:
    app: airgap-registry
EOF

oc -n "${namespace}" create route edge --service=airgap-registry --insecure-policy=Redirect --dry-run=client -o yaml | oc -n "${namespace}" apply -f -

echo "Registry exposed at: $(oc get route airgap-registry -n "${namespace}" --template='{{ .spec.host }}')"

# TODO
#echo oc edit image.config.openshift.io/cluster and add this registry as insecure if needed
#echo oc get secret/pull-secret -n openshift-config --template='{{index .data ".dockerconfigjson" | base64decode}}' > /tmp/my-global-pull-secret.yaml
#echo oc registry login --registry="<REG>" --auth-basic="<USER>:<PASS>" --to=/tmp/my-global-pull-secret.yaml
#echo oc set data secret/pull-secret -n openshift-config --from-file=.dockerconfigjson=/tmp/my-global-pull-secret.yaml