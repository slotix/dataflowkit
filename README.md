# Dataflow kit

![alt tag](https://raw.githubusercontent.com/slotix/dataflowkit/master/images/dfk-logo/logo-mini.png)

[![Build Status](https://travis-ci.org/slotix/dataflowkit.svg?branch=master)](https://travis-ci.org/slotix/dataflowkit)
[![GoDoc](https://godoc.org/github.com/slotix/dataflowkit?status.svg)](https://godoc.org/github.com/slotix/dataflowkit)
[![Go Report Card](https://goreportcard.com/badge/github.com/slotix/dataflowkit)](https://goreportcard.com/report/github.com/slotix/dataflowkit)
[![codecov](https://codecov.io/gh/slotix/dataflowkit/branch/master/graph/badge.svg)](https://codecov.io/gh/slotix/dataflowkit)


Dataflow kit extracts structured data from web pages, following the specified extractors.

It can be used in many ways for data mining, data processing or archiving.

The actual use case can be grabbing list of products on several pages and follow each product’s details page to retrieve additional information. Parse endpoint returns information as a JSON, XML or CSV data.

Try http://scrape.dataflowkit.org Front-end with Point-and-click interface to Dataflow kit services.  

DFK consists of two general services for fetching and parsing web pages content.

## Fetch service
**fetch.d** is the daemon that downloads html pages. It sends requests to [Splash server](https://github.com/scrapinghub/splash). Splash is a javascript rendering service. It is used to retrieve actual data before sending it to parse.d daemon. 

## Parse service
**parse.d** is the daemon that extracts data from downloaded web page following the rules described in configuration JSON file. Extracted data are returned in CSV, JSON or XML format.


![alt tag](https://raw.githubusercontent.com/slotix/dataflowkit/master/images/spider/Spider-White-BG.png)
