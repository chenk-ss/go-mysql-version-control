# go-mysql-version-control

# This a project to manage sql files. 
Notice: The v0.0.1 is based on gorm and MySQL
## How to use it?
### step1.
>go get github.com/chenk-ss/go-mysql-version-control

### step2.
```golang
import 	sqlVersionController "github.com/chenk-ss/go-mysql-version-control"

```
### step3.
```golang
sqlVersionController.NewController(&dbMySQL).Start("sql/")
```

The type of dbMySQL is *sql.DB, and the parameter of func Start() is your sql files' relative path from your module.