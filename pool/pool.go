package pool

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Jeffail/tunny"
)

// (1)
// sync.Pool
// 并发安全
// 参考：https://geektutu.com/post/hpg-sync-pool.html
// https://geektutu.com/post/hpg-concurrency-control.html
type Student struct {
	Name   string
	Age    int32
	Remark [1024]byte
}

var studentPool = sync.Pool{
	New: func() interface{} {
		return new(Student)
	},
}

func testSyncPool() {
	var buf, _ = json.Marshal(Student{Name: "Geektutu", Age: 25})
	stu := studentPool.Get().(*Student)
	json.Unmarshal(buf, stu)
	studentPool.Put(stu)
}

// (2)
// FuncPool 方法池接口
type FuncPool interface {
	Go(key interface{}, f func() error)
	Wait()
	Errs() map[interface{}]error
	Count() int
}

// NewFuncPool 实例化FuncPool，maxConcurrencyCount至少为1
func NewFuncPool(maxConcurrencyCount int32) FuncPool {
	if maxConcurrencyCount <= 0 {
		maxConcurrencyCount = 1
	}
	return &funcPoolImpl{
		errs:  map[interface{}]error{},
		chans: make(chan struct{}, maxConcurrencyCount),
	}
}

// funcPoolImpl 方法池实现
type funcPoolImpl struct {
	count int32
	errs  map[interface{}]error
	lock  sync.Mutex
	wg    sync.WaitGroup
	chans chan struct{}
}

// Go 异步执行
func (fc *funcPoolImpl) Go(key interface{}, f func() error) {
	fc.chans <- struct{}{}
	fc.wg.Add(1)
	go func() {
		var err error
		defer func() {
			if e := recover(); e != nil {
				err = fmt.Errorf("PANIC[%v]", e)
			}
			if err != nil && key != nil {
				fc.lock.Lock()
				fc.errs[key] = err
				fc.lock.Unlock()
			}
			atomic.AddInt32(&fc.count, 1)
			<-fc.chans
			fc.wg.Done()
		}()
		err = f()
	}()
}

// Wait 等待所有协程处理
func (fc *funcPoolImpl) Wait() {
	fc.wg.Wait()
}

// Errs 获取所有错误
func (fc *funcPoolImpl) Errs() map[interface{}]error {
	errs := make(map[interface{}]error, len(fc.errs))
	for k, v := range fc.errs {
		errs[k] = v
	}
	return errs
}

// Count 返回已处理的Func个数
func (fc *funcPoolImpl) Count() int {
	return int(atomic.LoadInt32(&fc.count))
}

// (3)开源库
func testTunnyPool() {
	pool := tunny.NewFunc(3, func(i interface{}) interface{} {
		log.Println(i)
		time.Sleep(time.Second)
		return nil
	})
	defer pool.Close()

	for i := 0; i < 10; i++ {
		go pool.Process(i)
	}
	time.Sleep(time.Second * 4)
}
