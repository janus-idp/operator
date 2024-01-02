# Backstage Operator Design [WIP]

The goal of Backstage Operator is to deploy Backstage workload to the Kubernetes namespace and keep this workload synced with the desired state defined by configuration. 

## Backstage Kubernetes Runtime

Backstage Kubernetes workload consists of set of Kubernetes resources (Runtime Objects).
Approximate set of Runtime Objects necessary for Backstage server on Kubernetes is shown on the diagram below:

![bso-runtime](https://github.com/gazarenkov/janus-idp-operator/assets/578124/9f72a5a5-fbdc-455c-9723-7fcb79734251)

The most important object is Backstage Pod created by Backstage Deployment. That is where we run 'backstage-backend' container with Backstage application inside.
This Backstage application is a web server which can be reached using Backstage Service.
Actually, those 2 are the core part of Backstage workload. 

Backstage application uses SQL database as a data storage and it is possible to install PostgreSQL DB on the same namespace as Backstage instance.
It brings PostgreSQL StatefulSet/Pod, Service to connect to Backstage and PV/PVC to store the data.

For providing external access to Backstage server it is possible, depending on underlying infrastructure, to use Openshift Route or 
K8s Ingress on top of Backstage Service. (As for v 0.01-0.02 only Route configuration is supported by the Operator out-of-the-box).

And, finally, Backstage Operator supports all the [Backstage configuration](https://backstage.io/docs/conf/writing) options, which can be provided by creating dedicated 
ConfigMaps and Secrets and contributing them to the Backstage Pod as mounted volumes or environment variables (see [Configuration](configuration.md) guide for details)  

## Configuration

### Configuration layers

Backstage Operator designed to be flexible in terms of what eactly it is going to deploy to make a workload.
At the same time we're trying to make a configuration as simple as possible for the case when user/admin just want to try it 
or make a personal or small group Backstage instance.

Taking it into account lead to create 3 layers configuration.

![bs-conf](https://github.com/gazarenkov/janus-idp-operator/assets/578124/d56cbbb0-781c-43fc-8624-8832893fede3)

As shown in the picture above:

- There is an Operator (Cluster) level Default Configuration implemented as a ConfigMap inside Backstage system namespace
  (where Backstage controller is launched). It allows to choose some optimal for most cases configuration which will be applied 
if there are no other config to override (i.e. Backstage CR is empty). 
- Another layer overriding default is instance (Backstage CR) scoped, implemented as a ConfigMap which
has the same as default structure but inside Backstage instance's namespace. The name of theis ConfigMap 
is specified on Backstage.Spec.RawConfig field. It offers very flexible way to configure certain Backstage instance  
- And finally, there are set of fields on Backstage.Spec to override configuration made on level 1 and 2.
It offers simple configuration of some parameters. So, user is not required to understand the
overall structure of Backstage runtime object and is able to simply configure "the most important" parameters.
  (see [configuration](configuration.md) for more details)

### Backstage Application

Backstage Application comes with advanced configuration features.

As it stated in [Backstage configuration](https://backstage.io/docs/conf/writing) document user can define and overload multiple _app-config.yaml_
files and flexible configuring it with inclusing and environment variables.
Backstage Operator supports this flexibility allowing to define these configurations components in all the configuration levels
(default, raw and CR)

![bs-pod](https://github.com/gazarenkov/janus-idp-operator/assets/578124/4ecf812b-28c7-4275-8c79-926b04fb94f8)

### Networking
TODO
