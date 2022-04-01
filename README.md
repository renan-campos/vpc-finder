# VPC Finder
In an Openshift cluster, network policies are set to restrict where pods can send traffic.
If pods needs to reach the nodes through their node ips, a static network policy allowing traffic to be sent within the node network would suffice.
When the cluster is deployed in AWS, the node network is the VPC CIDR.
The entity within the cluster creating the network policy must have a way to determine the VPC CIDR.

### Problem:
We need a way to determine the VPC CIDR from within an Openshift cluster without using AWS accounts to query the underlying infrastructure.

### vpc_finder DOES:
- Determines the VPC CIDR that the cluster is on from within the cluster.
- Writes the VPC CIDR to a aws-data ConfigMap so that other pods in the namespace can utilize that data.

### vpc_finder DOES NOT:
- Use AWS accounts to query the underlying infrastructure/

### Here is how it works:
- A pod, the vpc_finder is deployed in the host network namespace.
- The vpc_finder uses the AWS Instance Metadata and User data endpoint (IMDSv1[https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html]) to query for the VPC CIDR
- The pod writes the VPC CIDR into a aws-data configmap

### To deploy:
This approach has been tested on an Openshift cluster running in AWS:

1. Create the RBAC permissions needed to run the vpc-finder pod:
```kubectl apply -f manifests/rbac.yaml```                                                                                                       
2. Create the vpc-finder pod
```kubectl apply -f manifests/vpc-finder.yaml```


--- TODO start ---
To test locally on a minikube instance:
1. Create the RBAC permissions needed to run the vpc-finder pod:
```kubectl apply -f manifests/rbac.yaml```                                                                                                       
2. Create an imdsv1-mock pod, that will mock the data that the vpc_finder's IMDSv1 calls return.
   Run the vpc-finder pod configured to use a different endpoint for the IMDSv1 calls.
```kubectl apply -f manifests/minikube-vpc-finder.yaml```
--- TODO end ---
