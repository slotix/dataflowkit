// Dataflow kit - extract
//
//Copyright for portions of Dataflow kit are held by Andrew Dunham, 2016 as part of goscrape.
//All other copyright for Dataflow kit are held by Slotix s.r.o., 2017-2018
//
// All rights reserved. Use of this source code is governed
// by the BSD 3-Clause License license.

// Package extract of the Dataflow kit describes available extractors to retrieve a structured data from html web pages. The following extractor types are available: Text, HTML, OuterHTML, Attr, Link, Image, Regex.
//
//Filters are used to manipulate text data when extracting.
//Currently the following filters are available:
	//upperCase returns a copy of the Attr with all Unicode letters mapped to their upper case.
	//lowerCase returns a copy of the Attr with all Unicode letters mapped to their lower case.
	//trim returns a copy of the Attr, with all leading and trailing white space removed
//Filters are available for Text, Link and Image extractor types. 
//Image alt attribute, Link Text and Text are influenced by specified filters. 
package extract

// EOF
