package main

import "fmt"

// DNSRecord 统一的 DNS 记录结构体
type DNSRecord struct {
	ID      string `json:"id"`
	Type    string `json:"type"`    // A, CNAME, TXT 等
	Name    string `json:"name"`    // 主机记录
	Content string `json:"content"` // 记录值
	TTL     int    `json:"ttl"`
}

// DNSProvider 定义统一的厂商操作接口
type DNSProvider interface {
	GetDomains() ([]string, error)
	GetRecords(domain string) ([]DNSRecord, error)
	AddRecord(domain string, record DNSRecord) error
	UpdateRecord(domain string, record DNSRecord) error
	DeleteRecord(domain string, recordID string) error
}

// 统一调度服务，供前端 Vue 调用
type DNSService struct {
	providers map[string]DNSProvider
}

func NewDNSService(cfToken, aliAccess, aliSecret, txId, txKey, hwAccess, hwSecret string) *DNSService {
	// 这里实例化各个厂商的具体 API 客户端
	return &DNSService{
		providers: map[string]DNSProvider{
			"cloudflare": &CloudflareProvider{Token: cfToken},
			"aliyun":     &AliyunProvider{AccessKey: aliAccess, Secret: aliSecret},
			"tencent":    &TencentProvider{SecretId: txId, SecretKey: txKey},
			"huawei":     &HuaweiProvider{AccessKey: hwAccess, SecretKey: hwSecret},
		},
	}
}

// 前端调用的入口：获取域名列表
func (s *DNSService) GetDomains(provider string) ([]string, error) {
	p, ok := s.providers[provider]
	if !ok {
		return nil, fmt.Errorf("不支持的 DNS 厂商")
	}
	return p.GetDomains()
}

// 前端调用的入口：获取记录列表
func (s *DNSService) GetRecords(provider, domain string) ([]DNSRecord, error) {
	return s.providers[provider].GetRecords(domain)
}

// (Add/Update/Delete 的包装方法以此类推...)
