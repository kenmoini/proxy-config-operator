

```bash=
# Initialize the operator
operator-sdk init --domain k8s.kemo.dev --repo github.com/kenmoini/proxy-config-operator

# Create the ProxyConfig Controller
operator-sdk create api --group proxy --version v1alpha1 --kind ProxyConfig --resource --controller
```
