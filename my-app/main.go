package main

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var userStatus = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_request_get_user_status_count",
		Help: "Count of status returned by user.",
	},
	[]string{"user", "status"}, // labels
)

var averageResponseTime = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "http_request_average_response_time_seconds",
		Help: "Average response time for HTTP requests in seconds.",
	},
)

func init() {
	prometheus.MustRegister(userStatus)
	prometheus.MustRegister(averageResponseTime)
}

type MyRequest struct {
	User string
}

// the server will retrieve the user from the body, and randomly generate a status to return
func server(w http.ResponseWriter, r *http.Request) {
	var status string
	var user string

	// Record the start time
	startTime := time.Now()

	defer func() {
		// Calculate the response time in seconds
		responseTime := time.Since(startTime).Seconds()

		// Set the "average_response_time" metric to the calculated response time
		averageResponseTime.Set(responseTime)

		// Increment the "userStatus" counter metric
		userStatus.WithLabelValues(user, status).Inc()
	}()

	var mr MyRequest
	json.NewDecoder(r.Body).Decode(&mr)

	if rand.Float32() > 0.8 {
		status = "200"
	} else {
		status = "401"
	}

	user = mr.User
	log.Println(user, status)
	w.Write([]byte(status))
}

// the producer will randomly select a user from a pool of users and send it to the server
func producer() {
	userPool := []string{"Miras", "Meirkhan", "Alkey", "Meirkhan2.0", "Afton"}
	for {
		postBody, _ := json.Marshal(MyRequest{
			User: userPool[rand.Intn(len(userPool))],
		})
		requestBody := bytes.NewBuffer(postBody)
		http.Post("http://localhost:8081", "application/json", requestBody)
		time.Sleep(time.Second * 2)
	}
}

func main() {
	// the producer runs on its own goroutine
	go producer()

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", server)

	http.ListenAndServe(":8081", nil)
}
