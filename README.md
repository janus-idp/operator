# Backstage Operator

## Context
[Backstage](https://backstage.io) is an open platform for building developer portals. Powered by a centralized software catalog, Backstage restores order to your microservices and infrastructure and enables your product teams to ship high-quality code quickly — without compromising autonomy.

Backstage unifies all your infrastructure tooling, services, and documentation to create a streamlined development environment from end to end.

[Janus-IDP](https://janus-idp.io/) is a Red Hat sponsored community for building developer portals, built on Backstage. The set of Backstage plugins hand picked or created by the Janus IDP team include (but not limited to) ArgoCD, GitHub Issues, Keycloak, Kubernetes, OCM, Tekton, and Topology plugins. 

The purpose of [Janus Showcase](https://github.com/janus-idp/backstage-showcase) is to showcase the value of the plugins created is a part of Janus-IDP initiative and to demonstrate the power of an internal developer portal using Backstage as the solution.

## The Goal
The Goal of Backstage Operator project is creating Kubernetes Operator for configuring, installing and synchronizing Backstage instance on Kubernetes/OpenShift simple and flexible. 

The Operator should be flexible enough to install any correct Backstage instance (specific Kubernetes deployment with accompanying resources, see [Configuration](#)) but primary target is Janus-IDP Showcase on OpenShift, specifically supporting [dynamic-plugins](), so this kind of configuration may contain some specific objects (such as InitContainer(s) and dedicated ConfigMaps) to make this feature work.  

The Operator should provide clear and flexible configuration options to satisfy wide range of expectations: from "default (no)configuration for quick start" to "higly customized configuration for production".

Make sure namespace is created.

Local Database (PostgreSQL) is created by default, to disable


## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [Minikube](https://minikube.sigs.k8s.io/docs/) or [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

To test how it works locally on minikube:

0. Make sure your minikube instance is up and running and get your copy of Operator from GitHub: 
```sh
git clone https://github.com/janus-idp/operator
```
1. Deploy Operator on the cluster:
```sh
make deploy
```
Check if the Operator is up and running - TODO
2. Create Backstage Custom resource on some namespace
```sh
kubectl -n <your-namespace> apply -f examples/bs1.yaml
```
3. Tunnel Backstage Service and get URL for access Backstage
```sh
minikube service -n <your-namespace> backstage --url
Output:
>http://127.0.0.1:<port>
```
4. Access your Backstage instance using this URL.   

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

