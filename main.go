package main

import (
	"errors"
	"fmt"
	"helper"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/axgle/mahonia"
	"github.com/k0kubun/pp"
	"github.com/mozillazg/request"
)

var (
	username = "123"
	password = "123"
)

func main() {
	//开启HTTP服务器
	initServer()
}

//开启http服务器
func initServer() {
	http.HandleFunc("/post", getPostSource)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	pp.Print("server listening on :90...")
	err := http.ListenAndServe(":90", nil)
	if err != nil {
		pp.Print(err)
	}
	pp.Print("server started")
}

//获取源帖子
func getPostSource(resp http.ResponseWriter, req *http.Request) {
	sourceUrl := req.URL.Query()
	threadId, ok := sourceUrl["id"]
	if !ok {
		pp.Print("帖子获取失败...")
		return
	}
	pp.Print("帖子id:" + threadId[0])
	hipdaUrl := fmt.Sprintf("https://www.hi-pda.com/forum/viewthread.php?tid=%s", threadId[0])
	body, _, err := curl(hipdaUrl, nil)
	if err != nil {
		resp.Write([]byte("获取失败..."))
		return
	}
	if !isLogin(string(body)) {
		//登录!
		err := doLogin()
		pp.Print(err)
		if err != nil {
			resp.Write([]byte("登陆失败..."))
		}
	}
	//重组body
	reBody, err := rebuildBody(body)
	if err != nil {
		resp.Write([]byte("出错了."))
		return
	}
	resp.Write([]byte(reBody))
}

//判断是否登录 true 已登录 false未登录
func isLogin(body string) bool {
	login := strings.Contains(body, "您无权进行当前操作")
	if login {
		return false
	}

	return true
}

//没登陆 则登陆 保存cookie
func doLogin() error {
	loginPageUrl := "https://www.hi-pda.com/forum/logging.php?action=login"
	body, _, err := curl(loginPageUrl, nil)
	if err != nil {
		return err
	}
	dom, _ := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	hash, _ := dom.Find("[name='formhash']").Attr("value")
	if len(hash) < 1 {
		pp.Print("hash is:" + hash)
		return errors.New("hash获取失败")
	}
	pwd := helper.Md5(password)
	loginReq := map[string]string{
		"formhash":   hash,
		"referer":    "https://www.hi-pda.com/forum/",
		"loginfield": "username",
		"username":   username,
		"password":   pwd,
		"questionid": "0",
		"cookietime": "2592000",
	}

	doLoginUrl := "https://www.hi-pda.com/forum/logging.php?action=login&loginsubmit=yes&inajax=1"
	doResult, res, err := curl(doLoginUrl, loginReq)
	if strings.Contains(string(doResult), "欢迎您回来") {
		//登录成功 保存cookie
		var cookieString string
		for _, c := range res.Cookies() {
			cookieString += c.Name + ":" + c.Value + "\r\n"
		}
		pp.Print("登陆成功!")
		helper.FilePutContents("./cookie", cookieString, false)
	} else {
		pp.Print("登陆失败")
		return errors.New("登录失败。")
	}
	return err
}

//CURL请求
func curl(url string, reqData map[string]string) ([]byte, *request.Response, error) {
	c := new(http.Client)
	req := request.NewRequest(c)
	req.Headers = map[string]string{
		"Accept-Encoding": "deflate,br",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
		"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
		"Cache-Control":   "no-cache",
		"Connection":      "keep-alive",
		"Pragma":          "no-cache",
		"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36",
	}
	var resp *request.Response
	var err error
	//解析cookie
	if f, e := os.OpenFile("./cookie", os.O_RDWR, 0644); e == nil {
		cook, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, nil, err
		}
		cookie := parseCookie(string(cook))
		req.Cookies = cookie
	}
	//POST请求
	if reqData != nil {
		resp, err = req.PostForm(url, reqData)
	} else {
		resp, err = req.Get(url)
	}
	defer resp.Body.Close()

	resbody := mahonia.NewDecoder("gbk").NewReader(resp.Body)
	result, _ := ioutil.ReadAll(resbody)
	return result, resp, err
}

//从文本解析cookie
func parseCookie(str string) map[string]string {
	cookie := make(map[string]string)
	lines := strings.Split(str, "\n")
	for _, v := range lines {
		l := strings.Split(v, ":")
		if len(l) < 2 {
			continue
		}
		key := l[0]
		val := strings.Replace(l[1], "\r", "", -1)
		cookie[key] = val
	}
	return cookie
}

//重组body
func rebuildBody(html []byte) (string, error) {
	dom, _ := goquery.NewDocumentFromReader(strings.NewReader(string(html)))
	//标题部分
	title := dom.Find("#nav").Text()
	title = strings.Replace(title, "Hi!PDA", "", -1)
	title = strings.Replace(title, "Discovery ", "", -1)
	title = strings.Replace(title, "»", "", -1)
	title = strings.Trim(title, " ")
	//内容部分
	var body string
	dom.Find(".defaultpost").Each(func(i int, s *goquery.Selection) {
		h, _ := s.Html()
		body += h
	})
	//组装HTML
	header := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="description" content="">
    <title>%s</title>
	<link rel="stylesheet" type="text/css" href="https://img02.hi-pda.com/forum/forumdata/cache/style_1_common.css?nsP" />
	<link rel="stylesheet" type="text/css" href="https://img02.hi-pda.com/forum/forumdata/cache/scriptstyle_1_viewthread.css?nsP" />
    <link href="/static/custom.css" rel="stylesheet">
</head>
<body id="viewthread">
<div id="wrap" class="wrap s_clear threadfix">
	%s
</div>
</body>
</html>
`
	fullHtml := fmt.Sprintf(header, title, body)

	return fullHtml, nil
}
