package main

import (
	"bytes"
	"fmt"
	"github.com/gogf/gf/encoding/gjson"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
)

type Downloader struct {
	io.Reader
	Total   int64
	Current int64
}

func (d *Downloader) Read(p []byte) (n int, err error) {
	n, err = d.Reader.Read(p)
	d.Current += int64(n)
	fmt.Printf("\r正在下载，下载进度：%.2f%%", float64(d.Current*10000/d.Total)/100)
	if d.Current == d.Total {
		fmt.Printf("\r下载完成，下载进度：%.2f%%\n", float64(d.Current*10000/d.Total)/100)
	}
	return
}

func downloadFile(url, filePath string) {
	defer wg.Done()
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	file, err := os.Create(filePath)
	defer func() {
		_ = file.Close()
	}()
	downloader := &Downloader{
		Reader: resp.Body,
		Total:  resp.ContentLength,
	}
	if _, err := io.Copy(file, downloader); err != nil {
		log.Fatalln(err)
	}
}

var wg sync.WaitGroup

func main() {
	const mods = "./.minecraft/mods"

	defer func() {
		fmt.Println("程序结束,输入任意字符结束程序")
		var data string
		_, _ = fmt.Scanln(&data)
	}()

	if !IsExist(mods) {
		fmt.Println("找不到【.minecraft/mods】目录")
		return
	}

	headers := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
		"User-Agent":   "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/535.1 (KHTML, like Gecko) Chrome/14.0.835.163 Safari/535.1",
		"GGM":          "gg1",
	}
	resp,body:=HTTP("GET", "https://www.gongjubaike.cn/api/v1/mc/modsList", nil, headers)
	//fmt.Println(resp,body)
	if resp == nil {
		fmt.Println("请求最新mods信息失败")
	}

	JsonBody, err := gjson.DecodeToJson(body)
	if err != nil {
		fmt.Println("解析最新mods信息失败",err)
		return
	}


	task := make(map[string]string)
	for i := 0; i < JsonBody.Len("data"); i++ {
		name:=JsonBody.GetString(fmt.Sprint("data.",i,".Name"))
		Path:=fmt.Sprint(mods,"/",name)
		if IsExist(Path) {
			fmt.Println(name,"已存在")
			continue
		}
		task[fmt.Sprint("https://www.gongjubaike.cn/gg",JsonBody.GetString(fmt.Sprint("data.",i,".Url")))] = Path
	}
	for k, v := range task {
		wg.Add(1)
		fmt.Println("开始下载,",k)
		downloadFile(k, v)
	}
	wg.Wait()
}

//文件/目录是否存在
func IsExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}

//HTTP请求
func HTTP(method string, url string, data *gjson.Json, headers map[string]string) (*http.Response, string) {
	client := &http.Client{}
	var DyteData []byte
	if data != nil {
		DyteData = []byte(data.MustToJsonString())
	} else {
		DyteData = nil
	}
	bytesData2 := bytes.NewReader(DyteData)
	req, _ := http.NewRequest(method, url, bytesData2)
	for i, i2 := range headers {
		req.Header.Add(i, i2)
	}
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return resp, string(body)
}
