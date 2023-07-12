package version_control

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

var DB *gorm.DB
var versionMap = make(map[string]SQLVersion)
var TABLE_NAME = "sql_version"
var MAX_VERSION = "0.0.0"
var FILES_EXECUTED = []SQLVersion{}
var FILES_IN_PATH = []SQLVersion{}
var FILES_NOT_EXECUTED = []SQLVersion{}
var INIT_SQL = "CREATE TABLE `sql_version`  (" +
	"`version` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL," +
	"`name` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL," +
	"`hash` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL," +
	"`create_time` datetime NULL DEFAULT CURRENT_TIMESTAMP," +
	"PRIMARY KEY (`version`) USING BTREE" +
	") ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;"

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
func (c *MySQLVersionController) Start(path string) {
	if err := DB.Exec("SELECT * FROM `sql_version` LIMIT 1").Error; err != nil {
		exec(SQLVersion{
			Version:    "0.0.0",
			Name:       "init",
			Hash:       INIT_SQL,
			CreateTime: time.Now(),
		})
	}
	c.QueryAllVersion()
	c.ReadFilesInDisk(path)
	c.CheckSqlFiles()
	c.ExecuteSqlFiles()
}

// Query all executed versions
func (c *MySQLVersionController) QueryAllVersion() {
	versions := []SQLVersion{}
	DB.Table(TABLE_NAME).Find(&versions)
	sort.Slice(versions, func(i, j int) bool {
		return VersionLastThan(versions[j].Version, versions[i].Version)
	})
	for _, v := range versions {
		versionMap[v.Version] = v
		FILES_EXECUTED = append(FILES_EXECUTED, v)
	}
	MAX_VERSION = versions[len(versions)-1].Version
	log.Printf("The max version of executed sql is %s", MAX_VERSION)
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

func (c *MySQLVersionController) ReadFilesInDisk(path string) {
	files := c.QuerySqlFiles(path)
	for _, v := range files {
		bytes, err := os.ReadFile(path + v.Name())
		if err != nil {
			log.Fatalf("Read file failed, file is %s", path+v.Name())
		}
		version := strings.Split(v.Name(), "_")[0]
		FILES_IN_PATH = append(FILES_IN_PATH, SQLVersion{
			Version:    version,
			Name:       v.Name()[len(version)+1 : len(v.Name())-4],
			Hash:       string(bytes),
			CreateTime: time.Now(),
		})
	}
}

func (c *MySQLVersionController) CheckSqlFiles() {
	log.Println("Check files..........")
	for _, v := range FILES_IN_PATH {
		if val, ok := versionMap[v.Version]; ok {
			if val.Name != v.Name || val.Hash != Hash(v.Hash) {
				log.Fatalf("Check file failed, file is %s", v.Name)
			}
		} else {
			if VersionLastThan(v.Version, MAX_VERSION) {
				FILES_NOT_EXECUTED = append(FILES_NOT_EXECUTED, v)
			} else {
				log.Fatalf("Sql file version is outmoded. The file is %s", v.Name)
			}
		}
	}
	log.Println("Check files end..........")
}

func (c *MySQLVersionController) ExecuteSqlFiles() {
	log.Println("Executing files..........")
	log.Printf("Size is %d", len(FILES_NOT_EXECUTED))
	sort.Slice(FILES_NOT_EXECUTED, func(i, j int) bool {
		return VersionLastThan(FILES_NOT_EXECUTED[j].Version, FILES_NOT_EXECUTED[i].Version)
	})
	for _, v := range FILES_NOT_EXECUTED {
		log.Println(v.Version + v.Name)
		exec(v)
	}
	log.Println("Executing files end..........")
}

func exec(version SQLVersion) {
	// Note the use of tx as the database handle once you are within a transaction
	tx := DB.Begin()
	if err := tx.Error; err != nil {
		log.Panic("Execute sql error1")
	}
	if err := tx.Exec(version.Hash).Error; err != nil {
		tx.Rollback()
		log.Panic("Execute sql error2")
	}
	version.Hash = Hash(version.Hash)
	if err := tx.Table(TABLE_NAME).Create(version).Error; err != nil {
		tx.Rollback()
		log.Panic("Execute sql error3")
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		log.Panic("Execute sql error4")
	}
}

func VersionLastThan(a, b string) bool {
	if a == b {
		return false
	}
	lista, listb := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < 3; i++ {
		if lista[i] < listb[i] {
			return false
		}
	}
	return true
}

func Hash(val string) string {
	hash := sha256.Sum256([]byte(val))
	hashString := hex.EncodeToString(hash[:])
	return hashString
}
