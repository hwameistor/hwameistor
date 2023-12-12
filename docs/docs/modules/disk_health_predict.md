---
# Disk-Health-Prediction
---

Disk-Health-Prediction is one of the modules of HwameiStor. It is used to predict the health of storage cluster point disks. The current training accuracy can reach 85%. When deploying hwameister, automated deployment will not be performed. If users need to use it, please refer to the instructions for separate deployment. Note: The disk prediction function requires creating a pytorch job, and the estimated image size is 8 GB. Therefore, it is necessary to first evaluate the deployment node's image environment space size

![Disk-Health-Prediction](https://img-blog.csdnimg.cn/direct/3c951753b4354b4f8ac7ed18a611d734.png#pic_center)


## Usage



1. Deploy disk-health-prediction pod
   
   ```yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     labels:
       app: disk-health-prediction-dep
     name: disk-health-prediction-dep
     namespace: hwameistor
     
   spec:
     replicas: 1
     selector:
       matchLabels:
         app: disk-health-prediction-dep
         
     template:
       metadata:
         labels:
           app: disk-health-prediction-dep
       spec:
         serviceAccount: hwameistor-admin
         nodeName: poc-master3
         containers:
         - image: 10.6.118.138:5000/disk:v121912
           name: disk-health-prediction
   ```


2. Get predict-result-data 

```bash
kubectl get cm
NAME                         DATA   AGE
predict-result-poc-master1   22     7d19h
predict-result-poc-master2   22     7d19h
predict-result-poc-master3   22     7d19h
```
