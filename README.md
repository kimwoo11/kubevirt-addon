# kubevirt-addon

## Overview
The kubevirt-addon operator helps create Services, Routes, and Endpoints to expose services running on VMs, and create Endpoints to be aggregated to the MCM Hub cluster. 

## Deploy
1. Clone this repository
```bash
mkdir -p <project-directory>
cd <project-directory>
git clone https://github.ibm.com/steve-kim-ibm/kubevirt-addon.git
cd kubevirt-addon
```

2. Deploy the CRD and the operator
```bash
kubectl apply -f deploy/crds/app.ibm.com_kubevirtaddons_crd.yaml
kubectl apply -f deploy
```

3. Check that the operator is running
```bash
% kubectl get deploy 
NAME                        READY   UP-TO-DATE   AVAILABLE   AGE
kubevirt-addon              1/1     1            1           16s
```

## Guestbook Example
This example will go through the deployment of Kubernetes' (Guestbook Application)[https://kubernetes.io/docs/tutorials/stateless-application/guestbook/] deployed through KubeVirt VMs and the KubeVirt-Addon operator.

0. Deploy the frontend, redis-slave, and redis-master VMs found under [examples/guestbook/kubevirt-vms/](./examples/guestbook/kubevirt-vms)

```bash
kubectl apply -f ./examples/guestbook/kubevirt-vms/frontend.yaml
kubectl apply -f ./examples/guestbook/kubevirt-vms/redismst.yaml
kubectl apply -f ./examples/guestbook/kubevirt-vms/redisslv.yaml
```

1. Deploy the kubevirt-addon resources for each VM. These can be found under [examples/guestbook](./examples/guestbook/)

```bash
kubectl apply -f ./examples/guestbook/01-redis-master.yaml
kubectl apply -f ./examples/guestbook/02-redis-slave.yaml
kubectl apply -f ./examples/guestbook/03-frontend.yaml
```

2. Ensure that the intended services, routes, and endpoints have been deployed successfully

```bash
% kubectl get svc,routes,endpoints -l app=guestbook

# TODO: display output
```

3. Go to the URL defined in the endpoint resource: (http://guestbook.apps.folie.os.fyre.ibm.com)
```bash
% kubectl describe endpoints -l app=guestbook

# TODO: display output 
```
