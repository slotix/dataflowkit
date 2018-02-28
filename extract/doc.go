// Dataflow kit - extract
//
//Copyright for portions of Dataflow kit are held by Andrew Dunham, 2016 as part of goscrape.
//All other copyright for Dataflow kit are held by Slotix s.r.o., 2017-2018
//
// All rights reserved. Use of this source code is governed
// by the BSD 3-Clause License license.

// Package extract of the Dataflow kit describes available extractors to retrieve a structured data from html web pages. 
//
// Extractor types
//
// - Text Extractor returns the combined text contents of the given selection.
//
// - HTML Extractor returns the HTML from inside each part of the
// given selection, as a string.
//
// Note that this results in what is effectively the innerHTML of the element -
// i.e. if our selection consists of 
//	["<p><b>ONE</b></p>", "<p><i>TWO</i></p>"]
//
// then the output will be: 
//	"<b>ONE</b><i>TWO</i>".
//
// The return type is a string of all the inner HTML joined together.
//
// - OuterHTML Extractor returns the HTML of each part of the
// given selection, as a string.
//
//if our selection consists of
//	["<div><b>ONE</b></div>", "<p><i>TWO</i></p>"]
// then the output will be:
//	"<div><b>ONE</b></div><p><i>TWO</i></p>".
//The return type is a string of all the outer HTML joined together.
//
// - Attr extracts the value of a given HTML attribute from each part
// in the selection, and returns them as a list.
//
// The return type of the extractor is a list of attribute values (i.e. []string).
//
// - Regex  runs the given regex over the contents of each part in the
// given selection, and, for each match, extracts the given subexpression.
//
// The return type of the extractor is a list of string matches (i.e. []string).
//
//Filters
//
//Filters are used to manipulate text data when extracting.
//
//The following filters are available:
//
//- upperCase makes all of the letters in the Extractor's text/ Attr  uppercase.
//
//- lowerCase  makes all of the letters in the Extractor's text/ Attr   lowercase.
//
//- capitalize capitalizes the first letter of each word in the Extractor's text/ Attr 
//
//- trim returns a copy of the Extractor's text/ Attr, with all leading and trailing white space removed
//
//Filters are available for Text, Link and Image extractor types.
//
//Image alt attribute, Link Text and Text are influenced by specified filters.
package extract

// EOF
