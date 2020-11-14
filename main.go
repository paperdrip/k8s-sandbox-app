package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var redisHost string
var redisPort string
var redisPassword string
var configFile = "config.json"

func main() {
	var router = newServer()
	log.Println("Starting up server and listening on port 3000")
	log.Fatal(http.ListenAndServe(":3000", handlers.CORS(handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"}),
		handlers.AllowedOrigins([]string{"*"}))(router)))
}

func newServer() http.Handler {
	r := mux.NewRouter().StrictSlash(true)
	r.Use(commondMiddleware)
	r.HandleFunc("/api", handlePost).Methods("POST")
	r.HandleFunc("/api", handleQuery).Methods("GET")
	return r
}

// Configuration is a struct that represent the config file
type Configuration struct {
	RedisHost string
	RedisPort string
}

// PostData is a struct representing the data sent in request
type PostData struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

func setEnv() {
	file, err := os.Open(configFile)
	if err != nil {
		log.Fatal(err)
	}
	decoder := json.NewDecoder(file)
	var configuration Configuration
	err = decoder.Decode(&configuration)
	if err != nil {
		log.Fatal(err)
	}
	if redisHost = os.Getenv("REDIS_HOST"); redisHost == "" {
		if redisHost = configuration.RedisHost; redisHost == "" {
			redisHost = "localhost"
		}
	}

	if redisPort = os.Getenv("REDIS_PORT"); redisPort == "" {
		if redisPort = configuration.RedisPort; redisPort == "" {
			redisPort = "6379"
		}
	}

	redisPassword = os.Getenv("REDIS_PASSWORD")
}

func handlePost(w http.ResponseWriter, r *http.Request) {

}

func handleQuery(w http.ResponseWriter, r *http.Request) {
	pool := newPool(false)
	conn := pool.Get()
	defer conn.Close()
	content, _ := get(conn)
	json.NewEncoder(w).Encode(content)
}

func newPool(write bool) *redis.Pool {
	setEnv() // Creating a separate function for initializing the connection to redis, instead of coupling with the main function
	return &redis.Pool{
		// Maximum idle connections in the pool
		MaxIdle: 80,

		// Maximum number of connections
		MaxActive: 12000,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", redisHost+":"+redisPort)
			if err != nil {
				log.Println("cannot reach redis server", err)
			}
			_, err = c.Do("AUTH", redisPassword)
			if err != nil {
				log.Println("cannot authenticate to redis server", err)
			}
			return c, err
		},
	}
}

func get(c redis.Conn) ([]PostData, error) {
	results := []PostData{}
	_, err := c.Do("AUTH", redisPassword)
	if err != nil {
		log.Println("cannot authenticate to redis", err)
	}
	// Keys are in string array so we use redis.Strings
	keys, err := redis.Strings(c.Do("KEYS", "*"))
	if err != nil {
		return results, err
	}
	for _, k := range keys {
		if v, err := redis.String(c.Do("GET", k)); err == nil {
			results = append(results, PostData{k, v})
		}
	}
	return results, nil
}

func set(c redis.Conn, key string, value string) error {
	_, err := c.Do("AUTH", redisPassword)
	if err != nil {
		log.Println("cannot authenticate to redis", err)
	}
	_, err = c.DO("SET", key, value)
	if err != nil {
		return err
	}
	return nil
}

func commondMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
