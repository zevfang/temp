package main

type Request struct {
	fn   func() int
	data []byte
	op   int
	c    chan int
}

type Worker struct {
	req     chan Request
	pending int
	index   int
	done    chan struct{}
}

type Pool []*Worker

func (p Pool) Len() int {
	return len(p)
}

func (p Pool) Less(i, j int) bool {
	return p[i].pending < p[j].pending
}

func (p Pool) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
	p[i].index = i
	p[j].index = j
}

func (p *Pool) Push() int {
	return 0
}

func (p *Pool) Pop() int {
	return 0
}
func main() {

}
