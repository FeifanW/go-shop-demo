package common

import (
	"net/http"
	"strings"
)

// 声明一个新的数据类型（函数类型）
type FilterHandle func(rw http.ResponseWriter, req *http.Request) error

// 拦截结构体
type Filter struct {
	// 用来存储需要拦截的URI
	filterMap map[string]FilterHandle
}

// Filter初始化函数
func NewFilter() *Filter {
	return &Filter{make(map[string]FilterHandle)}
}

// 注册拦截器
func (f *Filter) RegisterFilterUri(uri string, handler FilterHandle) {
	f.filterMap[uri] = handler
}

// 根据uri获取对应的handle
func (f *Filter) GetFilterHandle(uri string) FilterHandle {
	return f.filterMap[uri]
}

// 声明新的函数类型
type WebHandle func(rw http.ResponseWriter, req *http.Request)

func (f *Filter) Handle(webHandle WebHandle) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		for path, handle := range f.filterMap {
			//if path == r.RequestURI {
			if strings.Contains(r.RequestURI, path) {
				// 执行拦截业务逻辑
				err := handle(rw, r)
				if err != nil {
					rw.Write([]byte(err.Error()))
					return
				}
				// 跳出循环
				break
			}
		}
		// 执行正常注册的函数
		webHandle(rw, r)
	}

}
