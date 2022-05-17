package mdpimg

import (
	"testing"
)

func TestImgs(t *testing.T) {
	imgs, err := Get("https://mdpr.jp/interview/detail/3115144")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(imgs)
}
