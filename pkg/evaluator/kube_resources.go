package evaluator

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const TEST_SERVICE_PORT = 8080

func (runner *testRunner) makeNamespace(namespace string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
}

func (runner *testRunner) makeDeployment(nodeName string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      makeName(nodeName),
			Namespace: runner.config.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: makeLabels(nodeName),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: makeLabels(nodeName),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-service",
							Image: runner.config.TestService.Image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http-port",
									ContainerPort: TEST_SERVICE_PORT,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "SERVICE_PORT",
									Value: fmt.Sprint(TEST_SERVICE_PORT),
								},
							},
						},
					},
					NodeName: nodeName,
				},
			},
		},
	}
}

func makeName(nodeName string) string {
	return fmt.Sprintf("test-service-%s", nodeName)
}

func makeLabels(nodeName string) map[string]string {
	return map[string]string{
		"node": nodeName,
		"type": "test-service",
		"app":  "k8s-node-perf-evaluator",
	}
}
