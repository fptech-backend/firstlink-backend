## Firstlink-backend

## 前提条件

- Docker
- Docker Compose

## 设置
1. **构建并启动容器:**

   ```sh
    docker compose build
    docker compose up -d ## Run the server in background
   ```
## Swagger API 门户

打开 http://localhost:8080/swagger/

## 简介

该后端服务提供了一个 Swagger API 门户，通过访问该门户，您可以查看、测试和文档化所有可用的 API 端点。Swagger 提供了一个用户友好的界面，使得开发人员和用户能够轻松地交互和理解 API。

在开始之前，请确保您的系统已经安装并配置好了 Docker 和 Docker Compose。如果尚未安装，可以参考官方文档进行安装和配置。

## 设置步骤详细说明
1. 构建并启动容器：

- docker compose build：此命令将读取 Docker Compose 文件中的配置，并根据这些配置构建所需的 Docker 镜像。
- docker compose up -d：此命令将启动所有配置的服务，并在后台运行它们。添加 -d 标志是为了让命令在启动服务后立即返回控制台，而不会阻塞终端。

2. 访问 Swagger API 门户：

- 打开浏览器，输入 http://localhost:8080/swagger/。
- 您将看到一个交互式的 API 文档界面，可以在此查看每个 API 的详细信息，包括请求参数、响应格式等。
- 通过此门户，您可以直接测试 API 请求，查看实际的响应结果，从而便于调试和开发。

通过这些步骤，您将能够快速启动并运行 Firstlink-backend 服务，并利用 Swagger 提供的便利进行 API 开发和测试。