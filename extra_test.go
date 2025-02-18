package main

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

/*
	Тест, предложенный одним из учащихся курса, Ilya Boltnev

	В чем его преимущество по сравнению с TestPipeline?
	1. Он проверяет то, что все функции действительно выполнились
	2. Он дает представление о влиянии time.Sleep в одном из звеньев конвейера на время работы

	возможно кому-то будет легче с ним
	при правильной реализации ваш код конечно же должен его проходить
*/

func TestByIlia(t *testing.T) {

	var recieved uint32
	freeFlowJobs := []job{
		job(func(in, out chan interface{}) {
			out <- uint32(1)
			out <- uint32(7)
			out <- uint32(11)
			fmt.Println("job1 finished")
		}),
		job(func(in, out chan interface{}) {
			for val := range in {
				multiplied := val.(uint32) * 3
				fmt.Println("multiplied", multiplied)
				out <- val.(uint32) * 3
				time.Sleep(time.Millisecond * 100)
			}
			fmt.Println("job2 finished")
		}),
		job(func(in, out chan interface{}) {
			for val := range in {
				fmt.Println("collected", val)
				atomic.AddUint32(&recieved, val.(uint32))
			}
			fmt.Println("job3 finished")
		}),
	}

	start := time.Now()

	ExecutePipeline(freeFlowJobs...)

	end := time.Since(start)

	expectedTime := time.Millisecond * 300

	if end > expectedTime {
		t.Errorf("execition too long\nGot: %s\nExpected: <%s", end, expectedTime)
	}

	if recieved != (1+7+11)*3 {
		t.Errorf("f3 have not collected inputs, recieved = %d", recieved)
	}
}
