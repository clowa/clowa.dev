ARG NODE_VERSION=18

FROM node:${NODE_VERSION}-bookworm as final

ARG TARGETARCH

ARG USERNAME=codespace
ARG USER_UID=1001
ARG USER_GID=$USER_UID

# APT caching - see: https://vsupalov.com/buildkit-cache-mount-dockerfile/
RUN --mount=target=/var/lib/apt/lists,type=cache,sharing=locked \
    --mount=target=/var/cache/apt,type=cache,sharing=locked \
    apt-get update \
    && apt-get upgrade -y \
    && apt-get install -y --no-install-recommends \
        sudo \
        lsb-release

# Create a non-root user to use - see: https://code.visualstudio.com/remote/advancedcontainers/add-nonroot-user#_creating-a-nonroot-user
RUN groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USER_GID -m --shell /bin/bash ${USERNAME} \
    && echo "${USERNAME} ALL=(root) NOPASSWD:ALL" > /etc/sudoers.d/${USERNAME} \
    && chmod 0440 /etc/sudoers.d/${USERNAME}
USER ${USERNAME}
