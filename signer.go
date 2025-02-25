package main

import (
	"runtime"
	"sort"
	"strconv"
	"sync"
)

const bufLen = 100

func SingleHash(in, out chan interface{}) {

	wg := sync.WaitGroup{}

	for value := range in {
		data := strconv.Itoa(value.(int))

		crc32First := make(chan string)
		crc32Second := make(chan string)

		go func() {
			crc32First <- DataSignerCrc32(data)
		}()

		md5 := DataSignerMd5(data)
		go func(md5 string) {
			crc32Second <- DataSignerCrc32(md5)
		}(md5)

		wg.Add(1)
		go func(crc32First, crc32Second chan string, wg *sync.WaitGroup) {
			defer wg.Done()
			result := <-crc32First + "~" + <-crc32Second
			out <- result
		}(crc32First, crc32Second, &wg)
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {

	hashNumbers := 6
	wgForMultiHash := sync.WaitGroup{}
	for value := range in {
		data := value.(string)
		results := make([]string, hashNumbers)

		wg := sync.WaitGroup{}
		for i := 0; i < hashNumbers; i++ {
			wg.Add(1)
			go func(i int, wg *sync.WaitGroup) {
				defer wg.Done()
				results[i] = DataSignerCrc32(strconv.Itoa(i) + data)
			}(i, &wg)
		}

		wgForMultiHash.Add(1)
		go func(out chan interface{}, wg, wgForMultiHash *sync.WaitGroup) {
			defer wgForMultiHash.Done()
			wg.Wait()
			var res string
			for _, value := range results {
				res += value
			}
			out <- res
		}(out, &wg, &wgForMultiHash)
	}
	wgForMultiHash.Wait()
}

func CombineResults(in, out chan interface{}) {

	var results []string
	for value := range in {
		results = append(results, value.(string))
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i] < results[j]
	})
	var res string
	for i, value := range results {
		res += value
		if i != len(results)-1 {
			res += "_"
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

func pipelineWorker(id int, ins, outs []chan interface{}, wg *sync.WaitGroup) {

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

func runJob(job job, in, out chan interface{}, wg *sync.WaitGroup) {
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
		ins[i] = make(chan interface{}, bufLen)
		outs[i] = make(chan interface{}, bufLen)
		wg.Add(1)
		go pipelineWorker(i, ins, outs, &wg)
	}

	for i, job := range freeFlowJobs {
		wg.Add(1)
		go runJob(job, ins[i], outs[i], &wg)
	}

	wg.Wait()
}

func main() {

}

// сюда писать код
