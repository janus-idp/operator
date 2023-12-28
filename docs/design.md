[WIP]
# Backstage Operator Design

The goal of Backstage Operator is to deploy Backstage workload to the Kubernetes namespace and keep this workload synced with the desired state defined by configuration. 

## Backstage Kubernetes Runtime

Backstage Kubernetes workload consists of set of Kubernetes resources (Runtime Objects).
Approximate set of Runtime Objects necessary for Backstage server on Kubernetes is shown on the diagram below:

![bso-runtime](https://github.com/gazarenkov/janus-idp-operator/assets/578124/9f72a5a5-fbdc-455c-9723-7fcb79734251)

The most important object you can see here is Backstage Pod created by Backstage Deployment. That is where we run 'backstage-backend' container with Backstage application inside.
This Backstage application is a web server which can be reached using Backstage Service.
Actually, those 2 are the core part of Backstage workload.  

## Configuration



### Layers

### Runtime configuration

![bs-conf](https://github.com/gazarenkov/janus-idp-operator/assets/578124/d56cbbb0-781c-43fc-8624-8832893fede3)

### Local Database

### Backstage App

![bs-pod](https://github.com/gazarenkov/janus-idp-operator/assets/578124/4ecf812b-28c7-4275-8c79-926b04fb94f8)

### Networking
