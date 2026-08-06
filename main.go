package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

type DNSRecord struct {
	ID      string `json:"id"`
	ZoneID  string `json:"zoneId"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Line    string `json:"line"`
	TTL     int32  `json:"ttl"`
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

func (s *DNSService) startup(ctx context.Context) { s.ctx = ctx }

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

func HexSha256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func HmacSha256(key []byte, data string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(data))
	return mac.Sum(nil)
}

func (s *DNSService) hwApiRequest(method, path string, body []byte) ([]byte, error) {
	region := s.config.Huawei.Region
	if region == "" {
		region = "cn-north-1"
	}
	host := fmt.Sprintf("dns.%s.myhuaweicloud.com", region)

	uri := path
	query := ""
	if idx := strings.Index(path, "?"); idx != -1 {
		uri = path[:idx]
		query = path[idx+1:]
	}

	urlStr := fmt.Sprintf("https://%s%s", host, path)
	req, err := http.NewRequest(method, urlStr, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	t := time.Now().UTC()
	dateStr := t.Format("20060102T150405Z")

	req.Header.Set("X-Sdk-Date", dateStr)
	req.Header.Set("Host", host)

	var signedHeaders, canonicalHeaders string
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
		signedHeaders = "content-type;host;x-sdk-date"
		canonicalHeaders = fmt.Sprintf("content-type:application/json\nhost:%s\nx-sdk-date:%s\n", host, dateStr)
	} else {
		signedHeaders = "host;x-sdk-date"
		canonicalHeaders = fmt.Sprintf("host:%s\nx-sdk-date:%s\n", host, dateStr)
	}

	if !strings.HasSuffix(uri, "/") {
		uri += "/"
	}

	bodyHash := HexSha256(body)
	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s", method, uri, query, canonicalHeaders, signedHeaders, bodyHash)

	stringToSign := fmt.Sprintf("SDK-HMAC-SHA256\n%s\n%s", dateStr, HexSha256([]byte(canonicalRequest)))
	signature := hex.EncodeToString(HmacSha256([]byte(s.config.Huawei.SecretKey), stringToSign))
	authHeader := fmt.Sprintf("SDK-HMAC-SHA256 Access=%s, SignedHeaders=%s, Signature=%s", s.config.Huawei.AccessKey, signedHeaders, signature)

	req.Header.Set("Authorization", authHeader)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

type HwZoneResp struct {
	Zones []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"zones"`
}

type HwRecordResp struct {
	Recordsets []struct {
		ID      string   `json:"id"`
		ZoneID  string   `json:"zone_id"`
		Name    string   `json:"name"`
		Type    string   `json:"type"`
		Records []string `json:"records"`
		Line    string   `json:"line"`
		TTL     int32    `json:"ttl"`
	} `json:"recordsets"`
}

type HwRecordPayload struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Records []string `json:"records"`
	TTL     int32    `json:"ttl"`
	Line    string   `json:"line"`
}

func (s *DNSService) GetDomains(provider string) []string {
	if provider != "huawei" {
		return []string{"[该厂商尚未接入]"}
	}
	if s.config.Huawei.AccessKey == "" {
		return []string{"[未配置有效密钥]"}
	}

	body, err := s.hwApiRequest("GET", "/v2/zones", nil)
	if err != nil {
		return []string{fmt.Sprintf("[API报错] %v", err)}
	}

	var data HwZoneResp
	json.Unmarshal(body, &data)

	var domains []string
	for _, z := range data.Zones {
		name := z.Name
		if len(name) > 0 && name[len(name)-1] == '.' {
			name = name[:len(name)-1]
		}
		domains = append(domains, name)
	}
	if len(domains) == 0 {
		return []string{"[账号下暂无域名]"}
	}
	return domains
}

func (s *DNSService) getZoneIdByName(domainName string) string {
	body, err := s.hwApiRequest("GET", "/v2/zones?name="+domainName, nil)
	if err != nil {
		return ""
	}
	var data HwZoneResp
	json.Unmarshal(body, &data)
	for _, z := range data.Zones {
		name := z.Name
		if len(name) > 0 && name[len(name)-1] == '.' {
			name = name[:len(name)-1]
		}
		if name == domainName {
			return z.ID
		}
	}
	return ""
}

func (s *DNSService) GetRecords(provider, domainName string) []DNSRecord {
	if domainName == "" || domainName[0] == '[' {
		return []DNSRecord{}
	}
	if provider != "huawei" {
		return []DNSRecord{}
	}

	zoneId := s.getZoneIdByName(domainName)
	if zoneId == "" {
		return []DNSRecord{{ID: "error", Type: "ERROR", Name: "查询失败", Content: "无法获取域名的Zone ID"}}
	}

	body, err := s.hwApiRequest("GET", fmt.Sprintf("/v2/zones/%s/recordsets?limit=500", zoneId), nil)
	if err != nil {
		return []DNSRecord{{ID: "error", Type: "ERROR", Name: "API请求失败", Content: err.Error()}}
	}

	var data HwRecordResp
	json.Unmarshal(body, &data)

	var records []DNSRecord
	for _, r := range data.Recordsets {
		rName := r.Name
		if len(rName) > 0 && rName[len(rName)-1] == '.' {
			rName = rName[:len(rName)-1]
		}

		if rName == domainName || (len(rName) > len(domainName) && rName[len(rName)-len(domainName)-1:] == "."+domainName) {
			host := "@"
			if rName != domainName {
				host = rName[:len(rName)-len(domainName)-1]
			}
			content := ""
			if len(r.Records) > 0 {
				content = r.Records[0]
			}
			line := r.Line
			if line == "" {
				line = "default"
			}

			records = append(records, DNSRecord{
				ID:      r.ID,
				ZoneID:  r.ZoneID,
				Type:    r.Type,
				Name:    host,
				Content: content,
				Line:    line,
				TTL:     r.TTL,
			})
		}
	}

	// 🛠️ 上帝视角调试模式：将原始报文强制作为一条记录展示在表格最下方
	records = append(records, DNSRecord{
		ID:      "debug_raw",
		ZoneID:  zoneId,
		Type:    "DEBUG",
		Name:    "底层原始JSON",
		Content: string(body),
		Line:    "上帝视角",
		TTL:     0,
	})

	return records
}

func (s *DNSService) AddRecord(provider, domainName string, record DNSRecord) error {
	if provider != "huawei" {
		return fmt.Errorf("暂未支持该厂商")
	}
	zoneId := record.ZoneID
	if zoneId == "" {
		zoneId = s.getZoneIdByName(domainName)
	}

	fullRecordName := domainName + "."
	if record.Name != "@" && record.Name != "" {
		fullRecordName = record.Name + "." + domainName + "."
	}

	ttl := record.TTL
	if ttl == 0 {
		ttl = 300
	}
	line := record.Line
	if line == "" {
		line = "default"
	}

	payload := HwRecordPayload{
		Name:    fullRecordName,
		Type:    record.Type,
		Records: []string{record.Content},
		TTL:     ttl,
		Line:    line,
	}

	jsonData, _ := json.Marshal(payload)
	_, err := s.hwApiRequest("POST", fmt.Sprintf("/v2/zones/%s/recordsets", zoneId), jsonData)
	return err
}

func (s *DNSService) UpdateRecord(provider, domainName string, record DNSRecord) error {
	if provider != "huawei" {
		return fmt.Errorf("暂未支持该厂商")
	}

	fullRecordName := domainName + "."
	if record.Name != "@" && record.Name != "" {
		fullRecordName = record.Name + "." + domainName + "."
	}

	line := record.Line
	if line == "" {
		line = "default"
	}

	payload := HwRecordPayload{
		Name:    fullRecordName,
		Type:    record.Type,
		Records: []string{record.Content},
		TTL:     record.TTL,
		Line:    line,
	}

	jsonData, _ := json.Marshal(payload)
	_, err := s.hwApiRequest("PUT", fmt.Sprintf("/v2/zones/%s/recordsets/%s", record.ZoneID, record.ID), jsonData)
	return err
}

func (s *DNSService) DeleteRecord(provider, zoneID, recordID string) error {
	if provider != "huawei" {
		return fmt.Errorf("暂未支持该厂商")
	}
	_, err := s.hwApiRequest("DELETE", fmt.Sprintf("/v2/zones/%s/recordsets/%s", zoneID, recordID), nil)
	return err
}

func main() {
	app := NewDNSService()
	err := wails.Run(&options.App{
		Title:  "多厂商 DNS 管理终端 - 纯净原生版",
		Width:  1200,
		Height: 850,
		AssetServer: &assetserver.Options{Assets: assets},
		OnStartup:   app.startup,
		Bind:        []interface{}{app},
	})
	if err != nil {
		log.Fatal(err)
	}
}
