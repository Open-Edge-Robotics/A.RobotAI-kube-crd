/*
Copyright 2023.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	ebmev1 "github.com/seoyoungchae/navi-crd/api/v1"
)

// NavigationReconciler reconciles a Navigation object
type NavigationReconciler struct {
	client.Client
	initSpec *ebmev1.NavigationSpec
	Log      logr.Logger
	Scheme   *runtime.Scheme
}

// +kubebuilder:rbac:groups=ebme.lge.com,resources=navigations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ebme.lge.com,resources=navigations/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ebme.lge.com,resources=navigations/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch

func (r *NavigationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logMsg := log.FromContext(ctx)

	foundNodes := &corev1.NodeList{}
	opts := []client.ListOption{}
	if err := r.List(ctx, foundNodes, opts...); err != nil {
		logMsg.Error(err, "Failed to list nodes")
		return ctrl.Result{}, err
	}

	// Fetch the Navigation instance.
	navi := &ebmev1.Navigation{}
	err := r.Get(ctx, req.NamespacedName, navi)
	if err != nil {
		if errors.IsNotFound(err) {
			logMsg.Info("Navigation resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Check if the deployment already exists, if not create a new deployment.
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: navi.Name, Namespace: navi.Namespace}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			// Define and create a new deployment.
			dep := r.deploymentForNavigation(navi, &foundNodes.Items[0].Name)
			if err = r.Create(ctx, dep); err != nil {
				logMsg.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
				return ctrl.Result{}, err
			}
			r.initSpec = &navi.Spec
			return ctrl.Result{Requeue: true}, nil
		} else {
			logMsg.Error(err, "Failed to get Deployment")
			return ctrl.Result{}, err
		}
	}

	// Check if the initial spec has changed. Then restart the deployment.
	currentSpec := navi.Spec
	if !reflect.DeepEqual(r.initSpec, &currentSpec) {
		// Define and create a new deployment.
		if err = r.Delete(ctx, found); err != nil {
			logMsg.Error(err, "Failed to delete Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		dep := r.deploymentForNavigation(navi, &foundNodes.Items[0].Name)
		if err = r.Create(ctx, dep); err != nil {
			logMsg.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		r.initSpec = &navi.Spec
		return ctrl.Result{Requeue: true}, nil
	}

	// Update the navi status with the pod names.
	// List the pods for this CR's deployment.
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(navi.Namespace),
		client.MatchingLabels(labelsForApp(navi.Name)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		logMsg.Error(err, "Failed to list pods", "Navigation.Namespace", navi.Namespace, "Navigation.Name", navi.Name)
		return ctrl.Result{}, err
	}

	// Update status.Nodes if needed.
	podNames := getPodNames(podList.Items)
	if !reflect.DeepEqual(podNames, navi.Status.Nodes) {
		navi.Status.Nodes = podNames
		if err := r.Status().Update(ctx, navi); err != nil {
			logMsg.Error(err, "Failed to update Navigation status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// deploymentForNavigation returns a Deployment object for data from m.
func (r *NavigationReconciler) deploymentForNavigation(m *ebmev1.Navigation, defaultNode *string) *appsv1.Deployment {
	paramPath := "/opt/cloi_ws/src/navigation2/nav2_bringup/params/" + m.Spec.RobotName + "_nav2_params.yaml"

	lbls := labelsForApp(m.Name)
	replicas := int32(1)
	nodeName := defaultNode
	network := ""
	image := "lgecloudroboticstask/nav2_humble:seocho_b1_2"
	command := []string{"/bin/bash", "-c"}
	nvArgs := []string{
		"echo $ROBOT_NAME;" +
			"sed s/acloi00/${ROBOT_NAME}/g /opt/cloi_ws/src/navigation2/nav2_bringup/params/acloi00_nav2_params.yaml > " + paramPath + ";" +
			"sed -i '/ *\\<x: /s/.*/      x: " + m.Spec.InitialPose.Position.X + "/g ' " + paramPath + ";" +
			"sed -i '/ *\\<y: /s/.*/      y: " + m.Spec.InitialPose.Position.Y + "/g ' " + paramPath + ";" +
			"sed -i '/ *\\<z: /s/.*/      z: " + m.Spec.InitialPose.Position.Z + "/g ' " + paramPath + ";" +
			"sed -i '/ *\\<yaw: /s/.*/      yaw: " + m.Spec.InitialPose.Orientation.W + "/g ' " + paramPath + ";" +
			"ln -s " + paramPath + " /opt/cloi_ws/install/nav2_bringup/share/nav2_bringup/params/${ROBOT_NAME}_nav2_params.yaml;" +
			"export PARAMS_FILE=/opt/cloi_ws/install/nav2_bringup/share/nav2_bringup/params/${ROBOT_NAME}_nav2_params.yaml;" +
			"export MAP=/opt/cloi_ws/maps/seocho_tower_B1F.yaml;" +
			"export ROS_DOMAIN_ID=$((ROS_DOMAIN_ID));" +
			". /opt/ros/humble/setup.bash;" +
			". /opt/cloi_ws/install/setup.bash;" +
			"ros2 launch nav2_bringup bringup_launch.py map:=$MAP use_sim_time:=True use_composition:=False use_namespace:=True namespace:=$ROBOT_NAME params_file:=$PARAMS_FILE;",
	}

	if m.Spec.Network != "" {
		network = m.Spec.Network
	}

	if m.Spec.Image != "" {
		image = m.Spec.Image
	}

	if m.Spec.NodeSelector != "" {
		nodeName = &m.Spec.NodeSelector
	}

	if m.Spec.Command != nil {
		command = m.Spec.Command
	}

	if m.Spec.Args != nil {
		nvArgs = m.Spec.Args
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: lbls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: lbls,
					Annotations: map[string]string{
						"k8s.v1.cni.cncf.io/networks": network,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:   image,
						Name:    "navigation",
						Command: command,
						Args:    nvArgs,
						Env: []corev1.EnvVar{{
							Name:  "ROBOT_NAME",
							Value: m.Spec.RobotName,
						}, {
							Name:  "ROS_DOMAIN_ID",
							Value: m.Spec.RosDomainId,
						}},
						EnvFrom: []corev1.EnvFromSource{{
							ConfigMapRef: &corev1.ConfigMapEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "ebme-config",
								},
							},
						}},
						ImagePullPolicy: corev1.PullAlways,
					}},
					ImagePullSecrets: []corev1.LocalObjectReference{{
						// Name: "regcred",
						Name: "dockerhub-lgecloudroboticstask",
					}},
					NodeSelector: map[string]string{
						"kubernetes.io/hostname": *nodeName,
					},
				},
			},
		},
	}

	// Set Navigation instance as the owner and controller.memcac
	// NOTE: calling SetControllerReference, and setting owner references in
	// general, is important as it allows deleted objects to be garbage collected.
	controllerutil.SetControllerReference(m, dep, r.Scheme)
	return dep
}

// labelsForApp creates a simple set of labels for Navigation.
func labelsForApp(name string) map[string]string {
	return map[string]string{"cr_name": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

// SetupWithManager sets up the controller with the Manager.
func (r *NavigationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ebmev1.Navigation{}).
		Owns(&appsv1.Deployment{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 2}).
		Complete(r)
}
