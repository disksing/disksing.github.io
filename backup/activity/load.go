package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type RssItem struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Link        string `xml:"link"`
}

type RssRes struct {
	Channel struct {
		Items []RssItem `xml:"item"`
	} `xml:"channel"`
}

func loadRSS(url string) (*RssRes, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var rssRes RssRes
	err = xml.Unmarshal(data, &rssRes)
	if err != nil {
		return nil, err
	}
	return &rssRes, nil
}

type ActivityItem struct {
	Title    string
	Date     time.Time
	Link     string
	Content  string
	Tags     []string
	FileName string
}

func loadItems() (items []ActivityItem, err error) {
	weiboRSS, err := loadRSS("http://vpn.disksing.com:255/weibo/user/2381077925/1")
	if err != nil {
		return
	}
	for _, item := range weiboRSS.Channel.Items {
		ai := ActivityItem{
			Title:    weiboTitle(item.Title),
			Date:     convertDate(item.PubDate),
			Link:     item.Link,
			Content:  item.Description,
			Tags:     []string{"新浪微博"},
			FileName: "weibo",
		}
		items = append(items, ai)
	}
	blRSS, err := loadRSS("http://vpn.disksing.com:255/bilibili/user/dynamic/2207710")
	if err != nil {
		return
	}
	for _, item := range blRSS.Channel.Items {
		ai := ActivityItem{
			Title:    item.Title,
			Date:     convertDate(item.PubDate),
			Link:     item.Link,
			Content:  item.Description,
			Tags:     []string{"Bilibili"},
			FileName: "bilibili",
		}
		items = append(items, ai)
	}
	doubanRSS, err := loadRSS("http://vpn.disksing.com:255/douban/people/80983646/status")
	if err != nil {
		return
	}
	for _, item := range doubanRSS.Channel.Items {
		ai := ActivityItem{
			Title:    "豆瓣广播",
			Date:     convertDate(item.PubDate),
			Link:     item.Link,
			Content:  strings.TrimPrefix(item.Title, "闰土 "),
			Tags:     []string{"豆瓣"},
			FileName: "douban",
		}
		items = append(items, ai)
	}
	return
}

func main() {
	items, err := loadItems()
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, item := range items {
		fName := item.FileName + "_" + item.Date.Format("20060102150405") + ".md"
		var b bytes.Buffer
		b.WriteString("+++\n")
		b.WriteString(fmt.Sprintf(`title = "%s"`, item.Title) + "\n")
		b.WriteString(fmt.Sprintf(`date = "%s"`, item.Date.Format("2006-01-02 15:04:05")) + "\n")
		b.WriteString(fmt.Sprintf(`link = "%s"`, item.Link) + "\n")
		b.WriteString(fmt.Sprintf(`description = "%s"`, convertContent(item.Content)) + "\n")
		if len(item.Tags) > 0 {
			b.WriteString("tags = [\"" + item.Tags[0])
			for i := 1; i < len(item.Tags); i++ {
				b.WriteString("\", \"" + item.Tags[i])
			}
			b.WriteString("\"]\n")
		}
		b.WriteString("+++\n")
		ioutil.WriteFile(fName, b.Bytes(), 0644)
	}
}

func weiboTitle(title string) string {
	if strings.Contains(title, "//") || strings.Contains(title, "转发微博") {
		return "转发微博"
	}
	return "发布微博"
}

func convertDate(date string) time.Time {
	t, _ := time.Parse("Mon, 2 Jan 2006 15:04:05 MST", date)
	return t
}

func convertContent(content string) string {
	content = strings.ReplaceAll(content, `"`, `\"`)
	content = strings.ReplaceAll(content, "\n", "</br>")
	return content
}
