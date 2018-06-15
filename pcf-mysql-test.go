package main 

import (
  "net/http"
  "os"
  "fmt"
  "encoding/json"
  "database/sql"
  "strconv"
  _ "github.com/go-sql-driver/mysql"
  "time"
)

var (
  Hostname string 
  User string 
  Password string 
  Database string 
  PingInterval int64
  Port = "3306"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("<html>hello</html>"))  
}

// Try to get app details from VCAP_APPLICATION.  If that fails then default to http://127.0.0.1:PORT
func setAppDetails() {
  env := os.Getenv("VCAP_SERVICES")
  if env == "" {
    fmt.Println("VCAP_SERVICES not found")
    os.Exit(2)
  }
  type Credentials struct {
          URI string `json:"uri"`
          Hostname string `:json:"hostname"`
          Port int `json:"port"`
          Database string `json:"name"`
          User string `json:"username"`
          Password string `json:"password"`
  }
  
  type SqlObjects struct {
    Creds Credentials `json:"credentials"`
  }
  type VCAP_SERVICE struct {
          Databases []SqlObjects `json:"p-mysql"`
  }
  
  MyDB := new(VCAP_SERVICE)
  err := json.Unmarshal([]byte(env), &MyDB)
  if err != nil {
          fmt.Printf("Failed to decode VCAP_SERVICE for database credentials: %s\n", err)
          os.Exit(3)
  }

  if len(MyDB.Databases) > 0  {
    Hostname = MyDB.Databases[0].Creds.Hostname
    User = MyDB.Databases[0].Creds.User
    Password = MyDB.Databases[0].Creds.Password
    Database = MyDB.Databases[0].Creds.Database 
    Port = fmt.Sprintf("%d", MyDB.Databases[0].Creds.Port)
  }
}

func overrideWithEnv() {
  if os.Getenv("HOSTNAME") != "" {
    Hostname = os.Getenv("HOSTNAME") 
  }
  if os.Getenv("SQLUSERNAME") != "" {
    User = os.Getenv("SQLUSERNAME") 
  }
  if os.Getenv("SQLPASSWORD") != "" {
    Password = os.Getenv("SQLPASSWORD") 
  }
  if os.Getenv("DATABASE") != "" {
    Database = os.Getenv("DATABASE") 
  }
  if os.Getenv("INTERVAL") != "" {
    Interval, _ := strconv.Atoi(os.Getenv("INTERVAL"))
    PingInterval = int64(Interval)
  } else {
    PingInterval = 315 // default 315 seconds
  }
}

func createTables(sess *sql.DB) {
	_, err := sess.Exec("CREATE TABLE IF NOT EXISTS t1 (a TEXT, b INT)")
	if err != nil {
		fmt.Printf("Failed to create table t1: %sn", err)
	}
}

func writeData(sess *sql.DB) {
	_, err := sess.Exec("INSERT INTO t1 VALUES ('test data', 1)")
	if err != nil {
		fmt.Printf("Failed to insert data into table t1: %s\n", err)
	}
}


func ConnectToDatabase() {
  //DBURL := fmt.Sprintf("mysql://%s:%s@%s:%s/%s?reconnect=true", User, Password, Hostname, Port, Database)
  DBURL := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", User, Password, Hostname, Port, Database)
  fmt.Printf("using uri: %s\n", DBURL)
  sess, err := sql.Open("mysql", DBURL)
  if err != nil {
    fmt.Printf("ERROR: Failed to connecto to database using:%s: %s\n", DBURL, err)
    os.Exit(1)
  }
  sess.SetMaxOpenConns(1) // make sure this is only ever one connection open to eliminate variables
  
  for {
    createTables(sess)
    writeData(sess)
    err = sess.Ping()
    if err != nil {
      fmt.Printf("Error pinging mysql server: %s\n", err)
    }
    rows, err := sess.Query("show full processlist")
    if err != nil {
      fmt.Printf("failed to query database: %s\n", err)
    } else {
      fmt.Println("Ping Successful")
      for rows.Next() {
        var id int 
        var user,host,db,command,state,info string  
        var t []uint8
        var progress float64
        err = rows.Scan(&id,&user,&host,&db,&command,&t,&state,&info,&progress)
        if err != nil {
          fmt.Printf("Failed to parse query results: %s\n", err)
        } else {
          fmt.Printf("%d | %s | %s | %s | %s | %s | %s | %f\n", id, user, host, db, command, state, info, progress)
        }
      }
    }
    time.Sleep(time.Duration(PingInterval * int64(time.Second) ))
  }
  
}

func main() {
  setAppDetails()
  overrideWithEnv()
  go ConnectToDatabase()
  
  fmt.Printf("Hostname: %s\nUser: %s\nPassword: %s\nDatabase: %s\n", Hostname, User, Password, Database)
  // start http services 
  http.HandleFunc("/", rootHandler)
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		fmt.Printf("Failed to start http server: %s\n", err)
	}
}
