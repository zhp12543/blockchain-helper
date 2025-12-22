package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zhp12543/fil-address/chain/oklink"
)

// OklinkCrawl 处理 OKLink 数据爬取请求
// 支持 GET 和 POST 方法
// 参数：
//   - symbol: 区块链符号（如 eth, btc, fil 等）
//   - txids: 交易哈希列表，多个用逗号分隔
//   - proxy: 可选的代理服务器地址（如 socks5://127.0.0.1:1080）
func OklinkCrawl(c *gin.Context) {
	ret := make(map[string]interface{})
	ret["code"] = "error"

	defer func() {
		c.JSON(http.StatusOK, ret)
	}()

	// 从 URL 参数获取参数
	symbol := strings.TrimSpace(c.Query("symbol"))
	txidsStr := strings.TrimSpace(c.Query("txids"))
	//proxyServer := strings.TrimSpace(c.Query("proxy"))

	// 验证必填参数
	if symbol == "" {
		ret["msg"] = "参数 symbol 不能为空"
		return
	}

	if txidsStr == "" {
		ret["msg"] = "参数 txids 不能为空"
		return
	}

	// 解析交易哈希列表
	txids := parseTxids(txidsStr)
	if len(txids) == 0 {
		ret["msg"] = "参数 txids 格式错误，请使用逗号分隔多个交易哈希"
		return
	}

	// 调用爬虫函数
	details, err := oklink.CrawlOklink(symbol, txids, "")
	if err != nil {
		ret["msg"] = err.Error()
		return
	}

	// 返回成功结果
	ret["code"] = "success"
	ret["data"] = details
	ret["count"] = len(details)
}

// parseTxids 解析交易哈希列表
// 支持逗号、分号、空格分隔
func parseTxids(txidsStr string) []string {
	// 替换分号和换行符为逗号
	txidsStr = strings.ReplaceAll(txidsStr, ";", ",")
	txidsStr = strings.ReplaceAll(txidsStr, "\n", ",")
	txidsStr = strings.ReplaceAll(txidsStr, "\r", ",")

	// 按逗号分割
	parts := strings.Split(txidsStr, ",")
	
	txids := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			txids = append(txids, part)
		}
	}

	return txids
}
