package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
	"time"
)

var (
	client = &fasthttp.Client{}
)

func DoHttp(url, method string, header map[string]string, body []byte) ([]byte, error) {
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}()

	req.SetRequestURI(url)
	req.Header.SetMethod(method)
	req.SetBody(body)
	for k, v := range header {
		req.Header.Set(k, v)
	}
	if err := client.DoTimeout(req, res, time.Duration(Config.Timeout)*time.Second); err != nil {
		return nil, err
	}

	if res.StatusCode() != 200 {
		var unescaped any
		if err := json.Unmarshal(res.Body(), &unescaped); err != nil {
			b, err := res.BodyUnbrotli()
			if err != nil {
				return nil, errors.New(fmt.Sprintf("failed to get request, body: %s", string(res.Body())))
			} else {
				return nil, errors.New(fmt.Sprintf("failed to get request, body: %s", string(b)))
			}
		}
		return nil, errors.New(fmt.Sprintf("failed to get request, body: %+v", unescaped))
	}
	raw, err := res.BodyUnbrotli()
	if err != nil {
		return nil, err
	}
	return CopyBody(raw), nil
}

// CopyBody in order to release the reference to src, make a deep copy of src and return
// For example: in concurrent circumstances,
func CopyBody(src []byte) []byte {
	dst := make([]byte, len(src))
	for i, b := range src {
		dst[i] = b
	}
	return dst
}

func StructToForm(raw any) (string, error) {
	jsonBytes, err := json.Marshal(raw)
	if err != nil {
		return "", err
	}
	jd, err := fastjson.ParseBytes(jsonBytes)
	if err != nil {
		return "", err
	}
	return parseValue(jd, ""), nil
}

func parseValue(jd *fastjson.Value, prefix string) string {
	result := ""
	if jd.Type() == fastjson.TypeString {
		return fmt.Sprintf("%s=%s", prefix, string(jd.GetStringBytes()))
	}
	jd.GetObject().Visit(func(k []byte, v *fastjson.Value) {
		key := ""
		if len(prefix) > 0 {
			key = fmt.Sprintf("%s[%s]", prefix, string(k))
		} else {
			key = string(k)
		}
		switch v.Type() {
		case fastjson.TypeNumber:
			result += fmt.Sprintf("%s=%d&", key, v.GetInt())
		case fastjson.TypeString:
			result += fmt.Sprintf("%s=%s&", key, string(v.GetStringBytes()))
		case fastjson.TypeArray:
			index := 0
			for _, each := range v.GetArray() {
				result += fmt.Sprintf("%s[%d]=%s&", string(k), index, string(each.GetStringBytes()))
				index++
			}
		case fastjson.TypeObject:
			v.GetObject().Visit(func(k2 []byte, v2 *fastjson.Value) {
				s := parseValue(v2, fmt.Sprintf("%s[%s]", key, string(k2)))
				result += fmt.Sprintf("%s&", s)
			})
		case fastjson.TypeNull:
			result += fmt.Sprintf("%s=&", key)
		}
	})
	return result[0 : len(result)-1]
}
