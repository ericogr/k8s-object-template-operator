# K8S Object Template Operator [![CircleCI](https://circleci.com/gh/ericogr/k8s-object-template-operator.svg?style=svg)](https://circleci.com/gh/ericogr/k8s-object-template-operator)
This operator can be used to create any kubernetes object dynamically. Build your own templates using Kubernetes specs and set parameters to create new objects based on it.


![demo](img/demo.git?raw=true "demo")

## Use case
Many kubernetes clusters are shared among many applications and teams. Sometimes services are available within the cluster scope and teams can use it to create or configure services using kubernetes spec (such as ConfigMap, Secret, PrometheusRule, ExternalDNS, etc.). Some of these specs are too complex or contains some configurations that we do not want to expose. You can automate it's creation using this operator.

This operator can create kubernete objects based on templates specs and simple namespaced parameters. You can give permissions to user create parameters specs but forbit templates specs and created objects from developers or users using the Kubernetes RBAC system.

# Installation
Use the file [specs/object-template-operator.yaml](specs/object-template-operator.yaml) to start deploy this operator with all permissions (dev/test mode). For production, see section about roles bellow.

```sh
kubectl apply -f https://raw.githubusercontent.com/ericogr/k8s-object-template-operator/master/specs/object-template-operator.yaml
```

## Additionals Kubernetes Roles
This operator should be allowed to create objects defined in templates. With default permission, it can create any object, but it can be a bit tricky. The ClusterRole ```k8s-ot-manager-role``` can be used to set permissions as needed.

See this example to add ConfigMap permission to this operator:

```yaml
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  creationTimestamp: null
  name: k8s-ot-manager-role
rules:
# >> HERE, ADDED CONFIGMAP PERMISSIONS
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - create
  - get
  - list
  - patch
  - update
# <<
- apiGroups:
  - template.k8s.ericogr.com.br
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
  - template.k8s.ericogr.com.br
  resources:
  - objecttemplateparams/status
  verbs:
  - get
  - patch
  - update
- apiGroups:
  - template.k8s.ericogr.com.br
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
  - template.k8s.ericogr.com.br
  resources:
  - objecttemplates/status
  verbs:
  - get
  - patch
  - update
```
# New Custom Resource Definitions (CRD's)
You have two new CRD's: [ObjectTemplate](config/crd/bases/template.k8s.ericogr.com.br_objecttemplates.yaml) and [ObjectTemplateParameters](config/crd/bases/template.k8s.ericogr.com.br_objecttemplateparams.yaml).

**ObjectTemplate (cluster scope):** template used to create kubernetes objects at users namespaces (can be used by k8s admins)

**ObjectTemplateParameters (namespaced):** parameters used to create objects in their namespace (can be used by k8s users/devs)

# Templates (ObjectTemplate)
Use templates as a base to create kubernetes objects. Users can define your own parameters to create new objects.

## Template example

```yaml
---
apiVersion: template.k8s.ericogr.com.br/v1
kind: ObjectTemplate
metadata:
  name: objecttemplate-configmap-test
spec:
  description: ConfigMap test
  objects:
  - kind: ConfigMap
    apiVersion: v1
    metadata:
      labels:
        label1: labelvalue1
        label2: labelvalue2
      annotations:
        annotation1: annotationvalue1
        annotation2: annotationvalue2
    name: configmap-test
    templateBody: |-
      data:
        name: '{{ .name }}'
        age: '{{ .age }}'
```

## Basic Template Substitution System
You can use sintax like ```{{ .variable }}``` to replace parameters. Let's say you created a template parameter with name/value ```name: foo```. You can use ```{{ .name }}``` inside ```templateBody``` template to be replaced in runtime. If you need to scape braces, use ```{{"{{anything}}"}}```.

### Library template functions (by Sprig)
There are many template functions library available to use. See some examples:

Remove spaces, convert to lowercase and truncate to 5 chars:

```template
{{ .username | trim | lower | trunc 5 }}
```
Convert text to base64:

```template
{{ .password | b64enc }}
```

Add 10 to age:

```template
{{ .age | add 10 }}
```

> More information: http://masterminds.github.io/sprig/

### System Runtime Variables

|Name         |Description       |
|-------------|------------------|
|__namespace  |Current namespace |
|__apiVersion |API Version       |
|__kind       |The name of kind  |
|__name       |Name of object    |

# Parameters (ObjectTemplateParams)
Users can define your own parameters to create new objects based on templates in their namespace.

## Parameters example

```yaml
---
apiVersion: template.k8s.ericogr.com.br/v1
kind: ObjectTemplateParams
metadata:
  name: objecttemplateparams-sample
  namespace: default
spec:
  templates:
  - name: objecttemplate-configmap-test
    values:
      name: foo
      age: '64'
 ```
