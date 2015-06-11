package main

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"log"
	"time"
)

type Dyrwood struct {
	db              *sql.DB
	refreshViewChan chan string
	refreshers      map[string]chan struct{}
	refreshDelay    time.Duration
}

type RefreshMaterializedViewRequest struct {
	ViewName string `json:"view_name"`
}

func NewDyrwood(db *sql.DB, refreshDelaySeconds float64) (out *Dyrwood) {
	out = &Dyrwood{
		db:              db,
		refreshViewChan: make(chan string),
		refreshers:      make(map[string]chan struct{}),
		refreshDelay:    time.Duration(refreshDelaySeconds * float64(time.Second)),
	}

	go out.refreshViewListener()

	return
}

func (self *Dyrwood) RefreshMaterializedView(req *RefreshMaterializedViewRequest, res *bool) error {
	self.refreshViewChan <- req.ViewName
	*res = true
	return nil
}

func (self *Dyrwood) refreshViewListener() {
	var refresher chan struct{}
	var ok bool

	for viewName := range self.refreshViewChan {
		log.Printf("refresh for '%s' requested", viewName)
		refresher, ok = self.refreshers[viewName]
		if !ok {
			refresher = make(chan struct{}, 1)
			go self.refresher(viewName, refresher)
			self.refreshers[viewName] = refresher
			log.Printf("started refresher for '%s'", viewName)
		}

		// non-blocking channel send, drop values over buffer size
		select {
		case refresher <- struct{}{}:
		default:
		}
	}
}

func (self *Dyrwood) refresher(viewName string, inChan chan struct{}) {
	var lastRefreshed, nextRefresh, now time.Time
	var err error

	for _ = range inChan {
		now = time.Now()

		if self.refreshDelay > 0 {
			nextRefresh = lastRefreshed.Add(self.refreshDelay)
			if now.Before(nextRefresh) {
				log.Printf("sleeping for '%s'", viewName)
				time.Sleep(nextRefresh.Sub(now))
			}
		}

		_, err = self.db.Exec(fmt.Sprintf(
			`REFRESH MATERIALIZED VIEW CONCURRENTLY %s`,
			pq.QuoteIdentifier(viewName)))
		if err == nil {
			log.Printf("refreshed view '%s'", viewName)
		} else {
			log.Printf("error refreshing view '%s': '%s'", viewName, err)
		}

		lastRefreshed = time.Now()
	}
}
