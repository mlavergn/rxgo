package rxgo

import (
	"log"
)

// ToInt export
func ToInt(value interface{}, failure int) (ret int) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			ret = failure
		}
	}()
	return value.(int)
}

// ToString export
func ToString(value interface{}, failure string) (ret string) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			ret = failure
		}
	}()
	return value.(string)
}
