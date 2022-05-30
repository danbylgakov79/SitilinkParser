package main

import (
	"database/sql"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/geziyor/geziyor"
	"github.com/geziyor/geziyor/client"
	"github.com/geziyor/geziyor/export"
	_ "github.com/lib/pq"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var database *sql.DB
var err error
var waitGroup sync.WaitGroup

func CitilinkParse() {

	connStr := "user=postgres password=root dbname=electronic sslmode=disable"
	db, error := sql.Open("postgres", connStr)
	if error != nil {
		panic(error)
	}
	defer db.Close()
	database = db
	runtime.GOMAXPROCS(6)
	t := time.Now()
	waitGroup.Add(2)
	startParse()
	waitGroup.Wait()
	fmt.Printf("time after:%.2f", time.Since(t).Seconds())
}
func startParse() {
	clearTable("planshety_sitilink")
	clearTable("laptops_sitilinc")
	go startParseLaptops()
	go startParsePlanshety()

}
func startParseLaptops() {
	defer waitGroup.Done()
	for i := 1; i <= 35; i++ {
		geziyor.NewGeziyor(&geziyor.Options{
			StartURLs: []string{"https://www.citilink.ru/catalog/noutbuki/?p=" + strconv.Itoa(i)},
			ParseFunc: parseLaptops,
			Exporters: []export.Exporter{&export.JSON{}},
		}).Start()
	}

}

func startParsePlanshety() {
	defer waitGroup.Done()
	for i := 1; i <= 6; i++ {
		geziyor.NewGeziyor(&geziyor.Options{
			StartURLs: []string{"https://www.citilink.ru/catalog/planshety/?view_type=grid&f=discount.any%2Crating.any&p=" + strconv.Itoa(i)},
			ParseFunc: parsePlanshety,
			Exporters: []export.Exporter{&export.JSON{}},
		}).Start()
	}
}
func parseLaptops(g *geziyor.Geziyor, r *client.Response) {
	r.HTMLDoc.Find("div.ProductCardVerticalLayout.ProductCardVertical__layout").Each(func(i int, s *goquery.Selection) {

		var title = strings.TrimSpace(s.Find("div.ProductCardVerticalLayout__header a.ProductCardVertical__name.Link.js--Link.Link_type_default").Text())
		var price = strings.TrimSpace(s.Find("div.ProductCardVerticalLayout__footer span.ProductCardVerticalPrice__price-current_current-price.js--ProductCardVerticalPrice__price-current_current-price ").Text())
		var href string = "https://www.citilink.ru/"
		if ref, ok := s.Find("div.ProductCardVerticalLayout__header a.ProductCardVertical__name.Link.js--Link.Link_type_default").Attr("href"); ok {
			href = href + ref
		}
		price = replace(price)

		_, err = database.Exec("insert into electronic.laptops.laptops_sitilinc (href, price, title) values ($1,$2,$3)",
			href, price, title)
		if err != nil {
			fmt.Println(err)
		}
	})
}
func replace(s string) (prefix string) {
	var strs = strings.SplitAfterN(s, "\n", 2)

	if len(strs) == 1 {
		return ""
	}
	strs[0] = strings.ReplaceAll(strs[0], "\n", "")
	strs[0] = strings.ReplaceAll(strs[0], " ", "")
	return strs[0]
}
func parsePlanshety(g *geziyor.Geziyor, r *client.Response) {
	r.HTMLDoc.Find("div.ProductCardVerticalLayout.ProductCardVertical__layout").Each(func(i int, s *goquery.Selection) {

		var title = strings.TrimSpace(s.Find("div.ProductCardVerticalLayout__header a.ProductCardVertical__name.Link.js--Link.Link_type_default").Text())
		var price = strings.TrimSpace(s.Find("div.ProductCardVerticalLayout__footer span.ProductCardVerticalPrice__price-current_current-price.js--ProductCardVerticalPrice__price-current_current-price ").Text())
		var href string = "https://www.citilink.ru/"
		if ref, ok := s.Find("div.ProductCardVerticalLayout__header a.ProductCardVertical__name.Link.js--Link.Link_type_default").Attr("href"); ok {
			href = href + ref
		}
		price = replace(price)
		_, err = database.Exec("insert into electronic.laptops.planshety_sitilink (title, price, href) values ($1,$2,$3)",
			title, price, href)
		if err != nil {
			fmt.Println(err)
		}

	})
}

func clearTable(s string) {
	var queryString string = "delete from electronic.laptops." + s
	_, error := database.Exec(queryString)
	if error != nil {
		panic(error)
	}
}
