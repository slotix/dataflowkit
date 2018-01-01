// Dataflow kit - extract
//
//Copyright for portions of Dataflow kit are held by Andrew Dunham, 2016 as part of goscrape.
//All other copyright for Dataflow kit are held by Slotix s.r.o., 2017-2018
//
// All rights reserved. Use of this source code is governed
// by the BSD 3-Clause License license.

// Package extract of the Dataflow kit describes available extractors to retrieve a structured data from html web pages. The following extractor types are available: Text, HTML, OuterHTML, Attr, Link, Image, Regex.
// Package dataflowkit is the Extractor types.
//
// text
//
// text returns the combined text contents of the given selection.
//
// params
//
// includeIfEmpty
//
// If text is empty in the selection, then return the empty string, instead of 'nil'.  This signals that the result of this Part should be included to the results, as opposed to omiting the empty string.
//
// multipleText
//
// multipleText extracts the text from each part in the given selection and returns the texts as an array.
//
// params
//
// includeIfEmpty
//
// If there are no items in the selection, then return empty list, instead of the 'nil'.  This signals that the result of this Part should be included to the results, as opposed to omiting the empty list.
//
// html
//
// html extracts and returns the HTML from inside each part of the given selection, as a string.
//
// i.e. if our selection consists of
//
//   ​```
//   [<p><b>ONE</b></p>,
//    <p><i>TWO</i></p>]
//   ​```
//   then the output will be:  
//   ​```
//   <b>ONE</b><i>TWO</i>
//   ​```  
//
// The return type is a string of all the inner HTML joined together.
//
// outerHtml
//
// outerHtml extracts and returns the HTML of each part of the given selection, as a string.
//
// If our selection consists of
//
//   ​```
//   ["<div><b>ONE</b></div>", 
//    "<p><i>TWO</i></p>"]
//   ​```
//   then the output will be:
//   ​```
//   "<div><b>ONE</b></div><p><i>TWO</i></p>"
//   ​```  
//
// The return type is a string of all the outer HTML joined together.
//
// attr
//
// attr extracts the value of a given HTML attribute from each part in the selection, and returns them as a list.
//
// The return type of the extractor is a list of attribute values (i.e. []string).
//
// params
//
// attr
//
// The HTML attribute to extract from each part.
//
// alwaysReturnList
//
// By default, if there is only a single attribute extracted, the match itself will be returned (as opposed to an array containing the single match).
//
// Set alwaysReturnList to true to disable this behaviour, ensuring that the Extract function always returns an array.
//
// includeIfEmpty
//
// If no parts with this attribute are found, then return the empty list instead of  'nil'. This signals that the result of this Part should be included to the results, as opposed to omiting the empty list.
//
// link
//
// link extracts attr="href" and text from specified field.
//
// image
//
// image extracts attr="src" and attr="alt" from specified field.
//
// regex
//
// regex runs the given regex over the contents of each part in the given selection, and, for each match, extracts the given subexpression.
//
// The return type of the extractor is a list of string matches (i.e. []string).
//
// params
//
// regexp
//
// The regular expression to match.  This regular expression must define exactly one parenthesized subexpression (sometimes known as a "capturing group"), which will be extracted.
//
// subExpression
//
// The subexpression of the regex to match.  If this value is not set, and if the given regex has more than one subexpression, an error will be thrown.
//
// onlyText
//
// When onlyText is true, only run the given regex over the text contents of each part in the selection, as opposed to the HTML contents.
//
// alwaysReturnList
//
// By default, if there is only a single match, Regex will return the match itself (as opposed to an array containing the single match).
//
// Set alwaysReturnList to true to disable this behaviour, ensuring that the Extract function always returns an array.
//
// includeIfEmpty
//
// If no matches of the provided regex could be extracted, then return the empty list, instead of 'nil'.  This signals that the result of this Part should be included to the results, as opposed to omiting the empty list.
//
// const
//
// const returns a constant value.
//
// params
//
// value
//
// The value to return.
//
// count
//
// count extracts the count of parts that are matched and returns it.
//
//
package extract

// EOF
