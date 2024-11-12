# **travel-from-sysu-backend**

中级实训中煮出去玩后端 好耶！出去玩！

基于 `Gin` 框架的 `Go` 后端项目。

### 0、进度

2024.11.12: register接口v1完成

### 一、安装和运行

1. **克隆项目**

```Bash
git clone https://github.com/your-username/your-repo-name.git
cd your-repo-name
```

1. **安装依赖**

使用以下命令安装项目所需的依赖项：

```Bash
go mod tidy
```

1. **配置环境**

   1. 先在mysql创建个叫`TravelFromSysu`的数据库

   2. 修改config.yaml改成你机子的配置：

   3. ```YAML
      app:
        name: TravelFromSysu
        port: :3000
      
      database:
        dsn : ur_username:ur_password@tcp(127.0.0.1:3306)/TravelFromSysu?charset=utf8mb4&parseTime=True&loc=Local
        MaxIdleConns : 11
        MaxOpenConns : 114
      ```

2. **运行项目**

   1. 执行以下命令启动服务器，用goland的话好多gui按钮可以运行：

   2. ```Bash
      go run main.go
      ```

   3. 启动后，服务器默认在 `http://localhost:3000` 运行，复制链接到apifox里mock测试api。

### 二、后端方便生成和修改api文档——Swagger 安装与使用

go可以通过写一些规定格式的注释通过swag工具生成可以导入apifox的api文档数据（swagger.json & swagger.yaml），本项目为了方便生成和修改api文档选择使用swagger。

更多基础概念参考博客：

[Go语言使用swagger生成接口文档 - Q1mi - 博客园](https://www.cnblogs.com/liwenzhou/p/13629767.html)

#### 1. 安装 Swagger

安装 `Swagger` 所需的工具 `swag`：

```Bash
go install github.com/swaggo/swag/cmd/swag@latest
```

> 注意：`swag` 将安装在 `$GOPATH/bin` 下，请确保 `$GOPATH/bin` 已经在你的系统路径中。

> 对于mac可参考下图加戏到系统路径里：

![img](https://svda6q665m8.feishu.cn/space/api/box/stream/download/asynccode/?code=YmFlOGYzYzRkNDk2NmFlNDliYzE3ZmI0ZjNlZmY4YTVfNUpzejUxS1hiWmFSdkhXdWNSTXZXQUM5VGRmZTE2dzBfVG9rZW46WWtVRmJ0VmJNb3ZMcUh4M2JRSGNhZXRrbnRlXzE3MzEzODg0MDk6MTczMTM5MjAwOV9WNA)

#### 2. 生成 Swagger 文档

在项目文件夹终端使用以下命令：

```Bash
swag init
```

该命令会在项目根目录下生成 `docs` 文件夹，包含 `swagger.json` 和 `swagger.yaml` 文件，以及用于 API 文档的静态文件。

#### 3. 将swagger文档导入apifox

[导入 OpenAPI (Swagger) 数据 | Apifox 帮助文档](https://apifox.com/help/api-docs/importing-api/swagger)

#### 4. 其他说明

- 每次更新或新增 API 注释后，重新运行 `swag init` 命令以更新文档。
- gorm.model不能被swagger识别所以抛弃了自己加了其提供的id, create time,...几个字段，将request请求的model和user的model分开了，后续哪个分到哪个文件夹可以下次讨论规定下。
- **报错问gpt(**
