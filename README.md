# github/zuki/cinii

CiNii Books APIを実装したパッケージ

## インストール

```sh
go get github/zuki/cinii
```

## 利用方法

- NCIDを指定してタイトルとタイトルの読みを出力

```go
package main

import (
	"fmt"
	"github.com/zuki/cinii"
)

func main() {
  record, err := cinii.Get("BB19132110", "Your CiNii appid")
  if err == nil {
    title := record.Title()
    fmt.Printf("%s (%s)\n", title[0], title[1])
  }
}
```

- OpenSearchを"Go言語"で検索してタイトルと著者を表示

```go
package main

import (
	"fmt"
  "net/url"
	"github.com/zuki/cinii"
)

func main() {
  appid := "Your CiNii appid"

  q := url.Values{}
	q.Set("q", "Go言語")
	q.Set("appid", appid)

  feed, err := cinii.Search(q)
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range feed.Entries {
		record, err := cinii.Get(entry.ID, appid)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("タイトル: %s\n", record.Title()[0])
    if authors, ok := record.Authors(); ok {
      for i, author := range authors {
        if i == 0 {
          fmt.Print("著者: ")
        } else {
          fmt.Print("; ")
        }
        fmt.Print(author[0])
      }
      fmt.Println("")
    }
    fmt.Println("")
	}
}
```

```
## 著作権

MIT
