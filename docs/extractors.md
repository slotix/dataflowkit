# Payload Extractor types

## Simple types
* ### text  
    returns the combined text contents of the given selection.
    * params
        * includeIfEmpty - If text is empty in the selection, then return the empty string, instead of 'nil'.  This signals that the result of this Piece should be included to the results, as opposed to omiting the empty string.
* ### multipleText - extracts the text from each element in the given selection and returns the texts as an array.
    * params
        * includeIfEmpty - If there are no items in the selection, then return empty list, instead of the 'nil'.  This signals that the result of this Piece should be included to the results, as opposed to omiting the empty list.
* ### html - extracts and returns the HTML from inside each element of the given selection, as a string.
    i.e. if our selection consists of 
    ```
    [<p><b>ONE</b></p>,
     <p><i>TWO</i></p>]
    ```
    then the output will be:  
    ```
    <b>ONE</b><i>TWO</i>
    ```
    The return type is a string of all the inner HTML joined together.
* ### outerHtml - extracts and returns the HTML of each element of the given selection, as a string.
    if our selection consists of 
    ```
    ["<div><b>ONE</b></div>", 
     "<p><i>TWO</i></p>"]
    ```
    then the output will be:
    ```
    "<div><b>ONE</b></div><p><i>TWO</i></p>"
    ```
    The return type is a string of all the outer HTML joined together.
* ### attr - extracts the value of a given HTML attribute from each element in the selection, and returns them as a list.  
    The return type of the extractor is a list of attribute values (i.e. ```[]string```).
    * params
        * attr - The HTML attribute to extract from each element.
        * alwaysReturnList - By default, if there is only a single attribute extracted, the match itself will be returned (as opposed to an array containing the single match).  
        Set alwaysReturnList to true to disable this behaviour, ensuring that the Extract function always returns an array.
        * includeIfEmpty - If no elements with this attribute are found, then return the empty list instead of  'nil'. This signals that the result of this Piece should be included to the results, as opposed to omiting the empty list.
* ### regex- runs the given regex over the contents of each element in the given selection, and, for each match, extracts the given subexpression.  
    The return type of the extractor is a list of string matches (i.e. ```[]string```).
    * params
        * regexp - The regular expression to match.  This regular expression must define exactly one parenthesized subexpression (sometimes known as a "capturing group"), which will be extracted.
        * subExpression - The subexpression of the regex to match.  If this value is not set, and if the given regex has more than one subexpression, an error will be thrown.
        * onlyText - When onlyText is true, only run the given regex over the text contents of each element in the selection, as opposed to the HTML contents.
        * alwaysReturnList - By default, if there is only a single match, Regex will return the match itself (as opposed to an array containing the single match).  
        Set alwaysReturnList to true to disable this behaviour, ensuring that the Extract function always returns an array.
        * includeIfEmpty - If no matches of the provided regex could be extracted, then return the empty list, instead of 'nil'.  This signals that the result of this Piece should be included to the results, as opposed to omiting the empty list.
* ### const - returns a constant value.  
    * params  
        * value - The value to return.
* ### count - extracts the count of elements that are matched and returns it.


## Complex types
* ### link - extracts ```attr="href"``` and text from specified field.
* ### image - extracts ```attr="src"``` and ```attr="alt"``` from specified field.