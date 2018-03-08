// Dataflow kit - splash
//
// Copyright © 2017-2018 Slotix s.r.o. <dm@slotix.sk>
//
// All rights reserved. Use of this source code is governed
// by the BSD 3-Clause License license.

// Package splash of the Dataflow kit is for accessing remote Splash server which is used  for fetching html web pages.
//
// In most cases it is not enough to just download page source. Actual data can't be retrieved without running javascript code on a downloaded page first.
//
// Splash is a javascript rendering service. It is used by Dataflow kit as a fetcher in opposite to Base fetcher which just download web page content as is.
//
// Read more at https://github.com/scrapinghub/splash
//
//	Request filters
//
//  Splash supports filtering requests based on Adblock Plus rules https://adblockplus.org/ . 
// You can use filters from EasyList to remove ads and tracking codes (and thus speedup page loading), and/or write filters manually to block some of the requests (e.g. to prevent rendering of images, mp3 files, custom fonts, etc.)
//  Read more about filters at http://splash.readthedocs.io/en/latest/api.html#request-filters
//
package splash

// EOF
