package main

import (
	"log"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
    "github.com/patrickmn/go-cache"
    "time"
    "github.com/gorilla/mux"
    "net/http"
	"os"
    "strings"
	"context"
	"net/http/httputil"
)



func main() {
	// database configs
	//dbUser, dbPassword, dbName := "root", "root", "project"
	
	db, err := sql.Open("mysql", "root:root/project")

	// unable to connect to database
	if err != nil {
		log.Fatalln(err)
	}
	
	defer db.Close()
	
	reader := kafkasubscriber()
	if reader != nil {
		//reload cache from sql DB
		rows, _ := db.Query(`SELECT id, content FROM project_table`)
		 ok := SetCache("key_1",rows)
		 fmt.Println(ok)
	}
	
	//handling APIs
	log.Println("caching service !!")
	
	router := mux.NewRouter()
    startServer(router)
    router.HandleFunc("/api/set", AddContext(setToCache)).Methods("POST")
    router.HandleFunc("/api/get/{id}", AddContext(getFromCache)).Methods("GET")
    startServer(router)

    select {}
	
}



func startServer(r *mux.Router){

        go func() {
              err := http.ListenAndServe(fmt.Sprintf(":%d", 8090), r)
              fmt.Println("8080")
              if err != nil {
                   fmt.Println("", "Starting Server Failed for %v", err)
                    os.Exit(1)
                  }
              }()
}



var Cache = cache.New(5*time.Minute, 5*time.Minute)
func SetCache(key string, data interface{}) bool {
    Cache.Set(key, data, cache.NoExpiration)
    return true
}



func GetCache(key string) (project_table, error) {
        data, _ := Cache.Get(key)
        fmt.Println("data :%v", data)
		
		dataDB := data.(project_table)
		return dataDB, nil
}


func setToCache(w http.ResponseWriter, request *http.Request) {

        fmt.Println("body in request:%v", request.Body)
		log.Println("user Id :%v",  request.URL.Path[strings.LastIndex(request.URL.Path[1:], "/")+2:])
        
		//
		db, err := sql.Open("mysql", "root:root/project")
		rows, err := db.Query(`SELECT id, content FROM project_table`)
		if err != nil {
			panic(err)
		}
		defer rows.Close()
		for rows.Next() {
			project_table := project_table{}
			err = rows.Scan(&project_table.Id, &project_table.Content)
			if err != nil {
				panic(err)
			}
			fmt.Println(project_table)
		}
		err = rows.Err()
		if err != nil {
			panic(err)
		}
		

        ok := SetCache("key_1",rows)
        fmt.Println(ok)
        w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "successfully! data has been saved in cache")

}

type project_table struct {
    Id int
    Content string
}


func getFromCache(w http.ResponseWriter, request *http.Request){
        fmt.Println("request is :%v", request)
        request.ParseForm()
        MyParam, ok  := request.Form["offSet"]
        if ok {
                fmt.Println(MyParam)
        }
        //to do pagination
        dataC, errFind := GetCache("key_1")
		if errFind != nil {
			fmt.Println("cache miss")
			_ = SetCache("key_1",dataC)
		}else {
			fmt.Println("cache hit")
			w.WriteHeader(http.StatusOK)
		}
}


func AddContext(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := "123"
		contextVal := "123"
		reqDump, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Println(contextVal, "httputil.DumpRequest Failed:  %v", err)
		}
		log.Println(contextVal, " Request RECVD --->  %v", string(reqDump))
		ctx := context.WithValue(r.Context(), key, contextVal)
		next.ServeHTTP(w, r.WithContext(ctx))

		
		log.Println(contextVal, " Response SENT <---  %s %s ", r.Method, r.URL.String())
		
	})
}


//kafka subscriber

//kafka configs









