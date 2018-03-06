# Dataflow kit

![alt tag](https://raw.githubusercontent.com/slotix/dataflowkit/master/images/dfk-logo/logo-mini.png)

[![Build Status](https://travis-ci.org/slotix/dataflowkit.svg?branch=master)](https://travis-ci.org/slotix/dataflowkit)
[![GoDoc](https://godoc.org/github.com/slotix/dataflowkit?status.svg)](https://godoc.org/github.com/slotix/dataflowkit)
[![Go Report Card](https://goreportcard.com/badge/github.com/slotix/dataflowkit)](https://goreportcard.com/report/github.com/slotix/dataflowkit)
[![codecov](https://codecov.io/gh/slotix/dataflowkit/branch/master/graph/badge.svg)](https://codecov.io/gh/slotix/dataflowkit)


Dataflow kit extracts structured data from web pages, following the specified extractors.

It can be used in many ways for data mining, data processing or archiving.

The actual use case can be grabbing list of products on several pages and follow each product’s details page to retrieve additional information. Parse endpoint returns information as a JSON, XML or CSV data.

DFK consists of two general services for fetching and parsing web pages content.

## Fetch service
**fetch.d** is the daemon that downloads html pages. It sends requests to [Splash server](https://github.com/scrapinghub/splash). Splash is a javascript rendering service. It is used to retrieve actual data before sending it to parse.d daemon. 

## Parse service
**parse.d** is the daemon that extracts data from downloaded web page following the rules described in configuration JSON file. Extracted data are returned in CSV, JSON or XML format.

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
1. Start Splash docker container 

``` docker run -d -it --rm -p 5023:5023 -p 8050:8050 -p 8051:8051 scrapinghub/splash```

[Splash](https://github.com/scrapinghub/splash) is used for fetching web pages to feed a Dataflow kit parser. 

2. Build and run fetch.d service
```
cd $GOPATH/src/github.com/slotix/dataflowkit/fetch/fetch.d && go build && ./fetch.d
```
3. In new terminal window build and run parse.d service
```
cd $GOPATH/src/github.com/slotix/dataflowkit/parse/parse.d && go build && ./parse.d
```
4. Launch parsing by sending POST request to parse daemon. Some json configuration files for testing are available in /examples folder.
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
			"paginateResults":false
		   }
```  
Read more information about scraper configuration JSON files at our [GoDoc reference](https://godoc.org/github.com/slotix/dataflowkit/parse/parse.d)

Extractors and filters are described at  [https://godoc.org/github.com/slotix/dataflowkit/extract](https://godoc.org/github.com/slotix/dataflowkit/extract)

## Front-End
Try http://scrape.dataflowkit.org Front-end with Point-and-click interface to Dataflow kit services. 

![alt tag](https://raw.githubusercontent.com/slotix/dataflowkit/master/images/dfk-screenshot2.png)

![alt tag](https://raw.githubusercontent.com/slotix/dataflowkit/master/images/dfk-screenshot1.png)


## License
This is Free Software, released under the BSD 3-Clause License.

## Contributing
You are welcome to contribute to our project. 
- Please submit [your issues](https://github.com/slotix/dataflowkit/issues) 
- Fork the [project](https://github.com/slotix/dataflowkit)

![alt tag](https://raw.githubusercontent.com/slotix/dataflowkit/master/images/spider/Spider-White-BG.png)
