package scrape

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/slotix/dataflowkit/fetch"
	"github.com/slotix/dataflowkit/splash"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/temoto/robotstxt"
)

var (
	randomize               bool
	delayFetch              time.Duration
	paginateResults         bool
	pJSON, pCSV_XML         Payload
	outJSON, outCSV, outXML []byte
)

func init() {
	viper.Set("SPLASH", "127.0.0.1:8050")
	viper.Set("SPLASH_TIMEOUT", 20)
	viper.Set("SPLASH_RESOURCE_TIMEOUT", 30)
	viper.Set("SPLASH_WAIT", 0.5)
	viper.Set("DFK_FETCH", "127.0.0.1:8000")
	randomize = true
	//delayFetch = 500 * time.Millisecond
	delayFetch = 0
	paginateResults = false
	pJSON = Payload{
		Name: "books-to-scrape",
		// Request: splash.Request{
		// 	URL: "http://books.toscrape.com",
		// },
		Request: fetch.BaseFetcherRequest{
			URL: "http://books.toscrape.com",
		},
		Fields: []Field{
			Field{
				Name:     "Title",
				Selector: "h3 a",
				Extractor: Extractor{
					Types:   []string{"text", "href"},
					Filters: []string{"trim"},
				},
				Details: &details{
					Fields: []Field{
						Field{
							Name:     "Availability",
							Selector: ".availability",
							Extractor: Extractor{
								Types:   []string{"text"},
								Filters: []string{"trim"},
							},
						},
					},
				},
			},
			Field{
				Name:     "Price",
				Selector: ".price_color",
				Extractor: Extractor{
					Types: []string{"regex"},
					Params: map[string]interface{}{
						"regexp": "([\\d\\.]+)",
					},
					Filters: []string{"trim"},
				},
			},
			Field{
				Name:     "Image",
				Selector: ".thumbnail",
				Extractor: Extractor{
					Types:   []string{"src", "alt"},
					Filters: []string{"trim"},
				},
			},
		},
		Paginator: &paginator{
			Selector:  ".next a",
			Attribute: "href",
			MaxPages:  2,
		},
		RandomizeFetchDelay: &randomize,
		FetchDelay:          &delayFetch,
		Format:              "json",
		PaginateResults:     &paginateResults,
	}
	pCSV_XML = Payload{
		Name: "books-to-scrape",
		Request: splash.Request{
			URL: "http://books.toscrape.com",
		},
		Fields: []Field{
			Field{
				Name:     "Title",
				Selector: "h3 a",
				Extractor: Extractor{
					Types: []string{"text"},
				},
			},
			Field{
				Name:     "Price",
				Selector: ".price_color",
				Extractor: Extractor{
					Types: []string{"text"},
				},
			},
		},
		RandomizeFetchDelay: &randomize,
		FetchDelay:          &delayFetch,
		Format:              "csv",
		PaginateResults:     &paginateResults,
	}
	outJSON = []byte(`[{"Image_alt":"A Light in the Attic","Image_src":"http://books.toscrape.com/media/cache/2c/da/2cdad67c44b002e7ead0cc35693c0e8b.jpg","Price_regex":"51.77","Title_href":"http://books.toscrape.com/catalogue/a-light-in-the-attic_1000/index.html","Title_href_details":[{"Availability_text":"In stock (22 available)"}],"Title_text":"A Light in the ..."},{"Image_alt":"Tipping the Velvet","Image_src":"http://books.toscrape.com/media/cache/26/0c/260c6ae16bce31c8f8c95daddd9f4a1c.jpg","Price_regex":"53.74","Title_href":"http://books.toscrape.com/catalogue/tipping-the-velvet_999/index.html","Title_href_details":[{"Availability_text":"In stock (20 available)"}],"Title_text":"Tipping the Velvet"},{"Image_alt":"Soumission","Image_src":"http://books.toscrape.com/media/cache/3e/ef/3eef99c9d9adef34639f510662022830.jpg","Price_regex":"50.10","Title_href":"http://books.toscrape.com/catalogue/soumission_998/index.html","Title_href_details":[{"Availability_text":"In stock (20 available)"}],"Title_text":"Soumission"},{"Image_alt":"Sharp Objects","Image_src":"http://books.toscrape.com/media/cache/32/51/3251cf3a3412f53f339e42cac2134093.jpg","Price_regex":"47.82","Title_href":"http://books.toscrape.com/catalogue/sharp-objects_997/index.html","Title_href_details":[{"Availability_text":"In stock (20 available)"}],"Title_text":"Sharp Objects"},{"Image_alt":"Sapiens: A Brief History of Humankind","Image_src":"http://books.toscrape.com/media/cache/be/a5/bea5697f2534a2f86a3ef27b5a8c12a6.jpg","Price_regex":"54.23","Title_href":"http://books.toscrape.com/catalogue/sapiens-a-brief-history-of-humankind_996/index.html","Title_href_details":[{"Availability_text":"In stock (20 available)"}],"Title_text":"Sapiens: A Brief History ..."},{"Image_alt":"The Requiem Red","Image_src":"http://books.toscrape.com/media/cache/68/33/68339b4c9bc034267e1da611ab3b34f8.jpg","Price_regex":"22.65","Title_href":"http://books.toscrape.com/catalogue/the-requiem-red_995/index.html","Title_href_details":[{"Availability_text":"In stock (19 available)"}],"Title_text":"The Requiem Red"},{"Image_alt":"The Dirty Little Secrets of Getting Your Dream Job","Image_src":"http://books.toscrape.com/media/cache/92/27/92274a95b7c251fea59a2b8a78275ab4.jpg","Price_regex":"33.34","Title_href":"http://books.toscrape.com/catalogue/the-dirty-little-secrets-of-getting-your-dream-job_994/index.html","Title_href_details":[{"Availability_text":"In stock (19 available)"}],"Title_text":"The Dirty Little Secrets ..."},{"Image_alt":"The Coming Woman: A Novel Based on the Life of the Infamous Feminist, Victoria Woodhull","Image_src":"http://books.toscrape.com/media/cache/3d/54/3d54940e57e662c4dd1f3ff00c78cc64.jpg","Price_regex":"17.93","Title_href":"http://books.toscrape.com/catalogue/the-coming-woman-a-novel-based-on-the-life-of-the-infamous-feminist-victoria-woodhull_993/index.html","Title_href_details":[{"Availability_text":"In stock (19 available)"}],"Title_text":"The Coming Woman: A ..."},{"Image_alt":"The Boys in the Boat: Nine Americans and Their Epic Quest for Gold at the 1936 Berlin Olympics","Image_src":"http://books.toscrape.com/media/cache/66/88/66883b91f6804b2323c8369331cb7dd1.jpg","Price_regex":"22.60","Title_href":"http://books.toscrape.com/catalogue/the-boys-in-the-boat-nine-americans-and-their-epic-quest-for-gold-at-the-1936-berlin-olympics_992/index.html","Title_href_details":[{"Availability_text":"In stock (19 available)"}],"Title_text":"The Boys in the ..."},{"Image_alt":"The Black Maria","Image_src":"http://books.toscrape.com/media/cache/58/46/5846057e28022268153beff6d352b06c.jpg","Price_regex":"52.15","Title_href":"http://books.toscrape.com/catalogue/the-black-maria_991/index.html","Title_href_details":[{"Availability_text":"In stock (19 available)"}],"Title_text":"The Black Maria"},{"Image_alt":"Starving Hearts (Triangular Trade Trilogy, #1)","Image_src":"http://books.toscrape.com/media/cache/be/f4/bef44da28c98f905a3ebec0b87be8530.jpg","Price_regex":"13.99","Title_href":"http://books.toscrape.com/catalogue/starving-hearts-triangular-trade-trilogy-1_990/index.html","Title_href_details":[{"Availability_text":"In stock (19 available)"}],"Title_text":"Starving Hearts (Triangular Trade ..."},{"Image_alt":"Shakespeare's Sonnets","Image_src":"http://books.toscrape.com/media/cache/10/48/1048f63d3b5061cd2f424d20b3f9b666.jpg","Price_regex":"20.66","Title_href":"http://books.toscrape.com/catalogue/shakespeares-sonnets_989/index.html","Title_href_details":[{"Availability_text":"In stock (19 available)"}],"Title_text":"Shakespeare's Sonnets"},{"Image_alt":"Set Me Free","Image_src":"http://books.toscrape.com/media/cache/5b/88/5b88c52633f53cacf162c15f4f823153.jpg","Price_regex":"17.46","Title_href":"http://books.toscrape.com/catalogue/set-me-free_988/index.html","Title_href_details":[{"Availability_text":"In stock (19 available)"}],"Title_text":"Set Me Free"},{"Image_alt":"Scott Pilgrim's Precious Little Life (Scott Pilgrim #1)","Image_src":"http://books.toscrape.com/media/cache/94/b1/94b1b8b244bce9677c2f29ccc890d4d2.jpg","Price_regex":"52.29","Title_href":"http://books.toscrape.com/catalogue/scott-pilgrims-precious-little-life-scott-pilgrim-1_987/index.html","Title_href_details":[{"Availability_text":"In stock (19 available)"}],"Title_text":"Scott Pilgrim's Precious Little ..."},{"Image_alt":"Rip it Up and Start Again","Image_src":"http://books.toscrape.com/media/cache/81/c4/81c4a973364e17d01f217e1188253d5e.jpg","Price_regex":"35.02","Title_href":"http://books.toscrape.com/catalogue/rip-it-up-and-start-again_986/index.html","Title_href_details":[{"Availability_text":"In stock (19 available)"}],"Title_text":"Rip it Up and ..."},{"Image_alt":"Our Band Could Be Your Life: Scenes from the American Indie Underground, 1981-1991","Image_src":"http://books.toscrape.com/media/cache/54/60/54607fe8945897cdcced0044103b10b6.jpg","Price_regex":"57.25","Title_href":"http://books.toscrape.com/catalogue/our-band-could-be-your-life-scenes-from-the-american-indie-underground-1981-1991_985/index.html","Title_href_details":[{"Availability_text":"In stock (19 available)"}],"Title_text":"Our Band Could Be ..."},{"Image_alt":"Olio","Image_src":"http://books.toscrape.com/media/cache/55/33/553310a7162dfbc2c6d19a84da0df9e1.jpg","Price_regex":"23.88","Title_href":"http://books.toscrape.com/catalogue/olio_984/index.html","Title_href_details":[{"Availability_text":"In stock (19 available)"}],"Title_text":"Olio"},{"Image_alt":"Mesaerion: The Best Science Fiction Stories 1800-1849","Image_src":"http://books.toscrape.com/media/cache/09/a3/09a3aef48557576e1a85ba7efea8ecb7.jpg","Price_regex":"37.59","Title_href":"http://books.toscrape.com/catalogue/mesaerion-the-best-science-fiction-stories-1800-1849_983/index.html","Title_href_details":[{"Availability_text":"In stock (19 available)"}],"Title_text":"Mesaerion: The Best Science ..."},{"Image_alt":"Libertarianism for Beginners","Image_src":"http://books.toscrape.com/media/cache/0b/bc/0bbcd0a6f4bcd81ccb1049a52736406e.jpg","Price_regex":"51.33","Title_href":"http://books.toscrape.com/catalogue/libertarianism-for-beginners_982/index.html","Title_href_details":[{"Availability_text":"In stock (19 available)"}],"Title_text":"Libertarianism for Beginners"},{"Image_alt":"It's Only the Himalayas","Image_src":"http://books.toscrape.com/media/cache/27/a5/27a53d0bb95bdd88288eaf66c9230d7e.jpg","Price_regex":"45.17","Title_href":"http://books.toscrape.com/catalogue/its-only-the-himalayas_981/index.html","Title_href_details":[{"Availability_text":"In stock (19 available)"}],"Title_text":"It's Only the Himalayas"},{"Image_alt":"In Her Wake","Image_src":"http://books.toscrape.com/media/cache/5d/72/5d72709c6a7a9584a4d1cf07648bfce1.jpg","Price_regex":"12.84","Title_href":"http://books.toscrape.com/catalogue/in-her-wake_980/index.html","Title_href_details":[{"Availability_text":"In stock (19 available)"}],"Title_text":"In Her Wake"},{"Image_alt":"How Music Works","Image_src":"http://books.toscrape.com/media/cache/5c/c8/5cc8e107246cb478960d4f0aba1e1c8e.jpg","Price_regex":"37.32","Title_href":"http://books.toscrape.com/catalogue/how-music-works_979/index.html","Title_href_details":[{"Availability_text":"In stock (19 available)"}],"Title_text":"How Music Works"},{"Image_alt":"Foolproof Preserving: A Guide to Small Batch Jams, Jellies, Pickles, Condiments, and More: A Foolproof Guide to Making Small Batch Jams, Jellies, Pickles, Condiments, and More","Image_src":"http://books.toscrape.com/media/cache/9f/59/9f59f01fa916a7bb8f0b28a4012179a4.jpg","Price_regex":"30.52","Title_href":"http://books.toscrape.com/catalogue/foolproof-preserving-a-guide-to-small-batch-jams-jellies-pickles-condiments-and-more-a-foolproof-guide-to-making-small-batch-jams-jellies-pickles-condiments-and-more_978/index.html","Title_href_details":[{"Availability_text":"In stock (19 available)"}],"Title_text":"Foolproof Preserving: A Guide ..."},{"Image_alt":"Chase Me (Paris Nights #2)","Image_src":"http://books.toscrape.com/media/cache/9c/2e/9c2e0eb8866b8e3f3b768994fd3d1c1a.jpg","Price_regex":"25.27","Title_href":"http://books.toscrape.com/catalogue/chase-me-paris-nights-2_977/index.html","Title_href_details":[{"Availability_text":"In stock (19 available)"}],"Title_text":"Chase Me (Paris Nights ..."},{"Image_alt":"Black Dust","Image_src":"http://books.toscrape.com/media/cache/44/cc/44ccc99c8f82c33d4f9d2afa4ef25787.jpg","Price_regex":"34.53","Title_href":"http://books.toscrape.com/catalogue/black-dust_976/index.html","Title_href_details":[{"Availability_text":"In stock (19 available)"}],"Title_text":"Black Dust"},{"Image_alt":"Birdsong: A Story in Pictures","Image_src":"http://books.toscrape.com/media/cache/af/6e/af6e796160fe63e0cf19d44395c7ddf2.jpg","Price_regex":"54.64","Title_href":"http://books.toscrape.com/catalogue/birdsong-a-story-in-pictures_975/index.html","Title_href_details":[{"Availability_text":"In stock (19 available)"}],"Title_text":"Birdsong: A Story in ..."},{"Image_alt":"America's Cradle of Quarterbacks: Western Pennsylvania's Football Factory from Johnny Unitas to Joe Montana","Image_src":"http://books.toscrape.com/media/cache/ef/0b/ef0bed08de4e083dba5e20fdb98d9c36.jpg","Price_regex":"22.50","Title_href":"http://books.toscrape.com/catalogue/americas-cradle-of-quarterbacks-western-pennsylvanias-football-factory-from-johnny-unitas-to-joe-montana_974/index.html","Title_href_details":[{"Availability_text":"In stock (19 available)"}],"Title_text":"America's Cradle of Quarterbacks: ..."},{"Image_alt":"Aladdin and His Wonderful Lamp","Image_src":"http://books.toscrape.com/media/cache/d6/da/d6da0371958068bbaf39ea9c174275cd.jpg","Price_regex":"53.13","Title_href":"http://books.toscrape.com/catalogue/aladdin-and-his-wonderful-lamp_973/index.html","Title_href_details":[{"Availability_text":"In stock (19 available)"}],"Title_text":"Aladdin and His Wonderful ..."},{"Image_alt":"Worlds Elsewhere: Journeys Around Shakespeare’s Globe","Image_src":"http://books.toscrape.com/media/cache/2e/98/2e98c332bf8563b584784971541c4445.jpg","Price_regex":"40.30","Title_href":"http://books.toscrape.com/catalogue/worlds-elsewhere-journeys-around-shakespeares-globe_972/index.html","Title_href_details":[{"Availability_text":"In stock (18 available)"}],"Title_text":"Worlds Elsewhere: Journeys Around ..."},{"Image_alt":"Wall and Piece","Image_src":"http://books.toscrape.com/media/cache/a5/41/a5416b9646aaa7287baa287ec2590270.jpg","Price_regex":"44.18","Title_href":"http://books.toscrape.com/catalogue/wall-and-piece_971/index.html","Title_href_details":[{"Availability_text":"In stock (18 available)"}],"Title_text":"Wall and Piece"},{"Image_alt":"The Four Agreements: A Practical Guide to Personal Freedom","Image_src":"http://books.toscrape.com/media/cache/0f/7e/0f7ee69495c0df1d35723f012624a9f8.jpg","Price_regex":"17.66","Title_href":"http://books.toscrape.com/catalogue/the-four-agreements-a-practical-guide-to-personal-freedom_970/index.html","Title_href_details":[{"Availability_text":"In stock (18 available)"}],"Title_text":"The Four Agreements: A ..."},{"Image_alt":"The Five Love Languages: How to Express Heartfelt Commitment to Your Mate","Image_src":"http://books.toscrape.com/media/cache/38/c5/38c56fba316c07305643a8065269594e.jpg","Price_regex":"31.05","Title_href":"http://books.toscrape.com/catalogue/the-five-love-languages-how-to-express-heartfelt-commitment-to-your-mate_969/index.html","Title_href_details":[{"Availability_text":"In stock (18 available)"}],"Title_text":"The Five Love Languages: ..."},{"Image_alt":"The Elephant Tree","Image_src":"http://books.toscrape.com/media/cache/5d/7e/5d7ecde8e81513eba8a64c9fe000744b.jpg","Price_regex":"23.82","Title_href":"http://books.toscrape.com/catalogue/the-elephant-tree_968/index.html","Title_href_details":[{"Availability_text":"In stock (18 available)"}],"Title_text":"The Elephant Tree"},{"Image_alt":"The Bear and the Piano","Image_src":"http://books.toscrape.com/media/cache/cf/bb/cfbb5e62715c6d888fd07794c9bab5d6.jpg","Price_regex":"36.89","Title_href":"http://books.toscrape.com/catalogue/the-bear-and-the-piano_967/index.html","Title_href_details":[{"Availability_text":"In stock (18 available)"}],"Title_text":"The Bear and the ..."},{"Image_alt":"Sophie's World","Image_src":"http://books.toscrape.com/media/cache/65/71/6571919836ec51ed54f0050c31d8a0cd.jpg","Price_regex":"15.94","Title_href":"http://books.toscrape.com/catalogue/sophies-world_966/index.html","Title_href_details":[{"Availability_text":"In stock (18 available)"}],"Title_text":"Sophie's World"},{"Image_alt":"Penny Maybe","Image_src":"http://books.toscrape.com/media/cache/12/53/1253c21c5ef3c6d075c5fa3f5fecee6a.jpg","Price_regex":"33.29","Title_href":"http://books.toscrape.com/catalogue/penny-maybe_965/index.html","Title_href_details":[{"Availability_text":"In stock (18 available)"}],"Title_text":"Penny Maybe"},{"Image_alt":"Maude (1883-1993):She Grew Up with the country","Image_src":"http://books.toscrape.com/media/cache/f5/88/f5889d038f5d8e949b494d147c2dcf54.jpg","Price_regex":"18.02","Title_href":"http://books.toscrape.com/catalogue/maude-1883-1993she-grew-up-with-the-country_964/index.html","Title_href_details":[{"Availability_text":"In stock (18 available)"}],"Title_text":"Maude (1883-1993):She Grew Up ..."},{"Image_alt":"In a Dark, Dark Wood","Image_src":"http://books.toscrape.com/media/cache/23/85/238570a1c284e730dbc737a7e631ae2b.jpg","Price_regex":"19.63","Title_href":"http://books.toscrape.com/catalogue/in-a-dark-dark-wood_963/index.html","Title_href_details":[{"Availability_text":"In stock (18 available)"}],"Title_text":"In a Dark, Dark ..."},{"Image_alt":"Behind Closed Doors","Image_src":"http://books.toscrape.com/media/cache/e1/5c/e15c289ba58cea38519e1281e859f0c1.jpg","Price_regex":"52.22","Title_href":"http://books.toscrape.com/catalogue/behind-closed-doors_962/index.html","Title_href_details":[{"Availability_text":"In stock (18 available)"}],"Title_text":"Behind Closed Doors"},{"Image_alt":"You can't bury them all: Poems","Image_src":"http://books.toscrape.com/media/cache/e9/20/e9203b733126c4a0832a1c7885dc27cf.jpg","Price_regex":"33.63","Title_href":"http://books.toscrape.com/catalogue/you-cant-bury-them-all-poems_961/index.html","Title_href_details":[{"Availability_text":"In stock (17 available)"}],"Title_text":"You can't bury them ..."}]` + "\n")

	outCSV = []byte("Title_text,Price_text\nA Light in the ...,£51.77\nTipping the Velvet,£53.74\nSoumission,£50.10\nSharp Objects,£47.82\nSapiens: A Brief History ...,£54.23\nThe Requiem Red,£22.65\nThe Dirty Little Secrets ...,£33.34\nThe Coming Woman: A ...,£17.93\nThe Boys in the ...,£22.60\nThe Black Maria,£52.15\nStarving Hearts (Triangular Trade ...,£13.99\nShakespeare's Sonnets,£20.66\nSet Me Free,£17.46\nScott Pilgrim's Precious Little ...,£52.29\nRip it Up and ...,£35.02\nOur Band Could Be ...,£57.25\nOlio,£23.88\nMesaerion: The Best Science ...,£37.59\nLibertarianism for Beginners,£51.33\nIt's Only the Himalayas,£45.17\n")

	outXML = []byte(`<?xml version="1.0" encoding="UTF-8"?><root><element><Price_text>£51.77</Price_text><Title_text>A Light in the ...</Title_text></element><element><Price_text>£53.74</Price_text><Title_text>Tipping the Velvet</Title_text></element><element><Price_text>£50.10</Price_text><Title_text>Soumission</Title_text></element><element><Price_text>£47.82</Price_text><Title_text>Sharp Objects</Title_text></element><element><Price_text>£54.23</Price_text><Title_text>Sapiens: A Brief History ...</Title_text></element><element><Price_text>£22.65</Price_text><Title_text>The Requiem Red</Title_text></element><element><Price_text>£33.34</Price_text><Title_text>The Dirty Little Secrets ...</Title_text></element><element><Price_text>£17.93</Price_text><Title_text>The Coming Woman: A ...</Title_text></element><element><Price_text>£22.60</Price_text><Title_text>The Boys in the ...</Title_text></element><element><Price_text>£52.15</Price_text><Title_text>The Black Maria</Title_text></element><element><Price_text>£13.99</Price_text><Title_text>Starving Hearts (Triangular Trade ...</Title_text></element><element><Price_text>£20.66</Price_text><Title_text>Shakespeare&apos;s Sonnets</Title_text></element><element><Price_text>£17.46</Price_text><Title_text>Set Me Free</Title_text></element><element><Price_text>£52.29</Price_text><Title_text>Scott Pilgrim&apos;s Precious Little ...</Title_text></element><element><Price_text>£35.02</Price_text><Title_text>Rip it Up and ...</Title_text></element><element><Price_text>£57.25</Price_text><Title_text>Our Band Could Be ...</Title_text></element><element><Price_text>£23.88</Price_text><Title_text>Olio</Title_text></element><element><Price_text>£37.59</Price_text><Title_text>Mesaerion: The Best Science ...</Title_text></element><element><Price_text>£51.33</Price_text><Title_text>Libertarianism for Beginners</Title_text></element><element><Price_text>£45.17</Price_text><Title_text>It&apos;s Only the Himalayas</Title_text></element></root>`)
}

func TestTask_ParseJSON(t *testing.T) {
	//start fetch server
	viper.Set("DFK_FETCH", "127.0.0.1:8000")
	fetchServerAddr := viper.GetString("DFK_FETCH")
	fetchServerCfg := fetch.Config{
		Host:         fetchServerAddr,
	//	ReadTimeout:  60 * time.Second,
	//	WriteTimeout: 60 * time.Second,
	}
	viper.Set("SKIP_STORAGE_MW", true)
	fetchServer := fetch.Start(fetchServerCfg)
	/////////
	type fields struct {
		ID      string
		Payload Payload
		Visited map[string]error
		Robots  map[string]*robotstxt.RobotsData
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "books-to-scrape-json",
			fields: fields{
				ID:      "111",
				Payload: pJSON,
				Visited: map[string]error{},
				Robots:  map[string]*robotstxt.RobotsData{},
			},
			want:    outJSON,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{
				ID:      tt.fields.ID,
				Payload: tt.fields.Payload,
				Visited: tt.fields.Visited,
				Robots:  tt.fields.Robots,
			}
			r, err := task.Parse()
			if err != nil {
				t.Error(err)
				return
			}
			buf := new(bytes.Buffer)
			buf.ReadFrom(r)
			got := buf.Bytes()
			//	t.Log(string(got))
			if (err != nil) != tt.wantErr {
				t.Errorf("Task.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Task.Parse() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
	fetchServer.Stop()
}

func TestTask_ParseCSV(t *testing.T) {
	//start fetch server
	viper.Set("DFK_FETCH", "127.0.0.1:8000")
	fetchServerAddr := viper.GetString("DFK_FETCH")
	fetchServerCfg := fetch.Config{
		Host:         fetchServerAddr,
	//	ReadTimeout:  60 * time.Second,
	//	WriteTimeout: 60 * time.Second,
	}
	viper.Set("SKIP_STORAGE_MW", true)
	fetchServer := fetch.Start(fetchServerCfg)
	///////
	type fields struct {
		ID      string
		Payload Payload
		Visited map[string]error
		Robots  map[string]*robotstxt.RobotsData
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "books-to-scrape-csv",
			fields: fields{
				ID:      "111",
				Payload: pCSV_XML,
				Visited: map[string]error{},
				Robots:  map[string]*robotstxt.RobotsData{},
			},
			want:    outCSV,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{
				ID:      tt.fields.ID,
				Payload: tt.fields.Payload,
				Visited: tt.fields.Visited,
				Robots:  tt.fields.Robots,
			}
			r, err := task.Parse()
			buf := new(bytes.Buffer)
			buf.ReadFrom(r)
			got := buf.Bytes()
			if (err != nil) != tt.wantErr {
				t.Errorf("Task.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Task.Parse() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
	fetchServer.Stop()
}

func TestTask_ParseXML(t *testing.T) {
	//start fetch server
	viper.Set("DFK_FETCH", "127.0.0.1:8000")
	fetchServerAddr := viper.GetString("DFK_FETCH")
	fetchServerCfg := fetch.Config{
		Host:         fetchServerAddr,
	//	ReadTimeout:  60 * time.Second,
	//	WriteTimeout: 60 * time.Second,
	}
	viper.Set("SKIP_STORAGE_MW", true)
	fetchServer := fetch.Start(fetchServerCfg)
	///////
	pCSV_XML.Format = "XML"

	type fields struct {
		ID      string
		Payload Payload
		Visited map[string]error
		Robots  map[string]*robotstxt.RobotsData
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "books-to-scrape-csv",
			fields: fields{
				ID:      "111",
				Payload: pCSV_XML,
				Visited: map[string]error{},
				Robots:  map[string]*robotstxt.RobotsData{},
			},
			want:    outXML,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{
				ID:      tt.fields.ID,
				Payload: tt.fields.Payload,
				Visited: tt.fields.Visited,
				Robots:  tt.fields.Robots,
			}
			r, err := task.Parse()
			buf := new(bytes.Buffer)
			_, err = buf.ReadFrom(r)
			if err != nil {
				t.Error(err)
			}
			got := buf.Bytes()
			if (err != nil) != tt.wantErr {
				t.Errorf("Task.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Task.Parse() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
	fetchServer.Stop()
}

func TestScraper_partNames(t *testing.T) {
	s := Scraper{}
	s.Parts = []Part{
		Part{Name: "1"},
		Part{Name: "2"},
		Part{Name: "3"},
		Part{Name: "4"},
	}
	parts := s.partNames()
	assert.Equal(t, []string{"1", "2", "3", "4"}, parts)

}

func TestPayload_selectors(t *testing.T) {
	p1 := Payload{
		Fields: []Field{
			Field{Selector: "sel1"},
			Field{Selector: "sel2"},
			Field{Selector: "sel3"},
			Field{Selector: "sel4"},
		},
	}
	p2 := Payload{
		Fields: []Field{
			Field{},
			Field{},
			Field{},
			Field{},
		},
	}

	s1, err := p1.selectors()
	assert.NoError(t, err)
	assert.Equal(t, []string{"sel1", "sel2", "sel3", "sel4"}, s1)
	s2, err := p2.selectors()
	assert.Error(t, err)
	assert.Equal(t, []string(nil), s2)

}

func TestNewTask(t *testing.T) {
	task := NewTask(Payload{})
	assert.NotEmpty(t, task.ID)
	start, err := task.startTime()
	assert.NoError(t, err)
	assert.NotNil(t, start, "task start time is not nil")
	//t.Log(start)
}

func TestParseTestServer12345(t *testing.T) {
	viper.Set("DFK_FETCH", "127.0.0.1:8000")
	fetchServerAddr := viper.GetString("DFK_FETCH")
	fetchServerCfg := fetch.Config{
		Host:         fetchServerAddr,
		//ReadTimeout:  60 * time.Second,
		//WriteTimeout: 60 * time.Second,
	}
	viper.Set("SKIP_STORAGE_MW", true)
	fetchServer := fetch.Start(fetchServerCfg)
	defer fetchServer.Stop()

	paginateResults = false
	p := Payload{
		Name: "persons Table",
		Request: fetch.BaseFetcherRequest{
			URL: "http://127.0.0.1:12345/persons/page-0",
		},
		Fields: []Field{
			Field{
				Name:     "Names",
				Selector: "#cards a",
				Extractor: Extractor{
					Types: []string{"text"},//"const", "outerHtml"},
					//Params: map[string]interface{}{
					//	"value": "2",
					//},
				},
			},
			Field{
				Name:     "IDs",
				Selector: ".badge-primary",
				Extractor: Extractor{
					Types: []string{"html"},
				},
			},
			// Field{
			// 	Name:     "Count",
			// 	Selector: "td:nth-child(1)",
			// 	Extractor: Extractor{
			// 		Types: []string{"count", "unknown"},
			// 	},
			// },
		},
		PaginateResults: &paginateResults,
		Format:          "json",
	}
	expected := []byte(`[{"Count_count":10,"Header_const":"1","Header_outerHtml":"\u003ch1\u003ePersons\u003c/h1\u003e","Warning_html":"\u003cstrong\u003eWarning!\u003c/strong\u003e This is a demo website for web scraping purposes. \u003cbr/\u003eThe data on this page has been randomly generated."}]` + "\n")
	task := &Task{
		ID:      "12345",
		Payload: p,
		Visited: map[string]error{},
		Robots:  map[string]*robotstxt.RobotsData{},
	}
	r, err := task.Parse()
	assert.NoError(t, err)
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	got := buf.Bytes()
	t.Log(string(got))
	assert.Equal(t, expected, got)
	///// No selectors
	badP := Payload{
		Name: "No Selectors",
		Request: fetch.BaseFetcherRequest{
			URL: "http://127.0.0.1:12345",
		},
		PaginateResults: &paginateResults,
		Format:          "json",
	}
	task = &Task{
		ID:      "12345",
		Payload: badP,
		Visited: map[string]error{},
		Robots:  map[string]*robotstxt.RobotsData{},
	}
	r, err = task.Parse()
	assert.Error(t, err, "400: no parts found")
	//Bad output format
	badOF := Payload{
		Name: "No Selectors",
		Request: fetch.BaseFetcherRequest{
			URL: "http://127.0.0.1:12345",
		},
		Fields: []Field{
			Field{
				Name:     "P",
				Selector: "p",
				Extractor: Extractor{
					Types: []string{"text"},
				},
			},
		},
		PaginateResults: &paginateResults,
		Format:          "BadOutputFormat",
	}
	task = &Task{
		ID:      "12345",
		Payload: badOF,
		Visited: map[string]error{},
		Robots:  map[string]*robotstxt.RobotsData{},
	}
	r, err = task.Parse()
	assert.Error(t, err, "invalid output format specified")
}

func TestParseSwitchFetchers(t *testing.T) {
	viper.Set("DFK_FETCH", "127.0.0.1:8000")
	fetchServerAddr := viper.GetString("DFK_FETCH")
	fetchServerCfg := fetch.Config{
		Host:         fetchServerAddr,
		//ReadTimeout:  60 * time.Second,
		//WriteTimeout: 60 * time.Second,
	}
	viper.Set("SKIP_STORAGE_MW", true)
	fetchServer := fetch.Start(fetchServerCfg)
	defer fetchServer.Stop()

	paginateResults = false
	p := Payload{
		Name: "quotes",
		Request: fetch.BaseFetcherRequest{
			URL: "http://quotes.toscrape.com/js/",
		},
		Fields: []Field{
			Field{
				Name:     "quotes",
				Selector: ".text",
				Extractor: Extractor{
					Types: []string{"text"},
				},
			},
		},
		PaginateResults: &paginateResults,
		Format:          "json",
	}
	task := &Task{
		ID:      "12345",
		Payload: p,
		Visited: map[string]error{},
		Robots:  map[string]*robotstxt.RobotsData{},
	}
	r, err := task.Parse()
	assert.NoError(t, err)
	assert.NotNil(t, r)
}
