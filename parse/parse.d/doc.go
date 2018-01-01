// Dataflow kit - main
//
// Copyright © 2017-2018 Slotix s.r.o. <dm@slotix.sk>
//
//
// All rights reserved. Use of this source code is governed
// by the BSD 3-Clause License license.

// Parse service of the Dataflow kit parses html content from web pages following the rules described in Payload JSON file. 
// Examples for accessing Parse endpoint are here :
//

/* 		{
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
		} */
//
package main

// EOF
