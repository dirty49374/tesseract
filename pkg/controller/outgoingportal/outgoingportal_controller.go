package outgoingportal

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	tesseractv1alpha1 "github.com/dirty49374/tesseract/pkg/apis/tesseract/v1alpha1"

	"github.com/dirty49374/tesseract/pkg/certs"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_outgoingportal")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new OutgoingPortal Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	certs, err := certs.LoadCerts(".")
	if err != nil {
		return err
	}

	return add(mgr, newReconciler(mgr, certs))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, certs *certs.Certs) reconcile.Reconciler {
	return &ReconcileOutgoingPortal{client: mgr.GetClient(), scheme: mgr.GetScheme(), certs: certs}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("outgoingportal-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource OutgoingPortal
	err = c.Watch(&source.Kind{Type: &tesseractv1alpha1.OutgoingPortal{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Deployments and requeue the owner OutgoingPortal
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tesseractv1alpha1.OutgoingPortal{},
	})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Services and requeue the owner OutgoingPortal
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tesseractv1alpha1.OutgoingPortal{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileOutgoingPortal{}

// ReconcileOutgoingPortal reconciles a OutgoingPortal object
type ReconcileOutgoingPortal struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	certs  *certs.Certs
}

// Reconcile reads that state of the cluster for a OutgoingPortal object and makes changes based on the state read
// and what is in the OutgoingPortal.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileOutgoingPortal) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling OutgoingPortal")

	// Fetch the OutgoingPortal instance
	instance := &tesseractv1alpha1.OutgoingPortal{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if r.reconcileConfigMap(instance, reqLogger) != nil {
		return reconcile.Result{}, err
	}

	if r.reconcileSecret(instance, reqLogger) != nil {
		return reconcile.Result{}, err
	}

	if r.reconcileDeployment(instance, reqLogger) != nil {
		return reconcile.Result{}, err
	}

	if r.reconcileService(instance, reqLogger) != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileOutgoingPortal) reconcileConfigMap(instance *tesseractv1alpha1.OutgoingPortal, reqLogger logr.Logger) error {
	configmap := newConfigMapForCR(instance)

	// Set OutgoingPortal instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, configmap, r.scheme); err != nil {
		return err
	}

	// Check if this ConfigMap already exists
	found := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: configmap.Name, Namespace: configmap.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new ConfigMap", "ConfigMap.Namespace", configmap.Namespace, "ConfigMap.Name", configmap.Name)
		err = r.client.Create(context.TODO(), configmap)
		if err != nil {
			return err
		}
		// ConfigMap created successfully - don't requeue
		return nil
	} else if err != nil {
		return err
	}

	// ConfigMap already exists - don't requeue
	reqLogger.Info("Skip reconcile: ConfigMap already exists", "ConfigMap.Namespace", found.Namespace, "ConfigMap.Name", found.Name)
	return nil
}

func (r *ReconcileOutgoingPortal) reconcileSecret(instance *tesseractv1alpha1.OutgoingPortal, reqLogger logr.Logger) error {
	secret := newSecretForCR(instance, r.certs)

	// Set OutgoingPortal instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, secret, r.scheme); err != nil {
		return err
	}

	// Check if this Secret already exists
	found := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Secret", "Secret.Namespace", secret.Namespace, "Secret.Name", secret.Name)
		err = r.client.Create(context.TODO(), secret)
		if err != nil {
			return err
		}
		// Secret created successfully - don't requeue
		return nil
	} else if err != nil {
		return err
	}

	// Secret already exists - don't requeue
	reqLogger.Info("Skip reconcile: Secret already exists", "Secret.Namespace", found.Namespace, "Secret.Name", found.Name)
	return nil
}

func (r *ReconcileOutgoingPortal) reconcileDeployment(instance *tesseractv1alpha1.OutgoingPortal, reqLogger logr.Logger) error {
	deployment := newDeploymentForCR(instance)

	// Set OutgoingPortal instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, deployment, r.scheme); err != nil {
		return err
	}

	// Check if this Deployment already exists
	found := &appsv1.Deployment{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		err = r.client.Create(context.TODO(), deployment)
		if err != nil {
			return err
		}
		// Deployment created successfully - don't requeue
		return nil
	} else if err != nil {
		return err
	}

	// Deployment already exists - don't requeue
	reqLogger.Info("Skip reconcile: Deployment already exists", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
	return nil
}

func (r *ReconcileOutgoingPortal) reconcileService(instance *tesseractv1alpha1.OutgoingPortal, reqLogger logr.Logger) error {
	service := newServiceForCR(instance)

	// Set OutgoingPortal instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, service, r.scheme); err != nil {
		return err
	}

	// Check if this Service already exists
	found := &corev1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		err = r.client.Create(context.TODO(), service)
		if err != nil {
			return err
		}
		// Service created successfully - don't requeue
		return nil
	} else if err != nil {
		return err
	}

	// Service already exists - don't requeue
	reqLogger.Info("Skip reconcile: Service already exists", "Service.Namespace", found.Namespace, "Service.Name", found.Name)
	return nil
}

const envoyConfig = `
static_resources:
  listeners:
  - address:
      socket_address:
        address: 0.0.0.0
        port_value: 80
    filter_chains:
    - filters:
      - name: envoy.http_connection_manager
        typed_config:
          "@type": type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager
          codec_type: auto
          stat_prefix: ingress_http
          access_log:
          - name: envoy.file_access_log
            config:
              path: "/dev/stdout"
          route_config:
            name: local_route
            virtual_hosts:
            - name: remote
              domains:
              - "*"
              routes:
              - match:
                  prefix: /
                route:
                  host_rewrite: {{ .Spec.ServiceFDQN }}:{{ .Spec.ServicePort }}
                  cluster: remote
          http_filters:
          - name: envoy.router
            typed_config: {}
  clusters:
  - name: remote
    connect_timeout: 0.25s
    type: strict_dns
    lb_policy: round_robin
    http2_protocol_options: {}
    load_assignment:
      cluster_name: remote
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: {{ .Spec.Gateway }}
                port_value: 443
    tls_context:
      common_tls_context:
        tls_certificates:
          certificate_chain: { "filename": "/secret/client.crt" }
          private_key: { "filename": "/secret/client.key" }
        validation_context:
          trusted_ca:
            filename: /secret/ca.crt

admin:
  access_log_path: "/dev/stdout"
  address:
    socket_address:
      address: 0.0.0.0
      port_value: 8001

`

func newConfigMapForCR(cr *tesseractv1alpha1.OutgoingPortal) *corev1.ConfigMap {

	var buf bytes.Buffer

	template := template.Must(template.New("envoyConfig").Parse(envoyConfig))
	err := template.Execute(&buf, cr)
	if err != nil {
		fmt.Println("==================================================================")
		fmt.Println(err)
		fmt.Println("==================================================================")
		return nil
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-portal",
			Namespace: cr.Namespace,
			Labels: map[string]string{
				"app": cr.Name + "-portal",
			},
		},
		Data: map[string]string{
			"envoy.yaml": buf.String(),
		},
	}
}

func newSecretForCR(cr *tesseractv1alpha1.OutgoingPortal, certs *certs.Certs) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-portal",
			Namespace: cr.Namespace,
			Labels: map[string]string{
				"app": cr.Name + "-portal",
			},
		},
		Data: map[string][]byte{
			"ca.crt":     []byte(certs.TrustedCa),
			"client.crt": []byte(certs.Certificate),
			"client.key": []byte(certs.PrivateKey),
		},
	}
}

// newDeploymentForCR returns a busybox pod with the same name/namespace as the cr
func newDeploymentForCR(cr *tesseractv1alpha1.OutgoingPortal) *appsv1.Deployment {
	var replicas int32 = 1

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-portal",
			Namespace: cr.Namespace,
			Labels: map[string]string{
				"app": cr.Name + "-portal",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": cr.Name + "-portal",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": cr.Name + "-portal",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "busybox",
							Image: "envoyproxy/envoy:v1.10.0",
							Ports: []corev1.ContainerPort{
								{
									Name:          "proxy",
									ContainerPort: 80,
								},
								{
									Name:          "admin",
									ContainerPort: 8001,
								},
							},
							Command: []string{
								"/usr/local/bin/envoy",
								"-c",
								"/config/envoy.yaml",
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "config",
									MountPath: "/config",
								},
								{
									Name:      "secret",
									MountPath: "/secret",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-portal",
									},
								},
							},
						},
						{
							Name: "secret",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: cr.Name + "-portal",
								},
							},
						},
					},
				},
			},
		},
	}
}

func newServiceForCR(cr *tesseractv1alpha1.OutgoingPortal) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels: map[string]string{
				"app": cr.Name + "-portal",
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"app": cr.Name + "-portal",
			},
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name:     cr.Name,
					Protocol: corev1.ProtocolTCP,
					Port:     cr.Spec.ServicePort,
				},
			},
		},
	}
}
