The `Content-Type` header must be set correctly, when using the polaris-scheduler or polaris-cluster-agent APIs, otherwise Gin will silently fail to parse the request body.

```sh
curl -XPOST localhost:8080/pods -H "Content-Type: application/json" -d '{
    "apiVersion": "v1",
    "kind": "Pod",
    "metadata": {
        "namespace": "default",
        "name": "myapp-01",
        "labels": {
            "name": "myapp-01"
        }
    },
    "spec": {
        "containers": [
            {
                "name": "myapp",
                "image": "gcr.io/google-containers/pause:3.2",
                "resources": {
                    "limits": {
                        "memory": "128Mi",
                        "cpu": "500m"
                    }
                }
            }
        ],
        "tolerations": [
            {
                "key": "fake-kubelet/provider",
                "operator": "Exists",
                "effect": "NoSchedule"
            }
        ]
    }
}'
```
