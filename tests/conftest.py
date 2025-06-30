from typing import Generator

import docker
import pytest

from .utils import build_image


@pytest.fixture(scope="module")
def docker_client() -> Generator[docker.DockerClient, None, None]:
    client = docker.from_env()
    yield client
    client.close()
