package pro_test

import (
	"fmt"
	"github.com/fatih/structtag"
)

func main() {
	tag := `json:"foo,omitempty,string"xml:"foo"`
	tags, err := structtag.Parse(string(tag))
	if err != nil {
		panic(err)
	}
	for _, t := range tags.Tags() {
		fmt.Printf("%+v\n", t)
	}
	fmt.Println(tags.Len())

	jsonTag, err := tags.Get("json")
	if err != nil {
		panic(err)
	}

	jsonTag.Name = "gar"
	jsonTag.Options = []string{"kai", "hui"}
	tags.Set(jsonTag)

	fmt.Println(tags)

	tags.Set(&structtag.Tag{
		Key:     "htc",
		Name:    "3620",
		Options: []string{"kaishou"},
	})
	fmt.Println(tags)
}
