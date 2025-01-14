package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mstgnz/sqlmapper"
	"github.com/mstgnz/sqlmapper/mysql"
	"github.com/mstgnz/sqlmapper/oracle"
	"github.com/mstgnz/sqlmapper/postgres"
	"github.com/mstgnz/sqlmapper/sqlite"
	"github.com/mstgnz/sqlmapper/sqlserver"
)

func main() {
	filePath := flag.String("file", "", "SQL dump dosyasının yolu")
	targetDB := flag.String("to", "", "Hedef veritabanı tipi (mysql, postgres, sqlite, oracle, sqlserver)")
	flag.Parse()

	if *filePath == "" || *targetDB == "" {
		fmt.Println("Kullanım: sqlmapper --file=<dosya_yolu> --to=<hedef_db>")
		fmt.Println("Örnek: sqlmapper --file=postgres.sql --to=mysql")
		flag.PrintDefaults()
		os.Exit(1)
	}

	content, err := os.ReadFile(*filePath)
	if err != nil {
		fmt.Printf("Dosya okuma hatası: %v\n", err)
		os.Exit(1)
	}

	sourceType := detectSourceType(string(content))
	if sourceType == "" {
		fmt.Println("Kaynak veritabanı tipi tespit edilemedi")
		os.Exit(1)
	}

	sourceParser := createParser(sourceType)
	if sourceParser == nil {
		fmt.Printf("Desteklenmeyen kaynak veritabanı tipi: %s\n", sourceType)
		os.Exit(1)
	}

	targetParser := createParser(*targetDB)
	if targetParser == nil {
		fmt.Printf("Desteklenmeyen hedef veritabanı tipi: %s\n", *targetDB)
		os.Exit(1)
	}

	schema, err := sourceParser.Parse(string(content))
	if err != nil {
		fmt.Printf("Parse hatası: %v\n", err)
		os.Exit(1)
	}

	result, err := targetParser.Generate(schema)
	if err != nil {
		fmt.Printf("SQL oluşturma hatası: %v\n", err)
		os.Exit(1)
	}

	outputPath := createOutputPath(*filePath, *targetDB)
	err = os.WriteFile(outputPath, []byte(result), 0644)
	if err != nil {
		fmt.Printf("Dosya yazma hatası: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Dönüşüm başarılı! Çıktı dosyası: %s\n", outputPath)
}

func detectSourceType(content string) string {
	content = strings.ToUpper(content)
	switch {
	case strings.Contains(content, "ENGINE=INNODB"):
		return "mysql"
	case strings.Contains(content, "AUTOINCREMENT"):
		return "sqlite"
	case strings.Contains(content, "IDENTITY"):
		return "sqlserver"
	case strings.Contains(content, "SERIAL"):
		return "postgres"
	case strings.Contains(content, "NUMBER("):
		return "oracle"
	default:
		return ""
	}
}

func createParser(dbType string) sqlmapper.Parser {
	switch strings.ToLower(dbType) {
	case "mysql":
		return mysql.NewMySQL()
	case "postgres":
		return postgres.NewPostgreSQL()
	case "sqlite":
		return sqlite.NewSQLite()
	case "oracle":
		return oracle.NewOracle()
	case "sqlserver":
		return sqlserver.NewSQLServer()
	default:
		return nil
	}
}

func createOutputPath(inputPath, targetDB string) string {
	dir := filepath.Dir(inputPath)
	filename := filepath.Base(inputPath)
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)
	return filepath.Join(dir, fmt.Sprintf("%s_%s%s", name, targetDB, ext))
}
