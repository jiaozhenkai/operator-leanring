/*
Copyright 2022.

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

	"k8s.io/apimachinery/pkg/types"

	"github.com/jiaozhenkai/webserver-operator/k8sdao"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mydomainv1 "github.com/jiaozhenkai/webserver-operator/api/v1"
)

// WebServerReconciler reconciles a WebServer object
type WebServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=my.domain,resources=webservers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=my.domain,resources=webservers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=my.domain,resources=webservers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the WebServer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile

const WebserverFinalizer = "webserver.finalizer"

func (r *WebServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	fmt.Println("start reconcile......")
	_ = log.FromContext(ctx)
	// TODO(user): your logic here
	var webserverList = mydomainv1.WebServerList{}

	err := r.List(ctx, &webserverList)

	if err != nil {
		log.Log.Error(err, "list webserver list error")
		return ctrl.Result{}, err
	}

	for _, item := range webserverList.Items {

		// 如果 DeletionTimestamp 是 0，则资源未被删除；检查 finalizer, 如果没有 finalizer 则添加
		if item.DeletionTimestamp.IsZero() {
			if len(item.Finalizers) == 0 {
				item.Finalizers = append(item.Finalizers, WebserverFinalizer)
			}

			hasFinalizer := containsFinalizer(item.Finalizers)
			if !hasFinalizer {
				item.Finalizers = append(item.Finalizers, WebserverFinalizer)
			}

			fmt.Println("first exec update")
			if err := r.Update(ctx, &item); err != nil {
				log.Log.Error(err, "update finalizer error")
				return ctrl.Result{}, err
			}
		} else {
			// DeletionTimestamp 不为 0 ，则对象正在被删除
			// 这里需要判断 CR 中的 finalizer 是不是与自定义的 finalizer 一致，如果一致则应该执行一个方法去删除依赖资源，然后再来删除 finalizer
			fmt.Println("delete webserver cr")
			item.Finalizers = removeFinalizer(item.Finalizers)
			fmt.Println("second exec update")
			if err := r.Update(ctx, &item); err != nil {
				log.Log.Error(err, "remove finalizer error")
				return ctrl.Result{}, err
			}
		}

		if item.Spec.Name != "replicas-update" {
			return ctrl.Result{}, fmt.Errorf("cr name must replicas-update will update replicas")
		}
		updateNamespacedName := types.NamespacedName{Namespace: req.Namespace, Name: "nginx-deployment"}
		err := k8sdao.UpdateReplicas(ctx, int32(item.Spec.Replicas), updateNamespacedName, r.Client, req.Namespace)
		if err != nil {
			item.Status.Reason = err.Error()
			item.Status.Message = fmt.Sprintf("%s %d %s", "update replicas to ", item.Spec.Replicas, " failed")
			r.Status().Update(ctx, &item)
			return ctrl.Result{}, err
		}

		item.Status.Reason = ""
		item.Status.Message = fmt.Sprintf("%s %d %s", "update replicas to ", item.Spec.Replicas, "success")

		item.SetFinalizers([]string{})
		fmt.Println("first exec update status")
		r.Status().Update(ctx, &item)
		fmt.Println("third exec update")
		r.Update(ctx, &item)
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WebServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mydomainv1.WebServer{}).
		Complete(r)
}

func removeFinalizer(items []string) []string {
	res := []string{}
	for _, item := range items {
		if item == WebserverFinalizer {
			continue
		}
		res = append(res, item)
	}
	return res
}

func containsFinalizer(items []string) bool {
	if len(items) == 0 {
		return false
	}

	var i int

	for i = 0; i < len(items); i++ {
		if items[i] == WebserverFinalizer {
			return true
		}
	}

	return false
}
