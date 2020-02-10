package kubevirtaddon

import (
	"context"

	appv1alpha1 "github.ibm.com/steve-kim-ibm/kubevirt-addon/pkg/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	vmiv1 "kubevirt.io/client-go/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_kubevirtaddon")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new KubevirtAddon Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKubevirtAddon{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("kubevirtaddon-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// // Watch for changes to primary resource KubevirtAddon
	// err = c.Watch(&source.Kind{Type: &appv1alpha1.KubevirtAddon{}}, &handler.EnqueueRequestForObject{})
	// if err != nil {
	// 	return err
	// }

	// Watch for changes to primary resource KubevirtAddon
	err = c.Watch(&source.Kind{Type: &vmiv1.VirtualMachineInstance{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner KubevirtAddon
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.KubevirtAddon{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileKubevirtAddon implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileKubevirtAddon{}

// ReconcileKubevirtAddon reconciles a KubevirtAddon object
type ReconcileKubevirtAddon struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a KubevirtAddon object and makes changes based on the state read
// and what is in the KubevirtAddon.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileKubevirtAddon) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling VMI")

	// Fetch the VMI instance
	instance := &vmiv1.VirtualMachineInstance{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	service := newServiceForVMI(instance)
	secret := newSecretforVMI(instance)

	foundSvc := &corev1.Service{}
	foundSecret := &corev1.Secret{}

	// Check if service already exists
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, foundSvc)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating new service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		err = r.client.Create(context.TODO(), service)
		if err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Check if secret already exists
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, foundSecret)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating new secret", "Secret.Namespace", secret.Namespace, "Secret.Name", secret.Name)
		err = r.client.Create(context.TODO(), secret)
		if err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func newServiceForVMI(vmi *vmiv1.VirtualMachineInstance) *corev1.Service {
	labels := map[string]string{
		"app": vmi.Name,
	}
	targetPort := intstr.IntOrString{
		IntVal: 22,
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vmi.Name + "-service",
			Namespace: vmi.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       27017,
					NodePort:   30000,
					Protocol:   "TCP",
					TargetPort: targetPort,
				},
			},
			Selector: vmi.Labels,
			Type:     "NodePort",
		},
	}
}

func newSecretforVMI(vmi *vmiv1.VirtualMachineInstance) *corev1.Secret {
	labels := map[string]string{
		"app": vmi.Name,
	}
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vmi.Name + "-service",
			Namespace: vmi.Namespace,
			Labels:    labels,
		},
		StringData: map[string]string{
			"username": "cirros",
			"password": "gocubsgo",
		},
	}
}
