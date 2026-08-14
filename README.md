# 项目介绍

本仓库是 [develop202/migu_video](https://github.com/develop202/migu_video) 的 Go 语言重构版本。

使用 Go 1.26 从零重写，零外部依赖，保留原项目全部环境变量配置和 API 兼容性，可直接替换原 Node.js 版本使用。

相比原版的主要改进：

- 并发处理：新增 singleflight 请求合并 + 令牌桶限速器，替代原有自旋锁串行机制
- 部署体积：静态编译单二进制，Docker 镜像基于 scratch，约 7MB（原版约 160MB）
- 多架构支持：CI 自动构建 amd64/arm64/arm 多架构镜像，发布到 ghcr.io 和 Docker Hub

## 技术栈

- 语言：Go 1.26，零外部依赖，仅使用标准库
- 运行时：静态编译单二进制，无需安装运行时环境
- 镜像：多阶段构建，基于 `scratch`，约 7MB
- CI/CD：GitHub Actions 自动构建并发布多架构 Docker 镜像到 ghcr.io 和 Docker Hub

# 本地部署

> [!warning]
> ⚠️ 注意事项
>
> 1. 登录会封号！登录会封号！登录会封号！为避免不必要的损失，请谨慎登录使用
> 2. 需要国内 IP 才可正常访问（非港澳台地区）

## 配置

默认本机和局域网可用，提供自定义 token，格式: `http://ip:port/mpass/userid/token`（未设置 mpass 请删除），使用此方式建议把画质改到蓝光或更高。

配置信息如下：

| 环境变量名 | 默认值 | 类型 | 介绍 |
| --- | --- | --- | --- |
| muserId | | string | 用户 id，可在网页端登录获取 |
| mtoken | | string | 用户 token，可在网页端登录获取 |
| mport | 1234 | number | 本地运行端口号 |
| mhost | | string | 公网/自定义访问地址，格式 `http://ip:port` |
| mrateType | 3 | number | 画质：2=标清，3=高清，4=蓝光，7=原画，9=4k。蓝光及以上需要登录且有 VIP |
| mpass | | string | 访问密码（大小写字母和数字），添加后访问格式 `http://ip:port/mpass/...` |
| menableHDR | true | boolean | 是否开启 HDR |
| menableH265 | true | boolean | 是否开启 h265（原画画质），开启后可能存在兼容性问题，比如浏览器播放没有画面 |
| mupdateInterval | 6 | number | 节目信息更新间隔，单位小时，不建议设置太短 |
| mignoreCategory | null | string | 屏蔽分类，每个分类名用逗号隔开。例如: `央视,卫视,体育-昨天`。TV 可屏蔽所有电视，PE 可屏蔽所有体育 |
| mmergeTVCategory | true | boolean | 是否将 TV 节目数量较少的分类合并到其他分类 |
| mcustomMergeCategory | null | string | 自定义合并分类到其他分类，每个分类名用逗号隔开，需先将 mmergeTVCategory 设置 false。格式: `熊猫,综艺,新闻` |
| mrefreshToken | false | boolean | 是否自动刷新 token（可能是导致封号的原因） |

## Docker 部署

### 拉取镜像

```shell
# GitHub Container Registry
docker pull ghcr.io/jipzeongit/migu-video-go:latest

# Docker Hub
docker pull jipzeongit/migu-video-go:latest
```

### 运行

```shell
# GitHub Container Registry
docker run -p 1234:1234 --name migu_video ghcr.io/jipzeongit/migu-video-go:latest

# Docker Hub
docker run -p 1234:1234 --name migu_video jipzeongit/migu-video-go:latest
```

若需要修改配置，可以使用以下命令：

```shell
# GitHub Container Registry
docker run -p 3000:3000 -e mport=3000 -e mhost="http://localhost:3000" --name migu_video ghcr.io/jipzeongit/migu-video-go:latest

# Docker Hub
docker run -p 3000:3000 -e mport=3000 -e mhost="http://localhost:3000" --name migu_video jipzeongit/migu-video-go:latest
```

### 自行构建镜像

```shell
docker build -t migu_video .
```

## 编译运行

### 环境要求

需要 Go 1.26+ 环境

### 获取代码

```shell
git clone https://github.com/JipZeonGit/migu-video-go.git
cd migu-video-go
```

### 编译

```shell
go build -o migu-server ./main.go
```

### 运行

```shell
./migu-server
```

若需要修改配置，可以使用以下命令。

Mac/Linux:

```shell
mport=3000 mhost=http://localhost:3000 ./migu-server
```

Windows 下使用 git-bash 等终端:

```shell
set mport=3000 && set mhost=http://localhost:3000 && ./migu-server.exe
```

Windows 下使用 PowerShell 等终端:

```shell
$Env:mport=3000; $Env:mhost="http://localhost:3000"; ./migu-server.exe
```

## 交叉编译

Go 原生支持交叉编译，无需 Docker QEMU：

```shell
# Linux amd64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o migu-server-linux-amd64 ./main.go

# Linux arm64
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o migu-server-linux-arm64 ./main.go

# Linux armv7
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-s -w" -o migu-server-linux-armv7 ./main.go

# Windows
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o migu-server.exe ./main.go
```

# 免责声明

> [!important]
>
> 1. 本仓库仅供学习使用，请尊重版权，请勿利用此仓库从事商业行为及非法用途！
> 2. 使用本仓库的过程中可能会产生版权数据。对于这些版权数据，本仓库不拥有它们的所有权。为了避免侵权，使用者务必在 24 小时内清除使用本仓库的过程中所产生的版权数据。
> 3. 由于使用本仓库产生的包括由于本协议或由于使用或无法使用本仓库而引起的任何性质的任何直接、间接、特殊、偶然或结果性损害（包括但不限于因商誉损失、停工、计算机故障或故障引起的损害赔偿，或任何及所有其他商业损害或损失）由使用者负责。
> 4. **禁止在违反当地法律法规的情况下使用本仓库。** 对于使用者在明知或不知当地法律法规不允许的情况下使用本仓库所造成的任何违法违规行为由使用者承担，本仓库不承担由此造成的任何直接、间接、特殊、偶然或结果性责任。
> 5. 如果官方平台觉得本仓库不妥，可联系本仓库更改或移除。
