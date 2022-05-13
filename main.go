package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

type HotelPrice struct{
	EndDate string
	Name string
	Price int
}
const dateFormat string = "02.01.2006"

func main() {
	today := time.Now().Local()
	startDate := flag.String("sd", today.Format(dateFormat), "Start date: must be in 05.06.2022 format")
	crawlDays := flag.Int("days", 3, "Number of days to crawl from start date, default to 3")
	flag.Parse()
	startDateTimeObject, err := time.Parse(dateFormat, *startDate)
	if err != nil {
		panic(err)
	}

	fName := "dump.json"
	file, err := os.Create(fName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	c := colly.NewCollector(
		// Visit only domains: hackerspaces.org, wiki.hackerspaces.org
		//colly.AllowedDomains("legoland.com", "coursera.org", "hackerspaces.org"),
	)

	dateGap := 0

	sd := startDateTimeObject
	priceCollection := make([]HotelPrice, 0, 200)
	for dateGap < *crawlDays {
		fmt.Println("START DATE IS: " + sd.String())
		ed := sd.AddDate(0,0,1)
		sd_f := sd.Format(dateFormat)
		ed_f := ed.Format(dateFormat)
		root_url := "https://vacations.legoland.com/california/LLCROOO/package/Results/main/01.html?mgnlPreview=true&=&trav=%5B%7B%22numberOfAdults%22%3A2%2C%22childAges%22%3A%5B7%5D%7D%5D&toDate=" + ed_f + "&upsell=false&stars=0&dis=50&fromDate=" + sd_f + "&duration=1&scc=ROOO&vi=f67d1e01-7a90-4ab0-bfb2-ddd97da2d435&packageCode=LLCROOO&minPrice=0&extra=false&maxPrice=5000&sortBy=random"

		c.OnHTML(`b[class=packageList_price--nowrap]`, func(e *colly.HTMLElement) {
			hotel := strings.TrimSpace(e.DOM.Parents().Eq(2).Find("a[href]").Eq(0).Text())
			price := strings.TrimSpace(e.Text)
			priceInt, _ := strconv.Atoi(strings.Split(price, "$")[1])
			fmt.Printf("Date from %s to %s, hotel is %s, price is %s\n", sd_f, ed_f, hotel, price)
			hotelPrice := HotelPrice{
				EndDate: ed_f,
				Name: hotel,
				Price: priceInt,
			}
			priceCollection = append(priceCollection, hotelPrice)
		})
	
		c.OnRequest(func(r *colly.Request) {
			//fmt.Println("Visiting", r.URL.String())
		})
	
		c.Visit(root_url)

		sd = sd.AddDate(0,0,1)
		dateGap++
		fmt.Printf("END START DATE IS: %v\n", dateGap)
	}
	sort.Slice(priceCollection, func(i, j int) bool { return priceCollection[i].Price < priceCollection[j].Price })
	fmt.Println(priceCollection)
	enc := json.NewEncoder(file)
	enc.SetIndent(""," ")
	enc.Encode(priceCollection)
}