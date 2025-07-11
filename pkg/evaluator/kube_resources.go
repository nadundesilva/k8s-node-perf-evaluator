package evaluator

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const testServicePort = 8080
const testServicePortName = "http-port"

func (runner *testRunner) makeNamespace(namespace string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
}

func (runner *testRunner) makeDeployment(testService TestService) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      makeName(testService),
			Namespace: runner.config.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: makeLabels(testService),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: makeLabels(testService),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-service",
							Image: runner.config.TestService.Image,
							Ports: []corev1.ContainerPort{
								{
									Name:          testServicePortName,
									ContainerPort: testServicePort,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "SERVICE_PORT",
									Value: fmt.Sprint(testServicePort),
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceEphemeralStorage: resource.MustParse("1Gi"),
									corev1.ResourceCPU:              resource.MustParse("1"),
									corev1.ResourceMemory:           resource.MustParse("1Gi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceEphemeralStorage: resource.MustParse("1Gi"),
									corev1.ResourceCPU:              resource.MustParse("1"),
									corev1.ResourceMemory:           resource.MustParse("1Gi"),
								},
							},
						},
					},
					NodeName: testService.NodeName,
				},
			},
		},
	}
}

func (runner *testRunner) makeService(testService TestService) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      makeName(testService),
			Namespace: runner.config.Namespace,
			Labels:    makeLabels(testService),
		},
		Spec: corev1.ServiceSpec{
			Selector: makeLabels(testService),
			Ports: []corev1.ServicePort{
				{
					Name:       testServicePortName,
					Port:       testServicePort,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromString(testServicePortName),
				},
			},
		},
	}
}

func (runner *testRunner) makeIngress(testService TestService) *networkingv1.Ingress {
	host := testService.UUID + runner.config.Ingress.HostnamePostfix
	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        makeName(testService),
			Namespace:   runner.config.Namespace,
			Labels:      makeLabels(testService),
			Annotations: runner.config.Ingress.Annotations,
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: runner.config.Ingress.ClassName,
			Rules: []networkingv1.IngressRule{
				{
					Host: host,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path: runner.config.Ingress.PathPrefix,
									PathType: func() *networkingv1.PathType {
										pathType := networkingv1.PathTypePrefix
										return &pathType

									}(),
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: makeName(testService),
											Port: networkingv1.ServiceBackendPort{
												Name: testServicePortName,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			TLS: []networkingv1.IngressTLS{
				{
					Hosts:      []string{host},
					SecretName: runner.config.Ingress.TLSSecretName,
				},
			},
		},
	}
}

func makeName(testService TestService) string {
	return fmt.Sprintf("test-service-%s", testService.UUID)
}

func makeLabels(testService TestService) map[string]string {
	return map[string]string{
		"node": testService.NodeName,
		"type": "test-service",
		"app":  "k8s-node-perf-evaluator",
	}
}
