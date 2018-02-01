/*
 * 说明：数据库工具类函数
 * 作者：zhe
 * 时间：2018-01-28 21:32
 * 更新：
 */

package dao

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// 自定义新类型 addrs 来重新实现flag.Value接口
// 使得通过一个命令行参数可以指定多个值，然后解析到slice中
type addrs []string

// 实现 String 方法
func (a *addrs) String() string {
	return fmt.Sprintf("%v", *a)
}

// 重写 Set 方法, 处理每个value, 使其追加到最终的切片中
// Note: Set接口决定了如何解析flag的值
func (a *addrs) Set(value string) error {
	*a = append(*a, value)
	return nil
}

const (
	TimeLayout = "2006-01-02 15:04:05"
	DateLayout = "2006-01-02"
)

// 获取当前时间
func Now() string {
	now := time.Now()
	local, err := time.LoadLocation("Local") // 服务器上设置的时区
	if err != nil {
		println(err)
	}
	return now.In(local).Format(TimeLayout)
}

// 获取当前日期
func Date() string {
	return time.Now().Format(DateLayout)
}

/*
 * 生成随机字符串
 */

const (
	StdLen  = 16
	UUIDLen = 20
)

var StdChars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

func NewStr() string {
	return NewLenChars(StdLen, StdChars)
}

func NewLenStr(length int) string {
	return NewLenChars(length, StdChars)
}

// NewLenChars returns a new random string of the provided length, consisting of the provided byte slice of allowed characters(maximum 256).
func NewLenChars(length int, chars []byte) string {
	if length == 0 {
		return ""
	}
	clen := len(chars)
	if clen < 2 || clen > 256 {
		panic("Wrong charset length for NewLenChars()")
	}
	maxrb := 255 - (256 % clen)
	b := make([]byte, length)
	r := make([]byte, length+(length/4)) // storage for random bytes.
	i := 0
	for {
		if _, err := rand.Read(r); err != nil {
			panic("Error reading random bytes: " + err.Error())
		}
		for _, rb := range r {
			c := int(rb)
			if c > maxrb {
				continue // Skip this number to avoid modulo bias.
			}
			b[i] = chars[c%clen]
			i++
			if i == length {
				return string(b)
			}
		}
	}
}

// RandNumMath return 6 bit random num by math
func RandomMath() string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	num := fmt.Sprintf("%06v", rnd.Int31n(1000000))
	return num
}

// Transform `struct` to `bson.M`
func StructToBsonMap(src interface{}, dst *bson.M) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, dst)
	if err != nil {
		return err
	}
	return nil
}

// bson.M to json
func BsonMapToJson(src ...interface{}) error {
	var err error
	var data []byte
	for _, v := range src {
		if data, err = json.Marshal(v); err != nil {
			break
		}
		if err = OutputJson(data); err != nil {
			break
		}
	}
	return err
}

// OutputJson 将数据转换为Json格式输出到标准输出
func OutputJson(data []byte) error {
	var out bytes.Buffer
	if err := json.Indent(&out, data, "", "  "); err != nil {
		return err
	}
	out.WriteTo(os.Stdout)
	fmt.Println()

	return nil
}
