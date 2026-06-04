package utilsdatatypes

import (
	"container/list"
	"sync"
)

type Queue struct {
	qlist      *list.List
	queueMutex sync.Mutex
}

func (ts *Queue) New() {
	ts.qlist = list.New()
}

func (ts *Queue) Enqueue(data interface{}) {
	ts.queueMutex.Lock()
	ts.qlist.PushBack(data)
	ts.queueMutex.Unlock()
}

func (ts *Queue) Dequeue() interface{} {
	ts.queueMutex.Lock()
	data := ts.qlist.Front()
	ts.qlist.Remove(data)
	ts.queueMutex.Unlock()
	return data
}

func (ts *Queue) Duplicate() list.List {
	ts.queueMutex.Lock()
	newList := list.New()
	newList.PushBackList(ts.qlist)
	ts.queueMutex.Unlock()
	return *newList
}

func (ts *Queue) Empty() {
	ts.queueMutex.Lock()
	ts.qlist = list.New()
	ts.queueMutex.Unlock()
}

func (ts *Queue) Length() int {
	ts.queueMutex.Lock()
	defer ts.queueMutex.Unlock()
	return ts.qlist.Len()

}

func (ts *Queue) Copy() []interface{} {
	//var s []byte
	s := make([]interface{}, 0)
	ts.queueMutex.Lock()
	defer ts.queueMutex.Unlock()

	for temp := ts.qlist.Front(); temp != nil; temp = temp.Next() {
		s = append(s, temp.Value)
	}
	return s
}
