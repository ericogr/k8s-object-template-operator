# K8S Object Template Operator [![CircleCI](https://circleci.com/gh/ericogr/k8s-object-template-operator.svg?style=svg)](https://circleci.com/gh/ericogr/k8s-object-template-operator)
This operator can be used to create any kubernetes object dynamically. Create your models and set parameters to create new objects.

## Use case
Many kubernetes clusters are shared among many applications and teams. Sometimes services are available within the cluster scope and teams can use it to create or configure services using kubernetes spec (such as PrometheusRule, ExternalDNS, etc.). Some of these specs are too complex or contain some configurations that we do not want to expose. You can automatize creation of many objects using one template.

Use this operator can create these kubernete objects based on templates and simple namespaced parameters. You can give permissions to user create parameters but hide templates and created objects from developers / users using the Kubernetes RBAC system.

## New Custom Resource Definitions (CRD's)
We have two CRD's: [ObjectTemplate](config/crd/bases/template.ericogr.github.com_objecttemplates.yaml) and [ObjectTemplateParameters](config/crd/bases/template.ericogr.template.ericogr.github.com_objecttemplateparams.yaml).

**ObjectTemplate (cluster scope):** used as model to create objects in namespaces (can be used by k8s admins)

**ObjectTemplateParameters (namespaced):** used as model parameters to create objects in their namespace (can be used by k8s users/devs)

## Basic Template Substitution System
You can use sintax like ```{{ .variable }}``` to replace parameters. Let's say you create ```app_name: myapp```. You can use {{ .app_name }} inside spec template to be replaced in runtime by this controller.

### System Runtime Variables

|Name         |Description       |
|-------------|------------------|
|__namespace  |Current namespace |
|__apiVersion |API Version       |
|__kind       |The name of kind  |
|__name       |Name of object    |

**Template example**

```sh
---
apiVersion: template.ericogr.github.com/v1
kind: ObjectTemplate
metadata:
  name: objecttemplate-sample
spec:
  template:
    name: prometheus-rules-default
    kind: PrometheusRule
    apiVersion: monitoring.coreos.com/v1
    metadata:
      labels:
        chave1: valor1
        chave2: valor2
      annotations:
        chave1a: valor1a
        chave2a: valor2a
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
            app_app: {{ .app_app }}
            app_route: slack
            app_severity: critical
            app_slack_channel: '{{ .app_slack_channel }}'

 ```

 **Parameters example**

 ```sh
 ---
apiVersion: template.ericogr.github.com/v1
kind: ObjectTemplateParams
metadata:
  name: objecttemplateparams-sample
  namespace: default
spec:
  templates:
  - name: prometheus-rules-default
    values:
      app_slack_channel: '#slack-channel'
      app_app: myapp
 ```
