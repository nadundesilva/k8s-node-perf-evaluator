import os
import tempfile
import time
from typing import IO, Tuple

import docker
import docker.client

from .docker import wait_for_container


class KindCluster:
    _kind_cluster_name: str
    _docker_client: docker.client.DockerClient

    _kind_container: docker.models.containers.Container
    _kubeconfig_file: IO[bytes]

    def __init__(self, **kwargs):
        self._kind_cluster_name = kwargs["cluster_name"]
        self._docker_client = kwargs["docker_client"]

    def start(self) -> None:
        self._kind_container = self._docker_client.containers.run(
            image="docker:latest",
            detach=True,
            volumes={
                "/var/run/docker.sock": {"bind": "/var/run/docker.sock", "mode": "rw"}
            },
            network="host",
            command=["sleep", "infinity"],
        )
        wait_for_container(self._kind_container)

        print("Installing Kind CLI")
        self._run_command("apk update && apk add kubectl kind")
        print(f"Creating Kind Cluster: {self._kind_cluster_name}")
        self._run_command(f"""cat <<EOF | kind create cluster --name {self._kind_cluster_name} --config=-
                              kind: Cluster
                              apiVersion: kind.x-k8s.io/v1alpha4
                              nodes:
                                - role: control-plane
                                  extraPortMappings:
                                    - containerPort: 80
                                      hostPort: 80
                                      protocol: TCP
                                    - containerPort: 443
                                      hostPort: 443
                                      protocol: TCP
                            EOF""")

        kindNodeContainer = self._docker_client.containers.get(
            f"{self._kind_cluster_name}-control-plane"
        )
        wait_for_container(kindNodeContainer)

        print("Waiting for Kind Control Plane to be Ready")
        max_retries = 15
        for retry in range(max_retries):
            exit_code, _ = self._run_command_and_get_output("kubectl get nodes")
            if exit_code == 0:
                print("Kind Control Plane ready")
                break
            assert retry + 1 < max_retries, (
                "Timed out waiting for Kind Control plane to startup"
            )

            print("Kind Control Plane not yet ready")
            time.sleep(2)

        print("Installing NginX Ingress Controller")
        self._run_command(
            "kubectl apply -f https://kind.sigs.k8s.io/examples/ingress/deploy-ingress-nginx.yaml"
        )
        print("Waiting for Ingress Controller Deployment to be Ready")
        self._run_command("""kubectl rollout status deployment \
                            --namespace ingress-nginx \
                            --selector=app.kubernetes.io/component=controller \
                            --timeout=180s""")
        print("Waiting for Ingress Controller Pod to be Ready")
        self._run_command("""kubectl wait pod \
                            --namespace ingress-nginx \
                            --for=condition=ready \
                            --selector=app.kubernetes.io/component=controller \
                            --timeout=180s""")

        print("Fetching Kube Config")
        exit_code, kubeconfig = self._run_command_and_get_output(
            f"kind get kubeconfig --name {self._kind_cluster_name}"
        )
        assert exit_code == 0, kubeconfig

        self._kubeconfig_file = tempfile.NamedTemporaryFile(delete=False)
        print(f"Creating Kube Config File {self._kubeconfig_file.name}")
        self._kubeconfig_file.write(kubeconfig)
        os.chmod(self._kubeconfig_file.name, 0o666)

    def _run_command_and_get_output(self, command: str) -> Tuple[int, bytes]:
        kind_container_user = (
            self._kind_container.image.attrs["Config"]["User"]
            if self._kind_container.image is not None
            else "root"
        )
        exit_code, output = self._kind_container.exec_run(
            [
                "/bin/sh",
                "-c",
                command,
            ],
            user=kind_container_user,
        )
        return exit_code, output

    def _run_command(self, command: str) -> None:
        exit_code, output = self._run_command_and_get_output(command)
        print(output.decode())
        assert exit_code == 0

    def load_images(self, image: docker.models.images.Image) -> None:
        print(f"Loading docker image {image.id} to Kind Cluster")
        self._run_command(
            f"kind load docker-image --name {self._kind_cluster_name} {image.id}"
        )

    def kube_config_path(self) -> str:
        return self._kubeconfig_file.name

    def stop(self) -> None:
        print("Shutting Down Kind Cluster")
        self._run_command(f"kind delete cluster --name {self._kind_cluster_name}")

        self._kind_container.stop()
        self._kind_container.wait()
        self._kind_container.remove()

        self._kubeconfig_file.close()
        os.unlink(self._kubeconfig_file.name)
