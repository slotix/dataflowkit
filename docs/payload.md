# Payload

## Paylod Example.

Here is a simple example of JSON formatted structure to be passed to Parse Endpoint.  

```json
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
            "type":"link",
            "params":{
               "includeIfEmpty":false
            }
         }
      },
      {
         "name":"Image",
         "selector":"#product-container img",
         "extractor":{
            "type":"image"
         }
      },
      {
         "name":"Buyinfo",
         "selector":".buy-info",
         "extractor":{
            "type":"text",
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



### Name.

Collection name.  

### Request.

Request parameters are passed to Fetch Endpoint for downloading html pages. 

URL field is required. All other fields including Params, Cookies, Func, SplashWait are optional.

#### url

url holds the URL address of the web page to be downloaded. 

#### params

Params is a string value for passing formdata parameters. 

For example it may be used for processing pages which require authentication

**Example:**

*"auth_key=880ea6a14ea49e853634fbdc5015a024&referer=http%3A%2F%2Fexample.com%2F&ips_username=user&ips_password=userpassword&rememberMe=1"*

#### cookies

Cookies contain cookies to be added to request  before sending it to browser. 

It may be used for processing pages after initial authentication. In the first step formdata with auth info is passed to a web page. 

Response object headers may contain an Object like

*name: "Set-Cookie"*

*value: "session_id=29d7b97879209ca89316181ed14eb01f; path=/; httponly"*

These cookie should be passed to the next pages on the same domain.

*"session_id", "29d7b97879209ca89316181ed14eb01f", "/", domain="example.com"*

#### func

Reserved

#### wait

Time in seconds to wait until java scripts loaded. Sometimes wait parameter should be set to more than default 0,5. It allows to finish js scripts execution on a web page. 

### Fields

A set of fields used to extract data from a web page.  

A Field represents a given chunk of data to be extracted from every block in each page of a scrape.

#### name

Field name is required, and will be used to aggregate results.

#### selector

A CSS selector within the given block to process.  Pass in "." to use the root block's selector.

#### extractor

Extractor contains the logic on how to extract some results from the selector that is provided to this Field. 

Have a look at [Extractor types](extractors.md) topic for more information.  

### Paginator

Paginator is used to scrape multiple pages. 

If Paginator is nil, then no pagination is performed and it is assumed that the initial URL is the only page.

Paginator extracts the next page from a document by querying a given CSS selector and extracting the given HTML attribute from the resulting element. 

#### selector

CSS selector for the next page

#### attr

HTML attribute for the next page

#### maxPages

The maximum number of pages to scrape. The scrape will proceed until either this number of pages have been scraped, or until the paginator returns no further URLs. 

Default value is 1. 

Set this value to 0 to indicate an unlimited number of pages to be scraped.

### format

The following Output formats are available: CSV, JSON, Microsoft Excel, XML

### paginateResults

Paginated results are returned if *paginateResults* is *true*. 

Single list of combined results from every block on all pages is returned by default.

Paginated results are applicable for JSON and XML output formats. 

Combined list of results is always returned for CSV format.  