from typing import Generator

import docker
import docker.models
import docker.models.images
import pytest
import requests

from .utils import build_image, wait_for_container

ping_response_body = {"status": "success"}
cpu_intensive_task_response_body = {"status": "success", "result": "-253290.33"}


@pytest.fixture(scope="module")
def test_service_image(
    docker_client: docker.DockerClient,
) -> Generator[docker.models.images.Image, None, None]:
    image = build_image(
        context_path="../",
        image_tag="test-service:test",
        dockerfile="docker/Dockerfile.test-service",
        docker_client=docker_client,
    )
    yield image
    docker_client.images.remove(image.id, force=True)


def test_service(
    docker_client: docker.DockerClient,
    test_service_image: docker.models.images.Image,
) -> None:
    server_bind_port = 18080

    container = docker_client.containers.run(
        image=test_service_image,
        detach=True,
        ports={str(8080): server_bind_port},
    )
    wait_for_container(container)

    resp = requests.get(f"http://localhost:{server_bind_port}/ping")
    assert resp.status_code == 200
    assert resp.json() == ping_response_body

    resp = requests.get(f"http://localhost:{server_bind_port}/cpu-intensive-task")
    assert resp.status_code == 200
    assert resp.json() == cpu_intensive_task_response_body

    container.stop()
    container.wait()
    container.remove()
