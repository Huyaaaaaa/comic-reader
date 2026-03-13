package crypto

import (
	"testing"
)
func TestExtractKeys(t *testing.T) {
	html := `
		<script>
		var aei = 'test_iv_1234567890123456';
		var aek = 'test_key_1234567890123456';
		</script>
	`

	iv, key, err := ExtractKeys(html)
	if err != nil {
		t.Fatalf("ExtractKeys failed: %v", err)
	}

	if iv != "test_iv_1234567890123456" {
		t.Errorf("Expected iv 'test_iv_1234567890123456', got '%s'", iv)
	}

	if key != "test_key_1234567890123456" {
		t.Errorf("Expected key 'test_key_1234567890123456', got '%s'", key)
	}
}

func TestDecrypt(t *testing.T) {
	// 这里需要实际的加密文本进行测试
	// 暂时跳过
	t.Skip("需要实际的加密数据进行测试")
}
