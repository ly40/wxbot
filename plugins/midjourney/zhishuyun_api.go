package midjourney

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

const (
	proxy_url = "http://192.168.2.98:7890"
	timeout   = 600 //超时时间600秒
)

type ApiResponse struct {
	TaskId         string   `json:"task_id"`
	ImageId        string   `json:"image_id"`
	ImageUrl       string   `json:"image_url"`
	ImageLocalPath string   `json:"image_local_path"`
	Actions        []string `json:"actions"`
}
type ApiRequest struct {
	Prompt  string `json:"prompt"`
	ImageId string `json:"image_id"`
	Action  string `json:"action"`
}

func imagine(reqData ApiRequest) (*ApiResponse, error) {
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		log.Fatal(err)
	}

	// 创建代理 URL
	proxyURL, err := url.Parse(proxy_url)
	if err != nil {
		log.Fatal(err)
	}

	// 创建自定义请求头
	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")

	// 创建HTTP客户端，并设置代理和超时时间
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: timeout * time.Second,
	}

	var tokens []Token
	if err := db.Orm.Table("token").Find(&tokens).Error; err != nil {
		log.Fatal(err)
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, errors.New("请联系管理员设置token")
	}

	// 创建POST请求
	request, err := http.NewRequest("POST", "https://api.zhishuyun.com/midjourney/imagine?token="+tokens[0].Token, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("创建请求失败:", err)
	}

	// 设置请求头
	request.Header = headers

	// 发送请求
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("发送请求失败:", err)
	}
	defer response.Body.Close()

	// 检查响应的状态码
	if response.StatusCode >= 200 && response.StatusCode <= 299 {
		// 读取响应数据
		var result ApiResponse
		err = json.NewDecoder(response.Body).Decode(&result)
		if err != nil {
			fmt.Println("解析响应失败:", err)
			return nil, nil
		}

		if localPath, err := save_image(result.ImageUrl); err == nil {
			result.ImageLocalPath = localPath
		}

		jsonRes, err := json.Marshal(result)
		if err != nil {
			log.Fatal(err)
			return nil, nil
		}

		fmt.Println(string(jsonRes))
		return &result, nil
	} else {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("读取响应体失败: %v", err)
		}
		// 打印响应体数据
		fmt.Println(string(body))
		var errResp struct {
			Detail string `json:"detail"`
			Code   string `json:"code"`
		}
		err = json.NewDecoder(bytes.NewBuffer(body)).Decode(&errResp)
		if err != nil {
			fmt.Println("解析响应失败:", err)
			return nil, errors.New("解析响应失败")
		}
		fmt.Println("请求失败，响应状态码:", response.StatusCode)
		return nil, errors.New(errResp.Code)
	}
}
