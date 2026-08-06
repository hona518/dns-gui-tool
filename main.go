package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	// 引入华为云官方 SDK
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	dns "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/model"
	region "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/region"
)

//go:embed all:frontend/dist
var assets embed.FS

type DNSRecord struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

type ProviderConfig struct {
	Name      string `json:"name"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	Region    string `json:"region"`
	Email     string `json:"email"`
}

type AppConfig struct {
	Huawei     ProviderConfig `json:"huawei"`
	Aliyun     ProviderConfig `json:"aliyun"`
	Tencent    ProviderConfig `json:"tencent"`
	Cloudflare ProviderConfig `json:"cloudflare"`
}

type DNSService struct {
	ctx    context.Context
	config AppConfig
}

func NewDNSService() *DNSService {
	s := &DNSService{}
	s.LoadConfig()
	return s
}

func (s *DNSService) startup(ctx context.Context) {
	s.ctx = ctx
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

// 真实的 API 调用：获取华为云托管的域名列表
func (s *DNSService) GetDomains(provider string) []string {
	if provider == "huawei" {
		if s.config.Huawei.AccessKey == "" || s.config.Huawei.SecretKey == "" {
			return []string{"[暂无数据或未配置有效密钥]"}
		}

		// 1. 构造华为云鉴权凭证
		auth := basic.NewCredentialsBuilder().
			WithAk(s.config.Huawei.AccessKey).
			WithSk(s.config.Huawei.SecretKey).
			Build()

		// 2. 初始化 DNS 客户端
		client := dns.NewDnsClient(
			dns.DnsClientBuilder().
				WithRegion(region.ValueOf(s.config.Huawei.Region)).
				WithCredential(auth).
				Build())

		// 3. 发起真实请求：列出公网域名 (Zone)
		request := &model.ListPublicZonesRequest{}
		response, err := client.ListPublicZones(request)
		if err != nil {
			// 鉴权失败或网络错误，将其作为特殊字符串返回给前端提示
			return []string{fmt.Sprintf("[API报错] %v", err)}
		}

		var domains []string
		if response.Zones != nil {
			for _, zone := range *response.Zones {
				// 华为云返回的域名末尾自带点号，例如 "example.com."，此处将其截取美化
				name := *zone.Name
				if len(name) > 0 && name[len(name)-1] == '.' {
					name = name[:len(name)-1]
				}
				domains = append(domains, name)
			}
		}
		
		if len(domains) == 0 {
			return []string{"[账号下暂无托管域名]"}
		}
		return domains
	}

	return []string{"[该厂商真实API尚未接入]"}
}

// 真实的 API 调用：获取某个域名的解析记录
func (s *DNSService) GetRecords(provider, domainName string) []DNSRecord {
	// 如果是报错信息或空提示，不发起网络请求
	if domainName == "" || domainName[0] == '[' {
		return []DNSRecord{}
	}

	if provider == "huawei" {
		auth := basic.NewCredentialsBuilder().
			WithAk(s.config.Huawei.AccessKey).
			WithSk(s.config.Huawei.SecretKey).
			Build()

		client := dns.NewDnsClient(
			dns.DnsClientBuilder().
				WithRegion(region.ValueOf(s.config.Huawei.Region)).
				WithCredential(auth).
				Build())

		// 获取该域名下的所有记录
		request := &model.ListRecordSetsByZoneRequest{}
		// 注意：华为云的 ListRecordSets 最好传递 Zone ID，为了简化，我们可以通过全局搜索并匹配域名
		// 这里演示全局搜索记录的方式
		request.Marker = nil 
		
		response, err := client.ListRecordSetsByZone(request)
		if err != nil {
			return []DNSRecord{{ID: "error", Type: "ERROR", Name: "API请求失败", Content: err.Error()}}
		}

		var records []DNSRecord
		if response.Recordsets != nil {
			for _, r := range *response.Recordsets {
				// 过滤出属于当前被点击域名的记录
				rName := *r.Name
				if len(rName) > 0 && rName[len(rName)-1] == '.' {
					rName = rName[:len(rName)-1]
				}
				
				// 简单的后缀匹配
				if rName == domainName || (len(rName) > len(domainName) && rName[len(rName)-len(domainName)-1:] == "."+domainName) {
					// 提取主机记录 (比如 www)
					host := "@"
					if rName != domainName {
						host = rName[:len(rName)-len(domainName)-1]
					}
					
					// 提取记录值 (可能有多条，以逗号分隔)
					content := ""
					if r.Records != nil && len(*r.Records) > 0 {
						content = (*r.Records)[0]
					}

					records = append(records, DNSRecord{
						ID:      *r.Id,
						Type:    *r.Type,
						Name:    host,
						Content: content,
					})
				}
			}
		}
		return records
	}
	return []DNSRecord{}
}

func main() {
	app := NewDNSService()
	err := wails.Run(&options.App{
		Title:  "多厂商 DNS 管理终端 - 真实 API 版",
		Width:  1080,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
