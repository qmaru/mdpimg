package mdpimg

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
)

type MdprWrapper struct{}

func (m *MdprWrapper) webHeader() http.Header {
	headers := make(http.Header)
	headers.Add("User-Agent", "Mozilla/5.0 (Linux; Android 7.1.1; E6533 Build/32.4.A.1.54; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/79.0.3945.136 Mobile Safari/537.36")
	headers.Add("X-Requested-With", "jp.mdpr.mdprviewer")
	return headers
}

func (m *MdprWrapper) apiHeader() http.Header {
	headers := make(http.Header)
	headers.Add("model", "E653325 (7.1.1)")
	headers.Add("mdpr-api", "3.0.0")
	headers.Add("mdpr-app", "android:37")
	headers.Add("User-Agent", "okhttp/4.2.2")
	return headers
}

func (m *MdprWrapper) httpGet(url string, header http.Header) []byte {
	httpClient := http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Panic(err)
	}
	req.Header = header
	res, err := httpClient.Do(req)
	if err != nil {
		log.Panic(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Panic(err)
	}
	defer res.Body.Close()
	return body
}

func (m *MdprWrapper) getArticleID(url string) (aid string) {
	urlPart := strings.Split(url, "/")
	aid = urlPart[len(urlPart)-1]
	return
}

func (m *MdprWrapper) URLCheck(url string) (string, bool) {
	newurl := strings.TrimSpace(url)
	if strings.Contains(newurl, "https://mdpr.jp/") {
		if strings.Contains(newurl, "photo/detail") {
			return "", false
		}
		return newurl, true
	}
	return "", false
}

func (m *MdprWrapper) getAPIURL(mdprURL string) (apiURL string) {
	aid := m.getArticleID(mdprURL)
	apiPrefix := "https://app2-mdpr.freetls.fastly.net"
	mobilePrefix := "https://app2-mdpr.freetls.fastly.net/articles/detail/"

	mobileIndex := mobilePrefix + aid
	resBody := m.httpGet(mobileIndex, m.webHeader())

	doc, _ := htmlquery.Parse(strings.NewReader(string(resBody)))
	nodes := htmlquery.Find(doc, `//div[@class="p-articleBody"]/a`)
	if len(nodes) != 0 {
		node := nodes[0]
		var apiJSON map[string]interface{}
		apiData := htmlquery.SelectAttr(node, "data-mdprapp-option")
		apiJSONString, _ := url.QueryUnescape(apiData)
		json.Unmarshal([]byte(apiJSONString), &apiJSON)
		apiURL = apiPrefix + apiJSON["url"].(string)
	} else {
		apiURL = ""
	}
	return
}

func GetImgs(mdprURL string) (imgs []string) {
	m := new(MdprWrapper)
	apiURL := m.getAPIURL(mdprURL)
	if apiURL != "" {
		resBody := m.httpGet(apiURL, m.apiHeader())
		var apiJSON map[string]interface{}
		json.Unmarshal(resBody, &apiJSON)
		mImgList := apiJSON["list"].([]interface{})
		for i := 0; i < len(mImgList); i++ {
			mData := mImgList[i].(map[string]interface{})
			imgs = append(imgs, mData["url"].(string))
		}
	}
	return
}
