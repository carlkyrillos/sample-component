/*
Copyright 2025.

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

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	helloworldv1 "github.com/opendatahub-io/sample-component/api/v1"
)

// HelloWorldReconciler reconciles a HelloWorld object
type HelloWorldReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=helloworld.opendatahub.io,resources=helloworlds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=helloworld.opendatahub.io,resources=helloworlds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=helloworld.opendatahub.io,resources=helloworlds/finalizers,verbs=update

func (r *HelloWorldReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Create a logger with the HelloWorld CR's name to keep track
	logger := log.FromContext(ctx).WithName(req.Name)

	// Capture the name and namespace of the incoming Request object
	ref := client.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Name,
	}
	hw := &helloworldv1.HelloWorld{}

	// Get the HelloWorld CR
	err := r.Client.Get(ctx, ref, hw)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Extract the message from the HelloWorld CR
	message := hw.Spec.Message

	// Print the message to the logs
	logger.Info(message)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HelloWorldReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&helloworldv1.HelloWorld{}).
		Named("helloworld").
		Complete(r)
}
