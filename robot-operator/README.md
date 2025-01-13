# Robot Operator

쿠버네티스에서 로봇과 내비게이션 운영을 위한 로봇 CRD와 컨트롤러 구현 프로젝트 입니다.

## 1. Overview

- 로봇 운영을 위한 로봇 CRD와 컨트롤러 구현
  - 로봇 CRD를 생성하면, 해당 정보를 바탕으로 로봇을 배포
  - 로봇 CRD의 정보를 변경하면, 해당 정보를 바탕으로 로봇을 수정하여 재배포
- former, navie, ros-kube 등 로봇 운영을 위한 다양한 Operator 제공

## Operator List

| Operator | Description | CRD | etc |
| --- | --- | --- | --- |
| Former Operator | Former 로봇 운영을 위한 Operator | [former_crd](./config/samples/_v1alpha1_former.yaml) | - |
| Navi Operator | Nav2를 배포하고 운영하기 위한 Operator | [navi_crd](./config/samples/_v1alpha1_navi.yaml) | - |
| ROS-Kube Operator | ROS-Kube 로봇 운영을 위한 Operator, Nav2로부터 amcl_pose 토픽을 받아와 ConfigMap과 Param에 업데이트 | [ros_kube_crd](./config/samples/_v1alpha1_roskube.yaml) | Ros와 K8s 연결에 활용 |

## 2. Robot Operator Deploy

- Kubernetes 클러스터 내부 환경에서 Makefile을 통해 Robot Operator 배포

```bash
# 설정 파일 생성
$ make generate
$ make manifests

# Docker image build and push
$ make docker-build docker-push # Private registry 사용 시, ./config/manager/manager.yaml 에 imagePullSecrets 추가 필요

# Robot Operator 배포
$ make deploy
```

### 2.1. Robot Operator 배포 확인

- Robot Operator 배포 확인

```bash
# Robot Operator 배포 확인
$ kubectl get crd
```

## 3. Robot CR 생성

### 3.1. Former Robot CR 생성

- Former Robot CR을 생성하여 로봇에 배포
- 설정값을 Spec 항목에 입력

```yaml
# former-robot-sample-1.yaml
apiVersion: lge.com/v1alpha1
kind: Former
metadata:   # Former 메타 데이터
  labels:
    app.kubernetes.io/name: former-0047
    app.kubernetes.io/instance: former-0047
    app.kubernetes.io/part-of: robot-operator
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: robot-operator
  name: former-0047     # Former Operator 이름
  namespace: lge-ebme   # Former Operator 배포할 Namespace
spec:
  robotName: "former_0047"  # 로봇 이름
  rosDomainId: "4"          # ROS Domain ID
  network: "ebme-network"   # Multus Network 이름
  image: "lgecloudroboticstask/former_docker:20240229"  # 로봇 Docker Image
  args: ". /opt/ros/humble/setup.bash && . /root/dev_ws/install/setup.bash && ros2 launch former_bringup bringup_robot.launch.py"   # 로봇 실행 명령어
  hostNetwork: false     # Host Network 사용 여부
  nodeSelector:
    kubernetes.io/hostname: "former-0047" # Node Selector
```

### 3.2. Navi Robot CR 생성

- Navi Robot CR을 생성하여 로봇에 배포
- Nav2 시작 시 저장된 Params, Map, Rviz 설정 파일을 불러와 실행
- 설정값을 Spec 항목에 입력

```yaml
# navi-robot-sample-1.yaml
apiVersion: lge.com/v1alpha1
kind: Navi
metadata:   # Navi 메타 데이터
  labels:
    app.kubernetes.io/name: former-0047-navi
    app.kubernetes.io/instance: former-0047-navi
    app.kubernetes.io/part-of: robot-operator
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: robot-operator
  name: former-0047-navi    # Navi Operator 이름
  namespace: lge-ebme       # Navi Operator 배포할 Namespace
spec:
  rosDomainId: "4"          # ROS Domain ID
  network: "ebme-network"   # Multus Network 이름
  image: "lgecloudroboticstask/navi:nav2-version2"  # Nav2 Docker Image
  args: ". /opt/ros/humble/setup.bash && ros2 launch nav2_bringup bringup_launch.py map:=/root/workspace/nav2/maps/seocho_test_19.yaml params_file:=/root/workspace/nav2/params/nav2_params.yaml"   # Nav2 실행 명령어
  paramPath: "/home/former/workspace/nav2/params"   # Nav2 params 파일 저장 경로
  mapPath: "/home/former/workspace/nav2/maps"    # Nav2 map 파일 저장 경로
  rvizPath: "/home/former/workspace/nav2/rviz"  # Nav2 rviz 설정 파일 저장 경로
  hostNetwork: false    # Host Network 사용 여부
  nodeSelector:
    kubernetes.io/hostname: "former-0047"   # Node Selector
```

### 3.3. ROS-Kube Robot CR 생성

- ROS-Kube Robot CR을 생성하여 로봇에 배포
- Nav2로부터 amcl_pose 토픽을 받아와 ConfigMap과 Param에 업데이트
- 설정값을 Spec 항목에 입력

```yaml
# ros-kube-robot-sample-1.yaml
apiVersion: lge.com/v1alpha1
kind: RosKube
metadata:   # RosKube 메타 데이터
  labels:
    app: former-0047-roskube
  name: former-0047-roskube     # RosKube Operator 이름
  namespace: lge-ebme           # RosKube Operator 배포할 Namespace
spec:
  naviName: "former-0047-navi-configmap"    # Navi Operator 이름
  rosDomainId: "4"          # ROS Domain ID
  network: "ebme-network"   # Multus Network 이름
  image: "lgecloudroboticstask/ros-kube:latest"     # ROS-Kube Docker Image
  paramPath: "/home/former/workspace/nav2/params"   # Nav2 params 파일 저장 경로
  amclPose: "/amcl_pose"    # AMCL Pose Topic 이름
  hostNetwork: false        # Host Network 사용 여부
  nodeSelector:
    kubernetes.io/hostname: "former-0047"   # Node Selector
```
