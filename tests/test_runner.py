from typing import Generator

import docker
import docker.models
import docker.models.images
import pytest

from .utils import build_image, wait_for_container


@pytest.fixture(scope="module")
def test_runner_image(
    docker_client: docker.DockerClient,
) -> Generator[docker.models.images.Image, None, None]:
    image = build_image(
        context_path="../",
        image_tag="test-runner:test",
        dockerfile="docker/Dockerfile.test-runner",
        docker_client=docker_client,
    )
    yield image
    docker_client.images.remove(image.id, force=True)


def test_runner(
    docker_client: docker.DockerClient,
    test_runner_image: docker.models.images.Image,
) -> None:
    container = docker_client.containers.run(image=test_runner_image, detach=True)
    wait_for_container(container)

    container.stop()
    container.wait()
    container.remove()
