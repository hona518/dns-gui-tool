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

// AppConfig 用于保存各个厂商的密钥
type AppConfig struct {
	HuaweiAK string `json:"huaweiAK"`
	HuaweiSK string `json:"huaweiSK"`
}

type DNSService struct {
	config AppConfig
}

func NewDNSService() *DNSService {
	s := &DNSService{}
	s.LoadConfig() // 启动时自动读取本地密钥
	return s
}

// 读取本地配置
func (s *DNSService) LoadConfig() AppConfig {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, "dns-manager-config.json")
	data, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(data, &s.config)
	}
	return s.config
}

// 保存配置到本地
func (s *DNSService) SaveConfig(config AppConfig) bool {
	s.config = config
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, "dns-manager-config.json")
	data, _ := json.MarshalIndent(config, "", "  ")
	err := os.WriteFile(configPath, data, 0644)
	return err == nil
}

// 联调测试：获取域名列表
func (s *DNSService) GetDomains(provider string) []string {
	if provider == "huawei" {
		// 校验：如果没有配置密钥，直接在界面上提示用户
		if s.config.HuaweiAK == "" || s.config.HuaweiSK == "" {
			return []string{"[请先点击右上角配置华为云密钥]"}
		}
		// 密钥存在，先返回联调占位符，下一步替换为真实的 HTTP 请求
		return []string{"huawei-real-test.cn"}
	}
	return []string{}
}

// 联调测试：获取解析记录
func (s *DNSService) GetRecords(provider, domain string) []DNSRecord {
	if provider == "huawei" && domain != "[请先点击右上角配置华为云密钥]" {
		return []DNSRecord{
			{ID: "1", Type: "A", Name: "测试连通性", Content: "获取成功，等待接入真实 API"},
		}
	}
	return []DNSRecord{}
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
			app,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
