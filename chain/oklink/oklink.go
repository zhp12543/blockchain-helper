package oklink

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/shopspring/decimal"
	"github.com/zhp12543/fil-address/logger"
)

const (
	defaultTimeout     = 60 * time.Second
	defaultWaitTime    = 5 * time.Second
	defaultUserAgent   = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

// getOkLinkURL 构建 OKLink 交易查询 URL
func getOkLinkURL(symbol, txid string) string {
	return fmt.Sprintf("https://www.oklink.com/zh-hans/%s/tx/%s", symbol, txid)
}

// CrawlOklink 抓取 OKLink 交易详情
// symbol: 区块链符号（如 eth, btc 等）
// txids: 交易哈希列表
// proxyServer: 可选的代理服务器地址（格式：socks5://host:port 或 http://host:port），空字符串表示不使用代理
func CrawlOklink(symbol string, txids []string, proxyServer string) ([]*TxDetails, error) {
	if len(txids) == 0 {
		return nil, fmt.Errorf("txids cannot be empty")
	}

	// 配置 Chrome 选项
	opts := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.UserAgent(defaultUserAgent),
	}

	if proxyServer != "" {
		logger.Infof("使用代理服务器: %s", proxyServer)
		opts = append(opts, chromedp.ProxyServer(proxyServer))
	}

	// 创建浏览器分配器上下文
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	results := make([]*TxDetails, 0, len(txids))
	
	// 遍历每个交易哈希
	for _, txid := range txids {
		details, err := fetchSingleTx(allocCtx, symbol, txid)
		if err != nil {
			logger.Errorf("抓取交易失败 [%s]: %v", txid, err)
			continue
		}
		
		results = append(results, details)
		logTxDetails(details)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("failed to fetch any transaction details")
	}

	return results, nil
}

// fetchSingleTx 抓取单个交易详情
func fetchSingleTx(allocCtx context.Context, symbol, txid string) (*TxDetails, error) {
	if txid == "" {
		return nil, fmt.Errorf("txid cannot be empty")
	}
	
	url := getOkLinkURL(symbol, txid)
	logger.Infof("开始抓取交易: %s", txid)
	
	// 创建浏览器上下文
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// 设置超时
	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, defaultTimeout)
	defer timeoutCancel()

	details, err := fetchTxDetails(timeoutCtx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch tx details for %s: %w", txid, err)
	}

	return &details, nil
}

// logTxDetails 记录交易详情日志
func logTxDetails(details *TxDetails) {
	logger.Info("--------------------------------------------------")
	logger.Infof("交易哈希: %s", details.Hash)
	logger.Infof("交易结果: %s", details.Status)
	logger.Infof("区块:    %s", details.Block)
	logger.Infof("时间:    %s", details.Time)

	if len(details.Transfers) > 0 {
		logger.Info("--- 转账列表 ---")
		for i, tr := range details.Transfers {
			logger.Infof("[%d] 发送方: %s", i+1, tr.Sender)
			logger.Infof("    接收方: %s", tr.Receiver)
			logger.Infof("    数量:   %s %s", tr.Amount, tr.Token)
		}
	} else {
		logger.Infof("发送方:  %s", details.Sender)
		logger.Infof("接收方:  %s", details.Receiver)
		logger.Infof("交易数量: %s %s", details.Amount, details.Token)
	}

	logger.Infof("手续费:  %s", details.Fee)
	logger.Info("--------------------------------------------------")
}

// fetchTxDetails 从页面抓取交易详情
func fetchTxDetails(ctx context.Context, url string) (TxDetails, error) {
	var details TxDetails
	
	// 抓取页面数据
	pageData, err := fetchPageData(ctx, url)
	if err != nil {
		return details, err
	}

	logger.Infof("页面标题: %s", pageData.Title)

	// 构建地址解析器
	resolver := newAddressResolver(pageData.Links)

	// 提取基础信息
	extractBasicInfo(&details, pageData.BodyText, resolver)

	// 提取金额信息
	extractAmountInfo(&details, pageData.BodyText)

	// 提取转账列表
	extractTransfers(&details, pageData.BodyText, resolver)

	return details, nil
}

// pageData 存储页面数据
type pageData struct {
	Title    string
	BodyText string
	Links    []LinkMap
}

// fetchPageData 抓取页面数据
func fetchPageData(ctx context.Context, url string) (*pageData, error) {
	var data pageData

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Title(&data.Title),
		chromedp.Sleep(defaultWaitTime),
		chromedp.Text(`body`, &data.BodyText, chromedp.ByQuery),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('a')).map(a => ({text: a.innerText.trim(), href: a.href}))`, &data.Links),
	)

	if err != nil {
		return nil, fmt.Errorf("chromedp run failed: %w", err)
	}

	return &data, nil
}

// addressResolver 地址解析器
type addressResolver struct {
	linkMap map[string]string
	addrRe  *regexp.Regexp
}

// newAddressResolver 创建地址解析器
func newAddressResolver(links []LinkMap) *addressResolver {
	linkMap := make(map[string]string)
	for _, l := range links {
		if l.Text != "" && l.Href != "" {
			linkMap[l.Text] = l.Href
		}
	}

	return &addressResolver{
		linkMap: linkMap,
		addrRe:  regexp.MustCompile(`(0x[0-9a-fA-F]{40})`),
	}
}

// resolve 解析地址
func (r *addressResolver) resolve(text string) string {
	text = cleanText(text)
	
	// 如果已经是 0x 开头的完整地址，直接返回
	if strings.HasPrefix(text, "0x") && len(text) >= 42 {
		return text
	}
	
	// 尝试从链接映射中查找
	if href, ok := r.linkMap[text]; ok {
		matches := r.addrRe.FindStringSubmatch(href)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	
	return text
}

// extractBasicInfo 提取基础信息
func extractBasicInfo(details *TxDetails, bodyText string, resolver *addressResolver) {
	details.Hash = cleanText(extractByRegex(bodyText, `(?:交易哈希|Transaction Hash)[：:]?\s*([0-9a-fxA-F]+)`))
	details.Status = cleanText(extractByRegex(bodyText, `(?:交易结果|Status)[：:]?\s*(\S+)`))
	details.Block = cleanText(extractByRegex(bodyText, `(?:区块|Block)[：:]?\s*(\d+)`))
	details.Time = cleanText(extractByRegex(bodyText, `(?:时间|Timestamp)[：:]?\s*([\d/ :]+)`))

	// 提取地址
	senderText := extractByRegex(bodyText, `(?:发送方|From)[：:]?\s*([0-9a-fxA-F]+)`)
	receiverText := extractByRegex(bodyText, `(?:接收方|To|Interacted With \(To\))[：:]?\s*([0-9a-fxA-F]+)`)
	
	details.Sender = resolver.resolve(senderText)
	details.Receiver = resolver.resolve(receiverText)
}

// extractAmountInfo 提取金额信息
func extractAmountInfo(details *TxDetails, bodyText string) {
	// 提取手续费
	feeStr := extractByRegex(bodyText, `(?:交易手续费|Transaction Fee)[：:]?\s*([\d,.]+)`)
	if feeStr != "" {
		feeStr = strings.ReplaceAll(feeStr, ",", "")
		details.Fee, _ = decimal.NewFromString(feeStr)
	}

	// 提取交易数量和代币符号
	amountRe := regexp.MustCompile(`(?:交易数量|Value)[：:]?\s*([\d,.]+)\s*([a-zA-Z]+)`)
	if matches := amountRe.FindStringSubmatch(bodyText); len(matches) > 2 {
		amountStr := strings.ReplaceAll(matches[1], ",", "")
		details.Amount, _ = decimal.NewFromString(amountStr)
		details.Token = matches[2]
	} else {
		// Fallback: 只提取数字
		amountStr := extractByRegex(bodyText, `(?:交易数量|Value)[：:]?\s*([\d,.]+)`)
		if amountStr != "" {
			amountStr = strings.ReplaceAll(amountStr, ",", "")
			details.Amount, _ = decimal.NewFromString(amountStr)
		}
	}
}

// extractTransfers 提取转账列表
func extractTransfers(details *TxDetails, bodyText string, resolver *addressResolver) {
	transferRe := regexp.MustCompile(`(?:发送方|From)\s*\n\s*(.*?)\s*\n\s*(?:接收方|To)\s*\n\s*(.*?)\s*\n\s*(?:转账|Value)\s*\n\s*([\d,.]+)\s*([A-Z]+)`)
	matches := transferRe.FindAllStringSubmatch(bodyText, -1)

	for _, m := range matches {
		if len(m) < 5 {
			continue
		}
		
		amountStr := strings.ReplaceAll(m[3], ",", "")
		amount, err := decimal.NewFromString(amountStr)
		if err != nil {
			continue
		}

		details.Transfers = append(details.Transfers, Transfer{
			Sender:   resolver.resolve(m[1]),
			Receiver: resolver.resolve(m[2]),
			Amount:   amount,
			Token:    cleanText(m[4]),
		})
	}
}

// cleanText 清理文本，移除前后空白
func cleanText(s string) string {
	return strings.TrimSpace(s)
}

// extractByRegex 使用正则表达式提取文本
func extractByRegex(text, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
