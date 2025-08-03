import json
import os
import shutil
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


@pytest.fixture(scope="module")
def dns_server(
    docker_client: docker.client.DockerClient,
    kind_cluster: KindCluster,
) -> Generator[str, None, None]:
    temp_dir = tempfile.mkdtemp()
    dns_container = None
    try:
        corefile_path = os.path.join(temp_dir, "Corefile")
        with open(corefile_path, "w") as f:
            f.write(
                f""".:53 {{
    template IN A k8s-node-per-evaluator.io {{
        answer "{{{{ .Name }}}} 60 IN A {kind_cluster.control_plane_node_ip()}"
    }}
    forward . /etc/resolv.conf
}}
"""
            )

        dns_container = docker_client.containers.run(
            image="coredns/coredns:latest",
            detach=True,
            volumes=[f"{corefile_path}:/Corefile"],
            network="kind",
            command=["-conf", "/Corefile"],
        )
        dns_container.reload()
        dns_server_ip = dns_container.attrs["NetworkSettings"]["Networks"]["kind"][
            "IPAddress"
        ]
        yield dns_server_ip
    finally:
        if dns_container:
            dns_container.stop()
            dns_container.remove()
        shutil.rmtree(temp_dir)


def test_runner(
    docker_client: docker.client.DockerClient,
    kind_cluster: KindCluster,
    test_runner_image: docker.models.images.Image,
    test_runner_config: RunnerConfig,
    dns_server: str,
) -> None:
    with tempfile.TemporaryDirectory() as output_dir:
        os.chmod(output_dir, 0o777)

        container_output_dir = "/app/output"
        report_file = "report.json"

        mounted_config_file = "/app/config.yaml"
        container = docker_client.containers.run(
            image=test_runner_image,
            detach=True,
            volumes=[
                f"{kind_cluster.kube_config_path()}:{test_runner_config.kubeconfig_path}:ro",
                f"{test_runner_config.file_path}:{mounted_config_file}:ro",
                f"{output_dir}:{container_output_dir}:rw",
            ],
            environment={
                "TEST_RUNNER_REPORT_FORMAT": "json",
                "TEST_RUNNER_REPORT_FILE": f"{container_output_dir}/{report_file}",
            },
            network="kind",
            dns=[dns_server],
            command=["--config", f"{mounted_config_file}"],
        )
        wait_for_container(container)

        for log_line in container.attach(
            stdout=True, stderr=True, stream=True, logs=True
        ):
            print(log_line.decode())

        with open(f"{output_dir}/{report_file}", "r") as f:
            report = json.load(f)

            for test in report:
                assert test["Name"] != ""
                for result in test["TestResults"]:
                    assert (
                        result["NodeName"]
                        == kind_cluster.control_plane_node_name()
                    )
                    assert result["AverageLatency"] > 0
                    assert result["FailedRequestCount"] == 0
                    assert result["FailedPercentage"] == 0

        container.reload()
        assert container.attrs["State"]["Status"] == "exited"
        assert container.attrs["State"]["ExitCode"] == 0
        assert container.attrs["State"]["Error"] == ""

        container.stop()
        container.wait()
        container.remove()
