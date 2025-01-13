# A.RobotAI-kube-crd
Kubernetes custom resource definition to deploy the specific robot engines and applications

## This Project to subject to the following contents.
```
연구개발과제명: 일상생활 공간에서 자율행동체의 복합작업 성공률 향상을 위한 자율행동체 엣지 AI SW 기술 개발

세부 개발 카테고리
● 지속적 지능 고도화를 위한 자율적 흐름제어 학습 프레임워크 기술 분석 및 설계
- 기밀성 데이터 활용 지능 고도화를 위한 엣지와 클라우드 분산 협업 학습 프레임워크 기술
- 엣지와 클라우드 협력 학습 간 최적 자원 활용 및 지속적 지능 배포를 위한 자율적 학습흐름제어 기술

개발 내용 
- 엣지와 클라우드 분산 협업을 위한 지속적 지능 배포 프레임워크 
- 자율행동체 엣지 기반 클러스터링 솔루션 및 분산 학습 프레임워크 개발
```

---

# Robot Operator

쿠버네티스에서 로봇 운영을 위한 로봇 CRD와 컨트롤러 구현 프로젝트 입니다.

## Operator List

| Operator | Description | etc |
| --- | --- | --- |
|[Cloud Bridge Operator](./cloud-bridge-crd) | 클라우드 브릿지를 위한 CRD와 컨트롤러 구현 | [README](./cloud-bridge-crd/README.md) |
|[Navigation Operator](./navi-crd) | 네비게이션을 위한 CRD와 컨트롤러 구현 | [README](./navi-crd/README.md) |
|[Robot Operator](./robot-operator) | 로봇 운영을 위한 로봇 CRD와 컨트롤러 구현 | [README](./robot-operator/README.md) |


---

## Related Project

```
https://github.com/Open-Edge-Robotics/A.EdgeAI-fl-perception
To deploy Perception engine,  which is model resulted from Federated Learning 

https://github.com/Open-Edge-Robotics/A.RobotAI-ros2-streamer
To make and send stream of ROS 2 images captured from carmera attached to Robot.

https://github.com/Open-Edge-Robotics/A.EdgeAI-rosbag-reader
Reader function to extract the data (including images) from rosbag of ROS2

https://github.com/Open-Edge-Robotics/A.CloudAI-fl-flower
Flower Framework, which is Federated Learning to be used as Distributed Collaborative Learing Framework

https://github.com/Open-Edge-Robotics/A.CloudAI-kube-multi-ctl
Customized kubectl to manage multiple k8s master node (standalone node)

https://github.com/Open-Edge-Robotics/A.RobotAI-kube-crd  (ebme-crd)
Kubernetes custom resource definition to deploy the specific robot engines and applications
```
