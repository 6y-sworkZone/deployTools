# DeployTools - Golang CLI部署工具

一个功能完善的Golang命令行部署工具，支持服务器管理、项目配置、项目组合管理和多种部署模式。

## 功能特性

### 项目组合管理
- 创建、保存和加载多个项目组合配置
- 交互式命令行界面选择现有组合或创建新组合
- 查看组合内包含的所有项目列表及基本信息

### 部署模式
- **组合部署模式**：一键部署组合内所有项目
- **单独部署模式**：从组合中选择特定项目进行部署
- 部署前确认机制，显示待部署项目清单供用户确认

### 项目配置管理
- 本地路径配置：支持Windows和Linux系统路径格式
- 远程服务器选择：从已配置服务器列表中选择目标部署服务器
- 远程路径配置：指定项目在远程服务器上的部署路径
- 更新方式选择：
  - **diff**：差量比对更新（仅传输变更文件）
  - **replace_file**：直接替换文件（覆盖单个文件）
  - **replace_dir**：直接替换文件夹（完全覆盖目标目录）
- 后处理脚本配置：部署完成后自动执行脚本命令
- 文件排除模式

### 远程服务器管理
- 添加、编辑、删除服务器信息
- 服务器信息包括：名称、IP地址、端口号、用户名、认证凭据（密码或SSH密钥）
- 服务器连接测试功能，验证配置的有效性

### 质量与性能
- 跨平台兼容性（Windows和Linux系统）
- 部署前备份选项，提高安全性
- 差量更新模式优化传输效率
- 详细的部署进度反馈
- 完整的日志记录功能

## 项目结构

```
deploytools/
├── cmd/                    # CLI命令入口
│   ├── main.go            # 主程序入口
│   ├── server.go          # 服务器管理命令
│   ├── project.go         # 项目管理命令
│   ├── group.go           # 项目组管理命令
│   ├── deploy.go          # 部署命令
│   └── list.go            # 列表命令
├── internal/              # 内部包
│   ├── config/            # 配置管理
│   │   ├── models.go      # 核心数据结构
│   │   └── manager.go     # 配置管理器
│   └── deploy/            # 部署逻辑
│       └── deployer.go    # 部署核心逻辑
├── pkg/                   # 公共包
│   ├── ssh/               # SSH/SFTP客户端
│   │   └── client.go      # SSH连接和文件传输
│   └── utils/             # 工具函数
│       ├── utils.go       # 通用工具函数
│       └── logger.go      # 日志记录器
├── configs/               # 配置文件目录
├── go.mod                 # Go模块依赖
├── go.sum                 # 依赖校验和
├── deploytools.exe        # 编译后的可执行文件
└── README.md              # 项目说明文档
```

## 核心数据结构

### Server（服务器）
- ID：唯一标识符
- Name：服务器名称
- IP：IP地址
- Port：SSH端口（默认22）
- Username：用户名
- AuthType：认证类型（password/key）
- Password：密码
- KeyPath：SSH私钥路径

### Project（项目）
- ID：唯一标识符
- Name：项目名称
- LocalPath：本地路径
- ServerID：关联服务器ID
- RemotePath：远程部署路径
- UpdateMode：更新模式（diff/replace_file/replace_dir）
- PostScript：后处理脚本
- ExcludePatterns：排除文件模式

### ProjectGroup（项目组）
- ID：唯一标识符
- Name：组名称
- Description：描述
- ProjectIDs：项目ID列表

## 快速开始

### 1. 编译项目

```bash
go build -o deploytools.exe ./cmd
```

### 2. 查看帮助

```bash
deploytools --help
deploytools server --help
deploytools project --help
deploytools group --help
deploytools deploy --help
deploytools list --help
```

### 3. 常用命令示例

#### 服务器管理
```bash
# 添加服务器（交互式）
deploytools server add

# 测试服务器连接
deploytools server test <server-id>

# 编辑服务器配置
deploytools server edit <server-id>

# 删除服务器
deploytools server delete <server-id>
```

#### 项目管理
```bash
# 添加项目（交互式）
deploytools project add

# 编辑项目配置
deploytools project edit <project-id>

# 删除项目
deploytools project delete <project-id>
```

#### 项目组管理
```bash
# 创建项目组
deploytools group add

# 查看项目组详情
deploytools group show <group-id>

# 编辑项目组
deploytools group edit <group-id>
```

#### 部署操作
```bash
# 部署单个项目
deploytools deploy project <project-id>

# 部署项目组
deploytools deploy group <group-id>

# 交互式部署
deploytools deploy interactive

# 跳过确认直接部署
deploytools deploy project <project-id> -y

# 禁用备份
deploytools deploy project <project-id> --no-backup
```

#### 查看资源
```bash
# 列出所有资源
deploytools list all

# 只列出服务器
deploytools list servers

# 只列出项目
deploytools list projects

# 只列出项目组
deploytools list groups
```

### 4. 配置文件位置

配置文件默认存储在用户主目录下：
- Windows: `C:\Users\<username>\.deploytools\config.yaml`
- Linux: `~/.deploytools/config.yaml`

也可以使用 `--config` 选项指定自定义配置文件路径。

## 部署模式说明

### diff（差量更新）
- 对本地和远程文件进行MD5比对
- 只传输有变更的文件
- 传输效率高，推荐日常使用

### replace_file（替换文件）
- 删除远程已存在的同名文件
- 上传所有本地文件
- 适合确保文件完全一致的场景

### replace_dir（替换目录）
- 先删除整个远程目标目录
- 重新上传所有文件
- 最彻底的更新方式，但耗时较长

## 日志

部署日志自动保存到配置目录下的 `logs/` 文件夹中，文件命名格式为：`deploy_YYYYMMDD_HHMMSS.log`

## 技术栈

- **Go 1.x**：编程语言
- **Cobra**：CLI框架
- **Viper**：配置管理
- **sftp**：文件传输
- **crypto/ssh**：SSH连接
- **uuid**：唯一ID生成

## 可扩展性

本项目设计具有良好的可扩展性：
- 插件机制：支持未来扩展更多部署方式
- API接口：预留支持未来开发Web界面
- 部署历史：可扩展版本回滚功能

## 许可证

MIT License
