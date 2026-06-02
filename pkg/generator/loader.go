package generator

import (
	"os"

	"gopkg.in/yaml.v3"
)

type GeneratorsConfig map[string]map[string]ColumnConfig

func LoadGeneratorsConfig(path string) (GeneratorsConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config GeneratorsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config, nil
}

func GeneratorsConfigToTableConfigs(config GeneratorsConfig) map[string]*TableConfig {
	result := make(map[string]*TableConfig)
	for tableName, columns := range config {
		tc := &TableConfig{
			TableName: tableName,
			Columns:   make(map[string]ColumnOverride),
		}
		for colName, colCfg := range columns {
			if colCfg.Disabled {
				tc.Columns[colName] = ColumnOverride{Disabled: true}
			} else {
				g, err := resolveBuiltinGenerator(colCfg.Generator, colCfg.Params)
				if err != nil {
					continue
				}
				tc.Columns[colName] = ColumnOverride{Generator: g}
			}
		}
		result[tableName] = tc
	}
	return result
}
