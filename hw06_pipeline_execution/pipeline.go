package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	input := make(Bi)
	output := make(Bi)

	// observe input channel
	go func() {
		// close input channel
		defer close(input)
		for v := range in {
			select {
			case <-done:
				return
			case input <- v:
			}
		}
	}()

	// observe output channel
	go func() {
		out := DoExecutePipeline(input, stages...)

		// wait until output channel is closed
		defer func() {
			for range out {
				continue
			}
		}()

		// close output channel
		defer close(output)
		for {
			select {
			case <-done:
				return
			case v := <-out:
				if v == nil {
					return
				}
				output <- v
			}
		}
	}()

	return output
}

func DoExecutePipeline(in In, stages ...Stage) Out {
	// pipe all stages
	out := in
	for _, stage := range stages {
		out = stage(out)
	}

	return out
}
