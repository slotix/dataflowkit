# Dataflow kit

![alt tag](https://raw.githubusercontent.com/slotix/dataflowkit/master/images/logo.png)

[![Build Status](https://travis-ci.org/slotix/dataflowkit.svg?branch=master)](https://travis-ci.org/slotix/dataflowkit)
[![GoDoc](https://godoc.org/github.com/slotix/dataflowkit?status.svg)](https://godoc.org/github.com/slotix/dataflowkit)
[![Go Report Card](https://goreportcard.com/badge/github.com/slotix/dataflowkit)](https://goreportcard.com/report/github.com/slotix/dataflowkit)
[![codecov](https://codecov.io/gh/slotix/dataflowkit/branch/master/graph/badge.svg)](https://codecov.io/gh/slotix/dataflowkit)


Dataflow kit ("DFK") is a Web Scraping framework for Gophers. It extracts data from web pages, following the specified CSS Selectors.

You can use it in many ways for data mining, data processing or archiving.

- Dataflow kit is fast. It takes about 4-6 seconds to fetch and then parse 50 pages.
- Dataflow kit is suitable to process quite large volumes of data. Our tests show the time needed to parse appr. 4 millions of pages is about 7 hours. 

## Dataflow kit benefits:

- Scraping of JavaScript generated pages;
- Data extraction from paginated websites;
- Processing infinite scrolled pages.
- Sсraping of websites behind login form;
- Cookies and sessions handling;
- Following links and detailed pages processing;
- Managing delays between requests per domain; 
- Following robots.txt directives; 
- Various storage types support including Diskv, Mongodb, Cassandra; 
Storage interface is flexible enough to add more storage types easily.
- Save results as CSV, MS Excel, JSON, XML;


DFK consists of two general services for fetching and parsing web pages content.

## Fetch service
**fetch.d** server is intended for html web pages content download. 
Depending on Fetcher type, web page content is downloaded using either Base Fetcher or Chrome fetcher. 

Base fetcher uses standard golang http client to fetch pages as is. 
It works faster than Chrome fetcher. But Base fetcher cannot render dynamic javascript driven web pages. 

Chrome fetcher is intended for rendering dynamic javascript based content. It sends requests to Chrome running in headless mode.  

Fetchers pass retrieved data to parse.d service. 

## Parse service
**parse.d** is the service that extracts data from downloaded web page following the rules listed in configuration JSON file. Extracted data is returned in CSV, MS Excel, JSON or XML format.

*Note: Sometimes Parse service cannot extract data from some pages retrieved by default Base fetcher. Empty results may be returned while parsing Java Script generated pages. Parse service then attempts to force Chrome fetcher to render the same dynamic javascript driven content automatically. Have a look at https://scrape.dataflowkit.org/persons/page-0 which is a sample of JavaScript driven web page.*   

## Installation
Using [dep](https://github.com/golang/dep)
```
dep ensure -add github.com/slotix/dataflowkit@master
```
or go get
```
go get -u github.com/slotix/dataflowkit
```

## Usage

### Docker
1. Install [Docker](https://www.docker.com) and [Docker Compose](https://docs.docker.com/compose/install/)

2. Start services.

```
cd $GOPATH/src/github.com/slotix/dataflowkit && docker-compose up
```
This command fetches docker images automatically and starts services.

3. Launch parsing in the second terminal window by sending POST request to parse daemon. Some json configuration files for testing are available in /examples folder.
```
curl -XPOST  127.0.0.1:8001/parse --data-binary "@$GOPATH/src/github.com/slotix/dataflowkit/examples/books.toscrape.com.json"
```
Here is the sample json configuration file:

```
{
	"name":"collection",
	"request":{
	   "url":"https://example.com"
	},
	"fields":[
	   {
		  "name":"Title",
		  "selector":".product-container a",
		  "extractor":{
			 "types":["text", "href"],
			 "filters":[
				"trim",
				"lowerCase"
			 ],
			 "params":{
				"includeIfEmpty":false
			 }
		  }
	   },
	   {
		  "name":"Image",
		  "selector":"#product-container img",
		  "extractor":{
			 "types":["alt","src","width","height"],
			 "filters":[
				"trim",
				"upperCase"
			 ]
		  }
	   },
	   {
		  "name":"Buyinfo",
		  "selector":".buy-info",
		  "extractor":{
			 "types":["text"],
			 "params":{
				"includeIfEmpty":false
			 }
		  }
	   }
	],
	"paginator":{
	   "selector":".next",
	   "attr":"href",
	   "maxPages":3
	},
	"format":"json",
	"fetcherType":"chrome",
	"paginateResults":false
}
```
Read more information about scraper configuration JSON files at our [GoDoc reference](https://godoc.org/github.com/slotix/dataflowkit/parse/parse.d)

Extractors and filters are described at  [https://godoc.org/github.com/slotix/dataflowkit/extract](https://godoc.org/github.com/slotix/dataflowkit/extract)

4. To stop services just press Ctrl+C and run 
``` 
cd $GOPATH/src/github.com/slotix/dataflowkit && docker-compose down --remove-orphans --volumes
```

[![IMAFGE ALT CLI Dataflow kit web scraping framework](https://raw.githubusercontent.com/slotix/dataflowkit/master/images/CLI-DFK.png)](https://youtu.be/lqFz1CbWzRs)

Click on image to see CLI in action.

### Manual way

1. Start Chrome docker container 
``` 
docker run --init -it --rm -d --name chrome --shm-size=1024m -p=127.0.0.1:9222:9222 --cap-add=SYS_ADMIN \
  yukinying/chrome-headless-browser:65.0.3322.3
```


[Headless Chrome](https://developers.google.com/web/updates/2017/04/headless-chrome) is used for fetching web pages to feed a Dataflow kit parser. 

2. Build and run fetch.d service
```
cd $GOPATH/src/github.com/slotix/dataflowkit/cmd/fetch.d && go build && ./fetch.d
```
3. In new terminal window build and run parse.d service
```
cd $GOPATH/src/github.com/slotix/dataflowkit/cmd/parse.d && go build && ./parse.d
```
4. Launch parsing. See step 3. from the previous section. 

### Run tests
- ```docker-compose -f test-docker-compose.yml up -d```
- ```./test.sh```
- To stop services just run ```docker-compose -f test-docker-compose.yml down```


## Front-End
Try https://dataflowkit.org/dfk Front-end with Point-and-click interface to Dataflow kit services. It generates JSON config file and sends POST request to DFK Parser 

[![IMAGE ALT Dataflow kit web scraping framework](https://raw.githubusercontent.com/slotix/dataflowkit/master/images/dfk-screenshot1.png)](https://youtu.be/SKBkclf1FxA)

Click on image to see Dataflow kit in action.

## License
This is Free Software, released under the BSD 3-Clause License.

## Contributing
You are welcome to contribute to our project. 
- Please submit [your issues](https://github.com/slotix/dataflowkit/issues) 
- Fork the [project](https://github.com/slotix/dataflowkit)


![alt tag](https://raw.githubusercontent.com/slotix/dataflowkit/master/images/Spider-White-BG.png)
