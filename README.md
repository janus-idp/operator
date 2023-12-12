# Backstage Operator

## The Goal
The Goal of [Backstage](https://backstage.io) Operator project is creating Kubernetes Operator for configuring, installing and synchronizing Backstage instance on Kubernetes/OpenShift. 
Primary target is supporting Red Hat's assemblies of Backstage (specifically supporting [dynamic-plugins]()) on OpenShift, such as [Janus-IDP](https://janus-idp.io/) and RHDH but is flexible enough to install any correct Backstage instance (specific Kubernetes deployment with accompanying resources, see [Configuration](#))
The Operator should provide clear and flexible configuration options to satisfy wide range of expectations: from "default (no)configuration for quick start" to "higly customized configuration for production".
[More documentation...](#More documentation)

## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [Minikube](https://minikube.sigs.k8s.io/docs/) or [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

To test it on minikube from the source code:

You need ready to use **kubectl** and **minikube** [tools](https://kubernetes.io/docs/tasks/tools/) as a prerequisite.

1.  Get your copy of Operator from GitHub: 
```sh
git clone https://github.com/janus-idp/operator
```
2. Deploy Operator on the minikube cluster:
```sh
cd <your-janus-idp-operator-project-dir>
make deploy
```
you can check if the Operator pod is up by running 
```sh
kubectl get pods -n backstage-system
It should be something like:
NAME                                           READY   STATUS    RESTARTS   AGE
backstage-controller-manager-cfc44bdfd-xzk8g   2/2     Running   0          32s
```
3. Create Backstage Custom resource on some namespace (make sure this namespace exists)
```sh
kubectl -n <your-namespace> apply -f examples/bs1.yaml
```
you can check if the Operator pod is up by running
```sh
kubectl get pods -n <your-namespace>
It should be something like:
NAME                         READY   STATUS    RESTARTS   AGE
backstage-85fc4657b5-lqk6r   1/1     Running   0          78s
backstage-psql-bs1-0         1/1     Running   0          79s

```
4. Tunnel Backstage Service and get URL for access Backstage
```sh
minikube service -n <your-namespace> backstage --url
Output:
>http://127.0.0.1:53245
```
5. Access your Backstage instance in your browser using this URL. 

## More documentation

[Openshift deployment](docs/openshift.md)
[Configuration](docs/configuration.md)
[Developer Guide](docs/developer.md)
[Operator Design](docs/developer.md)


## License

Copyright 2023 Red Hat Inc..

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

