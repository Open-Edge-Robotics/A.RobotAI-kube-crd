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
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ebmev1 "cloud-bridge-crd/api/v1"
)

// CloudBridgeReconciler reconciles a CloudBridge object
type CloudBridgeReconciler struct {
	initSpec *ebmev1.CloudBridgeSpec
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ebme.lge.com,resources=cloudbridges,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ebme.lge.com,resources=cloudbridges/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ebme.lge.com,resources=cloudbridges/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch

func (r *CloudBridgeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logMsg := log.FromContext(ctx)

	foundNodes := &corev1.NodeList{}
	opts := []client.ListOption{}
	if err := r.List(ctx, foundNodes, opts...); err != nil {
		logMsg.Error(err, "Failed to list nodes")
		return ctrl.Result{}, err
	}

	// Fetch the CloudBridge instance.
	cb := &ebmev1.CloudBridge{}
	err := r.Get(ctx, req.NamespacedName, cb)
	if err != nil {
		if errors.IsNotFound(err) {
			logMsg.Info("CloudBridge resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Check if the deployment already exists, if not create a new deployment.
	foundDeployment := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: cb.Name, Namespace: cb.Namespace}, foundDeployment)
	if err != nil {
		if errors.IsNotFound(err) {
			// Define and create a new deployment.
			dep := r.deploymentForCloudBridge(cb, &foundNodes.Items[0].Name)
			if err = r.Create(ctx, dep); err != nil {
				logMsg.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
				return ctrl.Result{}, err
			}
			logMsg.Info("Created a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			r.initSpec = &cb.Spec
			return ctrl.Result{Requeue: true}, nil
		} else {
			return ctrl.Result{}, err
		}
	}

	// Check if the service already exists, if not create a new service.
	foundService := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: cb.Name + "-svc", Namespace: cb.Namespace}, foundService)
	if err != nil {
		if errors.IsNotFound(err) {
			// Define and create a new service.
			svc := r.serviceForCloudBridge(cb)
			if err = r.Create(ctx, svc); err != nil {
				logMsg.Error(err, "Failed to create new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
				return ctrl.Result{}, err
			}
			logMsg.Info("Created a new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
			return ctrl.Result{Requeue: true}, nil
		} else {
			logMsg.Error(err, "Failed to get Service")
			return ctrl.Result{}, err
		}
	}

	// Check if the initial spec has changed. Then restart the deployment.
	currentSpec := cb.Spec
	if !reflect.DeepEqual(r.initSpec, &currentSpec) {
		// Update the deployment.
		if err = r.Update(ctx, foundDeployment); err != nil {
			logMsg.Error(err, "Failed to update Deployment", "Deployment.Namespace", foundDeployment.Namespace, "Deployment.Name", foundDeployment.Name)
			return ctrl.Result{}, err
		}
		logMsg.Info("Updated Deployment", "Deployment.Namespace", foundDeployment.Namespace, "Deployment.Name", foundDeployment.Name)

		// Update the service.
		if err = r.Update(ctx, foundService); err != nil {
			logMsg.Error(err, "Failed to update Service", "Service.Namespace", foundService.Namespace, "Service.Name", foundService.Name)
			return ctrl.Result{}, err
		}
		logMsg.Info("Updated Service", "Service.Namespace", foundService.Namespace, "Service.Name", foundService.Name)

		/** Hold on to the code for a while.
		// Define and create a new deployment.
		if err = r.Delete(ctx, foundDeployment); err != nil {
			logMsg.Error(err, "Failed to delete Deployment", "Deployment.Namespace", foundDeployment.Namespace, "Deployment.Name", foundDeployment.Name)
			return ctrl.Result{}, err
		}
		dep := r.deploymentForCloudBridge(cb, &foundNodes.Items[0].Name)
		if err = r.Create(ctx, dep); err != nil {
			logMsg.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		logMsg.Info("Created a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		*/
		r.initSpec = &cb.Spec
		return ctrl.Result{Requeue: true}, nil
	}

	// Update the navi status with the pod names.
	// List the pods for this CR's deployment.
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(cb.Namespace),
		client.MatchingLabels(labelsForApp(cb.Name)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		logMsg.Error(err, "Failed to list pods", "CloudBridge.Namespace", cb.Namespace, "CloudBridge.Name", cb.Name)
		return ctrl.Result{}, err
	}

	// Update status.Nodes if needed.
	podNames := getPodNames(podList.Items)
	if !reflect.DeepEqual(podNames, cb.Status.Nodes) {
		cb.Status.Nodes = podNames
		if err := r.Status().Update(ctx, cb); err != nil {
			logMsg.Error(err, "Failed to update CloudBridge status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// deploymentForCloudBridge returns a Deployment object for data from m.
func (r *CloudBridgeReconciler) deploymentForCloudBridge(m *ebmev1.CloudBridge, defaultNode *string) *appsv1.Deployment {
	logMsg := log.Log.WithName("deploymentForCloudBridge")

	lbls := labelsForApp(m.Name)
	replicas := int32(1)
	network := ""
	port := ebmev1.NewCloudBridgePort(m.Spec.Port)
	image := "lgecloudroboticstask/cloud_bridge:humble-offloading-20230508"
	configDir := ""
	command := []string{"/bin/bash", "-c"}
	cbArgs := []string{
		"tail -f /dev/null",
	}
	cloudIP := ""
	nodeSelector := defaultNode

	if m.Spec.Network != "" {
		network = m.Spec.Network
	}

	if m.Spec.Image != "" {
		image = m.Spec.Image
	}

	if m.Spec.ConfigDir != "" {
		configDir = m.Spec.ConfigDir
	}

	if m.Spec.Command != nil {
		command = m.Spec.Command
	}

	if m.Spec.Args != nil {
		cbArgs = m.Spec.Args
	}

	if m.Spec.CloudIP != "" {
		cloudIP = m.Spec.CloudIP
	}

	if m.Spec.NodeSelector != "" {
		nodeSelector = &m.Spec.NodeSelector
	}

	logMsg.Info("deploymentForCloudBridge", "RobotName", m.Spec.RobotName, "RosDomainId", m.Spec.RosDomainId)
	logMsg.Info("DeploymentForCloudBridge", "Image", image, "Port", m.Spec.Port, "ConfigDir", configDir, "NodeSelector", nodeSelector)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
			Labels:    lbls,
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
					Containers: []corev1.Container{
						{
							Image:      image,
							Name:       m.Name,
							WorkingDir: "/home/ubuntu/cloud_bridge",
							Command:    command,
							Args:       cbArgs,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: port.ContainerPort[ContainerS],
									Name:          ContainerS,
								},
								{
									ContainerPort: port.ContainerPort[ContainerP1],
									Name:          ContainerP1,
								},
								{
									ContainerPort: port.ContainerPort[ContainerP2],
									Name:          ContainerP2,
								},
								{
									ContainerPort: port.ContainerPort[ContainerP3],
									Name:          ContainerP3,
								},
								{
									ContainerPort: port.ContainerPort[ContainerP4],
									Name:          ContainerP4,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "ROBOT_NAME",
									Value: m.Spec.RobotName,
								},
								{
									Name:  "CONFIG_DIR",
									Value: configDir,
								},
								{
									Name:  "ROS_DOMAIN_ID",
									Value: m.Spec.RosDomainId,
								},
								{
									Name:  "CLOUD_IP",
									Value: cloudIP,
								},
								{
									Name:  HostSubPort,
									Value: strconv.Itoa(int(port.HostPort[HostSubPort])),
								},
								{
									Name:  HostPubPort,
									Value: strconv.Itoa(int(port.HostPort[HostPubPort])),
								},
								{
									Name:  HostReqPort,
									Value: strconv.Itoa(int(port.HostPort[HostReqPort])),
								},
								{
									Name:  HostRepPort,
									Value: strconv.Itoa(int(port.HostPort[HostRepPort])),
								},
							},
							EnvFrom: []corev1.EnvFromSource{{
								ConfigMapRef: &corev1.ConfigMapEnvSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "ebme-config",
									},
								},
							}},
							ImagePullPolicy: corev1.PullIfNotPresent,
						},
						{
							Name:            "cloud-bridge-monitor",
							Image:           "lgecloudroboticstask/alpine_cloud_bridge_monitor:latest",
							ImagePullPolicy: corev1.PullIfNotPresent,
							WorkingDir:      "/",
							Command:         []string{"/cloud_bridge_monitor.sh"},
						},
					},
					NodeSelector: map[string]string{
						"kubernetes.io/hostname": *nodeSelector,
					},
					ImagePullSecrets: []corev1.LocalObjectReference{{
						// Name: "regcred",
						Name: "dockerhub-lgecloudroboticstask",
					}},
					HostNetwork: false,
					Hostname:    m.Name,
				},
			},
		},
	}

	controllerutil.SetControllerReference(m, dep, r.Scheme)
	return dep
}

func (r *CloudBridgeReconciler) serviceForCloudBridge(m *ebmev1.CloudBridge) *corev1.Service {
	lbls := labelsForApp(m.Name + "-svc")
	port := ebmev1.NewCloudBridgePort(m.Spec.Port)

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name + "-svc",
			Namespace: m.Namespace,
			Labels:    lbls,
		},
		Spec: corev1.ServiceSpec{
			Selector: labelsForApp(m.Name),
			Ports: []corev1.ServicePort{
				{
					Name:       "lb-s",
					Port:       port.HostPort[HostPort],
					TargetPort: intstr.FromInt(int(port.ContainerPort[ContainerS])),
					NodePort:   port.HostPort[HostPort],
				},
				{
					Name:       "lb-p1",
					Port:       port.HostPort[HostSubPort],
					TargetPort: intstr.FromInt(int(port.ContainerPort[ContainerP1])),
					NodePort:   port.HostPort[HostSubPort],
				},
				{
					Name:       "lb-p2",
					Port:       port.HostPort[HostPubPort],
					TargetPort: intstr.FromInt(int(port.ContainerPort[ContainerP2])),
					NodePort:   port.HostPort[HostPubPort],
				},
				{
					Name:       "lb-p3",
					Port:       port.HostPort[HostReqPort],
					TargetPort: intstr.FromInt(int(port.ContainerPort[ContainerP3])),
					NodePort:   port.HostPort[HostReqPort],
				},
				{
					Name:       "lb-p4",
					Port:       port.HostPort[HostRepPort],
					TargetPort: intstr.FromInt(int(port.ContainerPort[ContainerP4])),
					NodePort:   port.HostPort[HostRepPort],
				},
			},
			Type: corev1.ServiceTypeNodePort,
		},
	}

	controllerutil.SetControllerReference(m, svc, r.Scheme)
	return svc
}

// labelsForApp creates a simple set of labels for CloudBridge.
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
func (r *CloudBridgeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ebmev1.CloudBridge{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 2}).
		Complete(r)
}
