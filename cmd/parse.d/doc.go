// Dataflow kit - main
//
// Copyright © 2017-2018 Slotix s.r.o. <dm@slotix.sk>
//
//
// All rights reserved. Use of this source code is governed
// by the BSD 3-Clause License license.

/*
Parse service of the Dataflow kit parses html content from web pages following the rules described in configuration JSON file.

Here is a simple example for requesting Parse endpoint:
  curl -XPOST  127.0.0.1:8001/parse -d '
  {
   "name":"collection",
   "request":{
      "userToken":"7383ef32-525c-55d9-ac12-1bac76e0c802",
      "url":"https://example.com",
      "type": "chrome"
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
      "type":"next",
      "selector":".next",
      "attr":"href"
   },
   "format":"json"
  }'

Name

Collection name

Request

Request parameters are passed to Fetch Endpoint for downloading html pages.
url holds the URL address of the web page to be downloaded. URL is required. All other fields are optional.
userToken identifies every unique user making requests. Cookies are stored as key/value for each unique user to handle multiple requests to a domain.
type specifies fetcher type which may be "base" or "chrome" value.

Fields

A set of fields used to extract data from a web page.
A Field represents a given chunk of data to be extracted from every block on each page.

Field name is required, and is used to aggregate results.

Selector represents a CSS selector for data extraction within the given block. Pass in "." to use the root block's selector.

Extractor contains the logic on how to extract some results from the selector that is provided to this Field.

Paginator

Paginator is used to scrape multiple pages.
If there is no paginator in Payload, then no pagination is performed and it is assumed that the initial URL is the only page.
Paginator extracts the next page from a document by querying a given CSS selector and extracting the given HTML attribute from the resulting element.
There are three paginator types.
"Next link" paginator type is used on pages containing Next Button Paginator link.
"Infinite scroll" automatically loads content while user scrolls page down.
"Load more Button" looks like "Next link" but loads content on its click.

Type represents paginator type. The following are available: "next", "more", "infinite"
Selector represents corresponding CSS selector for the "Next" link or "Load more" Button paginator types page along with
Attr belong exclusively to "Next" link paginator to define HTML element attribute for the next page.


Format

The following Output formats are available: CSV, MS Excel, JSON, XML
*/
//
// Flags and configuration settings
//
//General settings
//    DFK_PARSE: HTTP listen address of Parse service (defaults to "127.0.0.1:8001")
//
//    DFK_FETCH: HTTP listen address of Fetch service (defaults to "127.0.0.1:8000")
//
//Storage settings
//
//    STORAGE_TYPE: Storage backend for intermediary data passed to Dataflow
//    kit Parse service.
//    Types: Diskv, MongoDB
//    (defaults to "Diskv"). It is case insensitive.
//
//    ITEM_EXPIRE_IN: Default value for item expiration in seconds (defaults to 86400)
//
//    DISKV_BASE_DIR: diskv base directory for storing parsed results (defaults to "diskv").
//    RESULTS_DIR: Directory for storing results (defaults to "results").
//    Find more information about Diskv storage at https://github.com/peterbourgon/diskv
//    MONGO: MongoDB host address (defaults to "127.0.0.1")
//
//Crawler settings
//    MAX_PAGES: The maximum number of pages to scrape. The scrape will proceed
//    until either this number of pages have been scraped, or until the paginator
//    returns no further URLs. Set this value to 0 to indicate an unlimited number
//    of pages to be scraped.(defaults to 1)
//
//    FETCH_DELAY: FetchDelay should be used for a scraper to throttle the crawling
//    speed to avoid hitting the web servers too frequently.
//    FetchDelay specifies sleep time for multiple requests for the same domain.
//    It is equal to FetchDelay * random value between 500 and 1500 msec. (defaults to 500)
//
//    RANDOMIZE_FETCH_DELAY:  RandomizeFetchDelay setting decreases the chance of a
//    crawler being blocked. This way a random delay ranging from 0.5 * FetchDelay
//    to 1.5 * FetchDelay seconds is used between consecutive requests to the same
//    domain. If FetchDelay is zero this option has no effect. (defaults to true)
//
//    IGNORE_FETCH_DELAY: Ignores fetchDelay setting intended for debug purpose.
//    Please set it to false in Production
//
//Output settings
//    FORMAT: Format represents output format (CSV, JSON, XML)(defaults to "json")
//
//    PAGINATE_RESULTS: Paginated results are returned if true.
//    Single list of combined results from every block on all pages is returned by default.
//    Paginated results are applicable for JSON and XML output formats.
//    Combined list of results is always returned for CSV format. (defaults to false)
//
package main

// EOF
