// Dataflow kit - main
//
// Copyright © 2017-2018 Slotix s.r.o. <dm@slotix.sk>
//
//
// All rights reserved. Use of this source code is governed
// by the BSD 3-Clause License license.

// Fetcher CLI of the Dataflow kit downloads html content from web pages via Fetcher service endpoint.
//
// Currently two types of fetcher are available : Headless Chrome Fetcher and Base Fetcher.
//
// Base fetcher is used for downloading html web page using Go standard library's http.
//
// Chrome Fetcher connects to Headless Chrome which renders JavaScript pages.
//
// Accessing Fetcher endpoints
//
// Examples
//		./fetch.cli --URL http://example.com
//		./fetch.cli --URL http://example.com --FETCHER_TYPE base
//		./fetch.cli -u http://example.com -t base
//		./fetch.cli -u http://example.com -t chrome
//
// Flags and configuration settings
//		DFK_FETCH: HTTP listen address of Fetch service (defaults to "127.0.0.1:8000")
//		FETCHER_TYPE: DFK Fetcher type: chrome, base (defaults to base)
//		URL: URL to be fetched
// Request parameters for fetchers
//		FORMDATA: string value for passing formdata parameters.
//		For example the following string for processing pages which
//		require authentication may be passed:
//		"auth_key=880ea6a14ea49e853634fbdc5015a024&referer=http%3A%2F%2Fexample.com%2F&ips_username=user&ips_password=userpassword&rememberMe=1"
//		COOKIES: Cookies contain cookies to be added to request  before
//		sending it to browser.
//
package main

// EOF
