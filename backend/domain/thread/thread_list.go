package thread

type ThreadList struct {
	threads []*Thread
}

func NewThreadList(threads []*Thread) *ThreadList {
	return &ThreadList{threads: threads}
}

func (l *ThreadList) Threads() []*Thread {
	return l.threads
}
