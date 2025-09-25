package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func wrap(in In, done In, stage Stage) Out {
	out := make(Bi)
	stageOut := stage(in)
	go func() {
		defer close(out)
		for {
			select {
			case <-done:
				go func() {
					for data := range stageOut {
						_ = data
					}
				}()
				return
			case data, ok := <-stageOut:
				if !ok {
					return
				}
				select {
				case <-done:
					go func() {
						for data := range stageOut {
							_ = data
						}
					}()
					return
				case out <- data:
				}
			}
		}
	}()
	return out
}

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	out := in

	for _, stage := range stages {
		out = wrap(out, done, stage)
	}
	return out
}
