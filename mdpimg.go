package mdpimg

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
)

type MdprWrapper struct{}

func (m *MdprWrapper) webHeader() http.Header {
	headers := make(http.Header)
	headers.Add("User-Agent", "mdpr-user-agent: Mozilla/5.0 (Linux; Android 7.1.1; E6533 Build/32.4.A.1.54; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/94.0.4606.85 Mobile Safari/537.36")
	headers.Add("X-Requested-With", "jp.mdpr.mdprviewer")
	return headers
}

func (m *MdprWrapper) apiHeader() http.Header {
	headers := make(http.Header)
	headers.Add("mdpr-user-agent", "sony; E653325; android; 7.1.1; 3.10.4838(66);")
	headers.Add("User-Agent", "okhttp/4.9.1")
	return headers
}

func (m *MdprWrapper) fetch(url string, header http.Header) ([]byte, error) {
	httpClient := http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header = header
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return body, nil
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

func (m *MdprWrapper) GetImgURL(mdprURL string) (string, error) {
	aid := m.getArticleID(mdprURL)
	apiPrefix := "https://app2-mdpr.freetls.fastly.net"
	mobilePrefix := "https://app2-mdpr.freetls.fastly.net/articles/detail/"

	mobileIndex := mobilePrefix + aid
	resBody, err := m.fetch(mobileIndex, m.webHeader())
	if err != nil {
		return "", err
	}

	doc, err := htmlquery.Parse(strings.NewReader(string(resBody)))
	if err != nil {
		return "", err
	}
	nodes := htmlquery.Find(doc, `//div[@class="p-articleBody"]/a`)
	for _, node := range nodes {
		var appJson map[string]interface{}
		appData := htmlquery.SelectAttr(node, "data-mdprapp-option")
		appStr, err := url.QueryUnescape(appData)
		if err != nil {
			return "", err
		}
		err = json.Unmarshal([]byte(appStr), &appJson)
		if err != nil {
			return "", err
		}
		mdpURL := appJson["url"].(string)
		if strings.Contains(mdpURL, aid) {
			apiURL := apiPrefix + mdpURL
			return apiURL, nil
		}
	}
	return "", errors.New("url not found")
}

func (m *MdprWrapper) GetImgs(url string) ([]string, error) {
	resBody, err := m.fetch(url, m.apiHeader())
	if err != nil {
		return nil, err
	}
	var apiJSON map[string]interface{}
	err = json.Unmarshal(resBody, &apiJSON)
	if err != nil {
		return nil, err
	}
	imgList := apiJSON["list"].([]interface{})
	var imgs []string
	for i := 0; i < len(imgList); i++ {
		mData := imgList[i].(map[string]interface{})
		imgs = append(imgs, mData["url"].(string))
	}
	return imgs, nil
}

func Get(url string) ([]string, error) {
	m := new(MdprWrapper)
	url, ok := m.URLCheck(url)
	if ok {
		imgURL, err := m.GetImgURL(url)
		if err != nil {
			return nil, err
		}
		imgs, err := m.GetImgs(imgURL)
		if err != nil {
			return nil, err
		}
		return imgs, nil
	}
	return nil, errors.New("url is error")
}
