package main

import (
	"config"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var confUrl string
var lastTime int
var storePath string
var lastTimePath string

type jbListInfo struct {
	url   string
	time  string
	title string
}

func main() {
	myConfig := new(conf.Config)
	myConfig.InitConfig("./Spider.conf")
	confUrl = myConfig.Read("webInfo", "url")
	storePath = myConfig.Read("pathInfo", "rstInfo")
	lastTimePath = myConfig.Read("pathInfo", "oldPath")
	GetlastTime(myConfig)
	fmt.Println("-->爬取网址为 ：", confUrl)
	DoWork(confUrl)
	fmt.Println("***************爬取完毕*******10秒后关闭此窗口*******")
	time.Sleep(time.Second * 10)
}

//主工作方法 。
func DoWork(srcUrl string) {
	var dirName int
	fmt.Println("-->正在爬取。。。。")
	rst, err2 := HttpGet(srcUrl)
	if err2 != nil {
		fmt.Println("-->HttpGet(srcUrl)")
		return
	}
	listInfo := GetJbListInfo(rst)
	if listInfo == nil {
		fmt.Println("获取简报列表失败，可能是网站已经改版")
	}
	for _, v := range listInfo {
		temp, err22 := strconv.Atoi(v.time)
		if err22 != nil {
			fmt.Println("存在时间格式不正确", v.time)
			continue
		}
		if dirName < temp {
			dirName = temp
		}
	}
	fmt.Println("debug__________________________________")
	if lastTime >= dirName {
		return
	}
	fmt.Println("debug__________________________________time is ", lastTime, dirName)

	for _, v := range listInfo {
		rst, err11 := GetJbBody(v.url)
		if err11 != nil {
			fmt.Println(" GetJbBody() is ", err11, rst)
		}
		time, _ := strconv.Atoi(v.time)
		if time > lastTime {
			StoreTOFile(v, rst)
		}

	}
	SetlastTime(strconv.Itoa(dirName))
}

func HttpGet(url string) (rst string, err error) {
	resp, err1 := http.Get(url)
	if err1 != nil {
		err = err1
		return
	}
	defer resp.Body.Close()

	//读取网页内容
	buf := make([]byte, 4*1024)
	for {
		n, _ := resp.Body.Read(buf)
		if n == 0 {
			break
		}
		rst += string(buf[:n])
	}
	return rst, err
}

func GetJbListInfo(rst string) map[int]jbListInfo {
	//var ListInfo map(int)jbListInfo
	listInfo := make(map[int]jbListInfo)
	reg := regexp.MustCompile(`<li>(?s:(.*?))</li>`)
	if reg == nil {
		fmt.Errorf("-->匹配正则表达式出错 regexp.MustCompile err")
		return listInfo
	}
	htmlLi := reg.FindAllStringSubmatch(rst, -1)
	if htmlLi == nil {
		fmt.Println("reg.FindAllStringSubmatch(rst, -1) ERROR")
		return listInfo
	}
	for index, data := range htmlLi {
		var temp jbListInfo
		if !strings.Contains(data[1], "<span style=\"text-align:center\">[") {
			continue
		}
		//fmt.Println(data[1])

		//获取时间信息。 start
		regDate := regexp.MustCompile(`<span style="text-align:center">\[(.*)]</span>`)
		if regDate == nil {
			fmt.Println("regexp.MustCompile(`<span style=\"text-align:center\">[(?s:(.*?))]</span>`) error")
		}
		date := regDate.FindAllStringSubmatch(data[1], -1)

		if date == nil {
			fmt.Println("regDate.FindAllStringSubmatch(data[1], 1) return is nil")
			continue
		} else {
			date[0][1] = strings.Replace(date[0][1], "-", "", 2)

		}
		//获取时间信息   end

		//********************************************************************************************************

		//获取标题信息。 start
		regTitle := regexp.MustCompile(`">(?s:(.*?))</a>`)
		if regTitle == nil {
			fmt.Println("regexp.MustCompile(`\">(?s:(.*?))</a>`) return is nil")
			continue
		}
		title := regTitle.FindAllStringSubmatch(data[0], -1)
		if title == nil {
			fmt.Println("regTitle.FindAllStringSubmatch(data[0], -1) return is nil")
			continue
		} else {
			title[0][1] = strings.Replace(title[0][1], "\t", "", -1)
			title[0][1] = strings.Replace(title[0][1], "<br />", "", -1)
			title[0][1] = strings.Replace(title[0][1], "\n", "", -1)
			title[0][1] = strings.Replace(title[0][1], "&nbsp;", "", -1)
			fmt.Println(title[0][1])

		}
		//获取标题信息。 end

		//********************************************************************************************************

		//获取url信息。 start
		regUrl := regexp.MustCompile(`href=".(.*)target`)
		if regUrl == nil {
			fmt.Println("regexp.MustCompile(`href=\".(.*)target`) return is nil")
			continue
		}
		url := regUrl.FindAllStringSubmatch(data[1], -1)
		if url == nil {
			fmt.Println("regUrl.FindAllStringSubmatch(data[1], -1) return nil")
			continue
		} else {
			url[0][1] = strings.Replace(url[0][1], "\"", "", -1)
			fmt.Println(url[0][1])
		}

		//获取url信息。 end
		//**********************************************************************************************************

		//整合信息 start
		temp.time = string(date[0][1])
		temp.title = string(title[0][1])
		currentUrl := confUrl
		currentUrl += url[0][1]
		temp.url = currentUrl
		//整合信息 end
		fmt.Println(currentUrl)
		listInfo[index] = temp
		fmt.Printf("index=%d  %s\n", index, date[0][1])
		fmt.Println("************************************")
	}
	//fmt.Printf("完整列表信息为：", listInfo)
	return listInfo
}

func GetJbBody(url string) (string, error) {
	//下载网页。
	//url := "http://75.16.16.30/hebtax/swjb/201807/t20180724_1814760.html"
	pageStr, err1 := HttpGet(url)
	if err1 != nil {
		fmt.Println("下载网页。 err=", err1)
		return "", err1
	}
	var rst string
	//fmt.Println(pageStr)

	//过滤出正文	start
	regBody := regexp.MustCompile(`<meta name="ContentStart">(?s:(.*?))<meta name="ContentEnd">`)
	if regBody == nil {
		fmt.Println("过滤出正文	start return nil")
		return "", err1
	}
	body := regBody.FindStringSubmatch(pageStr)
	if body == nil {
		fmt.Println("过滤出正文 error")
		return "", err1
	}
	body[1] = strings.Replace(body[1], "\t", "", -1)
	body[1] = strings.Replace(body[1], "<br />", "", -1)
	body[1] = strings.Replace(body[1], "\n", "", -1)
	body[1] = strings.Replace(body[1], "\n\r", "", -1)
	body[1] = strings.Replace(body[1], "&nbsp;", "", -1)
	//body[1]是过滤结果
	//过滤出正文	end

	//*******************************************************************************************************

	//对正文再次进行过滤格式化   start

	regBody2 := regexp.MustCompile(`[0-9]*(pt")*>(?s:(.*?))<`)
	if regBody2 == nil {
		fmt.Println("对正文再次进行过滤格式化 return nil")
		return "", err1
	}
	body2 := regBody2.FindAllStringSubmatch(body[1], -1)

	if body2 == nil {
		fmt.Println("regBody2.FindAllStringSubmatch(body[1],-1) is nil err")
		return "", err1
	}
	for _, data := range body2 {
		//fmt.Println(data[2])
		//fmt.Println(data[0][:2])
		font, err6 := strconv.Atoi(data[0][:2])
		if err6 != nil {
			rst += data[2]
		} else {
			if font > 16 {
				data[2] += "\r\n"
				rst += data[2]
			} else if font == 16 {
				rst += data[2]
			}

		}

		//fmt.Println("_____________________________________________________________")
	}
	//对正文再次进行过滤格式化   end
	return rst, err1
}

func StoreTOFile(jbInfo jbListInfo, rst string) {
	var path string
	path += storePath
	path += "/"
	path += jbInfo.title
	path += ".txt"
	file, err1 := os.Create(path)
	defer file.Close()
	if err1 != nil {
		fmt.Println("创建文件失败", path)
		return
	}
	_, err2 := file.WriteString(rst)
	if err2 != nil {
		fmt.Println("写入内容失败。path=", path)
		return
	}
}
func GetlastTime(myconf *conf.Config) {
	path := myconf.Read("pathInfo", "oldPath")
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		fmt.Println("缺少lasttime.txt 请添加并修改配置文件。，err", err)
		return
	}
	buf := make([]byte, 1*1024)
	n, err2 := file.Read(buf)
	if err2 != nil {
		fmt.Println("", err2)
	}
	var err5 error
	lastTime, err5 = strconv.Atoi(string(buf[:n]))
	if err5 != nil {
		fmt.Println("lasttime.txt文件日期应为  式例“20001030”")
		return
	}
	fmt.Println(lastTime)
}

func SetlastTime(lasttime string) {

	file, err1 := os.OpenFile(lastTimePath, os.O_WRONLY|os.O_TRUNC, 0600)
	defer file.Close()
	if err1 != nil {
		fmt.Println("SetlastTime err=", err1)
		return
	}

	//	err := file.
	//	if err != nil {
	//		fmt.Println("file.Truncate(0)")
	//		return
	//	}
	fmt.Println("ddddddddddddddddddddddddddddddddddddddddddddd", lasttime)
	file.WriteString(lasttime)
}
