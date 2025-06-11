package controller

import (
	"context"

	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mcpserverv1 "github.com/opendatahub-io/sample-component/api/v1"
)

const (
	mcpServerAppLabelKey = "ocp-mcp-server"
)

func reconcileMCPServerDeployment(ctx context.Context, cli client.Client, mcpServer *mcpserverv1.MCPServer) error {
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      mcpServer.Name,
			Namespace: mcpServer.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: mcpServer.APIVersion,
					Kind:       mcpServer.Kind,
					Name:       mcpServer.Name,
					UID:        mcpServer.UID,
				},
			},
			Labels: map[string]string{
				mcpServerAppLabelKey: mcpServer.Name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					mcpServerAppLabelKey: mcpServer.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						mcpServerAppLabelKey: mcpServer.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "mcp-server",
							Image: mcpServer.Spec.Image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8000,
								},
							},
							Command: []string{"./kubernetes-mcp-server"},
							Args:    []string{"--sse-port", "8000"},
						},
					},
				},
			},
		},
	}

	err := cli.Create(ctx, deployment)
	if err != nil && !k8serr.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func reconcileMCPServerService(ctx context.Context, cli client.Client, mcpServer *mcpserverv1.MCPServer) error {
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      mcpServer.Name,
			Namespace: mcpServer.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: mcpServer.APIVersion,
					Kind:       mcpServer.Kind,
					Name:       mcpServer.Name,
					UID:        mcpServer.UID,
				},
			},
			Labels: map[string]string{
				mcpServerAppLabelKey: mcpServer.Name,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				mcpServerAppLabelKey: mcpServer.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   "TCP",
					Port:       8000,
					TargetPort: intstr.FromString("http"),
				},
			},
		},
	}

	err := cli.Create(ctx, service)
	if err != nil && !k8serr.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func reconcileMCPServerRoute(ctx context.Context, cli client.Client, mcpServer *mcpserverv1.MCPServer) error {
	route := &routev1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: "route.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      mcpServer.Name,
			Namespace: mcpServer.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: mcpServer.APIVersion,
					Kind:       mcpServer.Kind,
					Name:       mcpServer.Name,
					UID:        mcpServer.UID,
				},
			},
			Labels: map[string]string{
				mcpServerAppLabelKey: mcpServer.Name,
			},
		},
		Spec: routev1.RouteSpec{
			Path: "/sse",
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: mcpServer.Name,
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromInt32(8000),
			},
		},
	}

	err := cli.Create(ctx, route)
	if err != nil && !k8serr.IsAlreadyExists(err) {
		return err
	}

	return nil
}
