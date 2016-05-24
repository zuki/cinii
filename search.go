package cinii

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

// OpenSaerchEndpoint は、CiNii Books図書・雑誌書誌検索のOpenSearchのURI
const OpenSaerchEndpoint = "http://ci.nii.ac.jp/books/opensearch/search"

// AtomFeed はAtom1.0レスポンス構造体
type AtomFeed struct {
	XMLName      xml.Name `xml:"http://www.w3.org/2005/Atom feed"`
	TotalResults int      `xml:"http://a9.com/-/spec/opensearch/1.1/ totalResults"`
	StartIndex   int      `xml:"http://a9.com/-/spec/opensearch/1.1/ startIndex"`
	ItemsPerPage int      `xml:"http://a9.com/-/spec/opensearch/1.1/ itemsPerPage"`
	Entries      []Entry  `xml:"http://www.w3.org/2005/Atom entry"`
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

// Search はCiniiBooksをOpenSearchで検索する
func Search(q url.Values) (*AtomFeed, error) {
	if appid := q.Get("appid"); len(appid) == 0 {
		appid, err := GetAppID()
		if err != nil {
			return nil, err
		}
		q.Set("appid", appid)
	}
	url := fmt.Sprintf("%s?%s", OpenSaerchEndpoint, q.Encode())
	// URLを叩いてデータを取得
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

// GetAppID は環境変数 "CINII_APPID" からCiNii APIを利用するためのappidを取得
func GetAppID() (string, error) {
	appid := os.Getenv("CINII_APPID")
	if len(appid) == 0 {
		err := errors.New("Set your CiNii appid to CINII_APPID envrionmental variable")
		return "", err
	}
	return appid, nil
}
