package midjourney

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
)

func save_image(image_url string) (string, error) {

	// 代理 URL
	proxyUrl, err := url.Parse(proxy_url)
	if err != nil {
		return "", err
	}
	// 创建HTTP客户端
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		},
	}

	fileName := path.Base(image_url)

	// 创建文件
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	out, err := os.Create(path.Join(cwd, "data/plugins/midjourney/images", fileName))
	if err != nil {
		return "", err
	}
	defer out.Close()

	// 发送GET请求
	resp, err := client.Get(image_url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("HTTP status code")
	}

	// 下载图片并保存到文件
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return path.Join("data/plugins/midjourney/images", fileName), nil
}
