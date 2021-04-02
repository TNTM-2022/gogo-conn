package interfaces

import (
	"context"
	"sync"
)

type MainControl struct {
	Ctx context.Context
	Wg  *sync.WaitGroup
}
