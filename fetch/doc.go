// Dataflow kit - fetch
//
// Copyright © 2017-2018 Slotix s.r.o. <dm@slotix.sk>
//
//
// All rights reserved. Use of this source code is governed
// by the BSD 3-Clause License license.

// Package fetch of the Dataflow kit is used by fetch.d service which downloads html content from web pages to feed Dataflow kit scrapers.
//
// Fetcher is the interface that must be satisfied by things that can fetch
// remote URLs and return their contents.
//
// Currently two types of fetcher are available : Splash Fetcher and Base Fetcher.
//
// Base fetcher is used for downloading html web page using Go standard library's http.
//
// Splash Fetcher uses Scrapinghub splash to fetch URLs.
// Splash is a javascript rendering service.
// Read more at https://github.com/scrapinghub/splash
//
// RobotsTxtMiddleware checks if scraping of specified resource is allowed by robots.txt
//
// StorageMiddleware caches web pages content passed to Dataflow kit parser.
//
package fetch

// EOF
