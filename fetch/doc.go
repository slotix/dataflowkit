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
// Currently two types of fetcher are available : Chrome Fetcher and Base Fetcher.
//
// Base fetcher is used for downloading html web page using Go standard Http library.
//
// Chrome Fetcher connects to Headless Chrome which renders JavaScript pages.
//
// RobotsTxtMiddleware checks if scraping of specified resource is allowed by robots.txt
//
package fetch

// EOF
