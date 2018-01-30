// Dataflow kit - main
//
// Copyright © 2017-2018 Slotix s.r.o. <dm@slotix.sk>
//
//
// All rights reserved. Use of this source code is governed
// by the BSD 3-Clause License license.

// Fetcher CLI of the Dataflow kit downloads html content from web pages via Fetcher service endpoint.
//
// Currently two types of fetcher are available : Splash Fetcher and Base Fetcher.
//
// Base fetcher is used for downloading html web page using Go standard library's http.
//
// Splash Fetcher uses Scrapinghub splash to fetch URLs.
// Splash is a javascript rendering service.
// Read more at https://github.com/scrapinghub/splash
//
// Accessing Fetcher endpoints
//
// Examples
//		./fetch.cli --URL http://example.com
//		./fetch.cli --URL http://example.com --FETCHER_TYPE base
//		./fetch.cli -u http://example.com -t base
//
// Flags and configuration settings
//		DFK_FETCH: HTTP listen address of Fetch service (defaults to "127.0.0.1:8000")
//		FETCHER_TYPE: DFK Fetcher type: splash, base
//		URL: URL to be fetched
//
package main

// EOF
