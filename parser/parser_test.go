package parser

import (
	"log"
	"testing"

	"github.com/spf13/viper"
)

//TODO
//Article data extractions --
//http://www.wsj.com/
//http://www.bbc.com/news/
//http://www.cnn.com
//http://www.forbes.com
//http://www.bloomberg.com
//http://www.nytimes.com
//http://www.theatlantic.com
//http://www.wsj.com
//http://www.businessinsider.com
//http://thenextweb.com
//http://www.theverge.com
//http://www.xconomy.com
//https://medium.com
//http://techcrunch.com

//Product data extractions --
//http://shop.nordstrom.com
//http://www.sears.com
//http://www.walmart.com
//http://www1.macys.com
//http://www.target.com/
//http://www.bestbuy.com
//http://www.homedepot.com

func prepareData() Collections {
	/*
		viper.SetConfigName("../.dataflowkit") // name of config file (without extension)
		viper.AddConfigPath(".")              // look for config in the working directory
		if err := viper.ReadInConfig(); err != nil {
			fmt.Println(err)
			os.Exit(-1)
			// Handle errors reading the config file
		}
	*/
	viper.Set("splash","127.0.0.1:8050")
	viper.Set("splash-timeout","20")
	viper.Set("splash-resource-timeout","30")
	viper.Set("splash-wait","0,5")
	
	
	payloads := make(map[string][]byte)
	payloads["najnakup_heureka"] = []byte(`
	   { "format":"json",
	      "collections":[
	         {
	            "name":"Najnakup",
	            "url":"http://www.najnakup.sk/televizory",
	            "fields":[
	               {
	                  "name":"Recenzie",
	                  "css_selector":".item-review .ntip",
					  "type":1
	               },
	   			{
	                  "name":"Title",
	                  "css_selector":".item-title a",
					  "type":1
	               },
	   			{
	                  "name":"Percents",
	                  "css_selector":".product-view",
					  "type":1
	               }

	            ]
	         },
	         {
	            "name":"heureka",
	            "url":"http://drony.heureka.sk",
	            "fields":[
	   			 {
	                  "field_name":"Description",
	                  "css_selector":".desc .desc"
	               },
	               {
	                  "field_name":"Photo",
	                  "css_selector":".foto img"
	               },
	   			 {
	                  "field_name":"Price",
	                  "css_selector":".pricen"
	               },
	   			{
	                  "field_name":"Title",
	                  "css_selector":".product-container a"
	               },
	               {
	                  "field_name":"Count",
	                  "css_selector":".count a"
	               }
	            ]
	         }
	      ]
	   }
	   	   `)

	payloads["najnakup"] = []byte(`
		   {"format":"json",
		      "collections":[
		         {
		            "name":"Televizory",
		            "url":"https://www.najnakup.sk/televizory",
		            "fields":[
					   {
		                  "field_name":"Image",
		                  "css_selector":"#list-view-cont .lazy_activated"
		               },
		   			 {
		                  "field_name":"Percents",
		                  "css_selector":".product-view"
		               },
		                {
		                  "field_name":"Recenzie",
		                  "css_selector":".item-review .ntip"
		               },
					    {
		                  "field_name":"Title",
		                  "css_selector":".item-title a"
		               },
					   {
		                  "field_name":"Price",
		                  "css_selector":".cost strong"
		               }

					   
		            ]
		         }
		      ]
		   }
		   	   `)

	payloads["heureka"] = []byte(`
	   {"format":"json",
	      "collections":[
	         {
	            "name":"collection2",
	            "url":"http://drony.heureka.sk",
	            "fields":[
					 {
	                  "name":"Reviews",
	                  "css_selector":".review-count a",
					  "type":1
	               },
				   {
	                  "name":"Promo",
	                  "css_selector":".promo-top-ico__badge span",
					  "type":1
	               },
				   
				  {
	                  "name":"Title",
	                  "css_selector":".product-container a",
					  "type":1
	               },
					{
	                  "name":"TopIcon",
	                  "css_selector":".top-ico",
					  "type":1
	               },

					{
	                  "name":"BuyInfo",
	                  "css_selector":".buy-info",
					  "type":1
	               },
					{
	                  "name":"Price",
	                  "css_selector":".pricen",
					  "type":1
	               },
					{
	                  "name":"Photo",
	                  "css_selector":".foto img",
					  "type":1
	               }
				  
	            ]
	         }
	      ]
	   }
	   	`)
	/*

	 */

	payloads["ebay1"] = []byte(`
		   {"format":"json",
		      "collections":[
		         {
		            "name":"collection1",
		            "url":"http://www.ebay.com/sch/Computers-Tablets-Networking/58058/i.html?_nkw=Apple&_ipg=25&rt=nc",
		            "fields":[
		   			 {
		                  "field_name":"Title",
		                  "css_selector":".vip"
		               },
		                {
		                  "field_name":"Price",
		                  "css_selector":".prc .bold"
		               },
		   			{
		                  "field_name":"Sold",
		                  "css_selector":".red"
		               },
					    {
		                  "field_name":"Image",
		                  "css_selector":".img"
		               }
		            ]
		         }
		      ]
		   }
		   	   `)

	payloads["diesel"] = []byte(`
{
   "format":"xml",
   "collections":[
      {
         "name":"Collection1",
         "url":"http://diesel.elcat.kg",
         "fields":[
            {
               "name":"link1",
               "css":"h4 a",
               "type":2,
               "details":{
                  "name":"link1details",
                  "url":"http://diesel.elcat.kg/index.php?s=d274ba35aefc0c250968f376227468ba&showforum=29",
                  "fields":[
                     {
                        "name":"link1",
                        "css":"h4 a",
                        "type":2,
                        "details":null,
                        "count":24
                     }
                  ]
               },
               "count":144
            }
         ]
      },
      {
         "name":"Collection2",
         "url":"http://diesel.elcat.kg/index.php?s=d274ba35aefc0c250968f376227468ba&showforum=376",
         "fields":[
            {
               "name":"link1",
               "css":".col_c_forum a",
               "type":2,
               "details":{
                  "name":"link1details",
                  "url":"http://diesel.elcat.kg/index.php?s=d274ba35aefc0c250968f376227468ba&showforum=28",
                  "fields":[
                     {
                        "name":"text1",
                        "css":".topic_title span",
                        "type":1,
                        "details":null,
                        "count":40
                     }
                  ]
               },
               "count":4
            }
         ]
      }
   ]
}
`)

	payloads["amazon"] = []byte(`
				{"format":"json","collections":[{"name":"Amazon","url":"https://www.amazon.com/Best-Sellers-Electronics-Computer-Monitors/zgbs/electronics/1292115011/ref=s9_acss_bw_cg_PCMONSBC_1a1_w?pf_rd_m=ATVPDKIKX0DER&pf_rd_s=merchandised-search-2&pf_rd_r=G2MQWH6WP6X2TNHYC8KS&pf_rd_t=101&pf_rd_p=bc782040-3bc0-4dfb-bdb6-5eb097c1f272&pf_rd_i=1292115011","fields":[
				
				{"field_name":"Title","css_selector":".zg_title a"},
				{"field_name":"Price","css_selector":".zg_price .price"},
				{"field_name":"Image","css_selector":"#zg_centerListWrapper img"},
				{"field_name":"Reviews","css_selector":".a-size-small .a-link-normal"}]}]}
`)

	payloads["nbkrExchange"] = []byte(`
				{"format":"json","collections":[{"name":"NBKR Exchange","url":"http://nbkr.kg/index.jsp?lang=RUS","fields":[
				
				{"field_name":"Pairs","css_selector":".excurr:nth-child(1)"},
				{"field_name":"Yesterday","css_selector":".exrate:nth-child(2)"},
				{"field_name":"Today","css_selector":".exrate:nth-child(3)"},
				{"field_name":"UpDown","css_selector":"td:nth-child(4)"}]}]}
`)

	payloads["nbkrGold"] = []byte(`
				{"format":"json","collections":[{"name":"NBKR Gold","url":"http://nbkr.kg/index.jsp?lang=RUS","fields":[
				
				{"field_name":"Weight","css_selector":"#sticker-gold td:nth-child(1)"},
				{"field_name":"Buy","css_selector":"#sticker-gold td:nth-child(2)"},
				{"field_name":"Sell","css_selector":"#sticker-gold td:nth-child(3)"}]}]}
`)

	payloads["edPlane"] = []byte(`
				{"format":"json","collections":[{"name":"EdPlane24","url":"http://www.ebay.com/sch/edplane24/m.html?_nkw=&_armrs=1&_ipg=&_from=","fields":[
				
				
				{"field_name":"Title","css_selector":".vip"},
				{"field_name":"Price","css_selector":".prc .bold"},
				{"field_name":"Image","css_selector":".img"}]}]}
`)

	//Failed
	payloads["yahooPolitics"] = []byte(`
				{"format":"json","collections":[{"name":"Yahoo politics","url":"https://www.yahoo.com/news/politics/","fields":[			
				{"field_name":"Title","css_selector":""},
				{"field_name":"Text","css_selector":""},
				{"field_name":"Image","css_selector":""}]}]}
`)
	//Failed
	payloads["buysellcyprus"] = []byte(`
				{"format":"json","collections":[{"name":"buysellcyprus","url":"http://www.buysellcyprus.com/nqcontent.cfm?a_name=v4_map_search","fields":[			
				{"field_name":"Title","css_selector":".hometitle b"},
				{"field_name":"Price","css_selector":".red b"},
				{"field_name":"Image","css_selector":"#frmResults a > img"},
				{"field_name":"Type","css_selector":".price+ .homedetailsInside b"},
				{"field_name":"sqm","css_selector":".homedetailsInside b+ b"}]}]}
`)

	//
	payloads["wiki"] = []byte(`
				{"format":"json","collections":[{"name":"Wiki","url":"https://www.wikipedia.org","fields":[			
				{"field_name":"Title","css_selector":".link-box strong"},
				{"field_name":"Text","css_selector":".link-box small"}
			]}]}
`)

	/*

	 */

	var p Parser
	err := p.UnmarshalJSON(payloads["diesel"])
	if err != nil {

		log.Fatal(err)
	}
	out, err := p.Parse()
	if err != nil {
		log.Fatal(err)
	}
	return *out
}

func (out *Collections) marshalXML() {
	buf, err := out.MarshalXML()
	if err != nil {
		log.Println(err)
	}
	log.Println(string(buf))
}

/*
func (out *Out) marshalCSV() {
	buf, err := out.MarshalCSV()
	if err != nil {
		log.Println(err)
	}
	log.Println(string(buf))
}
*/

func (out *Collections) marshalJSON() {
	buf, err := out.MarshalJSON()
	if err != nil {
		log.Println(err)
	}
	log.Println(string(buf))
}

func TestOut(t *testing.T) {
	out := prepareData()
	out.marshalJSON()
	//out.marshalXML()
	//b, err := out.MarshalCSV()
	//_, err := out.SaveExcel("/tmp/out.xlsx")

	//if err == nil {
	//	log.Println(err)
	//}
	//log.Println(string(b))

}
