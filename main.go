package main

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	//tm "github.com/buger/goterm":
	"github.com/eidolon/wordwrap"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/mmcdole/gofeed"
	"io"
	"log"
	"net/http"
	"strings"
)

type News struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Link        string `json:"link"`
}

func main() {

	headlines, news := getHeadLines("https://www.thehindu.com/feeder/default.rss")

	index, err := promptUI("HeadLines", 20, headlines)

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	content, err := parseHTML(news[index].Link)

	if err != nil {
		fmt.Println("Some shit went wrong", err)
	}
	outPutToTerminal(content)
}

func getHeadLines(url string) ([]string, []News) {

	var headlines []string
	news, err := getNews(url)

	if err != nil {
		panic("Couldn't parse RSS feed")
	}

	green := color.New(color.FgGreen).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()

	for _, val := range news {
		title, description := val.Title, val.Description
		description = description[0:min(len(description), 200)] + "..."
		titleString := fmt.Sprintf("%s (%s)", white(title), green(description))
		headlines = append(headlines, titleString)
	}

	return headlines, news
}

func promptUI(label string, size int, items []string) (int, error) {

	prompt := promptui.Select{
		Label: label,
		Items: items,
		Size:  size,
	}

	index, _, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return -1, err
	}

	return index, nil
}

func getNews(url string) ([]News, error) {

	var news []News

	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)

	if err != nil {
		log.Println(err, "Some shit went wrong")
		return nil, errors.New("Couldn't parse RSS feed")
	}

	for _, item := range feed.Items {
		var (
			title       = strings.TrimSpace(item.Title)
			description = strings.TrimSpace(item.Description)
			link        = strings.TrimSpace(item.Link)
		)
		newsItem := News{Title: title, Description: description, Link: link}
		news = append(news, newsItem)
	}

	return news, nil
}

func parseHTML(url string) (string, error) {
	body, err := getArticle(url)
	if err != nil {
		return "", err
	}
	html, err := getText(body)

	if err != nil {
		return "", err
	}
	return html, nil
}

func getArticle(url string) (io.Reader, error) {
	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func getText(html io.Reader) (string, error) {
	var e error
	body := ""

	doc, err := goquery.NewDocumentFromReader(html)

	if err != nil {
		return "", err
	}
	wrapper := wordwrap.Wrapper(180, false)
	doc.Find("div.article").Children().Each(func(i int, s *goquery.Selection) {

		idString, _ := s.Attr("id")

		if strings.Contains(idString, "content-body-") {
			s.Children().Each(func(j int, el *goquery.Selection) {
				body = body + wrapper(el.Text())
				body = body + "\n\n"

			})

			return
		}
	})

	return body, e
}

func outPutToTerminal(text string) {
	//	tm.Clear()
	d := color.New(color.FgHiYellow, color.Italic)
	padded := d.Sprintf("%-72v", text)
	fmt.Println(padded)
	main()
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
