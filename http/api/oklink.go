package api

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhp12543/fil-address/chain/oklink"
	"github.com/zhp12543/fil-address/logger"
)

const (
	// 缓存过期时间：24小时
	cacheExpiration = 24 * time.Hour
)

// cacheEntry 缓存条目
type cacheEntry struct {
	data      *oklink.TxDetails
	timestamp time.Time
}

// txCache 交易缓存
type txCache struct {
	mu    sync.RWMutex
	cache map[string]*cacheEntry
}

// 全局缓存实例
var globalTxCache = &txCache{
	cache: make(map[string]*cacheEntry),
}

// getCacheKey 生成缓存键
func getCacheKey(symbol, txid string) string {
	return fmt.Sprintf("%s:%s", strings.ToLower(symbol), strings.ToLower(txid))
}

// get 从缓存获取数据
func (c *txCache) get(symbol, txid string) (*oklink.TxDetails, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := getCacheKey(symbol, txid)
	entry, exists := c.cache[key]
	
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Since(entry.timestamp) > cacheExpiration {
		return nil, false
	}

	return entry.data, true
}

// set 设置缓存
func (c *txCache) set(symbol, txid string, data *oklink.TxDetails) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := getCacheKey(symbol, txid)
	c.cache[key] = &cacheEntry{
		data:      data,
		timestamp: time.Now(),
	}
}

// cleanup 清理过期缓存
func (c *txCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.cache {
		if now.Sub(entry.timestamp) > cacheExpiration {
			delete(c.cache, key)
		}
	}
}

// 启动定期清理任务
func init() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		
		for range ticker.C {
			globalTxCache.cleanup()
			logger.Info("缓存清理完成")
		}
	}()
}

// OklinkCrawl 处理 OKLink 数据爬取请求
// 支持 GET 和 POST 方法
// 参数：
//   - symbol: 区块链符号（如 eth, btc, fil 等）
//   - txids: 交易哈希列表，多个用逗号分隔
//   - proxy: 可选的代理服务器地址（如 socks5://127.0.0.1:1080）
//   - nocache: 可选，设置为 "true" 时跳过缓存
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
	noCache := strings.ToLower(strings.TrimSpace(c.Query("nocache"))) == "true"

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

	// 处理缓存逻辑
	var results []*oklink.TxDetails
	var needFetch []string
	cacheHits := 0
	cacheMisses := 0

	if !noCache {
		// 尝试从缓存获取
		for _, txid := range txids {
			if cached, found := globalTxCache.get(symbol, txid); found {
				results = append(results, cached)
				cacheHits++
				logger.Infof("缓存命中: %s:%s", symbol, txid)
			} else {
				needFetch = append(needFetch, txid)
				cacheMisses++
			}
		}
	} else {
		needFetch = txids
		logger.Info("跳过缓存，强制抓取")
	}

	// 抓取未缓存的数据
	if len(needFetch) > 0 {
		logger.Infof("需要抓取 %d 个交易", len(needFetch))
		details, err := oklink.CrawlOklink(symbol, needFetch, "")
		if err != nil {
			// 如果有部分缓存数据，仍然返回
			if len(results) > 0 {
				ret["code"] = "partial_success"
				ret["msg"] = fmt.Sprintf("部分数据抓取失败: %v", err)
				ret["data"] = results
				ret["count"] = len(results)
				ret["cache_hits"] = cacheHits
				ret["cache_misses"] = cacheMisses
				return
			}
			ret["msg"] = err.Error()
			return
		}

		// 将新数据加入缓存
		for _, detail := range details {
			if detail != nil && detail.Hash != "" {
				globalTxCache.set(symbol, detail.Hash, detail)
				results = append(results, detail)
			}
		}
	}

	// 返回成功结果
	ret["code"] = "success"
	ret["data"] = results
	ret["count"] = len(results)
	ret["cache_hits"] = cacheHits
	ret["cache_misses"] = cacheMisses
	
	if cacheHits > 0 {
		logger.Infof("缓存统计 - 命中: %d, 未命中: %d", cacheHits, cacheMisses)
	}
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
