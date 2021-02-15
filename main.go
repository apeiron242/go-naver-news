package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/labstack/echo/v4"
)

var url1 string = "https://search.naver.com/search.naver?&where=news&query="
var url2 string = "&sm=tab_pge&sort=0&photo=0&field=0&reporter_article=&pd=0&ds=&de=&docid=&nso=so:r,p:all,a:all&mynews=0&cluster_rank=17&"
var url3 string = "&refresh_start=20"

type news struct {
	title string
	url   string
	data  string
}

func main() {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.File("index.html")
	})

	e.POST("/search", handlePost)

	PORT := os.Getenv("PORT")
	// PORT := ":8000"

	e.Logger.Fatal(e.Start(PORT))
}

func handlePost(c echo.Context) error {
	term := c.FormValue("term")
	scrape(term)
	defer os.Remove(term + ".csv")
	return c.Attachment("./"+term+".csv", term+".csv")
}

func scrape(term string) {
	c := make(chan []news)
	var newsData []news

	for i := 1; i < 711; i += 10 {
		go handleScrape(term, c, i)
	}

	for i := 1; i < 70; i++ {
		datas := <-c
		newsData = append(newsData, datas...)
	}

	writeFile(newsData, term)
}

func handleScrape(term string, c chan []news, i int) {
	var newsData []news

	fmt.Println("Scraping", i)
	res, err := http.Get(url1 + term + url2 + "start=" + strconv.Itoa(i) + url3)
	checkErr(err)
	checkStatus(res.StatusCode)

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".bx").Each(func(i int, s *goquery.Selection) {
		title := cleanStrings(s.Find(".news_tit").Text())
		url, _ := s.Find(".news_tit").Attr("href")
		data := cleanStrings(s.Find(".info").Text())
		newsData = append(newsData, news{title: title, url: url, data: data})
	})
	c <- newsData
}

func writeFile(data []news, term string) {
	file, err := os.Create(term + ".csv")
	checkErr(err)
	w := csv.NewWriter(file)

	defer w.Flush()

	headers := []string{"Title", "URL", "Data"}

	Err := w.Write(headers)
	checkErr(Err)

	for _, content := range data {
		if len(content.title) <= 0 {
			continue
		}
		contentSlice := []string{content.title, content.url, content.data}
		Err2 := w.Write(contentSlice)
		checkErr(Err2)
	}
}

func cleanStrings(str string) string {
	result := strings.TrimSpace(str)
	return result
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkStatus(code int) {
	if code != 200 {
		log.Fatalln("Status is not 200")
	}
}
