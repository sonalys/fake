package do_not_change

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"time"
)

func sample() {
	log.Printf("Hello, World!")
	fmt.Printf("time is %v", time.Now())
}
