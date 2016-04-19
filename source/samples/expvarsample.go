package main

import (
    "encoding/json"
    "expvar"
    "fmt"
    "net/http"
    "sync"
    "time"
)

// Stats is used to collect runtime metrics
type Stats struct {
    sync.Mutex
    TotalHit  int
    ErrorNums int
}

func (s *Stats) IncreaseTotalHit(i int) {
    s.Lock()
    defer s.Unlock()

    s.TotalHit += i
}

func (s *Stats) IncreaseErrorNums(i int) {
    s.Lock()
    defer s.Unlock()

    s.ErrorNums += i
}

func (s *Stats) String() string {
    s.Lock()
    defer s.Unlock()

    b, err := json.Marshal(*s)
    if err != nil {
        return "{}"
    } else {
        return string(b)
    }
}

var (
    stats *Stats
    hits  *expvar.Map
)

func init() {

    expvar.Publish("now", expvar.Func(func() interface{} {
        return time.Now().Format("\"2006-01-02 15:04:05\"")
    }))

    stats = &Stats{}
    expvar.Publish("stats", stats)

    hits = expvar.NewMap("hits").Init()
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    p := r.URL.Path[1:]
    hits.Add(p, 1)
    stats.IncreaseTotalHit(1)
    fmt.Fprintf(w, "Hey! I love %s! hits: %v\n", p, hits.Get(p))
}

func errHandler(w http.ResponseWriter, r *http.Request) {
    stats.IncreaseErrorNums(1)
    fmt.Fprintf(w, "Error Nums: %v\n", stats.ErrorNums)
}

func main() {
    http.HandleFunc("/err", errHandler)
    http.HandleFunc("/", homeHandler)
    http.ListenAndServe(":8080", nil)
}
