package main

import (
	"embed"
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

type DNSRecord struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

// ProviderConfig 统一各厂商配置格式
type ProviderConfig struct {
	Name      string `json:"name"`
	AccessKey string `json:"accessKey"` // 对应 AK / Secret ID
	SecretKey string `json:"secretKey"` // 对应 SK / Secret Key / API Token
	Region    string `json:"region"`    // 仅华为云
	Email     string `json:"email"`     // 仅 Cloudflare
}

// AppConfig 整体配置文件
type AppConfig struct {
	Huawei     ProviderConfig `json:"huawei"`
	Aliyun     ProviderConfig `json:"aliyun"`
	Tencent    ProviderConfig `json:"tencent"`
	Cloudflare ProviderConfig `json:"cloudflare"`
}

type DNSService struct {
	config AppConfig
}

func NewDNSService() *DNSService {
	s := &DNSService{}
	s.LoadConfig()
	return s
}

func (s *DNSService) LoadConfig() AppConfig {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, "dns-manager-config-v2.json")
	data, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(data, &s.config)
	}
	return s.config
}

func (s *DNSService) SaveConfig(config AppConfig) bool {
	s.config = config
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, "dns-manager-config-v2.json")
	data, _ := json.MarshalIndent(config, "", "  ")
	err := os.WriteFile(configPath, data, 0644)
	return err == nil
}

// 模拟 API 调用（下一版替换为真实请求）
func (s *DNSService) GetDomains(provider string) []string {
	if provider == "huawei" && s.config.Huawei.AccessKey != "" {
		return []string{"huawei-real-test.cn"}
	}
	if provider == "cloudflare" && s.config.Cloudflare.SecretKey != "" {
		return []string{"cf-test.com"}
	}
	return []string{"[暂无数据或未配置有效密钥]"}
}

func (s *DNSService) GetRecords(provider, domain string) []DNSRecord {
	if domain != "[暂无数据或未配置有效密钥]" {
		return []DNSRecord{
			{ID: "1", Type: "A", Name: "test", Content: "等待接入真实API"},
		}
	}
	return []DNSRecord{}
}

func main() {
	app := NewDNSService()
	err := wails.Run(&options.App{
		Title:  "多厂商 DNS 管理终端 - 1Panel 风格",
		Width:  1080,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
