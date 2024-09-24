package core

import (
	"fmt"
	"math"

	"golang.org/x/crypto/bcrypt"
)

func CompareHashAndPassword(e string, p string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(e), []byte(p))
	if err != nil {
		return false, err
	}
	return true, nil
}

func FormatBytes(bytes uint64) string {
	if bytes == 0 {
		return "0 B"
	}

	units := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	exponent := int(math.Log2(float64(bytes)) / 10) // 每 10 个 log2 刻度对应一个单位（2^10 = 1024）

	// 使用 math.Min 确保 exponent 不会超出 units 数组的范围
	exponent = int(math.Min(float64(exponent), float64(len(units)-1)))

	value := float64(bytes) / math.Pow(1024, float64(exponent))

	roundedValue := int(math.Round(value))
	return fmt.Sprintf("%d%s", roundedValue, units[exponent])
}

func GetIntPointer(n int) *int {
	return &n
}
