# Navigation CRD and Controller

## Overview

Navigation CRD 를 만들고 이를 controller 를 통하여 관리하는 프로젝트

## Project Git Repository

- [navi-crd](http://mod.lge.com/hub/seo.youngchae/navi-crd.git) lge mod repository

## Prerequisites

- [Kubernetes](https://kubernetes.io/) or [k3s](https://k3s.io/) cluster
- [Go](https://golang.org/) programming language
- [Docker](https://www.docker.com/) container runtime
- [Operator SDK](https://sdk.operatorframework.io/)

### Optional

- [Kustomize](https://kustomize.io/)
- [Kubebuilder](https://book.kubebuilder.io/)
- [Helm](https://helm.sh/)

## Getting Started

### Deploy Controller

- [Operator SDK](https://sdk.operatorframework.io/) 를 통하여 작성한 Controller 를 배포
  - Navigation CRD 를 생성하고 이를 관리하는 Controller 개발
  - Navigation CR이 생성되면 controller가 인지하여 Nav2 Deployment를 실행
  - 생성 시 CR에 입력된 Robot name, initialpose 값과 ROS Domain ID를 Nav2 Deployment에 전달
- make 명령어를 통하여 간단하게 관련 파일을 생성하고 배포 가능

```bash
# Config yaml files generate
$ make generate
$ make manifests

# Docker image build and push
$ make docker-build docker-push # Private registry 사용 시, ./config/manager/manager.yaml 에 imagePullSecrets 추가 필요

# Controller deploy
$ make deploy
```

### Deploy Navigation CR

- Navigation CR 설정 값에 따라 Nav2 Deployment를 실행
  - Robot name, initialpose, ROS Domain ID 등을 설정
  - initialpose 값은 Position(x, y, z)과 Orientation(w) 값으로 구성됨

```yaml
# Sample Navigation CR
apiVersion: ebme.lge.com/v1
kind: Navigation
metadata:
  labels:
    app.kubernetes.io/name: navigation
    app.kubernetes.io/instance: navigation-sample
    app.kubernetes.io/part-of: navi-crd
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: navi-crd
  name: navigation-sample
  namespace: lge-ebme
spec:
  netWork: "ebme-network"                   # (Required) Multus Network 이름
  robotName: cloi_31585                     # (Required) 로봇 이름
  rosDomainId: "72"                         # (Required) ROS Domain ID
  image: "lgecloudroboticstask/ros:nav2"    # (Optional) Nav2 Docker Image, 미입력 시 기본값 사용
  command: ["/bin/bash", "-c"]              # (Optional) Command, 미입력 시 기본값 사용
  args:                                     # (Optional) Args, 미입력 시 기본값 사용
    [
      ". /opt/ros/humble/setup.bash; . /opt/cloi_ws/install/setup.bash; ros2 launch nav2_bringup navigation_launch.py",
    ]
  initialPose:                              # (Required) 로봇 초기 위치
    position:
      "x": "33.1"
      "y": "22.2"
      "z": "11.3"
    orientation:
      "w": "22.4"
  nodeSelector: "cloud"                     # (Optional) Node Selector 설정, 미입력 시 기본값(Master Node) 사용

```

```bash
# Navigation CR deploy
$ kubectl apply -f ./config/samples/ebme_v1_navigation.yaml

navigation.ebme.lge.com/navigation-sample created
```

### Check Navigation CR and Controller

- Navigation CR 생성 및 Controller 동작 확인

```bash
# Check Navigation CR
$ kubectl get navigation

NAME               AGE
navigation-sample   2m

# Check Controller
$ kubectl get pods -n navi-crd-system

NAME                                          READY   STATUS    RESTARTS   AGE
navi-crd-controller-manager-6b5b4c6b6-4q9qf   2/2     Running   0          2m

# Check Navigation Pod
$ kubectl get pods -n lge-ebme

NAME                                       READY   STATUS    RESTARTS   AGE
navigation-sample-677f4c45f9-hrkhn         1/1     Running   0          111s
```

### ROS2 Navigation

- goal_pose 토픽을 사용하여 로봇을 이동
  - 현재 Cloud Bridge Client 에서 goal_pose 토픽을 사용하여 로봇을 이동 가능

```bash
# Publish Navigation Goal Pose
# LGE seocho b1 map
$ ros2 topic pub -1 /cloi_31585/goal_pose geometry_msgs/PoseStamped "{header: {stamp: {sec: 0}, frame_id: 'map'}, pose: {position: {x: 51.5, y: 28.0, z: 0.0}, orientation: {w: 0.0}}}"
```
