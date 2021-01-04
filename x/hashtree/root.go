package hashtree

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

// Root is the root node for hashtree
type Root struct {
	_node interface{}
	_mod  map[string]interface{}
	_bak  map[string]interface{}
}

// Load stores node into_node field,imports data from map,
// and set then into tree nodes.
// node is a ptr to struct.
func (r *Root) Load(node interface{}, hash map[string]string) error {
	r._node = node
	r._mod = make(map[string]interface{})
	r._bak = make(map[string]interface{})

	value := reflect.Indirect(reflect.ValueOf(r._node))
	key := ""
	r.setStructValue(value, key, hash)
	return nil
}

// Dump gets data from mod map that was modified recently,
// and this action will clear mod map for next use.
func (r *Root) Dump() (map[string]string, error) {
	hash := make(map[string]string)
	var err error
	for key, value := range r._mod {
		hash[key], err = interfaceToString(value)
		if err != nil {
			return nil, err
		}
	}
	r._mod = make(map[string]interface{})
	r._bak = make(map[string]interface{})
	return hash, nil
}

// Revert recover data from bak map to modified fields.
func (r *Root) Revert() error {
	for key, value := range r._bak {
		r.setKeyValue(key, value)
	}
	r._mod = make(map[string]interface{})
	r._bak = make(map[string]interface{})
	return nil
}

// Field returns field pointer by path
func (r *Root) Field(fields ...string) unsafe.Pointer {
	nodeValue := reflect.Indirect(reflect.ValueOf(r._node))
	value, ok := getValueByPath(nodeValue, fields)
	if !ok {
		return nil
	}
	return getValuePtr(value)
}

func (r *Root) setStructValue(v reflect.Value, key string,
	hash map[string]string) {

	for i := 0; i < v.NumField(); i++ {
		value := v.Field(i)
		field := v.Type().Field(i)

		if value.Kind() != reflect.Struct {
			continue
		}

		if value.FieldByName("_root").IsValid() {
			// It's a leaf, set field value
			r.setFieldValue(value, field, key, hash)
		} else {
			// It's a _node, use recursion
			r.setStructValue(value, r.appendKey(field, key), hash)
		}
	}
}

func (r *Root) setFieldValue(value reflect.Value, field reflect.StructField,
	key string, hash map[string]string) {

	// set _root *Root
	*(**Root)(getValueFieldPtr(value, "_root")) = r

	// set key string
	key = r.appendKey(field, key)
	setValueField(value, "_key", key)

	//set value custom type
	if s, ok := hash[key]; ok {
		v := value.FieldByName("_value")
		if !v.IsValid() {
			panic(fmt.Errorf("field %s is not valid", key))
		}
		if err := stringToValue(s, v); err != nil {
			panic(err)
		}
	}
}

func (r *Root) appendKey(field reflect.StructField, key string) string {
	var buf bytes.Buffer
	buf.WriteString(key)
	if key != "" {
		buf.WriteString(".")
	}
	buf.WriteString(field.Name)
	return buf.String()
}

func (r *Root) setKeyValue(key string, v interface{}) {
	nodeValue := reflect.Indirect(reflect.ValueOf(r._node))
	fields := strings.Split(key, ".")
	value, ok := getValueByPath(nodeValue, fields)
	if !ok {
		panic(fmt.Errorf("field %s is not valid", key))
	}
	value = value.FieldByName("_value")
	if !value.IsValid() {
		panic(fmt.Errorf("field %s is not valid", key))
	}
	setValue(value, v)
}
