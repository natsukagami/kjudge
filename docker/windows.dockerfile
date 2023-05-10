FROM mcr.microsoft.com/powershell:lts-nanoserver-ltsc2022 AS base
WORKDIR kjudge

USER ContainerAdministrator
SHELL ["pwsh", "--command"]

# Stage 1: Install front-end
FROM base AS frontend

COPY scripts/windows/add_path.ps1 scripts/windows/add_path.ps1

COPY scripts/windows/install_nodejs.ps1 scripts/windows/install_nodejs.ps1
RUN ./scripts/windows/install_nodejs.ps1
# RUN npm install yarn -g

COPY scripts/windows/install_python.ps1 scripts/windows/install_python.ps1
RUN ./scripts/windows/install_python.ps1
RUN python --version

# WORKDIR frontend
# COPY ./frontend/package.json ./frontend/yarn.lock ./
# RUN yarn install --frozen-lockfile

# COPY ./ /kjudge
# RUN yarn --prod build
