package mockx

import (
	"encoding/json"
	"fmt"
)

func Print(i interface{}) {
	b, err := json.MarshalIndent(i, "", "\t")
	if err != nil {
		fmt.Println("failed to marshal element")
		return
	}
	fmt.Printf("%s \n", string(b))
}
