# K8S Object Template Operator [![CircleCI](https://circleci.com/gh/ericogr/k8s-object-template-operator.svg?style=svg)](https://circleci.com/gh/ericogr/k8s-object-template-operator)
This operator can be used to create any kubernetes object dynamically. Build your templates and set parameters to create new k8s objects.

## Use case
Many kubernetes clusters are shared among many applications and teams. Sometimes services are available within the cluster scope and teams can use it to create or configure services using kubernetes spec (such as PrometheusRule, ExternalDNS, etc.). Some of these specs are too complex or contains some configurations that we do not want to expose. You can automate it's creation using templates.

This operator can create kubernete objects based on templates and simple namespaced parameters. You can give permissions to user create parameters but hide templates and created objects from developers or users using the Kubernetes RBAC system.

# Installation
Use the file [config/specs.yaml](specs/object-template-operator.yaml) as base to deploy this operator with all permissions (dev/test mode). See section about roles bellow.

```sh
kubectl apply -f https://raw.githubusercontent.com/ericogr/k8s-object-template-operator/master/specs/object-template-operator.yaml
```

## Additionals Kubernetes Roles
This operator should be allowed to create objects defined in templates. With default permission, it can create any object, but it can be a bit tricky. The ClusterRole ```k8s-ot-manager-role``` can be used to set permissions as needed.

See this example to add PrometheusRules permission to this operator:

```yaml
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  creationTimestamp: null
  name: k8s-ot-manager-role
rules:
- apiGroups:
  - monitoring.coreos.com
  resources:
  - prometheusrules
  verbs:
  - create
  - get
  - list
  - patch
  - update
- apiGroups:
  - template.ericogr.github.com
  resources:
  - objecttemplateparams
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - template.ericogr.github.com
  resources:
  - objecttemplateparams/status
  verbs:
  - get
  - patch
  - update
- apiGroups:
  - template.ericogr.github.com
  resources:
  - objecttemplates
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - template.ericogr.github.com
  resources:
  - objecttemplates/status
  verbs:
  - get
  - patch
  - update
```
# New Custom Resource Definitions (CRD's)
We have two CRD's: [ObjectTemplate](config/crd/bases/template.ericogr.github.com_objecttemplates.yaml) and [ObjectTemplateParameters](config/crd/bases/template.ericogr.github.com_objecttemplateparams.yaml).

**ObjectTemplate (cluster scope):** used as model to create objects in namespaces (can be used by k8s admins)

**ObjectTemplateParameters (namespaced):** used as model parameters to create objects in their namespace (can be used by k8s users/devs)

# Templates (ObjectTemplate)
Use templates as a base to create kubernetes objects. Users can define your own parameters to create new objects.

## Template example

```yaml
---
apiVersion: template.ericogr.github.com/v1
kind: ObjectTemplate
metadata:
  name: objecttemplate-prometheus-rules-default
spec:
  description: Default prometheus rule
  objects:
  - kind: PrometheusRule
    apiVersion: monitoring.coreos.com/v1
    metadata:
      labels:
        chave1: valor1
        chave2: valor2
      annotations:
        chave1a: valor1a
        chave2a: valor2a
    name: prometheus-rule-default
    spec: |-
      groups:
      - name: pods
        rules:
        - alert: pod_not_ready
          annotations:
            description: 'Pod not ready : {{"{{ $labels.pod }}"}}'
            summary: 'Pod not ready: {{"{{ $labels.pod }}"}}'
          expr: sum by(pod) (kube_pod_status_ready{namespace="{{ .__namespace }}"} == 0) != 0
          for: 10m
          labels:
            app_name: {{ .app_name }}
            app_route: slack
            app_severity: critical
            app_slack_channel: '{{ .app_slack_channel }}'
```

## Basic Template Substitution System
You can use sintax like ```{{ .variable }}``` to replace parameters. Let's say you create ```app_name: myapp```. You can use ```{{ .app_name }}``` inside spec template to be replaced in runtime by this controller. If you need to scape braces, use ```{{"{{anything}}"}}```

### System Runtime Variables

|Name         |Description       |
|-------------|------------------|
|__namespace  |Current namespace |
|__apiVersion |API Version       |
|__kind       |The name of kind  |
|__name       |Name of object    |

# Parameters (ObjectTemplateParams)
Users can define your own parameters to create new objects based on templates.

## Parameters example

```yaml
---
apiVersion: template.ericogr.github.com/v1
kind: ObjectTemplateParams
metadata:
  name: objecttemplateparams-sample
  namespace: default
spec:
  templates:
  - name: objecttemplate-prometheus-rules-default
    values:
      app_name: myapp
      app_slack_channel: '#slack-channel'
 ```
