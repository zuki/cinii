package cinii

import (
	"encoding/xml"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// OpenSaerchEndpoint は、CiNii Books図書・雑誌書誌検索のOpenSearchのURI
const OpenSaerchEndpoint = "http://ci.nii.ac.jp/books/opensearch/search"

// AtomFeed はAtom1.0レスポンス構造体
type AtomFeed struct {
	XMLName      xml.Name   `xml:"http://www.w3.org/2005/Atom feed"`
	Title        string     `xml:"http://www.w3.org/2005/Atom title"`
	Links        []Link     `xml:"http://www.w3.org/2005/Atom link"`
	ID           string     `xml:"http://www.w3.org/2005/Atom id"`
	Updated      customTime `xml:"http://www.w3.org/2005/Atom updated"`
	TotalResults int        `xml:"http://a9.com/-/spec/opensearch/1.1/ totalResults"`
	StartIndex   int        `xml:"http://a9.com/-/spec/opensearch/1.1/ startIndex"`
	ItemsPerPage int        `xml:"http://a9.com/-/spec/opensearch/1.1/ itemsPerPage"`
	Entries      []Entry    `xml:"http://www.w3.org/2005/Atom entry"`
}

// Link はAtomFeed Linkフィールド構造体
type Link struct {
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
	Href string `xml:"href,attr"`
}

// HTMLLink はAtomFeedからHTML Linkを返すメソッド
func (f *AtomFeed) HTMLLink() (link string, err error) {
	link = html.UnescapeString(f.Links[0].Href)
	link, err = url.QueryUnescape(link)
	return
}

// Entry はAtomFeedのエントリ構造体
type Entry struct {
	Title      string    `xml:"http://www.w3.org/2005/Atom title"`
	ID         string    `xml:"http://www.w3.org/2005/Atom id"`
	Authors    []EAuthor `xml:"http://www.w3.org/2005/Atom author"`
	Publisher  string    `xml:"http://purl.org/dc/elements/1.1/ publisher"`
	PubDate    string    `xml:"http://prismstandard.org/namespaces/basic/2.0/ publicationDate"`
	IsPartOf   []Parent  `xml:"http://purl.org/dc/terms/ isPartOf"`
	HasPart    []string  `xml:"http://purl.org/dc/terms/ hasPart"`
	OwnerCount int       `xml:"http://ci.nii.ac.jp/ns/1.0/ ownerCount"`
}

// EAuthor はAtomFeed Authorフィールド構造体
type EAuthor struct {
	Name string `xml:"http://www.w3.org/2005/Atom name"`
}

// Parent はAtomFeed 親書誌フィールド構造体
type Parent struct {
	Title string `xml:"title,attr"`
	Link  string `xml:",chardata"`
}

type customTime struct {
	time.Time
}

func (c *customTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	d.DecodeElement(&v, &start)
	parse, err := time.Parse("2006-01-02T15:04:05-0700", v)
	// RFC3339: 2006-01-02T15:04:05-07:00
	//parse, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return err
	}
	*c = customTime{parse}
	return nil
}

// Search はCiniiBooksをOpenSearchで検索する
func Search(q url.Values) (*AtomFeed, error) {
	url := fmt.Sprintf("%s?%s", OpenSaerchEndpoint, q.Encode())
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	feed, err := ParseAtomFeed(body)
	if err != nil {
		return nil, err
	}
	return feed, nil
}

// ParseAtomFeed はAtomFeedを含むbyte[]を受け取りAtomFeed構造体のポインタで返す関数
func ParseAtomFeed(body []byte) (*AtomFeed, error) {
	// 取得したデータをXMLデコード
	feed := &AtomFeed{}
	err := xml.Unmarshal(body, feed)
	if err != nil {
		return nil, err
	}

	return feed, nil
}
