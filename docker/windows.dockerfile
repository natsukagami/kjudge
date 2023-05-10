FROM mcr.microsoft.com/powershell:lts-nanoserver-ltsc2022 AS base
WORKDIR kjudge

USER ContainerAdministrator
SHELL ["pwsh", "--command"]

# Stage 1: Install front-end
FROM base AS frontend

COPY scripts/windows scripts/windows

RUN ./scripts/windows/install_nodejs.ps1

RUN npm install yarn -g