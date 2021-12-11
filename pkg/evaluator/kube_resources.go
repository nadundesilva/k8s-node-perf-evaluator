package evaluator

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const TEST_SERVICE_PORT = 8080
const TEST_SERVICE_PORT_NAME = "http-port"

var INGRESS_PATH_TYPE = networkingv1.PathTypePrefix

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
									Name:          TEST_SERVICE_PORT_NAME,
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

func (runner *testRunner) makeService(nodeName string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      makeName(nodeName),
			Namespace: runner.config.Namespace,
			Labels:    makeLabels(nodeName),
		},
		Spec: corev1.ServiceSpec{
			Selector: makeLabels(nodeName),
			Ports: []corev1.ServicePort{
				{
					Name:       TEST_SERVICE_PORT_NAME,
					Port:       TEST_SERVICE_PORT,
					TargetPort: intstr.FromString(TEST_SERVICE_PORT_NAME),
				},
			},
		},
	}
}

func (runner *testRunner) makeIngress(nodeName string) *networkingv1.Ingress {
	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      makeName(nodeName),
			Namespace: runner.config.Namespace,
			Labels:    makeLabels(nodeName),
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &runner.config.Ingress.ClassName,
			Rules: []networkingv1.IngressRule{
				{
					Host: runner.config.Ingress.Hostname,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     runner.config.Ingress.PathPrefix,
									PathType: &INGRESS_PATH_TYPE,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: makeName(nodeName),
											Port: networkingv1.ServiceBackendPort{
												Name: TEST_SERVICE_PORT_NAME,
											},
										},
									},
								},
							},
						},
					},
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
