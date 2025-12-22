package oklink

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/shopspring/decimal"
	"log"
	"regexp"
	"strings"
	"time"
)

func getOkLinkUrl(symbol, txid string) string {
	return fmt.Sprintf("https://www.oklink.com/zh-hans/%s/tx/%s", symbol, txid)
}

func crawlOklink(symbol string, txids []string) []*TxDetails {
	proxyServer := "socks5://127.0.0.1:51837"

	// 配置 Chrome 选项
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	if proxyServer != "" {
		log.Printf("使用代理: %s", proxyServer)
		opts = append(opts, chromedp.ProxyServer(proxyServer))
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ret := make([]*TxDetails, 0)
	for _, txid := range txids {
		func() {
			url := getOkLinkUrl(symbol, txid)
			ctx, cancel := chromedp.NewContext(allocCtx)
			defer cancel()

			ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
			defer cancel()

			details, err := fetchTxDetails(ctx, url)
			if err != nil {
				log.Printf("抓取失败 %s: %v", url, err)
				return
			}

			ret = append(ret, &details)
			// 打印结果
			fmt.Println("--------------------------------------------------")
			fmt.Printf("URL:     %s\n", url)
			fmt.Printf("交易哈希: %s\n", details.Hash)
			fmt.Printf("交易结果: %s\n", details.Status)
			fmt.Printf("区块:    %s\n", details.Block)
			fmt.Printf("时间:    %s\n", details.Time)

			if len(details.Transfers) > 0 {
				fmt.Println("--- 转账列表 ---")
				for i, tr := range details.Transfers {
					fmt.Printf("[%d] 发送方: %s\n", i+1, tr.Sender)
					fmt.Printf("    接收方: %s\n", tr.Receiver)
					fmt.Printf("    数量:   %s %s\n", tr.Amount, tr.Token)
				}
			} else {
				fmt.Printf("发送方:  %s\n", details.Sender)
				fmt.Printf("接收方:  %s\n", details.Receiver)
				fmt.Printf("交易数量: %s %s\n", details.Amount, details.Token)
			}

			fmt.Printf("手续费:  %s\n", details.Fee)
			fmt.Println("--------------------------------------------------")
		}()
	}
	return ret
}

func fetchTxDetails(ctx context.Context, url string) (TxDetails, error) {
	var details TxDetails
	var pageTitle string
	var bodyText string
	var links []LinkMap

	log.Printf("开始抓取: %s", url)

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Title(&pageTitle),
		chromedp.Sleep(5*time.Second),
		chromedp.Text(`body`, &bodyText, chromedp.ByQuery),
		// 获取所有链接的文本和 href，用于解析非 0x 地址
		chromedp.Evaluate(`Array.from(document.querySelectorAll('a')).map(a => ({text: a.innerText.trim(), href: a.href}))`, &links),
	)

	if err != nil {
		return details, err
	}

	log.Printf("页面标题: %s", pageTitle)

	// 构建链接映射表
	linkMap := make(map[string]string)
	for _, l := range links {
		if l.Text != "" && l.Href != "" {
			linkMap[l.Text] = l.Href
		}
	}

	// 辅助函数：解析地址
	resolveAddress := func(text string) string {
		text = cleanText(text)
		// 如果已经是 0x 开头，直接返回
		if strings.HasPrefix(text, "0x") && len(text) >= 42 {
			return text
		}
		// 尝试从链接映射中查找
		if href, ok := linkMap[text]; ok {
			// 尝试从 href 中提取地址 (例如 .../address/0x...)
			reAddr := regexp.MustCompile(`(0x[0-9a-fA-F]{40})`)
			matches := reAddr.FindStringSubmatch(href)
			if len(matches) > 1 {
				return matches[1]
			}
		}
		return text
	}

	// 基础信息提取
	details.Hash = cleanText(extractByRegex(bodyText, `(?:交易哈希|Transaction Hash)[：:]?\s*([0-9a-fxA-F]+)`))
	details.Status = cleanText(extractByRegex(bodyText, `(?:交易结果|Status)[：:]?\s*(\S+)`))
	details.Block = cleanText(extractByRegex(bodyText, `(?:区块|Block)[：:]?\s*(\d+)`))
	details.Time = cleanText(extractByRegex(bodyText, `(?:时间|Timestamp)[：:]?\s*([\d/ :]+)`))

	// 尝试提取概览中的发送方/接收方 (作为 fallback)
	details.Sender = resolveAddress(extractByRegex(bodyText, `(?:发送方|From)[：:]?\s*([0-9a-fxA-F]+)`))
	details.Receiver = resolveAddress(extractByRegex(bodyText, `(?:接收方|To|Interacted With \(To\))[：:]?\s*([0-9a-fxA-F]+)`))

	// 提取手续费
	feeStr := extractByRegex(bodyText, `(?:交易手续费|Transaction Fee)[：:]?\s*([\d,.]+)`)
	feeStr = strings.ReplaceAll(feeStr, ",", "")
	if feeStr != "" {
		details.Fee, _ = decimal.NewFromString(feeStr)
	}

	// 提取交易数量 (概览中的总数量)
	// 修改正则以同时捕获 Token
	reAmount := regexp.MustCompile(`(?:交易数量|Value)[：:]?\s*([\d,.]+)\s*([a-zA-Z]+)`)
	matchesAmount := reAmount.FindStringSubmatch(bodyText)
	if len(matchesAmount) > 2 {
		amountStr := strings.ReplaceAll(matchesAmount[1], ",", "")
		details.Amount, _ = decimal.NewFromString(amountStr)
		details.Token = matchesAmount[2]
	} else {
		// Fallback: 只提取数字
		amountStr := extractByRegex(bodyText, `(?:交易数量|Value)[：:]?\s*([\d,.]+)`)
		amountStr = strings.ReplaceAll(amountStr, ",", "")
		if amountStr != "" {
			details.Amount, _ = decimal.NewFromString(amountStr)
		}
	}

	// 提取转账列表 (针对多笔转账或新页面结构)
	// 模式：发送方\n<val>\n接收方\n<val>\n转账\n<val> <token>
	// 使用 (?s) 开启 dot matches newline 模式，或者显式匹配 \s*
	// 注意：(.*?) 是非贪婪匹配
	reTransfer := regexp.MustCompile(`(?:发送方|From)\s*\n\s*(.*?)\s*\n\s*(?:接收方|To)\s*\n\s*(.*?)\s*\n\s*(?:转账|Value)\s*\n\s*([\d,.]+)\s*([A-Z]+)`)
	matches := reTransfer.FindAllStringSubmatch(bodyText, -1)

	for _, m := range matches {
		amtStr := strings.ReplaceAll(m[3], ",", "")
		amt, _ := decimal.NewFromString(amtStr)
		details.Transfers = append(details.Transfers, Transfer{
			Sender:   resolveAddress(m[1]),
			Receiver: resolveAddress(m[2]),
			Amount:   amt,
			Token:    cleanText(m[4]),
		})
	}

	return details, nil
}

func cleanText(s string) string {
	return strings.TrimSpace(s)
}

func extractByRegex(text, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
