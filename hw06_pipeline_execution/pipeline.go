package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func helper(in In, done In) Out {
	out := make(Bi)
	go func() {
		defer close(out)
		for {
			select {
			case <-done:
				go func() {
					for data := range in {
						_ = data
					}
				}()
				return
			case data, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- data:
				case <-done:
					go func() {
						for data := range in {
							_ = data
						}
					}()
					return
				}
			}
		}
	}()
	return out
}

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	out := in

	for _, stage := range stages {
		out = stage(helper(out, done))
	}
	return out
}
