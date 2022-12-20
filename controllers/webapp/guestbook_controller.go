/*
Copyright 2022 Baiju Muthukadan.

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

package webapp

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	webappv1beta1 "github.com/baijum/guestbook/apis/webapp/v1beta1"
)

// GuestbookReconciler reconciles a Guestbook object
type GuestbookReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	ctrl   controller.Controller
}

var watchMap map[string]string = map[string]string{}

func validateLabels(fromSB, fromResource map[string]string) bool {
	fmt.Println("From SB:", fromSB)
	fmt.Println("From resource:", fromResource)
	return true
}

//+kubebuilder:rbac:groups=webapp.muthukadan.net,resources=guestbooks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=webapp.muthukadan.net,resources=guestbooks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=webapp.muthukadan.net,resources=guestbooks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Guestbook object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.1/pkg/reconcile
func (r *GuestbookReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	mapWorkloadToSB := func(a client.Object) []reconcile.Request {
		gbList := &webappv1beta1.GuestbookList{}
		opts := &client.ListOptions{}
		if err := r.List(context.Background(), gbList, opts); err != nil {
			return []reconcile.Request{}
		}
		reply := make([]reconcile.Request, 0, len(gbList.Items))
		for _, sb := range gbList.Items {
			if sb.Spec.Workload.Kind == a.GetObjectKind().GroupVersionKind().Kind &&
				validateLabels(sb.Spec.Workload.Selector.MatchLabels, a.GetLabels()) {
				reply = append(reply, reconcile.Request{NamespacedName: types.NamespacedName{
					Namespace: sb.Namespace,
					Name:      sb.Name,
				}})
			}
		}
		return reply
	}
	resource := &unstructured.Unstructured{}
	gvk := schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}
	resource.SetGroupVersionKind(gvk)
	if _, ok := watchMap[gvk.String()]; !ok {
		watchMap[gvk.String()] = ""
		r.ctrl.Watch(
			&source.Kind{Type: resource},
			handler.EnqueueRequestsFromMapFunc(mapWorkloadToSB))
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GuestbookReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctrl, err := ctrl.NewControllerManagedBy(mgr).
		For(&webappv1beta1.Guestbook{}).
		Build(r)
	r.ctrl = ctrl
	return err
}
