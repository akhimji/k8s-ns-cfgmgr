# k8s-cfgmgr
Simple (non informer) based operator to poll git repo for yaml files of Kind (Deployment and Service) and maintain state in a namespace. 


Usage:
```
git clone repo
cd deploy/
edit go-op-rc.yaml

# edit ENV var

          env:
          - name: URL
            value: "https://github.com/alyarctiq/k8s-ns-cfgmgr.git"
          - name: FOLDER
            value: "/yamls"
          - name: NAMESPACE
            value: "hipster"    

(leave the same for testing)

# Create Namesapce
$ kubectl create ns hipster
namespace/hipster created

# Apply RBAC
$ kubectl apply -f rbac-r.yaml -n hipster
serviceaccount/k8s-ns-cfgmgr created
role.rbac.authorization.k8s.io/k8s-ns-cfgmgr created
rolebinding.rbac.authorization.k8s.io/k8s-ns-cfgmgr created

# Deploy Op
$ kubectl apply -f go-op-rc.yaml  -n hipster
replicationcontroller/k8s-ns-cfgmgr created

$ kubectl get pods -n hipster
NAME                  READY   STATUS              RESTARTS   AGE
k8s-ns-cfgmgr-ppcfk   0/1     ContainerCreating   0          3s
```

Validate
```
# Tail Logs
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
```

Reconsiling Namespace 
```
2020/04/11 14:21:54 Repairing Missing Deployment: emailservice
Created deployment "emailservice".
2020/04/11 14:21:54 Repairing Missing Deployment: checkoutservice
Created deployment "checkoutservice".
2020/04/11 14:21:55 Repairing Missing Deployment: recommendationservice
Created deployment "recommendationservice".
2020/04/11 14:21:55 Repairing Missing Deployment: frontend
Created deployment "frontend".
2020/04/11 14:21:55 Repairing Missing Deployment: paymentservice
Created deployment "paymentservice".
2020/04/11 14:21:55 Repairing Missing Deployment: productcatalogservice
Created deployment "productcatalogservice".
2020/04/11 14:21:55 Repairing Missing Deployment: cartservice
Created deployment "cartservice".
2020/04/11 14:21:55 Repairing Missing Deployment: loadgenerator
Created deployment "loadgenerator".
2020/04/11 14:21:55 Repairing Missing Deployment: currencyservice
Created deployment "currencyservice".
2020/04/11 14:21:55 Repairing Missing Deployment: shippingservice
Created deployment "shippingservice".
2020/04/11 14:21:56 Repairing Missing Deployment: redis-cart
Created deployment "redis-cart".
2020/04/11 14:21:56 Repairing Missing Deployment: adservice
Created deployment "adservice".
2020/04/11 14:21:56 Repairing Missing Service: emailservice
Created Service "emailservice".
2020/04/11 14:21:57 Repairing Missing Service: checkoutservice
Created Service "checkoutservice".
.....

$ kubectl get pods -n hipster
NAME                                     READY   STATUS              RESTARTS   AGE
adservice-5897f58b66-wgq74               0/1     Pending             0          2s
cartservice-5d679b9449-z75s4             0/1     ContainerCreating   0          4s
checkoutservice-6f56ff8674-9cjrd         0/1     ContainerCreating   0          4s
currencyservice-f58f7f5d4-gvxsh          0/1     ContainerCreating   0          3s
emailservice-d66dc8fbb-mt4tn             0/1     ContainerCreating   0          4s
frontend-5bd68756d4-h4gwn                0/1     ContainerCreating   0          4s
k8s-ns-cfgmgr-jbhjh                      1/1     Running             0          12s
loadgenerator-6655f746fb-jc744           0/1     ContainerCreating   0          4s
paymentservice-794c8b9ccd-n6dnm          0/1     ContainerCreating   0          4s
productcatalogservice-6446f67666-29989   0/1     ContainerCreating   0          4s
recommendationservice-654cb4cdb4-fz68x   0/1     ContainerCreating   0          4s
redis-cart-65bf66b8fd-wmh47              0/1     ContainerCreating   0          2s
shippingservice-849db7fbb-s5r86          0/1     ContainerCreating   0          3s
```

Reconciles when a change is made to the repo. (eg yaml file is updated/deleted)
```
...
Delete Dep From Cluster: adservice
Deleted deployment.
Delete Dep From Cluster: cartservice
Deleted deployment.
Delete Dep From Cluster: currencyservice
Deleted deployment.
Delete Dep From Cluster: loadgenerator
Deleted deployment.
Delete Dep From Cluster: redis-cart
Deleted deployment.
Delete Dep From Cluster: shippingservice
Deleted deployment.
Deleting Svc: adservice
Delete Svc From Cluster: adservice
Deleting Svc: cartservice
Delete Svc From Cluster: cartservice
Deleting Svc: currencyservice
Delete Svc From Cluster: currencyservice
Deleting Svc: redis-cart
Delete Svc From Cluster: redis-cart
Deleting Svc: shippingservice
Delete Svc From Cluster: shippingservice
```


