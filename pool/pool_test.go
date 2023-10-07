package pool

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/panjf2000/ants/v2"
	. "github.com/smartystreets/goconvey/convey" //https://www.liwenzhou.com/posts/Go/unit-test-5/
)

// go test -v
// go test -v -run 函数名字 运行某个特定的函数
func TestFuncPool(t *testing.T) {
	failErr := errors.New("fail")
	panicErr := errors.New("PANIC[test panic]")

	testFcPanic := func() error {
		panic("test panic")
	}
	testFcSucc := func() error {
		return nil
	}
	testFcFail := func() error {
		return failErr
	}

	Convey("GoAndWait", t, func() {
		testCases := []struct {
			keys  []string
			funcs []func() error
			count int
			errs  map[interface{}]error
		}{
			{[]string{}, []func() error{}, 0, map[interface{}]error{}},
			{[]string{"1"}, []func() error{testFcSucc}, 1, map[interface{}]error{}},
			{[]string{"1"}, []func() error{testFcFail}, 1, map[interface{}]error{"1": failErr}},
			{[]string{"1"}, []func() error{testFcPanic}, 1, map[interface{}]error{"1": panicErr}},
			{[]string{"1", "2", "3"}, []func() error{testFcSucc, testFcSucc, testFcSucc}, 3, map[interface{}]error{}},
			{[]string{"1", "2", "3"}, []func() error{testFcSucc, testFcPanic, testFcFail}, 3, map[interface{}]error{"2": panicErr, "3": failErr}},
		}

		for _, testCase := range testCases {
			funcPool := NewFuncPool(3)
			for i, key := range testCase.keys {
				funcPool.Go(key, testCase.funcs[i])
			}
			funcPool.Wait()
			So(funcPool.Count(), ShouldEqual, testCase.count)
			So(funcPool.Errs(), ShouldResemble, testCase.errs)
		}
	})
}

func TestAntsWithCustomGo(t *testing.T) {
	// 创建自定义协程池
	pool, _ := ants.NewPool(2)
	defer pool.Release()
	for i := 0; i < 6; i++ {
		// 提交任务
		_ = pool.Submit(func() {
			time.Sleep(time.Second * 1)
			fmt.Println("time: ", time.Now().Format("2006-01-02 15:04:05"))
		})
		fmt.Println("i=", i, "当前运行协程数量: ", pool.Running())
	}
	time.Sleep(time.Second * 3)
	fmt.Println("finish")
}
