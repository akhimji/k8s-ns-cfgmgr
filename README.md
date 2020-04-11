# k8s-cfgmgr
Simple (non informer) based operator to poll git repo for yaml files of Kind (Deployment and Service) and maintain state in a namespace. 


Usage:
```
$ kubectl apply -f rbac-r.yaml -n hipster
serviceaccount/k8s-ns-cfgmgr created
role.rbac.authorization.k8s.io/k8s-ns-cfgmgr created
rolebinding.rbac.authorization.k8s.io/k8s-ns-cfgmgr created

$ kubectl apply -f go-op-rc.yaml  -n hipster
replicationcontroller/k8s-ns-cfgmgr created
```

Validate

```
$ kubectl logs -f k8s-ns-cfgmgr-8vvkc  -n hipster
2020/04/11 14:26:07 OS ENV NS:  hipster
2020/04/11 14:26:07 OS ENV URL:  https://github.com/alyarctiq/k8s-ns-cfgmgr.git
2020/04/11 14:26:07 Cloning Git Repo

git pull https://github.com/alyarctiq/k8s-ns-cfgmgr.git /tmp/repo --recursive

2020/04/11 14:26:08 Search Path: Found /tmp/repo/yamls
2020/04/11 14:26:08 Starting Watch Loop...
2020/04/11 14:26:08 Loading Yaml Files: /tmp/repo/yamls/hipster1of2.yaml
2020/04/11 14:26:08 Loading Yaml Files: /tmp/repo/yamls/hipster2of2.yaml
2020/04/11 14:26:08 Loading Master Files:  /tmp/master.yaml


2020/04/11 14:26:08 Repairing Missing Deployment: emailservice
Created deployment "emailservice".
2020/04/11 14:26:08 Repairing Missing Deployment: checkoutservice
Created deployment "checkoutservice".
2020/04/11 14:26:08 Repairing Missing Deployment: recommendationservice
Created deployment "recommendationservice".
2020/04/11 14:26:09 Repairing Missing Deployment: frontend
Created deployment "frontend".
.....

```


