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

// NaviReconciler reconciles a Navi object
type NaviReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=lge.com,resources=navis,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=lge.com,resources=navis/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=lge.com,resources=navis/finalizers,verbs=update
// add deployment, service, configmap rbac
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Navi object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *NaviReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	formerNavi := &lgecomv1alpha1.Navi{}
	err := r.Get(ctx, req.NamespacedName, formerNavi)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	logger.Info("Navi", "Navi", formerNavi)

	namespacedName := types.NamespacedName{Name: formerNavi.Name, Namespace: formerNavi.Namespace}

	deployNaviFound := &appsv1.Deployment{}
	err = r.Get(ctx, namespacedName, deployNaviFound)
	if err != nil {
		if errors.IsNotFound(err) {
			deploy := r.deploymentForNavi(formerNavi)
			if err := r.Create(ctx, deploy); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		} else {
			return ctrl.Result{}, err
		}
	}

	rosDomainId := formerNavi.Spec.RosDomainId
	if deployNaviFound.Spec.Template.Spec.Containers[0].Env[0].Value != rosDomainId {
		logger.Info("Update ROS_DOMAIN_ID", "ROS_DOMAIN_ID", rosDomainId)
		deployNaviFound.Spec.Template.Spec.Containers[0].Env[0].Value = rosDomainId
		if err := r.Update(ctx, deployNaviFound); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	network := formerNavi.Spec.NetWork
	if deployNaviFound.Annotations["k8s.v1.cni.cncf.io/networks"] != network {
		logger.Info("Update Network", "Network", network)
		deployNaviFound.Annotations["k8s.v1.cni.cncf.io/networks"] = network
		deployNaviFound.Spec.Template.Annotations["k8s.v1.cni.cncf.io/networks"] = network
		if err := r.Update(ctx, deployNaviFound); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	naviImage := formerNavi.Spec.Image
	if deployNaviFound.Spec.Template.Spec.Containers[0].Image != naviImage {
		logger.Info("Update Navi Image", "NaviImage", naviImage)
		deployNaviFound.Spec.Template.Spec.Containers[0].Image = naviImage
		if err := r.Update(ctx, deployNaviFound); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	naviArgs := formerNavi.Spec.Args
	if deployNaviFound.Spec.Template.Spec.Containers[0].Args[0] != naviArgs {
		logger.Info("Update Navi Args", "NaviArgs", naviArgs)
		deployNaviFound.Spec.Template.Spec.Containers[0].Args[0] = naviArgs
		if err := r.Update(ctx, deployNaviFound); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	hostNetwork := formerNavi.Spec.HostNetwork
	if deployNaviFound.Spec.Template.Spec.HostNetwork != hostNetwork {
		logger.Info("Update HostNetwork", "HostNetwork", hostNetwork)
		deployNaviFound.Spec.Template.Spec.HostNetwork = hostNetwork
		if err := r.Update(ctx, deployNaviFound); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	nodeSelector := formerNavi.Spec.NodeSelector
	if !reflect.DeepEqual(deployNaviFound.Spec.Template.Spec.NodeSelector, nodeSelector) {
		logger.Info("Update NodeSelector", "NodeSelector", nodeSelector)
		deployNaviFound.Spec.Template.Spec.NodeSelector = nodeSelector
		if err := r.Update(ctx, deployNaviFound); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	configNameSpaceName := types.NamespacedName{Name: formerNavi.Name + "-configmap", Namespace: formerNavi.Namespace}

	configMap := &corev1.ConfigMap{}
	err = r.Get(ctx, configNameSpaceName, configMap)
	if err != nil {
		if errors.IsNotFound(err) {
			cm := r.configMapForFormer(formerNavi)
			if err := r.Create(ctx, cm); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		} else {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *NaviReconciler) deploymentForNavi(navi *lgecomv1alpha1.Navi) *appsv1.Deployment {
	ls := labelsForNavi(navi.Name)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      navi.Name,
			Namespace: navi.Namespace,
			Annotations: map[string]string{
				"k8s.v1.cni.cncf.io/networks": navi.Spec.NetWork,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
					Annotations: map[string]string{
						"k8s.v1.cni.cncf.io/networks": navi.Spec.NetWork,
					},
				},
				Spec: corev1.PodSpec{
					HostNetwork: navi.Spec.HostNetwork,
					Containers: []corev1.Container{
						{
							Name:  "navi",
							Image: navi.Spec.Image,
							Args:  []string{navi.Spec.Args},
							Env: []corev1.EnvVar{
								{
									Name:  "ROS_DOMAIN_ID",
									Value: navi.Spec.RosDomainId,
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
									Name:      "params",
									MountPath: "/root/workspace/nav2/params",
								},
								{
									Name:      "maps",
									MountPath: "/root/workspace/nav2/maps",
								},
								{
									Name:      "rviz",
									MountPath: "/root/workspace/nav2/rviz",
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
							Name: "params",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: navi.Spec.ParamPath,
								},
							},
						},
						{
							Name: "maps",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: navi.Spec.MapPath,
								},
							},
						},
						{
							Name: "rviz",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: navi.Spec.RvizPath,
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
							Name: "regcred",
						},
					},
					NodeSelector: navi.Spec.NodeSelector,
				},
			},
		},
	}

	ctrl.SetControllerReference(navi, dep, r.Scheme)
	return dep
}

func (r *NaviReconciler) configMapForFormer(navi *lgecomv1alpha1.Navi) *corev1.ConfigMap {
	labels := map[string]string{
		"app": navi.Name,
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      navi.Name + "-configmap",
			Namespace: navi.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			"ros_domain_id":    navi.Spec.RosDomainId,
			"initial_pose_x":   "0.0",
			"initial_pose_y":   "0.0",
			"initial_pose_z":   "0.0",
			"initial_pose_yaw": "0.0",
		},
	}

	// Set Former instance as the owner and controller
	ctrl.SetControllerReference(navi, cm, r.Scheme)
	return cm
}

func labelsForNavi(name string) map[string]string {
	return map[string]string{"app": name}
}

// func (r *NaviReconciler) getPods(ctx context.Context, navi *lgecomv1alpha1.Navi, name string) ([]corev1.Pod, error) {
// 	podList := &corev1.PodList{}
// 	err := r.List(ctx, podList, client.InNamespace(navi.Namespace), client.MatchingLabels(labelsForNavi(name)))
// 	if err != nil {
// 		return nil, err
// 	}
// 	return podList.Items, nil
// }

// SetupWithManager sets up the controller with the Manager.
func (r *NaviReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&lgecomv1alpha1.Navi{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
