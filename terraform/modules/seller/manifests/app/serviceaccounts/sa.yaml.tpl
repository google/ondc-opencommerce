apiVersion: v1
kind: ServiceAccount
metadata:
  name: ${name} 
  namespace: ${namespace}
  annotations:
    iam.gke.io/gcp-service-account: ${service_account}