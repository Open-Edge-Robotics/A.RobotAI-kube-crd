# Cloud Bridge CRD

Cloud Bridge 를 Kubernetes CRD 로 구현한 프로젝트.
[cloud-bridge-crd](http://mod.lge.com/hub/seo.youngchae/cloud-bridge-crd.git) Repository 에서 관리.

## 1. Overview

- Cloud Bridge 배포와 관리를 간소화 하기 위해 Kubernetes CRD 로 구현
- Spec 정보를 입력한 Cloud Bridge CR 을 생성하면, 해당 정보를 바탕으로 Cloud Bridge 를 배포
- Cloud Bridge CR 의 Spec 정보를 변경하면, 해당 정보를 바탕으로 Cloud Bridge 를 수정하여 재배포

## 2. Cloud Bridge CRD Deploy

- Kubernetes 클러스터 내부 환경에서 Makefile 을 통해 Cloud Bridge CRD 배포

```bash
# 설정 파일 생성
$ make generate
$ make manifests

# Docker image build and push
$ make docker-build docker-push # Private registry 사용 시, ./config/manager/manager.yaml 에 imagePullSecrets 추가 필요

# Cloud Bridge CRD 배포
$ make deploy
```

### 2.1. Cloud Bridge CRD 배포 확인

- Cloud Bridge CRD 배포 확인

```bash
# Cloud Bridge CRD 배포 확인
$ kubectl get crd

NAME                              CREATED AT
cloudbridges.ebme.lge.com         2023-06-15T07:00:00Z
```

## 3. Cloud Bridge CR 생성

- Cloud Bridge CR 을 생성하여 Cloud Bridge 를 배포
- 설정값을 Spec 항목에 입력

```yaml
# cloudbridge-sample-1.yaml
apiVersion: ebme.lge.com/v1
kind: CloudBridge
metadata:
  labels:
    app.kubernetes.io/name: cloudbridge
    app.kubernetes.io/instance: cloudbridge-sample
    app.kubernetes.io/part-of: cloud-bridge-crd
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: cloud-bridge-crd
  name: cloudbridge-sample-1      # Cloud Bridge 이름
  namespace: lge-ebme             # Cloud Bridge 배포할 Namespace
spec:
  netWork: "ebme-network"         # Cloud Bridge 배포할 Multus Network 이름
  robotName: "cloi_31705"         # Cloud Bridge 배포할 Robot 이름
  rosDomainId: "72"               # ROS Domain ID
  image: "lgecloudroboticstask/cloud_bridge:humble-offloading-20230508"   # Cloud Bridge 배포할 Docker Image(선택) 미입력시 기본값 사용
  port: 31705                     # Cloud Bridge 배포할 Port
  configDir: "/home/ubuntu/cloud_bridge/config_offloading"                # Cloud Bridge 배포할 Config Directory(선택) 미입력시 기본값 사용
  command: ["/bin/bash", "-c"]    # Cloud Bridge 배포할 Command(선택) 미입력시 기본값 사용
  args:                           # Cloud Bridge 배포할 Args(선택) 미입력시 기본값 사용
    [
      "echo $ROBOT_NAME; . /opt/ros/$ROS_DISTRO/setup.bash; . /home/ubuntu/cloud_bridge/install/setup.bash; sed 's/microk8s_chatter/chatter2/g' /home/ubuntu/publisher.py > /home/ubuntu/cloud_bridge/publisher.py; ros2 launch cloud_bridge cloud_bridge_server.launch.py config_dir:=$CONFIG_DIR & python3 /home/ubuntu/cloud_bridge/publisher.py talker;",
    ]
  nodeSelector: "cloud"           # Cloud Bridge 배포할 Node Selector(선택) 미입력시 기본값(Master Node) 사용
```

- Cloud Bridge CR 생성

```bash
# Cloud Bridge CR 생성
$ kubectl apply -f cloudbridge-sample-1.yaml
```

### 3.1. Cloud Bridge CR 생성 확인

- Cloud Bridge CR 생성 확인

```bash
# Cloud Bridge CR 생성 확인
$ kubectl -n lge-ebme get cloudbridge

NAME                   AGE
cloudbridge-sample-1   23m
```

- Cloud Bridge CR 상세 정보 확인

```bash
# Cloud Bridge CR 상세 정보 확인
$ kubectl -n lge-ebme describe cloudbridge cloudbridge-sample-1
```

## 4. Cloud Bridge CR 수정

- Cloud Bridge CR 의 Spec 항목을 수정하여 Cloud Bridge 를 재배포

```bash
# Cloud Bridge CR 수정
$ kubectl edit cloudbridge cloudbridge-sample-1
```

- Cloud Bridge Spec 항목 수정

```yaml
# Please edit the object below. Lines beginning with a '#' will be ignored,
# and an empty file will abort the edit. If an error occurs while saving this file will be
# reopened with the relevant failures.
#
apiVersion: ebme.lge.com/v1
kind: CloudBridge
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ebme.lge.com/v1","kind":"CloudBridge","metadata":{"annotations":{},"name":"cloudbridge-sample-1","namespace":"lge-ebme"},"spec":{"port":31535,"robotName":"cloi_31535","rosDomainId":"72"}}
  creationTimestamp: "2023-06-15T05:21:45Z"
  generation: 2
  name: cloudbridge-sample-1
  namespace: lge-ebme
  resourceVersion: "29674232"
  uid: 5852cc47-5dc4-439b-9a00-d002992fc413
spec:
  port: 31545
  robotName: "cloi_31535"
  rosDomainId: "72"
  image: "lgecloudroboticstask/cloud_bridge:humble-offloading-20230508"
  configDir: "/home/ubuntu/cloud_bridge/config_offloading"
  nodeSelector: "node-name"
status:
  nodes:
  - cloudbridge-sample-1-6668647d89-mf4b2
```

- Cloud Bridge CR 수정 완료 시 Controller 가 인지하고 Cloud Bridge 를 재배포

## 5. Cloud Bridge CR 삭제

- Cloud Bridge CR 을 삭제하여 Cloud Bridge 를 삭제

```bash
# Cloud Bridge CR 삭제
$ kubectl delete -f cloudbridge-sample-1.yaml
```

## 6. Cloud Bridge CRD 삭제

- Cloud Bridge CRD 배포 취소하여 Cloud Bridge CRD 를 삭제

```bash
# Cloud Bridge CRD 배포 취소
$ make undeploy
```

## 7. Cloud Bridge ROS 연결 확인

- Cloud Bridge CR 을 생성하여 Cloud Bridge 를 배포한 후 ROS 연결 확인

```bash
# Cloud Bridge CR Shell 접속
# 생성된 pod 이름에 맞춰 변경
$ kubectl -n lge-ebme exec -it cloudbridge-sample-1-6668647d89-mf4b2 -- /bin/bash

# ROS 연결 확인
$ source ./install/setup.bash
$ ros2 topic list

# chatter topic echo
$ ros2 topic echo /chatter2
```
