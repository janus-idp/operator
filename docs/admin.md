# Administrator Guide

## Backstage Operator configuration

### Context

As it is described in Design doc (TODO), Backstage CR's desired state is defined using layered configuration approach, which means:
- By default each newly created Backstage CR uses Operator scope Default Configuration
- Which can be fully or partially overriden for particular CR instance using ConfigMap with the name pointed in BackstageCR.spec.RawConfig 
- Which in turn can be customized by other BackstageCR.spec fields (see Backstage API doc)

Cluster Administrator may want to customize Default Configuration due to internal preferences/limitations, for example:
- Preferences/restrictions for Backstage and|or PostgreSQL images due to  Airgapped environment.
- Existed Taints and tolerations policy, so Backstage Pods have to be configured with certain tolerations restrictions.
- ...

Default Configuration is implemented as a ConfigMap called *backstage-default-config*, deployed on *backstage-system* namespace and mounted to Backstage controller container as a */default-config* directory.
This config map contains the set of keys/values which maps to file names/contents in the */default-config*.
These files contain yaml manifests of objects used by Backstage controller as an initial desired state of Backstage CR according to Backstage Operator data model.
(TODO: link to the diagram here) 

Mapping of configMap keys (yaml files) to runtime objects (NOTE: for the time (Dec 20'23) it is a subject of change):
| Key/File name         | k8s/OCP Kind        | Mandatory*    |Notes                                      |
| ----------------------|:-------------------:| --------------|------------------------------------------:|
| deployment.yaml       | appsv1.Deployment   | Yes           | Backstage deployment |
| service.yaml          | corev1.Service      | Yes           | Backstage Service |
| db-statefulset.yaml   | appsv1.Statefulset  | For DB enabled| PostgreSQL statefulSet    |    
| db-service.yaml       | corev1.Service      | For DB enabled| PostgreSQL Service   |
| db-secret.yaml        | corev1.Secret       | For DB enabled| Secret to connect Backstage to PSQL   |
| route.yaml            | openshift.Route     | No (for OCP)  | Route exposing Backstage service    |
| app-config.yaml       | corev1.ConfigMap    | No            | Backstage app-config.yaml    |
| configmap-files.yaml  | corev1.ConfigMap    | No            | Backstage config file inclusions from configMap   |
| configmap-envs.yaml   | corev1.ConfigMap    | No            | Backstage env variables from configMap    |
| secret-files.yaml     | corev1.Secret       | No            | Backstage config file inclusions from Secret   |
| secret-envs.yaml      | corev1.Secret       | No            | Backstage env variables from Secret    |
| dynamic-plugins.yaml  | corev1.ConfigMap    | No            | dynamc-plugins config *    |


NOTES: 
 - Mandatory means it is needed to be present in either (or both) Default and CR Raw Configuration.
 - dynamic-plugins.yaml is a fragment of app-config.yaml provided with RHDH/Janus-IDP. The reason it mentioned separately is specific way it provided (via mounting to dedicated initContainer)  

### Operator Bundle configuration 

With Backstage Operator's Makefile you can generate bundle descriptor using *make bundle* command

Along with CSV manifest it generates default-config ConfigMap manifest, which can be modified and applied to Backstage Operator
TODO: is kubectl sufficient or there are some other tools?

### Kustomize deploy configuration

Make sure use the current context in your kubeconfig file is pointed to correct place, change necessary part of your config/manager/default-config or just replace some of the file(s) with yours and run
``
make deploy
``

### Direct ConfigMap configuration

You can change default configuration by directly changing the default-config ConfigMap with kubectl like:

 - retrieve the current `default-config` from the cluster

``
kubectl get -n backstage-system configmap default-config > my-config.yaml
``

- modify the file in your editor of choice

- apply the updated configuration to your cluster

``
  kubectl apply -n backstage-system -f my-config.yaml
``

It has to be re-applied to the controller's container after some time.


### Use Cases

#### Airgapped environment

Creating Backtage CR, the Operator will try to create Backstage Pod, deploying:
- Backstage Container from the image, configured in *(deployment.yaml).spec.template.spec.Containers[].image*
- Init Container (applied for RHDH/Janus-IDP configuration, usually the same as Backstage Container)
- 
Also, if Backstage CR configured as EnabledLocalDb it will create Database (PGSQL) Container, configured in *(db-deployment.yaml).spec.template.spec.Containers[].image*

By default Backstage Operator is configured to use publicly available images, which is not acceptable for Airgapped environment.
