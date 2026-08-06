package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

// DNSRecord 结构
type DNSRecord struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

// DNSService 绑定给前端
type DNSService struct{}

func NewDNSService() *DNSService { return &DNSService{} }

// 模拟获取域名列表（供前端交互测试）
func (s *DNSService) GetDomains(provider string) []string {
	switch provider {
	case "cloudflare":
		return []string{"cf-domain.com", "example.net"}
	case "aliyun":
		return []string{"aliyun-test.cn"}
	case "tencent":
		return []string{"tencent-cloud.com"}
	case "huawei":
		return []string{"huawei-tech.cn"}
	}
	return []string{}
}

// 模拟获取解析记录
func (s *DNSService) GetRecords(provider, domain string) []DNSRecord {
	return []DNSRecord{
		{ID: "1", Type: "A", Name: "www", Content: "192.168.1.1"},
		{ID: "2", Type: "CNAME", Name: "blog", Content: "cname.test.com"},
	}
}

// 模拟删除操作
func (s *DNSService) DeleteRecord(provider, domain, id string) bool {
	return true
}

func main() {
	app := NewDNSService()
	
	err := wails.Run(&options.App{
		Title:  "多厂商 DNS 管理终端",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Bind: []interface{}{
			app, // 绑定后端服务
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
