package main

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

var (
	// local环境
	dbUri = "root:mystic@tcp(127.0.0.1:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local"

	// 需要生成结构体的表
	tables = []string{
		"t_greeter",
	}
	genTable = map[string][]string{
		"test": tables,
	}
	genOutPath = map[string]string{
		"test": "./internal/data/",
	}
)

func getDBUri(db string) string {
	return fmt.Sprintf(dbUri, db)
}

func main() {
	for dbName, tables := range genTable {
		db, err := gorm.Open(mysql.Open(getDBUri(dbName)))
		if err != nil {
			panic(err)
		}
		g := gen.NewGenerator(gen.Config{
			OutPath:        genOutPath[dbName] + "/dao",
			ModelPkgPath:   genOutPath[dbName] + "/model",
			Mode:           gen.WithDefaultQuery | gen.WithQueryInterface,
			FieldSignable:  true, // 生成无符号类型
			FieldNullable:  true, // 当字段为空时，生成指针
			FieldCoverable: true, // 字段有默认值生成指针，解决无法赋值为零的问题
		})
		g.UseDB(db)
		for _, table := range tables {
			g.ApplyBasic(g.GenerateModel(table))
		}
		g.Execute()
	}
}
