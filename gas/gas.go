package gas

import (
	_ "embed"
	"gopkg.in/yaml.v3"
)

type gasSpec = struct {
	Methods           map[string]uint64 `yaml:"methods" json:"methods"`
	DefaultPrice      uint64            `yaml:"default_price" json:"default_price"`
}

//go:embed load_gas.yml
var loadGasYaml []byte

var GasSpec = func() *gasSpec {
	spec := &gasSpec{}
	err := yaml.Unmarshal(loadGasYaml, spec)
	if err != nil {
		panic(err)
	}
	return spec
}()


func CountGas(req string) uint64 {
	spec := GasSpec
	var price uint64 = 0
	method := req
	val, ok := spec.Methods[method]
	if ok {
		price += val
	} else {
		price += spec.DefaultPrice
	}
	return price
}
