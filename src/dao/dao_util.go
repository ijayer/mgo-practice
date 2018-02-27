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
	"regexp"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// 自定义新类型 addrs 来重新实现flag.Value接口
// 使得通过一个命令行参数可以指定多个值;然后解析到slice中
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
	fmt.Printf("total: %v\n", len(src))

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

const (
	RegexEmail            = `^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`                        // 邮箱
	RegexMobile           = `^(13[0-9]|14[5|7]|15[0|1|2|3|5|6|7|8|9]|18[0|1|2|3|5|6|7|8|9])\d{8}$` // 手机号
	RegexAnyNum           = "^[0-9]*$"                                                             // 数字
	RegexChinese          = "[\u4e00-\u9fa5]"                                                      // 汉字
	RegexAlphabet         = `^[A-Za-z]+$`                                                          // 26个英文字母
	RegexNumAlphabet      = `^[A-Za-z0-9]+$`                                                       // 数字+英文字母
	RegexSpecialAlpha     = `[^%&=?$\x22]+`                                                        // 允许这些特殊字符
	RegexMobile3Prefix    = `^(13[0-9]|14[5|7]|15[0|1|2|3|5|6|7|8|9]|18[0|1|2|3|5|6|7|8|9])\d{0}$` // 手机号前三位
	RegexCnEnNumUnderline = `^[\u4E00-\u9FA5A-Za-z0-9_]+$`                                         // 中+英+数+_
)

// 字段匹配
func MatchKeys(keys ...string) []bson.M {
	var ms []bson.M
	for _, key := range keys {
		ok1, _ := regexp.MatchString(RegexAlphabet, key)
		ok2, _ := regexp.MatchString(RegexNumAlphabet, key)
		ok3, _ := regexp.MatchString(RegexChinese, key)
		ok4, _ := regexp.MatchString(RegexCnEnNumUnderline, key)
		ok5, _ := regexp.MatchString(RegexAnyNum, key)
		ok6, _ := regexp.MatchString(RegexSpecialAlpha, key)
		ok7, _ := regexp.MatchString(RegexEmail, key)
		ok8, _ := regexp.MatchString(RegexMobile, key)
		ok9, _ := regexp.MatchString(RegexMobile3Prefix, key)

		if ok1 || ok2 || ok3 || ok4 || ok5 || ok6 {
			ms = append(ms, bson.M{"name": bsonRegex(key)}) // 姓名
			ms = append(ms, bson.M{"friends": bson.M{"$in": []string{key}}})
		}

		if ok1 || ok2 || ok7 || ok8 || ok9 {
			ms = append(ms, bson.M{"email": bsonRegex(key)}) // 邮箱
		}
	}
	return ms
}

// bson regex
func bsonRegex(key string) bson.M {
	return bson.M{"$regex": bson.RegEx{Pattern: key, Options: "i"}}
}
