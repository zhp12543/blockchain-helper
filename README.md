# Fil-Address

一个用于 Filecoin 地址转换和区块链数据处理的 HTTP 服务工具。

## 功能特性

- **Filecoin 地址转换**：支持 Filecoin 地址（f0/f4）与以太坊地址（0x）之间的相互转换
- **ABI 编解码**：支持 Filecoin Send 方法的 ABI 数据编解码
- **EOS 数据解码**：支持 EOS 区块链交易数据的十六进制解码
- **区块链浏览器爬虫**：集成 OKLink 区块浏览器数据爬取功能

## 技术栈

- **语言**：Go 1.24
- **Web 框架**：Gin
- **区块链库**：
  - filecoin-project/go-address
  - filecoin-project/lotus
  - ethereum/go-ethereum
  - eoscanada/eos-go
- **爬虫工具**：chromedp
- **日志**：uber.org/zap

## 项目结构

```
fil-address/
├── main.go              # 主程序入口
├── config.yaml          # 配置文件
├── go.mod              # Go 模块依赖
├── abiencode/          # ABI 编解码模块
│   └── abi.go
├── chain/              # 区块链相关功能
│   └── oklink/         # OKLink 爬虫
│       ├── oklink.go
│       ├── types.go
│       └── oklink_test.go
├── config/             # 配置管理
│   └── config.go
├── http/               # HTTP 服务
│   ├── api/            # API 处理器
│   │   ├── home.go
│   │   └── oklink.go
│   └── routers/        # 路由配置
│       └── router.go
└── logger/             # 日志模块
    └── logger.go
```

## 安装与运行

### 环境要求

- Go 1.24+
- macOS/Linux

### 安装依赖

```bash
go mod download
```

### 配置文件

编辑 `config.yaml` 配置服务监听地址：

```yaml
Tcp: "127.0.0.1:8988"
```

### 运行

```bash
# 使用默认配置文件 ./config.yaml
go run main.go

# 指定配置文件路径
go run main.go -c /path/to/config.yaml

# 查看版本
go run main.go -v
```

## API 接口

### 1. 健康检查

```
GET /api/ping
```

**响应**：

```
pong
```

### 2. 版本信息

```
GET /
```

**响应**：

```json
{
  "version": "v1.0.0"
}
```

### 3. Filecoin 地址转换

```
GET /filAddress?address=<地址>
```

**参数**：

- `address`: Filecoin 地址（f0/f4 开头）或以太坊地址（0x 开头）

**示例请求**：

```bash
# Filecoin 地址转以太坊地址
curl "http://127.0.0.1:8988/filAddress?address=f410fmdno4kbhxnj3rb2a6gvvso3bv7qinwmqlpqgvq"

# 以太坊地址转 Filecoin 地址
curl "http://127.0.0.1:8988/filAddress?address=0x1234567890123456789012345678901234567890"
```

**响应**：

```json
{
  "code": "success",
  "eth_address": "0x...",
  "fil_address": "f410..."
}
```

### 4. 解析 Filecoin Send 交易数据

```
GET /send/filAddress?data=<hex_data>
```

**参数**：

- `data`: ABI 编码的十六进制数据（可选 0x 前缀）

**示例请求**：

```bash
curl "http://127.0.0.1:8988/send/filAddress?data=0x..."
```

**响应**：

```json
{
  "code": "success",
  "eth_address": "0x...",
  "fil_address": "f410..."
}
```

### 5. EOS 交易数据解码

```
GET /eosDecodeHex?data=<hex_data>
```

**参数**：

- `data`: EOS 交易的十六进制编码数据

**示例请求**：

```bash
curl "http://127.0.0.1:8988/eosDecodeHex?data=<hex_string>"
```

**响应**：

```json
{
  "code": "success",
  "data": {
    "from": "account1",
    "to": "account2",
    "quantity": "10.0000 EOS",
    "memo": "transfer memo"
  }
}
```

## 核心模块说明

### ABI 编解码（abiencode）

提供 Filecoin Send 方法的 ABI 编码和解码功能：

- `ABISendFilEncode`: 将地址编码为 ABI 格式
- `ABISendFilDecode`: 从 ABI 数据中解析地址

### OKLink 爬虫（chain/oklink）

使用 chromedp 爬取 OKLink 区块浏览器的交易详情，支持：

- 交易哈希查询
- 交易状态、区块高度、时间戳
- 发送方、接收方地址
- 转账金额和手续费信息

## 错误处理

所有 API 接口返回统一格式：

**成功响应**：

```json
{
  "code": "success",
  ...
}
```

**错误响应**：

```json
{
  "code": "error",
  "msg": "错误描述信息"
}
```

## 开发说明

### 添加新的 API 接口

1. 在 `http/api/` 目录下创建处理函数
2. 在 `http/routers/router.go` 的 `initRouter` 函数中注册路由
3. 重启服务即可生效

### 日志配置

日志模块使用 zap，可在 `logger/logger.go` 中配置日志级别和输出格式。

## 许可证

本项目代码仅供学习和研究使用。

## 作者

zhp12543

## 版本

当前版本：v1.0.0
