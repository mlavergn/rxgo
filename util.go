package rx

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

// ToStringMap export
func ToStringMap(value interface{}, failure map[string]interface{}) (ret map[string]interface{}) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			ret = failure
		}
	}()
	return value.(map[string]interface{})
}

// ToByteArray export
func ToByteArray(value interface{}, failure []byte) (ret []byte) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			ret = failure
		}
	}()
	return value.([]byte)
}

// ToByteArrayArray export
func ToByteArrayArray(value interface{}, failure [][]byte) (ret [][]byte) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			ret = failure
		}
	}()
	return value.([][]byte)
}

// ToByteString export
func ToByteString(value interface{}, failure string) (ret string) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			ret = failure
		}
	}()
	bytes := value.([]byte)
	return string(bytes)
}
