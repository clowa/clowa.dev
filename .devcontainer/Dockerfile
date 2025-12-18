ARG NODE_VERSION=18

# FROM swacli/static-web-apps-cli:1.1.7
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

# Install Azure Static web app CLI
RUN npm install -g @azure/static-web-apps-cli

# Install azure function core tools
# This requires version 4.3.0 or higher - see: https://github.com/Azure/azure-functions-core-tools/issues/4279
RUN npm install -g azure-functions-core-tools@4

# Create a non-root user to use - see: https://code.visualstudio.com/remote/advancedcontainers/add-nonroot-user#_creating-a-nonroot-user
RUN groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USER_GID -m --shell /bin/bash ${USERNAME} \
    && echo "${USERNAME} ALL=(root) NOPASSWD:ALL" > /etc/sudoers.d/${USERNAME} \
    && chmod 0440 /etc/sudoers.d/${USERNAME}
USER ${USERNAME}
