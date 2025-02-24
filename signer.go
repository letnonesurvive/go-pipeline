package main

import (
	"runtime"
	"strconv"
	"strings"
	"sync"
)

const bufLen = 100

func SingleHash(in, out chan interface{}) {

	for value := range in {
		data := strconv.Itoa(value.(int))

		crc32First := DataSignerCrc32(data)
		md5 := DataSignerMd5(data)
		crc32Second := DataSignerCrc32(md5)

		out <- crc32First + "~" + crc32Second
	}
}

func MultiHash(in, out chan interface{}) {

	for value := range in {
		data := value.(string)
		var res string
		for i := 0; i < 6; i++ {
			res += DataSignerCrc32(strconv.Itoa(i) + data)
		}
		out <- res
	}
}

func CombineResults(in, out chan interface{}) {

	var res string
	for value := range in {
		hashes := strings.Split(res, "_")
		if value.(string) <= hashes[0] || len(hashes[0]) == 0 {
			res = value.(string) + "_" + res
		} else if value.(string) >= hashes[len(hashes)-1] {
			res += "_" + value.(string)
		}
	}
	out <- res
}

func ExecutePipeline1(freeFlowJobs ...job) { //works with accumulating data, not free flow

	tmp := make([]interface{}, 0, bufLen)

	wg := sync.WaitGroup{}
	for _, job := range freeFlowJobs {
		in := make(chan interface{}, bufLen)
		out := make(chan interface{}, bufLen)

		for _, value := range tmp {
			in <- value
		}
		tmp = tmp[:0] // clear the slice
		close(in)

		wg.Add(1)
		go func() {
			job(in, out)
			wg.Done()
		}()
		wg.Wait()

		close(out)
		for value := range out {
			tmp = append(tmp, value)
		}
	}
}

func worker(id int, ins, outs []chan interface{}, wg *sync.WaitGroup) {

	defer wg.Done()
	if (id + 1) == len(ins) {
		return
	}
	for value := range outs[id] {
		ins[id+1] <- value
		runtime.Gosched()
	}
	close(ins[id+1])
}

func RunJob(job job, in, out chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	job(in, out)
	close(out)
}

func ExecutePipeline(freeFlowJobs ...job) {

	wg := sync.WaitGroup{}
	numWorkers := len(freeFlowJobs)

	ins := make([]chan interface{}, numWorkers)
	outs := make([]chan interface{}, numWorkers)

	for i := 0; i < numWorkers; i++ {
		ins[i] = make(chan interface{}, 100)
		outs[i] = make(chan interface{}, 100)
		wg.Add(1)
		go worker(i, ins, outs, &wg)
	}

	for i, job := range freeFlowJobs {
		wg.Add(1)
		go RunJob(job, ins[i], outs[i], &wg)
	}

	wg.Wait()
}

func main() {

}

// сюда писать код
