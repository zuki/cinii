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

// Title はレコードから[タイトル, 読み]を返すメソッド
func (r *Record) Title() (ret []string) {
	ret = make([]string, 2)
	for _, title := range r.Descriptions[0].Title {
		if len(title.Lang) > 0 {
			ret[1] = title.Text
		} else {
			ret[0] = title.Text
		}
	}
	return
}

// Parents はレコードから[親書誌タイトル, NCID]の配列を返すメソッド
func (r *Record) Parents() (ret [][]string, ok bool) {
	fields := r.Descriptions[0].IsPartOf
	if len(fields) == 0 {
		return nil, false
	}
	ret = make([][]string, len(fields))
	for i, field := range fields {
		id := field.Resource
		id = strings.Replace(id, "http://ci.nii.ac.jp/ncid/", "", 1)
		id = strings.Replace(id, "#entity", "", 1)
		ret[i] = []string{field.Title, id}
	}
	return ret, true
}

// Volumes はレコードから[巻号等, ISNB]の配列を返すメソッド
func (r *Record) Volumes() (ret [][]string, ok bool) {
	fields := r.Descriptions[0].HasPart
	if len(fields) == 0 {
		return nil, false
	}
	ret = make([][]string, len(fields))
	for i, field := range fields {
		id := field.Resource
		id = strings.Replace(id, "urn:isbn:", "", 1)
		ret[i] = []string{field.Title, id}
	}
	return ret, true
}

// Topics はレコードからTopicの配列を返すメソッド
func (r *Record) Topics() (ret []string, ok bool) {
	fields := r.Descriptions[0].Topics
	if len(fields) == 0 {
		return nil, false
	}
	ret = make([]string, len(fields))
	for i, field := range fields {
		ret[i] = field.Title
	}
	return ret, true
}

// Authors はレコードから[著者名, 読み, ALID]の配列を返すメソッド
func (r *Record) Authors() (ret [][]string, ok bool) {
	// 書誌情報だけで著者情報はなし
	if len(r.Descriptions) == 1 {
		return nil, false
	}

	var fields []Author
	for _, description := range r.Descriptions {
		if len(description.Authors) == 0 {
			continue
		} else {
			fields = description.Authors
		}
	}
	// 書誌情報と所蔵情報のみで著者情報はなし
	if len(fields) == 0 {
		return nil, false
	}

	ret = make([][]string, len(fields))
	for i, field := range fields {
		id := field.Author.About
		id = strings.Replace(id, "http://ci.nii.ac.jp/author/", "", 1)
		id = strings.Replace(id, "#entity", "", 1)

		var author, yomi string
		for _, name := range field.Author.Name {
			if len(name.Lang) > 0 {
				yomi = name.Text
			} else {
				author = name.Text
			}
		}
		ret[i] = []string{author, yomi, id}
	}
	return ret, true
}

// Holdings はレコードから[所蔵館名, FAID, 所蔵館OPACURL]の配列を返すメソッド
func (r *Record) Holdings() (ret [][]string, ok bool) {
	// 書誌情報だけで所蔵館情報はなし
	if len(r.Descriptions) == 1 {
		return nil, false
	}

	var fields []Holding
	for _, description := range r.Descriptions {
		if len(description.Holdings) == 0 {
			continue
		} else {
			fields = description.Holdings
		}
	}
	// 書誌情報と著者情報のみで所蔵館情報はなし
	if len(fields) == 0 {
		return nil, false
	}

	ret = make([][]string, len(fields))
	for i, field := range fields {
		holding := field.Holding
		id := holding.About
		id = strings.Replace(id, "http://ci.nii.ac.jp/library/", "", 1)
		ret[i] = []string{holding.Name[0].Text, id, holding.SeeAlso.Resource}
	}
	return ret, true
}

// Get はレコードIDを受け取り、情報をRecord構造体のポインタで返す関数
func Get(url string, appid string) (*Record, error) {
	if !strings.HasPrefix(url, RetrieveEndopoint) {
		url = fmt.Sprintf("%s/%s", RetrieveEndopoint, url)
	}
	if !strings.HasSuffix(url, ".rdf") {
		url += ".rdf"
	}

	if len(appid) > 0 {
		url = fmt.Sprintf("%s?appid=%s", url, appid)
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
