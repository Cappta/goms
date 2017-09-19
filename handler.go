package goms

import "github.com/Cappta/gohelpgabs"

// handler is expected alter the container and the result will be considered the response
type handler func(container *gohelpgabs.Container)
