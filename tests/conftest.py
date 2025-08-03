from typing import Generator

import docker
import docker.client
import pytest

from .utils.docker import build_image


@pytest.fixture(scope="module")
def docker_client() -> Generator[docker.client.DockerClient, None, None]:
    client = docker.client.from_env()
    yield client
    client.close()


@pytest.fixture(scope="module")
def test_service_image(
    docker_client: docker.client.DockerClient,
) -> Generator[docker.models.images.Image, None, None]:
    image = build_image(
        context_path="../",
        image_tag="test-service:test",
        dockerfile="docker/Dockerfile.test-service",
        docker_client=docker_client,
    )
    yield image
    docker_client.images.remove(image.id, force=True)


@pytest.fixture(scope="module")
def test_runner_image(
    docker_client: docker.client.DockerClient,
) -> Generator[docker.models.images.Image, None, None]:
    image = build_image(
        context_path="../",
        image_tag="test-runner:test",
        dockerfile="docker/Dockerfile.test-runner",
        docker_client=docker_client,
    )
    yield image
    docker_client.images.remove(image.id, force=True)
