/*
Copyright 2021.

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
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	postgresv1 "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"

	"github.com/go-logr/logr"
	protoTypes "github.com/gogo/protobuf/types"
	reconcilehelper "github.com/pluralsh/controller-reconcile-helper"
	mlflowv1alpha1 "github.com/pluralsh/mlflow-operator/api/v1alpha1"
	istioNetworking "istio.io/api/networking/v1beta1"
	istioNetworkingClient "istio.io/client-go/pkg/apis/networking/v1beta1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
)

/*
We generally want to ignore (not requeue) NotFound errors, since we'll get a
reconciliation request once the object exists, and requeuing in the meantime
won't help.
*/
func ignoreNotFound(err error) error {
	if apierrs.IsNotFound(err) {
		return nil
	}
	return err
}

// TrackingServerReconciler reconciles a TrackingServer object
type TrackingServerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=acid.zalan.do,resources=postgresqls,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mlflow.plural.sh,resources=trackingservers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mlflow.plural.sh,resources=trackingservers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mlflow.plural.sh,resources=trackingservers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TrackingServer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *TrackingServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// _ = log.FromContext(ctx)
	log := r.Log.WithValues("TrackingServer", req.NamespacedName)

	tsInstance := &mlflowv1alpha1.TrackingServer{}
	if err := r.Get(ctx, req.NamespacedName, tsInstance); err != nil {
		if apierrs.IsNotFound(err) {
			log.Info("Unable to fetch tracking server - skipping", "namespace", tsInstance.Namespace, "name", tsInstance.Name)
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch tracking server")
		return ctrl.Result{}, ignoreNotFound(err)
	}

	// Reconcile PVC for the Tracking Server
	pvc := r.generatePVC(tsInstance)
	if err := ctrl.SetControllerReference(tsInstance, pvc, r.Scheme); err != nil {
		log.Error(err, "Error setting ControllerReference for PersistentVolumeClaim")
		return ctrl.Result{}, err
	}
	if err := reconcilehelper.PersistentVolumeClaim(ctx, r.Client, pvc, log); err != nil {
		log.Error(err, "Error reconciling PersistentVolumeClaim", "pvc", pvc.Name, "namespace", pvc.Namespace)
		return ctrl.Result{}, err
	}

	// Reconcile PostgreSQL database for the Tracking Server
	postgres := r.generatePostgres(tsInstance)
	if err := ctrl.SetControllerReference(tsInstance, postgres, r.Scheme); err != nil {
		log.Error(err, "Error setting ControllerReference for PostgreSQL")
		return ctrl.Result{}, err
	}
	if err := reconcilehelper.Postgresql(ctx, r.Client, postgres, log); err != nil {
		log.Error(err, "Error reconciling PostgreSQL", "postgres", postgres.Name, "namespace", postgres.Namespace)
		return ctrl.Result{}, err
	}

	// Reconcile Services for the PostgreSQL database
	postgresServices := r.generatePostgresService(tsInstance)
	for _, postgresService := range postgresServices.Items {
		service := &postgresService
		if err := ctrl.SetControllerReference(tsInstance, service, r.Scheme); err != nil {
			log.Error(err, "Error setting ControllerReference for PostgreSQL Service")
			return ctrl.Result{}, err
		}
		if err := reconcilehelper.Service(ctx, r.Client, service, log); err != nil {
			log.Error(err, "Error reconciling PostgreSQL Service", "service", service.Name, "namespace", service.Namespace)
			return ctrl.Result{}, err
		}
	}

	// Reconcile StatefulSet for the Tracking Server
	statefulset := r.generateStatefulset(tsInstance)
	if err := ctrl.SetControllerReference(tsInstance, statefulset, r.Scheme); err != nil {
		log.Error(err, "Error setting ControllerReference for StatefulSet")
		return ctrl.Result{}, err
	}
	if err := reconcilehelper.StatefulSet(ctx, r.Client, statefulset, log); err != nil {
		log.Error(err, "Error reconciling StatefulSet", "statefulset", statefulset.Name, "namespace", statefulset.Namespace)
		return ctrl.Result{}, err
	}

	// Reconcile Service for the Tracking Server
	service := r.generateService(tsInstance)
	if err := ctrl.SetControllerReference(tsInstance, service, r.Scheme); err != nil {
		log.Error(err, "Error setting ControllerReference for Service")
		return ctrl.Result{}, err
	}
	if err := reconcilehelper.Service(ctx, r.Client, service, log); err != nil {
		log.Error(err, "Error reconciling Service", "service", service.Name, "namespace", service.Namespace)
		return ctrl.Result{}, err
	}

	if tsInstance.Spec.Network.IstioEnabled {
		// Reconcile VirtualService for the Tracking Server
		virtualService := r.generateVirtualService(tsInstance)
		if err := ctrl.SetControllerReference(tsInstance, virtualService, r.Scheme); err != nil {
			log.Error(err, "Error setting ControllerReference for VirtualService")
			return ctrl.Result{}, err
		}
		if err := reconcilehelper.VirtualService(ctx, r.Client, virtualService, log); err != nil {
			log.Error(err, "Error reconciling VirtualService", "virtualservice", virtualService.Name, "namespace", virtualService.Namespace)
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TrackingServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mlflowv1alpha1.TrackingServer{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.Service{}).
		Owns(&postgresv1.Postgresql{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&istioNetworkingClient.VirtualService{}).
		Complete(r)
}

func (r *TrackingServerReconciler) generatePVC(tsInstance *mlflowv1alpha1.TrackingServer) *corev1.PersistentVolumeClaim {

	var storageClassName *string

	if tsInstance.Spec.StorageClass != "" {
		storageClassName = &tsInstance.Spec.StorageClass
	} else {
		storageClassName = nil
	}

	size := tsInstance.Spec.Size
	persistentVolumeClaim := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tsInstance.Name,
			Namespace: tsInstance.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       tsInstance.Name,
				"app.kubernetes.io/managed-by": "mlflow-operator",
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(size),
				},
			},
			StorageClassName: storageClassName,
		},
	}
	return persistentVolumeClaim
}

func (r *TrackingServerReconciler) generatePostgres(tsInstance *mlflowv1alpha1.TrackingServer) *postgresv1.Postgresql {

	resources := tsInstance.Spec.Postgres.Resources

	var RequestsCPU string
	var RequestsMem string
	var LimitCPU string
	var LimitMem string

	if resources.Requests.CPU == "" {
		RequestsCPU = "100m"
	} else {
		RequestsCPU = resources.Requests.CPU
	}

	if resources.Requests.Memory == "" {
		RequestsMem = "100Mi"
	} else {
		RequestsMem = resources.Requests.Memory
	}

	if resources.Limits.CPU == "" {
		LimitCPU = "1.5"
	} else {
		LimitCPU = resources.Limits.CPU
	}

	if resources.Limits.Memory == "" {
		LimitMem = "1Gi"
	} else {
		LimitMem = resources.Limits.Memory
	}

	postgres := &postgresv1.Postgresql{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("plural-postgres-mlflow-%s", tsInstance.Name),
			Namespace: tsInstance.Namespace,
		},
		Spec: postgresv1.PostgresSpec{
			TeamID: "plural",
			Volume: postgresv1.Volume{
				Size:         tsInstance.Spec.Postgres.Volume.Size,
				StorageClass: tsInstance.Spec.Postgres.Volume.StorageClass,
			},
			NumberOfInstances: tsInstance.Spec.Postgres.Instances,
			Users:             map[string]postgresv1.UserFlags{"mlflow": []string{"superuser", "createdb"}},
			Databases:         map[string]string{"mlflow": "mlflow"},
			PostgresqlParam: postgresv1.PostgresqlParam{
				PgVersion: tsInstance.Spec.Postgres.Version,
			},
			PodAnnotations: map[string]string{
				"sidecar.istio.io/inject": "false",
			},
			Resources: postgresv1.Resources{
				ResourceRequests: postgresv1.ResourceDescription{
					CPU:    RequestsCPU,
					Memory: RequestsMem,
				},
				ResourceLimits: postgresv1.ResourceDescription{
					CPU:    LimitCPU,
					Memory: LimitMem,
				},
			},
			Sidecars: []postgresv1.Sidecar{
				{
					DockerImage: "gcr.io/pluralsh/postgres-exporter:0.8.0",
					Name:        "exporter",
					Ports: []corev1.ContainerPort{
						{
							Name:          "http-metrics",
							ContainerPort: 9187,
							Protocol:      corev1.ProtocolTCP,
						},
					},
					Env: []corev1.EnvVar{
						{
							Name:  "DATA_SOURCE_URI",
							Value: "127.0.0.1:5432/mlflow?sslmode=disable",
						},
						{
							Name: "DATA_SOURCE_USER",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: fmt.Sprintf("postgres.plural-postgres-mlflow-%s.credentials.postgresql.acid.zalan.do", tsInstance.Name),
									},
									Key: "username",
								},
							},
						},
						{
							Name: "DATA_SOURCE_PASS",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: fmt.Sprintf("postgres.plural-postgres-mlflow-%s.credentials.postgresql.acid.zalan.do", tsInstance.Name),
									},
									Key: "password",
								},
							},
						},
					},
				},
			},
		},
	}
	return postgres
}

// Generate the desired Service object for the workspace
func (r *TrackingServerReconciler) generatePostgresService(tsInstance *mlflowv1alpha1.TrackingServer) *corev1.ServiceList {

	services := &corev1.ServiceList{
		Items: []corev1.Service{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("plural-postgres-mlflow-%s-master", tsInstance.Name),
					Namespace: tsInstance.Namespace,
					Labels:    map[string]string{"spilo-role": "master"},
				},
				Spec: corev1.ServiceSpec{
					Type:     "ClusterIP",
					Selector: map[string]string{"spilo-role": "master"},
					Ports: []corev1.ServicePort{
						{
							Name:       "postgres",
							Port:       5432,
							TargetPort: intstr.FromInt(5432),
							Protocol:   corev1.ProtocolTCP,
						},
						{
							Name:       "http-metrics",
							Port:       9187,
							TargetPort: intstr.FromString("http-metrics"),
							Protocol:   corev1.ProtocolTCP,
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("plural-postgres-mlflow-%s-replica", tsInstance.Name),
					Namespace: tsInstance.Namespace,
					Labels:    map[string]string{"spilo-role": "replica"},
				},
				Spec: corev1.ServiceSpec{
					Type:     "ClusterIP",
					Selector: map[string]string{"spilo-role": "replica"},
					Ports: []corev1.ServicePort{
						{
							Name:       "postgres",
							Port:       5432,
							TargetPort: intstr.FromInt(5432),
							Protocol:   corev1.ProtocolTCP,
						},
						{
							Name:       "http-metrics",
							Port:       9187,
							TargetPort: intstr.FromString("http-metrics"),
							Protocol:   corev1.ProtocolTCP,
						},
					},
				},
			},
		},
	}
	return services
}

func (r *TrackingServerReconciler) generateStatefulset(tsInstance *mlflowv1alpha1.TrackingServer) *appsv1.StatefulSet {
	replicas := tsInstance.Spec.Replicas
	labels := map[string]string{
		"statefulset": tsInstance.Name,
	}
	if tsInstance.Spec.ExtraPodLabels != nil {
		for k, v := range tsInstance.Spec.ExtraPodLabels {
			labels[k] = v
		}
	}
	container := []corev1.Container{
		{
			Image: tsInstance.Spec.Image,
			Name:  tsInstance.Name,
			Args: []string{
				"mlflow",
				"server",
				"--backend-store-uri",
				"$(BACKEND_STORE_URI)",
				"--default-artifact-root",
				"$(ARTIFACT_ROOT)",
				"--host",
				"0.0.0.0",
				"--port",
				"5000",
			},
			Env: []corev1.EnvVar{
				{
					Name:  "POSTGRES_MASTER_SERVICE",
					Value: fmt.Sprintf("plural-postgres-mlflow-%s-master", tsInstance.Name),
				},
				{
					Name: "POSTGRES_USERNAME",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: fmt.Sprintf("mlflow.plural-postgres-mlflow-%s.credentials.postgresql.acid.zalan.do", tsInstance.Name),
							},
							Key: "username",
						},
					},
				},
				{
					Name: "POSTGRES_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: fmt.Sprintf("mlflow.plural-postgres-mlflow-%s.credentials.postgresql.acid.zalan.do", tsInstance.Name),
							},
							Key: "password",
						},
					},
				},
				{
					Name:  "BACKEND_STORE_URI",
					Value: "postgresql://$(POSTGRES_USERNAME):$(POSTGRES_PASSWORD)@$(POSTGRES_MASTER_SERVICE)/mlflow",
				},
				{
					Name:  "ARTIFACT_ROOT",
					Value: tsInstance.Spec.DefaultArtifactRoot,
				},
			},
			Ports: []corev1.ContainerPort{
				{
					ContainerPort: 5000,
					Name:          "trackingserver",
					Protocol:      corev1.ProtocolTCP,
				},
			},
			VolumeMounts: []corev1.VolumeMount{
				{
					MountPath: "/mnt/mlruns",
					Name:      "files-mlflow",
				},
			},
		},
	}

	if len(tsInstance.Spec.S3secretName) != 0 {
		container[0].EnvFrom = []corev1.EnvFromSource{{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: tsInstance.Spec.S3secretName,
				},
			},
		}}
	}

	if len(tsInstance.Spec.S3endpointURL) != 0 {
		container[0].Env = []corev1.EnvVar{
			{
				Name:  "MLFLOW_S3_ENDPOINT_URL",
				Value: tsInstance.Spec.S3endpointURL,
			},
		}
	}

	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tsInstance.Name,
			Namespace: tsInstance.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: container,
					Volumes: []corev1.Volume{
						{
							Name: "files-mlflow",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: tsInstance.Name,
								},
							},
						},
					},
					SecurityContext:    &corev1.PodSecurityContext{},
					ServiceAccountName: tsInstance.Spec.ServiceAccountName,
				},
			},
		},
	}

	if len(tsInstance.Spec.ImagePullSecret) != 0 {
		statefulset.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
			{
				Name: tsInstance.Spec.ImagePullSecret,
			},
		}
	}
	return statefulset
}

// Generate the desired Service object for the workspace
func (r *TrackingServerReconciler) generateService(tsInstance *mlflowv1alpha1.TrackingServer) *corev1.Service {

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tsInstance.Name,
			Namespace: tsInstance.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Type:     "ClusterIP",
			Selector: map[string]string{"statefulset": tsInstance.Name},
			Ports: []corev1.ServicePort{
				{
					Name:       "http-trackingserver",
					Port:       5000,
					TargetPort: intstr.FromInt(5000),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
	return svc
}

// Generate the desired VirtualService object for the workspace
func (r *TrackingServerReconciler) generateVirtualService(tsInstance *mlflowv1alpha1.TrackingServer) *istioNetworkingClient.VirtualService {

	service := fmt.Sprintf("%s.%s.svc.cluster.local", tsInstance.Name, tsInstance.Namespace)

	httpRoutes := []*istioNetworking.HTTPRoute{}

	httpRoute := &istioNetworking.HTTPRoute{
		Match: []*istioNetworking.HTTPMatchRequest{
			{
				Uri: &istioNetworking.StringMatch{
					MatchType: &istioNetworking.StringMatch_Prefix{
						Prefix: fmt.Sprintf("/mlflow/%s/%s", tsInstance.Namespace, tsInstance.Name),
					},
				},
			},
		},
		Rewrite: &istioNetworking.HTTPRewrite{
			Uri: "/",
		},
		Route: []*istioNetworking.HTTPRouteDestination{
			{
				Destination: &istioNetworking.Destination{
					Host: service,
					Port: &istioNetworking.PortSelector{
						Number: uint32(5000),
					},
				},
			},
		},
		Timeout: &protoTypes.Duration{
			Seconds: 300,
		},
	}
	httpRoutes = append(httpRoutes, httpRoute)

	virtualservice := &istioNetworkingClient.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("mlflow-%s-%s", tsInstance.Namespace, tsInstance.Name),
			Namespace: tsInstance.Namespace,
		},
		Spec: istioNetworking.VirtualService{
			Hosts: []string{
				"*",
			},
			Gateways: []string{
				fmt.Sprintf("%s/%s", tsInstance.Spec.Network.IstioGatewayNamespace, tsInstance.Spec.Network.IstioGatewayName),
			},
			Http: httpRoutes,
		},
	}
	return virtualservice
}
