package obj

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestGet(t *testing.T) {
	jsonS := `{
	"str": "str",
	"int": 1,
	"float": 1.1,
	"bool": true,
	"list": [
		{"int": 2, "str": "str", "bool": true},
		1,
		1.1,
		"str"
	],
	"obj": {
		"str": "str",
		"int": 1,
		"float": 1.1,
		"bool": true
	}
}`
	var obj Obj
	err := json.Unmarshal([]byte(jsonS), &obj)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("obj.GetBool(\"bool\"): %v\n", obj.GetBool("bool"))
	fmt.Printf("obj.GetInt(\"int\"): %v\n", obj.GetInt("int"))
	fmt.Printf("obj.GetFloat(\"float\"): %v\n", obj.GetF64("float"))
	fmt.Printf("obj.GetString(\"str\"): %v\n", obj.GetStr("str"))
	fmt.Printf("obj.GetByPath(\"list.0.bool\"): %v\n", obj.GetBool("list.0.bool"))
	fmt.Printf("obj.GetByPath(\"list.0.int\"): %v\n", obj.GetInt("list.0.int"))
	fmt.Printf("obj.GetByPath(\"list.0.str\"): %v\n", obj.GetStr("list.0.str"))
	fmt.Printf("obj.GetByPath(\"list.1\"): %v\n", obj.GetInt("list.1"))
	fmt.Printf("obj.GetByPath(\"list.2\"): %v\n", obj.GetF64("list.2"))
	fmt.Printf("obj.GetByPath(\"list.3\"): %v\n", obj.GetStr("list.3"))
	fmt.Printf("obj.GetByPath(\"obj.str\"): %v\n", obj.GetStr("obj.str"))
	fmt.Printf("obj.GetByPath(\"obj.int\"): %v\n", obj.GetInt("obj.int"))
	fmt.Printf("obj.GetByPath(\"obj.float\"): %v\n", obj.GetF64("obj.float"))
	fmt.Printf("obj.GetByPath(\"obj.bool\"): %v\n", obj.GetBool("obj.bool"))
}

func TestSet(t *testing.T) {
	var obj = Obj{}
	obj.Set("a.b.c.d", 1)
	obj.Set("list", []any{1, 2, 3, 4})
	obj.Set("list.4", map[string]any{
		"str": "hello",
	})
	printObj(obj)
	obj.Set("list.4.str", "world")
	printObj(obj)
	if err := obj.Set("list.6", 5); err != nil {
		fmt.Printf("list.6 err: %v\n", err)
	}
	if err := obj.Set("list.key", 5); err != nil {
		fmt.Printf("list.key err: %v\n", err)
	}
	if err := obj.Set("list.-1", 5); err != nil {
		fmt.Printf("list.-1 err: %v\n", err)
	}
	if err := obj.Set("strLi", []string{}); err != nil {
		fmt.Printf("strLi err: %v\n", err)
	}
	if err := obj.Set("strMap", map[string]string{}); err != nil {
		fmt.Printf("strMap err: %v\n", err)
	}
	if err := obj.Set("a.b.c.d.e", 3); err != nil {
		fmt.Printf("a.b.c.d.e err: %v\n", err)
	}
}

func printObj(obj Obj) {
	bytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("JSON: %s\n", bytes)
}
