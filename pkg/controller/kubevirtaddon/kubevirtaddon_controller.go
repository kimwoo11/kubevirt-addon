package kubevirtaddon

import (
	"context"

	ocpv1 "github.com/openshift/api/route/v1"
	appv1alpha1 "github.ibm.com/steve-kim-ibm/kubevirt-addon/pkg/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	vmiv1 "kubevirt.io/client-go/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

	// Watch for changes to primary resource KubevirtAddon
	err = c.Watch(&source.Kind{Type: &appv1alpha1.KubevirtAddon{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// // Watch for changes to primary resource KubevirtAddon
	// err = c.Watch(&source.Kind{Type: &vmiv1.VirtualMachineInstance{}}, &handler.EnqueueRequestForObject{})
	// if err != nil {
	// 	return err
	// }

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner KubevirtAddon
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
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
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileKubevirtAddon) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling KubevirtAddon")

	// Fetch KubevirtAddon instance
	instance := &appv1alpha1.KubevirtAddon{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	vmi := &vmiv1.VirtualMachineInstance{}
	err = r.client.Get(context.Background(), client.ObjectKey{
		Name:      instance.Spec.VMI.Name,
		Namespace: instance.Spec.VMI.Namespace,
	}, vmi)

	generate := instance.Spec.Generate
	if generate != nil {
		if len(generate.Services) > 0 {
			for _, svcSpec := range generate.Services {
				if err != nil {
					if errors.IsNotFound(err) {
						return reconcile.Result{}, nil
					}
					return reconcile.Result{}, err
				}
				if len(svcSpec.Labels)+len(svcSpec.Selector) == 0 {
					reqLogger.Info("Labels and selectors not defined")
				}
				reqLogger.Info("Generating service " + svcSpec.Name)
				svc := generateService(&svcSpec, vmi)
				if err != nil {
					return reconcile.Result{}, err
				}
				err = r.client.Create(context.TODO(), svc)
				if err != nil {
					if errors.IsAlreadyExists(err) {
						reqLogger.Info("Service already exists")
					} else {
						return reconcile.Result{}, err
					}
				}
			}
		}
		if len(generate.Routes) > 0 {
			for _, routeSpec := range generate.Routes {
				svc := &corev1.Service{}
				err := r.client.Get(context.Background(), client.ObjectKey{
					Name:      routeSpec.ServiceRef.Name,
					Namespace: routeSpec.ServiceRef.Namespace,
				}, svc)
				if err != nil {
					return reconcile.Result{}, err
				}
				reqLogger.Info("Generating route " + svc.Name)
				route := generateRoute(&routeSpec, svc)
				err = r.client.Create(context.TODO(), route)
				if err != nil {
					if errors.IsAlreadyExists(err) {
						reqLogger.Info("Route already exists")
					} else {
						return reconcile.Result{}, err
					}
				}
				if routeSpec.GenerateEndpoint {
					reqLogger.Info("Generating endpoint " + routeSpec.Name)
					endpoint := generateEndpoint(svc, vmi, &routeSpec)
					if err := controllerutil.SetControllerReference(instance, endpoint, r.scheme); err != nil {
						reqLogger.Error(err, "unable to set owner reference on new pod")
						return reconcile.Result{}, err
					}
					err := r.client.Create(context.TODO(), endpoint)
					if err != nil {
						if errors.IsAlreadyExists(err) {
							reqLogger.Info("Endpoint already exists")
						} else {
							return reconcile.Result{}, err
						}
					}
				}
			}
		}
	}
	return reconcile.Result{}, nil
}

func generateService(svc *appv1alpha1.ServiceSpec, vmi *vmiv1.VirtualMachineInstance) *corev1.Service {

	targetPort := intstr.IntOrString{
		IntVal: svc.TargetPort,
	}

	selector := svc.Labels

	if len(svc.Selector) > 0 {
		selector = svc.Selector
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svc.Name,
			Namespace: svc.Namespace,
			Labels:    svc.Labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       svc.Port,
					TargetPort: targetPort,
				},
			},
			Selector: selector,
		},
	}
}

func generateRoute(routeSpec *appv1alpha1.RouteSpec, svc *corev1.Service) *ocpv1.Route {
	labels := svc.Labels

	if len(routeSpec.Labels) > 0 {
		labels = routeSpec.Labels
	}
	return &ocpv1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svc.Name,
			Namespace: svc.Namespace,
			Labels:    labels,
		},
		Spec: ocpv1.RouteSpec{
			Host: routeSpec.Host,
			To: ocpv1.RouteTargetReference{
				Kind: "Service",
				Name: svc.Name,
			},
		},
	}
}

func generateEndpoint(svc *corev1.Service, vmi *vmiv1.VirtualMachineInstance, routeSpec *appv1alpha1.RouteSpec) *corev1.Endpoints {
	anno := map[string]string{
		"hostURL": routeSpec.Host,
	}
	return &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:        svc.Name + "-endpoint",
			Namespace:   svc.Namespace,
			Labels:      svc.Labels,
			Annotations: anno,
		},
		Subsets: []corev1.EndpointSubset{
			corev1.EndpointSubset{
				Addresses: []corev1.EndpointAddress{
					corev1.EndpointAddress{
						IP: vmi.Status.Interfaces[0].IP,
					},
				},
				Ports: []corev1.EndpointPort{
					corev1.EndpointPort{
						Port: svc.Spec.Ports[0].TargetPort.IntVal,
					},
				},
			},
		},
	}
}
