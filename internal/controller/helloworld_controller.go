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
	"time"

	helloworldv1 "github.com/opendatahub-io/sample-component/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	defaultRequeueInterval   = 10 * time.Second
	configurationIntervalKey = "interval"
)

// HelloWorld Labels
const (
	SampleComponentPrefix = "sample-component.opendatahub.io"
	SampleComponentPartOf = SampleComponentPrefix + "/part-of"
	True                  = "true"
)

// HelloWorldReconciler reconciles a HelloWorld object
type HelloWorldReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=helloworld.opendatahub.io,resources=helloworlds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=helloworld.opendatahub.io,resources=helloworlds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=helloworld.opendatahub.io,resources=helloworlds/finalizers,verbs=update

// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch

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

	// Get the configuration ConfigMap
	cm, err := r.getConfigurationConfigMap(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Parse the refreshInterval from the ConfigMap
	interval, err := parseIntervalFromConfigMap(cm)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: interval}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HelloWorldReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&helloworldv1.HelloWorld{}).
		Named("helloworld").
		Complete(r)
}

// getConfigurationConfigMap Gets the configuration ConfigMap by label
func (r *HelloWorldReconciler) getConfigurationConfigMap(ctx context.Context) (*corev1.ConfigMap, error) {
	cmList := &corev1.ConfigMapList{}

	opts := []client.ListOption{
		client.MatchingLabels(map[string]string{
			SampleComponentPartOf: True,
		}),
	}
	err := r.Client.List(ctx, cmList, opts...)
	if err != nil {
		return nil, err
	}

	switch len(cmList.Items) {
	case 1:
		// Exactly 1 ConfigMap was found so return it
		return &cmList.Items[0], nil
	case 0:
		// No ConfigMaps were found so return nil but don't error
		return nil, nil
	default:
		return nil, fmt.Errorf("expected 1 configuration ConfigMap, got %d", len(cmList.Items))
	}
}

// parseIntervalFromConfigMap returns the parsed time interval if it exists and is valid
// A default value is returned if no interval exists
func parseIntervalFromConfigMap(cm *corev1.ConfigMap) (time.Duration, error) {
	if stringInterval, exists := cm.Data[configurationIntervalKey]; exists {
		duration, err := time.ParseDuration(stringInterval)
		if err != nil {
			return time.Duration(0), fmt.Errorf("unable to parse interval from configmap: %w", err)
		}
		return duration, nil
	}

	return defaultRequeueInterval, nil
}
