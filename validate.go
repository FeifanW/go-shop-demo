package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-shop-demo/common"
	"go-shop-demo/datamodels"
	"go-shop-demo/encrypt"
	"go-shop-demo/rabbitmq"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

// 设置集群地址，最好内外IP
var hostArray = []string{"127.0.0.1", "127.0.0.1"}
var localHost = "127.0.0.1"

// 数量控制接口服务器内网IP，或者getOne的SLB内网IP
var GetOneIp = "127.0.0.1"

var GetOnePort = "8084"

var port = "8083"

//var port = "8081"

var hashConsistent *common.Consistent

// rabbitmq
var rabbitMqValidate *rabbitmq.RabbitMQ

// 用来存放控制信息
type AccessControl struct {
	// 用来存放用户想要存放的信息
	sourcesArray map[int]time.Time
	*sync.RWMutex
}

// 服务器间隔时间，单位秒
var interval = 20

// 创建全局变量
var accessControl = &AccessControl{sourcesArray: make(map[int]time.Time)}

// 获取制定的数据
func (m *AccessControl) GetNewRecord(uid int) time.Time {
	m.RWMutex.RLock()
	defer m.RWMutex.Unlock()
	data := m.sourcesArray[uid]
	return data
}

// 设置记录
func (m *AccessControl) SetNewRecord(uid int) {
	m.RWMutex.Lock()
	m.sourcesArray[uid] = time.Now()
	m.RWMutex.Unlock()
}

// 黑名单结构体
type BlackList struct {
	listArray map[int]bool
	sync.RWMutex
}

var blackList = &BlackList{listArray: make(map[int]bool)}

// 获取黑名单
func (m *BlackList) GetBlackListByID(uid int) bool {
	m.RLock()
	defer m.RLock()
	return m.listArray[uid]
}

// 添加黑名单
func (m *BlackList) SetBlackListByID(uid int) bool {
	m.Lock()
	defer m.Unlock()
	m.listArray[uid] = true
	return true
}

func (m *AccessControl) GetDistributedRight(req *http.Request) bool {
	// 获取用户UID
	uid, err := req.Cookie("uid")
	if err != nil {
		return false
	}
	// 采用一致性hash算法，根据用户ID，判断具体机器
	hostRequest, err := hashConsistent.Get(uid.Value)
	if err != nil {
		return false
	}
	// 判断是否是本机
	if hostRequest == localHost {
		// 执行本机数据读取和教研
		return m.GetDataFromMap(uid.Value)
	} else {
		// 不是本机充当代理访问数据返回结果
		return GetDataFromOtherMap(hostRequest, req)
	}
}

// 获取本机map,并且处理逻辑，返回的结果类型为bool类型
func (m *AccessControl) GetDataFromMap(uid string) (isOK bool) {
	uidInt, err := strconv.Atoi(uid)
	if err != nil {
		return false
	}

	// 添加黑名单
	if blackList.GetBlackListByID(uidInt) {
		// 判断是否被添加到黑名单中
		return false
	}

	// 获取记录
	dataRecord := m.GetNewRecord(uidInt)
	// 执行判断逻辑
	if !dataRecord.IsZero() {
		// 业务判断，是否在指定直接之后
		if dataRecord.Add(time.Duration(interval) * time.Second).After(time.Now()) {
			return false
		}
	}
	m.SetNewRecord(uidInt)
	return true
}

// 获取其他节点处理结果
func GetDataFromOtherMap(host string, request *http.Request) bool {
	hostUrl := "http://" + host + ":" + port + "/checkRight"
	response, body, err := GetCurl(hostUrl, request)
	if err != nil {
		return false
	}
	// 判断状态
	if response.StatusCode == 200 {
		if string(body) == "true" {
			return true
		} else {
			return false
		}
	}
	return false
}

// 模拟请求
func GetCurl(hostUrl string, request *http.Request) (response *http.Response, body []byte, err error) {
	// 获取uid
	uidPre, err := request.Cookie("uid")
	if err != nil {
		return
	}
	// 获取sign
	uidSign, err := request.Cookie("sign")
	if err != nil {
		return
	}
	// 模拟接口访问
	client := &http.Client{}
	req, err := http.NewRequest("GET", hostUrl, nil)
	if err != nil {
		return
	}
	// 手动指定，排查多与cookies
	cookieUid := &http.Cookie{Name: "uid", Value: uidPre.Value, Path: "/"}
	cookieSign := &http.Cookie{Name: "sign", Value: uidSign.Value, Path: "/"}
	// 添加cookie到模拟的请求中
	req.AddCookie(cookieUid)
	req.AddCookie(cookieSign)
	// 获取返回结果
	response, err = client.Do(req)
	defer response.Body.Close()
	if err != nil {
		return
	}
	body, err = io.ReadAll(response.Body)
	return
}

func CheckRight(w http.ResponseWriter, r *http.Request) {
	right := accessControl.GetDistributedRight(r)
	if !right {
		w.Write([]byte("false"))
		return
	}
	w.Write([]byte("true"))
}

// 执行正常业务逻辑
func Check(w http.ResponseWriter, r *http.Request) {
	// 执行正常业务逻辑
	fmt.Println("执行check")
	queryForm, err := url.ParseQuery(r.URL.RawQuery)
	//if err != nil && len(queryForm["productID"]) <= 0 && len(queryForm["productID"][0]) <= 0 {
	if err != nil || len(queryForm["productID"]) <= 0 {
		w.Write([]byte("false"))
		return
	}
	productString := queryForm["productID"][0]
	fmt.Println(productString)
	// 获取用户cookie
	userCookie, err := r.Cookie("uid")
	if err != nil {
		w.Write([]byte("false"))
		return
	}
	// 1.分布式权限验证
	right := accessControl.GetDistributedRight(r)
	if right == false {
		w.Write([]byte("false"))
	}
	// 2.获取数量控制权限，防止秒杀出现超卖现象
	hostUrl := "http://" + GetOneIp + ":" + GetOnePort + "/getOne"
	responseValidate, validateBody, err := GetCurl(hostUrl, r)
	if err != nil {
		w.Write([]byte("false"))
		return
	}
	// 3.判断数量控制接口
	if responseValidate.StatusCode == 200 {
		if string(validateBody) == "true" {
			// 整合下单
			// 获取商品ID
			productID, err := strconv.ParseInt(productString, 10, 64)
			if err != nil {
				w.Write([]byte("false"))
				return
			}
			// 获取用户ID
			userID, err := strconv.ParseInt(userCookie.Value, 10, 64)
			if err != nil {
				w.Write([]byte("false"))
				return
			}
			// 创建消息体
			message := datamodels.Message{productID, userID}
			// 类型转化
			byteMessage, err := json.Marshal(message)
			if err != nil {
				w.Write([]byte("false"))
				return
			}
			// 生产消息
			err = rabbitMqValidate.PublishSimple(string(byteMessage))
			if err != nil {
				w.Write([]byte("false"))
				return
			}
			w.Write([]byte("true"))
			return
		}
	}
	w.Write([]byte("false"))
	return
}

// 统一验证拦截器，每个接口都需要提前验证
func Auth(w http.ResponseWriter, r *http.Request) error {
	fmt.Println("执行验证！")
	// 添加基于cookie的权限验证
	err := CheckUserInfo(r)
	if err != nil {
		return err
	}
	return nil
}

// 身份校验函数
func CheckUserInfo(r *http.Request) error {
	// 获取Uid,cookie
	uidCookie, err := r.Cookie("uid")
	if err != nil {
		return errors.New("用户UID Cookie获取失败")
	}
	// 获取用户加密串
	signCookie, err := r.Cookie("sign")
	if err != nil {
		return errors.New("用户加密串 Cookie 获取失败")
	}
	// 对信息进行解密
	signByte, err := encrypt.DePwdCode(signCookie.Value)
	if err != nil {
		return errors.New("加密串已被修改")
	}
	fmt.Println("结果比对")
	fmt.Println("用户ID：" + uidCookie.Value)
	fmt.Println("解密后用户ID：" + string(signByte))
	if checkInfo(uidCookie.Value, string(signByte)) {
		return nil
	}
	return errors.New("身份校验失败！")
}

// 自定义逻辑判断
func checkInfo(checkStr string, signStr string) bool {
	if checkStr == signStr {
		return true
	}
	return false
}

func main() {
	// 负载均衡器设置
	// 采用一致性哈希算法
	hashConsistent := common.NewConsistent()
	// 采用一致性hash算法，添加节点
	for _, v := range hostArray {
		hashConsistent.Add(v)
	}

	localIp, err := common.GetIntranceIp()
	if err != nil {
		fmt.Println(err)
	}
	localHost = localIp
	fmt.Println(localHost)
	rabbitMqValidate = rabbitmq.NewRabbitMQSimple("imoocProduct") // 后面用到rabbitmq直接用这个变量就行
	defer rabbitMqValidate.Destory()

	// 1.过滤器
	filter := common.NewFilter()
	// 注册拦截器
	filter.RegisterFilterUri("/check", Auth)
	filter.RegisterFilterUri("/checkRight", Auth)
	// 2.启动服务
	http.HandleFunc("/check", filter.Handle(Check))
	http.HandleFunc("/checkRight", filter.Handle(CheckRight))
	// 启动服务
	http.ListenAndServe(":8083", nil)
}
