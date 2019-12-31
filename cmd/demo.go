package main

import (
	"log"
	"rxgo"
)

func main() {
	rxgo.RxSetup(false)

	obvr := rxgo.NewRxObserver()
	obvb := rxgo.NewRxInterval(200)
	obvb.Take(10)
	obvb.Subscribe <- obvr
	for {
		select {
		case next := <-obvr.Next:
			log.Println("next", next.(int))
			break
		case <-obvr.Error:
			log.Println("error")
			return
		case <-obvr.Complete:
			log.Println("complete")
			return
		}
	}
}
