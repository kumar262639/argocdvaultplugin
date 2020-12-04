# argocd-vault-plugin
![Pipeline](https://github.com/IBM/argocd-vault-plugin/workflows/Pipeline/badge.svg) ![Code Scanning](https://github.com/IBM/argocd-vault-plugin/workflows/Code%20Scanning/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/IBM/argocd-vault-plugin)](https://goreportcard.com/report/github.com/IBM/argocd-vault-plugin) [![codecov](https://codecov.io/gh/IBM/argocd-vault-plugin/branch/main/graph/badge.svg?token=p8XJMcip6l)](https://codecov.io/gh/IBM/argocd-vault-plugin)

<img src="https://github.com/IBM/argocd-vault-plugin/raw/main/assets/argo_vault_logo.png" width="300">

An ArgoCD plugin to retrieve secrets from Hashicorp Vault and inject them into Kubernetes secrets

## Why use this plugin?
This plugin is aimed at helping to solve the issue of secret management with GitOps and ArgoCD. We wanted to find a simple way to utilize Vault without having to rely on an operator or custom resource definition. This plugin can be used not just for secrets but also for deployments, configMaps or any other Kubernetes resource.

## How it works
The argocd-vault-plugin works by taking a directory of yaml files that have been templated out using the pattern of `<thing-to-fill-in>` where you would want a value from Vault to go. The inside of the `<>` would be the actual key in vault.

An annotation is used to specify exactly where the plugin should look for the vault values. The annotation needs to be in the format `path: "path/to/vault"`.

For example, if you have a secret with the key `password` that you would want to pull from vault, you might have a yaml that looks something like the below code. In this yaml, the plugin will pull the value of `path/to/vault/password-vault-key` and inject it into the secret yaml.

```
kind: Secret
apiVersion: v1
metadata:
  name: example-secret
  annotations:
    avp_path: "path/to/secret"
type: Opaque
data:
  password: <password-vault-key>
```

And then once the plugin is done doing the substitutions, it outputs the yaml to standard out to then be applied by Argo CD. The resulting yaml would look like:
```
kind: Secret
apiVersion: v1
metadata:
  name: example-secret
  annotations:
    avp_path: "path/to/secret"
type: Opaque
data:
  password: cGFzc3dvcmQK # The Value from the key password-vault-key in vault
```

## Usage

### Environment Variables
The plugin requires a certain set of Environment Variables to be able to run correctly.

`AVP_VAULT_ADDR` - Address of your Vault

`AVP_PATH_PREFIX` - The prefix for the vault path. If you do not add the path as an annotation to your resource, the plugin expects a certain pattern of `AVP_PATH_PREFIX/${resource_type}` for example, if you are working with a Secret the plugin will look for `AVP_PATH_PREFIX/secret`.

`AVP_TYPE` - The type of Vault backend. As of now we only support `vault`. Other types are in the works.

`AVP_AUTH_TYPE` - The type of authentication mechanism. As of now we only support `approle` and `github` for `vault`. As time goes we plan on adding more options.

Make sure that these environment variables are available to the plugin when running it, whether that is in Argo CD or as a CLI tool.

### As a Vault Plugin
This plugin is meant to be used with Argo CD. In order to use the plugin you can add it to your Argo CD instance as a volume mount or build your own Argo CD image.
The Argo CD docs provide information on how to get started https://argoproj.github.io/argo-cd/operator-manual/custom_tools/.

Will look something like:
```
containers:
- name: argocd-repo-server
  volumeMounts:
  - name: custom-tools
    mountPath: /usr/local/bin/argocd-vault-plugin
    subPath: argocd-vault-plugin
  envFrom:
    - secretRef:
        name: argocd-vault-plugin-credentials
volumes:
- name: custom-tools
  emptyDir: {}
initContainers:
- name: download-tools
  image: alpine:3.8
  command: [sh, -c]
  args:
    - wget -O argocd-vault-plugin
      https://github.com/IBM/argocd-vault-plugin/releases/download/v0.1.0/argocd-vault-plugin_0.1.0_linux_amd64

      mv argocd-vault-plugin /custom-tools/
  volumeMounts:
    - mountPath: /custom-tools
      name: custom-tools
```

After making the plugin available, you must then register the plugin, documentation can be found at https://argoproj.github.io/argo-cd/user-guide/config-management-plugins/#plugins on how to do that.

For this plugin, you would add this:
```
data:
  configManagementPlugins: |-
    - name: argocd-vault
      generate:
        command: ["argocd-vault-plugin"]
        args: ["generate", "./"]
```

If you are using helm you might add something like:
```
configManagementPlugins: |
  - name: argocd-vault-helm
    generate:
      command: [sh, -c]
      args: ["helm template . > all.yaml && argocd-vault-plugin generate all.yaml"]
```

to the `argocd-cm` configMap.

Once that is done, the plugin has been registered with Argo CD and can be used by Applications.

To tell you Argo Cd Application to use the plugin you would specify it in the Application CRD
```
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: <appName>
spec:
  destination:
    name: ''
    namespace: default
    server: 'https://kubernetes.default.svc'
  source:
    path: .
    plugin:
      name: argocd-vault-plugin
      env:
        - name: AVP_PATH_PREFIX # Optional: Can be added to resource or a env var via a kubernetes secret
          value: some/path/to/vault
    repoURL: <repo>
    targetRevision: HEAD
  project: default
```
Or you can pass the config-management-plugin flag to the Argo CD CLI app create command:  
`argocd app create <appName> --config-management-plugin argocd-vault-plugin`

### As a CLI
The plugin can be used as just a cli tool if you are using a CI/CD system other than argo. You just run the tool like:

`argocd-vault-plugin generate ./path-to-templates`

And it will output the generated yaml files to standard out.

## Contributing

You can view the documentation on contibuting [here](./Contributing.md)
