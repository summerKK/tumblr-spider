package module

import "sync"

func Merge(done <-chan struct{}, cs []<-chan File) <-chan File {
	var wg sync.WaitGroup
	out := make(chan File, 100)

	output := func(c <-chan File) {
		defer wg.Done()
		for n := range c {
			select {
			case out <- n:
			case <-done:
				return
			}
		}
	}

	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
