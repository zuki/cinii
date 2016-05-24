package cinii

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// RetrieveEndopoint は、RDF形式のCiNii Bookレコードを書誌IDで取得するためのURI
const RetrieveEndopoint = "http://ci.nii.ac.jp/ncid"

// Record はRDFデータ用構造体
type Record struct {
	XMLName      xml.Name      `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# RDF"`
	Descriptions []Description `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# Description"`
}

// Description はコンテナ構造体
type Description struct {
	AboutAttr
	Type             ResourceAttr    `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# type"`
	IsPrimaryTopicOf ResourceAttr    `xml:"http://xmlns.com/foaf/0.1/ isPrimaryTopicOf"`
	Title            TextFields      `xml:"http://purl.org/dc/elements/1.1/ title"`
	Alternative      []string        `xml:"http://purl.org/dc/terms/ alternative"`
	Creator          string          `xml:"http://purl.org/dc/elements/1.1/ creator"`
	Publisher        []string        `xml:"http://purl.org/dc/elements/1.1/ publisher"`
	Language         string          `xml:"http://purl.org/dc/elements/1.1/ language"`
	Date             string          `xml:"http://purl.org/dc/elements/1.1/ date"`
	Topics           []ResourceField `xml:"http://xmlns.com/foaf/0.1/ topic"`
	NCID             string          `xml:"http://ci.nii.ac.jp/ns/1.0/ ncid"`
	Edition          string          `xml:"http://prismstandard.org/namespaces/basic/2.0/ edition"`
	IsPartOf         []ResourceField `xml:"http://purl.org/dc/terms/ isPartOf"`
	HasPart          []ResourceField `xml:"http://purl.org/dc/terms/ hasPart"`
	ContentOfWorks   []string        `xml:"http://ci.nii.ac.jp/ns/1.0/ contentOfWorks"`
	Medium           TitleAttr       `xml:"http://purl.org/dc/terms/ medium"`
	OwnerCount       int             `xml:"http://ci.nii.ac.jp/ns/1.0/ ownerCount"`
	LCCN             []int           `xml:"http://purl.org/ontology/bibo/ lccn"`
	SeeAlso          []ResourceAttr  `xml:"http://www.w3.org/2000/01/rdf-schema# seeAlso"`
	Authors          []Author        `xml:"http://xmlns.com/foaf/0.1/ maker"`
	Holdings         []Holding       `xml:"http://purl.org/ontology/bibo/ owner"`
}

// AboutAttr はabout sttribute構造体
type AboutAttr struct {
	About string `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# about,attr"`
}

// ResourceAttr はresource sttribute構造体
type ResourceAttr struct {
	Resource string `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# resource,attr"`
}

// TitleAttr はtitle attribute構造体
type TitleAttr struct {
	Title string `xml:"http://purl.org/dc/elements/1.1/ title,attr"`
}

// ResourceField はresource構造体
type ResourceField struct {
	ResourceAttr
	TitleAttr
}

// NameField はIDを持つ名前の構造体
type NameField struct {
	AboutAttr
	Name    TextFields   `xml:"http://xmlns.com/foaf/0.1/ name"`
	SeeAlso ResourceAttr `xml:"http://www.w3.org/2000/01/rdf-schema# seeAlso"`
}

// Stringerインターフェースの実装
func (n NameField) String() string {
	str := n.Name[0].Text
	if len(n.Name) > 1 {
		str += fmt.Sprintf(" (%s)", n.Name[1].Text)
	}
	if about := n.About; len(about) > 0 {
		about = strings.Replace(about, "http://ci.nii.ac.jp/author/", "", 1)
		about = strings.Replace(about, "http://ci.nii.ac.jp/library/", "", 1)
		about = strings.Replace(about, "#entity", "", 1)
		str += fmt.Sprintf(" [%s]", about)
	}
	/* 所蔵館におけるこの書誌のURI
	if sa := n.SeeAlso.Resource; len(sa) > 0 {
		str += fmt.Sprintf(" -> %s", sa)
	}
	*/
	return str
}

// TextField はよみを持つテキストフィールドの構造体
type TextField struct {
	Lang string `xml:"lang,attr"`
	Text string `xml:",chardata"`
}

// Author はmaker構造体
type Author struct {
	Author NameField `xml:"http://xmlns.com/foaf/0.1/ Person"`
}

// Holding は組織情報構造体
type Holding struct {
	Holding NameField `xml:"http://xmlns.com/foaf/0.1/ Organization"`
}

// TextFields は []TextFieldの別名
type TextFields []TextField

// Stringerインターフェースの実装
func (t TextFields) String() string {
	str := t[0].Text
	if len(t) > 1 {
		str += fmt.Sprintf(" (%s)", t[1].Text)
	}
	return str
}

// Get はレコードIDを受け取り、情報をRecord構造体のポインタで返す関数
func Get(url string, appid string) (*Record, error) {
	if !strings.HasPrefix(url, RetrieveEndopoint) {
		url = fmt.Sprintf("%s/%s?appid=%s", RetrieveEndopoint, url, appid)
	}
	if !strings.HasSuffix(url, ".rdf") {
		url += ".rdf"
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	record, err := Parse(body)
	if err != nil {
		return nil, err
	}

	return record, nil
}

// Parse はRecord情報を含むbyte[]を受け取りRecord構造体のポインタで返す関数
func Parse(body []byte) (*Record, error) {
	// 取得したデータをXMLデコード
	record := &Record{}
	err := xml.Unmarshal(body, record)
	if err != nil {
		return nil, err
	}

	return record, nil
}
