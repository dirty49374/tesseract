package incomingportal

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	tesseractv1alpha1 "github.com/dirty49374/tesseract/pkg/apis/tesseract/v1alpha1"
	"github.com/dirty49374/tesseract/pkg/certs"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type Empty struct{}

var log = logf.Log.WithName("controller_incomingportal")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new IncomingPortal Controller and adds it to the Manager. The Manager will set fields on the Controller
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
	return &ReconcileIncomingPortal{client: mgr.GetClient(), scheme: mgr.GetScheme(), certs: certs, portals: make(map[string][]int32)}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("incomingportal-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource IncomingPortal
	err = c.Watch(&source.Kind{Type: &tesseractv1alpha1.IncomingPortal{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner IncomingPortal
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tesseractv1alpha1.IncomingPortal{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileIncomingPortal{}

// ReconcileIncomingPortal reconciles a IncomingPortal object
type ReconcileIncomingPortal struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client  client.Client
	scheme  *runtime.Scheme
	certs   *certs.Certs
	portals map[string][]int32
}

// Reconcile reads that state of the cluster for a IncomingPortal object and makes changes based on the state read
// and what is in the IncomingPortal.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileIncomingPortal) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	addr := request.Name + "." + request.Namespace

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling IncomingPortal")

	// Fetch the IncomingPortal instance
	instance := &tesseractv1alpha1.IncomingPortal{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			delete(r.portals, addr)

			configMap := r.newConfigMapForCR()
			err = r.client.Update(context.TODO(), configMap)
			if err != nil {
				return reconcile.Result{}, err
			}

			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	r.portals[addr] = instance.Spec.ServicePorts
	if r.portals[addr] == nil {
		r.portals[addr] = make([]int32, 0)
	}
	configMap := r.newConfigMapForCR()
	err = r.client.Update(context.TODO(), configMap)
	if err != nil {
		return reconcile.Result{}, err
	}

	// // Define a new Pod object
	// pod := newPodForCR(instance)

	// // Set IncomingPortal instance as the owner and controller
	// if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
	// 	return reconcile.Result{}, err
	// }

	// // Check if this Pod already exists
	// found := &corev1.Pod{}
	// err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	// if err != nil && errors.IsNotFound(err) {
	// 	reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
	// 	err = r.client.Create(context.TODO(), pod)
	// 	if err != nil {
	// 		return reconcile.Result{}, err
	// 	}

	// 	// Pod created successfully - don't requeue
	// 	return reconcile.Result{}, nil
	// } else if err != nil {
	// 	return reconcile.Result{}, err
	// }

	// Pod already exists - don't requeue
	// reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	return reconcile.Result{}, nil
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
            {{- range $name, $ports := . }}
            {{- range $ports }}
            - name: "{{ $name }}-{{ . }}"
              domains:
              - "{{ $name }}:{{ . }}"
              {{- if eq . 80 }}
              - "{{ $name }}"
              {{- end }}
              routes:
              - match:
                  prefix: /
                route:
                  cluster: "{{ $name }}-{{ . }}"
            {{- end }}
            {{- end }}
          http_filters:
          - name: envoy.router
            typed_config: {}
  clusters:
  {{- range $name, $ports := . }}
  {{- range $ports }}
  - name: "{{ $name }}-{{ . }}"
    connect_timeout: 0.25s
    type: strict_dns
    lb_policy: round_robin
    load_assignment:
      cluster_name: "{{ $name }}-{{ . }}"
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: {{ $name }}
                port_value: {{ . }}
  {{- end }}
  {{- end }}
admin:
  access_log_path: "/dev/stdout"
  address:
    socket_address:
      address: 0.0.0.0
      port_value: 8001

`

func (r *ReconcileIncomingPortal) newConfigMapForCR() *corev1.ConfigMap {

	var buf bytes.Buffer

	template := template.Must(template.New("envoyConfig").Parse(envoyConfig))
	err := template.Execute(&buf, r.portals)
	if err != nil {
		fmt.Println("==================================================================")
		fmt.Println(err)
		fmt.Println("==================================================================")
		return nil
	}

	yaml := buf.String()
	fmt.Println(yaml)

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tesseract",
			Namespace: "tess",
		},
		Data: map[string]string{
			"envoy.yaml": yaml,
		},
	}
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *tesseractv1alpha1.IncomingPortal) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}
