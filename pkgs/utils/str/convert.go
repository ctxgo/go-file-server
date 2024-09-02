package str

import (
	"encoding/json"
	"strconv"
)

// convertToString 将各种数值类型转换为字符串
func ConvertToString(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case int:
		return strconv.Itoa(v), nil
	case int8:
		return strconv.FormatInt(int64(v), 10), nil
	case int16:
		return strconv.FormatInt(int64(v), 10), nil
	case int32:
		return strconv.FormatInt(int64(v), 10), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case uint:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint64:
		return strconv.FormatUint(v, 10), nil
	case float32:
		// 浮点数转换为字符串时，保留小数点后6位
		return strconv.FormatFloat(float64(v), 'f', 6, 32), nil
	case float64:
		// 浮点数转换为字符串时，保留小数点后15位
		return strconv.FormatFloat(v, 'f', 15, 64), nil
	default:
		// 对于未能直接处理的类型，使用 json.Marshal 序列化
		jsonData, err := json.Marshal(value)
		if err != nil {
			return "", err // 序列化失败返回错误
		}
		return string(jsonData), nil
	}
}
