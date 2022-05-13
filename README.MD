# Golang crawler using colly

- Get best prices from legoland!!!
- `./gocrawl` will crawl the prices for the next 3 days for 1 night stay.
- `./gocrawl -sd 01.08.2022 -days 15` will crawl from sd(startdate) to N number of days(-days)
- `./gocrawl -walk-months` will crawl from this month til 12/31 this year and return a monthly dump sorted by lowest price