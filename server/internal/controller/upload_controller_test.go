package controller

import "testing"

func TestDetectImageExt(t *testing.T) {
	pad := func(head []byte) []byte {
		buf := make([]byte, 512)
		copy(buf, head)
		return buf
	}
	cases := []struct {
		name string
		buf  []byte
		want string
		ok   bool
	}{
		{"jpeg", pad([]byte{0xFF, 0xD8, 0xFF, 0xE0}), ".jpg", true},
		{"png", pad([]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1A, '\n'}), ".png", true},
		{"gif", pad([]byte("GIF89a")), ".gif", true},
		{"webp", pad([]byte{'R', 'I', 'F', 'F', 0, 0, 0, 0, 'W', 'E', 'B', 'P'}), ".webp", true},
		{"text", []byte("<script>alert(1)</script>"), "", false},
		{"empty", nil, "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ext, ok := detectImageExt(c.buf)
			if ok != c.ok || ext != c.want {
				t.Fatalf("got (%q, %v), want (%q, %v)", ext, ok, c.want, c.ok)
			}
		})
	}
}

func TestUploadFilenamePattern(t *testing.T) {
	valid := []string{
		"0123456789abcdef0123456789abcdef.jpg",
		"0123456789abcdef0123456789abcdef.webp",
	}
	invalid := []string{
		"../etc/passwd",
		"0123456789abcdef0123456789abcdef.exe",
		"0123456789abcdef0123456789abcdef.JPG",
		"short.jpg",
		"0123456789abcdef0123456789abcdef.jpg.png",
	}
	for _, name := range valid {
		if !uploadFilenamePattern.MatchString(name) {
			t.Errorf("应接受安全文件名 %q", name)
		}
	}
	for _, name := range invalid {
		if uploadFilenamePattern.MatchString(name) {
			t.Errorf("应拒绝危险文件名 %q", name)
		}
	}
}
