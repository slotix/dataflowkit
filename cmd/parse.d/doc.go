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
   "fetcherType": "splash",
   "paginateResults":false
  }'


Name

Collection name

Request

Request parameters are passed to Fetch Endpoint for downloading html pages.
URL holds the URL address of the web page to be downloaded. URL is required. All other fields including Params, Cookies, Func are optional.

Params is a string value for passing formdata parameters.
For example it may be used for processing pages which require authentication.
	"auth_key=880ea6a14ea49e853634fbdc5015a024&referer=http%3A%2F%2Fexample.com%2F&ips_username=user&ips_password=userpassword&rememberMe=1"

Cookies contain cookies to be added to request  before sending it to browser.
It may be used for processing pages after initial authentication. In the first step formdata with auth info is passed to a web page.
Response object headers may contain an Object like
	name: "Set-Cookie"
	value: "session_id=29d7b97879209ca89316181ed14eb01f; path=/; httponly"
These cookies should be passed to the next pages on the same domain.
	"session_id", "29d7b97879209ca89316181ed14eb01f", "/", domain="example.com"

Fields

A set of fields used to extract data from a web page.
A Field represents a given chunk of data to be extracted from every block in each page of a scrape.

Field name is required, and will be used to aggregate results.

Selector represents a CSS selector within the given block to process.  Pass in "." to use the root block's selector.

Extractor contains the logic on how to extract some results from the selector that is provided to this Field.

Paginator

Paginator is used to scrape multiple pages.
If there is no paginator in Payload, then no pagination is performed and it is assumed that the initial URL is the only page.
Paginator extracts the next page from a document by querying a given CSS selector and extracting the given HTML attribute from the resulting element.

Selector represents corresponding CSS selector for the next page along with
Attr defining HTML attribute for the next page.
MaxPages sets upper bound to maximum number of pages to scrape. The scrape will proceed until either this number of pages have been scraped, or until the paginator returns no further URLs.
Default value is 1.
Set maxPages value to 0 to indicate an unlimited number of pages to be scraped.

Format

The following Output formats are available: CSV, JSON, XML

fetcherType

fetcherType represents fetcher which is used for document download. 
Set it to either "splash" or "base" value. 
If omited in Payload, default fetcher type is defined as FETCHER_TYPE variable of parse.d service. 
fetcherType from Payload structure takes precedence over FETCHER_TYPE flag value.


paginateResults

Paginated results are returned if paginateResults is true.
Single list of combined results from every block on all pages is returned by default.
Paginated results are applicable for JSON and XML output formats.
Combined list of results is always returned for CSV format.
*/
//
// Flags and configuration settings
//
//General settings
//    DFK_PARSE: HTTP listen address of Parse service (defaults to "127.0.0.1:8001")
//
//    DFK_FETCH: HTTP listen address of Fetch service (defaults to "127.0.0.1:8000")
//
//    FETCHER_TYPE: represent fetcher which is used for document download.
//    Set up it to either `base` or `splash` values
//    fetcherType from Payload structure takes precedence over FETCHER_TYPE flag value.
//
//Storage settings
//    SKIP_STORAGE_MW: If true no data will be saved to storage. 
//    This flag forces fetcher to bypass storage middleware.
//
//    STORAGE_TYPE: Storage backend for intermediary data passed to Dataflow 
//    kit Parse service. 
//    Types: S3, Digital Ocean Spaces, Redis, Diskv 
//    (defaults to "Diskv"). It is case insensitive.
//
//    ITEM_EXPIRE_IN: Default value for item expiration in seconds (defaults to 3600)
//
//    DISKV_BASE_DIR: diskv base directory for storing parsed results (defaults to "diskv").
//    Find more information about Diskv storage at https://github.com/peterbourgon/diskv
//
//    SPACES_ENDPOINT: Digital Ocean Spaces Endpoint Address.
//    Find more information about DO Spaces at https://www.digitalocean.com/community/tutorials/an-introduction-to-digitalocean-spaces
//
//    SPACES_CONFIG: Digital Ocean Spaces Configuration file location.
//    (defaults to "~/.spaces/credentials")
//
//    S3_REGION: AWS S3 or Digital Ocean Spaces region (defaults to "us-east-1")
//
//    DFK_BUCKET: Amazon AWS S3 or Digital Ocean Spaces bucket name for storing 
//    parsed results. (defaults to "dfk-storage")
//
//    REDIS: Redis host address (defaults to "127.0.0.1:6379")
//
//    REDIS_EXPIRE: Default Redis expire value in seconds  (defaults to 3600)
//
//    REDIS_NETWORK: Redis Network (defaults to "tcp")
//
//    REDIS_PASSWORD: Redis Password (defaults to "")
//
//    REDIS_DB: Redis database (defaults to 0)
//
//    REDIS_SOCKET_PATH: Redis Socket Path (defaults to "")
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
