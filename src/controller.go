package src

import (
	"io/fs"
	"log"
	"os"
	"regexp"
	"time"

	"gorm.io/gorm"
)

var DB *gorm.DB
var versionMap map[string]SQLVersion

type SQLVersion struct {
	Version    string `gorm:"primaryKey"`
	Name       string
	Hash       string
	CreateTime time.Time
}

type MySQLVersionController struct {
}

func NewController(db *gorm.DB) *MySQLVersionController {
	DB = db
	return &MySQLVersionController{}
}

// Init table
func (c *MySQLVersionController) Init() {
	if !DB.Migrator().HasTable(&SQLVersion{}) {
		DB.Migrator().CreateTable(&SQLVersion{})
	}
}

func (c *MySQLVersionController) QueryAllVersion() {
	versions := []SQLVersion{}
	DB.Take(&versions)
	for _, v := range versions {
		versionMap[v.Version] = v
		log.Printf("%s_%s", v.Version, v.Name)
	}
}

// Check sql files' names are legal
// file name must has the format x.x.x_name.sql x is number
func (c *MySQLVersionController) QuerySqlFiles(path string) []fs.DirEntry {
	files, err := os.ReadDir(path)
	if err != nil {
		log.Fatalf("Read files failed, path is %s", path)
		return nil
	}
	r, _ := regexp.Compile(`^\d+\.\d+\.\d+_.*.sql$`)
	for _, v := range files {
		if v.IsDir() {
			continue
		}
		name := v.Name()
		match := r.MatchString(name)
		if !match {
			log.Fatalf("File name is ilegal: %s", name)
		}
	}
	return files
}

func (c *MySQLVersionController) CheckSqlFiles() {

}
