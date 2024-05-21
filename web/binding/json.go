package binding

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
)

var Validator StructValidator = &defaultValidator{}

type jsonBinding struct {
	DisallowUnknownFields bool
	DisallowLessFiles     bool
}

func (b *jsonBinding) Name() string {
	return "json"
}

func (b *jsonBinding) Bind(req *http.Request, obj any) error {
	if req == nil || req.Body == nil {
		return errors.New("invalid request")
	}
	return b.decodeJson(req.Body, obj)
}

func (b *jsonBinding) decodeJson(body io.Reader, obj any) error {
	decoder := json.NewDecoder(body)
	if b.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	if b.DisallowLessFiles {
		err := validIsLessFields(obj, decoder)
		if err != nil {
			return err
		}
	} else {
		err := decoder.Decode(obj)
		if err != nil {
			return err
		}
	}
	return validate(obj)
}

func validate(obj any) error {
	return Validator.ValidateStruct(obj)
}

func validIsLessFields(obj any, decoder *json.Decoder) error {
	// 判断类型
	value := reflect.ValueOf(obj)
	if value.Kind() != reflect.Pointer {
		return errors.New("This argument need a pointer ")
	}

	// 这里取得了指针所指的对象
	elem := value.Elem().Interface()
	// 获取对象的类别
	valueOfElem := reflect.ValueOf(elem)

	switch valueOfElem.Kind() {
	case reflect.Struct:
		return checkStruct(valueOfElem, obj, decoder)
	case reflect.Slice, reflect.Array:
		elem := valueOfElem.Type().Elem()
		elemType := elem.Kind()
		if elemType == reflect.Struct {
			return checkSlice(elem, obj, decoder)
		}

	default:
		_ = decoder.Decode(obj)
	}
	return nil
}

func checkSlice(elem reflect.Type, obj any, decoder *json.Decoder) error {
	mapData := make([]map[string]interface{}, 0)
	_ = decoder.Decode(&mapData)
	if len(mapData) <= 0 {
		return nil
	}
	for _, item := range mapData {
		for i := 0; i < elem.NumField(); i++ {
			field := elem.Field(i)
			tag := field.Tag.Get("json")
			value := item[tag]
			if value == nil {
				return errors.New(fmt.Sprintf("filed [%s] is required", tag))
			}
		}
	}
	if obj != nil {
		marshal, _ := json.Marshal(mapData)
		_ = json.Unmarshal(marshal, obj)
	}
	return nil
}

func checkStruct(valueOfElem reflect.Value, obj any, decoder *json.Decoder) error {
	// 创建一个map存储key和elem里的字段进行比较
	m := make(map[string]any)
	err := decoder.Decode(&m)
	if err != nil {
		return err
	}

	for i := 0; i < valueOfElem.NumField(); i++ {
		field := valueOfElem.Type().Field(i)
		tag := field.Tag.Get("json")
		mapValue := m[tag]
		if mapValue == nil {
			return errors.New(fmt.Sprintf("filed [%s] is not exist", tag))
		}
	}
	// 因为decoder里面的数据流已经读完 不能二次重复读
	marshal, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
	}
	err = json.Unmarshal(marshal, obj)
	if err != nil {
		return err
	}
	// 切片和数组的情况
	return nil
}
