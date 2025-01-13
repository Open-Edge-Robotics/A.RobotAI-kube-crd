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

// RosKubeReconciler reconciles a RosKube object
type RosKubeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=lge.com,resources=roskubes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=lge.com,resources=roskubes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=lge.com,resources=roskubes/finalizers,verbs=update
// add deployment, service, configmap rbac
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RosKube object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *RosKubeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the RosKube instance
	rosKube := &lgecomv1alpha1.RosKube{}
	err := r.Get(ctx, req.NamespacedName, rosKube)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	namespacedName := types.NamespacedName{Name: rosKube.Name, Namespace: rosKube.Namespace}

	// Define a new Deployment object
	deployFound := &appsv1.Deployment{}
	err = r.Get(ctx, namespacedName, deployFound)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Creating a new Deployment", "Deployment.Namespace", rosKube.Namespace, "Deployment.Name", rosKube.Name)
			deploy := r.deploymentForRosKube(rosKube)
			if err = r.Create(ctx, deploy); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		} else {
			return ctrl.Result{}, err
		}
	}

	rosDomainId := rosKube.Spec.RosDomainId
	if rosDomainId != deployFound.Spec.Template.Spec.Containers[0].Env[0].Value {
		deployFound.Spec.Template.Spec.Containers[0].Env[0].Value = rosDomainId
		if err = r.Update(ctx, deployFound); err != nil {
			return ctrl.Result{}, err
		}
	}

	naviName := rosKube.Spec.NaviName
	if naviName != deployFound.Spec.Template.Spec.Containers[0].Env[1].Value {
		deployFound.Spec.Template.Spec.Containers[0].Env[1].Value = naviName
		if err = r.Update(ctx, deployFound); err != nil {
			return ctrl.Result{}, err
		}
	}

	network := rosKube.Spec.NetWork
	if network != deployFound.Spec.Template.ObjectMeta.Annotations["k8s.v1.cni.cncf.io/networks"] {
		deployFound.Annotations["k8s.v1.cni.cncf.io/networks"] = network
		deployFound.Spec.Template.ObjectMeta.Annotations["k8s.v1.cni.cncf.io/networks"] = network
		if err = r.Update(ctx, deployFound); err != nil {
			return ctrl.Result{}, err
		}
	}

	image := rosKube.Spec.Image
	if image != deployFound.Spec.Template.Spec.Containers[0].Image {
		deployFound.Spec.Template.Spec.Containers[0].Image = image
		if err = r.Update(ctx, deployFound); err != nil {
			return ctrl.Result{}, err
		}
	}

	nodeSelector := rosKube.Spec.NodeSelector
	if !reflect.DeepEqual(nodeSelector, deployFound.Spec.Template.Spec.NodeSelector) {
		deployFound.Spec.Template.Spec.NodeSelector = nodeSelector
		if err = r.Update(ctx, deployFound); err != nil {
			return ctrl.Result{}, err
		}
	}

	hostNetwork := rosKube.Spec.HostNetwork
	if hostNetwork != deployFound.Spec.Template.Spec.HostNetwork {
		deployFound.Spec.Template.Spec.HostNetwork = hostNetwork
		if err = r.Update(ctx, deployFound); err != nil {
			return ctrl.Result{}, err
		}
	}

	paramPath := rosKube.Spec.ParamPath
	if paramPath != deployFound.Spec.Template.Spec.Volumes[0].VolumeSource.HostPath.Path {
		deployFound.Spec.Template.Spec.Volumes[0].VolumeSource.HostPath.Path = paramPath
		if err = r.Update(ctx, deployFound); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// deploymentForRosKube returns a RosKube Deployment object
func (r *RosKubeReconciler) deploymentForRosKube(rosKube *lgecomv1alpha1.RosKube) *appsv1.Deployment {
	labels := map[string]string{
		"app": rosKube.Name,
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rosKube.Name,
			Namespace: rosKube.Namespace,
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
						"k8s.v1.cni.cncf.io/networks": rosKube.Spec.NetWork,
					},
				},
				Spec: corev1.PodSpec{
					HostNetwork: rosKube.Spec.HostNetwork,
					Containers: []corev1.Container{
						{
							Name:  rosKube.Name,
							Image: rosKube.Spec.Image,
							Args: []string{
								". /opt/ros/humble/setup.bash; . install/setup.bash; ros2 launch ros_kube ros_kube.launch.py kube_namespace:=" + rosKube.Namespace + " nav_pod_name:=" + rosKube.Spec.NaviName + " param_path:=/root/workspace/nav2/params/nav2_params.yaml" + " amcl_topic:=" + rosKube.Spec.AmclPose,
							},
							Env: []corev1.EnvVar{
								{
									Name:  "ROS_DOMAIN_ID",
									Value: rosKube.Spec.RosDomainId,
								},
								{
									Name:  "NAVI_NAME",
									Value: rosKube.Spec.NaviName,
								},
							},
							Command: []string{"/bin/bash", "-c"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "params",
									MountPath: "/root/workspace/nav2/params",
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1000m"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "params",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: rosKube.Spec.ParamPath,
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
					NodeSelector: rosKube.Spec.NodeSelector,
				},
			},
		},
	}

	// Set RosKube instance as the owner and controller
	ctrl.SetControllerReference(rosKube, dep, r.Scheme)
	return dep
}

// SetupWithManager sets up the controller with the Manager.
func (r *RosKubeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&lgecomv1alpha1.RosKube{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
