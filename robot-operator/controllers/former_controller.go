/*
Copyright 2024 seo.youngchae.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	lgecomv1alpha1 "lge.com/m/api/v1alpha1"
)

// FormerReconciler reconciles a Former object
type FormerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=lge.com,resources=formers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=lge.com,resources=formers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=lge.com,resources=formers/finalizers,verbs=update
// add deployment, service, configmap rbac
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Former object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *FormerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	former := &lgecomv1alpha1.Former{}
	err := r.Get(ctx, req.NamespacedName, former)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	namespacedName := types.NamespacedName{Name: former.Name, Namespace: former.Namespace}

	deployFound := &appsv1.Deployment{}
	err = r.Get(ctx, namespacedName, deployFound)
	if err != nil {
		if errors.IsNotFound(err) {
			// Define a new deployment
			deploy := r.deploymentForFormer(former)
			if err = r.Create(ctx, deploy); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		} else {
			return ctrl.Result{}, err
		}
	}

	robotName := former.Spec.RobotName
	if deployFound.Spec.Template.Spec.Containers[0].Env[0].Value != robotName {
		logger.Info("Updating Deployment", "Deployment.Namespace", deployFound.Namespace, "Deployment.Name", deployFound.Name)
		deployFound.Spec.Template.Spec.Containers[0].Env[0].Value = robotName
		if err = r.Update(ctx, deployFound); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	rosDomainId := former.Spec.RosDomainId
	if deployFound.Spec.Template.Spec.Containers[0].Env[1].Value != rosDomainId {
		logger.Info("Updating Deployment", "Deployment.Namespace", deployFound.Namespace, "Deployment.Name", deployFound.Name)
		deployFound.Spec.Template.Spec.Containers[0].Env[1].Value = rosDomainId
		if err = r.Update(ctx, deployFound); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	network := former.Spec.NetWork
	if deployFound.ObjectMeta.Annotations["k8s.v1.cni.cncf.io/networks"] != network {
		logger.Info("Updating Deployment", "Deployment.Namespace", deployFound.Namespace, "Deployment.Name", deployFound.Name)
		deployFound.ObjectMeta.Annotations["k8s.v1.cni.cncf.io/networks"] = network
		if err = r.Update(ctx, deployFound); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	image := former.Spec.Image
	if deployFound.Spec.Template.Spec.Containers[0].Image != image {
		logger.Info("Updating Deployment", "Deployment.Namespace", deployFound.Namespace, "Deployment.Name", deployFound.Name)
		deployFound.Spec.Template.Spec.Containers[0].Image = image
		if err = r.Update(ctx, deployFound); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	args := former.Spec.Args
	if deployFound.Spec.Template.Spec.Containers[0].Args[0] != args {
		logger.Info("Updating Deployment", "Deployment.Namespace", deployFound.Namespace, "Deployment.Name", deployFound.Name)
		deployFound.Spec.Template.Spec.Containers[0].Args[0] = args
		if err = r.Update(ctx, deployFound); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	nodeSelector := former.Spec.NodeSelector
	if !reflect.DeepEqual(deployFound.Spec.Template.Spec.NodeSelector, nodeSelector) {
		logger.Info("Updating Deployment", "Deployment.Namespace", deployFound.Namespace, "Deployment.Name", deployFound.Name)
		deployFound.Spec.Template.Spec.NodeSelector = nodeSelector
		if err = r.Update(ctx, deployFound); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	cmFound := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Namespace: former.Namespace, Name: former.Name + "-configmap"}, cmFound)
	if err != nil {
		if errors.IsNotFound(err) {
			// Define a new configmap
			cm := r.configMapForFormer(former)
			if err = r.Create(ctx, cm); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		} else {
			return ctrl.Result{}, err
		}
	}

	pods := &corev1.PodList{}
	pods, err = r.getPods(ctx, former, former.Name)
	if err != nil {
		logger.Error(err, "Failed to get Pods")
		return ctrl.Result{}, err
	}
	logger.Info("Pods", "Pods", pods.Items)

	// Pod restart when configmap is updated
	if cmFound.Data["robot_name"] != robotName || cmFound.Data["ros_domain_id"] != rosDomainId {
		for _, pod := range pods.Items {
			logger.Info("Deleting Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
			if pod.ObjectMeta.Labels["app"] == former.Name {
				if err = r.Delete(ctx, &pod); err != nil {
					return ctrl.Result{}, err
				}
			}
		}
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

func (r *FormerReconciler) deploymentForFormer(former *lgecomv1alpha1.Former) *appsv1.Deployment {
	labels := map[string]string{
		"app": former.Name,
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      former.Name,
			Namespace: former.Namespace,
			Labels:    labels,
			Annotations: map[string]string{
				"k8s.v1.cni.cncf.io/networks": former.Spec.NetWork,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
					Annotations: map[string]string{
						"k8s.v1.cni.cncf.io/networks": former.Spec.NetWork,
					},
				},
				Spec: corev1.PodSpec{
					HostNetwork: former.Spec.HostNetwork,
					Containers: []corev1.Container{
						{
							Name:  former.Name,
							Image: former.Spec.Image,
							Args:  []string{former.Spec.Args},
							Env: []corev1.EnvVar{
								{
									Name:  "ROBOT_NAME",
									Value: former.Spec.RobotName,
								},
								{
									Name:  "ROS_DOMAIN_ID",
									Value: former.Spec.RosDomainId,
								},
								{
									Name:  "NVIDIA_DRIVER_CAPABILITIES",
									Value: "all",
								},
								{
									Name:  "NVIDIA_VISIBLE_DEVICES",
									Value: "all",
								},
							},
							Command: []string{"/bin/bash", "-c"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "dev",
									MountPath: "/dev",
								},
								{
									Name:      "cyclonedds",
									MountPath: "/root/cyclonedds_conf",
								},
								{
									Name:      "src",
									MountPath: "/root/dev_ws/src",
								},
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: func(b bool) *bool { return &b }(true),
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("12000m"),
									corev1.ResourceMemory: resource.MustParse("64Gi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("3000m"),
									corev1.ResourceMemory: resource.MustParse("8Gi"),
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "dev",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/dev",
								},
							},
						},
						{
							Name: "cyclonedds",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/home/former/cyclonedds_conf",
								},
							},
						},
						{
							Name: "src",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/home/former/dev_ws/src",
								},
							},
						},
					},
					Tolerations: []corev1.Toleration{
						{
							Key:      "robot",
							Operator: corev1.TolerationOpEqual,
							Value:    "true",
							Effect:   corev1.TaintEffectNoSchedule,
						},
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{
							Name: "dockerhub-lgecloudroboticstask",
						},
					},
					NodeSelector: former.Spec.NodeSelector,
				},
			},
		},
	}

	// Set Former instance as the owner and controller
	ctrl.SetControllerReference(former, dep, r.Scheme)
	return dep
}

func (r *FormerReconciler) configMapForFormer(former *lgecomv1alpha1.Former) *corev1.ConfigMap {
	labels := map[string]string{
		"app": former.Name,
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      former.Name + "-configmap",
			Namespace: former.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			"robot_name":    former.Spec.RobotName,
			"ros_domain_id": former.Spec.RosDomainId,
		},
	}

	// Set Former instance as the owner and controller
	ctrl.SetControllerReference(former, cm, r.Scheme)
	return cm
}

func (r *FormerReconciler) getPods(ctx context.Context, former *lgecomv1alpha1.Former, name string) (*corev1.PodList, error) {
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(former.Namespace),
		client.MatchingLabels(map[string]string{"app": name}),
	}
	if err := r.List(ctx, podList, listOpts...); err != nil {
		return nil, err
	}
	return podList, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FormerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&lgecomv1alpha1.Former{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
