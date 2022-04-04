package gen

import (
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/collection"
	"github.com/zeromicro/go-zero/tools/goctl/model/sql/template"
	"github.com/zeromicro/go-zero/tools/goctl/util"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
	"github.com/zeromicro/go-zero/tools/goctl/util/stringx"
)

func genInsert(table Table, withCache, postgreSql bool) (string, string, error) {
	keySet := collection.NewSet()
	keyVariableSet := collection.NewSet()
	keySet.AddStr(table.PrimaryCacheKey.DataKeyExpression)
	keyVariableSet.AddStr(table.PrimaryCacheKey.KeyLeft)
	for _, key := range table.UniqueCacheKey {
		keySet.AddStr(key.DataKeyExpression)
		keyVariableSet.AddStr(key.KeyLeft)
	}

	expressions := make([]string, 0)
	expressionValues := make([]string, 0)
	expressionsAll := make([]string, 0)
	expressionAllValues := make([]string, 0)
	var count int
	flag := true
	for _, field := range table.Fields {
		camel := field.Name.ToCamel()
		if camel == "CreateTime" || camel == "UpdateTime" {
			flag = false
		}

		if field.Name.Source() == table.PrimaryKey.Name.Source() {
			if table.PrimaryKey.AutoIncrement {
				continue
			}
		}

		count += 1
		if postgreSql {
			if flag {
				expressions = append(expressions, fmt.Sprintf("$%d", count))
			}
			expressionsAll = append(expressionsAll, fmt.Sprintf("$%d", count))
		} else {
			if flag {
				expressions = append(expressions, "?")
			}
			expressionsAll = append(expressionsAll, "?")
		}
		if flag {
			expressionValues = append(expressionValues, "data."+camel)
		}
		expressionAllValues = append(expressionAllValues, "data."+camel)
	}

	camel := table.Name.ToCamel()
	text, err := pathx.LoadTemplate(category, insertTemplateFile, template.Insert)
	if err != nil {
		return "", "", err
	}

	output, err := util.With("insert").
		Parse(text).
		Execute(map[string]interface{}{
			"withCache":             withCache,
			"upperStartCamelObject": camel,
			"lowerStartCamelObject": stringx.From(camel).Untitle(),
			"expression":            strings.Join(expressions, ", "),
			"expressionValues":      strings.Join(expressionValues, ", "),
			"expressionsAll":        strings.Join(expressionsAll, ","),
			"expressionAllValues":   strings.Join(expressionAllValues, ","),
			"keys":                  strings.Join(keySet.KeysStr(), "\n"),
			"keyValues":             strings.Join(keyVariableSet.KeysStr(), ", "),
			"data":                  table,
		})
	if err != nil {
		return "", "", err
	}

	// interface method
	text, err = pathx.LoadTemplate(category, insertTemplateMethodFile, template.InsertMethod)
	if err != nil {
		return "", "", err
	}

	insertMethodOutput, err := util.With("insertMethod").Parse(text).Execute(map[string]interface{}{
		"upperStartCamelObject": camel,
		"data":                  table,
	})
	if err != nil {
		return "", "", err
	}

	return output.String(), insertMethodOutput.String(), nil
}
