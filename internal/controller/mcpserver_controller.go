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
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mcpserverv1 "github.com/opendatahub-io/sample-component/api/v1"
)

// MCPServerReconciler reconciles a MCPServer object
type MCPServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=mcpserver.opendatahub.io,resources=mcpservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mcpserver.opendatahub.io,resources=mcpservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mcpserver.opendatahub.io,resources=mcpservers/finalizers,verbs=update

// +kubebuilder:rbac:groups="",resources=services,verbs=create
// +kubebuilder:rbac:groups="apps",resources=deployments,verbs=create
// +kubebuilder:rbac:groups="route.openshift.io",resources=routes,verbs=create

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *MCPServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Create a logger with the MCPServer CR's name to keep track
	logger := log.FromContext(ctx).WithName(req.Name)

	logger.Info(fmt.Sprintf("Reconciling MCPServer %s in namespace %s", req.Name, req.Namespace))

	// Capture the name and namespace of the incoming Request object
	ref := client.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Name,
	}
	mcpServer := &mcpserverv1.MCPServer{}

	// Get the MCPServer CR
	err := r.Client.Get(ctx, ref, mcpServer)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Create Deployment
	err = reconcileMCPServerDeployment(ctx, r.Client, mcpServer)
	if err != nil {
		logger.Error(err, "Failed to reconcile MCPServer Deployment")
		return ctrl.Result{}, err
	}

	// Create Service
	err = reconcileMCPServerService(ctx, r.Client, mcpServer)
	if err != nil {
		logger.Error(err, "Failed to reconcile MCPServer Service")
		return ctrl.Result{}, err
	}

	// Create Route
	err = reconcileMCPServerRoute(ctx, r.Client, mcpServer)
	if err != nil {
		logger.Error(err, "Failed to reconcile MCPServer Route")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MCPServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mcpserverv1.MCPServer{}).
		Named("mcpserver").
		Complete(r)
}
