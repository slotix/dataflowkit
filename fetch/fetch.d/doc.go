// Dataflow kit - main
//
// Copyright © 2017-2018 Slotix s.r.o. <dm@slotix.sk>
//
//
// All rights reserved. Use of this source code is governed
// by the BSD 3-Clause License license.

// Fetcher service of the Dataflow kit downloads html content from web pages to feed Dataflow kit scrapers.
//
// Currently two types of fetcher are available : Splash Fetcher and Base Fetcher.
//
// Base fetcher is used for downloading html web page using Go standard library's http.
//
// Splash Fetcher uses Scrapinghub splash to fetch URLs.
// Splash is a javascript rendering service.
// Read more at https://github.com/scrapinghub/splash
//
// Here are two Examples for accessing Fetcher endpoints:
//		curl -XPOST  localhost:8000/fetch/splash -d '{"url":"http://example.com"}'
//		curl -XPOST  localhost:8000/fetch/base -d '{"url":"http://example.com"}'
//
package main

// EOF
