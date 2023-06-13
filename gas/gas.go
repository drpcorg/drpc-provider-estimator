package gas

import (
	_ "embed"
	"gopkg.in/yaml.v3"
)

type GasSpec = struct {
	Methods           map[string]uint64 `yaml:"methods" json:"methods"`
	DefaultPrice      uint64            `yaml:"default_price" json:"default_price"`
	ArchiveMultiplier uint64            `yaml:"archive_multiplier" json:"archive_multiplier"`
}

//go:embed load_gas.yml
var GAS_SPEC_RAW []byte
var GAS_SPEC *GasSpec

func ParseGasSpec() {
	GAS_SPEC = &GasSpec{}
	err := yaml.Unmarshal(GAS_SPEC_RAW, GAS_SPEC)
	if err != nil {
		panic(err)
	}
}

func GetGasSpec() *GasSpec {
	if GAS_SPEC == nil {
		ParseGasSpec()
	}
	return GAS_SPEC
}

func CountGas(req string) uint64 {
	spec := GetGasSpec()
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
