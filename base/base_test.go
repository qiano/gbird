package base

import (
	"fmt"
	"testing"
)

func TestMetadata(t *testing.T) {
	RegisterMetadata(&Base{})
	fmt.Println(Metadatas)
	fmt.Println(Metadata(&Base{}))
	fmt.Println(FieldMetadata(&Base{}, "IsDelete"))
	fmt.Println(GetTag(&Base{}, "IsDelete", "bson"))
}
