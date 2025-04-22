package controller

import (
	"context"
	"fmt"

	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	helloworldv1 "github.com/opendatahub-io/sample-component/api/v1"
)

const (
	helloWorldAppLabelKey = "app"
	helloWorldAppLabelVal = "hello-world"
)

func reconcileHelloWorldConfigMap(ctx context.Context, cli client.Client, hw *helloworldv1.HelloWorld) error {
	html := fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
      <head><title>Hello World</title></head>
      <body>
        <h1>%s</h1>
      </body>
    </html>`, hw.Spec.Message)

	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-html", hw.Name),
			Namespace: hw.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: hw.APIVersion,
					Kind:       hw.Kind,
					Name:       hw.Name,
					UID:        hw.UID,
				},
			},
		},
		Data: map[string]string{
			"index.html": html,
		},
	}

	err := cli.Create(ctx, cm)
	if err != nil && !k8serr.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func reconcileHelloWorldDeployment(ctx context.Context, cli client.Client, hw *helloworldv1.HelloWorld) error {
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-nginx", hw.Name),
			Namespace: hw.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: hw.APIVersion,
					Kind:       hw.Kind,
					Name:       hw.Name,
					UID:        hw.UID,
				},
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					helloWorldAppLabelKey: helloWorldAppLabelVal,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						helloWorldAppLabelKey: helloWorldAppLabelVal,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginxinc/nginx-unprivileged:latest",
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8080,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "html",
									MountPath: "/usr/share/nginx/html/index.html",
									SubPath:   "index.html",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "html",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: fmt.Sprintf("%s-html", hw.Name),
									},
								},
							},
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

func reconcileHelloWorldService(ctx context.Context, cli client.Client, hw *helloworldv1.HelloWorld) error {
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-nginx", hw.Name),
			Namespace: hw.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: hw.APIVersion,
					Kind:       hw.Kind,
					Name:       hw.Name,
					UID:        hw.UID,
				},
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				helloWorldAppLabelKey: helloWorldAppLabelVal,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   "TCP",
					Port:       8080,
					TargetPort: intstr.FromInt32(8080),
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

func reconcileHelloWorldRoute(ctx context.Context, cli client.Client, hw *helloworldv1.HelloWorld) error {
	route := &routev1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: "route.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-nginx", hw.Name),
			Namespace: hw.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: hw.APIVersion,
					Kind:       hw.Kind,
					Name:       hw.Name,
					UID:        hw.UID,
				},
			},
		},
		Spec: routev1.RouteSpec{
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromInt32(8080),
			},
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: fmt.Sprintf("%s-nginx", hw.Name),
			},
			TLS: &routev1.TLSConfig{
				Termination: routev1.TLSTerminationEdge,
			},
		},
	}

	err := cli.Create(ctx, route)
	if err != nil && !k8serr.IsAlreadyExists(err) {
		return err
	}

	return nil
}
