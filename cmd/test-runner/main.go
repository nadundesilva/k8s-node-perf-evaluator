package main

import (
	"flag"
	"path/filepath"

	"github.com/nadundesilva/k8s-node-perf-evaluator/pkg/k8s"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	k8sClient := k8s.NewFromKubeConfig(*kubeconfig)
}
