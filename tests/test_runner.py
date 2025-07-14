import os
import tempfile
from typing import Generator

import docker
import docker.client
import docker.models
import docker.models.images
import pytest
import yaml

from .utils.docker import wait_for_container
from .utils.kind_cluster import KindCluster


@pytest.fixture(scope="module")
def kind_cluster(
    test_service_image: docker.models.images.Image,
    docker_client: docker.client.DockerClient,
) -> Generator[KindCluster, None, None]:
    kind_cluster = KindCluster(
        cluster_name="k8s-node-perf-evaluator",
        docker_client=docker_client,
    )
    kind_cluster.start()
    kind_cluster.load_images(test_service_image)
    yield kind_cluster
    kind_cluster.stop()


class RunnerConfig:
    kubeconfig_path: str
    file_path: str

    def __init__(self, **kwargs):
        self.kubeconfig_path = kwargs["kubeconfig_path"]
        self.file_path = kwargs["file_path"]


@pytest.fixture(scope="module")
def test_runner_config() -> Generator[RunnerConfig, None, None]:
    with open("./config.yaml") as stream:
        conf = yaml.load(stream, yaml.Loader)

    conf["kubeConfig"] = "/app/kubeconfig"

    with tempfile.NamedTemporaryFile() as config_file:
        conf_yaml = yaml.dump(conf, Dumper=yaml.Dumper)
        config_file.write(conf_yaml.encode())
        config_file.flush()
        os.chmod(config_file.name, 0o666)

        yield RunnerConfig(
            file_path=config_file.name,
            kubeconfig_path=conf["kubeConfig"],
        )


def test_runner(
    docker_client: docker.client.DockerClient,
    kind_cluster: KindCluster,
    test_runner_image: docker.models.images.Image,
    test_runner_config: RunnerConfig,
) -> None:
    mounted_config_file = "/app/config.yaml"
    container = docker_client.containers.run(
        image=test_runner_image,
        detach=True,
        volumes=[
            f"{kind_cluster.kube_config_path()}:{test_runner_config.kubeconfig_path}:ro",
            f"{test_runner_config.file_path}:{mounted_config_file}:ro",
        ],
        network="host",
        command=["--config", f"{mounted_config_file}"],
    )
    wait_for_container(container)

    for log_line in container.attach(stdout=True, stderr=True, stream=True, logs=True):
        print(log_line.decode())

    container.reload()
    assert container.attrs["State"]["Status"] == "exited"
    assert container.attrs["State"]["ExitCode"] == 0
    assert container.attrs["State"]["Error"] == ""

    container.stop()
    container.wait()
    container.remove()
