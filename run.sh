kubectl delete secret op-portal
kubectl delete configmap op-portal
kubectl delete deployment -l app=op-portal

go build cmd/manager/main.go && ./main
