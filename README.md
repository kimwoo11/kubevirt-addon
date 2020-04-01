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
