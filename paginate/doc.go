// Dataflow kit - paginate
//
//Copyright for portions of Dataflow kit are held by Andrew Dunham, 2016 as part of goscrape.
//All other copyright for Dataflow kit are held by Slotix s.r.o., 2017-2018
//
// All rights reserved. Use of this source code is governed
// by the BSD 3-Clause License license.

// Package paginate of the Dataflow kit describes Paginator interface to retrieve the next page from the current one. 
//
//Next page can be obtained in several ways:
//
// BySelector returns a Paginator that extracts the next page from a document by
// querying a given CSS selector and extracting the given HTML attribute from the
// resulting element.
//
// ByQueryParam returns a Paginator that returns the next page from a document
// by incrementing a given query parameter.  Note that this will paginate
// infinitely - you probably want to specify a maximum number of pages to
// scrape by using MaxPages parameter of ScrapeOptions.
//
package paginate

// EOF
