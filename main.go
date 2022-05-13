package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

type HotelPrice struct {
	EndDate string
	Name    string
	Price   int
}

const dateFormat string = "02.01.2006"

func main() {
	var wg sync.WaitGroup

	today := time.Now().Local()
	startDate := flag.String("sd", today.Format(dateFormat), "Start date: must be in 05.06.2022 format")
	walkMonths := flag.Bool("walk-months", false, "Instead of crawl number of days, it will walk by months")
	crawlDays := flag.Int("days", 3, "Number of days to crawl from start date, default to 3")
	flag.Parse()
	startDateTimeObject, err := time.Parse(dateFormat, *startDate)
	if err != nil {
		panic(err)
	}
	sd := startDateTimeObject

	if !*walkMonths {
		priceCollection := make([]HotelPrice, 0, 200)
		file := mkfile(startDateTimeObject.Month().String())
		defer file.Close()

		dateGap := 0
		for dateGap < *crawlDays {
			fmt.Println("START DATE IS: " + sd.String())
			ed := sd.AddDate(0, 0, 1)
			sd_f := sd.Format(dateFormat)
			ed_f := ed.Format(dateFormat)
			root_url := "https://vacations.legoland.com/california/LLCROOO/package/Results/main/01.html?mgnlPreview=true&=&trav=%5B%7B%22numberOfAdults%22%3A2%2C%22childAges%22%3A%5B7%5D%7D%5D&toDate=" + ed_f + "&upsell=false&stars=0&dis=50&fromDate=" + sd_f + "&duration=1&scc=ROOO&vi=f67d1e01-7a90-4ab0-bfb2-ddd97da2d435&packageCode=LLCROOO&minPrice=0&extra=false&maxPrice=5000&sortBy=random"
			priceResult := crawl(root_url, sd_f, ed_f)
			priceCollection = append(priceCollection, priceResult...)

			sd = sd.AddDate(0, 0, 1)
			dateGap++
			fmt.Printf("END START DATE IS: %v\n", dateGap)
		}

		encodeResult(priceCollection, file)
	} else {
		remainingMonths := getRemainingMonths(today)
		for _, v := range remainingMonths {
			tt := getDaysFromMonth(v, dateFormat)
			fmt.Printf("This is %v", tt)
			wg.Add(1)
			go walkMonth(tt, v, &wg)
		}
	}
	wg.Wait()
}

func getRemainingMonths(t time.Time)[]time.Time {
	remainMonths := make([]time.Time, 0, 12)
	y, _, _ := t.Date()
	nextY := time.Date(y+1, 1, 1, 1, 1, 1, 1, time.Local)
	nt := t
	
	for nt.Before(nextY) {
		remainMonths = append(remainMonths, nt)
		nt = nt.AddDate(0, 1, 0)
	}
	return remainMonths
}

//given a date, get the 1st day of the month and last day of the month, return a list of all the dates
func walkMonth(days []string, startDateTimeObject time.Time, wg *sync.WaitGroup) {
	defer wg.Done()
	priceCollectionMonths := make([]HotelPrice, 0, 200)
	file := mkfile(startDateTimeObject.Month().String())
	defer file.Close()
	for _, v := range days {
		ed, err := time.Parse(dateFormat, v)
		if err != nil {
			panic(err)
		}
		ed_f := ed.AddDate(0, 0, 1).Format(dateFormat)
		root_url := "https://vacations.legoland.com/california/LLCROOO/package/Results/main/01.html?mgnlPreview=true&=&trav=%5B%7B%22numberOfAdults%22%3A2%2C%22childAges%22%3A%5B7%5D%7D%5D&toDate=" + ed_f + "&upsell=false&stars=0&dis=50&fromDate=" + v + "&duration=1&scc=ROOO&vi=f67d1e01-7a90-4ab0-bfb2-ddd97da2d435&packageCode=LLCROOO&minPrice=0&extra=false&maxPrice=5000&sortBy=random"
		priceResult := crawl(root_url, v, ed_f)
		priceCollectionMonths = append(priceCollectionMonths, priceResult...)
	}
	encodeResult(priceCollectionMonths, file)
}

func encodeResult(priceCollection []HotelPrice, file *os.File) {
	sort.Slice(priceCollection, func(i, j int) bool { return priceCollection[i].Price < priceCollection[j].Price })
	fmt.Println(priceCollection)
	enc := json.NewEncoder(file)
	enc.SetIndent("", " ")
	enc.Encode(priceCollection)
}
func getDaysFromMonth(startDate time.Time, df string) []string {
	days := make([]string, 0, 31)
	startYear, startMonth, _ := startDate.Date()
	first := time.Date(startYear, startMonth, 1, 0, 0, 0, 0, time.Local)
	last := first.AddDate(0, 1, -1)
	days = append(days, first.Format(df))
	dd := first.AddDate(0, 0, 1).Format(df)
	ll := last.Format(df)
	for dd != ll {
		days = append(days, dd)
		dt, _ := time.Parse(df, dd)
		dd = dt.AddDate(0, 0, 1).Format(df)
	}
	return days
}

func mkfile(month string) *os.File {
	fmt.Println("inside")
	fn := fmt.Sprintf("dump-%v-*.json", month)
	fName, err := ioutil.TempFile(".", fn)
	if err != nil {
		panic(err)
	}
	return fName
}

func crawl(root_url, sd_f, ed_f string) []HotelPrice {
	c := colly.NewCollector(
	// Visit only domains: hackerspaces.org, wiki.hackerspaces.org
	//colly.AllowedDomains("legoland.com", "coursera.org", "hackerspaces.org"),
	)
	var hotelPrice []HotelPrice
	c.OnHTML(`b[class=packageList_price--nowrap]`, func(e *colly.HTMLElement) {
		hotel := strings.TrimSpace(e.DOM.Parents().Eq(2).Find("a[href]").Eq(0).Text())
		price := strings.TrimSpace(e.Text)
		priceInt, _ := strconv.Atoi(strings.Split(price, "$")[1])
		fmt.Printf("Date from %s to %s, hotel is %s, price is %s\n", sd_f, ed_f, hotel, price)
		oneHotelPrice := HotelPrice{
			EndDate: ed_f,
			Name:    hotel,
			Price:   priceInt,
		}
		hotelPrice = append(hotelPrice, oneHotelPrice)
	})

	c.OnRequest(func(r *colly.Request) {
		//fmt.Println("Visiting", r.URL.String())
	})

	c.Visit(root_url)
	return hotelPrice
}
