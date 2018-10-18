package main

import (
	"os"

	"github.com/BurntSushi/toml"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	var tomlData = `
		[feature1]
		enable = true
		[feature2]
		enable = false`

	type feature struct {
		Enable bool
	}

	type tomlConfig struct {
		Title string
		F1    feature `toml:"feature1"`
		F2    feature `toml:"feature2"`
	}

	var conf tomlConfig
	if _, err := toml.Decode(tomlData, &conf); err != nil {
		panic(err)
	}

	f, err := os.Create("testfile")
	check(err)

	e := toml.NewEncoder(f)
	err = e.Encode(conf)
	check(err)

}
