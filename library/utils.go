package library

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"sort"
	//"strconv"
	//"strconv"

	//"encoding/base64"

	"crypto/md5"
	"crypto/sha1"
	"encoding/csv"
	"encoding/hex"
	"io/ioutil"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/cstockton/go-conv"
	"gopkg.in/mgo.v2/bson"
)

var (
	// ErrAbort custom error when user stop request handler manually.
	APP_PATH = GetCurrentPath()
)

// 获取当前时间
func GetCurrentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// 获取mongoid
func GetMongoId() string {
	id := bson.NewObjectId().Hex()
	fmt.Println("GetMongoId:", id)
	return id
}

func Explode(sep string, s string) []string {
	return strings.Split(s, sep)
}

func Urldecode(s string) string {
	tmp, err := url.QueryUnescape(s)
	if err != nil {
		panic(err)
	}
	return string(tmp)
}

func Urlencode(s string) string {
	return url.QueryEscape(s)
}

func Trim(s string, cutsets ...string) string {

	if len(cutsets) == 0 {
		return strings.TrimSpace(s)
	}
	tmp := s
	for _, cutset := range cutsets {
		tmp = strings.Trim(tmp, cutset)
	}

	return tmp

}

func Nl2br(val string) string {
	s := strings.Replace(val, "\n", "<br>", -1)
	return s
}

// 是否是boolean
func Is_bool(v interface{}) bool {
	t := reflect.TypeOf(v)
	return t.Kind() == reflect.Bool
}

// 是否是数组
func Is_array(v interface{}) bool {
	t := reflect.TypeOf(v)
	return t.Kind() == reflect.Slice || t.Kind() == reflect.Array
}

// 是否为空
func Empty(v interface{}) bool {
	if v == nil {
		return true
	}

	t := reflect.TypeOf(v)
	value := reflect.ValueOf(v)
	if t.Kind() == reflect.String {
		return value.Len() == 0
	} else if t.Kind() == reflect.Bool {
		return value.Bool() == false
	} else if t.Kind() == reflect.Slice {
		return value.Len() == 0
	} else if t.Kind() == reflect.Array {
		return value.Len() == 0
	} else if t.Kind() == reflect.Map {
		return value.Len() == 0
	} else if t.Kind() == reflect.Chan {
		return value.Len() == 0
	} else {
		panic(500)
		return false
	}
}

func Intval(n interface{}) int64 {
	if n == nil {
		return 0
	}
	val, err := conv.Int64(n)
	CheckErr(err)
	return val
}

func Strval(n interface{}) string {
	if n == nil {
		return ""
	}

	targetValue := reflect.ValueOf(n)
	switch reflect.TypeOf(n).Kind() {
	case reflect.String:
		return targetValue.String()
	case reflect.Bool:
		if targetValue.Bool() {
			return "1"
		} else {
			return ""
		}
	}
	val, err := conv.String(n)
	CheckErr(err)
	return val
}

func Boolval(n interface{}) bool {
	if n == nil {
		return false
	}
	val, err := conv.Bool(n)
	CheckErr(err)
	return val
}

func ArrayToCVS(datas map[string][][]string) (int, []byte) {
	f, err := ioutil.TempFile("", "export_csv_")
	//tmpname := os.TempDir() + "/export_csv_" + GetCurrentTime() + ".csv"
	//f, err := os.Create(tmpname) //创建文件
	tmpname := f.Name()
	fmt.Println("ArrayToCVS:", tmpname)
	CheckErr(err)

	defer func() {
		f.Close()
		os.Remove(tmpname)
	}()

	f.WriteString("\xEF\xBB\xBF") // 写入UTF-8 BOM
	w := csv.NewWriter(f)         //创建一个新的写入文件流
	w.WriteAll(datas["title"])    //写入列标题数据
	//	data := [][]string{
	//		{"1", "中国", "23"},
	//		{"2", "美国", "23"},
	//		{"3", "bb", "23"},
	//		{"4", "bb", "23"},
	//		{"5", "bb", "23"},
	//	}
	w.WriteAll(datas["result"]) //写入数据
	w.Flush()

	contents, err := ioutil.ReadFile(tmpname)
	CheckErr(err)
	return len(contents), contents
}

func GetCurrentPath() string {
	s, err := exec.LookPath(os.Args[0])
	CheckErr(err)
	i := strings.LastIndex(s, "\\")
	path := string(s[0 : i+1])
	return path
}

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

func Json_encode(str interface{}) string {
	b, err := json.Marshal(str)
	CheckErr(err)
	return (string(b))
}

func Json_decode(str string, b bool) map[string]interface{} {
	var obj interface{}
	json.Unmarshal([]byte(str), &obj)
	m := obj.(map[string]interface{})
	return m
}

func Array_keys(m interface{}) []string {
	keys := make([]string, 0)
	targetValue := reflect.ValueOf(m)
	switch reflect.TypeOf(m).Kind() {
	case reflect.Map:
		for _, key := range targetValue.MapKeys() {
			keys = append(keys, Strval(key.Interface()))
		}
	}
	return keys
}

func In_array(obj interface{}, target interface{}) bool {
	targetValue := reflect.ValueOf(target)
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true
		}
	}
	return false
}

func Ksort(m map[string]interface{}) map[string]interface{} {

	/* 排序输出 */
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	rmap1 := make(map[string]interface{}, 0)
	for _, k := range keys {
		//fmt.Println("Key:", k, "Value:", m[k])
		rmap1[k] = m[k]
	}
	return rmap1
}

func Array_values(m interface{}) []interface{} {

	//value := reflect.ValueOf(m)
	//type1 := reflect.TypeOf(m)
	//fmt.Println("reflect.ValueOf", value, "reflect.TypeOf", type1)
	switch m.(type) {
	case []string:
		//fmt.Println("reflect.ValueOf", value, "reflect.TypeOf", type1)
		var values []interface{}
		m1, _ := m.([]string)
		for _, v := range m1 {
			values = append(values, v)
		}
		return values
	case map[string]interface{}:
		var values []interface{}
		m1, _ := m.(map[string]interface{})
		for _, v := range m1 {
			values = append(values, v)
		}
		return values
	default:
		return nil
	}
}

//生成随机字符串
func Uniqid(num int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < num; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func Sha1(s string) string {
	r := sha1.Sum([]byte(s))
	return hex.EncodeToString(r[:])
}

func Md5(s string) string {
	r := md5.Sum([]byte(s))
	return hex.EncodeToString(r[:])
}
