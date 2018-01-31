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
// Accessing Fetcher endpoints
//
// Examples
//		curl -XPOST  localhost:8000/fetch/splash -d '{"url":"http://example.com"}'
//		curl -XPOST  localhost:8000/fetch/base -d '{"url":"http://example.com"}'
//
// Flags and configuration settings
//
//General settings
//		DFK_FETCH: HTTP listen address of Fetch service (defaults to "127.0.0.1:8000")
//		SPLASH: Splash server host address. (defaults to "127.0.0.1:8050")
//		SPLASH_TIMEOUT: Timeout in seconds for the page render (defaults to 20)
//		SPLASH_RESOURCE_TIMEOUT: Timeout in seconds for individual network requests 
//		(defaults to 30)
//		SPLASH_WAIT: Time in seconds to wait for updates after page is loaded
//		(defaults to 0.5). Increase this value if you expect pages to contain
//		setInterval/setTimeout javascript calls, because with wait=0 callbacks of
//		setInterval/setTimeout won’t be executed. SPLASH_WAIT time must be less than SPLASH_TIMEOUT
//Storage settings
//		SKIP_STORAGE_MW: If true no data will be saved to storage. 
//		This flag forces fetcher to bypass storage middleware.
//		STORAGE_TYPE: Storage backend for intermediary data passed to Dataflow 
//		kit Parse service. Types: S3, Digital Ocean Spaces, Redis, Diskv 
//		(defaults to "Diskv"). It is case insensitive.
//		IGNORE_CACHE_INFO:  If a website is not cachable by some reason, 
//		ignore this and use cached copy if any (defaults to false). 
//		Don't set it to true in production environment.
//		DISKV_BASE_DIR: diskv base directory for storing fetched html pages (defaults to "diskv").
//		Find more information about Diskv storage at https://github.com/peterbourgon/diskv
//		SPACES_ENDPOINT: Digital Ocean Spaces Endpoint Address.
//		Find more information about DO Spaces at https://www.digitalocean.com/community/tutorials/an-introduction-to-digitalocean-spaces
//		SPACES_CONFIG: Digital Ocean Spaces Configuration file location.
//		(defaults to "~/.spaces/credentials")
//		S3_REGION: AWS S3 or Digital Ocean Spaces region (defaults to "us-east-1")
//		DFK_BUCKET: Amazon AWS S3 or Digital Ocean Spaces bucket name for storing 
//		fetch results. (defaults to "dfk-storage")
//		REDIS: Redis host address (defaults to "127.0.0.1:6379")
//		REDIS_EXPIRE: Default Redis expire value in seconds  (defaults to 3600)
//		REDIS_NETWORK: Redis Network (defaults to "tcp")
//		REDIS_PASSWORD: Redis Password (defaults to "")
//		REDIS_DB: Redis database (defaults to 0)
//		REDIS_SOCKET_PATH: Redis Socket Path (defaults to "")
//
package main

// EOF
